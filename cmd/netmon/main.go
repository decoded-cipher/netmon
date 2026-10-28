package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"netmon/internal/monitor"
	"netmon/internal/server"
	"netmon/internal/store"
	"netmon/web"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z"
var version = "vX.Y.Z"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Println("netmon", version)
		os.Exit(0)
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	s, err := store.New("netmon.db")
	if err != nil {
		log.Error("database init failed", "error", err)
		os.Exit(1)
	}
	defer s.Close()

	// Priority: DB saved config > config.json > hardcoded defaults.
	cfg, found, err := monitor.LoadConfigFile("config.json")
	if err != nil {
		log.Error("config file error, using defaults", "error", err)
		cfg = monitor.DefaultConfig()
	} else if found {
		log.Info("loaded config from config.json")
	}
	if cs, err := s.GetConfig(); err == nil {
		cfg = monitor.ConfigFromStore(cs)
		log.Info("loaded config from database")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mon := monitor.New(cfg, s, log)
	mon.Start(ctx)

	h := server.NewHandler(s, mon, version)

	distFS, err := fs.Sub(web.FS, "dist")
	if err != nil {
		log.Error("failed to sub web FS", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/data", h.GetData)
	mux.HandleFunc("GET /api/config", h.GetConfig)
	mux.HandleFunc("POST /api/config", h.SaveConfig)
	mux.HandleFunc("POST /api/speedtest", h.RunSpeedTest)
	mux.HandleFunc("GET /api/version", h.GetVersion)
	mux.HandleFunc("POST /api/update", h.TriggerUpdate)
	mux.Handle("/", http.FileServer(http.FS(distFS)))

	srv := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		log.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server stopped", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown failed", "error", err)
	}
	mon.Wait()
	log.Info("shutdown complete")
}
