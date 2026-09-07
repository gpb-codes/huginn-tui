// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package profile

import (
	"os"
	"path/filepath"
	"time"

	"huginn/internal/domain/profile"
)

// Store gestiona el perfil en ~/.huginn/profile.md.
type Store struct {
	BaseDir string
}

func NewStore(baseDir string) *Store {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".huginn")
	}
	return &Store{BaseDir: baseDir}
}

func (s *Store) Path() string { return filepath.Join(s.BaseDir, "profile.md") }

func (s *Store) Load() (profile.Profile, error) {
	p := s.Path()
	if _, err := os.Stat(p); err != nil {
		// Devuelve el perfil por defecto si no existe el fichero.
		return profile.Default(), nil
	}
	// Retorna el perfil por defecto; el parseo de frontmatter queda pendiente.
	prof := profile.Default()
	prof.UpdatedAt = time.Now()
	return prof, nil
}

func (s *Store) Save(prof profile.Profile) error {
	if err := os.MkdirAll(s.BaseDir, 0755); err != nil {
		return err
	}
	prof.UpdatedAt = time.Now()
	content := `---
type: profile
version: 1
---

# User Profile

## Communication

- Language: ` + prof.Communication.Language + `
- Style: ` + prof.Communication.Style + `
- Verbosity: ` + prof.Communication.Verbosity + `
- Technical depth: ` + prof.Communication.TechnicalDepth + `

## Development

- Preferred languages:
`
	for _, l := range prof.Development.PreferredLanguages {
		content += "  - " + l + "\n"
	}
	content += "\n- Preferred editor:\n  - " + prof.Development.PreferredEditor + "\n"
	return os.WriteFile(s.Path(), []byte(content), 0644)
}
