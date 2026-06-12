"""Hermes System Monitor — thin API backend. Parses hermes CLI output."""
import json, subprocess, sys, re
from http.server import HTTPServer, BaseHTTPRequestHandler

def run(cmd):
    try:
        r = subprocess.run(cmd, capture_output=True, text=True, timeout=15)
        return r.stdout, r.stderr
    except Exception as e:
        return "", str(e)

class Handler(BaseHTTPRequestHandler):
    def _send(self, data):
        body = json.dumps(data).encode()
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_OPTIONS(self):
        self._send({})

    def do_GET(self):
        if self.path == '/api/health':
            self._send({"status": "ok"})
            
        elif self.path == '/api/profiles':
            out, _ = run(['hermes', 'profile', 'list'])
            profiles = []
            for line in out.split('\n'):
                line = line.strip()
                if not line or line.startswith('─') or line.startswith('═') or 'Profile' in line:
                    continue
                # Match profile rows: "◆default", " builder", " researcher"
                m = re.match(r'[◆ ]*(\w+)', line)
                if m:
                    profiles.append({"name": m.group(1)})
            self._send({"profiles": profiles, "count": len(profiles)})
            
        elif self.path == '/api/mcp':
            out, _ = run(['hermes', 'mcp', 'list'])
            servers = []
            in_table = False
            for line in out.split('\n'):
                line = line.strip()
                if not line:
                    continue
                if line.startswith('MCP Servers'):
                    continue
                if line.startswith('Name') and 'Transport' in line:
                    in_table = True
                    continue
                if in_table and line.startswith('─'):
                    continue
                if in_table and line:
                    parts = line.split()
                    if not parts or parts[0] in ('Name', 'MCP'):
                        continue
                    name = parts[0]
                    # Skip non-server-name lines
                    if name.startswith('─') or name.startswith('═'):
                        continue
                    status = 'enabled' if ('enabled' in line or '✓' in line) else 'failed'
                    tools = 'all'
                    if len(parts) >= 3:
                        tools = parts[2] if parts[2] != 'all' else 'all'
                    servers.append({"name": name, "status": status, "tools": tools})
                if in_table and not line.startswith(' ') and parts and 'enabled' not in line:
                    in_table = False
            self._send({"servers": servers, "count": len(servers)})
            
        elif self.path == '/api/kanban':
            out, _ = run(['hermes', 'kanban', 'list'])
            tasks = []
            counts = {'done': 0, 'running': 0, 'todo': 0, 'blocked': 0, 'ready': 0}
            for line in out.split('\n'):
                line = line.strip()
                parts = line.split()
                m = re.match(r'[◻●◆✓○]?\s*(t_[a-f0-9]+)', line)
                if m:
                    tid = m.group(1)
                    status = 'todo'
                    if 'done' in line: status = 'done'
                    elif '●' in line or 'running' in line: status = 'running'
                    elif 'blocked' in line: status = 'blocked'
                    elif '◻' in line: status = 'todo'
                    tasks.append({"id": tid, "status": status})
                    counts[status] = counts.get(status, 0) + 1
            self._send({"tasks": tasks, "counts": counts, "total": len(tasks)})
            
        elif self.path == '/api/cron':
            out, _ = run(['hermes', 'cron', 'list'])
            jobs = []
            current = None
            for line in out.split('\n'):
                line = line.strip()
                # Detect job ID line: hex string with optional [active] tag
                m = re.match(r'([a-f0-9]{12,})\s*(\[.*?\])?', line)
                if m and not line.startswith('┌') and not line.startswith('│') and not line.startswith('└'):
                    if current:
                        jobs.append(current)
                    current = {"id": m.group(1), "status": (m.group(2) or "[active]").strip("[]")}
                elif current:
                    # Key-value pair: "Key:  Value"
                    kv = re.match(r'(\w[\w\s]*?):\s+(.+)', line)
                    if kv:
                        current[kv.group(1).strip().lower()] = kv.group(2).strip()
            if current:
                jobs.append(current)
            self._send({"jobs": jobs, "count": len(jobs)})
            
        elif self.path == '/api/sessions':
            out, _ = run(['hermes', 'sessions', 'list'])
            sessions = []
            in_table = False
            for line in out.split('\n'):
                line = line.strip()
                if not line:
                    continue
                if 'ID' in line and 'Title' in line:
                    in_table = True
                    continue
                if in_table and line.startswith('─'):
                    continue
                if in_table and line:
                    # Parse: "Title Preview... LastActive ID"
                    # ID is always last word (format: YYYYMMDD_HHMMSS_xxxxx or cron_xxx_...)
                    parts = line.rsplit(None, 2)  # ["title...preview", "lastActive", "ID"]
                    if len(parts) >= 2:
                        # ID is the last part, text is everything before it
                        wid = parts[-1]
                        # Get everything before last 2 parts as the session text
                        text = ' '.join(parts[:-1])
                        sessions.append({"id": wid, "text": text})
            self._send({"sessions": sessions[:10], "count": len(sessions)})
            
        else:
            self._send({"error": "not found"})

    def log_message(self, *args):
        pass

if __name__ == '__main__':
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 9092
    server = HTTPServer(('127.0.0.1', port), Handler)
    print(f'API on :{port}')
    server.serve_forever()
