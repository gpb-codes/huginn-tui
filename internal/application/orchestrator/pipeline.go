// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
// Pipeline es el orquestador de ejecución planificada: el lado de Huginn de
// "Huginn piensa y coordina, los agentes ejecutan".
//
// Flujo: Planner.Plan -> niveles (etapas paralelas) -> Router.Route ->
// Agent.Execute -> Evaluator -> (reintento | replan único | puerta de aprobación).
// Reintentos y replanes están acotados; el pipeline nunca cicla para siempre.
package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"huginn/internal/application/evaluator"
	"huginn/internal/application/planner"
	"huginn/internal/application/router"
	"huginn/internal/domain/execution"
	"huginn/internal/domain/plan"
	"huginn/internal/domain/task"
)

// Tracer persiste registros de ejecución; lo implementa infrastructure/trace.
type Tracer interface {
	Append(execution.Record) error
}

// Approver resuelve las puertas human-in-the-loop; devolver false deniega la tarea.
// Un Approver nil deniega toda tarea protegida (default seguro).
type Approver interface {
	RequestApproval(ctx context.Context, a *execution.Approval) bool
}

// PermissionChecker aplica mínimo privilegio por agente; lo implementa infrastructure/security.
// Un checker nil permite la ejecución (modo abierto).
type PermissionChecker interface {
	Can(agentID, tool, permission string) bool
}

// Event es una notificación de ciclo de vida para las capas de presentación (TUI).
type Event struct {
	RunID   string
	TaskID  string
	AgentID string
	Type    string // planned, level, started, progress, retry, approved, denied, completed, failed, replan, done
	Message string
}

// EventHandler consume eventos del pipeline; debe ser rápido y no bloqueante.
type EventHandler func(Event)

// Pipeline conecta planificación, ruteo, evaluación y recuperación acotada.
type Pipeline struct {
	planner       *planner.Planner
	router        *router.Router
	evaluator     *evaluator.Evaluator
	retry         evaluator.RetryPolicy
	maxReplans    int
	maxConcurrent int
	taskTimeout   time.Duration
	tracer        Tracer
	approver      Approver
	permissions   PermissionChecker
	onEvent       EventHandler
}

// Options personaliza un Pipeline; los valores cero seleccionan defaults.
type Options struct {
	Retry         evaluator.RetryPolicy
	MaxReplans    int
	MaxConcurrent int
	TaskTimeout   time.Duration
	Tracer        Tracer
	Approver      Approver
	Permissions   PermissionChecker
	OnEvent       EventHandler
}

// NewPipeline construye un Pipeline; planner y router son obligatorios.
func NewPipeline(p *planner.Planner, r *router.Router, opts Options) *Pipeline {
	if opts.MaxReplans < 0 {
		opts.MaxReplans = 0
	}
	if opts.MaxReplans == 0 {
		opts.MaxReplans = 1
	}
	if opts.MaxConcurrent <= 0 {
		opts.MaxConcurrent = 4
	}
	if opts.TaskTimeout <= 0 {
		opts.TaskTimeout = 90 * time.Second
	}
	return &Pipeline{
		planner:       p,
		router:        r,
		evaluator:     evaluator.New(),
		retry:         opts.Retry.Normalize(),
		maxReplans:    opts.MaxReplans,
		maxConcurrent: opts.MaxConcurrent,
		taskTimeout:   opts.TaskTimeout,
		tracer:        opts.Tracer,
		approver:      opts.Approver,
		permissions:   opts.Permissions,
		onEvent:       opts.OnEvent,
	}
}

// Result resume una ejecución del pipeline.
type Result struct {
	Run      execution.Run
	Plan     plan.Plan
	Tasks    []*task.Task
	Verdicts map[string]evaluator.Verdict
	Success  bool
}

