package monitor

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestWithBytes(t *testing.T) {
	tests := []struct {
		url  string
		n    int64
		want string
	}{
		{"https://speed.cloudflare.com/__down?bytes=1000000", 0, "https://speed.cloudflare.com/__down?bytes=0"},
		{"https://speed.cloudflare.com/__down?bytes=1000000", 20_000_000, "https://speed.cloudflare.com/__down?bytes=20000000"},
		{"https://example.com/100MB.bin", 0, "https://example.com/100MB.bin"},
	}
	for _, tt := range tests {
		if got := withBytes(tt.url, tt.n); got != tt.want {
			t.Errorf("withBytes(%q, %d) = %q, want %q", tt.url, tt.n, got, tt.want)
		}
	}
}

func TestTailRateSkipsRamp(t *testing.T) {
	start := time.Now()
	var points []speedPoint
	var total int64
	for ms := 10; ms <= 2500; ms += 10 {
		if ms <= 1000 {
			total += 125_000
		} else {
			total += 62_500
		}
		points = append(points, speedPoint{t: start.Add(time.Duration(ms) * time.Millisecond), bytes: total})
	}

	got, err := tailRate(points)
	if err != nil {
		t.Fatal(err)
	}
	assertNear(t, "tail Mbps", got, 50)
}

func TestTailRateTooShort(t *testing.T) {
	now := time.Now()
	if _, err := tailRate([]speedPoint{{now, 100}}); err == nil {
		t.Error("expected error for a single point")
	}
	if _, err := tailRate([]speedPoint{{now, 100}, {now.Add(time.Millisecond), 200}}); err == nil {
		t.Error("expected error for a sub-10ms transfer")
	}
}

func TestUploadBodyStopsAtByteCap(t *testing.T) {
	b := &uploadBody{}
	n, err := io.Copy(io.Discard, b)
	if err != nil {
		t.Fatal(err)
	}
	if n != speedUploadBytes || b.n != speedUploadBytes {
		t.Errorf("sent %d bytes (counted %d), want %d", n, b.n, speedUploadBytes)
	}
}

func TestMeasureDownload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.ParseInt(r.URL.Query().Get("bytes"), 10, 64)
		w.Header().Set("Content-Length", strconv.FormatInt(n, 10))
		if r.Method == http.MethodHead {
			return
		}
		for sent := int64(0); sent < n; sent += 1_000_000 {
			io.CopyN(w, zeroReader{}, min(1_000_000, n-sent))
			time.Sleep(2 * time.Millisecond)
		}
	}))
	defer srv.Close()

	mbps, err := measureDownload(context.Background(), srv.Client(), srv.URL+"/__down?bytes=1")
	if err != nil {
		t.Fatal(err)
	}
	if mbps <= 0 {
		t.Errorf("mbps = %f, want > 0", mbps)
	}
}

func TestMeasureDownloadRejectsBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer srv.Close()

	if _, err := measureDownload(context.Background(), srv.Client(), srv.URL+"/__down?bytes=1"); err == nil {
		t.Error("expected error for 403 response")
	}
}

func TestMeasureUpload(t *testing.T) {
	var received int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			received, _ = io.Copy(io.Discard, r.Body)
			time.Sleep(20 * time.Millisecond)
		}
	}))
	defer srv.Close()

	mbps, err := measureUpload(context.Background(), srv.Client(), srv.URL+"/__up")
	if err != nil {
		t.Fatal(err)
	}
	if mbps <= 0 {
		t.Errorf("mbps = %f, want > 0", mbps)
	}
	if received != speedUploadBytes {
		t.Errorf("server received %d bytes, want %d", received, speedUploadBytes)
	}
}

func TestMeasureDownloadCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := measureDownload(ctx, http.DefaultClient, "http://127.0.0.1:1/__down?bytes=1"); err == nil {
		t.Error("expected error for cancelled context")
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}
