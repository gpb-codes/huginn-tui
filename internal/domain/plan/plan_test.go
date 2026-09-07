// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package plan

import (
	"testing"

	domainagent "huginn/internal/domain/agent"
	"huginn/internal/domain/task"
)

func TestPlan_Validate_OK(t *testing.T) {
	p := New("obj", 0.5, 0.2, []task.Task{
		task.NewWithCapability("a", "A", "", domainagent.CapabilityResearch),
		task.NewWithCapability("b", "B", "", domainagent.CapabilityCoding, "a"),
	})
	if err := p.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestPlan_Validate_UnknownDep(t *testing.T) {
	p := New("obj", 0.5, 0.2, []task.Task{
		task.NewWithCapability("a", "A", "", domainagent.CapabilityResearch, "ghost"),
	})
	if err := p.Validate(); err == nil {
		t.Fatal("expected unknown dep error")
	}
}

func TestPlan_Validate_Cycle(t *testing.T) {
	p := New("obj", 0.5, 0.2, []task.Task{
		task.NewWithCapability("a", "A", "", domainagent.CapabilityResearch, "b"),
		task.NewWithCapability("b", "B", "", domainagent.CapabilityCoding, "a"),
	})
	if err := p.Validate(); err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestPlan_Levels_Parallel(t *testing.T) {
	p := New("obj", 0.5, 0.2, []task.Task{
		task.NewWithCapability("a", "A", "", domainagent.CapabilityResearch),
		task.NewWithCapability("b", "B", "", domainagent.CapabilityCoding, "a"),
		task.NewWithCapability("c", "C", "", domainagent.CapabilityCoding, "a"),
		task.NewWithCapability("d", "D", "", domainagent.CapabilityTesting, "b", "c"),
	})
	levels := p.Levels()
	if len(levels) != 3 {
		t.Fatalf("expected 3 levels, got %d", len(levels))
	}
	if len(levels[0]) != 1 || len(levels[1]) != 2 || len(levels[2]) != 1 {
		t.Fatalf("wrong level sizes: %v", []int{len(levels[0]), len(levels[1]), len(levels[2])})
	}
}
