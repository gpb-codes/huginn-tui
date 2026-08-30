package ports

import "context"

// VaultPort is the hexagonal port for Agent Vault.
// Huginn talks to Vault only through this interface — no direct filesystem/DB code in domain/app.
type VaultPort interface {
	Path() string
	Exists(ctx context.Context) bool
	Search(ctx context.Context, query string) ([]string, error)
}

// MemoryPort abstracts persistent memory (Agent Vault).
type MemoryPort interface {
	Search(ctx context.Context, query string) ([]string, error)
	Save(ctx context.Context, text, project, importance string) error
}

// ToolPort abstracts MCP/LSP tools.
type ToolPort interface {
	Name() string
	Status() string
	Restart(ctx context.Context) error
}