// Run planifica el objetivo y lo ejecuta hasta completarlo o fallar de forma acotada.
func (pl *Pipeline) Run(ctx context.Context, objective, workspace string) (Result, error) {
	if pl.planner == nil || pl.router == nil {
		return Result{}, fmt.Errorf("pipeline: planner and router are required")
	}
	p, err := pl.planner.Plan(objective, workspace)
	if err != nil {
		return Result{}, err
	}
	byID := make(map[string]*task.Task, len(p.Tasks))
	ids := make([]string, 0, len(p.Tasks))
	for i := range p.Tasks {
		t := &p.Tasks[i]
		if workspace != "" {
			t.Workspace = workspace
		}
		byID[t.ID] = t
		ids = append(ids, t.ID)
	}
	run := execution.NewRun(objective, p.ID, ids)
	res := Result{Run: run, Plan: p, Verdicts: make(map[string]evaluator.Verdict)}
	pl.emit(Event{RunID: run.ID, Type: "planned",
		Message: fmt.Sprintf("plan %s: %d tasks, %d levels (complexity %.2f, risk %.2f)",
			p.ID, len(p.Tasks), len(p.Levels()), p.Complexity, p.Risk)})

	pl.executeLevels(ctx, &run, p.Levels(), res.Verdicts)

	// Replanificación acotada: una ronda de recuperación para fallos reintentables.
	replans := 0
	for replans < pl.maxReplans {
		failed := collectFailed(byID, res.Verdicts, true)
		if len(failed) == 0 {
			break
		}
		replans++
		run.Attempts++
		rp := pl.planner.RecoveryPlan(objective, failed)
		pl.emit(Event{RunID: run.ID, Type: "replan",
			Message: fmt.Sprintf("replan %d/%d: %d failed tasks", replans, pl.maxReplans, len(failed))})
		// Integra las tareas de recuperación en el seguimiento.
		for i := range rp.Tasks {
			rp.Tasks[i].Workspace = workspace
			rt := rp.Tasks[i]
			cp := rt
			byID[rt.ID] = &cp
			run.TaskIDs = append(run.TaskIDs, rt.ID)
		}
		pl.executeLevels(ctx, &run, rp.Levels(), res.Verdicts)
	}

	for _, t := range byID {
		res.Tasks = append(res.Tasks, t)
	}
	res.Success = len(collectFailed(byID, res.Verdicts, false)) == 0
	if ctx.Err() != nil {
		run.Finish(task.StatusCancelled, ctx.Err().Error())
		res.Success = false
	} else if res.Success {
		run.Finish(task.StatusCompleted, "")
	} else {
		run.Finish(task.StatusFailed, "one or more tasks failed after bounded retries")
	}
	res.Run = run
	pl.emit(Event{RunID: run.ID, Type: "done",
		Message: fmt.Sprintf("run %s success=%v", run.ID, res.Success)})
	return res, nil
}

func collectFailed(byID map[string]*task.Task, verdicts map[string]evaluator.Verdict, onlyRetryable bool) []task.Task {
	var out []task.Task
	for id, t := range byID {
		v, ok := verdicts[id]
		if !ok || v.Success {
			continue
		}
		if onlyRetryable && !v.Retryable {
			continue
		}
		out = append(out, *t)
	}
	return out
}

func (pl *Pipeline) executeLevels(ctx context.Context, run *execution.Run, levels [][]*task.Task, verdicts map[string]evaluator.Verdict) {
	for li, level := range levels {
		if ctx.Err() != nil {
			return
		}
		pl.emit(Event{RunID: run.ID, Type: "level",
			Message: fmt.Sprintf("level %d/%d: %d tasks", li+1, len(levels), len(level))})
		sem := make(chan struct{}, pl.maxConcurrent)
		var wg sync.WaitGroup
		for _, t := range level {
			// Omite tareas con dependencias fallidas: no tiene sentido ejecutarlas.
			if !depsSucceeded(t, verdicts) {
				t.Status = task.StatusCancelled
				verdicts[t.ID] = evaluator.Verdict{Success: false, Reasons: []string{"skipped: dependency failed"}}
				pl.emit(Event{RunID: run.ID, TaskID: t.ID, Type: "failed", Message: "skipped: dependency failed"})
				continue
			}
			wg.Add(1)
			sem <- struct{}{}
			go func(t *task.Task) {
				defer wg.Done()
				defer func() { <-sem }()
				verdicts[t.ID] = pl.executeTask(ctx, run.ID, t)
			}(t)
		}
		wg.Wait()
	}
}

func depsSucceeded(t *task.Task, verdicts map[string]evaluator.Verdict) bool {
	for _, dep := range t.Dependencies {
		v, ok := verdicts[dep]
		if !ok || !v.Success {
			return false
		}
	}
	return true
}

