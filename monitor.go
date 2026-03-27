package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MonitorConfig struct {
	PingTargets   []string
	DNSTargets    []string
	PingInterval  time.Duration
	SpeedInterval time.Duration
	PingCount     int
	DownloadURL   string
	UploadURL     string
}

func DefaultConfig() MonitorConfig {
	return MonitorConfig{
		PingTargets:   []string{"google.com", "cloudflare.com"},
		DNSTargets:    []string{"google.com", "cloudflare.com"},
		PingInterval:  60 * time.Second,
		SpeedInterval: 5 * time.Minute,
		PingCount:     5,
		DownloadURL:   "https://speed.cloudflare.com/__down?bytes=10000000",
		UploadURL:     "https://speed.cloudflare.com/__up",
	}
}

type Monitor struct {
	cfg   MonitorConfig
	store *Store
	log   *slog.Logger
	wg    sync.WaitGroup

	mu       sync.RWMutex
	lastDown float64
	lastUp   float64
}

func NewMonitor(cfg MonitorConfig, store *Store, log *slog.Logger) *Monitor {
	return &Monitor{
		cfg:   cfg,
		store: store,
		log:   log,
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

// pingWorker runs ping and DNS checks on a fixed interval.
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

// speedWorker runs download/upload speed tests on a longer interval.
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

type pingResult struct {
	Host       string
	IP         string
	AvgMs      float64
	JitterMs   float64
	PacketLoss float64
	Err        error
}

func (m *Monitor) runPingCycle() {
	m.log.Info("ping cycle: start")

	allTargets := make([]string, 0, len(m.cfg.PingTargets)+1)
	gw := defaultGateway()
	if gw != "" {
		allTargets = append(allTargets, gw)
	}
	allTargets = append(allTargets, m.cfg.PingTargets...)

	// Ping all targets concurrently
	results := make([]pingResult, len(allTargets))
	var wg sync.WaitGroup
	for i, target := range allTargets {
		wg.Add(1)
		go func(idx int, t string) {
			defer wg.Done()
			avg, jitter, loss, err := runPing(t, m.cfg.PingCount)
			ip := t
			if net.ParseIP(t) == nil {
				ip = resolveIP(t)
			}
			results[idx] = pingResult{Host: t, IP: ip, AvgMs: avg, JitterMs: jitter, PacketLoss: loss, Err: err}
		}(i, target)
	}

	// DNS lookups concurrently
	dnsResults := make([]float64, len(m.cfg.DNSTargets))
	for i, host := range m.cfg.DNSTargets {
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
			m.store.UpsertDNSCheck(DNSCheck{Host: h, TimeMs: round1(ms), Resolver: "system"})
		}(i, host)
	}

	wg.Wait()

	// Persist per-target results
	for i, r := range results {
		if r.Err != nil {
			m.log.Error("ping failed", "target", r.Host, "error", r.Err)
		}
		host := r.Host
		if i == 0 && gw != "" {
			host = "gateway"
		}
		status := "up"
		if r.Err != nil || r.PacketLoss >= 100 {
			status = "down"
		}
		m.store.UpsertPingTarget(PingTarget{
			Host: host, IP: r.IP, Latency: round1(r.AvgMs), Loss: r.PacketLoss, Status: status,
		})
		m.log.Info("ping", "target", host, "latency_ms", r.AvgMs, "loss_%", r.PacketLoss)
	}

	// Aggregate across external targets (skip gateway for overall stats)
	start := 0
	if gw != "" {
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

	meas := Measurement{
		Time:       time.Now().Format(time.RFC3339),
		Latency:    round1(avgLat),
		Jitter:     round1(avgJitter),
		PacketLoss: round1(avgLoss),
		Download:   round1(down),
		Upload:     round1(up),
		DNS:        round1(dnsAvg),
	}

	if err := m.store.SaveMeasurement(meas); err != nil {
		m.log.Error("save measurement failed", "error", err)
	}
	m.log.Info("ping cycle: done", "latency", meas.Latency, "loss", meas.PacketLoss)
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

// --- low-level network probes ---

func runPing(target string, count int) (avg, jitter, loss float64, err error) {
	out, _ := exec.Command("ping", "-c", strconv.Itoa(count), "-i", "0.2", target).CombinedOutput()
	text := string(out)

	lossRe := regexp.MustCompile(`([\d.]+)% packet loss`)
	if m := lossRe.FindStringSubmatch(text); m != nil {
		loss, _ = strconv.ParseFloat(m[1], 64)
	}

	rttRe := regexp.MustCompile(`min/avg/max/stddev = [\d.]+/([\d.]+)/[\d.]+/([\d.]+)`)
	if m := rttRe.FindStringSubmatch(text); m != nil {
		avg, _ = strconv.ParseFloat(m[1], 64)
		jitter, _ = strconv.ParseFloat(m[2], 64)
	}

	return avg, jitter, loss, nil
}

func defaultGateway() string {
	out, err := exec.Command("route", "get", "default").CombinedOutput()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gateway:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "gateway:"))
		}
	}
	return ""
}

func resolveIP(host string) string {
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return ""
	}
	return ips[0].String()
}

func measureDNS(host string) (time.Duration, error) {
	start := time.Now()
	_, err := net.LookupHost(host)
	return time.Since(start), err
}

func measureDownload(url string) (float64, error) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	n, err := io.Copy(io.Discard, resp.Body)
	dur := time.Since(start).Seconds()
	if dur == 0 {
		return 0, fmt.Errorf("zero duration")
	}
	return float64(n) / dur / 1e6 * 8, err
}

func measureUpload(url string) (float64, error) {
	data := make([]byte, 10*1024*1024)
	start := time.Now()
	resp, err := http.Post(url, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	dur := time.Since(start).Seconds()
	if dur == 0 {
		return 0, fmt.Errorf("zero duration")
	}
	return float64(len(data)) / dur / 1e6 * 8, nil
}
