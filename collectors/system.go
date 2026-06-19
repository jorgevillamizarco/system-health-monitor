package collectors

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var memoryPattern = regexp.MustCompile(`^Mem:\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)$`)
var swapPattern = regexp.MustCompile(`^Swap:\s+(\S+)\s+(\S+)\s+(\S+)$`)
var uptimePattern = regexp.MustCompile(`up\s+(.+?),\s+(\d+) user[s]?,\s+load average:\s+([\d.]+),\s+([\d.]+),\s+([\d.]+)`)

func CollectSystem(ctx context.Context, runner Runner) SectionResult[SystemData] {
	memoryCtx, cancelMemory := context.WithTimeout(ctx, 5*time.Second)
	defer cancelMemory()
	freeOutput, err := runner.Run(memoryCtx, "free", "-h")
	if err != nil {
		return SectionResult[SystemData]{Error: errorString(err)}
	}

	diskCtx, cancelDisk := context.WithTimeout(ctx, 5*time.Second)
	defer cancelDisk()
	diskOutput, err := runner.Run(diskCtx, "df", "-h", "/")
	if err != nil {
		return SectionResult[SystemData]{Error: errorString(err)}
	}

	uptimeCtx, cancelUptime := context.WithTimeout(ctx, 5*time.Second)
	defer cancelUptime()
	uptimeOutput, err := runner.Run(uptimeCtx, "uptime")
	if err != nil {
		return SectionResult[SystemData]{Error: errorString(err)}
	}

	data, parseErr := parseSystem(freeOutput, diskOutput, uptimeOutput)
	if parseErr != nil {
		return SectionResult[SystemData]{Error: parseErr.Error()}
	}
	return SectionResult[SystemData]{Data: &data}
}

func parseSystem(freeOutput, diskOutput, uptimeOutput []byte) (SystemData, error) {
	memory, swap, err := parseFree(freeOutput)
	if err != nil {
		return SystemData{}, err
	}
	disk, err := parseDisk(diskOutput)
	if err != nil {
		return SystemData{}, err
	}
	uptime, err := parseUptime(uptimeOutput)
	if err != nil {
		return SystemData{}, err
	}

	return SystemData{Memory: memory, Disk: disk, Swap: swap, Uptime: uptime}, nil
}

func parseFree(output []byte) (SystemMetrics, *SystemMetrics, error) {
	var memory SystemMetrics
	var swap *SystemMetrics

	for _, raw := range strings.Split(string(output), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if match := memoryPattern.FindStringSubmatch(line); len(match) == 7 {
			memory = SystemMetrics{
				Label:       "Memory",
				Total:       match[1],
				Used:        match[2],
				Available:   match[6],
				PercentUsed: percentInt(match[2], match[1]),
			}
		}
		if match := swapPattern.FindStringSubmatch(line); len(match) == 4 {
			total := match[1]
			used := match[2]
			metric := SystemMetrics{Label: "Swap", Total: total, Used: used, Available: match[3], PercentUsed: percentInt(used, total)}
			swap = &metric
		}
	}

	if memory.Total == "" {
		return SystemMetrics{}, nil, errors.New("unable to parse free -h output")
	}
	return memory, swap, nil
}

func parseDisk(output []byte) (SystemMetrics, error) {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return SystemMetrics{}, errors.New("unable to parse df -h output")
	}

	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < 6 {
		return SystemMetrics{}, errors.New("unable to parse root disk row")
	}

	percent := strings.TrimSuffix(fields[4], "%")
	percentIntValue, _ := strconv.Atoi(percent)
	return SystemMetrics{
		Label:       "Disk",
		Total:       fields[1],
		Used:        fields[2],
		Available:   fields[3],
		PercentUsed: percentIntValue,
	}, nil
}

func parseUptime(output []byte) (UptimeData, error) {
	match := uptimePattern.FindStringSubmatch(strings.TrimSpace(string(output)))
	if len(match) != 6 {
		return UptimeData{}, errors.New("unable to parse uptime output")
	}

	users, _ := strconv.Atoi(match[2])
	load1, _ := strconv.ParseFloat(match[3], 64)
	load5, _ := strconv.ParseFloat(match[4], 64)
	load15, _ := strconv.ParseFloat(match[5], 64)
	return UptimeData{Uptime: match[1], Users: users, Load1: load1, Load5: load5, Load15: load15}, nil
}

func percentInt(used, total string) int {
	usedValue, err := parseSizeValue(used)
	if err != nil {
		return 0
	}
	totalValue, err := parseSizeValue(total)
	if err != nil || totalValue == 0 {
		return 0
	}
	return int((usedValue / totalValue) * 100)
}

func parseSizeValue(input string) (float64, error) {
	if input == "0B" || input == "0" {
		return 0, nil
	}
	unit := input[len(input)-2:]
	numberPart := input[:len(input)-2]
	value, err := strconv.ParseFloat(numberPart, 64)
	if err != nil {
		return 0, fmt.Errorf("parse size %q: %w", input, err)
	}
	multiplier := 1.0
	switch unit {
	case "Ki":
		multiplier = 1.0 / 1024
	case "Mi":
		multiplier = 1
	case "Gi":
		multiplier = 1024
	case "Ti":
		multiplier = 1024 * 1024
	case "KB":
		multiplier = 1.0 / 1000
	case "MB":
		multiplier = 1
	case "GB":
		multiplier = 1000
	case "TB":
		multiplier = 1000 * 1000
	default:
		return 0, fmt.Errorf("unknown size unit %q", unit)
	}
	return value * multiplier, nil
}
