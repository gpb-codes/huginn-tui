// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agents

import (
	"context"
	"fmt"
	"time"

	"huginn/internal/domain/agent"
)

type CodingAgent struct{ providers []agent.Provider }

func NewCoding(providers ...agent.Provider) *CodingAgent { return &CodingAgent{providers: providers} }
func (c *CodingAgent) Name() string                      { return "coding" }
func (c *CodingAgent) Role() string                      { return "Implementa codigo via OpenCode/Kilo" }
func (c *CodingAgent) CanHandle(t string) bool           { return t == "coding" || t == "implement" }
func (c *CodingAgent) Providers() []agent.Provider       { return c.providers }
func (c *CodingAgent) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	resp, provName, err := invokeFirst(ctx, c.providers, agent.ProviderRequest{Prompt: task.Input, Context: task.Context})
	if err != nil {
		return agent.AgentResult{
			TaskID: task.ID, Agent: c.Name(), Provider: provName,
			Status: "error", Errors: []string{err.Error()},
			StartedAt: start, FinishedAt: time.Now(),
		}, err
	}
	return agent.AgentResult{
		TaskID: task.ID, Agent: c.Name(), Provider: provName, Status: "ok",
		Output:     agent.CodeResult{Summary: resp.Content},
		StartedAt:  start,
		FinishedAt: time.Now(),
	}, nil
}

type CodeReviewer struct{ providers []agent.Provider }

func NewCodeReviewer(providers ...agent.Provider) *CodeReviewer {
	return &CodeReviewer{providers: providers}
}
func (r *CodeReviewer) Name() string                { return "code-reviewer" }
func (r *CodeReviewer) Role() string                { return "Revisa codigo generado" }
func (r *CodeReviewer) CanHandle(t string) bool     { return t == "review" || t == "code-review" }
func (r *CodeReviewer) Providers() []agent.Provider { return r.providers }
func (r *CodeReviewer) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	// Sin provider no hay juicio posible: no aprobar a ciegas.
	if len(r.providers) == 0 {
		err := fmt.Errorf("code-reviewer: sin provider de revisión configurado")
		return agent.AgentResult{
			TaskID: task.ID, Agent: r.Name(), Provider: "reviewer",
			Status: "error", Errors: []string{err.Error()},
			StartedAt: start, FinishedAt: time.Now(),
		}, err
	}
	return agent.AgentResult{
		TaskID: task.ID, Agent: r.Name(), Provider: "reviewer", Status: "ok",
		Output: agent.ReviewResult{
			Approved:    false, // el humano o el evaluator deciden
			Suggestions: []string{"Derivar a runtime de revisión con el diff como input"},
		},
		StartedAt: start, FinishedAt: time.Now(),
	}, nil
}
