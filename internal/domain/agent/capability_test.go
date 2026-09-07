// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agent

import "testing"

func TestAgent_Supports(t *testing.T) {
	a := Agent{ID: "x", Capabilities: []string{"coding", "git"}}
	if !a.Supports(CapabilityCoding) {
		t.Fatal("should support coding")
	}
	if a.Supports(CapabilityResearch) {
		t.Fatal("should not support research")
	}
	if !a.SupportsAll(CapabilityCoding, CapabilityGit) {
		t.Fatal("should support all")
	}
	if a.SupportsAll(CapabilityCoding, CapabilityBrowser) {
		t.Fatal("should not support all")
	}
}

func TestHasCapabilities(t *testing.T) {
	have := []Capability{CapabilityCoding, CapabilityTesting}
	if !HasCapabilities(have, CapabilityCoding) {
		t.Fatal("expected true")
	}
	if HasCapabilities(have, CapabilityReview) {
		t.Fatal("expected false")
	}
	if HasCapabilities(have) {
		t.Fatal("empty want must be false")
	}
}
