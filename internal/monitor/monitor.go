package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net"
	"sync"
	"time"

	"netmon/internal/network"
	"netmon/internal/store"
)

type Config struct {
	PingTargets   []string
	DNSTargets    []string
	PingInterval  time.Duration
	SpeedInterval time.Duration
	PingCount     int
	DownloadURL   string
	UploadURL     string
}

func DefaultConfig() Config {
	return Config{
		PingTargets:   []string{"google.com", "cloudflare.com"},
		DNSTargets:    []string{"google.com", "cloudflare.com"},
		PingInterval:  60 * time.Second,
		SpeedInterval: 30 * time.Minute,
		PingCount:     5,
		DownloadURL:   "https://speed.cloudflare.com/__down?bytes=1000000",
		UploadURL:     "https://speed.cloudflare.com/__up",
	}
}

type Monitor struct {
	cfgMu sync.RWMutex
	cfg   Config

	store *store.Store
	log   *slog.Logger
	wg    sync.WaitGroup

	mu       sync.RWMutex
	lastDown float64
	lastUp   float64

	currentNetworkID string
}

func New(cfg Config, s *store.Store, log *slog.Logger) *Monitor {
	return &Monitor{cfg: cfg, store: s, log: log}
}

// GetConfig returns a snapshot of the current config.
func (m *Monitor) GetConfig() Config {
	m.cfgMu.RLock()
	defer m.cfgMu.RUnlock()
	return m.cfg
}

// SetConfig replaces the live config. Targets and ping count take effect on the
// next ping cycle. Interval changes take effect after a restart.
func (m *Monitor) SetConfig(cfg Config) {
	m.cfgMu.Lock()
	m.cfg = cfg
	m.cfgMu.Unlock()
}

// ConfigFromStore converts persisted settings into a full Config,
// filling URL fields from the compile-time defaults.
func ConfigFromStore(cs store.ConfigSettings) Config {
	def := DefaultConfig()
	return Config{
		PingTargets:   cs.PingTargets,
		DNSTargets:    cs.DNSTargets,
		PingInterval:  time.Duration(cs.PingIntervalS) * time.Second,
		SpeedInterval: time.Duration(cs.SpeedIntervalM) * time.Minute,
		PingCount:     cs.PingCount,
		DownloadURL:   def.DownloadURL,
		UploadURL:     def.UploadURL,
	}
}

func (m *Monitor) Start(ctx context.Context) {
	m.wg.Add(2)

	go func() {
		defer m.wg.Done()
		m.pingWorker(ctx)
	}()

	go func() {
		defer m.wg.Done()
		m.speedWorker(ctx)
	}()
}

func (m *Monitor) Wait() {
	m.wg.Wait()
}

func (m *Monitor) pingWorker(ctx context.Context) {
	m.runPingCycle()

	ticker := time.NewTicker(m.cfg.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.log.Info("ping worker stopped")
			return
		case <-ticker.C:
			m.runPingCycle()
		}
	}
}

func (m *Monitor) speedWorker(ctx context.Context) {
	m.runSpeedCycle()

	ticker := time.NewTicker(m.cfg.SpeedInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.log.Info("speed worker stopped")
			return
		case <-ticker.C:
			m.runSpeedCycle()
		}
	}
}

