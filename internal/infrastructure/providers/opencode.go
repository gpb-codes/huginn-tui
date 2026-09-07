// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package providers

import (
	"context"

	"huginn/internal/domain/agent"
	infraagents "huginn/internal/infrastructure/agents"
)

// OpenCodeRuntimeProvider expone el adaptador CLI real de OpenCode tras
// agent.Provider para que los roles ejecuten trabajo real en vez de texto simulado.
type OpenCodeRuntimeProvider struct {
	adapter *infraagents.OpenCodeAdapter
}

// NewOpenCodeRuntime construye el proveedor con la configuración por defecto.
func NewOpenCodeRuntime() *OpenCodeRuntimeProvider {
	return &OpenCodeRuntimeProvider{adapter: infraagents.NewOpenCodeAdapter()}
}

// NewOpenCodeRuntimeWithConfig construye el proveedor con configuración explícita.
func NewOpenCodeRuntimeWithConfig(cfg infraagents.Config) *OpenCodeRuntimeProvider {
	return &OpenCodeRuntimeProvider{adapter: infraagents.NewOpenCodeAdapterWithConfig(cfg)}
}

func (p *OpenCodeRuntimeProvider) Name() string { return "opencode" }

func (p *OpenCodeRuntimeProvider) Available(_ context.Context) bool {
	ok, _ := p.adapter.Detect()
	return ok
}

func (p *OpenCodeRuntimeProvider) Invoke(ctx context.Context, req agent.ProviderRequest) (agent.ProviderResponse, error) {
	res, err := p.adapter.Execute(ctx, agent.AgentTask{
		ID:      metaOr(req.Meta, "task_id", "chat"),
		Type:    metaOr(req.Meta, "task_type", "chat"),
		Input:   req.Prompt,
		Context: req.Context,
	})
	content := ""
	if out, ok := res.Output.(agent.CodeResult); ok {
		content = out.Summary
	}
	return agent.ProviderResponse{
		Content:  content,
		Provider: p.Name(),
		Latency:  res.FinishedAt.Sub(res.StartedAt),
	}, err
}

func metaOr(m map[string]string, k, def string) string {
	if m != nil {
		if v, ok := m[k]; ok && v != "" {
			return v
		}
	}
	return def
}
