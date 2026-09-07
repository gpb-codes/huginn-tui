// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package styles

import "charm.land/lipgloss/v2"

// Sistema de diseño HUGINN — spec §5 (tui-design-skill).
// Colores con significado semántico, no decoración. Primero monocromo, luego acento.
var (
	// Base — dark minimal, technical, premium (spec §5)
	Background   = lipgloss.Color("#0B0F14")
	Surface      = lipgloss.Color("#111820")
	SurfaceActive = lipgloss.Color("#161F29")

	// Aliases legacy para compatibilidad (mapean al nuevo sistema)
	Bg     = Background
	Panel  = Surface
	Panel2 = SurfaceActive
	Border = lipgloss.Color("#1E2A36")
	Border2 = lipgloss.Color("#243447")

	// Texto — jerarquía clara, funciona sin color
	TextPrimary   = lipgloss.Color("#E6EAF0")
	TextSecondary = lipgloss.Color("#9AA4B2")
	TextMuted     = lipgloss.Color("#626C78")
	// Aliases legacy
	Text   = TextPrimary
	Text2  = TextSecondary
	Muted  = TextMuted
	Muted2 = lipgloss.Color("#4A5563")

	// Acento — solo para acciones, estados, selección, identidad HUGINN
	Accent  = lipgloss.Color("#8B5CF6") // spec §5 — violeta HUGINN
	Accent2 = lipgloss.Color("#7C3AED") // violeta profundo para hover
	Purple  = Accent

	// Semánticos — distinguibles, no arcoíris
	Success = lipgloss.Color("#4ADE80")
	Warn    = lipgloss.Color("#FACC15")
	Error   = lipgloss.Color("#F87171")
	Info    = lipgloss.Color("#60A5FA")

	// Aliases legacy
	ColPanel  = Panel
	ColAccent = Accent
	ColPurple = Accent
)

// PanelStyle — panel principal con borde bronce redondeado.
func PanelStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(Panel).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(1, 2).
		Width(width)
}

// InsetStyle — tarjeta secundaria dentro de un panel.
func InsetStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(Panel2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(0, 1).
		Width(width)
}

// TitleStyle aplica títulos de sección en ámbar y negrita.
func TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(Accent)
}

// HintStyle — atajos tenues del footer.
func HintStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Muted)
}

// KeyStyle — tecla destacada dentro de un hint (p.ej. "enter").
func KeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(Text)
}

// SelectedStyle — fila seleccionada: fondo oro oscuro, texto canvas.
func SelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Bg).Background(Accent).Bold(true).Padding(0, 1)
}

// FocusedRowStyle — fila con foco suave (no roba atención del input).
func FocusedRowStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Text).Background(Panel2).Padding(0, 1)
}

// Espaciado en caracteres/líneas, no píxeles de TUI.
const (
	PadX   = 2
	PadY   = 1
	Gap    = 1
	Radius = 1 // para lipgloss rounded
)

// Motion documenta duraciones de referencia (no usadas directo en lipgloss).
const (
	DurationMicro = 150 // ms
	DurationPanel = 220
	DurationModal = 300
)
