package styles

import "charm.land/lipgloss/v2"

// HUGINN Bronze/Gold palette — premium warm theme
var (
	// Base
	Bg      = lipgloss.Color("#130E0A")
	Panel   = lipgloss.Color("#4D3217")
	Panel2  = lipgloss.Color("#4D3217")
	Border  = lipgloss.Color("#634924")
	Border2 = lipgloss.Color("#634924")

	// Texto
	Text  = lipgloss.Color("#FBE7AE")
	Text2 = lipgloss.Color("#9D8E69")
	Muted = lipgloss.Color("#634924")

	// Acentos
	Accent  = lipgloss.Color("#E1A451") // amber
	Accent2 = lipgloss.Color("#CD8D38") // orange

	Success = lipgloss.Color("#E1A451")
	Warn    = lipgloss.Color("#CD8D38")
	Error   = lipgloss.Color("#CD8D38")

	// Compat alias para migracion gradual desde main.go
	ColPanel  = Panel
	ColAccent = Accent
	ColPurple = Accent2
)

// Spacing (en chars/lines, no px en TUI)
const (
	PadX = 2
	PadY = 1
	Gap  = 1
	Radius = 1 // para lipgloss rounded
)

// Motion — duraciones para referencia (no usadas directo en lipgloss, pero documentan)
const (
	DurationMicro = 150 // ms
	DurationPanel = 220
	DurationModal = 300
)
