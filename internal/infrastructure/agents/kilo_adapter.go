package agents

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"huginn/internal/domain/agent"
)

type KiloAdapter struct{ bin string }

func NewKiloAdapter() *KiloAdapter { return &KiloAdapter{bin: resolveBin("kilocode")} }
func (a *KiloAdapter) ID() string   { return "kilo" }
func (a *KiloAdapter) Name() string { return "Kilo Code" }
func (a *KiloAdapter) Detect() (bool, string) {
	if a.bin == "" {
		if p, err := exec.LookPath("kilo"); err == nil {
			a.bin = p
		} else {
			return false, "NOT_INSTALLED"
		}
	}
	out, err := exec.Command(a.bin, "--version").CombinedOutput()
	if err != nil {
		return true, "unknown"
	}
	return true, strings.TrimSpace(string(out))
}
func (a *KiloAdapter) Capabilities() []string {
	return []string{"code_generation", "code_editing", "terminal_execution", "mcp", "subagents"}
}
func (a *KiloAdapter) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	if ok, _ := a.Detect(); !ok {
		return agent.AgentResult{TaskID: task.ID, Agent: a.Name(), Status: "error", Errors: []string{"kilo not installed"}, StartedAt: start, FinishedAt: time.Now()}, fmt.Errorf("kilo not installed")
	}
	// kilo run similar to opencode — usa opencode run como fallback si gateway no disponible
	args := []string{"run", "--format", "json", task.Input}
	cmd := exec.CommandContext(ctx, a.bin, args...)
	out, err := cmd.CombinedOutput()
	text := parseOpenCodeJSON(strings.TrimSpace(string(out)))
	if text == "" {
		text = strings.TrimSpace(string(out))
	}
	status := "ok"
	var errs []string
	if err != nil {
		status = "error"
		errs = []string{err.Error()}
	}
	return agent.AgentResult{TaskID: task.ID, Agent: a.Name(), Provider: a.ID(), Status: status, Output: agent.CodeResult{Summary: text}, Errors: errs, StartedAt: start, FinishedAt: time.Now()}, err
}
