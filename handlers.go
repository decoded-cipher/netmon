package main

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

type DashboardData struct {
	Summary Summary       `json:"summary"`
	Targets []PingTarget  `json:"targets"`
	DNS     []DNSCheck    `json:"dns"`
	History []Measurement `json:"history"`
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
		Summary: summary,
		Targets: targets,
		DNS:     dns,
		History: history,
	})
}
