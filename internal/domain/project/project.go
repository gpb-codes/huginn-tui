package project

// Project represents a detected project on disk.
type Project struct {
	Path           string
	Name           string
	PackageManager string
	Markers        []string
	IsDetected     bool
}
