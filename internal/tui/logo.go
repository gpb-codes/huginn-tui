// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package tui

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"strings"

	"charm.land/lipgloss/v2"
)

//go:embed assets/logo.png
var logoFS embed.FS

var cachedLogo image.Image

func loadLogo() image.Image {
	if cachedLogo != nil {
		return cachedLogo
	}
	b, err := logoFS.ReadFile("assets/logo.png")
	if err != nil {
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil
	}
	cachedLogo = img
	return img
}

// renderLogo ahora es ASCII puro como pide el diseño: HUG café, INN blanco.
// Mantiene el PNG como fallback si se necesita, pero el principal es el bloque ASCII.
func renderLogo(maxWidth int) string {
	return renderAsciiHugInn()
}

// renderAsciiHugInn devuelve el bloque HUGINN en ASCII de caja con HUG café y INN blanco.
func renderAsciiHugInn() string {
	// 6 líneas, HUG (H U G) a la izquierda en café, INN (I N N) a la derecha en blanco
	hug := []string{
		"██╗  ██╗██╗   ██╗ ██████╗",
		"██║  ██║██║   ██║██╔════╝",
		"███████║██║   ██║██║  ███╗",
		"██╔══██║██║   ██║██║   ██║",
		"██║  ██║╚██████╔╝╚██████╔╝",
		"╚═╝  ╚═╝ ╚═════╝  ╚═════╝",
	}
	inn := []string{
		" ██╗███╗   ██╗███╗   ██╗",
		" ██║████╗  ██║████╗  ██║",
		" ██║██╔██╗ ██║██╔██╗ ██║",
		" ██║██║╚██╗██║██║╚██╗██║",
		" ██║██║ ╚████║██║ ╚████║",
		" ╚═╝╚═╝  ╚═══╝╚═╝  ╚═══╝",
	}
	cafe := lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5A2B")).Bold(true)
	blanco := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	var sb strings.Builder
	for i := range hug {
		sb.WriteString(cafe.Render(hug[i]))
		sb.WriteString(blanco.Render(inn[i]))
		if i < len(hug)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func renderLogoImage(maxWidth int) string {
	img := loadLogo()
	if img == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FBE7AE")).Bold(true).Render("H U G I N N")
	}
	img = cropToContent(img)
	img = recolorHugInn(img)
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	if srcW == 0 || srcH == 0 {
		return ""
	}
	if maxWidth <= 0 {
		maxWidth = 48
	}
	if srcW > maxWidth {
		ratio := float64(maxWidth) / float64(srcW)
		newW := maxWidth
		newH := int(float64(srcH) * ratio)
		if newH%2 != 0 {
			newH++
		}
		if newH < 2 {
			newH = 2
		}
		img = scaleImage(img, newW, newH)
		bounds = img.Bounds()
	}

	var sb strings.Builder
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			top := img.At(x, y)
			var bottom color.Color = color.RGBA{0x13, 0x0E, 0x0A, 0xFF}
			if y+1 < bounds.Max.Y {
				bottom = img.At(x, y+1)
			}
			top = blendOnBg(top)
			bottom = blendOnBg(bottom)
			fg := colorToHex(top)
			bg := colorToHex(bottom)
			sb.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color(fg)).
				Background(lipgloss.Color(bg)).
				Render("▀"))
		}
		if y+2 < bounds.Max.Y {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func cropToContent(img image.Image) image.Image {
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y
	found := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			alpha := uint8(a >> 8)
			if alpha < 20 {
				continue
			}
			// Ignora el fondo sólido #0A0A0F / #130E0A cercano a negro
			rr, gg, bb := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			if rr < 20 && gg < 20 && bb < 20 {
				continue
			}
			// Considera contenido si no es fondo y alpha visible
			if alpha > 30 {
				found = true
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if !found {
		return img
	}
	// Sin padding extra: recorte ajustado al contenido para no desperdiciar filas
	// Asegura rect válido
	if maxX <= minX || maxY <= minY {
		return img
	}
	cropped := image.NewRGBA(image.Rect(0, 0, maxX-minX+1, maxY-minY+1))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			cropped.Set(x-minX, y-minY, img.At(x, y))
		}
	}
	return cropped
}

func recolorHugInn(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	// El texto ocupa el ~70% derecho del crop; HUG 30-62%, INN 62-100%
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.At(bounds.Min.X+x, bounds.Min.Y+y)
			r, g, b, a := c.RGBA()
			rr, gg, bb, aa := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)
			if aa < 20 {
				out.Set(x, y, c)
				continue
			}
			// Solo recolorea la zona de texto (derecha)
			isTextZone := x > w*3/10
			isHugZone := x > w*3/10 && x < w*62/100
			isInnZone := x >= w*62/100
			// Detecta naranja café del HUG (R 180-255, G 90-170, B 20-90)
			isOrange := rr > 170 && gg > 80 && gg < 180 && bb < 90
			// Detecta beige claro del INN (R>200, G>180, B>120)
			isLight := rr > 200 && gg > 180 && bb > 110
			if isTextZone && isHugZone && isOrange {
				// HUG → café profundo #8B5A2B
				out.Set(x, y, color.RGBA{0x8B, 0x5A, 0x2B, aa})
				continue
			}
			if isTextZone && isInnZone && isLight {
				// INN → blanco puro
				out.Set(x, y, color.RGBA{0xFF, 0xFF, 0xFF, aa})
				continue
			}
			out.Set(x, y, c)
		}
	}
	return out
}

func blendOnBg(c color.Color) color.Color {
	r, g, b, a := c.RGBA()
	alpha := uint8(a >> 8)
	if alpha > 200 {
		return c
	}
	if alpha < 20 {
		return color.RGBA{0x13, 0x0E, 0x0A, 0xFF}
	}
	// mezcla simple sobre Bg #130E0A
	bgR, bgG, bgB := uint8(0x13), uint8(0x0E), uint8(0x0A)
	fgR, fgG, fgB := uint8(r>>8), uint8(g>>8), uint8(b>>8)
	mix := func(fg, bg uint8) uint8 { return uint8((int(fg)*int(alpha) + int(bg)*(255-int(alpha))) / 255) }
	return color.RGBA{mix(fgR, bgR), mix(fgG, bgG), mix(fgB, bgB), 0xFF}
}

func scaleImage(src image.Image, newW, newH int) image.Image {
	bounds := src.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	ratioX := float64(srcW) / float64(newW)
	ratioY := float64(srcH) / float64(newH)
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := bounds.Min.X + int(float64(x)*ratioX)
			srcY := bounds.Min.Y + int(float64(y)*ratioY)
			if srcX >= bounds.Max.X {
				srcX = bounds.Max.X - 1
			}
			if srcY >= bounds.Max.Y {
				srcY = bounds.Max.Y - 1
			}
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}
