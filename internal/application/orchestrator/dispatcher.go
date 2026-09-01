package orchestrator

import (
	"context"

	"huginn/internal/domain/task"
)

// NoopDispatcher is a placeholder that marks tasks as running then completed.
type NoopDispatcher struct{}

func NewNoopDispatcher() *NoopDispatcher { return &NoopDispatcher{} }

func (d *NoopDispatcher) Dispatch(_ context.Context, t *task.Task) error {
	t.Status = task.StatusRunning
	// In real runtime, this would call AgentRuntime.Run
	t.Status = task.StatusCompleted
	t.Result = task.SuccessResult("dispatched (mock)")
	return nil
}
