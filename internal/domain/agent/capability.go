// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agent

// Capability describe lo que un runtime de agente puede hacer.
// El planner y el router trabajan con capacidades, nunca con nombres fijos
// de agente: `agent.Supports(CapabilityCoding)` en lugar de
// `agent == "opencode"`.
type Capability string

const (
	CapabilityCoding        Capability = "coding"
	CapabilityResearch      Capability = "research"
	CapabilityReasoning     Capability = "reasoning"
	CapabilityTesting       Capability = "testing"
	CapabilityReview        Capability = "review"
	CapabilityDevOps        Capability = "devops"
	CapabilityDocumentation Capability = "documentation"
	CapabilityBrowser       Capability = "browser"
	CapabilityFilesystem    Capability = "filesystem"
	CapabilityGit           Capability = "git"
)

// AllCapabilities enumera todas las capacidades conocidas.
var AllCapabilities = []Capability{
	CapabilityCoding,
	CapabilityResearch,
	CapabilityReasoning,
	CapabilityTesting,
	CapabilityReview,
	CapabilityDevOps,
	CapabilityDocumentation,
	CapabilityBrowser,
	CapabilityFilesystem,
	CapabilityGit,
}

// Supports indica si el agente declara una capacidad.
func (a Agent) Supports(c Capability) bool {
	for _, have := range a.Capabilities {
		if Capability(have) == c {
			return true
		}
	}
	return false
}

// SupportsAll indica si el agente declara todas las capacidades dadas.
func (a Agent) SupportsAll(cs ...Capability) bool {
	return HasCapabilities(agentCaps(a.Capabilities), cs...)
}

func agentCaps(names []string) []Capability {
	out := make([]Capability, 0, len(names))
	for _, n := range names {
		out = append(out, Capability(n))
	}
	return out
}

// HasCapabilities ayuda a implementaciones de runtime que mantienen las capacidades fuera del struct Agent.
func HasCapabilities(have []Capability, want ...Capability) bool {
	set := make(map[Capability]bool, len(have))
	for _, h := range have {
		set[h] = true
	}
	for _, w := range want {
		if !set[w] {
			return false
		}
	}
	return len(want) > 0
}
