package styles

import "charm.land/lipgloss/v2"

// Obsidian Glass — tokens premium para HUGINN TUI
var (
	// Base
	Bg      = lipgloss.Color("#080c14")
	Panel   = lipgloss.Color("#111827")
	Panel2  = lipgloss.Color("#151e2f")
	Border  = lipgloss.Color("#1f2a3a")
	Border2 = lipgloss.Color("#263449")

	// Texto
	Text  = lipgloss.Color("#e6edf3")
	Text2 = lipgloss.Color("#9aa8b8")
	Muted = lipgloss.Color("#5b6b82")

	// Acentos
	Accent  = lipgloss.Color("#22d3ee") // cian
	Accent2 = lipgloss.Color("#8b5cf6") // purpura

	Success = lipgloss.Color("#34d399")
	Warn    = lipgloss.Color("#fbbf24")
	Error   = lipgloss.Color("#f87171")

	// Compat alias para migración gradual desde main.go
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
