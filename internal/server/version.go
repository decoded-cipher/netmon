package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type versionChecker struct {
	current string
	mu      sync.RWMutex
	latest  string
}

func newVersionChecker(current string) *versionChecker {
	vc := &versionChecker{current: current}
	go vc.fetch()
	return vc
}

func (vc *versionChecker) fetch() {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/decoded-cipher/netmon/releases/latest")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return
	}
	vc.mu.Lock()
	vc.latest = payload.TagName
	vc.mu.Unlock()
}

type VersionResponse struct {
	Current         string `json:"current"`
	Latest          string `json:"latest"`
	UpdateAvailable bool   `json:"update_available"`
}

func (vc *versionChecker) response() VersionResponse {
	vc.mu.RLock()
	latest := vc.latest
	vc.mu.RUnlock()
	return VersionResponse{
		Current:         vc.current,
		Latest:          latest,
		UpdateAvailable: latest != "" && latest != vc.current,
	}
}

func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.version.response())
}
