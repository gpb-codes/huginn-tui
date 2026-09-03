package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/atotto/clipboard"

	"huginn/internal/bootstrap"
	agentpkg "huginn/internal/domain/agent"
	"huginn/internal/domain/execution"
	vaultpkg "huginn/internal/domain/vault"
	infravault "huginn/internal/infrastructure/vault"
	tuicomp "huginn/internal/tui/components"
)

const VERSION = "v0.2.0"

// ---------- clipboard fallback ----------
var clipboardFallback string

func copyToClipboard(s string) {
	clipboardFallback = s
	_ = clipboard.WriteAll(s)
}
func pasteFromClipboard() string {
	if t, err := clipboard.ReadAll(); err == nil && t != "" {
		return t
	}
	return clipboardFallback
}

// ---------- palette — HUGINN Bronze/Gold ----------
var (
	colPanel   = lipgloss.Color("#4D3217")
	colPanel2  = lipgloss.Color("#4D3217")
	colBorder  = lipgloss.Color("#634924")
	colBorder2 = lipgloss.Color("#634924")
	colText    = lipgloss.Color("#FBE7AE")
	colText2   = lipgloss.Color("#9D8E69")
	colMuted   = lipgloss.Color("#634924")
	colMuted2  = lipgloss.Color("#9D8E69")
	colAccent  = lipgloss.Color("#E1A451")
	colPurple  = lipgloss.Color("#CD8D38")
	colWhite   = lipgloss.Color("#FBE7AE")
	colSuccess = lipgloss.Color("#E1A451")
	colRaven   = lipgloss.Color("#634924")
	colWarn    = lipgloss.Color("#CD8D38")
	colError   = lipgloss.Color("#CD8D38")
)

// ---------- agents ----------
type agentStatus int

const (
	statusDone agentStatus = iota
	statusWorking
	statusWaiting
	statusTesting
)

type agent struct {
	name   string
	role   string
	status agentStatus
	pct    int
}

func (s agentStatus) label() string {
	switch s {
	case statusDone:
		return "Completed"
	case statusWorking:
		return "Working"
	case statusTesting:
		return "Running Tests"
	default:
		return "Queued"
	}
}

func (s agentStatus) color() color.Color {
	switch s {
	case statusDone:
		return colSuccess
	case statusWorking:
		return colAccent
	case statusTesting:
		return lipgloss.Color("#E1A451")
	default:
		return colMuted
	}
}

// ---------- backend agents (python parity) ----------
type backendAgent struct {
	Name        string
	Description string
	Command     string // empty = always online (ChatGPT)
}

var backendAgents = []backendAgent{
	{"ChatGPT", "central intelligence manager", ""},
	{"OpenCode", "terminal AI coding agent", "opencode"},
	{"KiloCode", "AI coding agent and workflow", "kilocode"},
}

func commandAvailable(cmd string) bool {
	if cmd == "" {
		return true
	}
	_, err := exec.LookPath(cmd)
	return err == nil
}

func openVault() error {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	candidates := []string{dir, filepath.Join(dir, ".."), "."}
	var root string
	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if _, err := os.Stat(filepath.Join(abs, "README.md")); err == nil {
			root = abs
			break
		}
		if _, err := os.Stat(filepath.Join(abs, ".git")); err == nil {
			root = abs
			break
		}
	}
	if root == "" {
		cwd, _ := os.Getwd()
		root = cwd
		if strings.HasSuffix(strings.ReplaceAll(cwd, "\\", "/"), "tui/huginn-tui") {
			root = filepath.Join(cwd, "..", "..")
		} else if strings.HasSuffix(strings.ReplaceAll(cwd, "\\", "/"), "huginn-tui") {
			root = filepath.Join(cwd, "..", "..")
		}
		root, _ = filepath.Abs(root)
	}

	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", root).Start()
	case "darwin":
		return exec.Command("open", root).Start()
	default:
		return exec.Command("xdg-open", root).Start()
	}
}

// ---------- commands ----------
type Command struct {
	Name        string
	Description string
	Shortcut    string
}

var commands = []Command{
	{"/help", "show help", "ctrl+x h"},
	{"/research", "manage research pipeline", "ctrl+x r"},
	{"/knowledge", "browse knowledge base", "ctrl+x k"},
	{"/graph", "show knowledge graph", "graph"},
	{"/agents", "manage connected agents", "ctrl+x a"},
	{"/search", "search in knowledge", "ctrl+x s"},
	{"/status", "show system status", "ctrl+x t"},
	{"/tools", "open tools & integrations", "ctrl+x o"},
	{"/settings", "configure huginn", "ctrl+x c"},
	{"/vault", "open vault in explorer", "ctrl+x v"},
	{"/chat", "talk with agents", "ctrl+x m"},
	{"/mcp", "MCP servers & tools", "ctrl+x m"},
	{"/lsp", "language servers (LSP)", "ctrl+x l"},
	{"/peers", "peer PCs & vault sync", "ctrl+x p"},
	{"/servers", "servers hub (MCP/LSP/Peers)", "ctrl+x s"},
	{"/exit", "exit huginn", "ctrl+x q"},
	{"/sessions", "manage sessions", "ctrl+x n"},
}

// ---------- settings tree (⚙ Settings) ----------
type settingsSection struct {
	Name  string
	Items []string
}

var huginnSettings = []settingsSection{
	{Name: "General", Items: []string{"Profile", "Language", "Timezone", "Startup behavior"}},
	{Name: "Appearance", Items: []string{"Theme", "Accent color", "Font size", "Chat density"}},
	{Name: "AI", Items: []string{"Providers", "Models", "Default model", "Reasoning", "Token limits"}},
	{Name: "Agents", Items: []string{"Default agent", "Agent permissions", "Agent routing", "Max agents", "Execution limits"}},
	{Name: "Memory", Items: []string{"Memory enabled", "Persistent memory", "Auto-save", "Context retrieval", "Knowledge Graph"}},
	{Name: "Tools", Items: []string{"Web", "Files", "GitHub", "Terminal", "MCP"}},
	{Name: "Vault", Items: []string{"Vault connection", "Default vault", "Sync", "Indexing", "Embeddings"}},
	{Name: "Advanced", Items: []string{"API", "Logs", "Debug", "Experimental"}},
}

var settingsValues = map[string]string{
	"Profile": "gabriel", "Language": "Español", "Timezone": "UTC-3", "Startup behavior": "Restore session",
	"Theme": "Dark", "Accent color": "#CD8D38", "Font size": "14px", "Chat density": "Comfortable",
	"Providers": "OpenCode Zen", "Models": "5 configured", "Default model": "mimo-v2.5-free", "Reasoning": "Enabled", "Token limits": "8k",
	"Default agent": "OpenCode", "Agent permissions": "Ask", "Agent routing": "Auto", "Max agents": "4", "Execution limits": "90s",
	"Memory enabled": "On", "Persistent memory": "On", "Auto-save": "On", "Context retrieval": "Hybrid", "Knowledge Graph": "Enabled",
	"Web": "Enabled", "Files": "Enabled", "GitHub": "Connected", "Terminal": "Enabled", "MCP": "5 servers",
	"Vault connection": "Connected", "Default vault": "~/huginn-vault", "Sync": "Auto", "Indexing": "On", "Embeddings": "Enabled",
	"API": "Local", "Logs": "Verbose", "Debug": "Off", "Experimental": "Off",
}

// ---------- MCP / LSP / PEERS — sistema de servidores ----------
type mcpTransport string

const (
	mcpStdio     mcpTransport = "stdio"
	mcpSSE       mcpTransport = "sse"
	mcpWebSocket mcpTransport = "websocket"
)

type mcpServer struct {
	Name      string
	Command   string
	Transport mcpTransport
	Status    string // Connected, Connecting, Needs auth, Error, Disabled
	Tools     int
	Latency   string
	ErrorMsg  string
	Enabled   bool
}

var mcpServers = []mcpServer{
	{"filesystem", "npx @modelcontextprotocol/server-filesystem", mcpStdio, "Connected", 8, "12ms", "", true},
	{"github", "npx @modelcontextprotocol/server-github", mcpStdio, "Connected", 6, "34ms", "", true},
	{"memory", "npx @modelcontextprotocol/server-memory", mcpStdio, "Connected", 4, "8ms", "", true},
	{"playwright", "npx @modelcontextprotocol/server-playwright", mcpStdio, "Connected", 7, "45ms", "", true},
	{"sequential", "node sequential-thinking", mcpStdio, "Connected", 1, "9ms", "", true},
	{"notion", "npx @modelcontextprotocol/server-notion", mcpSSE, "Needs auth", 0, "—", "Missing NOTION_API_KEY", true},
	{"vault-sync", "huginn-vault-sync", mcpWebSocket, "Disabled", 3, "—", "", false},
}

type lspServer struct {
	Language    string
	ServerName  string
	Command     string
	Status      string // Running, Stopped, Error, Installing
	Root        string
	Diagnostics int
	Version     string
	Enabled     bool
}

var lspServers = []lspServer{
	{"Go", "gopls", "gopls", "Running", "./", 0, "v0.15.2", true},
	{"TypeScript", "tsserver", "typescript-language-server --stdio", "Running", "web/", 2, "4.3.3", true},
	{"Python", "pyright", "pyright-langserver --stdio", "Stopped", "py/", 0, "1.1.350", false},
	{"Rust", "rust-analyzer", "rust-analyzer", "Running", "./", 1, "2024-08-28", true},
	{"Lua", "lua_ls", "lua-language-server", "Error", "lua/", 0, "3.7.4", true},
}

type peerStatus string

const (
	peerOnline  peerStatus = "Online"
	peerOffline peerStatus = "Offline"
	peerPairing peerStatus = "Pairing"
	peerSyncing peerStatus = "Syncing"
)

type peerServer struct {
	ID        string
	Hostname  string
	IP        string
	Status    peerStatus
	VaultSync string // Synced, Syncing 42%, Offline, Error
	Latency   string
	LastSeen  string
	Paired    bool
	Trusted   bool
}

var peerServers = []peerServer{
	{"peer_01JABC", "thinkpad-x1", "192.168.1.42", peerOnline, "Synced", "18ms", "now", true, true},
	{"peer_02JDEF", "macbook-pro", "192.168.1.88", peerSyncing, "Syncing 64%", "42ms", "2m ago", true, true},
	{"peer_03JGHI", "desk-linux", "10.0.0.12", peerOffline, "Offline", "—", "1h ago", true, false},
	{"peer_04JJKL", "huginn-remote", "100.64.1.5", peerPairing, "Pairing…", "—", "—", false, false},
}

var localPeerID = "peer_LOCAL_7F3A"
var localHostname = func() string {
	h, _ := os.Hostname()
	if h == "" {
		return "huginn-local"
	}
	return h
}()

// ---------- model ----------
type mode int

const (
	modeInput mode = iota
	modeRunning
	modeHelp
	modeAgents
	modeStatus
	modeMessage
	modeChat
	modeSettings
	modeServers
	modeGraph
	modeVault
)

// chat con todos los agentes
type chatMsg struct {
	From   string
	Text   string
	IsUser bool
	Time   string
}

var chatAgents = []string{"All", "ChatGPT", "OpenCode", "Mimo Code", "Kilo Code", "Muse Code"}

// menciones nativas estilo opencode/kilo: @chatgpt @opencode @kilo @mimo @muse @all
var mentionMap = map[string]string{
	"@all":         "All",
	"@chatgpt":     "ChatGPT",
	"@opencode":    "OpenCode",
	"@kilo":        "Kilo Code",
	"@kilocode":    "Kilo Code",
	"@mimo":        "Mimo Code",
	"@mimo_code":   "Mimo Code",
	"@mimocode":    "Mimo Code",
	"@muse":        "Muse Code",
	"@claude":      "Muse Code",
	"@claudecodes": "Muse Code",
}

func parseMentions(text string) []string {
	low := strings.ToLower(text)
	var found []string
	seen := map[string]bool{}
	for mention, agent := range mentionMap {
		if strings.Contains(low, mention) {
			if !seen[agent] {
				found = append(found, agent)
				seen[agent] = true
			}
		}
	}
	return found
}

func renderMentionSuggestions(partial string) []string {
	low := strings.ToLower(partial)
	var sug []string
	for m := range mentionMap {
		if strings.HasPrefix(m, low) {
			sug = append(sug, m)
		}
	}
	return sug
}

func highlightMentions(s string) string {
	res := s
	for m := range mentionMap {
		lowRes := strings.ToLower(res)
		lowM := strings.ToLower(m)
		if idx := strings.Index(lowRes, lowM); idx != -1 {
			styled := lipgloss.NewStyle().Foreground(colAccent).Bold(true).Render(m)
			var b strings.Builder
			i := 0
			for {
				lowRes = strings.ToLower(res[i:])
				idx = strings.Index(lowRes, lowM)
				if idx == -1 {
					b.WriteString(res[i:])
					break
				}
				b.WriteString(res[i : i+idx])
				b.WriteString(styled)
				i += idx + len(m)
				if i >= len(res) {
					break
				}
			}
			res = b.String()
		}
	}
	return res
}

func updateChatSuggests(m *model) {
	if m.chatCursor == 0 {
		m.chatSuggests = nil
		m.chatSuggestIdx = 0
		return
	}
	input := m.chatInput[:m.chatCursor]
	atIdx := strings.LastIndex(input, "@")
	if atIdx == -1 {
		m.chatSuggests = nil
		m.chatSuggestIdx = 0
		return
	}
	partial := input[atIdx:m.chatCursor]
	if strings.Contains(partial, " ") || len(partial) > 14 {
		m.chatSuggests = nil
		m.chatSuggestIdx = 0
		return
	}
	sugs := renderMentionSuggestions(partial)
	if len(sugs) == 0 {
		m.chatSuggests = nil
		m.chatSuggestIdx = 0
		return
	}
	m.chatSuggests = sugs
	if m.chatSuggestIdx >= len(sugs) {
		m.chatSuggestIdx = 0
	}
}

func applyChatAutocomplete(m *model) bool {
	if len(m.chatSuggests) == 0 {
		return false
	}
	input := m.chatInput[:m.chatCursor]
	atIdx := strings.LastIndex(input, "@")
	if atIdx == -1 {
		return false
	}
	chosen := m.chatSuggests[m.chatSuggestIdx]
	m.chatInput = m.chatInput[:atIdx] + chosen + " " + m.chatInput[m.chatCursor:]
	m.chatCursor = atIdx + len(chosen) + 1
	m.chatSuggests = nil
	m.chatSuggestIdx = 0
	return true
}

func updateMainSuggests(m *model) {
	if m.cursor == 0 {
		m.chatSuggests = nil
		m.chatSuggestIdx = 0
		return
	}
	input := m.input[:m.cursor]
	atIdx := strings.LastIndex(input, "@")
	if atIdx == -1 {
		m.chatSuggests = nil
		m.chatSuggestIdx = 0
		return
	}
	partial := input[atIdx:m.cursor]
	if strings.Contains(partial, " ") || len(partial) > 14 {
		m.chatSuggests = nil
		m.chatSuggestIdx = 0
		return
	}
	sugs := renderMentionSuggestions(partial)
	if len(sugs) == 0 {
		m.chatSuggests = nil
		m.chatSuggestIdx = 0
		return
	}
	m.chatSuggests = sugs
	if m.chatSuggestIdx >= len(sugs) {
		m.chatSuggestIdx = 0
	}
}

func applyMainAutocomplete(m *model) bool {
	if len(m.chatSuggests) == 0 {
		return false
	}
	input := m.input[:m.cursor]
	atIdx := strings.LastIndex(input, "@")
	if atIdx == -1 {
		return false
	}
	chosen := m.chatSuggests[m.chatSuggestIdx]
	m.input = m.input[:atIdx] + chosen + " " + m.input[m.cursor:]
	m.cursor = atIdx + len(chosen) + 1
	m.chatSuggests = nil
	m.chatSuggestIdx = 0
	return true
}

type model struct {
	input            string
	cursor           int
	width            int
	height           int
	mode             mode
	agents           []agent
	logs             []string
	quitting         bool
	msgTitle         string
	msgBody          string
	chatTarget       int
	chatInput        string
	chatCursor       int
	chatHistory      []chatMsg
	chatSuggests     []string
	chatSuggestIdx   int
	chatInputHistory []string
	chatHistIdx      int
	chatScroll       int
	// settings
	settingsCursor   int
	settingsExpanded []bool
	settingsOffset   int
	// servers hub (MCP/LSP/Peers)
	serversTab    int // 0=MCP 1=LSP 2=Peers
	serversCursor int
	serversOffset int
	serversMsg    string
	// CLI context (no duplica Vault: solo referencia)
	projectPath string
	projectName string
	vaultPath   string
	vaultOK     bool
	pkgManager  string
	// graph
	graphCursor int
	// vault wizard
	vaultWizardStep   int
	vaultWizardInput  string
	vaultWizardCursor int
	vaultWizardData   vaultConfig
	// orchestrator app — Registry + Context + Trace
	orchApp *bootstrap.App
	// mascot state
	mascotState string
}

