// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package models

import "strings"

// ChatKey es la clave de ModelMatrix para el chat conversacional.
// Las claves coinciden con los nombres de Capability de domain/agent.
const ChatKey = "chat"

// ModelMatrix asocia cada capacidad (o ChatKey) con un modelo Ollama.
// Es solo datos: config posee el JSON y los proveedores lo leen.
type ModelMatrix map[string]string

// For devuelve el modelo de una capacidad o "" si falta (el llamante usa defecto).
// Compara sin distinguir mayúsculas e ignora espacios laterales.
func (m ModelMatrix) For(capability string) string {
	if m == nil {
		return ""
	}
	key := strings.ToLower(strings.TrimSpace(capability))
	if key == "" {
		return ""
	}
	return m[key]
}
