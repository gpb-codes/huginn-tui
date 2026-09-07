// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
// Package plan es la salida estructurada del planificador Neural Engine:
// un objetivo descompuesto en un DAG de tareas con dependencias.
package plan

import (
	"errors"
	"fmt"
	"time"

	"huginn/internal/domain/task"
)

// Plan es un plan de ejecución estructurado producido por el planificador.
// Las tareas forman un DAG vía Dependencies; Levels() expone las etapas paralelas.
type Plan struct {
	ID         string
	Objective  string
	Complexity float64 // 0..1
	Risk       float64 // 0..1
	Tasks      []task.Task
	CreatedAt  time.Time
}

// New crea un plan con ID generado.
func New(objective string, complexity, risk float64, tasks []task.Task) Plan {
	return Plan{
		ID:         fmt.Sprintf("plan-%d", time.Now().UnixMilli()),
		Objective:  objective,
		Complexity: clamp01(complexity),
		Risk:       clamp01(risk),
		Tasks:      tasks,
		CreatedAt:  time.Now(),
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Validate comprueba que las dependencias referencien tareas conocidas y que el grafo sea acíclico.
func (p Plan) Validate() error {
	ids := make(map[string]bool, len(p.Tasks))
	for _, t := range p.Tasks {
		if t.ID == "" {
			return errors.New("plan: task with empty ID")
		}
		if ids[t.ID] {
			return fmt.Errorf("plan: duplicate task ID %q", t.ID)
		}
		ids[t.ID] = true
	}
	for _, t := range p.Tasks {
		for _, dep := range t.Dependencies {
			if !ids[dep] {
				return fmt.Errorf("plan: task %q depends on unknown task %q", t.ID, dep)
			}
			if dep == t.ID {
				return fmt.Errorf("plan: task %q depends on itself", t.ID)
			}
		}
	}
	if len(p.Tasks) > 0 && len(p.TopologicalOrder()) != len(p.Tasks) {
		return errors.New("plan: dependency cycle detected")
	}
	return nil
}

// Graph adapta las tareas del plan a un task.Graph.
func (p Plan) Graph() *task.Graph {
	g := task.NewGraph()
	for _, t := range p.Tasks {
		g.Add(t)
	}
	return g
}

// TopologicalOrder devuelve las tareas en orden de dependencias.
func (p Plan) TopologicalOrder() []*task.Task {
	return p.Graph().TopologicalOrder()
}

// Levels agrupa las tareas en etapas paralelas: cada tarea del nivel N solo depende de niveles < N.
// Un plan lineal produce una tarea por nivel; las independientes comparten nivel y pueden ejecutarse a la vez.
func (p Plan) Levels() [][]*task.Task {
	g := p.Graph()
	remaining := make(map[string]*task.Task, len(g.Tasks))
	for id, t := range g.Tasks {
		remaining[id] = t
	}
	done := make(map[string]bool)
	var levels [][]*task.Task
	for len(remaining) > 0 {
		var level []*task.Task
		for id, t := range remaining {
			ready := true
			for _, dep := range t.Dependencies {
				if !done[dep] {
					ready = false
					break
				}
			}
			if ready {
				level = append(level, t)
				_ = id
			}
		}
		if len(level) == 0 {
			// Ciclo: vacía el resto en un nivel para reportarlo sin bloquearse. Validate() marca el ciclo.
			for _, t := range remaining {
				level = append(level, t)
			}
			levels = append(levels, level)
			break
		}
		for _, t := range level {
			done[t.ID] = true
			delete(remaining, t.ID)
		}
		levels = append(levels, level)
	}
	return levels
}
