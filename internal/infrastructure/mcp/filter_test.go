// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package mcp

import (
	"testing"

	domainagent "huginn/internal/domain/agent"
)

func TestForCapability_CodingMinimal(t *testing.T) {
	got := ForCapability(domainagent.CapabilityCoding)
	names := map[string]bool{}
	for _, s := range got {
		names[s.Name] = true
	}
	for _, want := range []string{"filesystem", "github", "sequential"} {
		if !names[want] {
			t.Errorf("missing %s in %v", want, names)
		}
	}
	if names["notion"] || names["vault-sync"] {
		t.Errorf("over-provisioned: %v", names)
	}
}

func TestForCapability_UnknownDenies(t *testing.T) {
	if got := ForCapability(domainagent.Capability("telemetry")); got != nil {
		t.Fatalf("expected default-deny, got %v", got)
	}
}

func TestForCapability_Browser(t *testing.T) {
	got := ForCapability(domainagent.CapabilityBrowser)
	if len(got) != 1 || got[0].Name != "playwright" {
		t.Fatalf("expected playwright only, got %v", got)
	}
}
