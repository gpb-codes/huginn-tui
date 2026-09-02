package components

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
)

// ---------- States de la mascota ----------
type MascotState string

const (
	StateThinking   MascotState = "thinking"
	StateExplaining MascotState = "explaining"
	StateGuarding   MascotState = "guarding"
	StateObserving  MascotState = "observing"
	StateCurious    MascotState = "curious"
	StatePatient    MascotState = "patient"
	StateLoyal      MascotState = "loyal"
	StateWise       MascotState = "wise"
	StateIdle       MascotState = "idle"
)

// ---------- ASCII fallback ----------
var mascotSabio = []string{
	"   .--.      ",
	"  / o o \\   ",
	"  |  v  |   ",
	"   \\ ^ /    ",
	"   /_H_\\   ",
}

// ---------- Componente con sprites PNG ----------
type MascotComponent struct {
	state   MascotState
	sprites map[MascotState]image.Image
	width   int
	height  int
	thought string
}

func NewMascotComponent() *MascotComponent {
	m := &MascotComponent{
		state:   StateIdle,
		sprites: make(map[MascotState]image.Image),
		width:   32,
		height:  32,
	}
	m.loadSprites()
	return m
}

func (m *MascotComponent) loadSprites() {
	states := []MascotState{
		StateThinking, StateExplaining, StateGuarding, StateObserving,
		StateCurious, StatePatient, StateLoyal, StateWise, StateIdle,
	}
	for _, state := range states {
		path := filepath.Join("assets", "mascots", string(state)+".png")
		img, err := loadPNG(path)
		if err != nil {
			continue
		}
		m.sprites[state] = img
	}
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func (m *MascotComponent) SetState(state MascotState) { m.state = state }
func (m *MascotComponent) SetThought(thought string)  { m.thought = thought }

// renderImageToText — half-block rendering: 2 pixeles verticales por caracter
func (m *MascotComponent) renderImageToText(img image.Image) string {
	bounds := img.Bounds()
	w := bounds.Dx()
	_ = w

	var sb strings.Builder
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			top := img.At(x, y)
			var bottom color.Color = color.Black
			if y+1 < bounds.Max.Y {
				bottom = img.At(x, y+1)
			}
			fg := colorToHex(top)
			bg := colorToHex(bottom)
			sb.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color(fg)).
				Background(lipgloss.Color(bg)).
				Render("▀"))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// colorToHex — convierte color.Color a "#RRGGBB" para lipgloss v2
func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// View — renderiza mascota con thought bubble opcional
func (m *MascotComponent) View() string {
	var sb strings.Builder

	if m.thought != "" {
		border := strings.Repeat("─", len(m.thought)+2)
		purple := lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9"))
		sb.WriteString(purple.Render("┌" + border + "┐\n"))
		sb.WriteString(purple.Render("│ " + m.thought + " │\n"))
		sb.WriteString(purple.Render("└" + border + "┘\n"))
	}

	if img, ok := m.sprites[m.state]; ok {
		sb.WriteString(m.renderImageToText(img))
	} else {
		for _, l := range mascotSabio {
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#5b6b82")).Render(l) + "\n")
		}
	}
	return sb.String()
}

// ---------- API publica para el TUI ----------

// RenderMascotSmall — ASCII del cuervo sabio (fallback / splash rapido)
func RenderMascotSmall() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#5b6b82"))
	out := ""
	for _, l := range mascotSabio {
		out += style.Render(l) + "\n"
	}
	return out
}

// RenderMascotSprite — renderiza sprite PNG por estado, con optional thought
func RenderMascotSprite(state MascotState, thought string) string {
	m := NewMascotComponent()
	m.SetState(state)
	if thought != "" {
		m.SetThought(thought)
	}
	return m.View()
}

// RenderMascotScaled — renderiza sprite PNG escalado a ancho maximo en caracteres
func RenderMascotScaled(state MascotState, maxChars int) string {
	m := NewMascotComponent()
	m.SetState(state)
	if img, ok := m.sprites[state]; ok {
		scaled := scaleImage(img, maxChars)
		return m.renderImageToText(scaled)
	}
	return RenderMascotSmall()
}

// scaleImage — escala imagen para caber en maxChars de ancho (half-block: 1 char = 1 pixel)
func scaleImage(img image.Image, maxChars int) image.Image {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW <= maxChars {
		return img
	}
	ratio := float64(maxChars) / float64(srcW)
	newW := maxChars
	newH := int(float64(srcH) * ratio)
	if newH%2 != 0 {
		newH++
	}
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := bounds.Min.X + int(float64(x)/ratio)
			srcY := bounds.Min.Y + int(float64(y)/ratio)
			if srcX >= bounds.Max.X {
				srcX = bounds.Max.X - 1
			}
			if srcY >= bounds.Max.Y {
				srcY = bounds.Max.Y - 1
			}
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}
	return dst
}
