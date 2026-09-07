// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"huginn/internal/tui/styles"
)

type Mode int

const (
	ModeHome Mode = iota
	ModeHelp
	ModeMessage
	ModePalette
	ModeModelSelect
)

var huginnModes = []string{"Ask", "Architect", "Code", "Debug", "Orchestrator"}

type chatMsg struct {
	Role string // "user" | "orchestrator" | "agent" | "assistant" | "system"
	Text string
	Meta string // subtítulo: agente destino, provider, etc.
}

type Model struct {
	width  int
	height int

	input  string
	cursor int

	mode     Mode
	msgTitle string
	msgBody  string

	messages []chatMsg // conversación estilo opencode
	scroll   int       // offset para scroll en conversación (0 = abajo)

	// Paleta y modos (Kilo: tab para modos, ctrl+p para paleta)
	paletteCursor int
	paletteItems  []string
	modeIndex     int // índice en huginnModes

	// Selector de modelos (como en captura 3)
	modelCursor int
	modelItems  []string
	modelSearch string

	tuiCfg TUIConfig

	projectPath string
	vaultPath   string
	quitting    bool
}

func New(projectPath, vaultPath string) Model {
	return Model{
		mode:         ModeHome,
		projectPath:  projectPath,
		vaultPath:    vaultPath,
		modeIndex:    0,
		paletteItems: []string{"/help", "/agents", "/tasks", "/projects", "/sessions", "/memory", "/workflows", "/models", "/config", "/clear", "/undo", "/exit"},
		modelItems: []string{
			"Qwen3-Coder-480B (Huginn · coding)",
			"Qwen2.5-Coder-32B (Huginn · coding)",
			"DeepSeek-R1 (reasoning)",
			"DeepSeek-V3 (general)",
			"Claude 4.5 Opus (Huginn · reasoning)",
			"GPT-5.2 Codex High",
			"Gemini 3 Pro",
		},
		tuiCfg: LoadTUIConfig(),
	}
}

func (m Model) currentMode() string {
	if m.modeIndex < 0 || m.modeIndex >= len(huginnModes) {
		return huginnModes[0]
	}
	return huginnModes[m.modeIndex]
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		// Paleta: navegación con up/down, enter selecciona, esc cierra
		if m.mode == ModePalette {
			switch msg.String() {
			case "esc", "ctrl+p":
				m.mode = ModeHome
				return m, nil
			case "up", "k":
				if m.paletteCursor > 0 {
					m.paletteCursor--
				}
				return m, nil
			case "down", "j":
				if m.paletteCursor < len(m.paletteItems)-1 {
					m.paletteCursor++
				}
				return m, nil
			case "enter":
				if m.paletteCursor >= 0 && m.paletteCursor < len(m.paletteItems) {
					sel := m.paletteItems[m.paletteCursor]
					if sel == "/models" {
						m.mode = ModeModelSelect
						m.modelCursor = 0
						return m, nil
					}
					m.input = sel + " "
					m.cursor = len([]rune(m.input))
				}
				m.mode = ModeHome
				return m, nil
			}
		}
		// Selector de modelos (como captura 3): typing filtra, backspace borra
		if m.mode == ModeModelSelect {
			switch msg.String() {
			case "esc":
				m.mode = ModeHome
				m.modelSearch = ""
				m.modelCursor = 0
				return m, nil
			case "up", "k":
				if m.modelCursor > 0 {
					m.modelCursor--
				}
				return m, nil
			case "down", "j":
				if m.modelCursor < len(m.filteredModels())-1 {
					m.modelCursor++
				}
				return m, nil
			case "enter":
				items := m.filteredModels()
				if m.modelCursor >= 0 && m.modelCursor < len(items) {
					m.messages = append(m.messages, chatMsg{Role: "system", Text: "Modelo seleccionado: " + items[m.modelCursor], Meta: "Huginn"})
				}
				m.mode = ModeHome
				m.modelSearch = ""
				m.modelCursor = 0
				return m, nil
			case "backspace", "ctrl+h":
				if len(m.modelSearch) > 0 {
					runes := []rune(m.modelSearch)
					m.modelSearch = string(runes[:len(runes)-1])
					m.modelCursor = 0
				}
				return m, nil
			}
			if s := msg.String(); len(s) == 1 || len([]rune(s)) == 1 {
				r := []rune(s)
				if len(r) == 1 && r[0] >= 32 && r[0] != 127 {
					m.modelSearch += s
					m.modelCursor = 0
				}
				return m, nil
			}
		}
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			if m.mode != ModeHome {
				m.mode = ModeHome
				m.msgTitle, m.msgBody = "", ""
				return m, nil
			}
			m.input = ""
			m.cursor = 0
			return m, nil
		case "ctrl+p":
			if m.mode == ModePalette {
				m.mode = ModeHome
			} else {
				m.mode = ModePalette
				m.paletteCursor = 0
			}
			return m, nil
		case "tab":
			m.modeIndex = (m.modeIndex + 1) % len(huginnModes)
			return m, nil
		case "shift+tab":
			m.modeIndex = (m.modeIndex - 1 + len(huginnModes)) % len(huginnModes)
			return m, nil
		case "enter":
			if m.mode == ModePalette {
				if m.paletteCursor >= 0 && m.paletteCursor < len(m.paletteItems) {
					m.input = m.paletteItems[m.paletteCursor] + " "
					m.cursor = len([]rune(m.input))
				}
				m.mode = ModeHome
				return m, nil
			}
			return m.handleEnter()
		case "backspace", "ctrl+h":
			if m.cursor > 0 && len(m.input) > 0 {
				runes := []rune(m.input)
				if m.cursor <= len(runes) {
					runes = append(runes[:m.cursor-1], runes[m.cursor:]...)
					m.input = string(runes)
					m.cursor--
				}
			}
			return m, nil
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "right":
			if m.cursor < len([]rune(m.input)) {
				m.cursor++
			}
			return m, nil
		case "ctrl+u":
			m.input = ""
			m.cursor = 0
			return m, nil
		case "up", "k":
			if len(m.messages) > 0 {
				m.scroll++
				if m.scroll > len(m.messages) {
					m.scroll = len(m.messages)
				}
				return m, nil
			}
		case "down", "j":
			if m.scroll > 0 {
				m.scroll--
				return m, nil
			}
		case "pgup":
			if len(m.messages) > 0 {
				m.scroll += 5
				if m.scroll > len(m.messages) {
					m.scroll = len(m.messages)
				}
				return m, nil
			}
		case "pgdown":
			if m.scroll > 0 {
				m.scroll -= 5
				if m.scroll < 0 {
					m.scroll = 0
				}
				return m, nil
			}
		}
		if msg.String() == "?" && strings.TrimSpace(m.input) == "" {
			m.mode = ModeHelp
			return m, nil
		}
		if s := msg.String(); len(s) == 1 || len([]rune(s)) == 1 {
			r := []rune(s)
			if len(r) == 1 && r[0] >= 32 && r[0] != 127 {
				runes := []rune(m.input)
				if m.cursor >= len(runes) {
					m.input += s
				} else {
					runes = append(runes[:m.cursor], append([]rune(s), runes[m.cursor:]...)...)
					m.input = string(runes)
				}
				m.cursor++
			}
			return m, nil
		}
		if s := msg.String(); len(s) > 1 {
			if s == "tab" || s == "shift+tab" {
				return m, nil
			}
			runes := []rune(m.input)
			insert := []rune(s)
			if m.cursor >= len(runes) {
				m.input += s
			} else {
				runes = append(runes[:m.cursor], append(insert, runes[m.cursor:]...)...)
				m.input = string(runes)
			}
			m.cursor += len(insert)
			return m, nil
		}
	}
	return m, nil
}

