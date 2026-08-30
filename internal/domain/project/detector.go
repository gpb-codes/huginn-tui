package project

import (
	"os"
	"path/filepath"
	"strings"
)

// IsDirectory reports whether p is an existing directory.
// Used by CLI to distinguish `huginn <path>` vs `huginn "<prompt>"`.
// Handles Windows quoted args and C:\ paths.
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

// DetectPackageManager inspects lockfiles in root and returns "bun" | "pnpm" | "yarn" | "npm" | "".
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

// DetectProject checks for project markers; returns true and the marker that was found.
func DetectProject(root string) (bool, string) {
	markers := []string{"go.mod", "package.json", "pyproject.toml", "Cargo.toml", "README.md", ".git", "AGENTS.md", "huginn.json", "huginn.config.json"}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(root, m)); err == nil {
			return true, m
		}
	}
	return false, ""
}
