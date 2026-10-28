package monitor

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

func TestIsTrouble(t *testing.T) {
	tests := []struct {
		name                   string
		avgLat, baseline, loss float64
		latCnt                 int
		want                   bool
	}{
		{"healthy", 20, 20, 0, 2, false},
		{"loss at threshold", 20, 20, 5, 2, true},
		{"minor loss", 20, 20, 2, 2, false},
		{"total outage", 0, 20, 100, 0, false},
		{"latency doubled and +30ms", 60, 25, 0, 2, true},
		{"latency doubled but tiny", 8, 3, 0, 2, false},
		{"latency +40ms but not doubled", 100, 60, 0, 2, false},
		{"no baseline yet", 500, 0, 0, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTrouble(tt.avgLat, tt.baseline, tt.loss, tt.latCnt); got != tt.want {
				t.Errorf("isTrouble = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckTroubleRequestsOncePerEpisode(t *testing.T) {
	m := New(Config{}, nil, slog.New(slog.DiscardHandler))
	requested := func() bool {
		select {
		case <-m.speedReq:
			return true
		default:
			return false
		}
	}

	m.checkTrouble(20, 0, 2)
	if requested() {
		t.Fatal("healthy first cycle requested a test")
	}

	m.checkTrouble(20, 10, 2)
	if !requested() {
		t.Fatal("packet loss did not request a test")
	}

	m.checkTrouble(20, 10, 2)
	if requested() {
		t.Fatal("ongoing trouble requested a second test")
	}

	m.checkTrouble(20, 0, 2)
	m.checkTrouble(20, 10, 2)
	if requested() {
		t.Fatal("new episode within cooldown requested a test")
	}

	m.lastAutoAt = time.Now().Add(-speedAutoCooldown - time.Minute)
	m.checkTrouble(20, 10, 2)
	if !requested() {
		t.Fatal("trouble after cooldown did not request a test")
	}
}

func TestCheckTroubleLatencyBaseline(t *testing.T) {
	m := New(Config{}, nil, slog.New(slog.DiscardHandler))
	for range 5 {
		m.checkTrouble(20, 0, 2)
	}
	assertNear(t, "baseline", m.latBaseline, 20)

	m.checkTrouble(80, 0, 2)
	select {
	case <-m.speedReq:
	default:
		t.Fatal("latency spike did not request a test")
	}
	assertNear(t, "baseline after spike", m.latBaseline, 26)
}

func TestRunSpeedTestRejectsConcurrent(t *testing.T) {
	m := New(Config{}, nil, slog.New(slog.DiscardHandler))
	m.speedMu.Lock()
	defer m.speedMu.Unlock()
	if _, err := m.RunSpeedTest(context.Background()); !errors.Is(err, ErrSpeedTestRunning) {
		t.Fatalf("err = %v, want ErrSpeedTestRunning", err)
	}
}
