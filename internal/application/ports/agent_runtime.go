// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package ports

import (
	"context"

	"huginn/internal/domain/task"
)

// AgentEvent representa un evento de streaming emitido por un agente.
type AgentEvent struct {
	TaskID    string    `json:"task_id"`
	AgentID   string    `json:"agent_id"`
	Type      EventType `json:"type"`
	Content   string    `json:"content"`
	Timestamp int64     `json:"timestamp"`
}

type EventType string

const (
	EventAgentStarted    EventType = "agent.started"
	EventAgentThinking   EventType = "agent.thinking"
	EventAgentToolCall   EventType = "agent.tool_call"
	EventAgentToolResult EventType = "agent.tool_result"
	EventAgentMessage    EventType = "agent.message"
	EventAgentProgress   EventType = "agent.progress"
	EventAgentCompleted  EventType = "agent.completed"
	EventAgentFailed     EventType = "agent.failed"

	EventTaskCreated   EventType = "task.created"
	EventTaskStarted   EventType = "task.started"
	EventTaskCompleted EventType = "task.completed"
	EventTaskFailed    EventType = "task.failed"

	EventMemoryCreated EventType = "memory.created"
	EventMemoryUpdated EventType = "memory.updated"

	EventSessionStarted   EventType = "session.started"
	EventSessionCompleted EventType = "session.completed"
)

// AgentRuntime abstrae la ejecución de una tarea por un agente.
// Implementaciones: process (os/exec), opencode, claude, ollama, etc.
type AgentRuntime interface {
	Run(ctx context.Context, task task.Task) (<-chan AgentEvent, error)
}
