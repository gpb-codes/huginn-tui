package orchestrator

import (
	"fmt"
	"time"

	"huginn/internal/domain/task"
)

// DirectPlanner — crea una sola tarea directa (chat simple sin pipeline)
type DirectPlanner struct{}

func NewDirectPlanner() *DirectPlanner { return &DirectPlanner{} }

// PlanDirect — crea una tarea simple con agente resuelto por nombre
func (p *DirectPlanner) PlanDirect(prompt, agentName string) task.Task {
	return task.Task{
		ID:          fmt.Sprintf("chat-%d", time.Now().UnixMilli()),
		Title:       agentName,
		Description: prompt,
		AgentID:     agentName,
		Status:      task.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Plan implementa Planner interface — fallback al pipeline completo
func (p *DirectPlanner) Plan(prompt, projectPath string) []task.Task {
	return []task.Task{
		task.New("direct-1", "chat", prompt, "opencode"),
	}
}