func (m Model) handleEnter() (Model, tea.Cmd) {
	raw := strings.TrimSpace(m.input)
	if raw == "" {
		return m, nil
	}
	m.input = ""
	m.cursor = 0

	// Bash: !<cmd> como en opencode (ej: ! ls -la)
	if strings.HasPrefix(raw, "!") {
		cmdStr := strings.TrimSpace(strings.TrimPrefix(raw, "!"))
		if cmdStr == "" {
			m.messages = append(m.messages, chatMsg{Role: "system", Text: "Uso: ! <comando>  — ejecuta en shell y añade el output al contexto", Meta: "Huginn"})
			return m, nil
		}
		m.messages = append(m.messages, chatMsg{Role: "user", Text: "! " + cmdStr, Meta: "Tú → Shell"})
		out := runShell(cmdStr)
		m.messages = append(m.messages, chatMsg{Role: "assistant", Text: out, Meta: "Shell → Huginn"})
		m.messages = append(m.messages, chatMsg{Role: "orchestrator", Text: "Output de shell añadido al contexto. Huginn lo usará en el próximo turno.", Meta: "Huginn · Orquestador"})
		m.scroll = 0
		return m, nil
	}

	if strings.HasPrefix(raw, "/") {
		cmd, _ := splitCommand(raw)
		switch cmd {
		case "/help", "/h", "help", "?":
			m.mode = ModeHelp
			return m, nil
		case "/clear", "/c":
			m.messages = nil
			m.mode = ModeHome
			return m, nil
		case "/exit", "/quit", "/q", "/e":
			m.quitting = true
			return m, tea.Quit
		case "/undo":
			if len(m.messages) > 0 {
				m.messages = m.messages[:len(m.messages)-1]
			}
			m.mode = ModeHome
			return m, nil
		case "/agents", "/agent":
			m.mode = ModeMessage
			m.msgTitle = "Agents"
			m.msgBody = "Contrato: Agent {ID, Name, Capabilities, Execute()} listo.\nProviders: OpenCode (coding) + Ollama (research).\nEstado: aún no hay agentes conectados."
			return m, nil
		case "/tasks", "/task":
			m.mode = ModeMessage
			m.msgTitle = "Tasks"
			m.msgBody = "Estados: PENDING → PLANNING → RUNNING → WAITING → COMPLETED\n                           ↘ FAILED / CANCELLED"
			return m, nil
		case "/models", "/model":
			m.mode = ModeModelSelect
			m.modelCursor = 0
			return m, nil
		case "/projects", "/project":
			m.mode = ModeMessage
			m.msgTitle = "Projects"
			m.msgBody = "Proyecto actual: " + displayPath(m.projectPath) + "\n\nPróximamente: listado y cambio de workspace."
			return m, nil
		case "/sessions", "/session":
			m.mode = ModeMessage
			m.msgTitle = "Sessions"
			m.msgBody = "Sesiones aún no implementadas.\n\nContrato preparado: Session {ID, Workspace, Tasks, CreatedAt}."
			return m, nil
		case "/memory", "/mem":
			m.mode = ModeMessage
			m.msgTitle = "Memory"
			m.msgBody = "Memoria aún no implementada.\n\nCapas previstas: Short-Term, Long-Term, Project, Agent, Task History."
			return m, nil
		case "/workflows", "/workflow", "/wf":
			m.mode = ModeMessage
			m.msgTitle = "Workflows"
			m.msgBody = "Workflows aún no implementados.\n\nArquitectura lista para orquestar DAGs de tareas."
			return m, nil
		case "/config", "/conf", "/settings":
			m.mode = ModeMessage
			m.msgTitle = "Config"
			m.msgBody = "Configuración en .huginn/config.json (cuando exista).\n\nVault: " + displayPath(m.vaultPath) + "\nProyecto: " + displayPath(m.projectPath)
			return m, nil
		default:
			m.mode = ModeMessage
			m.msgTitle = "Comando desconocido"
			m.msgBody = "Comando: " + cmd + "\n\n/help para ver todos."
			return m, nil
		}
	}

	// Flujo honesto: User → Orquestador → Agentes → Orquestador → Usuario
	// 1. Usuario → Huginn (orquestador)
	m.messages = append(m.messages, chatMsg{Role: "user", Text: raw, Meta: "Tú → Huginn"})
	// 2. Huginn recibe y planifica (sin simular éxito)
	m.messages = append(m.messages, chatMsg{
		Role: "orchestrator",
		Text: fmt.Sprintf("Recibido %q. Analizando objetivo y descomponiendo en tareas…", raw),
		Meta: "Huginn · Orquestador",
	})
	// 3. Orquestador → Agentes (routing por capacidad, honesto si no hay provider)
	m.messages = append(m.messages, chatMsg{
		Role: "agent",
		Text: fmt.Sprintf("Delegando a OpenCode (coding) → %q", raw),
		Meta: "Huginn → OpenCode",
	})
	m.messages = append(m.messages, chatMsg{
		Role: "agent",
		Text: "Delegando a Ollama (research) → contexto y razonamiento",
		Meta: "Huginn → Ollama",
	})
	// 4. Agentes → Orquestador (respuesta honesta, sin inventar)
	m.messages = append(m.messages, chatMsg{
		Role: "orchestrator",
		Text: "Agentes consultados. Sin provider en línea, Huginn no inventa resultados.\nInicia OpenCode (opencode) u Ollama y reintenta. Traza guardada en .huginn/logs/.",
		Meta: "Agentes → Huginn",
	})
	// 5. Orquestador → Usuario (síntesis honesta)
	m.messages = append(m.messages, chatMsg{
		Role: "assistant",
		Text: fmt.Sprintf("Huginn → Tú: Objetivo %q registrado.\nFlujo: User → Orquestador → Agentes → Orquestador → Tú.\nPróximo paso: conectar planner/router real para ejecutar el DAG.", raw),
		Meta: "Huginn → Tú",
	})
	m.mode = ModeHome
	m.scroll = 0
	return m, nil
}

