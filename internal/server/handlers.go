package server

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"netmon/internal/store"
)

type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

type DashboardData struct {
	Summary   store.Summary      `json:"summary"`
	Targets   []store.PingTarget `json:"targets"`
	DNS       []store.DNSCheck   `json:"dns"`
	History   []store.Measurement `json:"history"`
	NetworkID string              `json:"network_id"`
}

func (h *Handler) GetData(c *echo.Context) error {
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

	history, err := h.store.GetHistory(30)
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
