package collectors

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var gatewayPIDPattern = regexp.MustCompile(`PID:\s*(\d+)`)
var gatewayProfilePattern = regexp.MustCompile(`^[✓●]\s+([\w-]+)\s+—\s+PID\s+(\d+)`)

func CollectGateway(ctx context.Context, runner Runner) SectionResult[GatewayData] {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	output, err := runner.Run(ctx, "hermes", "gateway", "status")
	if err != nil {
		return SectionResult[GatewayData]{Error: errorString(err)}
	}

	data, parseErr := parseGateway(output)
	if parseErr != nil {
		return SectionResult[GatewayData]{Error: parseErr.Error()}
	}
	return SectionResult[GatewayData]{Data: &data}
}

func parseGateway(output []byte) (GatewayData, error) {
	text := strings.TrimSpace(string(output))
	if text == "" {
		return GatewayData{}, errors.New("empty gateway output")
	}

	result := GatewayData{
		Running: strings.Contains(strings.ToLower(text), "gateway is running"),
		Summary: firstLine(text),
	}

	if pidMatch := gatewayPIDPattern.FindStringSubmatch(text); len(pidMatch) == 2 {
		pid, _ := strconv.Atoi(pidMatch[1])
		result.PID = pid
	}

	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		match := gatewayProfilePattern.FindStringSubmatch(line)
		if len(match) != 3 {
			continue
		}
		pid, _ := strconv.Atoi(match[2])
		result.Profiles = append(result.Profiles, GatewayProfile{Name: match[1], PID: pid})
	}

	return result, nil
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")
	return strings.TrimSpace(line)
}
