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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mon := monitor.New(monitor.DefaultConfig(), s, log)
	mon.Start(ctx)

	e := echo.New()
	e.Use(middleware.Recover())

	h := server.NewHandler(s)
	e.GET("/api/data", h.GetData)
	e.GET("/", func(c *echo.Context) error {
		data, err := web.FS.ReadFile("index.html")
		if err != nil {
			return err
		}
		return c.HTMLBlob(http.StatusOK, data)
	})

	log.Info("starting server", "addr", ":8080")
	if err := e.Start(":8080"); err != nil {
		log.Error("server stopped", "error", err)
	}

	cancel()
	mon.Wait()
}
