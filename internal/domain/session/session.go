package session

import "time"

// Session represents a Huginn execution.
type Session struct {
	ID        string    `json:"id"`
	Project   string    `json:"project"`
	CreatedAt time.Time `json:"created_at"`
	Messages  []Message `json:"messages"`
	Tasks     []string  `json:"tasks"`
	Agents    []string  `json:"agents"`
	Decisions []string  `json:"decisions"`
	Context   string    `json:"context"`
}

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func New(id, project string) Session {
	return Session{ID: id, Project: project, CreatedAt: time.Now()}
}
