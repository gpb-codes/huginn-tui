package orchestrator

import (
	"context"
	"fmt"
	"time"

	domainctx "huginn/internal/domain/context"
	"huginn/internal/domain/execution"
	"huginn/internal/domain/task"
	"huginn/internal/infrastructure/trace"

	appagents "huginn/internal/application/agents"
	agentpkg "huginn/internal/domain/agent"
)

// RegistryDispatcher — dispatch real via Registry + ContextManager + trace
type RegistryDispatcher struct {
	registry *appagents.Registry
	ctxMgr   *domainctx.Manager
	trace    *trace.Store
}

func NewRegistryDispatcher(r *appagents.Registry, ctxMgr *domainctx.Manager, t *trace.Store) *RegistryDispatcher {
	return &RegistryDispatcher{registry: r, ctxMgr: ctxMgr, trace: t}
}

func (d *RegistryDispatcher) Dispatch(ctx context.Context, t *task.Task) error {
	start := time.Now()
	t.Status = task.StatusRunning

	agentName := t.AgentID
	a, ok := d.registry.GetAgent(agentName)
	if !ok {
		// intenta resolver por tipo de tarea
		if a2, ok2 := d.registry.ResolveAgent(t.Title); ok2 {
			a = a2
			agentName = a.Name()
		} else {
			t.Status = task.StatusFailed
			d.appendTrace(t, "", "", start, "agent not found: "+t.AgentID)
			return fmt.Errorf("agent not found: %s", t.AgentID)
		}
	}

	// 1) seleccionar provider
	provider, err := d.registry.SelectProvider(ctx, agentName, a.Providers())
	if err != nil {
		t.Status = task.StatusFailed
		d.appendTrace(t, agentName, "", start, err.Error())
		return err
	}

	// 2) cargar contexto del vault
	var memCtx string
	if d.ctxMgr != nil {
		memCtx, _ = d.ctxMgr.Load()
	}

	req := agentpkg.AgentContext{
		VaultPath:   d.trace.VaultPath(),
		ProjectPath: t.Description,
		Memory:      []string{memCtx},
	}

	// 3) invocar provider
	result, err := provider.Invoke(ctx, agentpkg.ProviderRequest{
		Prompt:  t.Description,
		Context: req,
		Meta:    map[string]string{"task_id": t.ID, "task_type": t.Title},
	})

	// 4) trazabilidad
	d.appendTrace(t, agentName, provider.Name(), start, errMessage(err))

	if err != nil {
		t.Status = task.StatusFailed
		return err
	}

	// 5) resultado
	t.Status = task.StatusCompleted
	t.Result = task.SuccessResult(result.Content)
	return nil
}

func (d *RegistryDispatcher) appendTrace(t *task.Task, agentName, providerName string, start time.Time, errMsg string) {
	if d.trace == nil {
		return
	}
	rec := execution.Record{
		ExecutionID: fmt.Sprintf("exec-%s-%d", t.ID, start.UnixMilli()),
		Agent:       agentName,
		Provider:    providerName,
		TaskID:      t.ID,
		TaskType:    t.Title,
		Input:       t.Description,
		Status:      string(t.Status),
		StartedAt:   start,
	}
	if errMsg != "" {
		rec.Errors = []string{errMsg}
	}
	_ = d.trace.Append(rec)
}

func errMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
