// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package memory

import "time"

// Constantes de tipo de memoria.
const (
	TypePreference = "preference"
	TypeFact       = "fact"
	TypeDecision   = "decision"
	TypeLesson     = "lesson"
	TypeProfile    = "profile"
)

// Memory es la entidad canónica de dominio. Markdown es la fuente de verdad.
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

// Entry es un alias histórico para compatibilidad con la TUI actual.
type Entry = Memory
