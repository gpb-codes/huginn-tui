// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package personalization

import (
	"context"
	"fmt"
	"strings"
	"time"

	"huginn/internal/application/ports"
	"huginn/internal/domain/memory"
)

// ConservativeLearner solo aprende preferencias explícitas como "prefiero pnpm".
type ConservativeLearner struct {
	store ports.MemoryPort
}

func NewConservativeLearner(store ports.MemoryPort) *ConservativeLearner {
	return &ConservativeLearner{store: store}
}

func (l *ConservativeLearner) Learn(ctx context.Context, input string) (*memory.Memory, error) {
	lower := strings.ToLower(input)
	// Heurística: solo aprende si contiene "prefiero", "prefer" o "me gusta".
	if !strings.Contains(lower, "prefiero") && !strings.Contains(lower, "prefer") && !strings.Contains(lower, "me gusta") {
		return nil, nil
	}
	// Extrae un par clave=valor simple (p. ej. "prefiero pnpm" -> package_manager=pnpm).
	value := ""
	if strings.Contains(lower, "pnpm") {
		value = "pnpm"
	} else if strings.Contains(lower, "yarn") {
		value = "yarn"
	} else if strings.Contains(lower, "npm") {
		value = "npm"
	} else if strings.Contains(lower, "bun") {
		value = "bun"
	} else if strings.Contains(lower, "typescript") {
		value = "TypeScript"
	} else if strings.Contains(lower, "go") {
		value = "Go"
	}
	if value == "" {
		return nil, nil
	}
	m := memory.Memory{
		ID:         fmt.Sprintf("pref_%d", time.Now().UnixNano()),
		Type:       memory.TypePreference,
		Title:      "Preference: " + value,
		Content:    input,
		Importance: 0.8,
		Confidence: 0.9,
		Tags:       []string{"preference", "auto-learned"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := l.store.Save(ctx, m); err != nil {
		return nil, err
	}
	return &m, nil
}
