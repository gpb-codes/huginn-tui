package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	colPanel   = lipgloss.Color("#111317")
	colBorder  = lipgloss.Color("#1c232e")
	colText    = lipgloss.Color("#e7ecf2")
	colMuted   = lipgloss.Color("#5c6672")
	colRaven   = lipgloss.Color("#5f5f87")
	colPurple  = lipgloss.Color("#9061f9")
	colWhite   = lipgloss.Color("#f4f6f8")
	colSuccess = lipgloss.Color("#2fd67a")
)

// HeaderProps holds data needed for header rendering (extracted from main.go God file).
type HeaderProps struct {
	ProjectPath string
	ProjectName string
	VaultPath   string
	VaultOK     bool
	PkgManager  string
	Version     string
	Width       int
}

// RenderHeader renders the HUGINN wordmark + context line.
// This is the first step of splitting main.go God file (main.go:1595 → internal/tui/components).
func RenderHeader(p HeaderProps) string {
	hugRows := []string{"██  ██  ██  ██   █████", "██  ██  ██  ██  ██    ", "██████  ██  ██  ██ ███", "██  ██  ██  ██  ██  ██", "██  ██   ████    █████"}
	innRows := []string{"████  ██  ██  ██  ██", " ██   ███ ██  ███ ██", " ██   ██ ███  ██ ███", " ██   ██  ██  ██  ██", "████  ██  ██  ██  ██"}
	var bigRows []string
	for i := range hugRows {
		h := lipgloss.NewStyle().Bold(true).Foreground(colPurple).Render(hugRows[i])
		n := lipgloss.NewStyle().Bold(true).Foreground(colWhite).Render(innRows[i])
		gap := lipgloss.NewStyle().Foreground(colBorder).Render("   ")
		bigRows = append(bigRows, h+gap+n)
	}
	bigHuginn := strings.Join(bigRows, "\n")
	ravensTop := lipgloss.NewStyle().Foreground(colRaven).Width(76).Align(lipgloss.Center).Render(" 𓅃         𓅃 ")
	bigBlock := lipgloss.NewStyle().Padding(1, 0).Width(76).Align(lipgloss.Center).Render(bigHuginn)
	wordmark := lipgloss.JoinVertical(lipgloss.Center, ravensTop, bigBlock)
	subtitleInner := lipgloss.NewStyle().Bold(true).Foreground(colMuted).Render("AI  AGENT  ORCHESTRATION")
	subtitle := lipgloss.NewStyle().Padding(0, 1).Width(76).Align(lipgloss.Center).Render(subtitleInner)
	ver := lipgloss.NewStyle().Foreground(colRaven).Width(76).Align(lipgloss.Center).Render(p.Version + "  •  INTELLIGENCE & KNOWLEDGE INFRASTRUCTURE  •  𓅃𓅃")
	divider := lipgloss.NewStyle().Foreground(colBorder).Width(76).Align(lipgloss.Center).Render(strings.Repeat("─", 52))
	header := lipgloss.JoinVertical(lipgloss.Center, "", wordmark, subtitle, ver, "", divider)
	if p.ProjectPath != "" {
		ctx := p.ProjectPath
		if len(ctx) > 52 {
			ctx = "…" + ctx[len(ctx)-51:]
		}
		ctxLine := fmt.Sprintf("Project └─ %s", ctx)
		if p.PkgManager != "" {
			ctxLine += lipgloss.NewStyle().Foreground(colMuted).Render("  • ") + lipgloss.NewStyle().Foreground(colSuccess).Render(p.PkgManager)
		}
		if p.VaultPath != "" {
			vIcon := "○"
			vCol := colMuted
			if p.VaultOK {
				vIcon = "●"
				vCol = colSuccess
			}
			ctxLine += "  " + lipgloss.NewStyle().Foreground(vCol).Render(vIcon+" vault")
		}
		header = lipgloss.JoinVertical(lipgloss.Center, header, lipgloss.NewStyle().Foreground(colMuted).Width(76).Align(lipgloss.Center).Render(ctxLine), "")
	}
	_ = colPanel
	_ = colText
	return header
}
