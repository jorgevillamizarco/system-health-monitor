# System Health Monitor — Go Rewrite Research Handoff

> Conducted 2026-06-19 on `asus` (Linux 7.0.0-22-generic, 30Gi RAM, AMD Strix Halo)
> Researcher profile: deepseek-v4-pro
> Hermes version: current (has kanban --json)

## 1. Data Sources Table

| Section   | Command                   | Exact Fields Needed                          | Parse Strategy          | JSON? | Failure Mode                    | Timeout |
|-----------|---------------------------|----------------------------------------------|-------------------------|-------|---------------------------------|---------|
| Profiles  | `hermes profile list`     | name, model, gateway_status, active          | Regex on fixed columns  | NO    | Empty table, "unavailable" row  | 10s     |
| Kanban    | `hermes kanban list --json` | id, title, assignee, status, priority       | `json.Unmarshal`        | YES   | Empty table, "unavailable" row  | 15s     |
| MCP       | `hermes mcp list`         | name, transport, tools, status               | Regex table parser       | NO    | Empty table, "unavailable" row  | 10s     |
| Gateway   | `hermes gateway status`   | running (bool), pid, profiles                | Regex + line scan        | NO    | Red dot, "offline" text         | 5s      |
| System    | `free -h`                 | mem_total, mem_used, mem_avail, swap_used    | Regex Gi/Mi extraction   | N/A   | "--" for all values             | 5s      |
| System    | `df -h /`                 | disk_size, disk_used, disk_avail, disk_use%  | Regex on /dev/nvme* line | N/A   | "--" for all values             | 5s      |
| System    | `uptime`                  | uptime_str, users, load_1, load_5, load_15   | Regex + strconv          | N/A   | "--" for uptime, "0" for load   | 5s      |

**JSON availability verified 2026-06-19:**
- `hermes kanban list --json` — **WORKS**. Returns full JSON array of task objects with id, title, body, assignee, status, priority, workspace_*, timestamps, etc.
- `hermes profile list --json` — **DOES NOT EXIST**. `--help` shows only `-h`.
- `hermes mcp list --json` — **DOES NOT EXIST**. `--help` shows only `-h`.
- `hermes gateway status --json` — **DOES NOT EXIST**. Has `--deep` and `--full` but no structured output flag.

## 2. Exact Command Output Examples

### 2.1 hermes profile list
```
Profile          Model                        Gateway      Alias        Distribution
 ───────────────    ───────────────────────────    ───────────    ───────────    ────────────────────
  default         gpt-5.5                      running      —            —
  builder         gpt-5.4                      stopped      —            —
 ◆researcher      deepseek-v4-pro              stopped      —            —
  reviewer        deepseek-v4-pro              stopped      —            —
```
Exit code: 0. Column widths vary. Active profile marked with `◆` prefix.
Parser must handle: variable column alignment, em-dash `—` for empty fields, `◆` prefix on active profile.

### 2.2 hermes kanban list (text, no --json)
```
● t_d77413bd  running   researcher            research: system-health-monitor Go rewrite inputs
◻ t_3b5f0b85  todo      builder               build: rewrite system-health-monitor as Go embedded dashboard
◻ t_e07405c8  todo      reviewer              review: verify system-health-monitor Go dashboard E2E
✓ t_5cfc6d66  done      researcher            research: Ebitengine WASM best practices for Chrome 149+
⊘ t_1ea671b0  blocked   reviewer              review: System Health Monitor E2E
```
Status icons: ●=running, ◻=todo, ✓=done, ⊘=blocked. Exit code: 0.
**USE `--json` INSTEAD.** Text parsing is fragile; the JSON output (see below) is complete.

### 2.3 hermes kanban list --json (preferred path)
```json
[
  {
    "id": "t_d77413bd",
    "title": "research: system-health-monitor Go rewrite inputs",
    "body": "...",
    "assignee": "researcher",
    "status": "running",
    "priority": 100,
    "tenant": null,
    "workspace_kind": "dir",
    "workspace_path": "/home/jorge/Documents/projects/system-health-monitor",
    "created_by": "user",
    "created_at": 1781877226,
    "started_at": 1781877234,
    "completed_at": null,
    "result": null,
    "skills": [],
    "max_retries": null,
    "session_id": null
  }
]
```
All fields available. Key fields for dashboard: id, title, assignee, status, priority.
Body can be large (multi-KB) — omit from dashboard display, only use id/title/assignee/status.
Derive status counts by aggregating the `status` field across all items.
Filter archived tasks: `--archived` flag available but not needed for dashboard.

