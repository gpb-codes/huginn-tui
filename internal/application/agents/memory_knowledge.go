package agents

import (
	"context"
	"time"

	"huginn/internal/domain/agent"
)

type MemoryAgent struct{}
func NewMemory() *MemoryAgent { return &MemoryAgent{} }
func (m *MemoryAgent) Name() string { return "memory" }
func (m *MemoryAgent) Role() string { return "Lee/escribe memoria Vault" }
func (m *MemoryAgent) CanHandle(t string) bool { return t == "memory" }
func (m *MemoryAgent) Providers() []agent.Provider { return nil }
func (m *MemoryAgent) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	return agent.AgentResult{TaskID: task.ID, Agent: m.Name(), Provider: "memory", Status: "ok", Output: "memory ok", StartedAt: start, FinishedAt: time.Now()}, nil
}

type KnowledgeAgent struct{}
func NewKnowledge() *KnowledgeAgent { return &KnowledgeAgent{} }
func (k *KnowledgeAgent) Name() string { return "knowledge" }
func (k *KnowledgeAgent) Role() string { return "Organiza conocimiento Research" }
func (k *KnowledgeAgent) CanHandle(t string) bool { return t == "knowledge" }
func (k *KnowledgeAgent) Providers() []agent.Provider { return nil }
func (k *KnowledgeAgent) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	return agent.AgentResult{TaskID: task.ID, Agent: k.Name(), Provider: "knowledge", Status: "ok", Output: "knowledge ok", StartedAt: start, FinishedAt: time.Now()}, nil
}

type ContextAgent struct{}
func NewContext() *ContextAgent { return &ContextAgent{} }
func (c *ContextAgent) Name() string { return "context" }
func (c *ContextAgent) Role() string { return "Recupera contexto relevante del Vault" }
func (c *ContextAgent) CanHandle(t string) bool { return t == "context" }
func (c *ContextAgent) Providers() []agent.Provider { return nil }
func (c *ContextAgent) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	return agent.AgentResult{TaskID: task.ID, Agent: c.Name(), Provider: "context", Status: "ok", Output: "context ok para: " + task.Input, StartedAt: start, FinishedAt: time.Now()}, nil
}