type vaultConfig struct {
	Path       string
	Name       string
	Purpose    string // personal, equipo, proyecto
	Sync       string // Auto, Manual, Off
	Indexing   string // On, Off
	Embeddings string // OpenAI, Local, Ollama, Disabled
	VaultType  string // knowledge, memory, mixed
}

var vaultWizardQuestions = []struct {
	Title   string
	Key     string
	Prompt  string
	Options []string
	Default string
}{
	{"Vault path", "Path", "¿Dónde guardar el vault?", nil, "~/agent-vault"},
	{"Vault name", "Name", "Nombre del vault", nil, "agent-vault"},
	{"Propósito", "Purpose", "¿Para qué usarás el vault?", []string{"personal", "equipo", "proyecto"}, "proyecto"},
	{"Tipo de vault", "VaultType", "¿Qué guardará principalmente?", []string{"knowledge", "memory", "mixed"}, "mixed"},
	{"Sync", "Sync", "¿Sincronización entre PCs?", []string{"Auto", "Manual", "Off"}, "Auto"},
	{"Indexing", "Indexing", "¿Indexado automático de archivos?", []string{"On", "Off"}, "On"},
	{"Embeddings", "Embeddings", "¿Proveedor de embeddings?", []string{"OpenAI", "Local", "Ollama", "Disabled"}, "Local"},
}

var graphNodes = []struct {
	Label string
	Type  string
	Desc  string
	Color color.Color
}{
	{"Agent", "agent", "Coordinador principal", lipgloss.Color("#E1A451")},
	{"Memory", "memory", "Memoria persistente 4.2k tokens", lipgloss.Color("#CD8D38")},
	{"Context", "context", "Contexto activo del proyecto", lipgloss.Color("#E1A451")},
	{"Project", "project", "huginn-tui • Go/Bubbletea", lipgloss.Color("#E1A451")},
	{"embeddings", "file", "src/memory/embeddings.db", lipgloss.Color("#9D8E69")},
	{"session", "file", "sessions/sess_01JABC.json", lipgloss.Color("#9D8E69")},
	{"Task", "task", "Implement auth with JWT", lipgloss.Color("#E1A451")},
	{"Knowledge", "knowledge", "Grafo 23 relaciones", lipgloss.Color("#CD8D38")},
	{"recent.json", "file", "Memoria reciente", lipgloss.Color("#9D8E69")},
	{"graph.json", "file", "Grafo serializado", lipgloss.Color("#CD8D38")},
	{"Huginn", "agent", "Orquestador local", lipgloss.Color("#E1A451")},
	{"vault", "vault", "~/agent-vault • Synced", lipgloss.Color("#E1A451")},
	{"index.ts", "file", "src/knowledge/index.ts", lipgloss.Color("#CD8D38")},
	{"config.yaml", "config", "Vault config", lipgloss.Color("#9D8E69")},
	{"context.md", "context", "Contexto markdown", lipgloss.Color("#E1A451")},
	{"reviewer", "agent", "Code review", lipgloss.Color("#E1A451")},
	{"architect", "agent", "System design", lipgloss.Color("#E1A451")},
	{"developer", "agent", "Implementation", lipgloss.Color("#E1A451")},
	{"researcher", "agent", "Research", lipgloss.Color("#CD8D38")},
}

func initialModel() model {
	m := model{
		chatTarget: 0,
		agents: []agent{
			{"OpenCode", "Architecture & Backend", statusWaiting, 0},
			{"Mimo Code", "Implementation", statusWaiting, 0},
			{"Kilo Code", "Code Review", statusWaiting, 0},
			{"Muse Code", "Testing & Specialized Tasks", statusWaiting, 0},
		},
		chatHistory: []chatMsg{
			{From: "Hugin", Text: "Chat iniciado. Menciona con @chatgpt @opencode @kilo @mimo @muse @all  •  o usa Tab/1-6", IsUser: false, Time: time.Now().Format("15:04")},
		},
		settingsExpanded: make([]bool, len(huginnSettings)),
	}
	// por defecto expande General y Appearance para mostrar el árbol
	if len(m.settingsExpanded) > 0 {
		m.settingsExpanded[0] = true
		m.settingsExpanded[1] = true
	}
	return m
}

func initialModelWithContext(projectPath, prompt string) model {
	m := initialModel()
	abs, _ := filepath.Abs(projectPath)
	m.projectPath = abs
	m.projectName = filepath.Base(abs)
	// detect pkg manager y vault sin duplicar lógica de Vault (solo referencia)
	m.pkgManager = detectPackageManager(abs)
	m.vaultPath, m.vaultOK = resolveVaultPath()
	// bootstrap orchestrator: Registry + ContextManager + Trace
	if app, err := bootstrap.New(abs, m.vaultPath); err == nil {
		m.orchApp = app
	}
	m.mascotState = "idle"
	// si hay prompt directo, lo inyecta como tarea inicial para el orquestador
	if strings.TrimSpace(prompt) != "" {
		trim := strings.TrimSpace(prompt)
		m.chatHistory = append(m.chatHistory, chatMsg{From: "You", Text: trim, IsUser: true, Time: time.Now().Format("15:04")})
		m.mode = modeRunning
		m.logs = []string{
			"huginn      project: " + abs,
			"huginn      vault: " + m.vaultPath + (func() string {
				if m.vaultOK {
					return " (connected)"
				}
				return " (not found — usando memoria local)"
			})(),
			"huginn      task received: " + trim,
			"huginn      orchestrator → Planner → Coder → Researcher",
		}
		m.agents[0].status = statusWorking
	}
	return m
}

func (m model) Init() tea.Cmd {
	// si viene con prompt, arranca tick de orquestación inmediatamente
	if m.mode == modeRunning {
		return tick()
	}
	return nil
}

type tickMsg struct{}

