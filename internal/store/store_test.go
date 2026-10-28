package store

import (
	"database/sql"
	"math"
	"path/filepath"
	"testing"
	"time"
)

func ptr(v float64) *float64 { return &v }

func TestSummarySkipsMissingData(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "netmon.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	now := time.Now().Format(time.RFC3339)
	rows := []Measurement{
		{Time: now, Latency: ptr(20), Jitter: ptr(2), Download: ptr(60), Upload: ptr(40), DNS: ptr(10)},
		{Time: now, Latency: ptr(30), Jitter: ptr(4), Download: ptr(50), Upload: ptr(30), DNS: ptr(0)},
		{Time: now, PacketLoss: 100},
	}
	for _, m := range rows {
		if err := s.SaveMeasurement(m); err != nil {
			t.Fatal(err)
		}
	}

	sum, err := s.GetSummary()
	if err != nil {
		t.Fatal(err)
	}
	checks := []struct {
		name string
		got  *float64
		want float64
	}{
		{"latency_avg", sum.LatencyAvg, 25},
		{"latency_min", sum.LatencyMin, 20},
		{"latency_max", sum.LatencyMax, 30},
		{"latency_p95", sum.LatencyP95, 30},
		{"jitter_avg", sum.JitterAvg, 3},
		{"dns_avg", sum.DNSAvg, 5},
		{"download_avg", sum.DownloadAvg, 55},
		{"upload_avg", sum.UploadAvg, 35},
	}
	for _, c := range checks {
		if c.got == nil || *c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, deref(c.got), c.want)
		}
	}
	if sum.PacketLoss != 33.3 {
		t.Errorf("packet_loss = %v, want 33.3", sum.PacketLoss)
	}
	if sum.Outages24h != 1 {
		t.Errorf("outages_24h = %d, want 1", sum.Outages24h)
	}
}

func TestSummaryWithoutData(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "netmon.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := s.SaveMeasurement(Measurement{Time: time.Now().Format(time.RFC3339), Latency: ptr(15)}); err != nil {
		t.Fatal(err)
	}
	sum, err := s.GetSummary()
	if err != nil {
		t.Fatal(err)
	}
	if sum.DownloadAvg != nil || sum.UploadAvg != nil || sum.DNSAvg != nil {
		t.Errorf("unmeasured averages = %v/%v/%v, want nil", deref(sum.DownloadAvg), deref(sum.UploadAvg), deref(sum.DNSAvg))
	}

	h, err := s.GetHistory(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != 1 || h[0].Download != nil || deref(h[0].Latency) != 15 {
		t.Errorf("history = %+v, want one row with latency 15 and nil download", h)
	}
}

const legacySchema = `
CREATE TABLE measurements (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	ts             TEXT    NOT NULL,
	network_id     TEXT    NOT NULL DEFAULT '',
	latency        REAL    NOT NULL DEFAULT 0,
	jitter         REAL    NOT NULL DEFAULT 0,
	packet_loss    REAL    NOT NULL DEFAULT 0,
	download       REAL    NOT NULL DEFAULT 0,
	upload         REAL    NOT NULL DEFAULT 0,
	dns            REAL    NOT NULL DEFAULT 0,
	conn_type      TEXT    NOT NULL DEFAULT '',
	conn_rssi      INTEGER NOT NULL DEFAULT 0,
	conn_noise     INTEGER NOT NULL DEFAULT 0,
	conn_snr       INTEGER NOT NULL DEFAULT 0,
	conn_channel   INTEGER NOT NULL DEFAULT 0,
	conn_band      TEXT    NOT NULL DEFAULT '',
	conn_link_rate INTEGER NOT NULL DEFAULT 0,
	conn_duplex    TEXT    NOT NULL DEFAULT ''
);
CREATE INDEX idx_measurements_ts ON measurements(ts);
INSERT INTO measurements (ts, network_id, latency, jitter, packet_loss, download, upload, dns, conn_type, conn_rssi)
VALUES
	('2026-10-01T10:00:00Z', 'home', 20, 2, 0,   60, 40, 0.0, 'wifi', -55),
	('2026-10-01T10:01:00Z', 'home', 0,  0, 100, 0,  0,  12,  'wifi', -80);
`

func TestMigrateLegacyDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "netmon.db")
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(legacySchema); err != nil {
		t.Fatal(err)
	}
	raw.Close()

	for run := 1; run <= 2; run++ {
		s, err := New(path)
		if err != nil {
			t.Fatalf("open %d: %v", run, err)
		}
		var version, indexes int
		s.db.QueryRow(`PRAGMA user_version`).Scan(&version)
		s.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = 'idx_measurements_ts'`).Scan(&indexes)
		if version != 1 || indexes != 1 {
			t.Errorf("open %d: user_version = %d, index count = %d, want 1 and 1", run, version, indexes)
		}

		h, err := s.GetHistory(10)
		if err != nil {
			t.Fatal(err)
		}
		if len(h) != 2 {
			t.Fatalf("open %d: got %d rows, want 2", run, len(h))
		}
		ok, outage := h[0], h[1]
		if deref(ok.Latency) != 20 || deref(ok.Jitter) != 2 || deref(ok.Download) != 60 || deref(ok.Upload) != 40 {
			t.Errorf("open %d: measured row = %+v", run, ok)
		}
		if ok.DNS == nil || *ok.DNS != 0 {
			t.Errorf("open %d: legacy dns 0.0 should be kept, got %v", run, ok.DNS)
		}
		if ok.ConnType != "wifi" || ok.ConnRSSI != -55 || ok.NetworkID != "home" {
			t.Errorf("open %d: connection fields not preserved: %+v", run, ok)
		}
		if outage.Latency != nil || outage.Jitter != nil || outage.Download != nil || outage.Upload != nil {
			t.Errorf("open %d: outage row zeros not converted to NULL: %+v", run, outage)
		}
		if outage.PacketLoss != 100 || deref(outage.DNS) != 12 {
			t.Errorf("open %d: outage row = %+v", run, outage)
		}

		if run == 1 {
			if err := s.SaveMeasurement(Measurement{Time: time.Now().Format(time.RFC3339), Latency: ptr(9)}); err != nil {
				t.Fatal(err)
			}
			s.db.Exec(`DELETE FROM measurements WHERE latency = 9`)
		}
		s.Close()
	}
}

func deref(v *float64) float64 {
	if v == nil {
		return math.NaN()
	}
	return *v
}
