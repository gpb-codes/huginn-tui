package orchestrator

import (
	"fmt"

	"huginn/internal/domain/task"
)

// MockPlanner creates a deterministic task graph from a prompt.
// Real implementation will call an LLM via AgentRuntime.
type MockPlanner struct{}

func NewMockPlanner() *MockPlanner { return &MockPlanner{} }

func (p *MockPlanner) Plan(prompt, projectPath string) []task.Task {
	// Simple heuristic: split prompt into research → architecture → implementation → tests → review
	base := "TASK"
	return []task.Task{
		task.New(fmt.Sprintf("%s-001", base), "Research", "Research: "+prompt, "researcher"),
		task.New(fmt.Sprintf("%s-002", base), "Architecture", "Design architecture for: "+prompt, "planner", fmt.Sprintf("%s-001", base)),
		task.New(fmt.Sprintf("%s-003", base), "Implementation", "Implement: "+prompt, "coder", fmt.Sprintf("%s-002", base)),
		task.New(fmt.Sprintf("%s-004", base), "Tests", "Write tests for: "+prompt, "coder", fmt.Sprintf("%s-003", base)),
		task.New(fmt.Sprintf("%s-005", base), "Review", "Review implementation: "+prompt, "reviewer", fmt.Sprintf("%s-004", base)),
	}
}
