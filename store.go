package main

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
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
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS measurements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ts TEXT NOT NULL,
			latency REAL NOT NULL DEFAULT 0,
			jitter REAL NOT NULL DEFAULT 0,
			packet_loss REAL NOT NULL DEFAULT 0,
			download REAL NOT NULL DEFAULT 0,
			upload REAL NOT NULL DEFAULT 0,
			dns REAL NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS ping_targets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			host TEXT NOT NULL UNIQUE,
			ip TEXT NOT NULL DEFAULT '',
			latency REAL NOT NULL DEFAULT 0,
			loss REAL NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'unknown',
			updated_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS dns_checks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			host TEXT NOT NULL UNIQUE,
			time_ms REAL NOT NULL DEFAULT 0,
			resolver TEXT NOT NULL DEFAULT 'system',
			updated_at TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_measurements_ts ON measurements(ts);
	`)
	return err
}

type Measurement struct {
	ID         int64   `json:"id"`
	Time       string  `json:"time"`
	Latency    float64 `json:"latency"`
	Jitter     float64 `json:"jitter"`
	PacketLoss float64 `json:"loss"`
	Download   float64 `json:"download"`
	Upload     float64 `json:"upload"`
	DNS        float64 `json:"dns"`
}

func (s *Store) SaveMeasurement(m Measurement) error {
	_, err := s.db.Exec(
		`INSERT INTO measurements (ts, latency, jitter, packet_loss, download, upload, dns)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.Time, m.Latency, m.Jitter, m.PacketLoss, m.Download, m.Upload, m.DNS,
	)
	return err
}

func (s *Store) GetHistory(limit int) ([]Measurement, error) {
	rows, err := s.db.Query(
		`SELECT id, ts, latency, jitter, packet_loss, download, upload, dns
		 FROM measurements ORDER BY ts DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Measurement
	for rows.Next() {
		var m Measurement
		var ts string
		if err := rows.Scan(&m.ID, &ts, &m.Latency, &m.Jitter, &m.PacketLoss, &m.Download, &m.Upload, &m.DNS); err != nil {
			return nil, err
		}
		t, _ := time.Parse(time.RFC3339, ts)
		m.Time = t.Format("15:04")
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

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func (s *Store) GetSummary() (Summary, error) {
	var sum Summary
	cutoff := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

	err := s.db.QueryRow(`
		SELECT
			COALESCE(AVG(latency), 0),
			COALESCE(MIN(latency), 0),
			COALESCE(MAX(latency), 0),
			COALESCE(AVG(jitter), 0),
			COALESCE(AVG(packet_loss), 0),
			COALESCE(AVG(dns), 0),
			COALESCE(AVG(download), 0),
			COALESCE(AVG(upload), 0)
		FROM measurements WHERE ts >= ?
	`, cutoff).Scan(
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

	// P95 latency
	rows, err := s.db.Query(
		`SELECT latency FROM measurements WHERE ts >= ? ORDER BY latency`, cutoff,
	)
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
	s.db.QueryRow(`SELECT COUNT(*) FROM measurements WHERE ts >= ?`, cutoff).Scan(&total)
	s.db.QueryRow(`SELECT COUNT(*) FROM measurements WHERE ts >= ? AND packet_loss < 100`, cutoff).Scan(&up)
	if total > 0 {
		sum.Uptime24h = math.Round(float64(up)/float64(total)*10000) / 100
	} else {
		sum.Uptime24h = 100
	}

	s.db.QueryRow(
		`SELECT COUNT(*) FROM measurements WHERE ts >= ? AND packet_loss >= 50`, cutoff,
	).Scan(&sum.Outages24h)

	return sum, nil
}

type PingTarget struct {
	Host    string  `json:"host"`
	IP      string  `json:"ip"`
	Latency float64 `json:"latency"`
	Loss    float64 `json:"loss"`
	Status  string  `json:"status"`
}

func (s *Store) UpsertPingTarget(t PingTarget) error {
	_, err := s.db.Exec(`
		INSERT INTO ping_targets (host, ip, latency, loss, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(host) DO UPDATE SET
			ip = excluded.ip,
			latency = excluded.latency,
			loss = excluded.loss,
			status = excluded.status,
			updated_at = excluded.updated_at
	`, t.Host, t.IP, t.Latency, t.Loss, t.Status, time.Now().Format(time.RFC3339))
	return err
}

func (s *Store) GetPingTargets() ([]PingTarget, error) {
	rows, err := s.db.Query(`SELECT host, ip, latency, loss, status FROM ping_targets ORDER BY host`)
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

type DNSCheck struct {
	Host     string  `json:"host"`
	TimeMs   float64 `json:"time_ms"`
	Resolver string  `json:"resolver"`
}

func (s *Store) UpsertDNSCheck(d DNSCheck) error {
	_, err := s.db.Exec(`
		INSERT INTO dns_checks (host, time_ms, resolver, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(host) DO UPDATE SET
			time_ms = excluded.time_ms,
			resolver = excluded.resolver,
			updated_at = excluded.updated_at
	`, d.Host, d.TimeMs, d.Resolver, time.Now().Format(time.RFC3339))
	return err
}

func (s *Store) GetDNSChecks() ([]DNSCheck, error) {
	rows, err := s.db.Query(`SELECT host, time_ms, resolver FROM dns_checks ORDER BY host`)
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

func (s *Store) Close() error {
	return s.db.Close()
}
