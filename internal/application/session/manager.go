package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"huginn/internal/domain/session"
)

// Manager persists sessions to ~/.huginn/sessions/<id>.md (and .json for index)
type Manager struct {
	BaseDir string
}

func NewManager(baseDir string) *Manager {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".huginn", "sessions")
	}
	return &Manager{BaseDir: baseDir}
}

func (m *Manager) Save(s session.Session) error {
	if err := os.MkdirAll(m.BaseDir, 0755); err != nil {
		return err
	}
	name := s.CreatedAt.Format("2006-01-02T15-04-05") + ".md"
	if s.ID != "" {
		name = s.ID + ".md"
	}
	path := filepath.Join(m.BaseDir, name)
	content := fmt.Sprintf(`---
id: %s
project: %s
created: %s
---

# Session %s

Messages: %d
Tasks: %v
`, s.ID, s.Project, s.CreatedAt.Format(time.RFC3339), s.ID, len(s.Messages), s.Tasks)
	return os.WriteFile(path, []byte(content), 0644)
}

func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

func GenerateID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}
