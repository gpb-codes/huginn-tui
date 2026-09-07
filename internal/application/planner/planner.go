// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
// Package planner es la etapa de planificación del motor neuronal Huginn.
//
// Responde QUÉ / POR QUÉ / CUÁNDO: analiza intención, clasifica tareas,
// estima complejidad, descompone en un DAG de dependencias y define la
// estrategia de evaluación. Nunca ejecuta nada ni importa runtimes de
// agentes: el ruteo se expresa como capacidades requeridas.
package planner

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	domainagent "huginn/internal/domain/agent"
	"huginn/internal/domain/plan"
	"huginn/internal/domain/task"
	"huginn/internal/neural"
	"huginn/internal/routing"
)

// Intent es el significado clasificado de un objetivo del usuario.
type Intent struct {
	Kind       string // implement, fix, research, review, document, devops, question
	Complexity float64
	Risk       float64
}

// Planner genera planes DAG validados a partir de objetivos.
type Planner struct {
	router *routing.NeuralRouter
}

// New entrena (semilla sintética) y devuelve un Planner.
// Si el entrenamiento falla, sigue con heurística pura y Trained() devuelve false.
func New() *Planner {
	p := &Planner{
		router: routing.NewNeuralRouter(
			[]string{"coding", "research", "general"},
			nil,
		),
	}
	_ = p.train(context.Background())
	return p
}

func (p *Planner) train(ctx context.Context) error {
	ds := routing.NewSyntheticDatasetFor(p.routerAgentIDs())
	_, err := p.router.Train(ctx, ds, defaultTrainConfig())
	return err
}

func (p *Planner) routerAgentIDs() []string {
	return []string{"coding", "research", "general"}
}

func defaultTrainConfig() neural.TrainConfig {
	return neural.TrainConfig{Epochs: 200, LearningRate: 0.05, PrintEvery: 0}
}

// Trained indica si el router neuronal terminó el entrenamiento.
func (p *Planner) Trained() bool { return p.router.IsTrained() }

// Analyze clasifica el objetivo sin producir tareas.
func (p *Planner) Analyze(objective string) Intent {
	low := strings.ToLower(objective)
	kind := "implement"
	switch {
	case hasAny(low, "investiga", "research", "analiza", "compara", "estudia", "busca"):
		kind = "research"
	case hasAny(low, "fix", "bug", "error", "corrige", "repara", "falla"):
		kind = "fix"
	case hasAny(low, "review", "revisa", "audita"):
		kind = "review"
	case hasAny(low, "documenta", "docs", "documentación", "readme"):
		kind = "document"
	case hasAny(low, "deploy", "despliega", "producción", "production", "migra", "migration", "pipeline") ||
		hasWord(low, "ci", "cd"):
		kind = "devops"
	case hasAny(low, "qué", "qué es", "what", "why", "por qué", "explica", "cómo funciona"):
		kind = "question"
	}
	fe := routing.NewFeatureExtractor()
	f := fe.Extract(objective, objective)
	return Intent{
		Kind:       kind,
		Complexity: f.Complexity,
		Risk:       estimateRisk(low, f),
	}
}

func estimateRisk(low string, f routing.TaskFeatures) float64 {
	risk := 0.15 + 0.3*f.TerminalRequired
	for _, kw := range []string{"producción", "production", "deploy", "push", "migration", "credential", "secret", "rm -rf", "delete", "drop"} {
		if strings.Contains(low, kw) {
			risk += 0.2
		}
	}
	if risk > 1 {
		risk = 1
	}
	return risk
}

