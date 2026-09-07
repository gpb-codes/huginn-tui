// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package security

import "strings"

// Policy implementa ports.PermissionPolicy.
type Policy struct {
	// Denegación por defecto para escritura/ejecución/red salvo permiso explícito.
	Allowed map[string]map[string]bool // agentID -> tool -> allowed
}

func NewPolicy() *Policy {
	return &Policy{Allowed: make(map[string]map[string]bool)}
}

func (p *Policy) Allow(agentID, tool string) {
	if p.Allowed[agentID] == nil {
		p.Allowed[agentID] = make(map[string]bool)
	}
	p.Allowed[agentID][tool] = true
}

func (p *Policy) Can(agentID, tool, permission string) bool {
	// La lectura siempre está permitida.
	if permission == "read" {
		return true
	}
	// Comprueba el permiso explícito.
	if m, ok := p.Allowed[agentID]; ok {
		if m[tool] {
			return true
		}
		// Comodín para todas las herramientas.
		if m["*"] {
			return true
		}
	}
	// Permisos base por rol cuando no hay concesión explícita.
	switch agentID {
	case "planner":
		return permission == "read"
	case "coder", "developer":
		return permission == "read" || permission == "write"
	case "researcher":
		return permission == "read" || permission == "network"
	default:
		return permission == "read"
	}
}

// DetectSecret indica si un texto contiene posibles secretos.
func DetectSecret(s string) bool {
	lower := strings.ToLower(s)
	indicators := []string{"api_key", "apikey", "sk-", "ghp_", "gho_", "akia", "private_key", "password=", "passwd", "client_secret", "aws_secret", "bearer "}
	for _, ind := range indicators {
		if strings.Contains(lower, ind) {
			return true
		}
	}
	return false
}

// Redact sustituye valores de posibles secretos para no persistir credenciales.
// Opera por líneas y es conservadora: lo desconocido pasa intacto.
func Redact(s string) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lower := strings.ToLower(ln)
		for _, key := range []string{"api_key", "apikey", "api-key", "password", "passwd", "secret", "token"} {
			if idx := strings.Index(lower, key); idx >= 0 {
				if sep := strings.IndexAny(ln[idx:], "=:"); sep >= 0 {
					lines[i] = ln[:idx+sep+1] + "[redacted]"
					lower = strings.ToLower(lines[i])
				}
			}
		}
		if strings.Contains(lower, "sk-") || strings.Contains(lower, "ghp_") || strings.Contains(lower, "akia") {
			lines[i] = "[redacted: possible credential]"
		}
	}
	return strings.Join(lines, "\n")
}
