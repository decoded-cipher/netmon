package monitor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var speedSteps = []int64{100_000, 1_000_000, 5_000_000, 10_000_000}

const (
	speedRunsPerStep    = 2
	speedByteBudget     = 25_000_000
	speedTimeBudget     = 3 * time.Second
	speedFinishDuration = time.Second
	speedMinDuration    = 10 * time.Millisecond
	speedTestTimeout    = 60 * time.Second
)

type speedSample struct {
	bytes int64
	dur   time.Duration
}

func newSpeedClient() *http.Client {
	return &http.Client{Transport: http.DefaultTransport.(*http.Transport).Clone()}
}

func measureDownload(ctx context.Context, client *http.Client, rawURL string) (float64, error) {
	if err := warmUp(ctx, client, rawURL); err != nil {
		return 0, fmt.Errorf("warm-up: %w", err)
	}
	return runSpeedSteps(ctx, func(size int64) (speedSample, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, withBytes(rawURL, size), nil)
		if err != nil {
			return speedSample{}, err
		}
		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			return speedSample{}, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return speedSample{}, fmt.Errorf("status %s", resp.Status)
		}
		n, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			return speedSample{}, err
		}
		dur := time.Since(start) - serverTime(resp.Header)
		return speedSample{bytes: n, dur: dur}, nil
	})
}

func measureUpload(ctx context.Context, client *http.Client, rawURL string) (float64, error) {
	if err := warmUp(ctx, client, rawURL); err != nil {
		return 0, fmt.Errorf("warm-up: %w", err)
	}
	return runSpeedSteps(ctx, func(size int64) (speedSample, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(make([]byte, size)))
		if err != nil {
			return speedSample{}, err
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			return speedSample{}, err
		}
		defer resp.Body.Close()
		io.Copy(io.Discard, resp.Body)
		if resp.StatusCode != http.StatusOK {
			return speedSample{}, fmt.Errorf("status %s", resp.Status)
		}
		return speedSample{bytes: size, dur: time.Since(start)}, nil
	})
}

func runSpeedSteps(ctx context.Context, run func(size int64) (speedSample, error)) (float64, error) {
	var mbps []float64
	var used int64
	start := time.Now()

steps:
	for _, size := range speedSteps {
		for range speedRunsPerStep {
			if used+size > speedByteBudget || time.Since(start) > speedTimeBudget {
				break steps
			}
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			s, err := run(size)
			if err != nil {
				return 0, err
			}
			used += s.bytes
			if s.dur >= speedMinDuration {
				mbps = append(mbps, float64(s.bytes)*8/s.dur.Seconds()/1e6)
			}
			if s.dur >= speedFinishDuration {
				break steps
			}
		}
	}

	if len(mbps) == 0 {
		return 0, fmt.Errorf("no usable samples")
	}
	return percentile(mbps, 0.9), nil
}

func warmUp(ctx context.Context, client *http.Client, rawURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, withBytes(rawURL, 0), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}

func withBytes(rawURL string, n int64) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	if !q.Has("bytes") {
		return rawURL
	}
	q.Set("bytes", strconv.FormatInt(n, 10))
	u.RawQuery = q.Encode()
	return u.String()
}

func serverTime(h http.Header) time.Duration {
	var total float64
	for _, v := range h.Values("Server-Timing") {
		for _, metric := range strings.Split(v, ",") {
			parts := strings.Split(strings.TrimSpace(metric), ";")
			if !strings.HasPrefix(parts[0], "cfSpeed") {
				continue
			}
			for _, p := range parts[1:] {
				if d, ok := strings.CutPrefix(strings.TrimSpace(p), "dur="); ok {
					if ms, err := strconv.ParseFloat(d, 64); err == nil {
						total += ms
					}
				}
			}
		}
	}
	return time.Duration(total * float64(time.Millisecond))
}

func percentile(vals []float64, p float64) float64 {
	sorted := append([]float64(nil), vals...)
	sort.Float64s(sorted)
	idx := p * float64(len(sorted)-1)
	lo, hi := int(math.Floor(idx)), int(math.Ceil(idx))
	return sorted[lo] + (sorted[hi]-sorted[lo])*(idx-float64(lo))
}
