package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"netmon/internal/monitor"
	"netmon/internal/store"
)

type Handler struct {
	store *store.Store
	mon   *monitor.Monitor
}

func NewHandler(s *store.Store, mon *monitor.Monitor) *Handler {
	return &Handler{store: s, mon: mon}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type DashboardData struct {
	Summary   store.Summary       `json:"summary"`
	Targets   []store.PingTarget  `json:"targets"`
	DNS       []store.DNSCheck    `json:"dns"`
	History   []store.Measurement `json:"history"`
	NetworkID string              `json:"network_id"`
}

func (h *Handler) GetData(w http.ResponseWriter, r *http.Request) {
	minutes := 60
	if m := r.URL.Query().Get("minutes"); m != "" {
		if n, err := strconv.Atoi(m); err == nil && n > 0 && n <= 43200 {
			minutes = n
		}
	}

	summary, err := h.store.GetSummary()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	targets, err := h.store.GetPingTargets()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	dns, err := h.store.GetDNSChecks()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	history, err := h.store.GetHistoryWindow(minutes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, DashboardData{
		Summary:   summary,
		Targets:   targets,
		DNS:       dns,
		History:   history,
		NetworkID: h.store.GetCurrentNetworkID(),
	})
}

// ConfigPayload is the API shape for GET /api/config and POST /api/config.
type ConfigPayload struct {
	PingTargets    []string `json:"ping_targets"`
	DNSTargets     []string `json:"dns_targets"`
	PingIntervalS  int      `json:"ping_interval_s"`
	SpeedIntervalM int      `json:"speed_interval_m"`
	PingCount      int      `json:"ping_count"`
}

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := h.mon.GetConfig()
	writeJSON(w, http.StatusOK, ConfigPayload{
		PingTargets:    cfg.PingTargets,
		DNSTargets:     cfg.DNSTargets,
		PingIntervalS:  int(cfg.PingInterval.Seconds()),
		SpeedIntervalM: int(cfg.SpeedInterval.Minutes()),
		PingCount:      cfg.PingCount,
	})
}

func (h *Handler) SaveConfig(w http.ResponseWriter, r *http.Request) {
	var p ConfigPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if len(p.PingTargets) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ping_targets cannot be empty"})
		return
	}
	if p.PingIntervalS < 10 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ping_interval_s must be >= 10"})
		return
	}
	if p.SpeedIntervalM < 5 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "speed_interval_m must be >= 5"})
		return
	}
	if p.PingCount < 1 || p.PingCount > 20 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ping_count must be between 1 and 20"})
		return
	}

	if err := h.store.SaveConfig(store.ConfigSettings{
		PingTargets:    p.PingTargets,
		DNSTargets:     p.DNSTargets,
		PingIntervalS:  p.PingIntervalS,
		SpeedIntervalM: p.SpeedIntervalM,
		PingCount:      p.PingCount,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	def := monitor.DefaultConfig()
	h.mon.SetConfig(monitor.Config{
		PingTargets:   p.PingTargets,
		DNSTargets:    p.DNSTargets,
		PingInterval:  time.Duration(p.PingIntervalS) * time.Second,
		SpeedInterval: time.Duration(p.SpeedIntervalM) * time.Minute,
		PingCount:     p.PingCount,
		DownloadURL:   def.DownloadURL,
		UploadURL:     def.UploadURL,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}
