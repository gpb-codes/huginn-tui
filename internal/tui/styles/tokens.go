// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package styles

import "charm.land/lipgloss/v2"

// Paleta bronce/dorado de HUGINN; fuente única de verdad (DESIGN.md §2).
// El resto reexporta estos tokens sin redefinir hex.
var (
	// Base — canvas oscuro cálido, paneles con jerarquía real.
	// Panel es oscuro para que el borde bronce se lea; Panel2 es el inset claro.
	Bg      = lipgloss.Color("#130E0A")
	Panel   = lipgloss.Color("#20160E")
	Panel2  = lipgloss.Color("#2E2013")
	Border  = lipgloss.Color("#634924")
	Border2 = lipgloss.Color("#976629")

	// Texto — cream primario, dorado apagado secundario, bronce muted.
	Text   = lipgloss.Color("#FBE7AE")
	Text2  = lipgloss.Color("#C9A86C")
	Muted  = lipgloss.Color("#9D8E69")
	Muted2 = lipgloss.Color("#7A6950")

	// Acentos: ámbar primario y naranja secundario.
	Accent  = lipgloss.Color("#E1A451") // amber
	Accent2 = lipgloss.Color("#CD8D38") // orange

	// Semánticos — cálidos pero distinguibles entre sí.
	// Success es salvia cálida (no ámbar), Error terracota (no rojo plano).
	Success = lipgloss.Color("#9CAF7A")
	Warn    = lipgloss.Color("#E1A451")
	Error   = lipgloss.Color("#C96A4A")
	Info    = lipgloss.Color("#8AB4B8")
	Purple  = lipgloss.Color("#CD8D38")

	// Aliases de compatibilidad para la migración gradual desde main.go.
	ColPanel  = Panel
	ColAccent = Accent
	ColPurple = Accent2
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