### 2.4 hermes mcp list
```
MCP Servers:

  Name             Transport                      Tools        Status    
  ──────────────── ────────────────────────────── ──────────── ──────────
  discovery        http://localhost:8001/mcp      all          ✓ enabled
  research         http://localhost:8100/mcp      all          ✓ enabled
  chrome-devtools  /home/jorge/local/node/bi...   all          ✓ enabled
```
Exit code: 0. Transport path truncated with `...` if too long.
Parser: skip "MCP Servers:" header, skip separator lines (─), parse columns by position.
Status field: "✓ enabled" or error text. Tools is always "all" for current config.

### 2.5 hermes gateway status
```
✓ Gateway is running (PID: 79780)
  (Running manually, not as a system service)

To install as a service:
  hermes gateway install
  sudo hermes gateway install --system

Other profiles:
  ✓ default          — PID 79780
```
Exit code: 0 (when running). If not running, different output (likely exit code 1).
Parser: check for "running" substring OR "✓ Gateway is running". Extract PID from regex `PID: (\d+)`.
`--deep` flag exists but produces verbose output — not needed for dashboard.

### 2.6 free -h
```
               total        used        free      shared  buff/cache   available
Mem:            30Gi       7.3Gi       976Mi        19Mi        22Gi        23Gi
Swap:          4.0Gi       1.6Gi       2.4Gi
```
Exit code: 0. Parse Mem line: extract numeric + unit for total, used, available.
Swap line: extract used, total.

### 2.7 df -h /
```
Filesystem      Size  Used Avail Use% Mounted on
/dev/nvme0n1p4  466G  328G  115G  75% /
```
Need to filter: run `df -h /` to get only root filesystem, or parse and filter to mountpoint `/`.
Exit code: 0.

### 2.8 uptime
```
15:54:21 up  4:03,  1 user,  load average: 0.51, 0.45, 0.43
```
Exit code: 0. Parser: extract uptime duration ("4:03" or "1 day, 2:03"), user count, three load averages.

## 3. Recommended Dashboard Layout

Based on TUI monitor patterns (btop, htop, glances) and web dashboard conventions:

### 3.1 Overall Structure
```
┌─────────────────────────────────────────────────────┐
│  ● Hermes Monitor    uptime: 4h 03m  load: 0.51    │  ← Top strip
│  gateway: ● online   updated: 15:54:21              │
├──────────────────┬──────────────────────────────────┤
│  Profiles (4)    │  Kanban Board                    │  ← Row 1
│  ┌──────────────┐│  ┌──────┬──────┬──────┬───────┐│
│  │◆ researcher  ││  │ 12 ✓ │  1 ● │  8 ◻ │  1 ⊘  ││  Status strip
│  │  deepseek-v4 ││  │ Done  │ Run   │ Todo  │ Blk  ││
│  │  stopped     ││  └──────┴──────┴──────┴───────┘│
│  │ default •run ││  Latest tasks table (5 rows)     │
│  │  gpt-5.5     ││                                  │
│  └──────────────┘│                                  │
├──────────────────┼──────────────────────────────────┤
│  MCP Servers (3) │  System                          │  ← Row 2
│  ┌──────────────┐│  ┌────────────────────────────┐ │
│  │ discovery ✓  ││  │ Memory:  7.3/30 Gi  (24%)  │ │
│  │ research  ✓  ││  │ ████████░░░░░░░░░░░░░░░░░  │ │
│  │ chrome    ✓  ││  │ Disk:  328/466 G    (75%)  │ │
│  └──────────────┘│  │ ████████████████░░░░░░░░░░  │ │
│                  │  │ Swap:  1.6/4.0 Gi           │ │
│                  │  └────────────────────────────┘ │
└──────────────────┴──────────────────────────────────┘
```

