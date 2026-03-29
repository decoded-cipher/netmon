package store

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

//go:embed queries.sql
var queriesRaw string

// q holds all named queries parsed from queries.sql.
var q = parseQueries(queriesRaw)

// parseQueries splits queries.sql into a map keyed by "-- name: <key>" markers.
func parseQueries(data string) map[string]string {
	result := make(map[string]string)
	var name string
	var lines []string
	for _, line := range strings.Split(data, "\n") {
		if trimmed := strings.TrimSpace(line); strings.HasPrefix(trimmed, "-- name:") {
			if name != "" {
				result[name] = strings.TrimSpace(strings.Join(lines, "\n"))
			}
			name = strings.TrimSpace(strings.TrimPrefix(trimmed, "-- name:"))
			lines = nil
		} else if name != "" {
			lines = append(lines, line)
		}
	}
	if name != "" {
		result[name] = strings.TrimSpace(strings.Join(lines, "\n"))
	}
	return result
}

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	if _, err := s.db.Exec(schemaSQL); err != nil {
		return err
	}
	// Add network_id to measurements table for pre-WiFi databases (no-op if already exists).
	s.db.Exec(`ALTER TABLE measurements ADD COLUMN network_id TEXT NOT NULL DEFAULT ''`)
	return nil
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

// --- Measurement ---

type Measurement struct {
	ID           int64   `json:"id"`
	Time         string  `json:"time"`
	NetworkID    string  `json:"network_id"`
	Latency      float64 `json:"latency"`
	Jitter       float64 `json:"jitter"`
	PacketLoss   float64 `json:"loss"`
	Download     float64 `json:"download"`
	Upload       float64 `json:"upload"`
	DNS          float64 `json:"dns"`
	ConnType     string  `json:"conn_type"`
	ConnRSSI     int     `json:"conn_rssi"`
	ConnNoise    int     `json:"conn_noise"`
	ConnSNR      int     `json:"conn_snr"`
	ConnChannel  int     `json:"conn_channel"`
	ConnBand     string  `json:"conn_band"`
	ConnLinkRate int     `json:"conn_link_rate"`
	ConnDuplex   string  `json:"conn_duplex"`
}

func (s *Store) SaveMeasurement(m Measurement) error {
	_, err := s.db.Exec(q["insert_measurement"],
		m.Time, m.NetworkID, m.Latency, m.Jitter, m.PacketLoss, m.Download, m.Upload, m.DNS,
		m.ConnType, m.ConnRSSI, m.ConnNoise, m.ConnSNR, m.ConnChannel, m.ConnBand, m.ConnLinkRate, m.ConnDuplex,
	)
	return err
}

func (s *Store) GetHistory(limit int) ([]Measurement, error) {
	rows, err := s.db.Query(q["get_history"], limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Measurement
	for rows.Next() {
		var m Measurement
		var ts string
		if err := rows.Scan(
			&m.ID, &ts, &m.NetworkID, &m.Latency, &m.Jitter, &m.PacketLoss, &m.Download, &m.Upload, &m.DNS,
			&m.ConnType, &m.ConnRSSI, &m.ConnNoise, &m.ConnSNR, &m.ConnChannel, &m.ConnBand, &m.ConnLinkRate, &m.ConnDuplex,
		); err != nil {
			return nil, err
		}
		t, _ := time.Parse(time.RFC3339, ts)
		m.Time = t.Format("15:04:05")
		results = append(results, m)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})

	if results == nil {
		results = []Measurement{}
	}
	return results, nil
}

func (s *Store) GetHistoryWindow(minutes int) ([]Measurement, error) {
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute).Format(time.RFC3339)
	rows, err := s.db.Query(q["get_history_window"], cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timeFmt string
	switch {
	case minutes > 1440:
		timeFmt = "Jan 02 15:04"
	case minutes > 60:
		timeFmt = "15:04"
	default:
		timeFmt = "15:04:05"
	}

	var results []Measurement
	for rows.Next() {
		var m Measurement
		var ts string
		if err := rows.Scan(
			&m.ID, &ts, &m.NetworkID, &m.Latency, &m.Jitter, &m.PacketLoss, &m.Download, &m.Upload, &m.DNS,
			&m.ConnType, &m.ConnRSSI, &m.ConnNoise, &m.ConnSNR, &m.ConnChannel, &m.ConnBand, &m.ConnLinkRate, &m.ConnDuplex,
		); err != nil {
			return nil, err
		}
		t, _ := time.Parse(time.RFC3339, ts)
		m.Time = t.Format(timeFmt)
		results = append(results, m)
	}

	if results == nil {
		results = []Measurement{}
	}
	return results, nil
}

// --- Summary ---

type Summary struct {
	LatencyAvg  float64 `json:"latency_avg"`
	LatencyP95  float64 `json:"latency_p95"`
	LatencyMin  float64 `json:"latency_min"`
	LatencyMax  float64 `json:"latency_max"`
	JitterAvg   float64 `json:"jitter_avg"`
	PacketLoss  float64 `json:"packet_loss"`
	DNSAvg      float64 `json:"dns_avg"`
	DownloadAvg float64 `json:"download_avg"`
	UploadAvg   float64 `json:"upload_avg"`
	Uptime24h   float64 `json:"uptime_24h"`
	Outages24h  int     `json:"outages_24h"`
}

