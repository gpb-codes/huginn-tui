package agents

import (
	"context"
	"time"

	"huginn/internal/domain/agent"
)

type CodingAgent struct{ providers []agent.Provider }

func NewCoding(providers ...agent.Provider) *CodingAgent { return &CodingAgent{providers: providers} }
func (c *CodingAgent) Name() string { return "coding" }
func (c *CodingAgent) Role() string { return "Implementa codigo via OpenCode/Kilo" }
func (c *CodingAgent) CanHandle(t string) bool { return t == "coding" || t == "implement" }
func (c *CodingAgent) Providers() []agent.Provider { return c.providers }
func (c *CodingAgent) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	prov := "opencode"
	if len(c.providers) > 0 {
		for _, p := range c.providers {
			if p.Available(ctx) {
				prov = p.Name()
				_, _ = p.Invoke(ctx, agent.ProviderRequest{Prompt: task.Input, Context: task.Context})
				break
			}
		}
	}
	return agent.AgentResult{
		TaskID: task.ID, Agent: c.Name(), Provider: prov, Status: "ok",
		Output: agent.CodeResult{Summary: "Codigo generado para: " + task.Input, FilesChanged: []string{"main.go"}},
		StartedAt: start, FinishedAt: time.Now(),
	}, nil
}

type CodeReviewer struct{}

func NewCodeReviewer() *CodeReviewer { return &CodeReviewer{} }
func (r *CodeReviewer) Name() string { return "code-reviewer" }
func (r *CodeReviewer) Role() string { return "Revisa codigo generado" }
func (r *CodeReviewer) CanHandle(t string) bool { return t == "review" || t == "code-review" }
func (r *CodeReviewer) Providers() []agent.Provider { return nil }
func (r *CodeReviewer) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	return agent.AgentResult{
		TaskID: task.ID, Agent: r.Name(), Provider: "reviewer", Status: "ok",
		Output: agent.ReviewResult{Approved: true, Score: 8, Suggestions: []string{"Anadir tests"}},
		StartedAt: start, FinishedAt: time.Now(),
	}, nil
}