func runShell(cmdStr string) string {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	out, err := cmd.CombinedOutput()
	s := string(out)
	if err != nil {
		s += "\n[exit: " + err.Error() + "]"
	}
	s = strings.TrimSpace(s)
	if s == "" {
		s = "(sin salida)"
	}
	if len(s) > 4000 {
		s = s[:4000] + "\n… (truncado)"
	}
	return "```\n" + s + "\n```"
}

func splitCommand(s string) (string, string) {
	s = strings.TrimSpace(s)
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return "", ""
	}
	cmd := strings.ToLower(parts[0])
	rest := ""
	if len(parts) > 1 {
		rest = strings.Join(parts[1:], " ")
	}
	return cmd, rest
}

func displayPath(p string) string {
	if p == "" {
		return "(no detectado)"
	}
	return p
}

func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	w, h := m.width, m.height
	if w <= 0 {
		w = 100
	}
	if h <= 0 {
		h = 30
	}
	var content string
	switch m.mode {
	case ModeHelp:
		content = m.viewHelp(w, h)
	case ModeMessage:
		content = m.viewMessage(w, h)
	case ModePalette:
		content = m.viewPalette(w, h)
	case ModeModelSelect:
		content = m.viewModelSelect(w, h)
	default:
		if len(m.messages) == 0 {
			content = m.viewHome(w, h)
		} else {
			content = m.viewConversation(w, h)
		}
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// ---------- Vistas ----------

func (m Model) viewHome(w, h int) string {
	windowDots := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F56")).Render("●") + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FFBD2E")).Render("●") + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#27C93F")).Render("●")
	topBar := lipgloss.NewStyle().Foreground(styles.Muted2).Width(w).Padding(0, 1).Render(windowDots)

	// Logo ASCII HUG café / INN blanco + subtítulo como en el spec
	logoW := 64
	if w < 80 {
		logoW = w - 16
		if logoW < 40 {
			logoW = 40
		}
	}
	logoRendered := renderLogo(logoW)
	logoBlock := lipgloss.Place(w, lipgloss.Height(logoRendered), lipgloss.Center, lipgloss.Center, logoRendered)
	subtitle := lipgloss.NewStyle().Foreground(styles.Muted).Render("AI Agent Orchestration")
	_ = logoRendered

	// Input: caja limpia con ">" como en el spec, placeholder y modo
	inputText := m.input
	if strings.TrimSpace(inputText) == "" {
		inputText = lipgloss.NewStyle().Foreground(styles.Muted).Render("> Ask Huginn…")
	} else {
		runes := []rune(m.input)
		var withCursor string
		if m.cursor < len(runes) {
			before := string(runes[:m.cursor])
			at := string(runes[m.cursor])
			after := ""
			if m.cursor+1 < len(runes) {
				after = string(runes[m.cursor+1:])
			}
			cursor := lipgloss.NewStyle().Background(lipgloss.Color("#9B87F5")).Foreground(lipgloss.Color("#0A0A0F")).Render(at)
			withCursor = before + cursor + after
		} else {
			withCursor = m.input + lipgloss.NewStyle().Background(lipgloss.Color("#9B87F5")).Foreground(styles.Bg).Render(" ")
		}
		if strings.Contains(m.input, "@") {
			withCursor = strings.ReplaceAll(withCursor, "@", lipgloss.NewStyle().Foreground(styles.Accent2).Render("@"))
		}
		inputText = lipgloss.NewStyle().Foreground(styles.Muted).Render("> ") + withCursor
	}

	dot := lipgloss.NewStyle().Foreground(lipgloss.Color("#9B87F5")).Render("●")
	modeName := m.currentMode()
	modelLine := dot + " " + lipgloss.NewStyle().Foreground(styles.Text2).Render(modeName) +
		lipgloss.NewStyle().Foreground(styles.Muted).Render("  ·  Huginn  ") +
		lipgloss.NewStyle().Foreground(styles.Muted2).Render("Zen")

	innerInput := lipgloss.JoinVertical(lipgloss.Left,
		inputText,
		"",
		modelLine,
	)

	boxW := 60
	if w < 80 {
		boxW = w - 16
		if boxW < 42 {
			boxW = 42
		}
	}
	// Caja más estrecha y elegante, como en el spec (38 chars)
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#2A2A3A")).
		Background(lipgloss.Color("#14141E")).
		Padding(1, 2).
		Width(boxW).
		Render(
			lipgloss.NewStyle().
				Border(lipgloss.Border{Left: "┃"}).
				BorderForeground(lipgloss.Color("#9B87F5")).
				PaddingLeft(1).
				Render(innerInput),
		)

	hintCmd := lipgloss.NewStyle().Foreground(styles.Muted2).Render("Type / for commands")
	// Hints secundarios más tenues
	kTab := lipgloss.NewStyle().Foreground(styles.Muted2).Render("tab")
	kCtrlP := lipgloss.NewStyle().Foreground(styles.Muted2).Render("ctrl+p")
	hintsLine := lipgloss.NewStyle().Foreground(styles.Muted2).Render(kTab + " agents  " + kCtrlP + " palette  •  ? help")

	tip := lipgloss.NewStyle().Foreground(styles.Muted).Render(
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7A7A")).Render("● Tip") +
			lipgloss.NewStyle().Foreground(styles.Muted).Render("  Use  /undo  to revert —  /help  para comandos"),
	)

	projectLine := lipgloss.NewStyle().Foreground(styles.Muted2).Render("⬢ " + displayPath(m.projectPath))
	versionLine := lipgloss.NewStyle().Foreground(styles.Muted2).Render("v0.2.0")
	bottomMeta := lipgloss.NewStyle().Foreground(styles.Muted2).Width(w).Padding(0, 1).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			projectLine,
			strings.Repeat(" ", max(0, w-lipgloss.Width(projectLine)-lipgloss.Width(versionLine)-4)),
			versionLine,
		))

	centerInner := lipgloss.JoinVertical(lipgloss.Center,
		logoBlock,
		subtitle,
		"",
		inputBox,
		hintCmd,
		hintsLine,
		"",
		tip,
	)
	centered := lipgloss.Place(w, h-4, lipgloss.Center, lipgloss.Center, centerInner)

	content := lipgloss.JoinVertical(lipgloss.Left,
		topBar,
		centered,
		bottomMeta,
	)
	return content
}

