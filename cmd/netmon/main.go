package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

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

	h := server.NewHandler(s, mon)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/data", h.GetData)
	mux.HandleFunc("GET /api/config", h.GetConfig)
	mux.HandleFunc("POST /api/config", h.SaveConfig)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		data, err := web.FS.ReadFile("index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})
	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		data, err := web.FS.ReadFile("style.css")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/css")
		w.Write(data)
	})
	mux.HandleFunc("GET /script.js", func(w http.ResponseWriter, r *http.Request) {
		data, err := web.FS.ReadFile("script.js")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(data)
	})

	log.Info("starting server", "addr", ":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Error("server stopped", "error", err)
	}

	cancel()
	mon.Wait()
}
