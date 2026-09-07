// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agents

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"huginn/internal/domain/agent"
)

// HermesAdapter es el agente Hermes de Nous Research; solo detección.
// Sin forma CLI verificada, Execute rehúsa falsear éxito y devuelve error descriptivo.
type HermesAdapter struct{ bin string }

func NewHermesAdapter() *HermesAdapter { return &HermesAdapter{bin: resolveBin("hermes")} }
func (a *HermesAdapter) ID() string    { return "hermes" }
func (a *HermesAdapter) Name() string  { return "Hermes" }
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
func (a *HermesAdapter) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	err := fmt.Errorf("hermes: execution not implemented — CLI shape unverified, refusing to fake success")
	return agent.AgentResult{
		TaskID: task.ID, Agent: a.Name(), Status: "error",
		Errors:     []string{err.Error()},
		StartedAt:  start,
		FinishedAt: time.Now(),
	}, err
}
