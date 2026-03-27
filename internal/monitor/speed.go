package monitor

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

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
	data := make([]byte, 1*1024*1024) // 1 MB
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
