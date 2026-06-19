# Hermes System Health Monitor

Local Go dashboard for Hermes profile, kanban, MCP, gateway, and host health.

## What changed

The old Python API (`api.py`) and standalone root `index.html` are no longer the main path. The monitor is now a single Go service that serves both the JSON API and the embedded dashboard from one binary.

## Features

- Single binary HTTP service, default `127.0.0.1:9090`
- Embedded dark-theme dashboard served from Go `embed.FS`
- `GET /health` JSON health check
- `GET /api/status` consolidated snapshot with per-section error reporting
- Browser auto-refresh every 5 seconds
- Independent failure handling for Profiles, Kanban, MCP, Gateway, and System sections
- Command execution via `exec.CommandContext` with per-command timeouts

## Project layout

- `main.go` — HTTP server, route registration, dashboard embedding
- `dashboard/index.html` — embedded dashboard HTML/CSS/JS
- `collectors/` — Hermes/system command adapters and parsers
- `collectors/collectors_test.go` — parser and failure-mode tests
- `Makefile` — `build`, `run`, `test`, `fmt`, `clean`

## Requirements

- Go 1.23+
- `hermes` CLI in `PATH`
- Linux commands: `free`, `df`, `uptime`

## Run

```bash
make run
```

Then open:

- `http://localhost:9090/`
- `http://localhost:9090/health`
- `http://localhost:9090/api/status`

To override the port:

```bash
PORT=9191 make run
```

## Build and test

```bash
make build
make test
```

## API shape

### GET /health

Returns HTTP 200 with:

```json
{"status":"ok","timestamp":"2026-06-19T16:00:00Z"}
```

### GET /api/status

Returns one JSON object with independent section payloads:

- `profiles`
- `kanban`
- `mcp`
- `gateway`
- `system`
- `updated`

Each section has either `data` or `error`.

## Limitations

- The dashboard depends on the `hermes` CLI output formats documented in `.hermes/system-health-monitor-research.md`.
- `hermes kanban list --json` is used first; if that fails, the service falls back to text parsing.
- `hermes profile list`, `hermes mcp list`, and `hermes gateway status` do not expose JSON on this machine, so those sections still rely on table/free-text parsing.
- This is a local monitor. No auth, persistence, or historical charts.

## Removed old main path

- `api.py` was removed; the Python service is no longer part of the runtime path.
- The old root `index.html` was removed; the embedded dashboard lives at `dashboard/index.html` and is served by the Go binary.
