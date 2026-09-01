package task

import "time"

// Task represents real work to be done by an Agent.
// It is the core of the Task Graph.
type Task struct {
	ID           string
	Title        string
	Description  string
	AgentID      string
	Status       Status
	Dependencies []string
	Result       *Result
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// New creates a Task with pending status.
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
