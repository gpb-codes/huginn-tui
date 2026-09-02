package delegation

import (
	"context"
	"sync"
)

// Role controla si un hijo puede delegar más
type Role string

const (
	RoleLeaf         Role = "leaf"
	RoleOrchestrator Role = "orchestrator"
)

// Delegate — fork de subagente con contexto heredado (Claude fork mode)
type Delegate struct {
	mu            sync.Mutex
	maxDepth      int
	maxConcurrent int
	active        map[string]*Child
}

type Child struct {
	ID     string
	Role   Role
	Depth  int
	Status string // running, completed, stopped
	Result any
}

func New(maxDepth, maxConcurrent int) *Delegate {
	if maxDepth < 1 {
		maxDepth = 1
	}
	if maxConcurrent < 1 {
		maxConcurrent = 10
	}
	return &Delegate{maxDepth: maxDepth, maxConcurrent: maxConcurrent, active: make(map[string]*Child)}
}

func (d *Delegate) Spawn(ctx context.Context, id string, role Role, depth int) (*Child, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if depth > d.maxDepth || len(d.active) >= d.maxConcurrent {
		return nil, false
	}
	c := &Child{ID: id, Role: role, Depth: depth, Status: "running"}
	d.active[id] = c
	return c, true
}

func (d *Delegate) List() []*Child {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]*Child, 0, len(d.active))
	for _, c := range d.active {
		out = append(out, c)
	}
	return out
}

func (d *Delegate) SendMessage(id, msg string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if c, ok := d.active[id]; ok && c.Status == "running" {
		c.Result = msg
		return true
	}
	return false
}

func (d *Delegate) Stop(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if c, ok := d.active[id]; ok {
		c.Status = "stopped"
		return true
	}
	return false
}
