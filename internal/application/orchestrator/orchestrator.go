package orchestrator

import (
	"context"

	"huginn/internal/domain/task"
)

// Orchestrator coordinates Planner → Scheduler → Runtime → Synthesizer.
type Orchestrator struct {
	planner     Planner
	scheduler   Scheduler
	dispatcher  Dispatcher
	synthesizer Synthesizer
}

func New(planner Planner, scheduler Scheduler, dispatcher Dispatcher, synthesizer Synthesizer) *Orchestrator {
	return &Orchestrator{planner: planner, scheduler: scheduler, dispatcher: dispatcher, synthesizer: synthesizer}
}

// Run plans tasks from a prompt and executes them.
func (o *Orchestrator) Run(ctx context.Context, prompt, projectPath string) ([]*task.Task, error) {
	tasks := o.planner.Plan(prompt, projectPath)
	graph := task.NewGraph()
	for _, t := range tasks {
		graph.Add(t)
	}
	ordered := graph.TopologicalOrder()
	for _, t := range ordered {
		if err := o.scheduler.Schedule(ctx, t); err != nil {
			return nil, err
		}
		if err := o.dispatcher.Dispatch(ctx, t); err != nil {
			return nil, err
		}
	}
	if err := o.synthesizer.Synthesize(ctx, ordered); err != nil {
		return ordered, err
	}
	return ordered, nil
}

type Planner interface {
	Plan(prompt, projectPath string) []task.Task
}

type Scheduler interface {
	Schedule(ctx context.Context, task *task.Task) error
}

type Dispatcher interface {
	Dispatch(ctx context.Context, task *task.Task) error
}

type Synthesizer interface {
	Synthesize(ctx context.Context, tasks []*task.Task) error
}
