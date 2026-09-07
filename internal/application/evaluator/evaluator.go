// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
// Package evaluator juzga el resultado de cada tarea y decide entre reintentar, replanificar o detenerse.
//
// Tras cada ejecución evalúa el resultado; los reintentos y la recuperación están acotados y las
// tareas sensibles pasan por una puerta de aprobación previa a la ejecución.
package evaluator

import (
	"strings"
	"time"

	"huginn/internal/domain/execution"
	"huginn/internal/domain/task"
)

// Verdict es el juicio sobre el resultado de una tarea.
type Verdict struct {
	Success   bool
	Retryable bool
	Reasons   []string
}

// Evaluator aplica comprobaciones deterministas sin llamar a modelos:
// los criterios de éxito deben seguir siendo auditables.
type Evaluator struct{}

// New devuelve un Evaluator listo para usar.
func New() *Evaluator { return &Evaluator{} }

// Evaluate juzga el resultado de una tarea.
func (e *Evaluator) Evaluate(t task.Task, res task.Result) Verdict {
	v := Verdict{Success: true}
	fail := func(reason string, retryable bool) {
		v.Success = false
		v.Reasons = append(v.Reasons, reason)
		if retryable {
			v.Retryable = true
		}
	}
	if !res.Success {
		retryable := isRetryableError(res.Error)
		if res.Error == "" {
			fail("result marked failed without error detail", true)
		} else {
			fail("agent reported failure: "+truncate(res.Error, 300), retryable)
		}
	}
	if res.Success && strings.TrimSpace(res.Output) == "" && len(res.Artifacts) == 0 {
		fail("empty output and no artifacts", true)
	}
	if res.Error != "" && res.Success {
		// Los avisos se adjuntan sin suspender la tarea.
		v.Reasons = append(v.Reasons, "warning: "+truncate(res.Error, 200))
	}
	if res.Duration < 0 {
		fail("negative duration (clock skew or runtime bug)", false)
	}
	return v
}

func isRetryableError(msg string) bool {
	low := strings.ToLower(msg)
	for _, kw := range []string{
		"timeout", "deadline exceeded", "temporarily", "temporarily unavailable",
		"connection refused", "connection reset", "econn", "try again",
		"rate limit", "429", "503", "502", "504", "empty output",
		"not installed", "not found", "no such file",
	} {
		if strings.Contains(low, kw) {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// RetryPolicy acota los reintentos; los valores cero seleccionan defaults razonables.
type RetryPolicy struct {
	MaxAttempts int
	BackoffBase time.Duration
}

// DefaultRetryPolicy reintenta dos veces con backoff lineal.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{MaxAttempts: 2, BackoffBase: 2 * time.Second}
}

// Normalize completa con defaults los valores cero.
func (p RetryPolicy) Normalize() RetryPolicy {
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = 2
	}
	if p.BackoffBase <= 0 {
		p.BackoffBase = 2 * time.Second
	}
	return p
}

// ShouldRetry indica si se permite otro intento (attempt en base 1: 1 es la primera ejecución).
// MaxAttempts de la tarea prevalece sobre la política.
func (p RetryPolicy) ShouldRetry(t task.Task, attempt int, v Verdict) bool {
	if v.Success || !v.Retryable {
		return false
	}
	max := p.Normalize().MaxAttempts
	if t.MaxAttempts > 0 {
		max = t.MaxAttempts
	}
	return attempt < max
}

// Backoff devuelve la espera previa al intento indicado (base 1).
func (p RetryPolicy) Backoff(attempt int) time.Duration {
	return time.Duration(attempt) * p.Normalize().BackoffBase
}

// ApprovalGate adapta execution.RequiresApproval para uso del pipeline.
func ApprovalGate(t task.Task) *execution.Approval {
	if !execution.RequiresApproval(t) {
		return nil
	}
	return &execution.Approval{
		TaskID:    t.ID,
		AgentID:   t.AgentID,
		Operation: string(t.Capability),
		Reason:    "sensitive operation heuristic matched (see execution.SensitiveKeywords)",
		CreatedAt: time.Now(),
	}
}
