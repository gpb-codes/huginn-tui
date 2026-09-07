// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package project

// Project representa un proyecto detectado en disco.
type Project struct {
	Path           string
	Name           string
	PackageManager string
	Markers        []string
	IsDetected     bool
}