// handleCommand migra la lógica de huginn.py:execute() a Go
func handleCommand(raw string, m model) (model, tea.Cmd, bool) {
	cmd := strings.TrimSpace(strings.ToLower(raw))
	if cmd == "" {
		return m, nil, false
	}
	if cmd == "q" || cmd == "quit" || cmd == "exit" || cmd == "/quit" || cmd == "/exit" {
		m.quitting = true
		return m, tea.Quit, true
	}
	if cmd == "/help" || cmd == "help" {
		m.mode = modeHelp
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if cmd == "/agents" || cmd == "agents" || cmd == "/connect" || cmd == "connect" {
		m.mode = modeAgents
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if cmd == "/status" || cmd == "status" {
		m.mode = modeStatus
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if cmd == "/vault" || cmd == "vault" {
		// wizard interactivo — no abre directo, pregunta paso a paso
		m.mode = modeVault
		m.vaultWizardStep = 0
		m.vaultWizardCursor = 0
		m.vaultWizardData = vaultConfig{
			Path:       "~/agent-vault",
			Name:       "agent-vault",
			Purpose:    "proyecto",
			VaultType:  "mixed",
			Sync:       "Auto",
			Indexing:   "On",
			Embeddings: "Local",
		}
		// si ya hay vault detectado, precarga
		if m.vaultPath != "" {
			m.vaultWizardData.Path = m.vaultPath
			m.vaultWizardData.Name = filepath.Base(m.vaultPath)
		}
		m.vaultWizardInput = vaultWizardQuestions[0].Default
		if m.vaultWizardData.Path != "" {
			m.vaultWizardInput = m.vaultWizardData.Path
		}
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if strings.HasPrefix(cmd, "/research") {
		m.mode = modeMessage
		m.msgTitle = "Research"
		m.msgBody = "Research workspace selected.\n\nResearch provider integration is next.\nManaged by ChatGPT — synthesis & coordination."
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if strings.HasPrefix(cmd, "/knowledge") {
		m.mode = modeMessage
		m.msgTitle = "Knowledge"
		m.msgBody = "Knowledge workspace selected.\n\nSearch/index integration is next.\nVault: 00-inbox → 01-knowledge → 02-research"
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if strings.HasPrefix(cmd, "/search") {
		q := strings.TrimSpace(raw[len("/search"):])
		if q == "" {
			q = "(no query)"
		}
		m.mode = modeMessage
		m.msgTitle = "Search"
		m.msgBody = fmt.Sprintf("Knowledge search: %s\n\nVault search integration is next.", q)
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if cmd == "/graph" || cmd == "graph" {
		m.mode = modeGraph
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if cmd == "/tools" {
		m.mode = modeMessage
		m.msgTitle = "Tools"
		m.msgBody = "Tools & integrations workspace selected.\n\n• 03-tools/  • OpenCode  • KiloCode  • ChatGPT"
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if cmd == "/settings" {
		m.mode = modeSettings
		m.input = ""
		m.cursor = 0
		if m.settingsExpanded == nil || len(m.settingsExpanded) != len(huginnSettings) {
			m.settingsExpanded = make([]bool, len(huginnSettings))
			if len(m.settingsExpanded) > 0 {
				m.settingsExpanded[0] = true
				m.settingsExpanded[1] = true
			}
		}
		m.settingsCursor = 0
		m.settingsOffset = 0
		return m, nil, false
	}
	if cmd == "/sessions" {
		m.mode = modeMessage
		m.msgTitle = "Sessions"
		m.msgBody = "No persisted HUGINN sessions yet.\n\nSessions will be stored in the vault."
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if cmd == "/mcp" {
		m.mode = modeServers
		m.serversTab = 0
		m.serversCursor = 0
		m.serversOffset = 0
		m.input = ""
		m.cursor = 0
		return m, tickServers(), false
	}
	if cmd == "/lsp" {
		m.mode = modeServers
		m.serversTab = 1
		m.serversCursor = 0
		m.serversOffset = 0
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if cmd == "/peers" || cmd == "/servers" || cmd == "/server" {
		m.mode = modeServers
		if cmd == "/servers" || cmd == "/server" {
			m.serversTab = 0
		} else {
			m.serversTab = 2
		}
		m.serversCursor = 0
		m.serversOffset = 0
		m.input = ""
		m.cursor = 0
		return m, tickServers(), false
	}
	if cmd == "/chat" || cmd == "/talk" || cmd == "chat" || cmd == "talk" {
		m.mode = modeChat
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	if strings.HasPrefix(raw, "/") {
		m.mode = modeMessage
		m.msgTitle = "Unknown command"
		m.msgBody = fmt.Sprintf("Unknown command: %s\n\nType /help for available commands.", raw)
		m.input = ""
		m.cursor = 0
		return m, nil, false
	}
	return m, nil, false
}

type agentReplyMsg struct {
	Agent string
	Text  string
	Err   error
}

func callAgentCmd(agent, prompt string, app *bootstrap.App) tea.Cmd {
	return func() tea.Msg {
		// fallback inmediato si opencode no está instalado — usa mock local
		if agent == "ChatGPT" {
			time.Sleep(500 * time.Millisecond)
			return agentReplyMsg{Agent: agent, Text: mockReply(agent, prompt)}
		}

		// si hay app, usa Registry + Context + Trace
		if app != nil && app.Registry != nil {
			return dispatchViaRegistry(agent, prompt, app)
		}

		// fallback: exec directo (pre-orchestrator)
		return dispatchViaExec(agent, prompt)
	}
}

// dispatchViaRegistry — Fase 7: Registry.SelectProvider → ContextManager.Load → Provider.Invoke → trace.Append
func dispatchViaRegistry(agentName, prompt string, app *bootstrap.App) agentReplyMsg {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// 1) resolver provider por nombre de agente del chat
	providerName := mapAgentToProvider(agentName)
	p, ok := app.Registry.GetProvider(providerName)
	if !ok || !p.Available(ctx) {
		// fallback a opencode si no disponible
		if op, ok2 := app.Registry.GetProvider("opencode"); ok2 && op.Available(ctx) {
			p = op
		} else {
			p, _ = app.Registry.GetProvider("chatgpt")
			if p == nil {
				return agentReplyMsg{Agent: agentName, Text: mockReply(agentName, prompt), Err: fmt.Errorf("no provider available")}
			}
		}
	}

	// 2) contexto del vault
	var vaultCtx string
	if app.ContextManager != nil {
		vaultCtx, _ = app.ContextManager.Load()
	}
	actx := agentpkg.AgentContext{
		VaultPath:   app.Trace.VaultPath(),
		ProjectPath: "",
		Memory:      []string{vaultCtx},
	}

	// 3) invocar provider
	resp, err := p.Invoke(ctx, agentpkg.ProviderRequest{
		Prompt:  prompt,
		Context: actx,
		Meta:    map[string]string{"agent": agentName},
	})

	// 4) trazabilidad
	if app.Trace != nil {
		rec := execution.Record{
			ExecutionID: fmt.Sprintf("chat-%d", start.UnixMilli()),
			Agent:       agentName,
			Provider:    p.Name(),
			TaskID:      fmt.Sprintf("chat-%d", start.UnixMilli()),
			TaskType:    "chat",
			Input:       prompt,
			Status:      statusFromErr(err),
			StartedAt:   start,
		}
		_ = app.Trace.Append(rec)
	}

	if err != nil {
		return agentReplyMsg{Agent: agentName, Text: fmt.Sprintf("Error %s: %v", agentName, err), Err: err}
	}

	text := strings.TrimSpace(resp.Content)
	if len(text) > 800 {
		text = text[:800] + " …"
	}
	return agentReplyMsg{Agent: agentName, Text: text}
}

// dispatchViaExec — fallback pre-orchestrator (exec.Command directo)
func dispatchViaExec(agent, prompt string) agentReplyMsg {
	opencodeBin := "opencode"
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("opencode.cmd"); err == nil {
			opencodeBin = "opencode.cmd"
		}
	}
	if !commandAvailable(opencodeBin) && !commandAvailable("opencode") && !commandAvailable("opencode.cmd") {
		time.Sleep(400 * time.Millisecond)
		return agentReplyMsg{Agent: agent, Text: mockReply(agent, prompt)}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	var model string
	switch agent {
	case "ChatGPT":
		model = "opencode/mimo-v2.5-free"
	case "OpenCode":
		model = "opencode/mimo-v2.5-free"
	case "Mimo Code":
		model = "opencode/mimo-v2.5-free"
	case "Kilo Code":
		model = "opencode/nemotron-3.5-lightning-free"
	case "Muse Code":
		model = "opencode/muse-spark-1.2-contributor-free"
	default:
		model = "opencode/mimo-v2.5-free"
	}
	cmd := exec.CommandContext(ctx, opencodeBin, "run", "-m", model, "--format", "json", prompt)
	cmd.Dir, _ = os.Getwd()
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return agentReplyMsg{Agent: agent, Text: "⏱ timeout 90s — el agente no respondió a tiempo", Err: ctx.Err()}
	}
	raw := strings.TrimSpace(string(out))
	if err != nil && raw == "" {
		return agentReplyMsg{Agent: agent, Text: fmt.Sprintf("Error %s: %v", agent, err), Err: err}
	}
	text := ""
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var obj map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(line), &obj); jsonErr == nil {
			if typ, ok := obj["type"].(string); ok && typ == "text" {
				if part, ok := obj["part"].(map[string]interface{}); ok {
					if t, ok := part["text"].(string); ok {
						text += t
					}
				}
			}
			if typ, ok := obj["type"].(string); ok && typ == "error" {
				if e, ok := obj["error"].(map[string]interface{}); ok {
					if d, ok := e["data"].(map[string]interface{}); ok {
						if m, ok := d["message"].(string); ok {
							text = "Error: " + m
						}
					}
				}
			}
		}
	}
	if text == "" {
		text = raw
		text = strings.ReplaceAll(text, "\x1b[0m", "")
		text = strings.ReplaceAll(text, "\x1b[90m", "")
		text = strings.TrimSpace(text)
		if text == "" {
			text = fmt.Sprintf("(sin respuesta %s — raw: %d bytes)", agent, len(raw))
		}
	}
	text = strings.TrimSpace(text)
	if len(text) > 800 {
		text = text[:800] + " …"
	}
	return agentReplyMsg{Agent: agent, Text: text, Err: err}
}

func mapAgentToProvider(agent string) string {
	switch agent {
	case "ChatGPT":
		return "chatgpt"
	case "OpenCode":
		return "opencode"
	case "Mimo Code":
		return "muse-spark"
	case "Kilo Code":
		return "kilo"
	case "Muse Code":
		return "muse-spark"
	default:
		return "opencode"
	}
}

func statusFromErr(err error) string {
	if err != nil {
		return "error"
	}
	return "ok"
}

func mockReply(agent, userText string) string {
	low := strings.ToLower(userText)
	switch agent {
	case "ChatGPT":
		if strings.Contains(low, "huginn") || strings.Contains(low, "vault") {
			return "HUGINN es la memoria compartida. Puedo sintetizar y organizar tu vault (00-inbox → 01-knowledge)."
		}
		if strings.Contains(low, "invest") {
			return "Puedo liderar la investigación: verifico fuentes primarias, separo hechos de análisis y propongo siguientes pasos."
		}
		return "Recibido. Como manager, coordinaré la tarea y la descompondré para los ejecutores."
	case "OpenCode":
		return "OpenCode listo — puedo implementar, refactorizar y testear en el repo. Decime el stack y arranco."
	case "Mimo Code":
		return "Mimo Code acá — me encargo de la implementación directa. Pasame los requisitos y codeamos."
	case "Kilo Code":
		return "Kilo Code — reviso y ejecuto workflows. Puedo hacer code review y automatizar."
	case "Muse Code":
		return "Muse Code — testing & QA. Armo tests, corro suites y reporto coverage."
	default:
		return "Agente listo para ayudar."
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		if m.mode == modeChat {
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "esc":
				if len(m.chatSuggests) > 0 {
					m.chatSuggests = nil
					return m, nil
				}
				m.mode = modeInput
				return m, nil
			case "enter":
				if len(m.chatSuggests) > 0 {
					applyChatAutocomplete(&m)
					return m, nil
				}
				if strings.TrimSpace(m.chatInput) != "" {
					userText := m.chatInput
					now := time.Now().Format("15:04")
					m.chatHistory = append(m.chatHistory, chatMsg{From: "You", Text: userText, IsUser: true, Time: now})
					m.chatInputHistory = append(m.chatInputHistory, userText)
					m.chatHistIdx = len(m.chatInputHistory)
					mentions := parseMentions(userText)
					var targets []string
					if len(mentions) > 0 {
						for _, mm := range mentions {
							if mm == "All" {
								targets = chatAgents[1:]
								break
							}
						}
						if len(targets) == 0 {
							targets = mentions
						}
					} else {
						t := chatAgents[m.chatTarget]
						if t == "All" {
							targets = chatAgents[1:]
						} else {
							targets = []string{t}
						}
					}
				var cmds []tea.Cmd
					m.mascotState = "thinking"
					for _, ag := range targets {
						m.chatHistory = append(m.chatHistory, chatMsg{From: ag, Text: "… escribiendo (real) — llamando " + ag + "…", IsUser: false, Time: now})
					cmds = append(cmds, callAgentCmd(ag, userText, m.orchApp))
				}
				m.chatInput = ""
				m.chatCursor = 0
					m.chatSuggests = nil
					if len(cmds) > 0 {
						return m, tea.Batch(cmds...)
					}
				}
				return m, nil
			case "tab":
				if len(m.chatSuggests) > 0 {
					applyChatAutocomplete(&m)
					return m, nil
				}
				m.chatTarget = (m.chatTarget + 1) % len(chatAgents)
				return m, nil
			case "shift+tab":
				if len(m.chatSuggests) > 0 {
					applyChatAutocomplete(&m)
					return m, nil
				}
				m.chatTarget = (m.chatTarget - 1 + len(chatAgents)) % len(chatAgents)
				return m, nil
			case "up":
				if len(m.chatSuggests) > 0 {
					m.chatSuggestIdx = (m.chatSuggestIdx - 1 + len(m.chatSuggests)) % len(m.chatSuggests)
					return m, nil
				}
				if len(m.chatInputHistory) > 0 && m.chatCursor == len(m.chatInput) {
					if m.chatHistIdx > 0 {
						m.chatHistIdx--
						m.chatInput = m.chatInputHistory[m.chatHistIdx]
						m.chatCursor = len(m.chatInput)
						updateChatSuggests(&m)
					}
					return m, nil
				}
				if m.chatScroll < len(m.chatHistory)-5 {
					m.chatScroll++
				}
				return m, nil
			case "down":
				if len(m.chatSuggests) > 0 {
					m.chatSuggestIdx = (m.chatSuggestIdx + 1) % len(m.chatSuggests)
					return m, nil
				}
				if len(m.chatInputHistory) > 0 && m.chatHistIdx < len(m.chatInputHistory)-1 {
					m.chatHistIdx++
					m.chatInput = m.chatInputHistory[m.chatHistIdx]
					m.chatCursor = len(m.chatInput)
					updateChatSuggests(&m)
					return m, nil
				} else if m.chatHistIdx == len(m.chatInputHistory)-1 {
					m.chatHistIdx = len(m.chatInputHistory)
					m.chatInput = ""
					m.chatCursor = 0
					m.chatSuggests = nil
					return m, nil
				}
				if m.chatScroll > 0 {
					m.chatScroll--
				}
				return m, nil
			case "pgup":
				if m.chatScroll < len(m.chatHistory)-5 {
					m.chatScroll += 3
				}
				return m, nil
			case "pgdown":
				if m.chatScroll > 0 {
					m.chatScroll -= 3
					if m.chatScroll < 0 {
						m.chatScroll = 0
					}
				}
				return m, nil
			case "backspace":
				if m.chatCursor > 0 {
					m.chatInput = m.chatInput[:m.chatCursor-1] + m.chatInput[m.chatCursor:]
					m.chatCursor--
					updateChatSuggests(&m)
				}
				return m, nil
			case "left":
				if m.chatCursor > 0 {
					m.chatCursor--
					updateChatSuggests(&m)
				}
				return m, nil
			case "right":
				if m.chatCursor < len(m.chatInput) {
					m.chatCursor++
					updateChatSuggests(&m)
				}
				return m, nil
			case "space":
				m.chatInput = m.chatInput[:m.chatCursor] + " " + m.chatInput[m.chatCursor:]
				m.chatCursor++
				m.chatSuggests = nil
				return m, nil
			case "ctrl+l":
				m.chatHistory = []chatMsg{{From: "Hugin", Text: "Chat limpiado.", IsUser: false, Time: time.Now().Format("15:04")}}
				m.chatScroll = 0
				return m, nil
			default:
				s := msg.String()
				if len(s) == 1 && s >= "1" && s <= "6" {
					idx := int(s[0] - '1')
					if idx < len(chatAgents) {
						m.chatTarget = idx
					}
					return m, nil
				}
				if len(s) == 1 {
					m.chatInput = m.chatInput[:m.chatCursor] + s + m.chatInput[m.chatCursor:]
					m.chatCursor++
					updateChatSuggests(&m)
				}
				return m, nil
			}
		}
		// ---------- SETTINGS NAVIGATION ----------
		if m.mode == modeSettings {
			// helpers inline para lista aplanada
			flatLen := func() int {
				n := 0
				for i, s := range huginnSettings {
					n++ // header
					if m.settingsExpanded[i] {
						n += len(s.Items)
					}
				}
				return n
			}
			ensureVisible := func() {
				// viewport approx 18 filas visibles
				visible := 18
				if m.settingsCursor < m.settingsOffset {
					m.settingsOffset = m.settingsCursor
				} else if m.settingsCursor >= m.settingsOffset+visible {
					m.settingsOffset = m.settingsCursor - visible + 1
				}
			}
			getRow := func(idx int) (secIdx int, itemIdx int) {
				pos := 0
				for si, sec := range huginnSettings {
					if pos == idx {
						return si, -1
					}
					pos++
					if m.settingsExpanded[si] {
						if idx < pos+len(sec.Items) {
							return si, idx - pos
						}
						pos += len(sec.Items)
					}
				}
				return 0, -1
			}
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "esc", "q", "Q":
				m.mode = modeInput
				m.input = ""
				m.cursor = 0
				return m, nil
			case "up", "k":
				m.settingsCursor--
				if m.settingsCursor < 0 {
					m.settingsCursor = flatLen() - 1
				}
				ensureVisible()
				return m, nil
			case "down", "j":
				m.settingsCursor++
				if m.settingsCursor >= flatLen() {
					m.settingsCursor = 0
				}
				ensureVisible()
				return m, nil
			case "left", "h":
				si, it := getRow(m.settingsCursor)
				if it == -1 {
					// colapsa sección
					m.settingsExpanded[si] = false
				} else {
					// salta al header
					// busca inicio de sección
					pos := 0
					for i := 0; i < si; i++ {
						pos++
						if m.settingsExpanded[i] {
							pos += len(huginnSettings[i].Items)
						}
					}
					m.settingsCursor = pos
				}
				ensureVisible()
				return m, nil
			case "right", "l":
				si, it := getRow(m.settingsCursor)
				if it == -1 {
					m.settingsExpanded[si] = true
					ensureVisible()
				}
				return m, nil
			case "enter", " ":
				si, it := getRow(m.settingsCursor)
				if it == -1 {
					m.settingsExpanded[si] = !m.settingsExpanded[si]
					// si colapsa y cursor queda dentro, ajusta
					if !m.settingsExpanded[si] && m.settingsCursor > 0 {
						// mantiene cursor en header
					}
				} else {
					// toggle / edita valor simple: si es booleano, alterna
					key := huginnSettings[si].Items[it]
					val := settingsValues[key]
					switch val {
					case "On":
						settingsValues[key] = "Off"
					case "Off":
						settingsValues[key] = "On"
					case "Enabled":
						settingsValues[key] = "Disabled"
					case "Disabled":
						settingsValues[key] = "Enabled"
					case "Dark":
						settingsValues[key] = "Light"
					case "Light":
						settingsValues[key] = "Dark"
					default:
						m.msgTitle = key
						m.msgBody = fmt.Sprintf("%s\n\nValor actual: %s\n\n[Enter] para editar (próximamente)", key, val)
						m.mode = modeMessage
						return m, nil
					}
				}
				ensureVisible()
				return m, nil
			case "pgup":
				m.settingsCursor -= 5
				if m.settingsCursor < 0 {
					m.settingsCursor = 0
				}
				ensureVisible()
				return m, nil
			case "pgdown":
				m.settingsCursor += 5
				if m.settingsCursor >= flatLen() {
					m.settingsCursor = flatLen() - 1
				}
				ensureVisible()
				return m, nil
			case "ctrl+a":
				// expande todas
				for i := range m.settingsExpanded {
					m.settingsExpanded[i] = true
				}
				return m, nil
			case "ctrl+d":
				for i := range m.settingsExpanded {
					m.settingsExpanded[i] = false
				}
				m.settingsCursor = 0
				m.settingsOffset = 0
				return m, nil
			default:
				return m, nil
			}
		}
		// ---------- SERVERS HUB (MCP/LSP/Peers) ----------
		if m.mode == modeServers {
			listLen := func() int {
				switch m.serversTab {
				case 0:
					return len(mcpServers)
				case 1:
					return len(lspServers)
				default:
					return len(peerServers)
				}
			}
			ensureVisible := func() {
				visible := 12
				if m.serversCursor < m.serversOffset {
					m.serversOffset = m.serversCursor
				} else if m.serversCursor >= m.serversOffset+visible {
					m.serversOffset = m.serversCursor - visible + 1
				}
			}
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "esc", "q", "Q":
				m.mode = modeInput
				m.input = ""
				m.cursor = 0
				return m, nil
			case "tab", "right", "l":
				// tab cambia de pestaña MCP→LSP→Peers
				if msg.String() == "tab" {
					m.serversTab = (m.serversTab + 1) % 3
					m.serversCursor = 0
					m.serversOffset = 0
					m.serversMsg = ""
					return m, nil
				}
				// right/l en peers/mcp no hace nada extra, solo tab maneja
				return m, nil
			case "shift+tab", "left", "h":
				if msg.String() == "shift+tab" || msg.String() == "left" || msg.String() == "h" {
					if msg.String() == "shift+tab" {
						m.serversTab = (m.serversTab - 1 + 3) % 3
						m.serversCursor = 0
						m.serversOffset = 0
						m.serversMsg = ""
						return m, nil
					}
				}
				return m, nil
			case "1", "2", "3":
				tab := int(msg.String()[0] - '1')
				if tab >= 0 && tab < 3 {
					m.serversTab = tab
					m.serversCursor = 0
					m.serversOffset = 0
					m.serversMsg = ""
				}
				return m, nil
			case "up", "k":
				m.serversCursor--
				if m.serversCursor < 0 {
					m.serversCursor = listLen() - 1
				}
				ensureVisible()
				return m, nil
			case "down", "j":
				m.serversCursor++
				if m.serversCursor >= listLen() {
					m.serversCursor = 0
				}
				ensureVisible()
				return m, nil
			case "pgup":
				m.serversCursor -= 4
				if m.serversCursor < 0 {
					m.serversCursor = 0
				}
				ensureVisible()
				return m, nil
			case "pgdown":
				m.serversCursor += 4
				if m.serversCursor >= listLen() {
					m.serversCursor = listLen() - 1
				}
				ensureVisible()
				return m, nil
			case "enter", " ":
				// acción contextual: toggle enable / restart / pair
				needsTick := false
				switch m.serversTab {
				case 0: // MCP
					s := &mcpServers[m.serversCursor]
					if s.Status == "Needs auth" {
						m.serversMsg = fmt.Sprintf("→ Abriendo auth para %s … (configura %s en .env)", s.Name, s.Name)
						s.Status = "Connecting"
						needsTick = true
					} else if !s.Enabled {
						s.Enabled = true
						s.Status = "Connecting"
						m.serversMsg = fmt.Sprintf("→ %s habilitado, conectando…", s.Name)
						needsTick = true
					} else if s.Status == "Connected" {
						s.Status = "Connecting"
						m.serversMsg = fmt.Sprintf("→ Reiniciando %s …", s.Name)
						needsTick = true
					} else {
						s.Status = "Connected"
						s.Latency = "15ms"
						m.serversMsg = fmt.Sprintf("✓ %s reconectado (%s, %d tools)", s.Name, s.Latency, s.Tools)
					}
				case 1: // LSP
					s := &lspServers[m.serversCursor]
					if s.Status == "Running" {
						s.Status = "Stopped"
						m.serversMsg = fmt.Sprintf("■ %s (%s) detenido", s.ServerName, s.Language)
					} else if s.Status == "Stopped" {
						s.Status = "Running"
						m.serversMsg = fmt.Sprintf("▶ %s (%s) iniciado", s.ServerName, s.Language)
					} else if s.Status == "Error" {
						s.Status = "Running"
						s.Diagnostics = 0
						m.serversMsg = fmt.Sprintf("↻ %s reinstalado y corriendo", s.ServerName)
					}
					s.Enabled = s.Status == "Running"
				case 2: // Peers
					p := &peerServers[m.serversCursor]
					if !p.Paired {
						p.Status = peerPairing
						m.serversMsg = fmt.Sprintf("🔗 Código de emparejamiento para %s: %s — comparte este código en el otro PC", p.Hostname, strings.ToUpper(p.ID[len(p.ID)-4:])+"-HUGS")
						p.Paired = true
						p.Status = peerOnline
						p.VaultSync = "Synced"
					} else if p.Status == peerOnline {
						p.Status = peerOffline
						p.VaultSync = "Offline"
						m.serversMsg = fmt.Sprintf("⏸ %s desconectado", p.Hostname)
					} else {
						p.Status = peerOnline
						p.VaultSync = "Synced"
						p.Latency = "21ms"
						m.serversMsg = fmt.Sprintf("✓ %s reconectado (%s)", p.Hostname, p.IP)
					}
				}
				if needsTick {
					return m, tickServers()
				}
				return m, nil
			case "r", "R":
				// refresh / restart seleccionado
				switch m.serversTab {
				case 0:
					mcpServers[m.serversCursor].Status = "Connecting"
					m.serversMsg = fmt.Sprintf("↻ Reiniciando MCP %s …", mcpServers[m.serversCursor].Name)
				case 1:
					lspServers[m.serversCursor].Status = "Running"
					m.serversMsg = fmt.Sprintf("↻ Reiniciando LSP %s …", lspServers[m.serversCursor].ServerName)
				case 2:
					peerServers[m.serversCursor].Latency = fmt.Sprintf("%dms", 15+m.serversCursor*7)
					m.serversMsg = "↻ Ping a peers actualizado"
				}
				if m.serversTab == 0 {
					return m, tickServers()
				}
				return m, nil
			case "a", "A":
				if m.serversTab == 2 {
					// añadir peer
					newID := fmt.Sprintf("peer_%02dJNEW", len(peerServers)+1)
					peerServers = append(peerServers, peerServer{ID: newID, Hostname: fmt.Sprintf("pc-%d", len(peerServers)+1), IP: fmt.Sprintf("192.168.1.%d", 100+len(peerServers)), Status: peerPairing, VaultSync: "Pairing…", Latency: "—", LastSeen: "—", Paired: false})
					m.serversCursor = len(peerServers) - 1
					m.serversMsg = "＋ Nuevo peer añadido — presiona Enter para emparejar"
					ensureVisible()
				}
				return m, nil
			case "d", "D":
				if m.serversTab == 0 && len(mcpServers) > 1 {
					// toggle disable
					s := &mcpServers[m.serversCursor]
					s.Enabled = !s.Enabled
					if !s.Enabled {
						s.Status = "Disabled"
						m.serversMsg = fmt.Sprintf("⦸ %s deshabilitado", s.Name)
					} else {
						s.Status = "Connected"
						m.serversMsg = fmt.Sprintf("✓ %s habilitado", s.Name)
					}
				}
				return m, nil
			default:
				return m, nil
			}
		}
		// ---------- VAULT WIZARD ----------
		if m.mode == modeVault {
			q := vaultWizardQuestions[m.vaultWizardStep]
			hasOptions := len(q.Options) > 0
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "esc":
				if m.vaultWizardStep > 0 {
					m.vaultWizardStep--
					nq := vaultWizardQuestions[m.vaultWizardStep]
					// restaura valor previo
					switch nq.Key {
					case "Path":
						m.vaultWizardInput = m.vaultWizardData.Path
					case "Name":
						m.vaultWizardInput = m.vaultWizardData.Name
					case "Purpose":
						m.vaultWizardInput = m.vaultWizardData.Purpose
					case "VaultType":
						m.vaultWizardInput = m.vaultWizardData.VaultType
					case "Sync":
						m.vaultWizardInput = m.vaultWizardData.Sync
					case "Indexing":
						m.vaultWizardInput = m.vaultWizardData.Indexing
					case "Embeddings":
						m.vaultWizardInput = m.vaultWizardData.Embeddings
					default:
						m.vaultWizardInput = nq.Default
					}
					m.vaultWizardCursor = 0
				} else {
					m.mode = modeInput
					m.input = ""
					m.cursor = 0
				}
				return m, nil
			case "up", "k":
				if hasOptions {
					m.vaultWizardCursor = (m.vaultWizardCursor - 1 + len(q.Options)) % len(q.Options)
					m.vaultWizardInput = q.Options[m.vaultWizardCursor]
				}
				return m, nil
			case "down", "j":
				if hasOptions {
					m.vaultWizardCursor = (m.vaultWizardCursor + 1) % len(q.Options)
					m.vaultWizardInput = q.Options[m.vaultWizardCursor]
				}
				return m, nil
			case "enter":
				val := strings.TrimSpace(m.vaultWizardInput)
				if val == "" {
					val = q.Default
				}
				switch q.Key {
				case "Path":
					m.vaultWizardData.Path = val
				case "Name":
					m.vaultWizardData.Name = val
				case "Purpose":
					m.vaultWizardData.Purpose = val
				case "VaultType":
					m.vaultWizardData.VaultType = val
				case "Sync":
					m.vaultWizardData.Sync = val
				case "Indexing":
					m.vaultWizardData.Indexing = val
				case "Embeddings":
					m.vaultWizardData.Embeddings = val
				}
				if m.vaultWizardStep < len(vaultWizardQuestions)-1 {
					m.vaultWizardStep++
					nq := vaultWizardQuestions[m.vaultWizardStep]
					// precarga siguiente
					switch nq.Key {
					case "Path":
						m.vaultWizardInput = m.vaultWizardData.Path
					case "Name":
						m.vaultWizardInput = m.vaultWizardData.Name
					case "Purpose":
						m.vaultWizardInput = m.vaultWizardData.Purpose
					case "VaultType":
						m.vaultWizardInput = m.vaultWizardData.VaultType
					case "Sync":
						m.vaultWizardInput = m.vaultWizardData.Sync
					case "Indexing":
						m.vaultWizardInput = m.vaultWizardData.Indexing
					case "Embeddings":
						m.vaultWizardInput = m.vaultWizardData.Embeddings
					default:
						m.vaultWizardInput = nq.Default
					}
					// set cursor to current value index
					m.vaultWizardCursor = 0
					if len(nq.Options) > 0 {
						for i, opt := range nq.Options {
							if strings.EqualFold(opt, m.vaultWizardInput) {
								m.vaultWizardCursor = i
								break
							}
						}
					}
				} else {
					// wizard completo — usa VaultManager para materializar .huginn en disco
					vaultPath := m.vaultWizardData.Path
					if strings.HasPrefix(vaultPath, "~/") {
						home, _ := os.UserHomeDir()
						vaultPath = filepath.Join(home, strings.TrimPrefix(vaultPath, "~/"))
					}
					mgr := infravault.NewFilesystemManager()
					var v *vaultpkg.Vault
					var err error
					if _, err2 := os.Stat(vaultPath); os.IsNotExist(err2) {
						parent := filepath.Dir(vaultPath)
						name := filepath.Base(vaultPath)
						_ = os.MkdirAll(parent, 0755)
						v, err = mgr.Create(context.Background(), parent, name)
					} else {
						if mgr.IsInitialized(vaultPath) {
							v, err = mgr.Open(context.Background(), vaultPath)
						} else {
							v, err = mgr.Initialize(context.Background(), vaultPath)
						}
					}
					if err != nil {
						m.mode = modeMessage
						m.msgTitle = "Vault error"
						m.msgBody = fmt.Sprintf("No se pudo crear vault: %v\n\nPath: %s", err, vaultPath)
						m.vaultWizardStep = 0
						m.vaultWizardInput = ""
						return m, nil
					}
					m.vaultPath = v.Path
					m.vaultOK = true
					m.projectPath = v.Path
					m.projectName = v.Name
					settingsValues["Vault connection"] = "Connected"
					settingsValues["Default vault"] = v.Path
					settingsValues["Sync"] = m.vaultWizardData.Sync
					settingsValues["Indexing"] = m.vaultWizardData.Indexing
					settingsValues["Embeddings"] = m.vaultWizardData.Embeddings
					m.mode = modeMessage
					m.msgTitle = "Vault configurado"
					m.msgBody = fmt.Sprintf("Vault: %s\nName: %s\nID: %s\nPropósito: %s\nTipo: %s\nSync: %s\nIndexing: %s\nEmbeddings: %s\n\n✓ .huginn/config.json\n✓ .huginn/vault.json\n✓ .huginn/agents, memory, plugins, cache, logs, runtime\n✓ notes/projects/agents/memory/attachments\n\nVault listo. Presiona 'o' para abrir en explorer.", v.Path, v.Name, v.ID, m.vaultWizardData.Purpose, m.vaultWizardData.VaultType, m.vaultWizardData.Sync, m.vaultWizardData.Indexing, m.vaultWizardData.Embeddings)
					m.vaultWizardStep = 0
					m.vaultWizardInput = ""
				}
				return m, nil
			case "o", "O":
				_ = openVault()
				m.msgTitle = "Vault"
				m.msgBody = "Vault abierto en explorer: " + m.vaultWizardData.Path
				m.mode = modeMessage
				return m, nil
			case "backspace":
				if !hasOptions && len(m.vaultWizardInput) > 0 {
					m.vaultWizardInput = m.vaultWizardInput[:len(m.vaultWizardInput)-1]
				}
				return m, nil
			case "space":
				if !hasOptions {
					m.vaultWizardInput += " "
				}
				return m, nil
			default:
				s := msg.String()
				if len(s) == 1 && !hasOptions {
					m.vaultWizardInput += s
				}
				return m, nil
			}
		}
		// ---------- GRAPH INTERACTIVO ----------
		if m.mode == modeGraph {
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "esc", "q", "Q":
				m.mode = modeInput
				m.input = ""
				m.cursor = 0
				return m, nil
			case "up", "k":
				m.graphCursor = (m.graphCursor - 1 + len(graphNodes)) % len(graphNodes)
				return m, nil
			case "down", "j":
				m.graphCursor = (m.graphCursor + 1) % len(graphNodes)
				return m, nil
			case "left", "h":
				m.graphCursor = (m.graphCursor - 3 + len(graphNodes)) % len(graphNodes)
				return m, nil
			case "right", "l":
				m.graphCursor = (m.graphCursor + 3) % len(graphNodes)
				return m, nil
			case "enter", " ":
				n := graphNodes[m.graphCursor]
				m.mode = modeMessage
				m.msgTitle = n.Label
				m.msgBody = fmt.Sprintf("%s (%s)\n\n%s\n\nConexiones: %d\n\nPresiona esc para volver al grafo.", n.Label, n.Type, n.Desc, 2+m.graphCursor%4)
				return m, nil
			case "r", "R":
				m.graphCursor = 0
				return m, nil
			default:
				return m, nil
			}
		}
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.mode == modeInput && len(m.chatSuggests) > 0 {
				applyMainAutocomplete(&m)
				return m, nil
			}
			if m.mode == modeInput && strings.TrimSpace(m.input) != "" {
				raw := m.input
				trim := strings.TrimSpace(raw)
				low := strings.ToLower(trim)
				if strings.HasPrefix(trim, "/") || low == "q" || low == "quit" || low == "exit" || low == "help" || low == "agents" || low == "status" || low == "vault" || low == "graph" || low == "knowledge" {
					nm, c, quit := handleCommand(raw, m)
					if quit {
						return nm, c
					}
					return nm, c
				}
				// sin @ va directo al chat (no requiere mencionar agente)
				now := time.Now().Format("15:04")
				m.chatHistory = append(m.chatHistory, chatMsg{From: "You", Text: raw, IsUser: true, Time: now})
				m.chatInputHistory = append(m.chatInputHistory, raw)
				m.chatHistIdx = len(m.chatInputHistory)
				mentions := parseMentions(trim)
				var targets []string
				if len(mentions) > 0 {
					for _, mm := range mentions {
						if mm == "All" {
							targets = chatAgents[1:]
							break
						}
					}
					if len(targets) == 0 {
						targets = mentions
					}
				} else {
					// sin mencion → usa agente seleccionado en el chat (Tab/1-6) o All por defecto
					t := chatAgents[m.chatTarget]
					if t == "All" {
						targets = chatAgents[1:]
					} else {
						targets = []string{t}
					}
				}
			var cmds []tea.Cmd
			m.mascotState = "thinking"
			for _, ag := range targets {
					m.chatHistory = append(m.chatHistory, chatMsg{From: ag, Text: "… escribiendo (real) — llamando " + ag + "…", IsUser: false, Time: now})
				cmds = append(cmds, callAgentCmd(ag, raw, m.orchApp))
			}
			// lleva al chat y limpia el input — el chat ya muestra el input propio
				m.chatInput = ""
				m.chatCursor = 0
				m.mode = modeChat
				m.input = ""
				m.cursor = 0
				m.chatSuggests = nil
				m.chatScroll = 0
				if len(cmds) > 0 {
					return m, tea.Batch(cmds...)
				}
				return m, nil
			}
			if m.mode == modeHelp || m.mode == modeAgents || m.mode == modeStatus || m.mode == modeMessage {
				m.mode = modeInput
				m.input = ""
				m.cursor = 0
				return m, nil
			}
			return m, nil

		case "esc":
			if len(m.chatSuggests) > 0 {
				m.chatSuggests = nil
				return m, nil
			}
			if m.mode != modeInput {
				if m.mode == modeRunning {
					w, h := m.width, m.height
					m = initialModel()
					m.width, m.height = w, h
				} else {
					m.mode = modeInput
					m.input = ""
					m.cursor = 0
				}
			}
			return m, nil

		case "tab":
			if m.mode == modeInput && len(m.chatSuggests) > 0 {
				applyMainAutocomplete(&m)
				return m, nil
			}
			return m, nil

		case "up":
			if m.mode == modeInput && len(m.chatSuggests) > 0 {
				m.chatSuggestIdx = (m.chatSuggestIdx - 1 + len(m.chatSuggests)) % len(m.chatSuggests)
				return m, nil
			}
			return m, nil

		case "down":
			if m.mode == modeInput && len(m.chatSuggests) > 0 {
				m.chatSuggestIdx = (m.chatSuggestIdx + 1) % len(m.chatSuggests)
				return m, nil
			}
			return m, nil

		case "backspace":
			if m.mode == modeInput && m.cursor > 0 {
				m.input = m.input[:m.cursor-1] + m.input[m.cursor:]
				m.cursor--
				updateMainSuggests(&m)
			}
			return m, nil
		case "left":
			if m.mode == modeInput && m.cursor > 0 {
				m.cursor--
				updateMainSuggests(&m)
			}
			return m, nil
		case "right":
			if m.mode == modeInput && m.cursor < len(m.input) {
				m.cursor++
				updateMainSuggests(&m)
			}
			return m, nil
		case "space":
			if m.mode == modeInput {
				m.input = m.input[:m.cursor] + " " + m.input[m.cursor:]
				m.cursor++
				m.chatSuggests = nil
			}
			return m, nil
		default:
			if m.mode == modeInput && len(msg.String()) == 1 {
				m.input = m.input[:m.cursor] + msg.String() + m.input[m.cursor:]
				m.cursor++
				updateMainSuggests(&m)
			} else if m.mode != modeInput && (msg.String() == "q" || msg.String() == "Q") {
				if m.mode != modeInput {
					m.mode = modeInput
					m.input = ""
					m.cursor = 0
					return m, nil
				}
			}
			return m, nil
		}

	case agentReplyMsg:
		m.mascotState = "idle"
		replaced := false
		for i := len(m.chatHistory) - 1; i >= 0; i-- {
			if m.chatHistory[i].From == msg.Agent && strings.Contains(m.chatHistory[i].Text, "… escribiendo") {
				m.chatHistory[i] = chatMsg{From: msg.Agent, Text: msg.Text, IsUser: false, Time: time.Now().Format("15:04")}
				replaced = true
				break
			}
		}
		if !replaced {
			m.chatHistory = append(m.chatHistory, chatMsg{From: msg.Agent, Text: msg.Text, IsUser: false, Time: time.Now().Format("15:04")})
		}
		m.chatScroll = 0
		return m, nil

	case tickMsg:
		if m.mode == modeRunning {
			return advance(m)
		}
		if m.mode == modeServers {
			// simula reconexión MCP y progreso de sync en peers
			changed := false
			for i := range mcpServers {
				if mcpServers[i].Status == "Connecting" {
					mcpServers[i].Status = "Connected"
					mcpServers[i].Latency = fmt.Sprintf("%dms", 8+(i*5)%40)
					if mcpServers[i].Name == "notion" && mcpServers[i].ErrorMsg != "" {
						// sigue needing auth hasta que se configure, lo dejamos en Needs auth si tenía error
						// pero para demo, lo conectamos
						mcpServers[i].ErrorMsg = ""
					}
					changed = true
					m.serversMsg = fmt.Sprintf("✓ %s conectado (%s)", mcpServers[i].Name, mcpServers[i].Latency)
				}
			}
			for i := range peerServers {
				if peerServers[i].Status == peerSyncing {
					// simula avance 64% → Synced
					peerServers[i].VaultSync = "Synced"
					peerServers[i].Status = peerOnline
					peerServers[i].Latency = "19ms"
					peerServers[i].LastSeen = "now"
					changed = true
				}
			}
			if changed {
				return m, nil
			}
			return m, nil
		}
		return m, nil
	}
	return m, nil
}

func tick() tea.Cmd {
	return tea.Tick(180*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} })
}

func tickServers() tea.Cmd {
	return tea.Tick(900*time.Millisecond, func(_ time.Time) tea.Msg { return tickMsg{} })
}

func advance(m model) (tea.Model, tea.Cmd) {
	allDone := true
	for i := range m.agents {
		a := &m.agents[i]
		switch a.status {
		case statusWaiting:
			if i == 0 || m.agents[i-1].status == statusDone {
				a.status = statusWorking
				m.logs = append(m.logs, fmt.Sprintf("%-12s assigned: %s", a.name, a.role))
			}
			allDone = false
		case statusWorking:
			a.pct += 20
			if a.pct >= 100 {
				a.pct = 100
				a.status = statusDone
				m.logs = append(m.logs, fmt.Sprintf("%-12s done", a.name))
			}
			allDone = false
		case statusDone:
		}
	}
	if allDone {
		m.logs = append(m.logs, "huginn      all agents completed — ready for review")
		return m, nil
	}
	return m, tick()
}

// ---------- view ----------

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	header := tuicomp.RenderHeader(tuicomp.HeaderProps{
		ProjectPath: m.projectPath,
		ProjectName: m.projectName,
		VaultPath:   m.vaultPath,
		VaultOK:     m.vaultOK,
		PkgManager:  m.pkgManager,
		Version:     VERSION,
		Width:       76,
	})

	// chat tiene layout propio de 86 cols (dashboard) — no usa header global
	if m.mode == modeChat {
		v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, viewChat(m)))
		v.AltScreen = true
		return v
	}
	var body string
	switch m.mode {
	case modeInput:
		body = viewInput(m)
	case modeRunning:
		body = viewRunning(m)
	case modeHelp:
		body = viewHelp(m)
	case modeAgents:
		body = viewAgents(m)
	case modeStatus:
		body = viewStatus(m)
	case modeMessage:
		body = viewMessage(m)
	case modeSettings:
		body = viewSettings(m)
	case modeServers:
		body = viewServers(m)
	case modeGraph:
		body = viewGraph(m)
	case modeVault:
		body = viewVaultWizard(m)
	}

	content := lipgloss.JoinVertical(lipgloss.Center, header, "", body)
	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}

