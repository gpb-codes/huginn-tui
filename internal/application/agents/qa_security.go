// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agents

import (
	"context"
	"time"

	"huginn/internal/domain/agent"
	"huginn/internal/infrastructure/security"
)

// QAAgent analiza QA mediante un provider de modelo.
// Sin provider devuelve un error honesto en vez de un PASS inventado.
type QAAgent struct{ providers []agent.Provider }

func NewQA(providers ...agent.Provider) *QAAgent { return &QAAgent{providers: providers} }
func (q *QAAgent) Name() string                  { return "qa" }
func (q *QAAgent) Role() string                  { return "Ejecuta tests/lint/build" }
func (q *QAAgent) CanHandle(t string) bool       { return t == "qa" || t == "test" }
func (q *QAAgent) Providers() []agent.Provider   { return q.providers }
func (q *QAAgent) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	resp, provName, err := invokeFirst(ctx, q.providers, agent.ProviderRequest{Prompt: task.Input, Context: task.Context})
	if err != nil {
		return agent.AgentResult{
			TaskID: task.ID, Agent: q.Name(), Provider: provName,
			Status: "error", Errors: []string{err.Error()},
			StartedAt: start, FinishedAt: time.Now(),
		}, err
	}
	return agent.AgentResult{
		TaskID: task.ID, Agent: q.Name(), Provider: provName, Status: "ok",
		Output: agent.QAResult{
			Status:          "WARN", // model judgement, unverified by execution
			Recommendations: []string{truncate(resp.Content, 800)},
		},
		StartedAt: start, FinishedAt: time.Now(),
	}, nil
}

// SecurityAgent combina el escaneo local determinista de secretos con la revisión por provider si hay uno configurado.
// El escaneo es heurístico y documentado como tal, no inferencia de modelo.
type SecurityAgent struct{ providers []agent.Provider }

func NewSecurity(providers ...agent.Provider) *SecurityAgent {
	return &SecurityAgent{providers: providers}
}
func (s *SecurityAgent) Name() string                { return "security" }
func (s *SecurityAgent) Role() string                { return "Revisa secrets y vulnerabilidades" }
func (s *SecurityAgent) CanHandle(t string) bool     { return t == "security" }
func (s *SecurityAgent) Providers() []agent.Provider { return s.providers }
func (s *SecurityAgent) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	res := ScanSecrets(task.Input)
	if len(s.providers) > 0 {
		if resp, provName, err := invokeFirst(ctx, s.providers, agent.ProviderRequest{Prompt: task.Input, Context: task.Context}); err == nil {
			res.Issues = append(res.Issues, "model review ("+provName+"): "+truncate(resp.Content, 500))
		}
	}
	return agent.AgentResult{
		TaskID: task.ID, Agent: s.Name(), Provider: "security", Status: "ok",
		Output:     res,
		StartedAt:  start,
		FinishedAt: time.Now(),
	}, nil
}

// ScanSecrets aplica la heurística determinista de secretos sobre el input.
// Escaneo local real, sin intervención de modelo.
func ScanSecrets(input string) agent.SecurityResult {
	found := security.DetectSecret(input)
	status := "ok"
	score := 9
	if found {
		status = "WARN"
		score = 4
	}
	issues := []string{}
	if found {
		issues = append(issues, "posible secreto en el input — no persistir en memoria/traza sin redactar")
	}
	return agent.SecurityResult{Status: status, Issues: issues, SecretsFound: found, Score: score}
}

// GithubAgent interactúa con GitHub mediante un provider.
// Sin provider devuelve un error honesto en vez de un "github ok" inventado.
type GithubAgent struct{ providers []agent.Provider }

func NewGithub(providers ...agent.Provider) *GithubAgent {
	return &GithubAgent{providers: providers}
}
func (g *GithubAgent) Name() string                { return "github" }
func (g *GithubAgent) Role() string                { return "Interacciona con GitHub" }
func (g *GithubAgent) CanHandle(t string) bool     { return t == "github" }
func (g *GithubAgent) Providers() []agent.Provider { return g.providers }
func (g *GithubAgent) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	resp, provName, err := invokeFirst(ctx, g.providers, agent.ProviderRequest{Prompt: task.Input, Context: task.Context})
	if err != nil {
		return agent.AgentResult{
			TaskID: task.ID, Agent: g.Name(), Provider: provName,
			Status: "error", Errors: []string{err.Error()},
			StartedAt: start, FinishedAt: time.Now(),
		}, err
	}
	return agent.AgentResult{TaskID: task.ID, Agent: g.Name(), Provider: provName, Status: "ok",
		Output: truncate(resp.Content, 1200), StartedAt: start, FinishedAt: time.Now()}, nil
}

// DocumentationAgent genera documentación mediante un provider o devuelve un error honesto.
type DocumentationAgent struct{ providers []agent.Provider }

func NewDocs(providers ...agent.Provider) *DocumentationAgent {
	return &DocumentationAgent{providers: providers}
}
func (d *DocumentationAgent) Name() string                { return "documentation" }
func (d *DocumentationAgent) Role() string                { return "Genera docs" }
func (d *DocumentationAgent) CanHandle(t string) bool     { return t == "docs" || t == "documentation" }
func (d *DocumentationAgent) Providers() []agent.Provider { return d.providers }
func (d *DocumentationAgent) Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	resp, provName, err := invokeFirst(ctx, d.providers, agent.ProviderRequest{Prompt: task.Input, Context: task.Context})
	if err != nil {
		return agent.AgentResult{
			TaskID: task.ID, Agent: d.Name(), Provider: provName,
			Status: "error", Errors: []string{err.Error()},
			StartedAt: start, FinishedAt: time.Now(),
		}, err
	}
	return agent.AgentResult{TaskID: task.ID, Agent: d.Name(), Provider: provName, Status: "ok",
		Output: truncate(resp.Content, 2000), StartedAt: start, FinishedAt: time.Now()}, nil
}
