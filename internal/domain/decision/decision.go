package decision

import "time"

// Decision represents an ADR (Architecture Decision Record) stored as Markdown.
type Decision struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"` // proposed, accepted, rejected, superseded
	Tags         []string  `json:"tags"`
	Content      string    `json:"content"`
	Reason       string    `json:"reason"`
	Alternatives []string  `json:"alternatives"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func New(id, title, status string) Decision {
	now := time.Now()
	return Decision{
		ID:        id,
		Title:     title,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