func viewInput(m model) string {
	line := m.input
	if line == "" {
		line = lipgloss.NewStyle().Foreground(colMuted).Render("escribe para chatear — Enter para enviar  •  @chatgpt @opencode @all opcional")
	} else {
		cursorChar := " "
		if m.cursor < len(line) {
			cursorChar = string(line[m.cursor])
		}
		before := line[:m.cursor]
		after := ""
		if m.cursor < len(line) {
			after = line[m.cursor+1:]
		}
		caret := lipgloss.NewStyle().Background(colWhite).Foreground(lipgloss.Color("#130E0A")).Render(cursorChar)
		if strings.Contains(before+after, "@") {
			before = highlightMentions(before)
			after = highlightMentions(after)
		}
		line = before + caret + after
	}

	cmdHint := lipgloss.NewStyle().Foreground(colMuted).Render("Try ") +
		lipgloss.NewStyle().Bold(true).Foreground(colAccent).Render("/help") + lipgloss.NewStyle().Foreground(colMuted).Render("  ") +
		lipgloss.NewStyle().Foreground(colText).Render("/agents") + lipgloss.NewStyle().Foreground(colMuted).Render("  ") +
		lipgloss.NewStyle().Foreground(colText).Render("@all") + lipgloss.NewStyle().Foreground(colMuted).Render(" chat  ") +
		lipgloss.NewStyle().Foreground(colText).Render("/vault")

	panel := lipgloss.NewStyle().
		Background(colPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(colAccent).
		Padding(1, 3).
		Width(76).
		Render(line + "\n\n" + cmdHint)

	mentionSuggest := ""
	if atIdx := strings.LastIndex(m.input[:m.cursor], "@"); atIdx != -1 {
		partial := m.input[atIdx:m.cursor]
		if !strings.Contains(partial, " ") && len(partial) <= 12 {
			sugs := renderMentionSuggestions(partial)
			if len(sugs) > 0 {
				if len(sugs) > 5 {
					sugs = sugs[:5]
				}
				var styled []string
				for i, s := range sugs {
					if i == m.chatSuggestIdx && len(m.chatSuggests) > 0 {
						styled = append(styled, lipgloss.NewStyle().Foreground(colWhite).Background(colAccent).Padding(0, 1).Bold(true).Render(s+" ↵"))
					} else {
						styled = append(styled, lipgloss.NewStyle().Foreground(colAccent).Bold(true).Render(s))
					}
				}
				mentionSuggest = lipgloss.NewStyle().Foreground(colMuted).Render(strings.Join(styled, "  ")) + lipgloss.NewStyle().Foreground(colMuted).Render("  Tab/Enter autocompleta")
			}
		}
	}

	hints := lipgloss.NewStyle().Foreground(colMuted).Render(
		lipgloss.NewStyle().Bold(true).Foreground(colText).Render("enter") + " run / @chat    " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("ctrl+c") + " quit")

	if mentionSuggest != "" {
		return lipgloss.JoinVertical(lipgloss.Right, panel, mentionSuggest, hints)
	}
	return lipgloss.JoinVertical(lipgloss.Right, panel, hints)
}

func viewRunning(m model) string {
	var rows []string
	for _, a := range m.agents {
		dot := lipgloss.NewStyle().Foreground(a.status.color()).Render("●")
		name := lipgloss.NewStyle().Bold(true).Foreground(colText).Width(12).Render(a.name)
		role := lipgloss.NewStyle().Foreground(colMuted).Width(30).Render(a.role)
		status := lipgloss.NewStyle().Foreground(a.status.color()).Render(a.status.label())
		rows = append(rows, fmt.Sprintf("%s %s %s %s", dot, name, role, status))
	}
	agentsBlock := lipgloss.NewStyle().
		Background(colPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(colAccent).
		Padding(1, 3).
		Width(76).
		Render(strings.Join(rows, "\n"))

	logLines := m.logs
	if len(logLines) > 8 {
		logLines = logLines[len(logLines)-8:]
	}
	logStyle := lipgloss.NewStyle().Foreground(colMuted)
	logsBlock := lipgloss.NewStyle().
		Background(lipgloss.Color("#130E0A")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder).
		Padding(1, 2).
		Width(76).
		Render(logStyle.Render(strings.Join(logLines, "\n")))

	hints := lipgloss.NewStyle().Foreground(colMuted).Render(
		lipgloss.NewStyle().Bold(true).Foreground(colText).Render("esc") + " new task    " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("ctrl+c") + " quit")

	return lipgloss.JoinVertical(lipgloss.Right, agentsBlock, "", logsBlock, hints)
}

func viewHelp(m model) string {
	var rows []string
	// mascota cabecera — Sabio con libro (sprite PNG)
	mascot := tuicomp.RenderMascotSprite(tuicomp.StateWise, "")
	mascotTitle := lipgloss.NewStyle().Foreground(colMuted).Italic(true).Render("Huginn vuela, recuerda, conecta mundos.  — La mirada que todo lo ve")
	rows = append(rows, mascot, mascotTitle, "")
	title := lipgloss.NewStyle().Bold(true).Foreground(colAccent).Render("HUGINN commands")
	rows = append(rows, title, "")
	for _, c := range commands {
		left := lipgloss.NewStyle().Bold(true).Foreground(colPurple).Width(15).Render(c.Name)
		mid := lipgloss.NewStyle().Foreground(colText).Width(38).Render(c.Description)
		right := lipgloss.NewStyle().Foreground(colMuted).Render(c.Shortcut)
		rows = append(rows, fmt.Sprintf("%s %s %s", left, mid, right))
	}
	body := strings.Join(rows, "\n")
	panel := lipgloss.NewStyle().
		Background(colPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder2).
		Padding(1, 3).
		Width(76).
		Render(body)
	hints := lipgloss.NewStyle().Foreground(colMuted).Render(lipgloss.NewStyle().Bold(true).Foreground(colText).Render("esc") + " back    " + lipgloss.NewStyle().Bold(true).Foreground(colText).Render("enter") + " back")
	return lipgloss.JoinVertical(lipgloss.Right, panel, hints)
}

func viewAgents(m model) string {
	var rows []string
	title := lipgloss.NewStyle().Bold(true).Foreground(colAccent).Render("AGENTS  —  Connected")
	rows = append(rows, title, "")
	for _, a := range backendAgents {
		online := commandAvailable(a.Command)
		state := "online"
		stateCol := colSuccess
		if !online {
			state = "not installed"
			stateCol = colMuted
		}
		name := lipgloss.NewStyle().Bold(true).Foreground(colText).Width(12).Render(a.Name)
		stateStr := lipgloss.NewStyle().Foreground(stateCol).Width(16).Render(state)
		desc := lipgloss.NewStyle().Foreground(colMuted).Render(a.Description)
		rows = append(rows, fmt.Sprintf("%s %s %s", name, stateStr, desc))
	}
	rows = append(rows, "", lipgloss.NewStyle().Foreground(colMuted).Render("Vault: shared knowledge source of truth"))
	body := strings.Join(rows, "\n")
	panel := lipgloss.NewStyle().
		Background(colPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder2).
		Padding(1, 3).
		Width(76).
		Render(body)
	hints := lipgloss.NewStyle().Foreground(colMuted).Render(lipgloss.NewStyle().Bold(true).Foreground(colText).Render("esc") + " back")
	return lipgloss.JoinVertical(lipgloss.Right, panel, hints)
}

func viewStatus(m model) string {
	var rows []string
	mascot := tuicomp.RenderMascotSprite(tuicomp.StateObserving, "")
	rows = append(rows, mascot, "")
	title := lipgloss.NewStyle().Bold(true).Foreground(colAccent).Render("HUGINN STATUS")
	rows = append(rows, title, "")
	rows = append(rows, fmt.Sprintf("%-14s %s", "version", VERSION))
	rows = append(rows, fmt.Sprintf("%-14s %s", "knowledge", "ready"))
	rows = append(rows, fmt.Sprintf("%-14s %s", "research", "ready"))
	rows = append(rows, fmt.Sprintf("%-14s %s", "agent layer", "ready"))
	rows = append(rows, "")
	for _, a := range backendAgents {
		online := commandAvailable(a.Command)
		state := "online"
		stateCol := colSuccess
		if !online {
			state = "not installed"
			stateCol = colMuted
		}
		name := lipgloss.NewStyle().Bold(true).Foreground(colText).Width(12).Render(a.Name)
		stateStr := lipgloss.NewStyle().Foreground(stateCol).Width(16).Render(state)
		rows = append(rows, fmt.Sprintf("%s %s", name, stateStr))
	}
	body := strings.Join(rows, "\n")
	panel := lipgloss.NewStyle().
		Background(colPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder2).
		Padding(1, 3).
		Width(76).
		Render(body)
	hints := lipgloss.NewStyle().Foreground(colMuted).Render(lipgloss.NewStyle().Bold(true).Foreground(colText).Render("esc") + " back")
	return lipgloss.JoinVertical(lipgloss.Right, panel, hints)
}

func viewMessage(m model) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(colAccent).Render(m.msgTitle)
	body := lipgloss.NewStyle().Foreground(colText).Render(m.msgBody)
	content := title + "\n\n" + body
	panel := lipgloss.NewStyle().
		Background(colPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder2).
		Padding(1, 3).
		Width(76).
		Render(content)
	hints := lipgloss.NewStyle().Foreground(colMuted).Render(lipgloss.NewStyle().Bold(true).Foreground(colText).Render("esc") + " back    " + lipgloss.NewStyle().Bold(true).Foreground(colText).Render("enter") + " back")
	return lipgloss.JoinVertical(lipgloss.Right, panel, hints)
}

func viewSettings(m model) string {
	// header
	header := lipgloss.NewStyle().Bold(true).Foreground(colAccent).Render("⚙ Settings")
	sub := lipgloss.NewStyle().Foreground(colMuted).Render(fmt.Sprintf("%d secciones  •  %d ajustes", len(huginnSettings), 36))
	top := lipgloss.JoinVertical(lipgloss.Left, header, sub, "")

	// construir lista aplanada para render + scroll
	type flatRow struct {
		secIdx  int
		itemIdx int // -1 = header
	}
	var flat []flatRow
	for si := range huginnSettings {
		flat = append(flat, flatRow{si, -1})
		if m.settingsExpanded[si] {
			for ii := range huginnSettings[si].Items {
				flat = append(flat, flatRow{si, ii})
			}
		}
	}
	visible := 18
	offset := m.settingsOffset
	if offset < 0 {
		offset = 0
	}
	if offset > len(flat)-visible {
		offset = len(flat) - visible
	}
	if offset < 0 {
		offset = 0
	}
	end := offset + visible
	if end > len(flat) {
		end = len(flat)
	}
	var rows []string
	for idx := offset; idx < end; idx++ {
		r := flat[idx]
		sec := huginnSettings[r.secIdx]
		isLastSec := r.secIdx == len(huginnSettings)-1
		selected := idx == m.settingsCursor

		var line string
		if r.itemIdx == -1 {
			// header fila
			prefix := "├── "
			if isLastSec {
				prefix = "└── "
			}
			arrow := "▶"
			if m.settingsExpanded[r.secIdx] {
				arrow = "▼"
			}
			label := fmt.Sprintf("%s %s %s", arrow, sec.Name, lipgloss.NewStyle().Foreground(colMuted).Render(fmt.Sprintf("(%d)", len(sec.Items))))
			// si está expandido, el header muestra flecha ▼
			text := prefix + label
			style := lipgloss.NewStyle().Foreground(colWhite).Bold(true)
			if selected {
				style = lipgloss.NewStyle().Foreground(colWhite).Background(colBorder2).Bold(true).Padding(0, 1)
				text = style.Render(text)
			} else {
				text = style.Render(text)
			}
			// marca selección con ▶ a la izquierda
			if selected {
				line = lipgloss.NewStyle().Foreground(colAccent).Render("▶ ") + text
			} else {
				line = "  " + text
			}
		} else {
			// item fila
			base := "│   "
			if isLastSec {
				base = "    "
			}
			isLastItem := r.itemIdx == len(sec.Items)-1
			branch := "├── "
			if isLastItem {
				branch = "└── "
			}
			itemName := sec.Items[r.itemIdx]
			val := settingsValues[itemName]
			// truncar valor si es largo
			if len(val) > 18 {
				val = val[:18] + "…"
			}
			left := fmt.Sprintf("%s%s%s", base, branch, itemName)
			// columnas: nombre a la izq, valor a la dcha
			valStyled := lipgloss.NewStyle().Foreground(colMuted2).Render(val)
			// booleans con color
			if val == "On" || val == "Enabled" || val == "Connected" || val == "Auto" {
				valStyled = lipgloss.NewStyle().Foreground(colSuccess).Render(val)
			} else if val == "Off" || val == "Disabled" {
				valStyled = lipgloss.NewStyle().Foreground(colMuted).Render(val)
			} else if val == "Dark" {
				valStyled = lipgloss.NewStyle().Foreground(colPurple).Render(val)
			}
			// ancho disponible 76 - padding 6 => 70 para left+val
			gap := 44 - len(itemName)
			if gap < 2 {
				gap = 2
			}
			lineContent := left + strings.Repeat(" ", gap) + valStyled
			if selected {
				// highlight fila completa
				style := lipgloss.NewStyle().Foreground(colAccent).Background(colPanel2).Padding(0, 1)
				lineContent = style.Render(strings.TrimSpace(lineContent))
				if val != "" {
					// reañadir valor con highlight? mantenemos simple
					lineContent = style.Render(fmt.Sprintf("%s%s  %s", base+branch, itemName, val))
				}
				line = lipgloss.NewStyle().Foreground(colAccent).Render("▶ ") + lineContent
			} else {
				line = "  " + lipgloss.NewStyle().Foreground(colText2).Render(left) + strings.Repeat(" ", gap) + valStyled
			}
		}
		rows = append(rows, line)
	}
	if len(rows) == 0 {
		rows = []string{lipgloss.NewStyle().Foreground(colMuted).Render("No settings")}
	}
	// indicador scroll
	scrollInfo := ""
	if len(flat) > visible {
		scrollInfo = lipgloss.NewStyle().Foreground(colMuted).Render(fmt.Sprintf(" %d/%d ", m.settingsCursor+1, len(flat)))
	}
	// panel principal
	body := strings.Join(rows, "\n")
	panel := lipgloss.NewStyle().
		Background(colPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder2).
		Padding(1, 2).
		Width(76).
		Height(22).
		Render(top + "\n" + body + "\n" + scrollInfo)

	hints := lipgloss.NewStyle().Foreground(colMuted).Render(
		lipgloss.NewStyle().Bold(true).Foreground(colText).Render("↑↓") + " navegar  " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("←→") + " colapsar/expandir  " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("enter") + " toggle  " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("esc") + " volver",
	)
	hints2 := lipgloss.NewStyle().Foreground(colMuted).Render(
		lipgloss.NewStyle().Bold(true).Foreground(colText).Render("ctrl+a") + " expandir todo  " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("ctrl+d") + " colapsar todo",
	)
	return lipgloss.JoinVertical(lipgloss.Right, panel, hints, hints2)
}

func viewServers(m model) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(colAccent).Render("SERVERS  —  MCP · LSP · Peers")
	sub := lipgloss.NewStyle().Foreground(colMuted).Render(fmt.Sprintf("MCP · LSP · Peers  •  local %s (%s)", localHostname, localPeerID))
	header := lipgloss.JoinVertical(lipgloss.Left, title, sub, "")

	// tabs
	tabNames := []string{fmt.Sprintf("MCP (%d)", len(mcpServers)), fmt.Sprintf("LSP (%d)", len(lspServers)), fmt.Sprintf("Peers (%d)", len(peerServers))}
	var tabs []string
	for i, n := range tabNames {
		active := i == m.serversTab
		icon := "◇"
		if i == 0 {
			icon = "⬢"
		} else if i == 1 {
			icon = "λ"
		} else {
			icon = "◈"
		}
		label := fmt.Sprintf("%s %d:%s", icon, i+1, n)
		var st lipgloss.Style
		if active {
			st = lipgloss.NewStyle().Foreground(colPanel).Background(colAccent).Padding(0, 1).Bold(true)
		} else {
			st = lipgloss.NewStyle().Foreground(colMuted2).Background(colPanel2).Padding(0, 1)
		}
		tabs = append(tabs, st.Render(label))
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	tabsBar := lipgloss.PlaceHorizontal(76, lipgloss.Center, tabsRow)
	if lipgloss.Width(tabsRow) > 76 {
		tabsBar = lipgloss.PlaceHorizontal(76, lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
	}
	// list box + detail box side by side
	visible := 9
	var listRows []string
	var detailRows []string
	switch m.serversTab {
	case 0: // MCP
		start := m.serversOffset
		end := start + visible
		if end > len(mcpServers) {
			end = len(mcpServers)
		}
		if start < 0 {
			start = 0
		}
		for i := start; i < end; i++ {
			s := mcpServers[i]
			sel := i == m.serversCursor
			dot := "●"
			col := colSuccess
			switch s.Status {
			case "Connected":
				col = colSuccess
			case "Needs auth":
				col = lipgloss.Color("#E1A451")
			case "Connecting":
				col = colAccent
			case "Disabled":
				col = colMuted
				dot = "○"
			case "Error":
				col = colError
			}
			dotS := lipgloss.NewStyle().Foreground(col).Render(dot)
			name := s.Name
			if !s.Enabled {
				name = lipgloss.NewStyle().Foreground(colMuted).Render(name + " (off)")
			} else {
				name = lipgloss.NewStyle().Bold(true).Foreground(colText).Render(name)
			}
			transport := lipgloss.NewStyle().Foreground(colMuted).Render(string(s.Transport))
			tools := lipgloss.NewStyle().Foreground(colMuted2).Render(fmt.Sprintf("%d tools", s.Tools))
			lat := lipgloss.NewStyle().Foreground(colMuted).Render(s.Latency)
			status := lipgloss.NewStyle().Foreground(col).Render(s.Status)
			line := fmt.Sprintf("%s %-11s %-7s %-8s %s", dotS, name, transport, tools, lat+" "+status)
			if sel {
				line = lipgloss.NewStyle().Foreground(colWhite).Background(colBorder2).Padding(0, 1).Render(line)
				listRows = append(listRows, lipgloss.NewStyle().Foreground(colAccent).Render("▶ ")+line)
			} else {
				listRows = append(listRows, "  "+line)
			}
		}
		// detail
		if len(mcpServers) > 0 {
			s := mcpServers[m.serversCursor]
			detailRows = []string{
				lipgloss.NewStyle().Bold(true).Foreground(colWhite).Render(s.Name),
				lipgloss.NewStyle().Foreground(colMuted).Render(s.Command),
				"",
				fmt.Sprintf("Transport: %s  •  Tools: %d  •  %s", s.Transport, s.Tools, s.Status),
				fmt.Sprintf("Latency: %s", s.Latency),
			}
			if s.ErrorMsg != "" {
				detailRows = append(detailRows, lipgloss.NewStyle().Foreground(colError).Render("Error: "+s.ErrorMsg))
			}
			detailRows = append(detailRows, "", lipgloss.NewStyle().Foreground(colMuted).Italic(true).Render("Enter: restart/auth  d: toggle  r: refresh"))
		}
	case 1: // LSP
		start := m.serversOffset
		end := start + visible
		if end > len(lspServers) {
			end = len(lspServers)
		}
		for i := start; i < end; i++ {
			s := lspServers[i]
			sel := i == m.serversCursor
			dot := "●"
			col := colSuccess
			switch s.Status {
			case "Running":
				col = colSuccess
			case "Stopped":
				col = colMuted
				dot = "○"
			case "Error":
				col = colError
			case "Installing":
				col = colAccent
				dot = "◌"
			}
			dotS := lipgloss.NewStyle().Foreground(col).Render(dot)
			lang := lipgloss.NewStyle().Bold(true).Foreground(colText).Width(11).Render(s.Language)
			srv := lipgloss.NewStyle().Foreground(colText2).Width(14).Render(s.ServerName)
			ver := lipgloss.NewStyle().Foreground(colMuted).Render(s.Version)
			diag := ""
			if s.Diagnostics > 0 {
				diag = lipgloss.NewStyle().Foreground(colWarn).Render(fmt.Sprintf("%d diag", s.Diagnostics))
			} else {
				diag = lipgloss.NewStyle().Foreground(colMuted).Render("0 diag")
			}
			status := lipgloss.NewStyle().Foreground(col).Render(s.Status)
			line := fmt.Sprintf("%s %s %s %s %s", dotS, lang, srv, diag, status+" "+ver)
			if sel {
				line = lipgloss.NewStyle().Foreground(colWhite).Background(colBorder2).Padding(0, 1).Render(line)
				listRows = append(listRows, lipgloss.NewStyle().Foreground(colAccent).Render("▶ ")+line)
			} else {
				listRows = append(listRows, "  "+line)
			}
		}
		if len(lspServers) > 0 {
			s := lspServers[m.serversCursor]
			detailRows = []string{
				lipgloss.NewStyle().Bold(true).Foreground(colWhite).Render(s.Language + " — " + s.ServerName),
				lipgloss.NewStyle().Foreground(colMuted).Render(s.Command),
				fmt.Sprintf("Root: %s  •  %s  •  %s", s.Root, s.Version, s.Status),
				fmt.Sprintf("Diagnostics: %d", s.Diagnostics),
				"", lipgloss.NewStyle().Foreground(colMuted).Italic(true).Render("Enter: start/stop  r: restart"),
			}
		}
	case 2: // Peers
		start := m.serversOffset
		end := start + visible
		if end > len(peerServers) {
			end = len(peerServers)
		}
		for i := start; i < end; i++ {
			p := peerServers[i]
			sel := i == m.serversCursor
			dot := "●"
			col := colSuccess
			switch p.Status {
			case peerOnline:
				col = colSuccess
			case peerOffline:
				col = colMuted
				dot = "○"
			case peerPairing:
				col = colWarn
				dot = "◌"
			case peerSyncing:
				col = colAccent
			}
			dotS := lipgloss.NewStyle().Foreground(col).Render(dot)
			host := lipgloss.NewStyle().Bold(true).Foreground(colText).Width(13).Render(p.Hostname)
			ip := lipgloss.NewStyle().Foreground(colMuted).Width(13).Render(p.IP)
			sync := lipgloss.NewStyle().Foreground(colMuted2).Render(p.VaultSync)
			if p.VaultSync == "Synced" {
				sync = lipgloss.NewStyle().Foreground(colSuccess).Render(p.VaultSync)
			} else if p.VaultSync == "Syncing 64%" || p.Status == peerSyncing {
				sync = lipgloss.NewStyle().Foreground(colAccent).Render(p.VaultSync)
			}
			lat := lipgloss.NewStyle().Foreground(colMuted).Render(p.Latency)
			stat := lipgloss.NewStyle().Foreground(col).Render(string(p.Status))
			line := fmt.Sprintf("%s %s %s %s %s", dotS, host, ip, sync, lat+" "+stat)
			if sel {
				line = lipgloss.NewStyle().Foreground(colWhite).Background(colBorder2).Padding(0, 1).Render(line)
				listRows = append(listRows, lipgloss.NewStyle().Foreground(colAccent).Render("▶ ")+line)
			} else {
				listRows = append(listRows, "  "+line)
			}
		}
		if len(peerServers) > 0 {
			p := peerServers[m.serversCursor]
			trust := "Untrusted"
			if p.Trusted {
				trust = lipgloss.NewStyle().Foreground(colSuccess).Render("Trusted")
			}
			detailRows = []string{
				lipgloss.NewStyle().Bold(true).Foreground(colWhite).Render(p.Hostname + "  " + p.IP),
				fmt.Sprintf("ID: %s  •  %s  •  %s", p.ID, p.Status, trust),
				fmt.Sprintf("Vault: %s  •  Latency: %s  •  Seen: %s", p.VaultSync, p.Latency, p.LastSeen),
				"", lipgloss.NewStyle().Foreground(colMuted).Italic(true).Render("Enter: pair/connect  a: add peer  r: ping"),
			}
			if !p.Paired {
				detailRows = append(detailRows, lipgloss.NewStyle().Foreground(colWarn).Render("⚠ No emparejado — Enter para generar código"))
			}
		}
	}
	if len(listRows) == 0 {
		listRows = []string{lipgloss.NewStyle().Foreground(colMuted).Render("No servers")}
	}
	listBox := lipgloss.NewStyle().
		Background(colPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder2).
		Padding(1, 1).
		Width(46).
		Height(13).
		Render(strings.Join(listRows, "\n"))
	detailBox := lipgloss.NewStyle().
		Background(colPanel2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder).
		Padding(1, 1).
		Width(28).
		Height(13).
		Render(strings.Join(detailRows, "\n"))
	middle := lipgloss.JoinHorizontal(lipgloss.Top, listBox, " ", detailBox)
	msgLine := ""
	if m.serversMsg != "" {
		msgLine = lipgloss.NewStyle().Foreground(colWarn).Italic(true).Width(76).Render(m.serversMsg)
	} else {
		msgLine = lipgloss.NewStyle().Foreground(colMuted).Width(76).Align(lipgloss.Center).Render("tip: 1/2/3 cambia pestaña  •  tab rota  •  conexión real via WebSocket/mDNS (simulado)")
	}
	// vault sync footer
	localInfo := lipgloss.NewStyle().Foreground(colMuted).Render(fmt.Sprintf("local: %s  •  %d MCP  %d LSP  %d peers  •  vault ~/huginn-vault", localHostname, len(mcpServers), len(lspServers), len(peerServers)))
	hints := lipgloss.NewStyle().Foreground(colMuted).Render(
		lipgloss.NewStyle().Bold(true).Foreground(colText).Render("↑↓") + " navegar  " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("enter") + " acción  " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("tab") + " pestaña  " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("r") + " refresh  " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("esc") + " volver",
	)
	content := lipgloss.JoinVertical(lipgloss.Left, header, tabsBar, "", middle, msgLine, localInfo)
	panel := lipgloss.NewStyle().
		Background(colPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder).
		Padding(0, 1).
		Width(76).
		Render(content)
	return lipgloss.JoinVertical(lipgloss.Right, panel, hints)
}

func viewGraph(m model) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(colAccent).Render("KNOWLEDGE GRAPH")
	sub := lipgloss.NewStyle().Foreground(colMuted).Render(fmt.Sprintf("12 nodes  •  23 edges  •  4 agents  •  synced  •  %d/%d", m.graphCursor+1, len(graphNodes)))
	header := lipgloss.JoinVertical(lipgloss.Left, title, sub, "")
	// grafo base
	baseGraph := []string{
		lipgloss.NewStyle().Foreground(colMuted).Render("                         ") + lipgloss.NewStyle().Foreground(colPurple).Bold(true).Render("● Agent"),
		lipgloss.NewStyle().Foreground(colBorder).Render("                            │"),
		lipgloss.NewStyle().Foreground(colBorder).Render("                ┌────────────┼────────────┐"),
		lipgloss.NewStyle().Foreground(colAccent).Render("              ◉ Memory") + lipgloss.NewStyle().Foreground(colMuted).Render("     ") + lipgloss.NewStyle().Foreground(colWarn).Render("● Context") + lipgloss.NewStyle().Foreground(colMuted).Render("     ") + lipgloss.NewStyle().Foreground(colSuccess).Render("○ Project"),
		lipgloss.NewStyle().Foreground(colBorder).Render("                │            │            │"),
		lipgloss.NewStyle().Foreground(colBorder).Render("         ┌──────┴──────┐     │     ┌──────┴──────┐"),
		lipgloss.NewStyle().Foreground(colMuted2).Render("     ● embeddings  ○ session") + lipgloss.NewStyle().Foreground(colBorder).Render("     │     ") + lipgloss.NewStyle().Foreground(colWarn).Render("● Task") + lipgloss.NewStyle().Foreground(colBorder).Render("   ") + lipgloss.NewStyle().Foreground(lipgloss.Color("#CD8D38")).Render("◉ Knowledge"),
		lipgloss.NewStyle().Foreground(colBorder).Render("         │           │     │     │           │"),
		lipgloss.NewStyle().Foreground(colBorder).Render("    ┌────┴────┐ ┌────┴────┐│┌────┴────┐ ┌────┴────┐"),
		lipgloss.NewStyle().Foreground(colMuted2).Render("  ○ recent  ● graph") + lipgloss.NewStyle().Foreground(colPurple).Render("  ◉ Huginn") + lipgloss.NewStyle().Foreground(colSuccess).Render("   ● vault  ○ index.ts"),
		lipgloss.NewStyle().Foreground(colBorder).Render("    │       │  │       │ │  │       │  │       │"),
		lipgloss.NewStyle().Foreground(colMuted2).Render("  ○ config ● context") + lipgloss.NewStyle().Foreground(colWarn).Render(" ● reviewer") + lipgloss.NewStyle().Foreground(colPurple).Render(" ● architect") + lipgloss.NewStyle().Foreground(colSuccess).Render(" ○ dev") + lipgloss.NewStyle().Foreground(lipgloss.Color("#CD8D38")).Render(" ● researcher"),
	}
	// resalta línea que contiene el nodo seleccionado
	sel := graphNodes[m.graphCursor]
	highlighted := make([]string, len(baseGraph))
	for i, line := range baseGraph {
		if strings.Contains(line, sel.Label) {
			highlighted[i] = lipgloss.NewStyle().Background(colPanel2).Foreground(colWhite).Bold(true).Padding(0, 1).Render(line)
		} else {
			highlighted[i] = line
		}
	}
	content := strings.Join(highlighted, "\n")
	graphBox := lipgloss.NewStyle().
		Background(colPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder2).
		Padding(1, 1).
		Width(76).
		Height(16).
		Render(content)
	// panel de detalle del nodo seleccionado
	selDetail := lipgloss.NewStyle().Bold(true).Foreground(sel.Color).Render("▶ "+sel.Label) + lipgloss.NewStyle().Foreground(colMuted).Render("  •  "+sel.Type) + "\n" + lipgloss.NewStyle().Foreground(colText2).Render(sel.Desc)
	detailBox := lipgloss.NewStyle().Background(colPanel2).BorderStyle(lipgloss.NormalBorder()).BorderForeground(colBorder).Padding(0, 1).Width(76).Render(selDetail)
	stats := lipgloss.NewStyle().Foreground(colMuted).Render("◉ 12 nodes  •  23 edges  •  4 agents  •  synced") +
		lipgloss.NewStyle().Foreground(colMuted2).Render("    │    ") +
		lipgloss.NewStyle().Foreground(colAccent).Render("vault: ~/agent-vault")
	statsLine := lipgloss.NewStyle().Width(76).Align(lipgloss.Center).Render(stats)
	legend := lipgloss.NewStyle().Foreground(colMuted).Render("● agent  ◉ memory  ○ project  ─ connection") + lipgloss.NewStyle().Foreground(colMuted2).Render("  •  ↑↓←→ navega")
	hints := lipgloss.NewStyle().Foreground(colMuted).Render(
		lipgloss.NewStyle().Bold(true).Foreground(colText).Render("esc") + " volver  " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("↑↓←→") + " navegar  " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("enter") + " inspeccionar  " +
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("r") + " reset",
	)
	panel := lipgloss.NewStyle().
		Background(colPanel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder).
		Padding(0, 1).
		Width(76).
		Render(lipgloss.JoinVertical(lipgloss.Left, header, "", graphBox, detailBox, legend, statsLine))
	return lipgloss.JoinVertical(lipgloss.Right, panel, hints)
}

func viewVaultWizard(m model) string {
	q := vaultWizardQuestions[m.vaultWizardStep]
	progress := fmt.Sprintf("Paso %d/%d", m.vaultWizardStep+1, len(vaultWizardQuestions))
	title := lipgloss.NewStyle().Bold(true).Foreground(colAccent).Render("VAULT  —  Configuración")
	sub := lipgloss.NewStyle().Foreground(colMuted).Render(progress + "  •  /vault  •  " + q.Title + "  •  " + m.vaultWizardData.Name)
	header := lipgloss.JoinVertical(lipgloss.Left, title, sub, "")
	filled := (m.vaultWizardStep + 1) * 20 / len(vaultWizardQuestions)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 20-filled)
	barLine := lipgloss.NewStyle().Foreground(colAccent).Render(bar) + lipgloss.NewStyle().Foreground(colMuted).Render(fmt.Sprintf(" %d%%", (m.vaultWizardStep+1)*100/len(vaultWizardQuestions)))
	question := lipgloss.NewStyle().Bold(true).Foreground(colWhite).Render(q.Prompt)
	var body string
	if len(q.Options) > 0 {
		var opts []string
		for i, opt := range q.Options {
			style := lipgloss.NewStyle().Foreground(colText2).Padding(0, 1).Width(28)
			if i == m.vaultWizardCursor {
				style = lipgloss.NewStyle().Foreground(colPanel).Background(colAccent).Bold(true).Padding(0, 1).Width(28)
			}
			check := "  "
			if opt == m.vaultWizardInput {
				check = "▶ "
			}
			desc := ""
			if q.Key == "Purpose" {
				if opt == "personal" { desc = "  privado" }
				if opt == "equipo" { desc = "  compartido" }
				if opt == "proyecto" { desc = "  por repo" }
			}
			if q.Key == "VaultType" {
				if opt == "knowledge" { desc = "  docs" }
				if opt == "memory" { desc = "  prefs" }
				if opt == "mixed" { desc = "  ambos" }
			}
			opts = append(opts, style.Render(check+opt+lipgloss.NewStyle().Foreground(colMuted).Render(desc)))
		}
		body = strings.Join(opts, "\n")
	} else {
		cursor := "█"
		input := m.vaultWizardInput
		if input == "" {
			input = lipgloss.NewStyle().Foreground(colMuted).Italic(true).Render(q.Default)
		} else {
			input = lipgloss.NewStyle().Foreground(colWhite).Render(input) + lipgloss.NewStyle().Background(colAccent).Foreground(colPanel).Render(cursor)
		}
		box := lipgloss.NewStyle().Background(colPanel2).BorderStyle(lipgloss.RoundedBorder()).BorderForeground(colAccent).Padding(0, 1).Width(52).Render(input)
		body = box + "\n" + lipgloss.NewStyle().Foreground(colMuted).Italic(true).Render("Escribe • Enter confirma • Esc atrás")
	}
	var summary []string
	for i := 0; i < m.vaultWizardStep; i++ {
		qq := vaultWizardQuestions[i]
		var val string
		switch qq.Key {
		case "Path": val = m.vaultWizardData.Path
		case "Name": val = m.vaultWizardData.Name
		case "Purpose": val = m.vaultWizardData.Purpose
		case "VaultType": val = m.vaultWizardData.VaultType
		case "Sync": val = m.vaultWizardData.Sync
		case "Indexing": val = m.vaultWizardData.Indexing
		case "Embeddings": val = m.vaultWizardData.Embeddings
		}
		icon := lipgloss.NewStyle().Foreground(colSuccess).Render("✓ ")
		summary = append(summary, icon+lipgloss.NewStyle().Foreground(colMuted).Render(qq.Title+": ")+lipgloss.NewStyle().Foreground(colWhite).Render(val))
	}
	summaryBox := ""
	if len(summary) > 0 {
		summaryBox = lipgloss.NewStyle().Background(colPanel2).BorderStyle(lipgloss.NormalBorder()).BorderForeground(colBorder).Padding(0, 1).Width(52).Render(strings.Join(summary, "\n"))
	}
	// preview vault structure (derecha)
	previewTitle := lipgloss.NewStyle().Bold(true).Foreground(colWhite).Render("Preview  ") + lipgloss.NewStyle().Foreground(colMuted).Render(m.vaultWizardData.Path)
	previewTree := []string{
		lipgloss.NewStyle().Foreground(colAccent).Render(m.vaultWizardData.Name+"/"),
		lipgloss.NewStyle().Foreground(colMuted).Render("├── ") + lipgloss.NewStyle().Foreground(colSuccess).Render("memory/"),
		lipgloss.NewStyle().Foreground(colMuted).Render("│   ├── ") + lipgloss.NewStyle().Foreground(colMuted2).Render("preferences.md"),
		lipgloss.NewStyle().Foreground(colMuted).Render("│   └── ") + lipgloss.NewStyle().Foreground(colMuted2).Render("index.jsonl"),
		lipgloss.NewStyle().Foreground(colMuted).Render("├── ") + lipgloss.NewStyle().Foreground(colPurple).Render("agents/"),
		lipgloss.NewStyle().Foreground(colMuted).Render("├── ") + lipgloss.NewStyle().Foreground(colWarn).Render("knowledge/"),
		lipgloss.NewStyle().Foreground(colMuted).Render("└── ") + lipgloss.NewStyle().Foreground(colMuted2).Render("config.json"),
	}
	previewBox := lipgloss.NewStyle().Background(colPanel2).BorderStyle(lipgloss.NormalBorder()).BorderForeground(colBorder).Padding(0, 1).Width(28).Height(9).Render(previewTitle + "\n" + strings.Join(previewTree, "\n"))
	leftContent := lipgloss.JoinVertical(lipgloss.Left, question, "", body)
	if summaryBox != "" {
		leftContent = lipgloss.JoinVertical(lipgloss.Left, leftContent, "", summaryBox)
	}
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(42).Render(leftContent), "  ", previewBox)
	panel := lipgloss.NewStyle().Background(colPanel).BorderStyle(lipgloss.NormalBorder()).BorderForeground(colBorder2).Padding(1, 2).Width(76).Render(lipgloss.JoinVertical(lipgloss.Left, header, barLine, "", mainRow))
	hints := lipgloss.NewStyle().Foreground(colMuted).Render(
		lipgloss.NewStyle().Bold(true).Foreground(colText).Render("enter")+" confirmar  "+
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("↑↓")+" elegir  "+
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("esc")+" atrás  "+
			lipgloss.NewStyle().Bold(true).Foreground(colText).Render("o")+" abrir",
	)
	return lipgloss.JoinVertical(lipgloss.Right, panel, hints)
}

