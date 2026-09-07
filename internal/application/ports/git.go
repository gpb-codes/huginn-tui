// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package ports

import "context"

// GitPort abstrae las operaciones git; el agente debe pasar por Tool, no usar git directo.
type GitPort interface {
	Status(ctx context.Context) (string, error)
	Diff(ctx context.Context) (string, error)
	Branch(ctx context.Context) (string, error)
	Commit(ctx context.Context, message string) (string, error)
}
