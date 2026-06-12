# Hermes System Health Monitor

Build a real-time system monitoring dashboard at `/home/jorge/Documents/projects/system-health-monitor/`.

## Architecture

**Backend:** Python thin API server (`api.py`) that runs `hermes` CLI commands via subprocess and exposes JSON endpoints. Serves on port 9092.

**Frontend:** Single HTML file (`index.html`) with dark/light cards, shadow-as-border technique, font-feature-settings ligatures. Polls the API every 30 seconds. Served on port 9091.

## What to Build

### 1. Backend API (`api.py`)
- `/api/health` — returns `{"status": "ok"}`
- `/api/profiles` — runs `hermes profile list`, parses output, returns profile names and models
- `/api/mcp` — runs `hermes mcp list`, returns server names and status
- `/api/kanban` — runs `hermes kanban list`, returns task counts (done/running/todo/blocked) and task list
- `/api/cron` — runs `hermes cron list`, returns job list
- `/api/sessions` — runs `hermes sessions list`, returns recent sessions
- All endpoints include CORS headers (`Access-Control-Allow-Origin: *`)
- Parse Hermes CLI table output using regex patterns for the Unicode box-drawing format

### 2. Frontend Dashboard (`index.html`)
- Clean card grid layout (2 columns, full-width cards for wide sections)
- Cards: Profiles, MCP Servers, Kanban Board (with status strip), Cron Jobs, Recent Sessions
- Status dots (green/red/amber) for MCP servers
- Kanban status strip: Done (green), Running (blue), Todo (gray), Blocked (amber) counts
- Auto-refresh every 30 seconds
- Gateway status indicator in footer
- Use `#fafafa` background, `#171717` text, shadow-as-border cards

### 3. Deploy
- Serve with `python3 -m http.server 9091` for frontend
- Serve API with `python3 api.py 9092`

## Already Done
- Frontend HTML exists at `index.html` — polish and fix data binding
- Backend `api.py` exists — polish parsing logic
- API tested: returns 4 profiles, 14 kanban tasks, 4+ MCP servers

## Acceptance Criteria
- [ ] Dashboard loads and shows live data from API
- [ ] All 6 API endpoints return valid JSON with correct CORS headers
- [ ] MCP server list excludes table headers
- [ ] Kanban counts reflect real board state
- [ ] Profiles show all configured profiles (default, builder, reviewer, researcher)
- [ ] Responsive grid layout
- [ ] git init, commit, push with clear commit message