func viewChat(m model) string {
    // Fullscreen chat — opencode-style but with Huginn improvements
    w := m.width
    h := m.height
    if w < 80 {
        w = 80
    }
    if h < 24 {
        h = 24
    }
    innerW := w - 4
    if innerW < 80 {
        innerW = 80
    }
    leftW := int(float64(innerW) * 0.68)
    rightW := innerW - leftW - 1
    if leftW < 50 {
        leftW = 50
    }
    if rightW < 20 {
        rightW = 20
    }
    // Use full height for rows
    totalRows := h - 8
    if totalRows < 20 {
        totalRows = 20
    }
    pad := func(s string, w int) string {
        lw := lipgloss.Width(s)
        if lw >= w {
            return s
        }
        return s + strings.Repeat(" ", w-lw)
    }
    truncate := func(s string, w int) string {
        if lipgloss.Width(s) <= w {
            return s
        }
        runes := []rune(s)
        for len(runes) > 0 && lipgloss.Width(string(runes)) > w-1 {
            runes = runes[:len(runes)-1]
        }
        return string(runes) + "…"
    }
    // wrap para respuesta completa responsive — rompe por palabras, respeta ancho leftW
    wrapLines := func(s string, w int) []string {
        if w <= 10 {
            w = 10
        }
        var lines []string
        words := strings.Fields(s)
        if len(words) == 0 {
            return []string{s}
        }
        cur := words[0]
        for _, wd := range words[1:] {
            if lipgloss.Width(cur+" "+wd) <= w {
                cur += " " + wd
            } else {
                lines = append(lines, cur)
                cur = wd
                // palabra muy larga — corta duro
                for lipgloss.Width(cur) > w {
                    runes := []rune(cur)
                    cut := ""
                    for _, r := range runes {
                        if lipgloss.Width(cut+string(r)) > w-1 {
                            break
                        }
                        cut += string(r)
                    }
                    lines = append(lines, cut)
                    cur = string([]rune(cur)[len([]rune(cut)):])
                }
            }
        }
        lines = append(lines, cur)
        return lines
    }
    row := func(left, right string) string {
        return "│" + pad(left, leftW) + "│" + pad(right, rightW) + "│"
    }
    top := "┌" + strings.Repeat("─", innerW) + "┐"
    midDiv := "├" + strings.Repeat("─", leftW) + "┬" + strings.Repeat("─", rightW) + "┤"
    botMid := "├" + strings.Repeat("─", leftW) + "┴" + strings.Repeat("─", rightW) + "┤"
    bottom := "└" + strings.Repeat("─", innerW) + "┘"
    hdrLeft := lipgloss.NewStyle().Bold(true).Foreground(colWhite).Render(" HUGINN")
    hdrMid := lipgloss.NewStyle().Foreground(colMuted).Render("AI ORCHESTRATOR")
    hdrRight := lipgloss.NewStyle().Foreground(colSuccess).Render("● CONNECTED")
    // header centered
    hdrContentWidth := leftW + 1 + rightW
    hdrInner := pad(hdrLeft, 20) + pad(hdrMid, hdrContentWidth-46) + pad(hdrRight, 26)
    hdr := "│" + hdrInner + "│"
    prompt := "Implement authentication with JWT"
    if len(m.chatHistory) > 0 {
        for i:=len(m.chatHistory)-1; i>=0; i-- {
            if m.chatHistory[i].IsUser && strings.TrimSpace(m.chatHistory[i].Text) != "" {
                prompt = m.chatHistory[i].Text
                break
            }
        }
    }
    if lipgloss.Width(prompt) > 40 {
        prompt = truncate(prompt, 40)
    }
    agentStatus := func(name string) (string, color.Color) {
        for _, a := range m.agents {
            if strings.EqualFold(a.name, name) {
                switch a.status {
                case statusWorking:
                    return "WORKING", colAccent
                case statusDone:
                    return "READY", colSuccess
                case statusTesting:
                    return "REVIEW", lipgloss.Color("#E1A451")
                default:
                    return "READY", colSuccess
                }
            }
        }
        if name=="ChatGPT" {
            return "READY", colSuccess
        }
        return "READY", colSuccess
    }
    inputLine := m.chatInput
    if strings.TrimSpace(inputLine)=="" {
        inputLine = lipgloss.NewStyle().Foreground(colMuted).Italic(true).Render("escribe con @chatgpt @opencode @all …")
    } else {
        before := inputLine[:m.chatCursor]
        after := ""
        cur := " "
        if m.chatCursor < len(inputLine) {
            cur = string(inputLine[m.chatCursor])
            if m.chatCursor+1 < len(inputLine) {
                after = inputLine[m.chatCursor+1:]
            }
        }
        caret := lipgloss.NewStyle().Background(colWhite).Foreground(lipgloss.Color("#130E0A")).Render(cur)
        inputLine = lipgloss.NewStyle().Foreground(colWhite).Render(before) + caret + lipgloss.NewStyle().Foreground(colWhite).Render(after)
    }
    _ = lipgloss.NewStyle().Foreground(colAccent).Render("> ") + inputLine
    // híbrido opencode: left limpio solo chat history + input (sin dispatch duplicado)
    agents := []string{"ChatGPT", "OpenCode", "Kilo Code", "Mimo Code", "Muse Code"}
    leftLines := []string{""}
    if len(m.chatHistory) <= 1 {
        // splash mascota en vacio — state-based sprite
        mascotState := tuicomp.StateWise
        if m.mascotState == "thinking" {
            mascotState = tuicomp.StateThinking
        }
        for _, l := range strings.Split(tuicomp.RenderMascotSprite(mascotState, ""), "\n") {
            if strings.TrimSpace(l) != "" {
                leftLines = append(leftLines, "  "+l)
            }
        }
        leftLines = append(leftLines, lipgloss.NewStyle().Foreground(colMuted).Italic(true).Render("  Huginn vuela, recuerda, conecta mundos."))
        leftLines = append(leftLines, "")
    }
    if lipgloss.Width(prompt) > 0 && len(m.chatHistory) <= 1 {
        leftLines = append(leftLines, lipgloss.NewStyle().Foreground(colMuted).Render("  › ")+lipgloss.NewStyle().Foreground(colText2).Italic(true).Render(truncate(prompt, leftW-8)))
        leftLines = append(leftLines, "")
    }
    if len(m.chatHistory) > 0 {
        // muestra todas las que caben según altura disponible — responsive
        // altura disponible para mensajes = totalRows - inputBox(3) - rightHeader(2) - padding
        availRows := totalRows - 10
        if availRows < 8 {
            availRows = 8
        }
        hist := m.chatHistory
        // recolecta líneas envueltas y corta por availRows con scroll
        var rendered []string
        for _, msg := range hist {
            ts := ""
            if msg.Time != "" {
                ts = lipgloss.NewStyle().Foreground(colMuted).Render("["+msg.Time+"] ")
            }
            if msg.IsUser {
                prefix := ts + lipgloss.NewStyle().Foreground(colAccent).Render("▶ You: ")
                pw := leftW - lipgloss.Width(prefix) - 4
                if pw < 20 {
                    pw = 20
                }
                for i, wl := range wrapLines(msg.Text, pw) {
                    if i == 0 {
                        rendered = append(rendered, prefix+lipgloss.NewStyle().Foreground(colWhite).Render(wl))
                    } else {
                        rendered = append(rendered, strings.Repeat(" ", lipgloss.Width(prefix))+lipgloss.NewStyle().Foreground(colWhite).Render(wl))
                    }
                }
            } else {
                if strings.Contains(msg.Text, "… escribiendo") {
                    line := lipgloss.NewStyle().Foreground(colMuted).Render("    ◌ " + msg.From + " escribiendo…")
                    rendered = append(rendered, line)
                } else {
                    col := colText2
                    switch msg.From {
                    case "ChatGPT":
                        col = colPurple
                    case "OpenCode":
                        col = colAccent
                    case "Kilo Code":
                        col = colSuccess
                    case "Mimo Code":
                        col = lipgloss.Color("#E1A451")
                    case "Muse Code":
                        col = lipgloss.Color("#E1A451")
                    case "Hugin":
                        col = colWarn
                    }
                    prefix := ts + lipgloss.NewStyle().Bold(true).Foreground(col).Render(" " + msg.From + ": ")
                    pw := leftW - lipgloss.Width(prefix) - 4
                    if pw < 20 {
                        pw = 20
                    }
                    for i, wl := range wrapLines(msg.Text, pw) {
                        if i == 0 {
                            rendered = append(rendered, prefix+lipgloss.NewStyle().Foreground(colText2).Render(wl))
                        } else {
                            rendered = append(rendered, strings.Repeat(" ", lipgloss.Width(prefix))+lipgloss.NewStyle().Foreground(colText2).Render(wl))
                        }
                    }
                }
            }
        }
        // responsive: corta por altura disponible y respeta scroll (PgUp/PgDn)
        if len(rendered) > availRows {
            start := len(rendered) - availRows - m.chatScroll
            if start < 0 {
                start = 0
            }
            end := start + availRows
            if end > len(rendered) {
                end = len(rendered)
            }
            rendered = rendered[start:end]
        } else if m.chatScroll > 0 && m.chatScroll < len(rendered) {
            // si hay menos líneas que avail pero hay scroll, no hacer nada extra
        }
        leftLines = append(leftLines, rendered...)
        leftLines = append(leftLines, "")
    }
    inputBoxContent := lipgloss.NewStyle().Foreground(colAccent).Render("> ") + inputLine + "\n" + lipgloss.NewStyle().Foreground(colMuted).Render("  "+m.pkgManager+" · Enter enviar • Esc limpiar • Tab agente")
    inputBox := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(colAccent).Background(colPanel2).Padding(0, 1).Width(leftW-4).Render(inputBoxContent)
    for _, line := range strings.Split(inputBox, "\n") {
        leftLines = append(leftLines, line)
    }
    rightLines := []string{
        lipgloss.NewStyle().Bold(true).Foreground(colWhite).Render(" AGENTS"),
        "",
    }
    for _, an := range agents {
        st, col := agentStatus(an)
        dot := lipgloss.NewStyle().Foreground(col).Render("●")
        name := lipgloss.NewStyle().Foreground(colText).Render(pad(an, 10))
        status := lipgloss.NewStyle().Foreground(col).Render(st)
        rightLines = append(rightLines, fmt.Sprintf(" %s %s %s", dot, name, status))
    }
    rightLines = append(rightLines, "")
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colBorder).Render(strings.Repeat("─", rightW-2)))
    rightLines = append(rightLines, lipgloss.NewStyle().Bold(true).Foreground(colWhite).Render(" TASK PIPELINE"))
    rightLines = append(rightLines, "")
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colSuccess).Render(" ● Research"))
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colSuccess).Render("   ✓ completed"))
    rightLines = append(rightLines, "")
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colAccent).Render(" ● Architecture"))
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colAccent).Render("   ● running"))
    rightLines = append(rightLines, "")
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colMuted).Render(" ○ Implementation"))
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colMuted).Render("   queued"))
    rightLines = append(rightLines, "")
    rightLines = append(rightLines, lipgloss.NewStyle().Bold(true).Foreground(colWhite).Render(" CONTEXT"))
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colBorder).Render(strings.Repeat("─", rightW-2)))
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colMuted).Render(" Project"))
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colText2).Render(" └─ huginn-tui"))
    rightLines = append(rightLines, "")
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colMuted).Render(" Memory"))
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colMuted).Render(" ├─ auth decisions"))
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colMuted).Render(" ├─ agent history"))
    rightLines = append(rightLines, lipgloss.NewStyle().Foreground(colMuted).Render(" └─ project context"))
    // dynamic rows based on height
    for len(leftLines) < totalRows {
        leftLines = append(leftLines, "")
    }
    for len(rightLines) < totalRows {
        rightLines = append(rightLines, "")
    }
    if len(leftLines) > totalRows {
        leftLines = leftLines[len(leftLines)-totalRows:]
    }
    if len(rightLines) > totalRows {
        rightLines = rightLines[:totalRows]
    }
    var rows []string
    rows = append(rows, top)
    rows = append(rows, hdr)
    rows = append(rows, midDiv)
    for i := 0; i < totalRows; i++ {
        l := leftLines[i]
        r := rightLines[i]
        rows = append(rows, row(l, r))
    }
    rows = append(rows, botMid)
    bottomContent := lipgloss.NewStyle().Foreground(colMuted).Render(" TAB") + lipgloss.NewStyle().Foreground(colText).Render(" Agent   ") +
        lipgloss.NewStyle().Foreground(colMuted).Render("CTRL+P") + lipgloss.NewStyle().Foreground(colText).Render(" Commands   ") +
        lipgloss.NewStyle().Foreground(colMuted).Render("CTRL+B") + lipgloss.NewStyle().Foreground(colText).Render(" Background   ") +
        lipgloss.NewStyle().Foreground(colMuted).Render("CTRL+G") + lipgloss.NewStyle().Foreground(colText).Render(" Graph   ") +
        lipgloss.NewStyle().Foreground(colMuted).Render("CTRL+L") + lipgloss.NewStyle().Foreground(colText).Render(" Clear   ") +
        lipgloss.NewStyle().Foreground(colMuted).Render("?") + lipgloss.NewStyle().Foreground(colText).Render(" Help")
    rows = append(rows, "│" + pad(bottomContent, innerW) + "│")
    rows = append(rows, bottom)
    content := strings.Join(rows, "\n")
    // Fullscreen: use full terminal size, no extra outer border
    return content
}


