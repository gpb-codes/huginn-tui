// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package providers

import (
	"context"
	"time"

	"huginn/internal/domain/agent"
)

// MockProvider es un doble de prueba trazable solo para tests y desarrollo offline.
// Nunca va en producción: allí se usan proveedores reales (opencode, ollama).
type MockProvider struct {
	name string
}

// NewMock devuelve un doble de prueba con el nombre dado.
func NewMock(name string) *MockProvider { return &MockProvider{name: name} }

func (m *MockProvider) Name() string { return m.name }

func (m *MockProvider) Available(_ context.Context) bool { return true }

func (m *MockProvider) Invoke(_ context.Context, req agent.ProviderRequest) (agent.ProviderResponse, error) {
	start := time.Now()
	return agent.ProviderResponse{
		Content:  "[mock " + m.name + "] " + req.Prompt,
		Provider: m.name,
		Latency:  time.Since(start),
	}, nil
}
