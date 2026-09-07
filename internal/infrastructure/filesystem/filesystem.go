// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package filesystem

import (
	"os"
	"path/filepath"
)

// Service expone operaciones de ficheros tras ToolPort.
// El dominio nunca usa os directo; pasa por este adaptador.
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
