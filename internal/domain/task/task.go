// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package task

import (
	"time"

	"huginn/internal/domain/agent"
)

// Task representa trabajo real pendiente de ejecución por un Agent.
// Es el núcleo del grafo de tareas.
type Task struct {
	ID           string
	Title        string
	Description  string
	AgentID      string
	Capability   agent.Capability
	Status       Status
	Dependencies []string
	// NeedsApproval exige aprobación humana antes de ejecutar.
	NeedsApproval bool
	// MaxAttempts redefine la política de reintentos del pipeline (0 = valor por defecto).
	MaxAttempts int
	// Workspace es el directorio donde ejecutan los runtimes (opencode --dir, …).
	Workspace string
	Result    *Result
	CreatedAt time.Time
	UpdatedAt time.Time
}

// New crea una Task en estado pendiente.
func New(id, title, description, agentID string, deps ...string) Task {
	now := time.Now()
	return Task{
		ID:           id,
		Title:        title,
		Description:  description,
		AgentID:      agentID,
		Status:       StatusPending,
		Dependencies: deps,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewWithCapability crea una Task enrutada por capacidad en lugar de por nombre de agente.
func NewWithCapability(id, title, description string, cap agent.Capability, deps ...string) Task {
	t := New(id, title, description, "", deps...)
	t.Capability = cap
	t.Status = StatusReady
	return t
}
