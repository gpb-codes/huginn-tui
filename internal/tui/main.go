// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"huginn/internal/cli"
	"huginn/internal/domain/project"
	"huginn/internal/domain/vault"
	"huginn/internal/infrastructure/config"
	infravault "huginn/internal/infrastructure/vault"
)

const Version = "v0.2.0"

// Main es la entrada de la TUI. Parsea args con internal/cli y lanza bubbletea.
func Main() {
	if len(os.Args) > 1 && os.Args[1] == "--dump-ansi" {
		// Dump homescreen + chat con sidebar como en la captura (para landing/screenshots)
		m := New(".", "")
		m.width = 120
		m.height = 36
		fmt.Println("=== HOMESCREEN ===")
		fmt.Println(m.View().Content)
		// Chat de ejemplo con sidebar como en la captura
		m2 := New("huginn-tui", "~/agent-vault")
		m2.width = 120
		m2.height = 36
		m2.messages = []chatMsg{
			{Role: "system", Text: "Chat iniciado. Menciona con @chatgpt @opencode @kilo @mimo @muse @all • o usa Tab/1-6", Meta: "Hugin"},
			{Role: "user", Text: "analiza este proyecto en detalle y explica la arquitectura hexagonal y el sistema de Vault", Meta: "Tú → Huginn"},
			{Role: "assistant", Text: "Detectado Go 1.25 + Bubble Tea v2. Proyecto limpio con 4 agentes (ChatGPT, OpenCode, Kilo y Muse). Arquitectura hexagonal: cmd/huginn → cli → application/orchestrator → domain/agent/vault → infrastructure. Vault .huginn con config.json, vault.json, memory.jsonl y user/*.md.", Meta: "ChatGPT"},
		}
		m2.input = "explica el vault"
		fmt.Println("\n=== CHAT ===")
		fmt.Println(m2.View().Content)
		os.Exit(0)
	}

	args, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "huginn: %v\n", err)
		fmt.Fprintln(os.Stderr, "Usa: huginn --help")
		os.Exit(2)
	}
	if args.Help {
		cli.PrintHelp()
		os.Exit(0)
	}
	if args.Version {
		cli.PrintVersion(Version)
		os.Exit(0)
	}
	if args.Subcommand != "" {
		handleSubcommand(args)
	}

	projectPath := args.Path
	if projectPath == "" {
		projectPath = "."
	}
	prompt := args.Prompt

	// Detección de contexto (proyecto/vault) delegada a domain, no a la TUI.
	abs, _ := filepath.Abs(projectPath)
	vaultPath, _ := vault.ResolveVaultPath()
	// Carga config si existe (no falla si no existe)
	_, _ = config.Load(abs)

	if err := runTUI(abs, vaultPath, prompt); err != nil {
		fmt.Fprintf(os.Stderr, "huginn: %v\n", err)
		os.Exit(1)
	}
}

func handleSubcommand(args cli.Args) {
	switch args.Subcommand {
	case "vault":
		mgr := infravault.NewFilesystemManager()
		if len(args.SubArgs) == 0 {
			if v, ok := mgr.GetCurrent(); ok {
				fmt.Printf("Vault actual: %s\nPath: %s\nID: %s\n", v.Name, v.Path, v.ID)
			} else if v, ok := mgr.Detect("."); ok {
				fmt.Printf("Vault detectado: %s\nPath: %s\n", v.Name, v.Path)
			} else {
				fmt.Println("No Vault selected. Usa: huginn vault open <path> | huginn vault create <name>")
			}
			os.Exit(0)
		}
		sub := args.SubArgs[0]
		switch sub {
		case "open":
			path := "."
			if len(args.SubArgs) > 1 {
				path = args.SubArgs[1]
			}
			v, err := mgr.Open(nil, path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "vault open failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Vault abierto: %s (%s)\n", v.Name, v.Path)
		case "create":
			if len(args.SubArgs) < 2 {
				fmt.Fprintln(os.Stderr, "uso: huginn vault create <name> [parentDir]")
				os.Exit(2)
			}
			name := args.SubArgs[1]
			parent := "."
			if len(args.SubArgs) > 2 {
				parent = args.SubArgs[2]
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
				for _, r := range recent {
					fmt.Printf("  %s — %s\n", r.Name, r.Path)
				}
			}
		default:
			v, err := mgr.Open(nil, sub)
			if err != nil {
				fmt.Fprintf(os.Stderr, "vault open failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Vault abierto: %s (%s)\n", v.Name, v.Path)
		}
		os.Exit(0)
	case "agent", "agents", "memory", "config", "tasks", "sessions", "workflows", "models", "projects":
		fmt.Printf("huginn %s — aún no implementado como subcomando CLI.\n", args.Subcommand)
		fmt.Println("Usa la TUI: huginn  →  escribe /" + args.Subcommand + " o ? para ayuda.")
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "huginn: subcomando desconocido: %s\n", args.Subcommand)
		os.Exit(2)
	}
}

func runTUI(projectPath, vaultPath, prompt string) error {
	// Valida que el path exista si no es "."
	if projectPath != "." {
		if _, err := os.Stat(projectPath); err != nil {
			return fmt.Errorf("ruta no encontrada: %s", projectPath)
		}
		if !project.IsDirectory(projectPath) {
			return fmt.Errorf("la ruta no es un directorio: %s", projectPath)
		}
	}

	m := New(projectPath, vaultPath)
	if strings.TrimSpace(prompt) != "" {
		m.input = prompt
		m.cursor = len([]rune(prompt))
	}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

// Run expone runTUI para entradas finas.
func Run(projectPath, prompt string) error {
	vaultPath, _ := vault.ResolveVaultPath()
	abs, _ := filepath.Abs(projectPath)
	return runTUI(abs, vaultPath, prompt)
}