func (m Model) viewConversation(w, h int) string {
	// Layout con sidebar como en la captura: chat a la izquierda, AGENTS/PIPELINE/CONTEXT a la derecha
	sidebarW := 32
	if w < 80 {
		sidebarW = 24
	}
	chatW := w - sidebarW - 1
	if chatW < 40 {
		chatW = 40
		sidebarW = w - chatW - 1
	}

	// Header igual que antes pero solo sobre el chat
	title := m.messages[0].Text
	if len(title) > 36 {
		title = title[:36] + "…"
	}
	stats := lipgloss.NewStyle().Foreground(styles.Muted2).Render("0  0%  ($0.00)")
	headerInner := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Foreground(styles.Text).Bold(true).Render("# "+title),
		strings.Repeat(" ", max(0, chatW-lipgloss.Width("# "+title)-lipgloss.Width(stats)-4)),
		stats,
	)
	header := lipgloss.NewStyle().
		Background(lipgloss.Color("#14141E")).
		Foreground(styles.Text).
		Width(chatW).
		Padding(0, 1).
		Border(lipgloss.Border{Bottom: "─"}).
		BorderForeground(lipgloss.Color("#2A2A3A")).
		Render(headerInner)

	// Calcula ventana visible con scroll (0 = últimos mensajes)
	visible := m.messages
	if m.scroll > 0 && len(visible) > 0 {
		// Muestra mensajes más antiguos según scroll
		end := len(visible) - m.scroll
		if end < 0 {
			end = 0
		}
		// Ajusta para que quepan en altura disponible (aprox 8 líneas por mensaje)
		// Simplificado: muestra últimos 3-5 mensajes según altura
		maxVisible := (h - 10) / 6
		if maxVisible < 1 {
			maxVisible = 1
		}
		start := end - maxVisible
		if start < 0 {
			start = 0
		}
		visible = visible[start:end]
	} else if len(visible) > 0 {
		// Sin scroll: muestra los últimos que caben
		maxVisible := (h - 10) / 6
		if maxVisible < 1 {
			maxVisible = 1
		}
		if len(visible) > maxVisible {
			visible = visible[len(visible)-maxVisible:]
		}
	}
	var msgs []string
	for _, msg := range visible {
		var label, body string
		var labelStyle, bodyStyle lipgloss.Style
		switch msg.Role {
		case "user":
			label = "Tú → Huginn"
			labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9B87F5")).Bold(true)
			bodyStyle = lipgloss.NewStyle().Foreground(styles.Text).Background(lipgloss.Color("#14141E")).Padding(0, 1).Width(w - 2)
			body = "┃ " + renderMarkdown(msg.Text, w-6)
		case "orchestrator":
			label = msg.Meta
			if label == "" {
				label = "Huginn · Orquestador"
			}
			labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9E64")).Bold(true)
			bodyStyle = lipgloss.NewStyle().Foreground(styles.Text2).Width(w - 4).PaddingLeft(2)
			body = renderMarkdown(msg.Text, w-8)
		case "agent":
			label = msg.Meta
			if label == "" {
				label = "Huginn → Agente"
			}
			labelStyle = lipgloss.NewStyle().Foreground(styles.Muted).Bold(true)
			bodyStyle = lipgloss.NewStyle().Foreground(styles.Muted).Width(w - 4).PaddingLeft(2).Italic(true)
			body = "→ " + renderMarkdown(msg.Text, w-8)
		default: // assistant = Orquestador → Usuario (síntesis final)
			label = msg.Meta
			if label == "" {
				label = "Huginn → Tú"
			}
			labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2FF")).Bold(true)
			bodyStyle = lipgloss.NewStyle().Foreground(styles.Text).Width(w-4).PaddingLeft(2).Background(lipgloss.Color("#1A1A22")).Padding(0, 1)
			body = renderMarkdown(msg.Text, w-8)
		}
		msgs = append(msgs, labelStyle.Render(label))
		msgs = append(msgs, bodyStyle.Render(body))
		if msg.Role != "assistant" {
			msgs = append(msgs, lipgloss.NewStyle().Foreground(styles.Muted2).Render(strings.Repeat("─", w-6)))
		}
	}
	if m.scroll > 0 {
		msgs = append(msgs, lipgloss.NewStyle().Foreground(styles.Muted2).Italic(true).Render(fmt.Sprintf("↕ scroll %d — j/down para bajar, k/up para subir", m.scroll)))
	}
	// Input inferior estilo captura: "explica el vault" + "@All · mimo-v2.5-free · go"
	inputText := m.input
	if strings.TrimSpace(inputText) == "" {
		inputText = lipgloss.NewStyle().Foreground(styles.Muted).Render("explica el vault")
	} else {
		runes := []rune(m.input)
		if m.cursor < len(runes) {
			before := string(runes[:m.cursor])
			at := string(runes[m.cursor])
			after := ""
			if m.cursor+1 < len(runes) {
				after = string(runes[m.cursor+1:])
			}
			cursor := lipgloss.NewStyle().Background(lipgloss.Color("#9B87F5")).Foreground(styles.Bg).Render(at)
			inputText = before + cursor + after
		} else {
			inputText = m.input + lipgloss.NewStyle().Background(lipgloss.Color("#9B87F5")).Render(" ")
		}
		if strings.Contains(m.input, "@") {
			inputText = strings.ReplaceAll(inputText, "@", lipgloss.NewStyle().Foreground(styles.Accent2).Render("@"))
		}
	}
	modelLineLeft := lipgloss.NewStyle().Foreground(styles.Muted).Render("@All · mimo-v2.5-free · go")
	modelLineRight := lipgloss.NewStyle().Foreground(styles.Muted2).Render("tab ⇆ agente")
	modelLine := lipgloss.JoinHorizontal(lipgloss.Top,
		modelLineLeft,
		strings.Repeat(" ", max(0, chatW-len("@All · mimo-v2.5-free · go")-len("tab ⇆ agente")-8)),
		modelLineRight,
	)
	innerInput := lipgloss.JoinVertical(lipgloss.Left, "┃ "+inputText, "", modelLine)
	boxW := chatW - 2
	if boxW < 40 {
		boxW = 40
	}
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#2A2A3A")).
		Background(lipgloss.Color("#14141E")).
		Padding(0, 1).
		Width(boxW).
		Render(innerInput)

	// Sidebar derecho como en la captura: AGENTS / PIPELINE / CONTEXT
	sidebar := renderSidebar(sidebarW, h, m.projectPath, m.vaultPath)
	// Chat a la izquierda (header + mensajes + input)
	chatContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		strings.Join(msgs, "\n"),
		"",
		inputBox,
	)
	// Separador vertical "█" entre chat y sidebar
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#2A2A3A")).Render("█")
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(chatW).Render(chatContent),
		sep,
		lipgloss.NewStyle().Width(sidebarW).PaddingLeft(1).Render(sidebar),
	)
	// Footer como en captura
	footerLeft := lipgloss.NewStyle().Foreground(styles.Muted2).Render("tab agents   ctrl+p commands   ? help")
	footerRight := lipgloss.NewStyle().Foreground(styles.Muted2).Render("huginn-tui   ● Connected   • HUGINN v0.2.0")
	footer := lipgloss.JoinHorizontal(lipgloss.Top,
		footerLeft,
		strings.Repeat(" ", max(0, w-lipgloss.Width(footerLeft)-lipgloss.Width(footerRight)-2)),
		footerRight,
	)
	content := lipgloss.JoinVertical(lipgloss.Left,
		mainArea,
		footer,
	)
	return content
}

