package personalization

import (
	"context"

	"huginn/internal/application/ports"
	"huginn/internal/domain/memory"
)

// Engine orchestrates Retriever + Learner + ContextBuilder.
type Engine struct {
	retriever Retriever
	learner   Learner
	builder   ContextBuilder
	store     ports.MemoryPort
}

func NewEngine(retriever Retriever, learner Learner, builder ContextBuilder, store ports.MemoryPort) *Engine {
	return &Engine{retriever: retriever, learner: learner, builder: builder, store: store}
}

type Retriever interface {
	Retrieve(ctx context.Context, query string, limit int) ([]memory.Memory, error)
}

type Learner interface {
	Learn(ctx context.Context, input string) (*memory.Memory, error)
}

type ContextBuilder interface {
	Build(ctx context.Context, prompt, project string) (string, error)
}

func (e *Engine) Search(ctx context.Context, query string) ([]memory.Memory, error) {
	return e.retriever.Retrieve(ctx, query, 10)
}

func (e *Engine) Learn(ctx context.Context, input string) (*memory.Memory, error) {
	return e.learner.Learn(ctx, input)
}

func (e *Engine) Context(ctx context.Context, prompt, project string) (string, error) {
	return e.builder.Build(ctx, prompt, project)
}
