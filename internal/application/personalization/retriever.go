// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package personalization

import (
	"context"
	"sort"
	"strings"

	"huginn/internal/application/ports"
	"huginn/internal/domain/memory"
)

// SimpleRetriever busca por texto y ordena por importancia, sin embeddings.
type SimpleRetriever struct {
	store ports.MemoryPort
}

func NewSimpleRetriever(store ports.MemoryPort) *SimpleRetriever {
	return &SimpleRetriever{store: store}
}

func (r *SimpleRetriever) Retrieve(ctx context.Context, query string, limit int) ([]memory.Memory, error) {
	all, err := r.store.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	// Ordena por importancia*confianza y coincidencia textual.
	q := strings.ToLower(query)
	sort.Slice(all, func(i, j int) bool {
		si := score(all[i], q)
		sj := score(all[j], q)
		return si > sj
	})
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func score(m memory.Memory, q string) float64 {
	base := m.Importance*0.6 + m.Confidence*0.4
	// Bonifica coincidencias en título, contenido o etiquetas.
	if strings.Contains(strings.ToLower(m.Title), q) {
		base += 0.3
	}
	if strings.Contains(strings.ToLower(m.Content), q) {
		base += 0.2
	}
	for _, t := range m.Tags {
		if strings.Contains(strings.ToLower(t), q) {
			base += 0.1
			break
		}
	}
	return base
}
