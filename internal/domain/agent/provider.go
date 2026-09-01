package agent

import (
	"context"
	"time"
)

// ProviderConfig — configuracion estructurada por provider (JSON)
type ProviderConfig struct {
	Name    string         `json:"name"` // chatgpt, claude, gemini, ollama, perplexity, opencode, kilo
	Enabled bool           `json:"enabled"`
	Priority int           `json:"priority"` // menor = preferido
	Model   string         `json:"model,omitempty"`
	Extras  map[string]any `json:"extras,omitempty"`
}

// Provider — implementacion intercambiable (LLM o tool)
type Provider interface {
	Name() string
	Available(ctx context.Context) bool
	Invoke(ctx context.Context, req ProviderRequest) (ProviderResponse, error)
}

type ProviderRequest struct {
	Prompt  string            `json:"prompt"`
	Context AgentContext      `json:"context"`
	Meta    map[string]string `json:"meta,omitempty"`
}

type ProviderResponse struct {
	Content  string            `json:"content"`
	Meta     map[string]string `json:"meta,omitempty"`
	Latency  time.Duration     `json:"latency"`
	Provider string            `json:"provider"`
}

// AgentContext — contexto que recibe cada agente (del Vault/Memory)
type AgentContext struct {
	VaultPath   string            `json:"vault_path"`
	ProjectPath string            `json:"project_path"`
	Memory      []string          `json:"memory,omitempty"` // fragmentos markdown relevantes
	Profile     map[string]string `json:"profile,omitempty"`
	Extras      map[string]any    `json:"extras,omitempty"`
}

// AgentTask — tarea delegada por Orchestrator
type AgentTask struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"` // research, coding, review, qa, security, docs, memory, github
	Input       string            `json:"input"`
	Context     AgentContext      `json:"context"`
	Deps        []string          `json:"deps,omitempty"`
	Criteria    string            `json:"criteria,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// AgentResult — resultado estructurado trazable
type AgentResult struct {
	TaskID    string         `json:"task_id"`
	Agent     string         `json:"agent"`
	Provider  string         `json:"provider"`
	Status    string         `json:"status"` // ok, error, warn
	Output    any            `json:"output"` // ResearchResult, CodeResult, etc.
	Errors    []string       `json:"errors,omitempty"`
	StartedAt time.Time      `json:"started_at"`
	FinishedAt time.Time     `json:"finished_at"`
	Latency   time.Duration  `json:"latency"`
}

// Tipos de resultado especializados

type ResearchResult struct {
	Summary           string   `json:"summary"`
	Findings          []string `json:"findings"`
	Sources           []string `json:"sources"`
	Contradictions    []string `json:"contradictions,omitempty"`
	Confidence        float64  `json:"confidence"`
	RecommendedActions []string `json:"recommended_actions,omitempty"`
}

type CodeResult struct {
	FilesChanged []string `json:"files_changed"`
	Diff         string   `json:"diff,omitempty"`
	Summary      string   `json:"summary"`
}

type ReviewResult struct {
	Approved bool     `json:"approved"`
	Issues   []string `json:"issues"`
	Suggestions []string `json:"suggestions"`
	Score    int      `json:"score"` // 0-10
}

type QAResult struct {
	Status         string   `json:"status"` // PASS, FAIL, WARN
	Tests          int      `json:"tests"`
	Errors         []string `json:"errors,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

type SecurityResult struct {
	Status  string   `json:"status"`
	Issues  []string `json:"issues"`
	SecretsFound bool  `json:"secrets_found"`
	Score   int      `json:"score"`
}
