// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package project

import (
	"os"
	"path/filepath"
	"strings"
)

// IsDirectory indica si p es un directorio existente.
// Usado por el CLI para distinguir `huginn <path>` de `huginn "<prompt>"`.
// Tolera argumentos entrecomillados de Windows y rutas C:\.
func IsDirectory(p string) bool {
	if p == "" {
		return false
	}
	p = strings.Trim(p, `"'`)
	clean := filepath.Clean(p)
	info, err := os.Stat(clean)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// DetectPackageManager inspecciona los lockfiles de root y devuelve "bun" | "pnpm" | "yarn" | "npm" | "".
func DetectPackageManager(root string) string {
	if _, err := os.Stat(filepath.Join(root, "bun.lockb")); err == nil {
		return "bun"
	}
	if _, err := os.Stat(filepath.Join(root, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(root, "yarn.lock")); err == nil {
		return "yarn"
	}
	if _, err := os.Stat(filepath.Join(root, "package-lock.json")); err == nil {
		return "npm"
	}
	if _, err := os.Stat(filepath.Join(root, "package.json")); err == nil {
		return "npm"
	}
	return ""
}

// DetectProject busca marcadores de proyecto y devuelve si lo es más el marcador hallado.
func DetectProject(root string) (bool, string) {
	markers := []string{"go.mod", "package.json", "pyproject.toml", "Cargo.toml", "README.md", ".git", "AGENTS.md", "huginn.json", "huginn.config.json"}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(root, m)); err == nil {
			return true, m
		}
	}
	return false, ""
}