func (m *Monitor) runPingCycle() {
	fmt.Println()
	m.log.Info("ping cycle: start")

	cfg := m.GetConfig() // snapshot config for this cycle

	// Detect current network; log an event if it changed.
	netInfo := network.Detect()
	if netInfo.ID != m.currentNetworkID {
		if m.currentNetworkID != "" {
			m.log.Info("network changed", "from", m.currentNetworkID, "to", netInfo.ID)
		}
		m.currentNetworkID = netInfo.ID
		if err := m.store.LogNetworkEvent(store.NetworkEvent{
			Time:    time.Now().Format(time.RFC3339),
			Network: netInfo.ID,
			SSID:    netInfo.SSID,
			Gateway: netInfo.Gateway,
		}); err != nil {
			m.log.Error("log network event failed", "error", err)
		}
	}

	allTargets := make([]string, 0, len(cfg.PingTargets)+1)
	if netInfo.Gateway != "" {
		allTargets = append(allTargets, netInfo.Gateway)
	}
	allTargets = append(allTargets, cfg.PingTargets...)

	results := make([]pingResult, len(allTargets))
	var wg sync.WaitGroup
	var connInfo network.ConnectionInfo
	wg.Go(func() {
		connInfo, _ = network.GetConnectionInfo()
	})

	for i, target := range allTargets {
		wg.Add(1)
		go func(idx int, t string) {
			defer wg.Done()
			avg, jitter, loss, err := runPing(t, cfg.PingCount)
			ip := t
			if net.ParseIP(t) == nil {
				ip = resolveIP(t)
			}
			results[idx] = pingResult{Host: t, IP: ip, AvgMs: avg, JitterMs: jitter, PacketLoss: loss, Err: err}
		}(i, target)
	}

	dnsResults := make([]float64, len(cfg.DNSTargets))
	for i, host := range cfg.DNSTargets {
		wg.Add(1)
		go func(idx int, h string) {
			defer wg.Done()
			d, err := measureDNS(h)
			ms := float64(d.Microseconds()) / 1000.0
			dnsResults[idx] = ms
			if err != nil {
				m.log.Error("dns failed", "host", h, "error", err)
				return
			}
			m.store.UpsertDNSCheck(store.DNSCheck{Host: h, TimeMs: round1(ms), Resolver: "system"})
		}(i, host)
	}

	wg.Wait()

	for i, r := range results {
		if r.Err != nil {
			m.log.Error("ping failed", "target", r.Host, "error", r.Err)
		}
		host := r.Host
		if i == 0 && netInfo.Gateway != "" {
			host = "gateway"
		}
		status := "up"
		if r.Err != nil || r.PacketLoss >= 100 {
			status = "down"
		}
		m.store.UpsertPingTarget(store.PingTarget{
			Host: host, IP: r.IP, Latency: round1(r.AvgMs), Loss: r.PacketLoss, Status: status,
		})
		m.log.Info("ping", "target", host, "latency_ms", r.AvgMs, "loss_%", r.PacketLoss)
	}

	start := 0
	if netInfo.Gateway != "" {
		start = 1
	}
	var totalLat, totalJitter, totalLoss float64
	var cnt int
	for _, r := range results[start:] {
		if r.Err != nil {
			continue
		}
		totalLat += r.AvgMs
		totalJitter += r.JitterMs
		totalLoss += r.PacketLoss
		cnt++
	}

	var dnsAvg float64
	for _, d := range dnsResults {
		dnsAvg += d
	}
	if len(dnsResults) > 0 {
		dnsAvg /= float64(len(dnsResults))
	}

	avgLat := totalLat / float64(max(cnt, 1))
	avgJitter := totalJitter / float64(max(cnt, 1))
	avgLoss := totalLoss / float64(max(cnt, 1))

	m.mu.RLock()
	down := m.lastDown
	up := m.lastUp
	m.mu.RUnlock()

	meas := store.Measurement{
		Time:         time.Now().Format(time.RFC3339),
		NetworkID:    m.currentNetworkID,
		Latency:      round1(avgLat),
		Jitter:       round1(avgJitter),
		PacketLoss:   round1(avgLoss),
		Download:     round1(down),
		Upload:       round1(up),
		DNS:          round1(dnsAvg),
		ConnType:     connInfo.Type,
		ConnRSSI:     connInfo.RSSI,
		ConnNoise:    connInfo.Noise,
		ConnSNR:      connInfo.SNR,
		ConnChannel:  connInfo.Channel,
		ConnBand:     connInfo.Band,
		ConnLinkRate: connInfo.LinkRate,
		ConnDuplex:   connInfo.Duplex,
	}

	if err := m.store.SaveMeasurement(meas); err != nil {
		m.log.Error("save measurement failed", "error", err)
	}
	m.log.Info("ping cycle: done", "latency", meas.Latency, "loss", meas.PacketLoss, "network", meas.NetworkID)
}

func (m *Monitor) runSpeedCycle() {
	m.log.Info("speed test: start")

	down, err := measureDownload(m.cfg.DownloadURL)
	if err != nil {
		m.log.Error("download test failed", "error", err)
	} else {
		m.log.Info("download", "mbps", math.Round(down*100)/100)
	}

	up, err := measureUpload(m.cfg.UploadURL)
	if err != nil {
		m.log.Error("upload test failed", "error", err)
	} else {
		m.log.Info("upload", "mbps", math.Round(up*100)/100)
	}

	m.mu.Lock()
	if down > 0 {
		m.lastDown = math.Round(down*100) / 100
	}
	if up > 0 {
		m.lastUp = math.Round(up*100) / 100
	}
	m.mu.Unlock()

	m.log.Info("speed test: done")
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
