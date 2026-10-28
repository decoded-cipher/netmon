package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSummarySkipsMissingData(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "netmon.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	now := time.Now().Format(time.RFC3339)
	rows := []Measurement{
		{Time: now, Latency: 20, Jitter: 2, Download: 60, Upload: 40},
		{Time: now, Latency: 30, Jitter: 4, Download: 50, Upload: 30},
		{Time: now, Latency: 0, Jitter: 0, PacketLoss: 100, Download: 0, Upload: 0},
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
		name      string
		got, want float64
	}{
		{"latency_avg", sum.LatencyAvg, 25},
		{"latency_min", sum.LatencyMin, 20},
		{"latency_max", sum.LatencyMax, 30},
		{"latency_p95", sum.LatencyP95, 30},
		{"jitter_avg", sum.JitterAvg, 3},
		{"download_avg", sum.DownloadAvg, 55},
		{"upload_avg", sum.UploadAvg, 35},
		{"packet_loss", sum.PacketLoss, 33.3},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
	if sum.Outages24h != 1 {
		t.Errorf("outages_24h = %d, want 1", sum.Outages24h)
	}
}