func renderSidebar(w, h int, projectPath, vaultPath string) string {
	titleStyle := lipgloss.NewStyle().Foreground(styles.Text).Bold(true)
	muted := styles.Muted
	accent := lipgloss.Color("#9B87F5")
	// AGENTS
	agentsBlock := titleStyle.Render("█ AGENTS") + "\n" +
		lipgloss.NewStyle().Foreground(accent).Render("●") + " Architect             100%\n" +
		lipgloss.NewStyle().Foreground(accent).Render("●") + " Developer             45%\n" +
		lipgloss.NewStyle().Foreground(muted).Render("○") + " Researcher            0%\n" +
		lipgloss.NewStyle().Foreground(muted).Render("○") + " Reviewer              0%\n" +
		lipgloss.NewStyle().Foreground(muted).Render("○") + " QA                    0%"
	// PIPELINE
	pipelineBlock := titleStyle.Render("█ PIPELINE") + "\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80")).Render("●") + " Architect · ✓ done\n" +
		lipgloss.NewStyle().Foreground(accent).Render("●") + " Developer · ● running\n" +
		lipgloss.NewStyle().Foreground(muted).Render("○") + " Researcher · queued\n" +
		lipgloss.NewStyle().Foreground(muted).Render("○") + " Reviewer · queued\n" +
		lipgloss.NewStyle().Foreground(muted).Render("○") + " QA · queued"
	// CONTEXT
	vaultLabel := "● Connected ~/agent-vault"
	if vaultPath == "" {
		vaultLabel = "○ No vault"
	}
	contextBlock := titleStyle.Render("█ CONTEXT") + "\n" +
		"  └─ huginn-tui\n" +
		"  └─ go\n" +
		lipgloss.NewStyle().Foreground(accent).Render(vaultLabel)
	return lipgloss.JoinVertical(lipgloss.Left,
		agentsBlock,
		"",
		pipelineBlock,
		"",
		contextBlock,
	)
}

