package monitor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

const (
	speedDownloadBytes = 20_000_000
	speedUploadBytes   = 10_000_000
	speedMaxDuration   = 2 * time.Second
	speedRampFraction  = 0.4
	speedMinDuration   = 10 * time.Millisecond
	speedTestTimeout   = 60 * time.Second
)

type speedPoint struct {
	t     time.Time
	bytes int64
}

func newSpeedClient() *http.Client {
	return &http.Client{Transport: http.DefaultTransport.(*http.Transport).Clone()}
}

// measureDownload returns sustained download throughput in Mbps.
func measureDownload(ctx context.Context, client *http.Client, rawURL string) (float64, error) {
	if err := warmUp(ctx, client, withBytes(rawURL, 0)); err != nil {
		return 0, fmt.Errorf("warm-up: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, withBytes(rawURL, speedDownloadBytes), nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status %s", resp.Status)
	}

	buf := make([]byte, 32*1024)
	var points []speedPoint
	var total int64
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			total += int64(n)
			points = append(points, speedPoint{t: time.Now(), bytes: total})
			if points[len(points)-1].t.Sub(points[0].t) >= speedMaxDuration {
				break
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return tailRate(points)
}

// measureUpload returns upload throughput in Mbps.
func measureUpload(ctx context.Context, client *http.Client, rawURL string) (float64, error) {
	if err := warmUp(ctx, client, rawURL); err != nil {
		return 0, fmt.Errorf("warm-up: %w", err)
	}
	body := &uploadBody{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, body)
	if err != nil {
		return 0, err
	}
	req.ContentLength = -1
	req.Header.Set("Content-Type", "application/octet-stream")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status %s", resp.Status)
	}
	dur := time.Since(start)
	if dur < speedMinDuration {
		return 0, errors.New("transfer too short to measure")
	}
	return float64(body.n) * 8 / dur.Seconds() / 1e6, nil
}

type uploadBody struct {
	n     int64
	start time.Time
}

func (b *uploadBody) Read(p []byte) (int, error) {
	if b.start.IsZero() {
		b.start = time.Now()
	}
	if b.n >= speedUploadBytes || time.Since(b.start) >= speedMaxDuration {
		return 0, io.EOF
	}
	p = p[:min(int64(len(p)), speedUploadBytes-b.n)]
	clear(p)
	b.n += int64(len(p))
	return len(p), nil
}

func tailRate(points []speedPoint) (float64, error) {
	if len(points) < 2 {
		return 0, errors.New("transfer too short to measure")
	}
	first, last := points[0], points[len(points)-1]
	cut := first.t.Add(time.Duration(float64(last.t.Sub(first.t)) * speedRampFraction))
	i := sort.Search(len(points), func(i int) bool { return !points[i].t.Before(cut) })
	from := points[min(i, len(points)-2)]
	dur := last.t.Sub(from.t)
	if dur < speedMinDuration {
		return 0, errors.New("transfer too short to measure")
	}
	return float64(last.bytes-from.bytes) * 8 / dur.Seconds() / 1e6, nil
}

func warmUp(ctx context.Context, client *http.Client, rawURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
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
