package filesystem

import (
	"os"
	"path/filepath"
)

// Service provides filesystem operations behind ToolPort.
// In Clean Architecture, domain never touches os directly; it goes through this adapter.
type Service struct {
	Root string
}

func New(root string) *Service { return &Service{Root: root} }

func (s *Service) Read(path string) ([]byte, error) {
	full := filepath.Join(s.Root, path)
	return os.ReadFile(full)
}

func (s *Service) Write(path string, data []byte) error {
	full := filepath.Join(s.Root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0644)
}

func (s *Service) List(dir string) ([]string, error) {
	full := filepath.Join(s.Root, dir)
	entries, err := os.ReadDir(full)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out, nil
}