### 3.2 Card Layout Rules
- **Dark theme** — background `#0d1117` (GitHub dark), cards `#161b22`, borders `#30363d`
- **2x2 grid** — 2 columns on desktop, 1 column on narrow (<768px)
- **Card padding**: 16px, border-radius: 8px
- **Status dot colors**: green (#10b981), amber (#f59e0b), red (#ef4444)
- **Table font**: 13px, monospace for IDs and paths
- **Badge colors**: done=green-bg, running=blue-bg, blocked=amber-bg, todo=gray-bg
- **Top strip**: single-row bar with gateway dot, uptime, load avg, last-updated timestamp

### 3.3 Section: Profiles
- Table: Profile | Model | Gateway
- Active profile row highlighted with left border accent
- Gateway: green dot + "running" / red dot + "stopped"
- Count badge in card header: "Profiles (4)"

### 3.4 Section: Kanban
- Status strip: 4 metric cards (Done, Running, Todo, Blocked) with large numbers
- Below strip: table of 5 most recent non-done tasks (id, title, assignee, status badge)
- Derived from JSON aggregation — no text parsing needed

### 3.5 Section: MCP
- Table: Server | Transport | Status
- Transport: show full URL (no truncation — web has space)
- Status: green dot + "enabled" / red dot + "error"
- Count badge: "MCP Servers (3/3)" — enabled/total

### 3.6 Section: System
- Memory: label + used/total + percentage + CSS progress bar
- Disk (root): label + used/total + percentage + CSS progress bar
- Swap: label + used/total (only show if swap > 0)
- Bar colors: green < 60%, amber 60-85%, red > 85%

### 3.7 Error/Empty States
- Each card independently handles failure — one broken command doesn't blank the page
- Error state: show "⚠ unavailable" with muted styling, keep card structure
- Loading state: subtle pulse animation on placeholder content
- Empty state (e.g., no MCP servers): "No servers configured" with muted text

### 3.8 Auto-Refresh
- Poll `/api/status` every 5 seconds from browser JS
- Show "Updated: HH:MM:SS" in top strip
- If fetch fails 3 consecutive times, show "⚠ connection lost" and stop polling (avoid retry storms)
- Re-poll on user click or after 30s pause

## 4. Implementation Constraints for Builder

### 4.1 Command Execution
- Use `os/exec` with `CommandContext(ctx, ...)` — NEVER bare `Command`
- Every command gets a context with timeout matching the table above
- Do NOT shell out through `sh -c`; use direct args: `exec.CommandContext(ctx, "hermes", "kanban", "list", "--json")`
- Strip/sanitize stderr from error messages — don't leak full command output to dashboard

### 4.2 Parser Paths — Safest to Riskiest

**TIER 1 — JSON native (use these):**
1. `hermes kanban list --json` → `json.Unmarshal` into `[]KanbanTask` struct. ZERO parsing risk.
   **This is the single biggest improvement over the current Python implementation.**

**TIER 2 — Fixed regex with known stable output:**
2. `free -h` — regex: `Mem:\s+([\d.]+)([GM]i)\s+([\d.]+)([GM]i)\s+.*\s+([\d.]+)([GM]i)` for total/used/available
3. `df -h /` — regex: filter line containing mountpoint `/`, extract Size/Used/Avail/Use%
4. `uptime` — regex: `up\s+(.+?),\s+(\d+) user.*load average:\s+([\d.]+),\s+([\d.]+),\s+([\d.]+)`

**TIER 3 — Table column parsing (fragile, but stable for known output):**
5. `hermes profile list` — split lines, skip separator (─), parse columns by position. 
   Active profile: check for `◆` prefix. Columns: Profile (0), Model (1), Gateway (2), Alias (3), Distribution (4)
6. `hermes mcp list` — skip header lines, parse table rows. Columns: Name (0), Transport (1-?), Tools (?-?), Status (last)

**TIER 4 — Free-text scanning (most fragile):**
7. `hermes gateway status` — scan for "running" keyword, extract PID via regex.

### 4.3 Graceful Degradation
```go
type SectionResult struct {
    Data  interface{} `json:"data"`
    Error string      `json:"error,omitempty"`
}

type StatusResponse struct {
    Profiles SectionResult `json:"profiles"`
    Kanban   SectionResult `json:"kanban"`
    MCP      SectionResult `json:"mcp"`
    Gateway  SectionResult `json:"gateway"`
    System   SectionResult `json:"system"`
    Updated  string        `json:"updated"`
}
```
- Each section is populated independently
- If a command fails: set `Error` field, leave `Data` as nil
- Frontend renders "⚠ error" card for failed sections
- /health endpoint always returns 200 with `{"status":"ok","timestamp":"..."}` even if some sections fail

### 4.4 Go Module Structure
```
system-health-monitor/
├── main.go              # HTTP server, route registration, embed.FS
├── dashboard/
│   └── index.html       # Embedded dashboard (single file, no build step)
├── collectors/
│   ├── profiles.go      # hermes profile list parser
│   ├── kanban.go        # hermes kanban list --json parser
│   ├── mcp.go           # hermes mcp list parser
│   ├── gateway.go       # hermes gateway status parser
│   └── system.go        # free, df, uptime parsers
├── collectors_test.go   # Tests for each parser
├── go.mod
├── go.sum
├── Makefile
├── .gitignore
└── README.md
```

### 4.5 Timeout Strategy
- Single global timeout context (10s) wrapping all collectors
- Each collector runs in a goroutine, results collected via channel
- If a collector exceeds global timeout, its result is the error state
- Frontend fetch timeout: 8s (slightly less than server timeout)

### 4.6 Embedded Dashboard
- Use Go 1.16+ `embed` package: `//go:embed dashboard/index.html`
- Single HTML file — no React build step, no npm
- Inline CSS and JS (keep it under ~300 lines total)
- Dark theme CSS variables
- Fetch `/api/status` every 5s, render all sections

## 5. Risks and Assumptions

### Assumptions
1. **Hermes CLI is in $PATH** — Go binary assumes `hermes` is available. No hardcoded paths.
2. **kanban --json format is stable** — the JSON structure won't change between Hermes versions. If it does, `json.Unmarshal` will fail gracefully (unmarshal error → error state).
3. **Dashboard is local only** — no auth, no rate limiting, bind to 127.0.0.1 only.
4. **Single file dashboard** — sufficient for this scope. No React/Vite build complexity.
5. **Port 9090** — configurable via PORT env var. Default 9090.

### Risks (with mitigations)
| Risk | Impact | Mitigation |
|------|--------|------------|
| Hermes CLI not in PATH for Go process | All Hermes sections fail | Check at startup, log warning. Graceful error in dashboard. |
| `hermes kanban list --json` deprecated/changed | Kanban section broken | Fallback: attempt text parse if JSON unmarshal fails. This is why collector should try JSON first, then text. |
| `hermes profile list` column format changes | Profiles data missing/incorrect | Parse by column position with lenient whitespace splitting. If column count < 4, show error. |
| Large kanban board (100+ tasks) — JSON payload huge | Slow response, memory pressure | Limit to 50 tasks via `--status` filtering or truncate in Go after parsing. Only send id/title/assignee/status to frontend (strip body). |
| Command hangs indefinitely (e.g., gateway unresponsive) | Request handler hangs | Context timeout (15s max) enforced on all exec calls. |
| `df -h` slow on network mounts | System section hangs | Run `df -h /` to only query root filesystem. If still slow, timeout at 5s. |
| Binary size bloated by embedded dashboard | Unnecessary concern for local tool | Single HTML file < 15KB. Negligible. |

### Known Gaps (not in scope for v1)
- No process-level monitoring (CPU per process, top-like view)
- No network I/O monitoring
- No historical data / time series
- No alerting or threshold notifications
- No GPU monitoring (though system has AMD Strix Halo)
- No multiple gateway profile support beyond `hermes gateway status` output

## Appendix: Current Implementation Analysis

The existing `api.py` (149 lines) has these known weaknesses the Go rewrite will fix:

1. **Text parsing for kanban** — parses icon characters and regex; Go will use `--json`
2. **Brittle profile parsing** — regex `[◆ ]*(\w+)` only extracts name; Go will parse all 5 columns
3. **MCP parser bugs** — line 58 splits on all whitespace, line 70 has dead condition; Go will use structured column parsing
4. **No system metrics** — free/df/uptime not collected; Go will add them
5. **No graceful degradation** — any command failure returns partial/empty JSON without error indicators; Go will use SectionResult with error fields
6. **No timeouts on commands** — subprocess.run timeout=15 only on top-level, but no per-command differentiation; Go will use context timeouts per collector
7. **Sessions and cron endpoints** — out of scope for Go rewrite per task spec (only Profiles, Kanban, MCP, System required)