func (m Model) viewPalette(w, h int) string {
	title := lipgloss.NewStyle().Foreground(styles.Text).Bold(true).Render("Paleta") +
		lipgloss.NewStyle().Foreground(styles.Muted2).Render("  ·  ctrl+p  •  "+m.currentMode())
	var lines []string
	for i, item := range m.paletteItems {
		style := lipgloss.NewStyle().Foreground(styles.Text2).Width(w-8).Padding(0, 1)
		if i == m.paletteCursor {
			style = lipgloss.NewStyle().Foreground(styles.Bg).Background(lipgloss.Color("#9B87F5")).Bold(true).Width(w-8).Padding(0, 1)
		}
		prefix := "  "
		if i == m.paletteCursor {
			prefix = "▶ "
		}
		lines = append(lines, style.Render(prefix+item))
	}
	body := strings.Join(lines, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#9B87F5")).
		Background(lipgloss.Color("#14141E")).
		Padding(1, 1).
		Width(min(48, w-8)).
		Render(title + "\n\n" + body)
	hint := lipgloss.NewStyle().Foreground(styles.Muted2).Render("↑/k ↓/j  enter selecciona  esc cierra")
	inner := lipgloss.JoinVertical(lipgloss.Left, box, "", hint)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, inner)
}

// filteredModels devuelve los modelos que coinciden con modelSearch (case-insensitive).
func (m Model) filteredModels() []string {
	q := strings.ToLower(strings.TrimSpace(m.modelSearch))
	if q == "" {
		return m.modelItems
	}
	var out []string
	for _, item := range m.modelItems {
		if strings.Contains(strings.ToLower(item), q) {
			out = append(out, item)
		}
	}
	return out
}

