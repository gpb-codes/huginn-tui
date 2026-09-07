// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package vault

import "context"

// Manager abstrae las operaciones sobre Vault.
// La infraestructura lo implementa sobre el sistema de ficheros.
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