// ===================== CLI LAYER =====================
const huginnBanner = "Huginn — Agent Vault Intelligence"

type cliArgs struct {
    Path       string
    Prompt     string
    Subcommand string
    SubArgs    []string
    Help       bool
    Version    bool
    Debug      bool
}

var futureSubcommands = map[string]bool{
    "chat": true, "agent": true, "agents": true, "memory": true, "vault": true, "config": true,
}

func isDebug() bool {
    if os.Getenv("HUGINN_DEBUG") == "1" || os.Getenv("HUGINN_DEBUG") == "true" {
        return true
    }
    for _, a := range os.Args[1:] {
        if a == "--debug" {
            return true
        }
    }
    return false
}

func isDirectory(p string) bool {
    if p == "" {
        return false
    }
    p = strings.Trim(p, `"'`)
    clean := filepath.Clean(p)
    info, err := os.Stat(clean)
    if err != nil {
        return false
    }
    return info.IsDir()
}

func detectPackageManager(root string) string {
    if _, err := os.Stat(filepath.Join(root, "bun.lockb")); err == nil {
        return "bun"
    }
    if _, err := os.Stat(filepath.Join(root, "pnpm-lock.yaml")); err == nil {
        return "pnpm"
    }
    if _, err := os.Stat(filepath.Join(root, "yarn.lock")); err == nil {
        return "yarn"
    }
    if _, err := os.Stat(filepath.Join(root, "package-lock.json")); err == nil {
        return "npm"
    }
    if _, err := os.Stat(filepath.Join(root, "package.json")); err == nil {
        return "npm"
    }
    return ""
}