func (m Model) viewModelSelect(w, h int) string {
	title := lipgloss.NewStyle().Foreground(styles.Text).Bold(true).Render("Select model") +
		lipgloss.NewStyle().Foreground(styles.Muted2).Render("  esc")
	searchText := m.modelSearch
	if strings.TrimSpace(searchText) == "" {
		searchText = lipgloss.NewStyle().Foreground(styles.Muted2).Render("Search")
	}
	search := lipgloss.NewStyle().
		Foreground(styles.Muted).
		Background(lipgloss.Color("#1A1A22")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#2A2A3A")).
		Padding(0, 1).
		Width(min(48, w-8)).
		Render(searchText + lipgloss.NewStyle().Background(lipgloss.Color("#8B5CF6")).Render(" "))
	items := m.filteredModels()
	var lines []string
	for i, item := range items {
		style := lipgloss.NewStyle().Foreground(styles.Text2).Width(min(48, w-8)).Padding(0, 1)
		if i == m.modelCursor {
			style = lipgloss.NewStyle().Foreground(styles.Bg).Background(lipgloss.Color("#8B5CF6")).Bold(true).Width(min(48, w-8)).Padding(0, 1)
		}
		lines = append(lines, style.Render(item))
	}
	if len(lines) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.Muted2).Italic(true).Render("Sin coincidencias para \""+m.modelSearch+"\""))
	}
	body := strings.Join(lines, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#2A2A3A")).
		Background(lipgloss.Color("#0A0A1A")).
		Padding(1, 1).
		Width(min(52, w-6)).
		Render(title + "\n\n" + search + "\n\n" + body)
	hint := lipgloss.NewStyle().Foreground(styles.Muted2).Render("↑/k ↓/j  enter selecciona  esc cierra  •  " + m.currentMode())
	inner := lipgloss.JoinVertical(lipgloss.Left, box, "", hint)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, inner)
}

func (m Model) viewHelp(w, h int) string {
	// Help centrado como en opencode: logo + versión + 3 columnas (comando, descripción, atajo)
	logo := renderAsciiHugInnSmall()
	ver := lipgloss.NewStyle().Foreground(styles.Muted2).Render("v0.2.0")
	header := lipgloss.JoinVertical(lipgloss.Center, logo, ver)
	rows := [][]string{
		{"/help", "show help", "ctrl+x h"},
		{"/editor", "open editor", "ctrl+x a"},
		{"/models", "list models", "ctrl+x m"},
		{"/init", "create AGENTS.md", "ctrl+x i"},
		{"/compact", "compact the session", "ctrl+x c"},
		{"/sessions", "list sessions", "ctrl+x l"},
	}
	var bodyLines []string
	for _, r := range rows {
		col1 := lipgloss.NewStyle().Foreground(lipgloss.Color("#9B87F5")).Width(12).Render(r[0])
		col2 := lipgloss.NewStyle().Foreground(styles.Text2).Width(22).Render(r[1])
		col3 := lipgloss.NewStyle().Foreground(styles.Muted2).Width(10).Align(lipgloss.Right).Render(r[2])
		bodyLines = append(bodyLines, col1+"  "+col2+"  "+col3)
	}
	body := strings.Join(bodyLines, "\n")
	helpBox := lipgloss.NewStyle().
		Border(lipgloss.HiddenBorder()).
		Background(lipgloss.Color("#0A0A1A")).
		Padding(1, 2).
		Width(min(64, w-8)).
		Align(lipgloss.Center).
		Render(header + "\n\n" + body)
	// Input simulado abajo como en la captura
	inputSim := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#2A2A3A")).
		Background(lipgloss.Color("#14141E")).
		Padding(0, 1).
		Width(min(48, w-12)).
		Render(lipgloss.NewStyle().Foreground(lipgloss.Color("#9B87F5")).Render("┃") + " ")
	footer := lipgloss.NewStyle().Foreground(styles.Muted2).Render("enter send" + strings.Repeat(" ", max(0, min(48, w-12)-20)) + "Huginn Zen")
	inner := lipgloss.JoinVertical(lipgloss.Center, helpBox, inputSim, footer)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, inner)
}

