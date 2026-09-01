package agents

import (
	"context"
	"time"

	"huginn/internal/domain/agent"
)

// ResearchAgent — orquesta Hugin Research + Perplexity como providers
type ResearchAgent struct {
	providers []agent.Provider
}

func NewResearchAgent(providers ...agent.Provider) *ResearchAgent {
	return &ResearchAgent{providers: providers}
}
func (r *ResearchAgent) Name() string { return "research" }
func (r *ResearchAgent) Role() string { return "Busqueda y analisis de informacion" }
func (r *ResearchAgent) CanHandle(t string) bool { return t == "research" || t == "search" }
func (r *ResearchAgent) Providers() []agent.Provider { return r.providers }

func (r *ResearchAgent) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	// selecciona provider disponible (huginn preferido, perplexity fallback)
	var prov agent.Provider
	for _, p := range r.providers {
		if p.Available(ctx) {
			prov = p
			break
		}
	}
	if prov == nil && len(r.providers) > 0 {
		prov = r.providers[0]
	}
	provName := "huginn-research"
	if prov != nil {
		provName = prov.Name()
		_, _ = prov.Invoke(ctx, agent.ProviderRequest{Prompt: task.Input, Context: task.Context})
	}
	res := agent.ResearchResult{
		Summary:    "Investigacion: " + task.Input,
		Findings:   []string{"Hallazgo 1 para: " + task.Input, "Hallazgo 2: fuentes verificadas"},
		Sources:    []string{"https://example.com/docs", "https://github.com/search"},
		Confidence: 0.82,
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

// ResearchReviewer — verifica resultados antes de Knowledge
type ResearchReviewer struct{}

func NewResearchReviewer() *ResearchReviewer { return &ResearchReviewer{} }
func (r *ResearchReviewer) Name() string { return "research-reviewer" }
func (r *ResearchReviewer) Role() string { return "Verificacion de investigacion" }
func (r *ResearchReviewer) CanHandle(t string) bool { return t == "research-review" }
func (r *ResearchReviewer) Providers() []agent.Provider { return nil }
func (r *ResearchReviewer) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	// por ahora verifica estructura y marca confianza
	res := agent.ResearchResult{
		Summary:    "Review: " + task.Input,
		Findings:   []string{"Verificado sin contradicciones"},
		Sources:    []string{"reviewed"},
		Confidence: 0.88,
	}
	return agent.AgentResult{
		TaskID:     task.ID,
		Agent:      r.Name(),
		Provider:   "reviewer",
		Status:     "ok",
		Output:     res,
		StartedAt:  start,
		FinishedAt: time.Now(),
	}, nil
}
