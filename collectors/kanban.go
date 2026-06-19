package collectors

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"
)

var kanbanTextPattern = regexp.MustCompile(`^(?:[●◻✓⊘◆○]\s+)?(t_[a-f0-9]+)\s+([a-z_]+)\s+(\S+)\s+(.+)$`)

func CollectKanban(ctx context.Context, runner Runner) SectionResult[KanbanData] {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	output, err := runner.Run(ctx, "hermes", "kanban", "list", "--json")
	if err == nil {
		data, parseErr := parseKanbanJSON(output)
		if parseErr == nil {
			return SectionResult[KanbanData]{Data: &data}
		}
	}

	fallbackOutput, fallbackErr := runner.Run(ctx, "hermes", "kanban", "list")
	if fallbackErr != nil {
		if err != nil {
			return SectionResult[KanbanData]{Error: errorString(err)}
		}
		return SectionResult[KanbanData]{Error: errorString(fallbackErr)}
	}

	data, parseErr := parseKanbanText(fallbackOutput)
	if parseErr != nil {
		if err != nil {
			return SectionResult[KanbanData]{Error: parseErr.Error() + "; json path: " + errorString(err)}
		}
		return SectionResult[KanbanData]{Error: parseErr.Error()}
	}
	return SectionResult[KanbanData]{Data: &data}
}

func parseKanbanJSON(output []byte) (KanbanData, error) {
	var raw []KanbanTask
	if err := json.Unmarshal(output, &raw); err != nil {
		return KanbanData{}, err
	}

	counts := map[string]int{}
	for _, task := range raw {
		counts[task.Status]++
	}

	sort.Slice(raw, func(i, j int) bool {
		return raw[i].Priority > raw[j].Priority
	})

	tasks := make([]KanbanTask, 0, len(raw))
	for _, task := range raw {
		if task.Status == "done" {
			continue
		}
		tasks = append(tasks, KanbanTask{
			ID:       task.ID,
			Title:    task.Title,
			Assignee: task.Assignee,
			Status:   task.Status,
			Priority: task.Priority,
		})
		if len(tasks) == 8 {
			break
		}
	}

	return KanbanData{Counts: counts, Tasks: tasks}, nil
}

func parseKanbanText(output []byte) (KanbanData, error) {
	lines := strings.Split(string(output), "\n")
	counts := map[string]int{}
	tasks := make([]KanbanTask, 0)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		match := kanbanTextPattern.FindStringSubmatch(line)
		if len(match) != 5 {
			continue
		}

		status := match[2]
		counts[status]++
		if status == "done" {
			continue
		}
		tasks = append(tasks, KanbanTask{
			ID:       match[1],
			Status:   status,
			Assignee: match[3],
			Title:    match[4],
		})
	}

	if len(tasks) == 0 && len(counts) == 0 {
		return KanbanData{}, errors.New("unable to parse kanban output")
	}
	if len(tasks) > 8 {
		tasks = tasks[:8]
	}
	return KanbanData{Counts: counts, Tasks: tasks}, nil
}
