# Contributing to netmon

## Prerequisites

- **Go 1.21+** — [install](https://go.dev/dl/)
- **Node.js 18+** — [install](https://nodejs.org/) (for building the Vue frontend)

No C compiler needed — the project uses `modernc.org/sqlite`, a pure-Go SQLite driver (`CGO_ENABLED=0`).

## Run locally

```bash
git clone https://github.com/decoded-cipher/netmon.git
cd netmon
make run          # go run ./cmd/netmon
# or
make build && ./netmon
```

Dashboard: **http://localhost:8080**

## Project structure

```
cmd/netmon/        main.go — entry point, HTTP server setup
internal/
  monitor/         Background workers: ping (60s) and speed test (30m)
    monitor.go     Monitor struct, worker goroutines, ping cycle orchestration
    ping.go        runPing, cross-platform output parsing, DNS measurement
    speed.go       measureDownload, measureUpload
  network/         Cross-platform network identity detection
    network.go     Detect(), DefaultGateway(), SSID helpers per OS
  server/          HTTP layer
    handlers.go    JSON API handlers
    version.go     Version check against GitHub releases
  store/           SQLite persistence
    store.go       Schema, migrations, CRUD for all tables
web/               Vue 3 + Vite frontend
  src/             Components, composables, and utilities
  web.go           embed.FS — bundles web/dist into the binary at build time
```

## Making changes

- **Backend changes** — edit files under `internal/`. Run `make vet` and `make build` to verify.
- **Frontend changes** — edit files under `web/src/`. Run `make dev` to get hot-reload (Vite proxies `/api/*` to the Go backend).
- **Adding a new metric** — add a column to the `measurements` table in `store.go`, update `Measurement` struct, `SaveMeasurement`, and `GetHistory`; wire it through `monitor.go → runPingCycle()`; add a chart in the Vue components.

## Pull requests

1. Fork the repo and create a branch from `master`
2. Keep changes focused — one logical change per PR
3. Run `make vet` before pushing
4. Describe *why* the change is needed in the PR description, not just what changed

## Reporting issues

Open an issue with:
- OS and Go version (`go version`)
- Steps to reproduce
- Relevant log output (the structured log lines from the terminal)
