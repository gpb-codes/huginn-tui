package context

import (
	"os"
	"path/filepath"
	"strings"
)

// Manager — lee Markdown directo sin embeddings/RAG (Fase 1)
type Manager struct {
	vaultPath string
}

func New(vaultPath string) *Manager { return &Manager{vaultPath: vaultPath} }

// Load lee los 5 archivos de memoria y retorna contexto concatenado para agente
func (m *Manager) Load() (string, error) {
	if m.vaultPath == "" {
		return "", nil
	}
	var parts []string
	for _, name := range []string{"profile.md", "development.md", "git.md", "ai.md", "workflow.md"} {
		p := filepath.Join(m.vaultPath, ".huginn", "user", name)
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		s := strings.TrimSpace(string(b))
		if s != "" {
			parts = append(parts, "## "+name+"\n"+s)
		}
	}
	if len(parts) == 0 {
		return "", nil
	}
	return strings.Join(parts, "\n\n---\n\n"), nil
}

// NeedsUpdate detecta divergencia simple (ej: git strategy cambió)
func (m *Manager) NeedsUpdate(currentGitStrategy string) (bool, string) {
	ctx, _ := m.Load()
	if ctx == "" {
		return false, ""
	}
	if currentGitStrategy != "" && !strings.Contains(strings.ToLower(ctx), strings.ToLower(currentGitStrategy)) {
		return true, "HUGINN detecto que tu configuracion actual utiliza una estrategia diferente de Git. ¿Actualizar memoria?"
	}
	return false, ""
}