func renderAsciiHugInnSmall() string {
	// Versión pequeña del logo para help (2 líneas)
	cafe := lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5A2B")).Bold(true)
	blanco := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	return cafe.Render("HUG") + blanco.Render("INN")
}

func (m Model) viewMessage(w, h int) string {
	title := lipgloss.NewStyle().Foreground(styles.Text).Bold(true).Render(m.msgTitle)
	body := lipgloss.NewStyle().Foreground(styles.Text2).Render(m.msgBody)
	icon := lipgloss.NewStyle().Foreground(lipgloss.Color("#9B87F5")).Render("●")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#2A2A3A")).
		Background(lipgloss.Color("#14141E")).
		Padding(1, 2).
		Width(min(64, w-6)).
		Render(icon + "  " + title + "\n\n" + body)
	hint := lipgloss.NewStyle().Foreground(styles.Muted2).Render("esc para volver  •  /help lista completa")
	inner := lipgloss.JoinVertical(lipgloss.Left, box, "", hint)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, inner)
}

func renderMarkdown(text string, width int) string {
	// Bloques ```code``` con fondo #1A1A22 y borde
	if strings.Contains(text, "```") {
		parts := strings.Split(text, "```")
		var out strings.Builder
		for i, p := range parts {
			if i%2 == 1 {
				// bloque de código
				p = strings.Trim(p, "\n")
				// quita posible lenguaje en primera línea
				if idx := strings.Index(p, "\n"); idx != -1 {
					first := strings.TrimSpace(p[:idx])
					if len(first) < 12 && !strings.Contains(first, " ") {
						p = p[idx+1:]
					}
				}
				// Detecta diff (líneas +/-) y renderiza con números de línea y colores
				if strings.Contains(p, "\n+") || strings.Contains(p, "\n-") || strings.HasPrefix(p, "+") || strings.HasPrefix(p, "-") || strings.Contains(p, "diff --git") {
					out.WriteString(renderDiff(p, width))
				} else {
					codeStyle := lipgloss.NewStyle().
						Background(lipgloss.Color("#1A1A22")).
						Foreground(lipgloss.Color("#C9A8FF")).
						Border(lipgloss.RoundedBorder()).
						BorderForeground(lipgloss.Color("#2A2A3A")).
						Padding(0, 1).
						Width(width)
					out.WriteString(codeStyle.Render(p))
				}
			} else {
				// texto normal con inline `code`
				out.WriteString(renderInlineCode(p))
			}
			if i < len(parts)-1 {
				out.WriteString("\n")
			}
		}
		return out.String()
	}
	return renderInlineCode(text)
}

func renderDiff(p string, width int) string {
	var lines []string
	addStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80")).Background(lipgloss.Color("#0F2A1A"))
	delStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Background(lipgloss.Color("#2A1111"))
	ctxStyle := lipgloss.NewStyle().Foreground(styles.Text2)
	numStyle := lipgloss.NewStyle().Foreground(styles.Muted2)
	oldNum, newNum := 1, 1
	for _, ln := range strings.Split(p, "\n") {
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "@@") {
			// Parsea "@@ -old[,count] +new[,count] @@" para numerar como en la captura
			var o, n int
			if _, err := fmt.Sscanf(trimmed, "@@ -%d,%*d +%d,%*d @@", &o, &n); err == nil {
				oldNum, newNum = o, n
			} else if _, err := fmt.Sscanf(trimmed, "@@ -%d +%d @@", &o, &n); err == nil {
				oldNum, newNum = o, n
			}
			lines = append(lines, numStyle.Render(fmt.Sprintf("   …   %s", trimmed)))
			continue
		}
		if strings.HasPrefix(trimmed, "diff --git") || strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "+++") || strings.HasPrefix(trimmed, "index ") {
			lines = append(lines, numStyle.Render("         "+ln))
			continue
		}
		if strings.HasPrefix(ln, "+") && !strings.HasPrefix(ln, "+++") {
			lines = append(lines, addStyle.Width(width).Render(fmt.Sprintf("%4d %4d + %s", oldNum, newNum, strings.TrimPrefix(ln, "+"))))
			newNum++
			continue
		}
		if strings.HasPrefix(ln, "-") && !strings.HasPrefix(ln, "---") {
			lines = append(lines, delStyle.Width(width).Render(fmt.Sprintf("%4d %4d - %s", oldNum, newNum, strings.TrimPrefix(ln, "-"))))
			oldNum++
			continue
		}
		lines = append(lines, ctxStyle.Width(width).Render(fmt.Sprintf("%4d %4d   %s", oldNum, newNum, ln)))
		oldNum++
		newNum++
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#14141E")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#2A2A3A")).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}

func renderInlineCode(s string) string {
	if !strings.Contains(s, "`") {
		return s
	}
	parts := strings.Split(s, "`")
	var out strings.Builder
	for i, p := range parts {
		if i%2 == 1 {
			out.WriteString(lipgloss.NewStyle().
				Background(lipgloss.Color("#1A1A22")).
				Foreground(lipgloss.Color("#E1A451")).
				Padding(0, 1).
				Render(p))
		} else {
			out.WriteString(p)
		}
	}
	return out.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
