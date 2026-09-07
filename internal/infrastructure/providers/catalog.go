// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package providers

import "sync"

// Catalog guarda overrides de modelo (ventana/precio) sin requerir release.
type Catalog struct {
	mu        sync.RWMutex
	overrides map[string]map[string]any // model -> field -> value
}

func NewCatalog() *Catalog { return &Catalog{overrides: make(map[string]map[string]any)} }

func (c *Catalog) Set(model, field string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.overrides[model]; !ok {
		c.overrides[model] = make(map[string]any)
	}
	c.overrides[model][field] = value
}

func (c *Catalog) Get(model, field string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if m, ok := c.overrides[model]; ok {
		v, ok := m[field]
		return v, ok
	}
	return nil, false
}
