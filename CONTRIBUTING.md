# Contributing to netmon

## Prerequisites

- **Go 1.21+** — [install](https://go.dev/dl/)
- **C compiler** — required by the SQLite driver (`go-sqlite3` uses CGO)
  - macOS: `xcode-select --install`
  - Linux: `sudo apt install gcc` / `sudo dnf install gcc`
  - Windows: [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or WSL

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
    handlers.go    GET /api/data handler, DashboardData response type
  store/           SQLite persistence
    store.go       Schema, migrations, CRUD for all 4 tables
web/
  index.html       Single-page dashboard (TailwindCSS + ApexCharts)
  web.go           embed.FS — bundles index.html into the binary
```

## Making changes

- **Backend changes** — edit files under `internal/`. Run `make vet` and `make build` to verify.
- **Dashboard changes** — edit `web/index.html`. Rebuild and refresh the browser.
- **Adding a new metric** — add a column to the `measurements` table in `store.go → migrate()`, update `Measurement` struct, `SaveMeasurement`, and `GetHistory`; wire it through `monitor.go → runPingCycle()`; render it in `index.html`.

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