func (pl *Pipeline) executeTask(ctx context.Context, runID string, t *task.Task) evaluator.Verdict {
	// Puerta human-in-the-loop previa a la ejecución.
	if gate := evaluator.ApprovalGate(*t); gate != nil {
		gate.RunID = runID
		allowed := false
		if pl.approver != nil {
			allowed = pl.approver.RequestApproval(ctx, gate)
		}
		if !allowed {
			t.Status = task.StatusCancelled
			pl.emit(Event{RunID: runID, TaskID: t.ID, Type: "denied",
				Message: "denied by approval gate: " + gate.Reason})
			return evaluator.Verdict{Success: false, Reasons: []string{"denied: " + gate.Reason}}
		}
		pl.emit(Event{RunID: runID, TaskID: t.ID, Type: "approved", Message: "approved by human"})
	}
	// Mínimo privilegio: los escritores necesitan permiso de escritura.
	if pl.permissions != nil && (t.Capability == "coding" || t.Capability == "devops") {
		agentID := t.AgentID
		if agentID == "" {
			agentID = "pipeline-agent"
		}
		if !pl.permissions.Can(agentID, "filesystem", "write") {
			t.Status = task.StatusFailed
			return evaluator.Verdict{Success: false,
				Reasons: []string{"denied by permission policy: no write grant for " + agentID}}
		}
	}
	ag, err := pl.router.Route(ctx, *t)
	if err != nil {
		t.Status = task.StatusFailed
		t.Result = task.FailedResult(err.Error())
		pl.trace(runID, t, "", "", err.Error())
		return evaluator.Verdict{Success: false, Retryable: false, Reasons: []string{err.Error()}}
	}
	t.AgentID = ag.ID()
	pl.emit(Event{RunID: runID, TaskID: t.ID, AgentID: ag.ID(), Type: "started",
		Message: fmt.Sprintf("%s -> %s", t.Title, ag.DisplayName())})

	var last evaluator.Verdict
	for attempt := 1; ; attempt++ {
		t.Status = task.StatusRunning
		if attempt > 1 {
			t.Status = task.StatusRetrying
			pl.emit(Event{RunID: runID, TaskID: t.ID, Type: "retry",
				Message: fmt.Sprintf("attempt %d", attempt)})
		}
		callCtx, cancel := context.WithTimeout(ctx, pl.taskTimeout)
		start := time.Now()
		res, execErr := ag.Execute(callCtx, *t)
		cancel()
		if execErr != nil && !res.Success && res.Error == "" {
			res.Error = execErr.Error()
			res.Success = false
		}
		if res.CompletedAt.IsZero() {
			res.Duration = time.Since(start)
		}
		t.Result = &res
		t.UpdatedAt = time.Now()
		last = pl.evaluator.Evaluate(*t, res)
		if last.Success {
			t.Status = task.StatusCompleted
			pl.trace(runID, t, ag.ID(), "", "")
			pl.emit(Event{RunID: runID, TaskID: t.ID, AgentID: ag.ID(), Type: "completed",
				Message: truncateOut(res.Output)})
			return last
		}
		t.Status = task.StatusFailed
		pl.trace(runID, t, ag.ID(), "", firstReason(last))
		pl.emit(Event{RunID: runID, TaskID: t.ID, AgentID: ag.ID(), Type: "failed",
			Message: firstReason(last)})
		if !pl.retry.ShouldRetry(*t, attempt, last) {
			return last
		}
		select {
		case <-ctx.Done():
			last.Reasons = append(last.Reasons, "cancelled during backoff")
			return last
		case <-time.After(pl.retry.Backoff(attempt)):
		}
	}
}

func (pl *Pipeline) trace(runID string, t *task.Task, agentID, provider, errMsg string) {
	if pl.tracer == nil {
		return
	}
	rec := execution.Record{
		ExecutionID: runID + ":" + t.ID,
		Agent:       agentID,
		Provider:    provider,
		TaskID:      t.ID,
		TaskType:    t.Title,
		Input:       t.Description,
		Status:      string(t.Status),
		StartedAt:   t.CreatedAt,
	}
	if errMsg != "" {
		rec.Errors = []string{errMsg}
	}
	_ = pl.tracer.Append(rec)
}

func (pl *Pipeline) emit(e Event) {
	if pl.onEvent != nil {
		pl.onEvent(e)
	}
}

func firstReason(v evaluator.Verdict) string {
	if len(v.Reasons) == 0 {
		return "failed"
	}
	return v.Reasons[0]
}

func truncateOut(s string) string {
	const n = 200
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
