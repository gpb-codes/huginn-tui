// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package task

// Status representa el ciclo de vida de una Task.
type Status string

const (
	StatusPending   Status = "pending"
	StatusPlanning  Status = "planning"
	StatusReady     Status = "ready"
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusWaiting   Status = "waiting"
	StatusReview    Status = "review"
	StatusRetrying  Status = "retrying"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

func (s Status) IsTerminal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCancelled
}

func (s Status) IsActive() bool {
	return s == StatusRunning || s == StatusWaiting || s == StatusQueued
}
