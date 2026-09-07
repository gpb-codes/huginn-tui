// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agents

import (
	"context"
	"time"

	"huginn/internal/application/router"
	domainagent "huginn/internal/domain/agent"
	"huginn/internal/domain/task"
)

// OpenCodeRuntime expone el adaptador CLI de OpenCode como runtime enrutable.
// Es el runtime principal de código para el pipeline.
type OpenCodeRuntime struct {
	adapter *OpenCodeAdapter
}

// NewOpenCodeRuntime construye el runtime con la configuración por defecto.
func NewOpenCodeRuntime() *OpenCodeRuntime {
	return &OpenCodeRuntime{adapter: NewOpenCodeAdapter()}
}

// NewOpenCodeRuntimeWithConfig construye el runtime con configuración explícita.
func NewOpenCodeRuntimeWithConfig(cfg Config) *OpenCodeRuntime {
	return &OpenCodeRuntime{adapter: NewOpenCodeAdapterWithConfig(cfg)}
}

func (r *OpenCodeRuntime) ID() string          { return "opencode" }
func (r *OpenCodeRuntime) DisplayName() string { return "OpenCode" }

func (r *OpenCodeRuntime) Capabilities() []domainagent.Capability {
	return []domainagent.Capability{
		domainagent.CapabilityCoding,
		domainagent.CapabilityTesting,
		domainagent.CapabilityDevOps,
		domainagent.CapabilityFilesystem,
		domainagent.CapabilityGit,
	}
}

func (r *OpenCodeRuntime) Available(_ context.Context) bool {
	ok, _ := r.adapter.Detect()
	return ok
}

// Execute ejecuta la tarea con OpenCode en el workspace de la tarea.
func (r *OpenCodeRuntime) Execute(ctx context.Context, t task.Task) (task.Result, error) {
	start := time.Now()
	res, err := r.adapter.Execute(ctx, domainagent.AgentTask{
		ID:    t.ID,
		Type:  t.Title,
		Input: t.Description,
		Context: domainagent.AgentContext{
			ProjectPath: t.Workspace,
		},
	})
	out := task.Result{Success: err == nil, CompletedAt: time.Now(), Duration: time.Since(start)}
	if code, ok := res.Output.(domainagent.CodeResult); ok {
		out.Output = code.Summary
	}
	if err != nil {
		out.Success = false
		out.Error = err.Error()
		if len(res.Errors) > 0 {
			out.Error = res.Errors[0]
		}
		if res.Output != nil {
			if code, ok := res.Output.(domainagent.CodeResult); ok && code.Summary != "" {
				out.Output = code.Summary
			}
		}
	}
	return out, err
}

// ModelRuntime adapta un proveedor de modelos como runtime para capacidades de texto.
// Cubre research, reasoning, documentación y review.
type ModelRuntime struct {
	id       string
	name     string
	provider domainagent.Provider
	caps     []domainagent.Capability
}

// NewModelRuntime construye un runtime sobre un proveedor con capacidades explícitas.
func NewModelRuntime(id, name string, p domainagent.Provider, caps ...domainagent.Capability) *ModelRuntime {
	return &ModelRuntime{id: id, name: name, provider: p, caps: caps}
}

func (r *ModelRuntime) ID() string          { return r.id }
func (r *ModelRuntime) DisplayName() string { return r.name }

func (r *ModelRuntime) Capabilities() []domainagent.Capability { return r.caps }

func (r *ModelRuntime) Available(ctx context.Context) bool {
	return r.provider.Available(ctx)
}

// Execute invoca el modelo y convierte la respuesta en resultado de tarea.
func (r *ModelRuntime) Execute(ctx context.Context, t task.Task) (task.Result, error) {
	start := time.Now()
	resp, err := r.provider.Invoke(ctx, domainagent.ProviderRequest{
		Prompt: t.Description,
		Context: domainagent.AgentContext{
			ProjectPath: t.Workspace,
		},
		Meta: map[string]string{"task_id": t.ID, "capability": string(t.Capability)},
	})
	out := task.Result{CompletedAt: time.Now(), Duration: time.Since(start)}
	if err != nil {
		out.Success = false
		out.Error = err.Error()
		return out, err
	}
	out.Success = true
	out.Output = resp.Content
	return out, nil
}

var _ router.Agent = (*OpenCodeRuntime)(nil)
var _ router.Agent = (*ModelRuntime)(nil)

// KiloRuntime es el runtime opcional de kilo; registra disponibilidad.
// La ejecución devuelve los errores honestos del adaptador.
type KiloRuntime struct{ adapter *KiloAdapter }

// NewKiloRuntime construye el runtime opcional de kilo.
func NewKiloRuntime() *KiloRuntime { return &KiloRuntime{adapter: NewKiloAdapter()} }

func (r *KiloRuntime) ID() string          { return "kilo" }
func (r *KiloRuntime) DisplayName() string { return "Kilo Code" }

func (r *KiloRuntime) Capabilities() []domainagent.Capability {
	return []domainagent.Capability{
		domainagent.CapabilityCoding,
		domainagent.CapabilityTesting,
		domainagent.CapabilityFilesystem,
		domainagent.CapabilityGit,
	}
}

func (r *KiloRuntime) Available(_ context.Context) bool {
	ok, _ := r.adapter.Detect()
	return ok
}

func (r *KiloRuntime) Execute(ctx context.Context, t task.Task) (task.Result, error) {
	start := time.Now()
	_, err := r.adapter.Execute(ctx, domainagent.AgentTask{
		ID:    t.ID,
		Type:  t.Title,
		Input: t.Description,
		Context: domainagent.AgentContext{
			ProjectPath: t.Workspace,
		},
	})
	out := task.Result{CompletedAt: time.Now(), Duration: time.Since(start)}
	if err != nil {
		out.Success = false
		out.Error = err.Error()
		return out, err
	}
	out.Success = true
	return out, nil
}

var _ router.Agent = (*KiloRuntime)(nil)
