// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agents

import (
	"context"
	"fmt"

	"huginn/internal/domain/agent"
)

// invokeFirst invoca al primer provider disponible.
// Si ninguno puede ejecutar, devuelve un error descriptivo: los agentes de rol nunca inventan resultados.
func invokeFirst(ctx context.Context, providers []agent.Provider, req agent.ProviderRequest) (agent.ProviderResponse, string, error) {
	for _, p := range providers {
		if p.Available(ctx) {
			resp, err := p.Invoke(ctx, req)
			if err != nil {
				return agent.ProviderResponse{}, p.Name(), err
			}
			return resp, p.Name(), nil
		}
	}
	names := make([]string, 0, len(providers))
	for _, p := range providers {
		names = append(names, p.Name())
	}
	return agent.ProviderResponse{}, "", fmt.Errorf("no model provider available (tried: %v) — is ollama running? is opencode installed?", names)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
