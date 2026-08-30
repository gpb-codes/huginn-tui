package ports

import "context"

// GitPort abstracts git operations. Agent must go through Tool, not direct git.
type GitPort interface {
	Status(ctx context.Context) (string, error)
	Diff(ctx context.Context) (string, error)
	Branch(ctx context.Context) (string, error)
	Commit(ctx context.Context, message string) (string, error)
}
