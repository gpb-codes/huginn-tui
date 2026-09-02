package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"huginn/internal/domain/agent"
)

// OpenCodeAdapter — referencia arquitectónica para los demás adapters
type OpenCodeAdapter struct {
	bin string
}

func NewOpenCodeAdapter() *OpenCodeAdapter { return &OpenCodeAdapter{bin: resolveBin("opencode")} }

func (a *OpenCodeAdapter) ID() string   { return "opencode" }
func (a *OpenCodeAdapter) Name() string { return "OpenCode" }

func (a *OpenCodeAdapter) Detect() (bool, string) {
	if a.bin == "" {
		return false, "NOT_INSTALLED"
	}
	out, err := exec.Command(a.bin, "--version").CombinedOutput()
	if err != nil {
		return false, "ERROR"
	}
	return true, strings.TrimSpace(string(out))
}

func (a *OpenCodeAdapter) Capabilities() []string {
	return []string{"code_generation", "code_editing", "terminal_execution", "file_operations", "git", "streaming", "tool_calling"}
}

func (a *OpenCodeAdapter) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	if ok, _ := a.Detect(); !ok {
		return agent.AgentResult{TaskID: task.ID, Agent: a.Name(), Provider: a.ID(), Status: "error", Errors: []string{"opencode not installed"}, StartedAt: start, FinishedAt: time.Now()}, fmt.Errorf("opencode not installed")
	}
	// opencode run [message..] --format json
	args := []string{"run", "--format", "json", task.Input}
	cmd := exec.CommandContext(ctx, a.bin, args...)
	out, err := cmd.CombinedOutput()
	raw := strings.TrimSpace(string(out))
	text := parseOpenCodeJSON(raw)
	if text == "" {
		text = raw
	}
	status := "ok"
	var errs []string
	if err != nil {
		status = "error"
		errs = []string{err.Error()}
	}
	return agent.AgentResult{
		TaskID: task.ID, Agent: a.Name(), Provider: a.ID(), Status: status,
		Output: agent.CodeResult{Summary: text, FilesChanged: []string{}},
		Errors: errs, StartedAt: start, FinishedAt: time.Now(),
	}, err
}

func parseOpenCodeJSON(raw string) string {
	var text string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err == nil {
			if typ, ok := obj["type"].(string); ok && typ == "text" {
				if part, ok := obj["part"].(map[string]interface{}); ok {
					if t, ok := part["text"].(string); ok {
						text += t
					}
				}
			}
		}
	}
	return strings.TrimSpace(text)
}

func resolveBin(name string) string {
	if runtime.GOOS == "windows" {
		for _, cand := range []string{name + ".cmd", name + ".ps1", name + ".exe", name} {
			if _, err := exec.LookPath(cand); err == nil {
				return cand
			}
		}
		return ""
	}
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	return ""
}
