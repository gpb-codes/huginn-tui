// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
// Package router elige qué runtime de agente ejecuta cada tarea.
//
// El ruteo es por capacidades: el planner declara lo que la tarea necesita
// (task.Capability) y el router elige el runtime disponible con mejor solape.
// Ningún `if agent == "opencode"` vive en el núcleo de Huginn;
// los runtimes autodeclaran lo que soportan.
package router

import (
	"context"
	"errors"
	"sort"

	domainagent "huginn/internal/domain/agent"
	"huginn/internal/domain/task"
)

// ErrNoAgent se devuelve cuando ningún runtime registrado puede atender la tarea.
var ErrNoAgent = errors.New("router: no agent available for task")

// Agent es un runtime capaz de ejecutar tareas; lo implementan los adaptadores de infraestructura.
// (OpenCode, roles sobre Ollama, futuros Goose/Aider/…).
type Agent interface {
	ID() string
	DisplayName() string
	Capabilities() []domainagent.Capability
	Available(ctx context.Context) bool
	Execute(ctx context.Context, t task.Task) (task.Result, error)
}

// Supports indica si un conjunto de capacidades incluye la capacidad pedida.
func Supports(caps []domainagent.Capability, want domainagent.Capability) bool {
	return domainagent.HasCapabilities(caps, want)
}

// Router mantiene el registro de runtimes y la estrategia de selección.
type Router struct {
	agents []Agent
}

// New devuelve un router vacío; los runtimes se añaden con Register.
func New() *Router { return &Router{} }

// Register añade un runtime; los IDs duplicados reemplazan la entrada previa.
// Así bootstrap puede redefinir defaults sin tocar el planner.
func (r *Router) Register(a Agent) {
	for i, existing := range r.agents {
		if existing.ID() == a.ID() {
			r.agents[i] = a
			return
		}
	}
	r.agents = append(r.agents, a)
}

// Agents devuelve los runtimes registrados en orden de registro.
func (r *Router) Agents() []Agent {
	out := make([]Agent, len(r.agents))
	copy(out, r.agents)
	return out
}

// Route elige el mejor runtime disponible para una tarea:
//  1. compatibles con la capacidad requerida, disponibles primero;
//  2. mayor solape de capacidades (el más especializado gana empates);
//  3. si ninguno coincide, cualquier runtime disponible como reserva;
//  4. si no, ErrNoAgent.
func (r *Router) Route(ctx context.Context, t task.Task) (Agent, error) {
	if len(r.agents) == 0 {
		return nil, ErrNoAgent
	}
	type scored struct {
		a         Agent
		available bool
		overlap   int
		total     int
	}
	var candidates []scored
	for _, a := range r.agents {
		caps := a.Capabilities()
		overlap := 0
		for _, c := range caps {
			if t.Capability != "" && c == t.Capability {
				overlap++
			}
		}
		if t.Capability == "" || overlap > 0 {
			candidates = append(candidates, scored{a, a.Available(ctx), overlap, len(caps)})
		}
	}
	if len(candidates) == 0 {
		// Reserva: cualquier runtime disponible.
		for _, a := range r.agents {
			if a.Available(ctx) {
				return a, nil
			}
		}
		return nil, ErrNoAgent
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].available != candidates[j].available {
			return candidates[i].available
		}
		if candidates[i].overlap != candidates[j].overlap {
			return candidates[i].overlap > candidates[j].overlap
		}
		return candidates[i].total < candidates[j].total
	})
	if !candidates[0].available {
		// La mejor coincidencia está offline: se devuelve igual para que Execute reporte un error honesto.
		return candidates[0].a, nil
	}
	return candidates[0].a, nil
}
