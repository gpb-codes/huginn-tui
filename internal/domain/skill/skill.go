package skill

// Skill represents a reusable capability — SKILL.md + assets (Hermes/Mimo style).
type Skill struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Agents       []string `json:"agents"`
	Tools        []string `json:"tools"`
	Version      string   `json:"version"`
	Environments []string `json:"environments,omitempty"` // kanban, s6, docker — empty = all
	Trust        string   `json:"trust,omitempty"`        // builtin, trusted, community
	Tags         []string `json:"tags,omitempty"`
	UsageCount   int      `json:"usage_count"`
	LastUsed     string   `json:"last_used,omitempty"`
}

// Example: coding, debugging, architecture, research, code-review
var Builtin = []Skill{
	{Name: "coding", Description: "Write and refactor code", Agents: []string{"coder"}, Tools: []string{"filesystem", "shell"}},
	{Name: "debugging", Description: "Diagnose and fix bugs", Agents: []string{"coder", "researcher"}, Tools: []string{"filesystem", "shell", "git"}},
	{Name: "architecture", Description: "Design system architecture", Agents: []string{"planner"}, Tools: []string{"filesystem"}},
	{Name: "research", Description: "Research and synthesize knowledge", Agents: []string{"researcher"}, Tools: []string{"filesystem", "memory"}},
	{Name: "code-review", Description: "Review source code for bugs, security and maintainability", Agents: []string{"researcher", "coder"}, Tools: []string{"filesystem", "git"}},
}
