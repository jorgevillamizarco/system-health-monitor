package collectors

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
)

var mcpLinePattern = regexp.MustCompile(`^([\w-]+)\s{2,}(.+?)\s{2,}(all|\d+)\s{2,}(.+)$`)

func CollectMCP(ctx context.Context, runner Runner) SectionResult[MCPData] {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	output, err := runner.Run(ctx, "hermes", "mcp", "list")
	if err != nil {
		return SectionResult[MCPData]{Error: errorString(err)}
	}

	data, parseErr := parseMCP(output)
	if parseErr != nil {
		return SectionResult[MCPData]{Error: parseErr.Error()}
	}
	return SectionResult[MCPData]{Data: &data}
}

func parseMCP(output []byte) (MCPData, error) {
	lines := strings.Split(string(output), "\n")
	servers := make([]MCPServer, 0)
	enabled := 0

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || line == "MCP Servers:" || strings.HasPrefix(line, "Name") || strings.HasPrefix(line, "─") {
			continue
		}

		match := mcpLinePattern.FindStringSubmatch(line)
		if len(match) != 5 {
			continue
		}

		status := strings.TrimSpace(match[4])
		isEnabled := strings.Contains(status, "enabled") || strings.Contains(status, "✓")
		if isEnabled {
			enabled++
		}
		servers = append(servers, MCPServer{
			Name:      strings.TrimSpace(match[1]),
			Transport: strings.TrimSpace(match[2]),
			Tools:     strings.TrimSpace(match[3]),
			Status:    status,
			Enabled:   isEnabled,
		})
	}

	if len(servers) == 0 {
		return MCPData{}, errors.New("unable to parse mcp output")
	}
	return MCPData{Servers: servers, Enabled: enabled, Total: len(servers)}, nil
}
