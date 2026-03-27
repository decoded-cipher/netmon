# netmon

A lightweight, self-hosted network monitoring dashboard. Tracks latency, jitter, packet loss, DNS resolution times, and bandwidth — displayed in a live web UI with no external services required.

![Build](https://github.com/decoded-cipher/netmon/actions/workflows/ci.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)


## Features

- **Live dashboard** — auto-refreshing charts for latency, throughput, packet loss, jitter, and DNS
- **Network change detection** — automatically detects Wi-Fi/network switches and tags measurements per network; works on macOS, Linux, Windows, Docker, and Raspberry Pi
- **Self-contained binary** — single executable with embedded UI, no config files needed
- **Lightweight** — pings every 60s, speed test every 30min (1 MB); negligible network overhead
- **SQLite storage** — no database server; data persists in a single file
- **Cross-platform** — runs on Linux (x86, ARM/Pi), macOS, Windows, and Docker


## Quick Start

### Docker (recommended)

```bash
docker run -d \
  --name netmon \
  --network host \
  -v netmon-data:/data \
  ghcr.io/decoded-cipher/netmon:latest

# create the volume first if it doesn't exist
docker volume create netmon-data
```

Or with Docker Compose:

```bash
docker compose up -d
```

Then open **http://localhost:8080**.

### Build from source

Requires Go 1.21+ and a C compiler (`gcc`) for the SQLite driver.

```bash
git clone https://github.com/decoded-cipher/netmon.git
cd netmon
make build
./netmon
```

### Makefile targets

```
make build      # compile binary → ./netmon
make run        # go run ./cmd/netmon (no build step)
make clean      # remove binary and local database
make vet        # run go vet on all packages
make docker     # build Docker image
make docker-run # build + run Docker container
```


## Configuration

All defaults live in `internal/monitor/monitor.go → DefaultConfig()`. There is no config file — rebuild or set env vars (support planned).

| Setting | Default | Description |
|---------|---------|-------------|
| Ping targets | `google.com`, `cloudflare.com` | Hosts to measure latency/loss against |
| DNS targets | `google.com`, `cloudflare.com` | Hosts to measure DNS resolution time |
| Ping interval | `60s` | How often to run a ping cycle |
| Speed test interval | `30m` | How often to test download/upload |
| Speed test size | `1 MB` | Payload size (kept small to avoid hogging the link) |
| HTTP port | `:8080` | Dashboard address |


## Architecture

```
cmd/netmon/          Entry point — wires packages, starts HTTP server
internal/
  monitor/           Ping + speed workers (concurrent, 60s / 30m intervals)
  network/           Cross-platform gateway detection and SSID identification
  server/            HTTP handler, JSON API response types
  store/             SQLite layer — schema, UPSERT patterns, aggregation queries
web/
  index.html         Single-page dashboard (TailwindCSS + ApexCharts, embedded at build time)
```

**Data flow:**
1. `pingWorker` pings all targets concurrently, resolves DNS, detects network — saves a `Measurement` row every 60s
2. `speedWorker` downloads/uploads 1 MB to Cloudflare every 30 min — updates latest bandwidth values
3. Dashboard polls `GET /api/data` every 30s and renders charts client-side


## Platforms

| Platform | Tested | Notes |
|----------|--------|-------|
| macOS (Apple Silicon / Intel) | Yes | Native binary |
| Linux x86-64 | Yes | Includes Docker |
| Linux ARM (Raspberry Pi) | Yes | Build with `GOARCH=arm64` |
| Windows | Partial | Ping parsing works; SSID detection via `netsh` |
| Docker | Yes | See `Dockerfile` and `docker-compose.yml` |

### Raspberry Pi

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc \
  go build -o netmon ./cmd/netmon
```


## Requirements

- **Go 1.21+** for building from source
- **CGO enabled** (`CGO_ENABLED=1`) — required by the SQLite driver (`go-sqlite3`)
- A C compiler (`gcc` / `musl-gcc` / `clang`) at build time; not needed at runtime
- `ping` available in `PATH` at runtime (standard on all platforms)
