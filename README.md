# Hermes System Health Monitor

Real-time dashboard for monitoring Hermes Agent system health.

## Architecture

- **api.py** — Python HTTP API (port 9092) that wraps `hermes` CLI commands into JSON endpoints
- **index.html** — Single-page dashboard (port 9091) with auto-refreshing data

## Endpoints

| Endpoint | Description |
|---|---|
| `/api/health` | Gateway health check |
| `/api/profiles` | Active Hermes profiles |
| `/api/mcp` | MCP server status (with tools count) |
| `/api/kanban` | Kanban board task counts and recent tasks |
| `/api/cron` | Scheduled cron jobs with schedule and last run |
| `/api/sessions` | Recent session history |

## Running

```bash
# Start API
python3 api.py 9092 &

# Serve frontend
python3 -m http.server 9091 --bind 127.0.0.1 &

# Open http://localhost:9091
```

## Features

- 6 API endpoints with CORS headers and Content-Length
- Responsive 3-column grid (900px → 2-col, 600px → 1-col)
- Status strip with Done/Running/Todo/Blocked counts
- Profile count and MCP enabled/total badges
- 30-second auto-refresh
- Gateway online/offline indicator
