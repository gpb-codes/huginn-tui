package agents

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"huginn/internal/domain/agent"
)

// HermesAdapter — Nous Research Hermes Agent, frontera clara, no copia interna
type HermesAdapter struct{ bin string }

func NewHermesAdapter() *HermesAdapter { return &HermesAdapter{bin: resolveBin("hermes")} }
func (a *HermesAdapter) ID() string   { return "hermes" }
func (a *HermesAdapter) Name() string { return "Hermes" }
func (a *HermesAdapter) Detect() (bool, string) {
	if a.bin == "" {
		return false, "NOT_INSTALLED"
	}
	out, err := exec.Command(a.bin, "--version").CombinedOutput()
	if err != nil {
		return true, "unknown"
	}
	return true, string(out)
}
func (a *HermesAdapter) Capabilities() []string {
	return []string{"memory", "skills", "subagents", "cron", "mcp", "tool_calling", "streaming"}
}
func (a *HermesAdapter) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	if ok, _ := a.Detect(); !ok {
		// no instalado — no crash, retorna UNKNOWN con instrucción
		return agent.AgentResult{
			TaskID: task.ID, Agent: a.Name(), Status: "error",
			Errors: []string{"hermes not installed — instala con curl -fsSL https://hermes-agent.nousresearch.com/install | bash"},
			StartedAt: start, FinishedAt: time.Now(),
		}, fmt.Errorf("hermes not installed")
	}
	// hermes run <prompt> — placeholder, por ahora delega a mock
	_ = ctx
	return agent.AgentResult{TaskID: task.ID, Agent: a.Name(), Status: "ok", Output: "hermes ok: " + task.Input, StartedAt: start, FinishedAt: time.Now()}, nil
}
