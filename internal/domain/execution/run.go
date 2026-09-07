// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package execution

import (
	"strings"
	"time"

	"huginn/internal/domain/task"
)

// Run registra una ejecución planificada completa: plan -> tareas -> agentes -> herramientas -> eventos -> errores -> resultado.
// Cada ejecución tiene ID para alimentar depuración, analítica, memoria y aprendizaje.
type Run struct {
	ID         string
	Objective  string
	PlanID     string
	Status     task.Status
	TaskIDs    []string
	Attempts   int
	StartedAt  time.Time
	FinishedAt time.Time
	Error      string
}

// NewRun crea una ejecución en estado de planificación.
func NewRun(objective, planID string, taskIDs []string) Run {
	return Run{
		ID:        "run-" + time.Now().Format("20060102-150405.000"),
		Objective: objective,
		PlanID:    planID,
		Status:    task.StatusPlanning,
		TaskIDs:   taskIDs,
		StartedAt: time.Now(),
	}
}

// Finish marca la ejecución como terminal.
func (r *Run) Finish(status task.Status, errMsg string) {
	r.Status = status
	r.Error = errMsg
	r.FinishedAt = time.Now()
}

// Approval es una compuerta con humano en el bucle: la ejecución se pausa hasta que un humano
// aprueba o rechaza la operación sensible.
type Approval struct {
	RunID     string
	TaskID    string
	AgentID   string
	Operation string
	Reason    string
	Approved  *bool
	CreatedAt time.Time
}

// SensitiveKeywords se compara (sin distinguir mayúsculas) con título y descripción como heurística.
// Los runtimes con metadatos de permiso explícitos deben fijar Task.NeedsApproval en lugar de usar esta lista.
var SensitiveKeywords = []string{
	"git push",
	"git reset --hard",
	"rm -rf",
	"del /s",
	"format",
	"drop table",
	"drop database",
	"migration",
	"production",
	"deploy",
	"publish",
	"credential",
	"secret",
	"api_key",
	"token",
	"password",
}

// RequiresApproval indica si una tarea debe pausarse para aprobación humana.
func RequiresApproval(t task.Task) bool {
	if t.NeedsApproval {
		return true
	}
	haystack := strings.ToLower(t.Title + "\n" + t.Description)
	for _, kw := range SensitiveKeywords {
		if strings.Contains(haystack, kw) {
			return true
		}
	}
	return false
}