// Plan descompone el objetivo en un DAG de tareas validado.
// Las preguntas simples colapsan a una tarea; el resto fluye por research -> diseño ->
// implementación (+ bifurcaciones paralelas según complejidad) -> tests -> review.
func (p *Planner) Plan(objective, workspace string) (plan.Plan, error) {
	intent := p.Analyze(objective)
	var tasks []task.Task
	switch intent.Kind {
	case "question", "research":
		tasks = []task.Task{
			capTask("task-1-research", "Research", objective, domainagent.CapabilityResearch),
			capTask("task-2-synthesize", "Synthesize", "Sintetizar hallazgos en respuesta clara: "+objective, domainagent.CapabilityReasoning, "task-1-research"),
		}
	case "review":
		tasks = []task.Task{
			capTask("task-1-analyze", "Analyze", "Analizar alcance para revisión: "+objective, domainagent.CapabilityResearch),
			capTask("task-2-review", "Review", objective, domainagent.CapabilityReview, "task-1-analyze"),
		}
	case "document":
		tasks = []task.Task{
			capTask("task-1-gather", "Gather", "Recolectar contexto del código: "+objective, domainagent.CapabilityResearch),
			capTask("task-2-write", "Write docs", objective, domainagent.CapabilityDocumentation, "task-1-gather"),
			capTask("task-3-review", "Review docs", "Revisar precisión de la documentación", domainagent.CapabilityReview, "task-2-write"),
		}
	case "devops":
		t := capTask("task-3-apply", "Apply change", objective, domainagent.CapabilityDevOps, "task-2-design")
		t.NeedsApproval = true
		tasks = []task.Task{
			capTask("task-1-analyze", "Analyze", "Analizar estado actual: "+objective, domainagent.CapabilityResearch),
			capTask("task-2-design", "Design", "Diseñar el cambio operativo: "+objective, domainagent.CapabilityReasoning, "task-1-analyze"),
			t,
			capTask("task-4-verify", "Verify", "Verificar el cambio aplicado", domainagent.CapabilityTesting, "task-3-apply"),
		}
	default: // implement, fix
		tasks = []task.Task{
			capTask("task-1-research", "Research", "Investigar contexto: "+objective, domainagent.CapabilityResearch),
			capTask("task-2-design", "Design", "Diseñar arquitectura del cambio: "+objective, domainagent.CapabilityReasoning, "task-1-research"),
		}
		impl := capTask("task-3-implement", "Implement", objective, domainagent.CapabilityCoding, "task-2-design")
		tasks = append(tasks, impl)
		if intent.Complexity >= 0.55 {
			// Bifurcación paralela: frentes de implementación independientes.
			frontA := capTask("task-3a-frontend", "Implement frontend", objective+" (frontend)", domainagent.CapabilityCoding, "task-2-design")
			frontB := capTask("task-3b-backend", "Implement backend", objective+" (backend)", domainagent.CapabilityCoding, "task-2-design")
			implDeps := []string{"task-3a-frontend", "task-3b-backend"}
			tasks = append(tasks, frontA, frontB)
			tasks = append(tasks, capTask("task-4-tests", "Tests", "Probar: "+objective, domainagent.CapabilityTesting, implDeps...))
			tasks = append(tasks, capTask("task-5-review", "Review", "Revisar: "+objective, domainagent.CapabilityReview, "task-4-tests"))
		} else {
			tasks = append(tasks, capTask("task-4-tests", "Tests", "Probar: "+objective, domainagent.CapabilityTesting, "task-3-implement"))
			tasks = append(tasks, capTask("task-5-review", "Review", "Revisar: "+objective, domainagent.CapabilityReview, "task-4-tests"))
		}
		if intent.Kind == "fix" {
			repro := capTask("task-0-repro", "Reproduce", "Reproducir el fallo: "+objective, domainagent.CapabilityTesting)
			tasks = append([]task.Task{repro}, tasks...)
			for i := range tasks {
				if tasks[i].ID == "task-1-research" {
					tasks[i].Dependencies = append(tasks[i].Dependencies, "task-0-repro")
				}
			}
		}
	}
	if intent.Risk >= 0.6 {
		for i := range tasks {
			if tasks[i].Capability == domainagent.CapabilityCoding ||
				tasks[i].Capability == domainagent.CapabilityDevOps {
				tasks[i].NeedsApproval = true
			}
		}
	}
	for i := range tasks {
		if workspace != "" {
			tasks[i].Workspace = workspace
		}
	}
	pl := plan.New(objective, intent.Complexity, intent.Risk, tasks)
	if err := pl.Validate(); err != nil {
		return plan.Plan{}, err
	}
	return pl, nil
}

// RecoveryPlan construye un plan mínimo de seguimiento para las tareas fallidas.
// Permite al pipeline replanificar una vez en vez de entrar en bucle.
func (p *Planner) RecoveryPlan(objective string, failed []task.Task) plan.Plan {
	tasks := make([]task.Task, 0, len(failed))
	for i, f := range failed {
		rt := task.NewWithCapability(
			fmt.Sprintf("retry-%d-%s", i+1, f.ID),
			"Recover: "+f.Title,
			"Reintento con contexto de fallo previo ("+firstLine(failureSummary(f))+"). Original: "+f.Description,
			f.Capability,
		)
		rt.MaxAttempts = 1
		tasks = append(tasks, rt)
	}
	if len(tasks) == 0 {
		return plan.New(objective+" (recovery: nothing to recover)", 0.1, 0.1, nil)
	}
	return plan.New(objective+" (recovery)", 0.5, 0.5, tasks)
}

func failureSummary(f task.Task) string {
	if f.Result != nil && f.Result.Error != "" {
		return f.Result.Error
	}
	return "sin detalle de error"
}

func firstLine(s string) string {
	if i := strings.Index(s, "\n"); i >= 0 {
		return s[:i]
	}
	return s
}

func capTask(id, title, desc string, cap domainagent.Capability, deps ...string) task.Task {
	t := task.NewWithCapability(id, title, desc, cap, deps...)
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt
	return t
}

func hasAny(s string, kws ...string) bool {
	for _, k := range kws {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

// hasWord solo coincide palabras completas (evita falsos positivos como "ci" en "autenticación").
func hasWord(s string, words ...string) bool {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	set := make(map[string]bool, len(fields))
	for _, f := range fields {
		set[f] = true
	}
	for _, w := range words {
		if set[w] {
			return true
		}
	}
	return false
}
