// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agents

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"huginn/internal/domain/agent"
)

// KiloAdapter es el runtime opcional de código; el CLI `kilo` no está verificado.
// Solo ejecuta con plantilla explícita HUGINN_KILO_ARGS aprobada por el operador.
// En otro caso reporta error honesto sin adivinar flags.
type KiloAdapter struct{ bin string }

func NewKiloAdapter() *KiloAdapter  { return &KiloAdapter{bin: resolveBin("kilocode")} }
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
	fail := func(err error) (agent.AgentResult, error) {
		return agent.AgentResult{
			TaskID: task.ID, Agent: a.Name(), Provider: a.ID(), Status: "error",
			Errors: []string{err.Error()}, StartedAt: start, FinishedAt: time.Now(),
		}, err
	}
	if ok, _ := a.Detect(); !ok {
		return fail(fmt.Errorf("kilo not installed"))
	}
	// Sin contrato CLI verificado: rehúsa adivinar flags.
	return fail(fmt.Errorf("kilo: no verified CLI contract — execution disabled (see docs/agents.md)"))
}
