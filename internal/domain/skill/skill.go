package skill

// Skill represents a reusable capability.
type Skill struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Agents      []string `json:"agents"`
	Tools       []string `json:"tools"`
	Version     string   `json:"version"`
}

// Example: coding, debugging, architecture, research, code-review
var Builtin = []Skill{
	{Name: "coding", Description: "Write and refactor code", Agents: []string{"coder"}, Tools: []string{"filesystem", "shell"}},
	{Name: "debugging", Description: "Diagnose and fix bugs", Agents: []string{"coder", "researcher"}, Tools: []string{"filesystem", "shell", "git"}},
	{Name: "architecture", Description: "Design system architecture", Agents: []string{"planner"}, Tools: []string{"filesystem"}},
	{Name: "research", Description: "Research and synthesize knowledge", Agents: []string{"researcher"}, Tools: []string{"filesystem", "memory"}},
	{Name: "code-review", Description: "Review source code for bugs, security and maintainability", Agents: []string{"researcher", "coder"}, Tools: []string{"filesystem", "git"}},
}
