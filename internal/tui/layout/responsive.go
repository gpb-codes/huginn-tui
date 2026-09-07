// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package layout

// Breakpoint define anchos conceptuales para responsive (spec §22).
type Breakpoint string

const (
	Compact Breakpoint = "compact" // 80x24 — oculta sidebar, reduce metadata
	Normal  Breakpoint = "normal"  // 100x30 — layout estándar
	Wide    Breakpoint = "wide"    // 120x40+ — muestra todo
)

// At determina el breakpoint según el ancho del terminal.
func At(width int) Breakpoint {
	if width < 90 {
		return Compact
	}
	if width < 120 {
		return Normal
	}
	return Wide
}

// SidebarVisible indica si el sidebar debe mostrarse en este breakpoint.
func SidebarVisible(bp Breakpoint, userToggled bool) bool {
	if bp == Compact {
		return false
	}
	return userToggled
}

// InputWidth calcula el ancho del input según el breakpoint.
func InputWidth(totalWidth int, bp Breakpoint) int {
	switch bp {
	case Compact:
		w := totalWidth - 8
		if w < 30 {
			w = 30
		}
		return w
	case Normal:
		w := 60
		if totalWidth < 80 {
			w = totalWidth - 16
		}
		return w
	default: // Wide
		return 68
	}
}
