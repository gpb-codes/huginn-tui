package main

import (
	"fmt"
	"os"

	"huginn/internal/cli"
)

// VERSION is single source of truth — domain layer, not CLI.
const VERSION = "v0.2.0"

func main() {
	args := os.Args[1:]
	parsed, err := cli.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "huginn: %v\n", err)
		fmt.Fprintln(os.Stderr, "Usa: huginn --help")
		os.Exit(2)
	}
	if parsed.Help {
		cli.PrintHelp()
		os.Exit(0)
	}
	if parsed.Version {
		cli.PrintVersion(VERSION)
		os.Exit(0)
	}
	if parsed.Subcommand != "" {
		switch parsed.Subcommand {
		case "chat":
			path := "."
			prompt := ""
			if len(parsed.SubArgs) > 0 {
				// delegate to same parsing as root CLI
				subParsed, _ := cli.ParseArgs(parsed.SubArgs)
				path = subParsed.Path
				if path == "" {
					path = "."
				}
				prompt = subParsed.Prompt
				if prompt == "" && len(parsed.SubArgs) == 1 && subParsed.Path == "." {
					// was prompt, not path
					prompt = parsed.SubArgs[0]
				}
			}
			abs, _, _, _, err := cli.ResolveContext(path)
			if err != nil {
				cli.HuginnError(err.Error(), parsed.Debug || cli.IsDebug(args), err)
				os.Exit(1)
			}
			fmt.Printf("Huginn chat — %s\n", abs)
			if prompt != "" {
				fmt.Printf("Prompt: %s\n", prompt)
			}
			fmt.Println("(TUI se abriría aquí — usa `huginn` en la raíz para la UI completa)")
			os.Exit(0)
		case "graph":
			path := "."
			if len(parsed.SubArgs) > 0 {
				subParsed, _ := cli.ParseArgs(parsed.SubArgs)
				if subParsed.Path != "" {
					path = subParsed.Path
				} else if len(parsed.SubArgs) == 1 {
					path = parsed.SubArgs[0]
				}
			}
			abs, _, _, _, err := cli.ResolveContext(path)
			if err != nil {
				cli.HuginnError(err.Error(), parsed.Debug || cli.IsDebug(args), err)
				os.Exit(1)
			}
			fmt.Printf("Huginn graph — %s\n", abs)
			fmt.Println("Knowledge graph: 12 nodes • 23 edges • Agent Vault")
			fmt.Println("(En TUI: escribe `graph` o `/graph` o Ctrl+G)")
			os.Exit(0)
		case "agent", "agents", "memory", "vault", "config":
			fmt.Printf("huginn %s — (subcomando preparado, aún no implementado)\n", parsed.Subcommand)
			fmt.Println("Usa: huginn --help para ver comandos disponibles.")
			fmt.Println("El CLI está preparado para agregar este subcomando sin reescribir la arquitectura.")
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "huginn: subcomando desconocido: %s\n", parsed.Subcommand)
			os.Exit(2)
		}
	}
	projectPath := parsed.Path
	if projectPath == "" {
		projectPath = "."
	}
	abs, pkg, vaultPath, vaultOK, err := cli.ResolveContext(projectPath)
	if err != nil {
		cli.HuginnError(err.Error(), parsed.Debug || cli.IsDebug(args), err)
		os.Exit(1)
	}
	// Por ahora el TUI vive en la raíz (main.go legacy). Este binario demuestra la capa CLI limpia.
	// En la arquitectura final, aquí se llamaría a internal/tui.Run(ctx)
	fmt.Printf("\n  Huginn %s — Agent Vault Intelligence\n", VERSION)
	fmt.Printf("  Project  └─ %s\n", abs)
	if pkg != "" {
		fmt.Printf("  Package manager: %s\n", pkg)
	}
	if vaultPath != "" {
		status := "not found"
		if vaultOK {
			status = "connected"
		}
		fmt.Printf("  Vault    └─ %s (%s)\n", vaultPath, status)
	}
	if parsed.Prompt != "" {
		fmt.Printf("\n  Prompt: %s\n", parsed.Prompt)
		fmt.Println("  Orchestrator → Planner → Coder → Researcher (via Agent Vault)")
	}
	fmt.Println("\n  (TUI completa disponible ejecutando `huginn` desde la raíz — este binario es la capa CLI limpia)")
	fmt.Println("  Usa: huginn --help")
}
