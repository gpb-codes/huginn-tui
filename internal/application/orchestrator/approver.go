// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package orchestrator

import (
	"context"

	"huginn/internal/domain/execution"
)

// AllowApprover aprueba todas las puertas; úsalo solo con auto_approve explícito del operador.
// Equivale a la config permissions.auto_approve.
type AllowApprover struct{}

func (AllowApprover) RequestApproval(_ context.Context, _ *execution.Approval) bool {
	return true
}

// DenyApprover deniega todas las puertas (default seguro para ejecuciones desatendidas).
type DenyApprover struct{}

func (DenyApprover) RequestApproval(_ context.Context, _ *execution.Approval) bool {
	return false
}
