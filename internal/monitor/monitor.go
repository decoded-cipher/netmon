package monitor

import (
	"context"
	"errors"
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
		SpeedInterval: 60 * time.Minute,
		PingCount:     5,
		DownloadURL:   "https://speed.cloudflare.com/__down?bytes=1000000",
		UploadURL:     "https://speed.cloudflare.com/__up",
	}
}

const (
	speedAutoCooldown    = time.Hour
	troubleLossPct       = 5
	troubleLatencyRatio  = 2
	troubleLatencyMinMs  = 30
	latencyBaselineAlpha = 0.1
)

var ErrSpeedTestRunning = errors.New("a speed test is already running")

// SpeedTestStatus describes the most recent speed test.
type SpeedTestStatus struct {
	Running  bool    `json:"running"`
	LastAt   string  `json:"last_at"`
	Trigger  string  `json:"trigger"`
	Download float64 `json:"download"`
	Upload   float64 `json:"upload"`
	Error    string  `json:"error,omitempty"`
}

type Monitor struct {
	cfgMu sync.RWMutex
	cfg   Config

	store *store.Store
	log   *slog.Logger
	wg    sync.WaitGroup

	speedMu  sync.Mutex
	speedReq chan struct{}

	mu        sync.RWMutex
	lastDown  float64
	lastUp    float64
	speedStat SpeedTestStatus

	currentNetworkID string
	latBaseline      float64
	troubleTested    bool
	lastAutoAt       time.Time
}

func New(cfg Config, s *store.Store, log *slog.Logger) *Monitor {
	return &Monitor{cfg: cfg, store: s, log: log, speedReq: make(chan struct{}, 1)}
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
	m.runSpeedCycle(ctx, "startup")

	ticker := time.NewTicker(m.cfg.SpeedInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.log.Info("speed worker stopped")
			return
		case <-ticker.C:
			m.runSpeedCycle(ctx, "scheduled")
		case <-m.speedReq:
			m.runSpeedCycle(ctx, "auto")
		}
	}
}

// RunSpeedTest runs a speed test immediately and returns its result.
func (m *Monitor) RunSpeedTest(ctx context.Context) (SpeedTestStatus, error) {
	return m.runSpeedCycle(ctx, "manual")
}

// SpeedTestStatus returns the state of the most recent speed test.
func (m *Monitor) SpeedTestStatus() SpeedTestStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.speedStat
}

// isTrouble reports whether a ping cycle looks degraded enough to warrant a speed test.
func isTrouble(avgLat, baseline, avgLoss float64, latCnt int) bool {
	if latCnt == 0 {
		return false
	}
	if avgLoss >= troubleLossPct {
		return true
	}
	return baseline > 0 && avgLat > baseline*troubleLatencyRatio && avgLat-baseline >= troubleLatencyMinMs
}

