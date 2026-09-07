// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agents

import (
	"context"
	"time"

	"huginn/internal/domain/agent"
)

// MemoryAgent lee y escribe la memoria del Vault.
// Sin provider devuelve un error honesto en vez de un "memory ok" inventado.
type MemoryAgent struct{ providers []agent.Provider }

func NewMemory(providers ...agent.Provider) *MemoryAgent {
	return &MemoryAgent{providers: providers}
}
func (m *MemoryAgent) Name() string                { return "memory" }
func (m *MemoryAgent) Role() string                { return "Lee/escribe memoria Vault" }
func (m *MemoryAgent) CanHandle(t string) bool     { return t == "memory" }
func (m *MemoryAgent) Providers() []agent.Provider { return m.providers }
func (m *MemoryAgent) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	resp, provName, err := invokeFirst(ctx, m.providers, agent.ProviderRequest{Prompt: task.Input, Context: task.Context})
	if err != nil {
		return agent.AgentResult{
			TaskID: task.ID, Agent: m.Name(), Provider: provName,
			Status: "error", Errors: []string{err.Error()},
			StartedAt: start, FinishedAt: time.Now(),
		}, err
	}
	return agent.AgentResult{TaskID: task.ID, Agent: m.Name(), Provider: provName, Status: "ok",
		Output: truncate(resp.Content, 1200), StartedAt: start, FinishedAt: time.Now()}, nil
}

// KnowledgeAgent organiza la investigación en conocimiento con el mismo contrato de honestidad.
type KnowledgeAgent struct{ providers []agent.Provider }

func NewKnowledge(providers ...agent.Provider) *KnowledgeAgent {
	return &KnowledgeAgent{providers: providers}
}
func (k *KnowledgeAgent) Name() string                { return "knowledge" }
func (k *KnowledgeAgent) Role() string                { return "Organiza conocimiento Research" }
func (k *KnowledgeAgent) CanHandle(t string) bool     { return t == "knowledge" }
func (k *KnowledgeAgent) Providers() []agent.Provider { return k.providers }
func (k *KnowledgeAgent) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	resp, provName, err := invokeFirst(ctx, k.providers, agent.ProviderRequest{Prompt: task.Input, Context: task.Context})
	if err != nil {
		return agent.AgentResult{
			TaskID: task.ID, Agent: k.Name(), Provider: provName,
			Status: "error", Errors: []string{err.Error()},
			StartedAt: start, FinishedAt: time.Now(),
		}, err
	}
	return agent.AgentResult{TaskID: task.ID, Agent: k.Name(), Provider: provName, Status: "ok",
		Output: truncate(resp.Content, 1200), StartedAt: start, FinishedAt: time.Now()}, nil
}

// ContextAgent recupera el contexto relevante del Vault a partir de los fragmentos recibidos.
// La síntesis con modelo requiere un provider configurado.
type ContextAgent struct{ providers []agent.Provider }

func NewContext(providers ...agent.Provider) *ContextAgent {
	return &ContextAgent{providers: providers}
}
func (c *ContextAgent) Name() string                { return "context" }
func (c *ContextAgent) Role() string                { return "Recupera contexto relevante del Vault" }
func (c *ContextAgent) CanHandle(t string) bool     { return t == "context" }
func (c *ContextAgent) Providers() []agent.Provider { return c.providers }
func (c *ContextAgent) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	if len(task.Context.Memory) > 0 && len(c.providers) == 0 {
		// Sin modelo expone los fragmentos recuperados: comportamiento real y útil.
		combined := ""
		for i, frag := range task.Context.Memory {
			if i > 0 {
				combined += "\n---\n"
			}
			combined += frag
		}
		return agent.AgentResult{TaskID: task.ID, Agent: c.Name(), Provider: "vault", Status: "ok",
			Output: truncate(combined, 2000), StartedAt: start, FinishedAt: time.Now()}, nil
	}
	resp, provName, err := invokeFirst(ctx, c.providers, agent.ProviderRequest{Prompt: task.Input, Context: task.Context})
	if err != nil {
		return agent.AgentResult{
			TaskID: task.ID, Agent: c.Name(), Provider: provName,
			Status: "error", Errors: []string{err.Error()},
			StartedAt: start, FinishedAt: time.Now(),
		}, err
	}
	return agent.AgentResult{TaskID: task.ID, Agent: c.Name(), Provider: provName, Status: "ok",
		Output: truncate(resp.Content, 2000), StartedAt: start, FinishedAt: time.Now()}, nil
}
