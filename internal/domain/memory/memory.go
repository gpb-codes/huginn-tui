package memory

import "time"

// Type constants
const (
	TypePreference = "preference"
	TypeFact       = "fact"
	TypeDecision   = "decision"
	TypeLesson     = "lesson"
	TypeProfile    = "profile"
)

// Memory is the canonical domain entity. Markdown is source of truth.
type Memory struct {
	ID         string    `json:"id"`
	File       string    `json:"file"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Importance float64   `json:"importance"`
	Confidence float64   `json:"confidence"`
	Tags       []string  `json:"tags"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Entry is a legacy alias kept for backward compatibility with current TUI.
type Entry = Memory
