// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package mcp

import (
	domainagent "huginn/internal/domain/agent"
)

// ForCapability devuelve la allowlist mínima de servidores MCP por capacidad.
// Activa solo lo necesario por tarea, pues cada servidor consume contexto.
//
// Denegación por defecto: las capacidades desconocidas no reciben servidores.
func ForCapability(cap domainagent.Capability) []Server {
	allow := map[string]bool{}
	switch cap {
	case domainagent.CapabilityCoding, domainagent.CapabilityTesting, domainagent.CapabilityDevOps:
		allow["filesystem"] = true
		allow["github"] = true
		allow["sequential"] = true
	case domainagent.CapabilityResearch, domainagent.CapabilityReasoning:
		allow["memory"] = true
		allow["sequential"] = true
	case domainagent.CapabilityDocumentation:
		allow["filesystem"] = true
		allow["memory"] = true
	case domainagent.CapabilityReview:
		allow["filesystem"] = true
		allow["github"] = true
		allow["sequential"] = true
	case domainagent.CapabilityGit:
		allow["github"] = true
	case domainagent.CapabilityBrowser:
		allow["playwright"] = true
	case domainagent.CapabilityFilesystem:
		allow["filesystem"] = true
	default:
		return nil
	}
	var out []Server
	for _, s := range DefaultServers {
		if allow[s.Name] && s.Enabled {
			out = append(out, s)
		}
	}
	return out
}
