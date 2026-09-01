package ports

import (
	"context"

	"huginn/internal/domain/memory"
)

// MemoryPort is the canonical port for memory (Markdown + JSONL).
type MemoryPort interface {
	Get(ctx context.Context, id string) (*memory.Memory, error)
	List(ctx context.Context, memoryType string) ([]memory.Memory, error)
	Search(ctx context.Context, query string) ([]memory.Memory, error)
	Save(ctx context.Context, m memory.Memory) error
	Delete(ctx context.Context, id string) error
}
