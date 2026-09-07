// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package skill

// Skill representa una capacidad reutilizable: SKILL.md más recursos (estilo Hermes/Mimo).
type Skill struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Agents       []string `json:"agents"`
	Tools        []string `json:"tools"`
	Version      string   `json:"version"`
	Environments []string `json:"environments,omitempty"` // kanban, s6, docker; vacío = todos
	Trust        string   `json:"trust,omitempty"`        // builtin, trusted o community
	Tags         []string `json:"tags,omitempty"`
	UsageCount   int      `json:"usage_count"`
	LastUsed     string   `json:"last_used,omitempty"`
}

// Builtin agrupa skills de ejemplo: coding, debugging, architecture, research y code-review.
var Builtin = []Skill{
	{Name: "coding", Description: "Write and refactor code", Agents: []string{"coder"}, Tools: []string{"filesystem", "shell"}},
	{Name: "debugging", Description: "Diagnose and fix bugs", Agents: []string{"coder", "researcher"}, Tools: []string{"filesystem", "shell", "git"}},
	{Name: "architecture", Description: "Design system architecture", Agents: []string{"planner"}, Tools: []string{"filesystem"}},
	{Name: "research", Description: "Research and synthesize knowledge", Agents: []string{"researcher"}, Tools: []string{"filesystem", "memory"}},
	{Name: "code-review", Description: "Review source code for bugs, security and maintainability", Agents: []string{"researcher", "coder"}, Tools: []string{"filesystem", "git"}},
}
