package collectors

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return output, nil
	}

	trimmed := sanitizeCommandOutput(output)
	if trimmed == "" {
		return nil, err
	}
	return nil, fmt.Errorf("%w: %s", err, trimmed)
}

func sanitizeCommandOutput(output []byte) string {
	text := strings.TrimSpace(string(output))
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "\n", " | ")
	text = strings.ReplaceAll(text, "\r", "")
	if len(text) > 240 {
		text = text[:240] + "..."
	}
	return text
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "command timed out"
	}
	msg := err.Error()
	if len(msg) > 240 {
		msg = msg[:240] + "..."
	}
	return msg
}
