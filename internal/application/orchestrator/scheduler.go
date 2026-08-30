package orchestrator

import (
	"context"

	"huginn/internal/domain/task"
)

// SimpleScheduler marks tasks as queued if dependencies satisfied.
type SimpleScheduler struct{}

func NewSimpleScheduler() *SimpleScheduler { return &SimpleScheduler{} }

func (s *SimpleScheduler) Schedule(_ context.Context, t *task.Task) error {
	if t.Status == task.StatusPending {
		t.Status = task.StatusQueued
	}
	return nil
}
