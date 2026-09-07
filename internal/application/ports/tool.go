// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package ports

import "context"

// Tool representa una capacidad que un agente puede usar.
type Tool struct {
	Name        string
	Description string
	Permissions []string // read, write, execute, network, git
}

// ToolExecutor ejecuta una llamada a herramienta.
type ToolExecutor interface {
	Execute(ctx context.Context, tool string, args map[string]string) (string, error)
}

// PermissionPolicy decide si un agente puede usar una herramienta.
type PermissionPolicy interface {
	Can(agentID, tool string, permission string) bool
}

// ToolResult es el resultado de una ejecución de herramienta.
type ToolResult struct {
	Tool    string
	Success bool
	Output  string
	Error   string
}
