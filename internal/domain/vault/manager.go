package vault

import "context"

// Manager is the abstraction for Vault operations.
// Infrastructure implements this via filesystem.
type Manager interface {
	Open(ctx context.Context, path string) (*Vault, error)
	Create(ctx context.Context, parentDir, name string) (*Vault, error)
	Close(ctx context.Context) error
	Initialize(ctx context.Context, path string) (*Vault, error)
	Exists(path string) bool
	IsInitialized(path string) bool
	GetCurrent() (*Vault, bool)
	GetPath() string
	Detect(startPath string) (*Vault, bool)
	Recent() ([]Vault, error)
	AddRecent(vault Vault) error
}
