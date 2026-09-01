package orchestrator

import "context"

import "huginn/internal/domain/task"

// NoopSynthesizer does nothing (future: LLM synthesis).
type NoopSynthesizer struct{}

func NewNoopSynthesizer() *NoopSynthesizer { return &NoopSynthesizer{} }

func (s *NoopSynthesizer) Synthesize(_ context.Context, _ []*task.Task) error { return nil }