func detectProject(root string) (bool, string) {
    markers := []string{"go.mod", "package.json", "pyproject.toml", "Cargo.toml", "README.md", ".git", "AGENTS.md", "huginn.json", "huginn.config.json"}
    for _, m := range markers {
        if _, err := os.Stat(filepath.Join(root, m)); err == nil {
            return true, m
        }
    }
    return false, ""
}

func resolveVaultPath() (string, bool) {
    if v := os.Getenv("HUGINN_VAULT"); v != "" {
        if _, err := os.Stat(v); err == nil {
            return v, true
        }
        return v, false
    }
    if v := os.Getenv("AGENT_VAULT"); v != "" {
        if _, err := os.Stat(v); err == nil {
            return v, true
        }
        return v, false
    }
    home, _ := os.UserHomeDir()
    candidates := []string{
        filepath.Join(home, "agent-vault"),
        filepath.Join(home, "huginn-vault"),
        filepath.Join(home, ".huginn", "vault"),
    }
    for _, c := range candidates {
        if _, err := os.Stat(c); err == nil {
            return c, true
        }
    }
    if len(candidates) > 0 {
        return candidates[0], false
    }
    return "", false
}

func parseArgs(raw []string) (cliArgs, error) {
    var out cliArgs
    var positional []string
    for _, a := range raw {
        switch a {
        case "--help", "-h", "-help", "help":
            if len(positional) == 0 && len(raw) == 1 {
                out.Help = true
                return out, nil
            }
            out.Help = true
        case "--version", "-v", "-version", "version":
            if len(positional) == 0 && len(raw) == 1 {
                out.Version = true
                return out, nil
            }
            out.Version = true
        case "--debug":
            out.Debug = true
        default:
            if strings.HasPrefix(a, "--") {
                return out, fmt.Errorf("flag desconocido: %s", a)
            }
            if strings.HasPrefix(a, "-") && len(a) > 1 {
                return out, fmt.Errorf("flag desconocido: %s", a)
            }
            positional = append(positional, a)
        }
    }
    if out.Help || out.Version {
        return out, nil
    }
    if len(positional) > 0 && futureSubcommands[positional[0]] {
        out.Subcommand = positional[0]
        out.SubArgs = positional[1:]
        return out, nil
    }
    if len(positional) == 0 {
        out.Path = "."
        return out, nil
    }
    if len(positional) == 1 {
        p := positional[0]
        if p == "." || p == "./" || isDirectory(p) {
            out.Path = p
        } else {
            isAbs := filepath.IsAbs(p) || (len(p) >= 2 && p[1] == ':')
            if isAbs || (strings.Contains(p, string(os.PathSeparator)) && isDirectory(p)) {
                out.Path = p
            } else {
                out.Prompt = p
                out.Path = "."
            }
        }
        return out, nil
    }
    first := positional[0]
    if first == "." || first == "./" || isDirectory(first) || filepath.IsAbs(first) || (len(first) >= 2 && first[1] == ':') {
        if isDirectory(first) || filepath.IsAbs(first) || strings.Contains(first, string(os.PathSeparator)) {
            out.Path = first
            out.Prompt = strings.Join(positional[1:], " ")
            return out, nil
        }
    }
    out.Prompt = strings.Join(positional, " ")
    out.Path = "."
    return out, nil
}

