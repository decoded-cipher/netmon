package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"

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

type DashboardData struct {
	Summary   store.Summary       `json:"summary"`
	Targets   []store.PingTarget  `json:"targets"`
	DNS       []store.DNSCheck    `json:"dns"`
	History   []store.Measurement `json:"history"`
	NetworkID string              `json:"network_id"`
}

func (h *Handler) GetData(c *echo.Context) error {
	minutes := 60
	if m := c.QueryParam("minutes"); m != "" {
		if n, err := strconv.Atoi(m); err == nil && n > 0 && n <= 43200 {
			minutes = n
		}
	}

	summary, err := h.store.GetSummary()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	targets, err := h.store.GetPingTargets()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	dns, err := h.store.GetDNSChecks()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	history, err := h.store.GetHistoryWindow(minutes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, DashboardData{
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

func (h *Handler) GetConfig(c *echo.Context) error {
	cfg := h.mon.GetConfig()
	return c.JSON(http.StatusOK, ConfigPayload{
		PingTargets:    cfg.PingTargets,
		DNSTargets:     cfg.DNSTargets,
		PingIntervalS:  int(cfg.PingInterval.Seconds()),
		SpeedIntervalM: int(cfg.SpeedInterval.Minutes()),
		PingCount:      cfg.PingCount,
	})
}

func (h *Handler) SaveConfig(c *echo.Context) error {
	var p ConfigPayload
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if len(p.PingTargets) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ping_targets cannot be empty"})
	}
	if p.PingIntervalS < 10 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ping_interval_s must be >= 10"})
	}
	if p.SpeedIntervalM < 5 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "speed_interval_m must be >= 5"})
	}
	if p.PingCount < 1 || p.PingCount > 20 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ping_count must be between 1 and 20"})
	}

	if err := h.store.SaveConfig(store.ConfigSettings{
		PingTargets:    p.PingTargets,
		DNSTargets:     p.DNSTargets,
		PingIntervalS:  p.PingIntervalS,
		SpeedIntervalM: p.SpeedIntervalM,
		PingCount:      p.PingCount,
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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

	return c.JSON(http.StatusOK, map[string]string{"status": "saved"})
}
