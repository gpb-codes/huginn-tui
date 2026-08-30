package orchestrator

import (
	"context"

	"huginn/internal/application/ports"
	"huginn/internal/domain/task"
)

// Executor bridges Dispatcher and AgentRuntime.
type Executor struct {
	runtime ports.AgentRuntime
}

func NewExecutor(rt ports.AgentRuntime) *Executor {
	return &Executor{runtime: rt}
}

func (e *Executor) Dispatch(ctx context.Context, t *task.Task) error {
	ch, err := e.runtime.Run(ctx, *t)
	if err != nil {
		t.Status = task.StatusFailed
		t.Result = task.FailedResult(err.Error())
		return err
	}
	t.Status = task.StatusRunning
	go func() {
		for range ch {
			// consume events; TUI will also consume via ports
		}
		t.Status = task.StatusCompleted
	}()
	return nil
}