func (m *Monitor) checkTrouble(avgLat, avgLoss float64, latCnt int) {
	trouble := isTrouble(avgLat, m.latBaseline, avgLoss, latCnt)
	if latCnt > 0 {
		if m.latBaseline == 0 {
			m.latBaseline = avgLat
		} else {
			m.latBaseline += latencyBaselineAlpha * (avgLat - m.latBaseline)
		}
	}

	if !trouble {
		m.troubleTested = false
		return
	}
	if m.troubleTested || time.Since(m.lastAutoAt) < speedAutoCooldown {
		return
	}
	select {
	case m.speedReq <- struct{}{}:
		m.log.Info("connection degraded, requesting speed test", "latency", round1(avgLat), "loss", round1(avgLoss))
		m.troubleTested = true
		m.lastAutoAt = time.Now()
	default:
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
		m.latBaseline = 0
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
	dnsOK := make([]bool, len(cfg.DNSTargets))
	for i, host := range cfg.DNSTargets {
		wg.Add(1)
		go func(idx int, h string) {
			defer wg.Done()
			d, err := measureDNS(h)
			ms := float64(d.Microseconds()) / 1000.0
			if err != nil {
				m.log.Error("dns failed", "host", h, "error", err)
				return
			}
			dnsResults[idx] = ms
			dnsOK[idx] = true
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
	var cnt, latCnt int
	for _, r := range results[start:] {
		if r.Err != nil {
			continue
		}
		totalLoss += r.PacketLoss
		cnt++
		if r.PacketLoss >= 100 {
			continue
		}
		totalLat += r.AvgMs
		totalJitter += r.JitterMs
		latCnt++
	}

	var dnsAvg float64
	var dnsCnt int
	for i, d := range dnsResults {
		if dnsOK[i] {
			dnsAvg += d
			dnsCnt++
		}
	}
	dnsAvg /= float64(max(dnsCnt, 1))

	avgLat := totalLat / float64(max(latCnt, 1))
	avgJitter := totalJitter / float64(max(latCnt, 1))
	avgLoss := totalLoss / float64(max(cnt, 1))
	if cnt == 0 && len(results) > start {
		avgLoss = 100
	}
	m.checkTrouble(avgLat, avgLoss, latCnt)

	m.mu.RLock()
	down := m.lastDown
	up := m.lastUp
	m.mu.RUnlock()

	meas := store.Measurement{
		Time:         time.Now().Format(time.RFC3339),
		NetworkID:    m.currentNetworkID,
		Latency:      measured(avgLat, latCnt > 0),
		Jitter:       measured(avgJitter, latCnt > 0),
		PacketLoss:   round1(avgLoss),
		Download:     measured(down, down > 0),
		Upload:       measured(up, up > 0),
		DNS:          measured(dnsAvg, dnsCnt > 0),
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
	m.log.Info("ping cycle: done", "latency", round1(avgLat), "loss", meas.PacketLoss, "network", meas.NetworkID)
}

func (m *Monitor) runSpeedCycle(ctx context.Context, trigger string) (SpeedTestStatus, error) {
	if !m.speedMu.TryLock() {
		m.log.Info("speed test skipped, one is already running", "trigger", trigger)
		return m.SpeedTestStatus(), ErrSpeedTestRunning
	}
	defer m.speedMu.Unlock()

	m.mu.Lock()
	m.speedStat.Running = true
	m.mu.Unlock()

	m.log.Info("speed test: start", "trigger", trigger)

	cfg := m.GetConfig()
	ctx, cancel := context.WithTimeout(ctx, speedTestTimeout)
	defer cancel()
	client := newSpeedClient()
	defer client.CloseIdleConnections()

	down, downErr := measureDownload(ctx, client, cfg.DownloadURL)
	if downErr != nil {
		m.log.Error("download test failed", "error", downErr)
		downErr = fmt.Errorf("download: %w", downErr)
	} else {
		m.log.Info("download", "mbps", math.Round(down*100)/100)
	}

	up, upErr := measureUpload(ctx, client, cfg.UploadURL)
	if upErr != nil {
		m.log.Error("upload test failed", "error", upErr)
		upErr = fmt.Errorf("upload: %w", upErr)
	} else {
		m.log.Info("upload", "mbps", math.Round(up*100)/100)
	}
	err := errors.Join(downErr, upErr)

	m.mu.Lock()
	m.lastDown = math.Round(down*100) / 100
	m.lastUp = math.Round(up*100) / 100
	m.speedStat = SpeedTestStatus{
		LastAt:   time.Now().Format(time.RFC3339),
		Trigger:  trigger,
		Download: round1(down),
		Upload:   round1(up),
	}
	if err != nil {
		m.speedStat.Error = err.Error()
	}
	stat := m.speedStat
	m.mu.Unlock()

	m.log.Info("speed test: done", "trigger", trigger)
	return stat, err
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func measured(v float64, ok bool) *float64 {
	if !ok {
		return nil
	}
	r := round1(v)
	return &r
}
