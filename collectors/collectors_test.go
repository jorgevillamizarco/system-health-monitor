package collectors

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRunner struct {
	responses map[string][]byte
	errors    map[string]error
}

func (f fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	key := name
	for _, arg := range args {
		key += " " + arg
	}
	if err := f.errors[key]; err != nil {
		return nil, err
	}
	if output, ok := f.responses[key]; ok {
		return output, nil
	}
	return nil, errors.New("missing fake response")
}

func TestParseProfiles(t *testing.T) {
	input := []byte(`Profile          Model                        Gateway      Alias        Distribution
 ───────────────    ───────────────────────────    ───────────    ───────────    ────────────────────
  default         gpt-5.5                      running      —            —
  builder         gpt-5.4                      stopped      —            —
 ◆researcher      deepseek-v4-pro              stopped      —            —
`)

	data, err := parseProfiles(input)
	if err != nil {
		t.Fatalf("parseProfiles returned error: %v", err)
	}
	if len(data.Profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(data.Profiles))
	}
	if !data.Profiles[2].Active || data.Profiles[2].Name != "researcher" {
		t.Fatalf("expected researcher to be active, got %+v", data.Profiles[2])
	}
}

func TestParseKanbanJSON(t *testing.T) {
	input := []byte(`[
  {"id":"t_a","title":"first","assignee":"builder","status":"running","priority":80},
  {"id":"t_b","title":"done task","assignee":"reviewer","status":"done","priority":70},
  {"id":"t_c","title":"todo task","assignee":"builder","status":"todo","priority":100}
]`)

	data, err := parseKanbanJSON(input)
	if err != nil {
		t.Fatalf("parseKanbanJSON returned error: %v", err)
	}
	if data.Counts["running"] != 1 || data.Counts["done"] != 1 || data.Counts["todo"] != 1 {
		t.Fatalf("unexpected counts: %+v", data.Counts)
	}
	if len(data.Tasks) != 2 || data.Tasks[0].ID != "t_c" {
		t.Fatalf("expected todo task first by priority, got %+v", data.Tasks)
	}
}

func TestParseMCP(t *testing.T) {
	input := []byte(`MCP Servers:

  Name             Transport                      Tools        Status
  ──────────────── ────────────────────────────── ──────────── ──────────
  discovery        http://localhost:8001/mcp      all          ✓ enabled
  chrome-devtools  /home/jorge/local/node/bin     all          failed to connect
`)

	data, err := parseMCP(input)
	if err != nil {
		t.Fatalf("parseMCP returned error: %v", err)
	}
	if data.Enabled != 1 || data.Total != 2 {
		t.Fatalf("unexpected counts: %+v", data)
	}
	if data.Servers[1].Enabled {
		t.Fatalf("expected second server to be disabled: %+v", data.Servers[1])
	}
}

func TestParseGateway(t *testing.T) {
	input := []byte(`✓ Gateway is running (PID: 79780)
  (Running manually, not as a system service)

Other profiles:
  ✓ default          — PID 79780
`)

	data, err := parseGateway(input)
	if err != nil {
		t.Fatalf("parseGateway returned error: %v", err)
	}
	if !data.Running || data.PID != 79780 {
		t.Fatalf("unexpected gateway data: %+v", data)
	}
	if len(data.Profiles) != 1 || data.Profiles[0].Name != "default" {
		t.Fatalf("unexpected profiles: %+v", data.Profiles)
	}
}

func TestParseSystem(t *testing.T) {
	freeOutput := []byte(`               total        used        free      shared  buff/cache   available
Mem:            30Gi       7.3Gi       976Mi        19Mi        22Gi        23Gi
Swap:          4.0Gi       1.6Gi       2.4Gi
`)
	diskOutput := []byte(`Filesystem      Size  Used Avail Use% Mounted on
/dev/nvme0n1p4  466G  328G  115G  75% /
`)
	uptimeOutput := []byte(`15:54:21 up  4:03,  1 user,  load average: 0.51, 0.45, 0.43`)

	data, err := parseSystem(freeOutput, diskOutput, uptimeOutput)
	if err != nil {
		t.Fatalf("parseSystem returned error: %v", err)
	}
	if data.Memory.PercentUsed == 0 || data.Disk.PercentUsed != 75 {
		t.Fatalf("unexpected metrics: %+v", data)
	}
	if data.Uptime.Users != 1 || data.Uptime.Load1 != 0.51 {
		t.Fatalf("unexpected uptime: %+v", data.Uptime)
	}
}

func TestCollectSystemCommandFailure(t *testing.T) {
	runner := fakeRunner{errors: map[string]error{"free -h": context.DeadlineExceeded}}
	section := CollectSystem(context.Background(), runner)
	if section.Error == "" {
		t.Fatal("expected error on free -h failure")
	}
}

func TestErrorStringTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	<-ctx.Done()
	if got := errorString(ctx.Err()); got != "command timed out" {
		t.Fatalf("expected timeout message, got %q", got)
	}
}
