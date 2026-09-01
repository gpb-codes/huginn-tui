package agents

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"huginn/internal/domain/agent"
)

// Agent — rol con responsabilidad (planner, research, coding, etc.)
type Agent interface {
	Name() string
	Role() string
	CanHandle(taskType string) bool
	Execute(ctx context.Context, task agent.AgentTask) (agent.AgentResult, error)
	Providers() []agent.Provider
}

// Registry — registro extensible de agentes y providers
type Registry struct {
	mu        sync.RWMutex
	agents    map[string]Agent
	providers map[string]agent.Provider
	config    map[string]agent.ProviderConfig // agent -> preferred provider name
}

func NewRegistry() *Registry {
	return &Registry{
		agents:    make(map[string]Agent),
		providers: make(map[string]agent.Provider),
		config:    make(map[string]agent.ProviderConfig),
	}
}

func (r *Registry) RegisterAgent(a Agent) error {
	if a == nil || a.Name() == "" {
		return fmt.Errorf("agent invalido")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.agents[a.Name()]; exists {
		return fmt.Errorf("agente ya registrado: %s", a.Name())
	}
	r.agents[a.Name()] = a
	return nil
}

func (r *Registry) RegisterProvider(p agent.Provider) error {
	if p == nil || p.Name() == "" {
		return fmt.Errorf("provider invalido")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[p.Name()]; exists {
		return fmt.Errorf("provider ya registrado: %s", p.Name())
	}
	r.providers[p.Name()] = p
	return nil
}

func (r *Registry) GetAgent(name string) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.agents[name]
	return a, ok
}

func (r *Registry) GetProvider(name string) (agent.Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

func (r *Registry) ListAgents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.agents))
	for k := range r.agents {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (r *Registry) SetPreferred(agentName, providerName string, cfg agent.ProviderConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cfg.Name = providerName
	r.config[agentName] = cfg
}

// SelectProvider — selecciona provider segun config, disponibilidad y prioridad
func (r *Registry) SelectProvider(ctx context.Context, agentName string, candidates []agent.Provider) (agent.Provider, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("sin candidates para %s", agentName)
	}
	r.mu.RLock()
	pref, hasPref := r.config[agentName]
	r.mu.RUnlock()

	// 1) preferido si esta disponible
	if hasPref {
		if p, ok := r.GetProvider(pref.Name); ok && p.Available(ctx) {
			for _, c := range candidates {
				if c.Name() == pref.Name {
					return p, nil
				}
			}
		}
	}
	// 2) primer disponible por prioridad
	sort.Slice(candidates, func(i, j int) bool {
		// usa config priority si existe
		pi, pj := 0, 0
		r.mu.RLock()
		if ca, ok := r.config[agentName+"::"+candidates[i].Name()]; ok {
			pi = ca.Priority
		}
		if ca, ok := r.config[agentName+"::"+candidates[j].Name()]; ok {
			pj = ca.Priority
		}
		r.mu.RUnlock()
		return pi < pj
	})
	for _, c := range candidates {
		if c.Available(ctx) {
			return c, nil
		}
	}
	// 3) fallback al primero aunque no disponible (provider hara error trazable)
	return candidates[0], nil
}

// Delegate — resuelve agente por tipo de tarea
func (r *Registry) ResolveAgent(taskType string) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.agents {
		if a.CanHandle(taskType) {
			return a, true
		}
	}
	return nil, false
}
