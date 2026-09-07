// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package vault

import (
	"os"
	"path/filepath"
)

// ResolveVaultPath devuelve la ruta del vault e indica si existe.
// Prioridad: env HUGINN_VAULT / AGENT_VAULT > ~/agent-vault > ~/huginn-vault > ~/.huginn/vault.
// Capa de dominio: solo comprueba el sistema de ficheros, sin detalles de infraestructura.
func ResolveVaultPath() (string, bool) {
	if v := os.Getenv("HUGINN_VAULT"); v != "" {
		if _, err := os.Stat(v); err == nil {
			return v, true
		}
		return v, false
	}
	if v := os.Getenv("AGENT_VAULT"); v != "" {
		if _, err := os.Stat(v); err == nil {
			return v, true
		}
		return v, false
	}
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, "agent-vault"),
		filepath.Join(home, "huginn-vault"),
		filepath.Join(home, ".huginn", "vault"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, true
		}
	}
	if len(candidates) > 0 {
		return candidates[0], false
	}
	return "", false
}
