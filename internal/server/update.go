package server

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const maxBinarySize = 100 << 20 // 100 MB

func (h *Handler) TriggerUpdate(w http.ResponseWriter, r *http.Request) {
	vr := h.version.response()
	if !vr.UpdateAvailable {
		writeJSON(w, http.StatusOK, map[string]string{"status": "up_to_date"})
		return
	}

	if runtime.GOOS == "windows" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "self-update is not supported on Windows; download the latest release manually"})
		return
	}

	execPath, err := os.Executable()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "cannot determine executable path: " + err.Error()})
		return
	}

	tag    := vr.Latest
	goos   := runtime.GOOS
	goarch := runtime.GOARCH
	archiveURL := fmt.Sprintf(
		"https://github.com/decoded-cipher/netmon/releases/download/%s/netmon_%s_%s.tar.gz",
		tag, goos, goarch,
	)

	// Respond before restarting so the client gets the response.
	writeJSON(w, http.StatusOK, map[string]string{"status": "updating", "version": tag})
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	go func() {
		time.Sleep(300 * time.Millisecond)
		if err := downloadAndReplace(archiveURL, execPath); err != nil {
			// Cannot write to the client at this point; the update silently failed.
			// The polling loop on the frontend will time out and show an error.
			return
		}
		restart(execPath)
	}()
}

func downloadAndReplace(url, execPath string) error {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}

	dir := filepath.Dir(execPath)
	tmp, err := os.CreateTemp(dir, ".netmon-update-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	success := false
	defer func() {
		tmp.Close()
		if !success {
			os.Remove(tmpPath)
		}
	}()

	gz, err := gzip.NewReader(io.LimitReader(resp.Body, maxBinarySize))
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	found := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}
		if filepath.Base(hdr.Name) == "netmon" && hdr.Typeflag == tar.TypeReg {
			if _, err := io.Copy(tmp, io.LimitReader(tr, maxBinarySize)); err != nil {
				return fmt.Errorf("extract: %w", err)
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("netmon binary not found in release archive")
	}

	if err := tmp.Chmod(0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}
	tmp.Close()

	if err := os.Rename(tmpPath, execPath); err != nil {
		return fmt.Errorf("replace binary (check write permissions): %w", err)
	}
	success = true
	return nil
}
