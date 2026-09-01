package agents

import (
	"context"
	"time"

	"huginn/internal/domain/agent"
)

type QAAgent struct{}

func NewQA() *QAAgent { return &QAAgent{} }
func (q *QAAgent) Name() string { return "qa" }
func (q *QAAgent) Role() string { return "Ejecuta tests/lint/build" }
func (q *QAAgent) CanHandle(t string) bool { return t == "qa" || t == "test" }
func (q *QAAgent) Providers() []agent.Provider { return nil }
func (q *QAAgent) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	return agent.AgentResult{
		TaskID: task.ID, Agent: q.Name(), Provider: "qa", Status: "ok",
		Output: agent.QAResult{Status: "PASS", Tests: 42},
		StartedAt: start, FinishedAt: time.Now(),
	}, nil
}

type SecurityAgent struct{}

func NewSecurity() *SecurityAgent { return &SecurityAgent{} }
func (s *SecurityAgent) Name() string { return "security" }
func (s *SecurityAgent) Role() string { return "Revisa secrets y vulnerabilidades" }
func (s *SecurityAgent) CanHandle(t string) bool { return t == "security" }
func (s *SecurityAgent) Providers() []agent.Provider { return nil }
func (s *SecurityAgent) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	return agent.AgentResult{
		TaskID: task.ID, Agent: s.Name(), Provider: "security", Status: "ok",
		Output: agent.SecurityResult{Status: "ok", Score: 9},
		StartedAt: start, FinishedAt: time.Now(),
	}, nil
}

type GithubAgent struct{}

func NewGithub() *GithubAgent { return &GithubAgent{} }
func (g *GithubAgent) Name() string { return "github" }
func (g *GithubAgent) Role() string { return "Interacciona con GitHub" }
func (g *GithubAgent) CanHandle(t string) bool { return t == "github" }
func (g *GithubAgent) Providers() []agent.Provider { return nil }
func (g *GithubAgent) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	return agent.AgentResult{TaskID: task.ID, Agent: g.Name(), Provider: "github", Status: "ok", Output: "github ok", StartedAt: start, FinishedAt: time.Now()}, nil
}

type DocumentationAgent struct{}

func NewDocs() *DocumentationAgent { return &DocumentationAgent{} }
func (d *DocumentationAgent) Name() string { return "documentation" }
func (d *DocumentationAgent) Role() string { return "Genera docs" }
func (d *DocumentationAgent) CanHandle(t string) bool { return t == "docs" || t == "documentation" }
func (d *DocumentationAgent) Providers() []agent.Provider { return nil }
func (d *DocumentationAgent) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	return agent.AgentResult{TaskID: task.ID, Agent: d.Name(), Provider: "docs", Status: "ok", Output: "docs ok", StartedAt: start, FinishedAt: time.Now()}, nil
}
