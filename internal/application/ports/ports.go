// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package ports

import "context"

// VaultPort es el puerto hexagonal del Vault de agentes.
// Huginn solo habla con el Vault por esta interfaz, sin ficheros ni DB directos en dominio/app.
type VaultPort interface {
	Path() string
	Exists(ctx context.Context) bool
	Search(ctx context.Context, query string) ([]string, error)
}

// ToolPort abstrae las herramientas MCP/LSP.
type ToolPort interface {
	Name() string
	Status() string
	Restart(ctx context.Context) error
}