func printHelp() {
    useColor := os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
    accent := ""
    reset := ""
    if useColor {
        accent = "\x1b[36m"
        reset = "\x1b[0m"
    }
    fmt.Printf(`
  %sHuginn%s — Agent Vault Intelligence
  CLI oficial de Huginn (interfaz de Agent Vault)

%sUso:%s
  huginn                      Abre Huginn en el directorio actual
  huginn .                    Igual que arriba (explícito)
  huginn <path>               Abre Huginn con <path> como contexto
  huginn "<prompt>"           Envía prompt directo al orquestador
  huginn <path> "<prompt>"    Combina contexto + prompt
  huginn --help               Muestra esta ayuda
  huginn --version            Muestra versión

%sComandos futuros (arquitectura preparada):%s
  huginn chat                 Chat interactivo (alias de huginn)
  huginn agent                Seleccionar agente
  huginn agents               Listar agentes
  huginn memory               Consultar memoria
  huginn vault                Explorar vault
  huginn config               Gestionar configuración

%sEjemplos:%s
  huginn
  huginn .
  huginn ./mi-proyecto
  huginn C:\Projects\mi-proyecto
  huginn "analiza este proyecto"
  huginn "crea un sistema de autenticación"
  huginn ./mi-proyecto "analiza este proyecto"

%sFlags:%s
  -h, --help                  Ayuda
  -v, --version               Versión
  --debug                     Modo debug (stack traces)

`, accent, reset, accent, reset, accent, reset, accent, reset, accent, reset)
}

func printVersion() {
    fmt.Printf("huginn %s\n", VERSION)
}

func huginnError(msg string, debug bool, err error) {
    fmt.Fprintf(os.Stderr, "\n  Huginn error: %s\n\n", msg)
    if err != nil && debug {
        fmt.Fprintf(os.Stderr, "  detalle: %v\n\n", err)
    }
    if strings.Contains(strings.ToLower(msg), "vault") {
        fmt.Fprintln(os.Stderr, "Check:")
        fmt.Fprintln(os.Stderr, "  • Agent Vault is running")
        fmt.Fprintln(os.Stderr, "  • Your configuration is correct")
        fmt.Fprintln(os.Stderr, "  • The configured endpoint is reachable")
        fmt.Fprintln(os.Stderr, "")
        fmt.Fprintln(os.Stderr, "Usa --debug para más detalle.")
    }
}

func runTUI(projectPath, prompt string) error {
    abs, err := filepath.Abs(projectPath)
    if err != nil {
        abs = projectPath
    }
    if projectPath != "." {
        if _, err := os.Stat(abs); err != nil {
            return fmt.Errorf("ruta no encontrada: %s", abs)
        }
        if !isDirectory(abs) {
            return fmt.Errorf("la ruta no es un directorio: %s", abs)
        }
    }
    m := initialModelWithContext(abs, prompt)
    p := tea.NewProgram(m)
    if _, err := p.Run(); err != nil {
        return err
    }
    return nil
}

func dumpViews() {
	cases := []struct {
		name string
		mode mode
		setup func(*model)
	}{
		{"chat", modeChat, func(m *model) {
			m.chatHistory = []chatMsg{
				{From: "Hugin", Text: "Chat iniciado. Menciona con @chatgpt @opencode @kilo @mimo @muse @all  •  o usa Tab/1-6", IsUser: false, Time: "20:24"},
				{From: "You", Text: "analiza este proyecto en detalle y explica la arquitectura hexagonal y el sistema de Vault", IsUser: true, Time: "20:24"},
				{From: "ChatGPT", Text: "Detectado Go 1.25 + Bubble Tea v2. Proyecto limpio con 4 agentes (ChatGPT, OpenCode, Kilo y Muse). Arquitectura hexagonal: cmd/huginn → cli → application/orchestrator → domain/agent/vault → infrastructure. Vault .huginn con config.json, vault.json, memory.jsonl y user/*.md. Este texto largo debe envolverse automaticamente y adaptarse al ancho de la pantalla sin truncarse, mostrando la respuesta completa de forma fluida y premium, con scroll via PgUp/PgDn.", IsUser: false, Time: "20:25"},
			}
			m.chatInput = "explica el vault"
			m.chatCursor = len(m.chatInput)
		}},
		{"vault", modeVault, func(m *model) { m.vaultWizardStep = 2 }},
		{"graph", modeGraph, func(m *model) { m.graphCursor = 3 }},
		{"settings", modeSettings, func(m *model) {}},
		{"servers", modeServers, func(m *model) { m.serversTab = 0; m.serversCursor = 1 }},
		{"memory", modeStatus, func(m *model) {}},
	}
	for _, c := range cases {
		m := initialModelWithContext(".", "")
		m.width = 120
		m.height = 38
		m.mode = c.mode
		m.projectPath, _ = filepath.Abs(".")
		m.projectName = "huginn-tui"
		m.pkgManager = "go"
		m.vaultPath = "~/agent-vault"
		m.vaultOK = true
		if c.setup != nil {
			c.setup(&m)
		}
		var out string
		switch c.mode {
		case modeChat:
			out = viewChat(m)
		case modeVault:
			out = viewVaultWizard(m)
		case modeGraph:
			out = viewGraph(m)
		case modeSettings:
			out = viewSettings(m)
		case modeServers:
			out = viewServers(m)
		case modeStatus:
			out = viewStatus(m)
		default:
			v := m.View()
			out = fmt.Sprint(v)
		}
		// strip ANSI for preview and also save raw ANSI
		_ = os.WriteFile("screenshot_"+c.name+".ansi", []byte(out), 0644)
		// also generate text preview without ANSI for Python fallback
		plain := stripAnsi(out)
		_ = os.WriteFile("screenshot_"+c.name+".txt", []byte(plain), 0644)
	}
	fmt.Println("dumped views to screenshot_*.ansi/.txt")
	os.Exit(0)
}

func stripAnsi(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inEsc := false
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			inEsc = true
			i++
			continue
		}
		if inEsc {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEsc = false
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--dump-ansi" {
		dumpViews()
	}
	args := os.Args[1:]
	parsed, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "huginn: %v\n", err)
		fmt.Fprintln(os.Stderr, "Usa: huginn --help")
		os.Exit(2)
	}
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	if parsed.Version {
		printVersion()
		os.Exit(0)
	}
	if parsed.Subcommand != "" {
		// arquitectura extensible: stubs para futuros subcomandos sin reescribir CLI
		switch parsed.Subcommand {
		case "chat":
			// alias a huginn [path] [prompt]
			path := "."
			prompt := ""
			if len(parsed.SubArgs) > 0 {
				// primer subArg puede ser path
				if isDirectory(parsed.SubArgs[0]) || parsed.SubArgs[0] == "." {
					path = parsed.SubArgs[0]
					if len(parsed.SubArgs) > 1 {
						prompt = strings.Join(parsed.SubArgs[1:], " ")
					}
				} else {
					prompt = strings.Join(parsed.SubArgs, " ")
				}
			}
			if err := runTUI(path, prompt); err != nil {
				huginnError(err.Error(), parsed.Debug || isDebug(), err)
				os.Exit(1)
			}
			os.Exit(0)
		case "graph":
			path := "."
			if len(parsed.SubArgs) > 0 && (isDirectory(parsed.SubArgs[0]) || parsed.SubArgs[0] == ".") {
				path = parsed.SubArgs[0]
			}
			if err := runTUI(path, ""); err != nil {
				huginnError(err.Error(), parsed.Debug || isDebug(), err)
				os.Exit(1)
			}
			os.Exit(0)
		case "vault":
			mgr := infravault.NewFilesystemManager()
			if len(parsed.SubArgs) == 0 {
				// huginn vault → muestra vault actual o detecta
				if v, ok := mgr.GetCurrent(); ok {
					fmt.Printf("Vault actual: %s\nPath: %s\nID: %s\n", v.Name, v.Path, v.ID)
				} else if v, ok := mgr.Detect("."); ok {
					fmt.Printf("Vault detectado: %s\nPath: %s\n", v.Name, v.Path)
				} else {
					fmt.Println("No Vault selected. Usa: huginn vault open <path> | huginn vault create <name>")
					recent, _ := mgr.Recent()
					if len(recent) > 0 {
						fmt.Println("\nRecent Vaults:")
						for _, r := range recent {
							fmt.Printf("  - %s (%s)\n", r.Name, r.Path)
						}
					}
				}
				os.Exit(0)
			}
			sub := parsed.SubArgs[0]
			switch sub {
			case "open":
				path := "."
				if len(parsed.SubArgs) > 1 {
					path = parsed.SubArgs[1]
				}
				v, err := mgr.Open(nil, path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "vault open failed: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Vault abierto: %s (%s)\n", v.Name, v.Path)
			case "create":
				if len(parsed.SubArgs) < 2 {
					fmt.Fprintln(os.Stderr, "uso: huginn vault create <name> [parentDir]")
					os.Exit(2)
				}
				name := parsed.SubArgs[1]
				parent := "."
				if len(parsed.SubArgs) > 2 {
					parent = parsed.SubArgs[2]
				}
				v, err := mgr.Create(nil, parent, name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "vault create failed: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Vault creado: %s (%s)\n", v.Name, v.Path)
			case "list", "recent":
				recent, _ := mgr.Recent()
				if len(recent) == 0 {
					fmt.Println("No recent vaults")
				} else {
					fmt.Println("Recent Vaults:")
					for _, r := range recent {
						fmt.Printf("  %s — %s\n", r.Name, r.Path)
					}
				}
			case "current":
				v, ok := mgr.GetCurrent()
				if !ok {
					fmt.Println("No vault open")
				} else {
					fmt.Printf("%s (%s)\n", v.Name, v.Path)
				}
			case "init":
				path := "."
				if len(parsed.SubArgs) > 1 {
					path = parsed.SubArgs[1]
				}
				v, err := mgr.Initialize(nil, path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "vault init failed: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Vault inicializado: %s (%s)\n", v.Name, v.Path)
			default:
				// huginn vault <path> → open
				v, err := mgr.Open(nil, sub)
				if err != nil {
					fmt.Fprintf(os.Stderr, "vault open failed: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Vault abierto: %s (%s)\n", v.Name, v.Path)
			}
			os.Exit(0)
		case "agent", "agents", "memory", "config":
			fmt.Printf("huginn %s — (subcomando preparado, aún no implementado)\n", parsed.Subcommand)
			fmt.Println("Usa: huginn --help para ver comandos disponibles.")
			fmt.Println("El CLI está preparado para agregar este subcomando sin reescribir la arquitectura.")
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "huginn: subcomando desconocido: %s\n", parsed.Subcommand)
			os.Exit(2)
		}
	}
	// flujo principal: huginn | huginn <path> | huginn "<prompt>" | huginn <path> "<prompt>"
	projectPath := parsed.Path
	if projectPath == "" {
		projectPath = "."
	}
	prompt := parsed.Prompt

	// modo debug global
	_ = parsed.Debug

	if err := runTUI(projectPath, prompt); err != nil {
		huginnError(err.Error(), parsed.Debug || isDebug(), err)
		os.Exit(1)
	}
}
