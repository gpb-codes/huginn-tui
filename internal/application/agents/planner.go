package agents

import (
	"context"
	"time"

	"huginn/internal/domain/agent"
)

type PlannerAgent struct{}

func NewPlanner() *PlannerAgent { return &PlannerAgent{} }
func (p *PlannerAgent) Name() string { return "planner" }
func (p *PlannerAgent) Role() string { return "Descompone objetivos en subtareas" }
func (p *PlannerAgent) CanHandle(t string) bool { return t == "plan" || t == "planner" }
func (p *PlannerAgent) Providers() []agent.Provider { return nil }
func (p *PlannerAgent) Execute(_ context.Context, task agent.AgentTask) (agent.AgentResult, error) {
	start := time.Now()
	// ejemplo: objetivo -> Research -> Coding -> QA -> Review
	return agent.AgentResult{
		TaskID: task.ID, Agent: p.Name(), Provider: "planner", Status: "ok",
		Output: []string{"Research", "Coding", "QA", "Review"},
		StartedAt: start, FinishedAt: time.Now(),
	}, nil
}
