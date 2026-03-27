package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	store, err := NewStore("netmon.db")
	if err != nil {
		log.Error("database init failed", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mon := NewMonitor(DefaultConfig(), store, log)
	mon.Start(ctx)

	e := echo.New()
	e.Use(middleware.Recover())

	h := NewHandler(store)
	e.GET("/api/data", h.GetData)
	e.File("/", "index.html")

	log.Info("starting server", "addr", ":8080")
	if err := e.Start(":8080"); err != nil {
		log.Error("server stopped", "error", err)
	}

	cancel()
	mon.Wait()
}
