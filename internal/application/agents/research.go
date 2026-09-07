// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agents

import (
	"context"
	"time"

	"huginn/internal/domain/agent"
)

// ResearchAgent orquesta Hugin Research y Perplexity como providers.
type ResearchAgent struct {
	providers []agent.Provider
}

func NewResearchAgent(providers ...agent.Provider) *ResearchAgent {
	return &ResearchAgent{providers: providers}
}
func (r *ResearchAgent) Name() string                { return "research" }
func (r *ResearchAgent) Role() string                { return "Busqueda y analisis de informacion" }
func (r *ResearchAgent) CanHandle(t string) bool     { return t == "research" || t == "search" }
func (r *ResearchAgent) Providers() []agent.Provider { return r.providers }

func (r *ResearchAgent) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	resp, provName, err := invokeFirst(ctx, r.providers, agent.ProviderRequest{Prompt: task.Input, Context: task.Context})
	if err != nil {
		return agent.AgentResult{
			TaskID: task.ID, Agent: r.Name(), Provider: provName,
			Status: "error", Errors: []string{err.Error()},
			StartedAt: start, FinishedAt: time.Now(),
		}, err
	}
	res := agent.ResearchResult{
		Summary:            truncate(resp.Content, 1200),
		Confidence:         0.5, // model output, unverified — reviewer decides
		RecommendedActions: []string{"Revisar con research-reviewer antes de persistir en Vault"},
	}
	return agent.AgentResult{
		TaskID:     task.ID,
		Agent:      r.Name(),
		Provider:   provName,
		Status:     "ok",
		Output:     res,
		StartedAt:  start,
		FinishedAt: time.Now(),
	}, nil
}

// ResearchReviewer verifica los resultados antes de pasar a Knowledge.
type ResearchReviewer struct{}

func NewResearchReviewer() *ResearchReviewer            { return &ResearchReviewer{} }
func (r *ResearchReviewer) Name() string                { return "research-reviewer" }
func (r *ResearchReviewer) Role() string                { return "Verificacion de investigacion" }
func (r *ResearchReviewer) CanHandle(t string) bool     { return t == "research-review" }
func (r *ResearchReviewer) Providers() []agent.Provider { return nil }
func (r *ResearchReviewer) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	// Revisión estructural: comprueba que el input tenga contenido revisable.
	// La revisión semántica la hace el pipeline (reviewer runtime + evaluator).
	issues := []string{}
	if len(task.Input) < 10 {
		issues = append(issues, "contenido insuficiente para revisar")
	}
	out := agent.ReviewResult{
		Approved:    len(issues) == 0,
		Issues:      issues,
		Suggestions: []string{"Derivar a runtime de revisión (ollama/opencode) para juicio semántico"},
		Score:       6,
	}
	return agent.AgentResult{
		TaskID:     task.ID,
		Agent:      r.Name(),
		Provider:   "reviewer",
		Status:     "ok",
		Output:     out,
		StartedAt:  start,
		FinishedAt: time.Now(),
	}, nil
}
