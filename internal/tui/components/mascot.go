package components

import "charm.land/lipgloss/v2"

var mascotSabio = []string{
	"   .--.      ",
	"  / o o \\   ",
	"  |  v  |   ",
	"   \\ ^ /    ",
	"   /_H_\\   ",
}

// RenderMascotSmall retorna ASCII del cuervo sabio para splash/help
func RenderMascotSmall() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#5b6b82"))
	out := ""
	for _, l := range mascotSabio {
		out += style.Render(l) + "\n"
	}
	return out
}
