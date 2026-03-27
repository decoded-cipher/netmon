package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"netmon/internal/monitor"
	"netmon/internal/server"
	"netmon/internal/store"
	"netmon/web"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	s, err := store.New("netmon.db")
	if err != nil {
		log.Error("database init failed", "error", err)
		os.Exit(1)
	}
	defer s.Close()

	// Priority: DB saved config > config.json > hardcoded defaults.
	cfg, err := monitor.LoadConfigFile("config.json")
	if err != nil {
		log.Error("config file error, using defaults", "error", err)
		cfg = monitor.DefaultConfig()
	} else {
		log.Info("loaded config from config.json")
	}
	if cs, err := s.GetConfig(); err == nil {
		cfg = monitor.ConfigFromStore(cs)
		log.Info("loaded config from database")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mon := monitor.New(cfg, s, log)
	mon.Start(ctx)

	e := echo.New()
	e.Use(middleware.Recover())

	h := server.NewHandler(s, mon)
	e.GET("/api/data", h.GetData)
	e.GET("/api/config", h.GetConfig)
	e.POST("/api/config", h.SaveConfig)
	e.GET("/", func(c *echo.Context) error {
		data, err := web.FS.ReadFile("index.html")
		if err != nil {
			return err
		}
		return c.HTMLBlob(http.StatusOK, data)
	})
	e.GET("/style.css", func(c *echo.Context) error {
		data, err := web.FS.ReadFile("style.css")
		if err != nil {
			return err
		}
		return c.Blob(http.StatusOK, "text/css", data)
	})
	e.GET("/script.js", func(c *echo.Context) error {
		data, err := web.FS.ReadFile("script.js")
		if err != nil {
			return err
		}
		return c.Blob(http.StatusOK, "application/javascript", data)
	})

	log.Info("starting server", "addr", ":8080")
	if err := e.Start(":8080"); err != nil {
		log.Error("server stopped", "error", err)
	}

	cancel()
	mon.Wait()
}