func (s *Store) GetSummary() (Summary, error) {
	var sum Summary
	cutoff := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

	err := s.db.QueryRow(q["get_summary_stats"], cutoff).Scan(
		&sum.LatencyAvg, &sum.LatencyMin, &sum.LatencyMax,
		&sum.JitterAvg, &sum.PacketLoss, &sum.DNSAvg,
		&sum.DownloadAvg, &sum.UploadAvg,
	)
	if err != nil {
		return sum, err
	}

	sum.LatencyAvg = round1(sum.LatencyAvg)
	sum.LatencyMin = round1(sum.LatencyMin)
	sum.LatencyMax = round1(sum.LatencyMax)
	sum.JitterAvg = round1(sum.JitterAvg)
	sum.PacketLoss = round1(sum.PacketLoss)
	sum.DNSAvg = round1(sum.DNSAvg)
	sum.DownloadAvg = round1(sum.DownloadAvg)
	sum.UploadAvg = round1(sum.UploadAvg)

	rows, err := s.db.Query(q["get_latencies"], cutoff)
	if err == nil {
		defer rows.Close()
		var latencies []float64
		for rows.Next() {
			var l float64
			rows.Scan(&l)
			latencies = append(latencies, l)
		}
		if len(latencies) > 0 {
			idx := int(float64(len(latencies)) * 0.95)
			if idx >= len(latencies) {
				idx = len(latencies) - 1
			}
			sum.LatencyP95 = round1(latencies[idx])
		}
	}

	var total, up int
	s.db.QueryRow(q["count_measurements"], cutoff).Scan(&total)
	s.db.QueryRow(q["count_up_measurements"], cutoff).Scan(&up)
	if total > 0 {
		sum.Uptime24h = math.Round(float64(up)/float64(total)*10000) / 100
	} else {
		sum.Uptime24h = 100
	}

	s.db.QueryRow(q["count_outages"], cutoff).Scan(&sum.Outages24h)

	return sum, nil
}

// --- PingTarget ---

type PingTarget struct {
	Host    string  `json:"host"`
	IP      string  `json:"ip"`
	Latency float64 `json:"latency"`
	Loss    float64 `json:"loss"`
	Status  string  `json:"status"`
}

func (s *Store) UpsertPingTarget(t PingTarget) error {
	_, err := s.db.Exec(q["upsert_ping_target"],
		t.Host, t.IP, t.Latency, t.Loss, t.Status, time.Now().Format(time.RFC3339))
	return err
}

func (s *Store) GetPingTargets() ([]PingTarget, error) {
	rows, err := s.db.Query(q["get_ping_targets"])
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	targets := []PingTarget{}
	for rows.Next() {
		var t PingTarget
		if err := rows.Scan(&t.Host, &t.IP, &t.Latency, &t.Loss, &t.Status); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}

// --- DNSCheck ---

type DNSCheck struct {
	Host     string  `json:"host"`
	TimeMs   float64 `json:"time_ms"`
	Resolver string  `json:"resolver"`
}

func (s *Store) UpsertDNSCheck(d DNSCheck) error {
	_, err := s.db.Exec(q["upsert_dns_check"],
		d.Host, d.TimeMs, d.Resolver, time.Now().Format(time.RFC3339))
	return err
}

func (s *Store) GetDNSChecks() ([]DNSCheck, error) {
	rows, err := s.db.Query(q["get_dns_checks"])
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checks := []DNSCheck{}
	for rows.Next() {
		var d DNSCheck
		if err := rows.Scan(&d.Host, &d.TimeMs, &d.Resolver); err != nil {
			return nil, err
		}
		checks = append(checks, d)
	}
	return checks, nil
}

// --- NetworkEvent ---

type NetworkEvent struct {
	ID      int64  `json:"id"`
	Time    string `json:"time"`
	Network string `json:"network_id"`
	SSID    string `json:"ssid"`
	Gateway string `json:"gateway"`
}

func (s *Store) LogNetworkEvent(e NetworkEvent) error {
	_, err := s.db.Exec(q["log_network_event"], e.Time, e.Network, e.SSID, e.Gateway)
	return err
}

func (s *Store) GetCurrentNetworkID() string {
	var id string
	s.db.QueryRow(q["get_current_network_id"]).Scan(&id)
	return id
}

// --- ConfigSettings ---

// ConfigSettings is the persisted, user-editable subset of the monitor config.
type ConfigSettings struct {
	PingTargets    []string `json:"ping_targets"`
	DNSTargets     []string `json:"dns_targets"`
	PingIntervalS  int      `json:"ping_interval_s"`
	SpeedIntervalM int      `json:"speed_interval_m"`
	PingCount      int      `json:"ping_count"`
}

func (s *Store) GetConfig() (ConfigSettings, error) {
	var raw string
	err := s.db.QueryRow(q["get_config"]).Scan(&raw)
	if err == sql.ErrNoRows {
		return ConfigSettings{}, sql.ErrNoRows
	}
	if err != nil {
		return ConfigSettings{}, err
	}
	var cs ConfigSettings
	return cs, json.Unmarshal([]byte(raw), &cs)
}

func (s *Store) SaveConfig(cs ConfigSettings) error {
	b, err := json.Marshal(cs)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(q["save_config"], string(b))
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}
