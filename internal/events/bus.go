// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package events

import (
	"context"
	"sync"
	"time"
)

// Type tipa cada evento del sistema.
type Type string

const (
	TaskCreated       Type = "task.created"
	TaskStarted       Type = "task.started"
	TaskCompleted     Type = "task.completed"
	TaskFailed        Type = "task.failed"
	AgentStarted      Type = "agent.started"
	AgentStopped      Type = "agent.stopped"
	AgentMessage      Type = "agent.message"
	WorkflowStarted   Type = "workflow.started"
	WorkflowCompleted Type = "workflow.completed"
	MemoryUpdated     Type = "memory.updated"
)

// Event es el sobre tipado interno para observabilidad, logs y UI en tiempo real.
type Event struct {
	Type      Type           `json:"type"`
	TaskID    string         `json:"task_id,omitempty"`
	AgentID   string         `json:"agent_id,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	At        time.Time      `json:"at"`
}

// Handler consume eventos. Debe ser rápido y no bloqueante.
type Handler func(Event)

// Bus es un pub/sub en memoria, sin dependencias externas.
// Pensado para logs, TUI en tiempo real, métricas y debugging.
type Bus struct {
	mu       sync.RWMutex
	handlers []Handler
}

func New() *Bus { return &Bus{} }

func (b *Bus) Subscribe(h Handler) {
	b.mu.Lock()
	b.handlers = append(b.handlers, h)
	b.mu.Unlock()
}

func (b *Bus) Publish(ctx context.Context, e Event) {
	if e.At.IsZero() {
		e.At = time.Now()
	}
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers...)
	b.mu.RUnlock()
	for _, h := range handlers {
		select {
		case <-ctx.Done():
			return
		default:
			h(e)
		}
	}
}
