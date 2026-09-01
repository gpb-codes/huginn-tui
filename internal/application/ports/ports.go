package ports

import "context"

// VaultPort is the hexagonal port for Agent Vault.
// Huginn talks to Vault only through this interface — no direct filesystem/DB code in domain/app.
type VaultPort interface {
	Path() string
	Exists(ctx context.Context) bool
	Search(ctx context.Context, query string) ([]string, error)
}

// ToolPort abstracts MCP/LSP tools.
type ToolPort interface {
	Name() string
	Status() string
	Restart(ctx context.Context) error
}
