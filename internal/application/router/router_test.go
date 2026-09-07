// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package router

import (
	"context"
	"errors"
	"testing"

	domainagent "huginn/internal/domain/agent"
	"huginn/internal/domain/task"
)

type fakeAgent struct {
	id        string
	caps      []domainagent.Capability
	available bool
}

func (f *fakeAgent) ID() string          { return f.id }
func (f *fakeAgent) DisplayName() string { return f.id }
func (f *fakeAgent) Capabilities() []domainagent.Capability {
	return f.caps
}
func (f *fakeAgent) Available(_ context.Context) bool { return f.available }
func (f *fakeAgent) Execute(_ context.Context, t task.Task) (task.Result, error) {
	return *task.SuccessResult("ok"), nil
}

func TestRoute_ByCapability(t *testing.T) {
	r := New()
	r.Register(&fakeAgent{id: "coder", caps: []domainagent.Capability{domainagent.CapabilityCoding}, available: true})
	r.Register(&fakeAgent{id: "sage", caps: []domainagent.Capability{domainagent.CapabilityResearch}, available: true})
	got, err := r.Route(context.Background(), task.NewWithCapability("1", "T", "", domainagent.CapabilityResearch))
	if err != nil || got.ID() != "sage" {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestRoute_PrefersAvailable(t *testing.T) {
	r := New()
	r.Register(&fakeAgent{id: "offline-coder", caps: []domainagent.Capability{domainagent.CapabilityCoding}, available: false})
	r.Register(&fakeAgent{id: "general", caps: []domainagent.Capability{domainagent.CapabilityCoding, domainagent.CapabilityResearch}, available: true})
	got, err := r.Route(context.Background(), task.NewWithCapability("1", "T", "", domainagent.CapabilityCoding))
	if err != nil || got.ID() != "general" {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestRoute_FallbackWhenNoCapabilityMatch(t *testing.T) {
	r := New()
	r.Register(&fakeAgent{id: "general", caps: []domainagent.Capability{domainagent.CapabilityResearch}, available: true})
	got, err := r.Route(context.Background(), task.NewWithCapability("1", "T", "", domainagent.CapabilityBrowser))
	if err != nil || got.ID() != "general" {
		t.Fatalf("expected honest fallback, got %v %v", got, err)
	}
}

func TestRoute_NoAgents(t *testing.T) {
	r := New()
	_, err := r.Route(context.Background(), task.NewWithCapability("1", "T", "", domainagent.CapabilityCoding))
	if !errors.Is(err, ErrNoAgent) {
		t.Fatalf("expected ErrNoAgent, got %v", err)
	}
}

func TestRegister_ReplacesDuplicate(t *testing.T) {
	r := New()
	r.Register(&fakeAgent{id: "a", caps: []domainagent.Capability{domainagent.CapabilityCoding}, available: false})
	r.Register(&fakeAgent{id: "a", caps: []domainagent.Capability{domainagent.CapabilityCoding}, available: true})
	if len(r.Agents()) != 1 {
		t.Fatal("duplicate must replace")
	}
	if !r.Agents()[0].Available(context.Background()) {
		t.Fatal("replacement must win")
	}
}
