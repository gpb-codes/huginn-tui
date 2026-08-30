package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"huginn/internal/domain/project"
	"huginn/internal/domain/vault"
)

// Args parsed from os.Args.
type Args struct {
	Path       string
	Prompt     string
	Subcommand string
	SubArgs    []string
	Help       bool
	Version    bool
	Debug      bool
}

var FutureSubcommands = map[string]bool{
	"chat": true, "agent": true, "agents": true, "memory": true, "vault": true, "config": true, "graph": true,
}

func IsDebug(args []string) bool {
	if os.Getenv("HUGINN_DEBUG") == "1" || os.Getenv("HUGINN_DEBUG") == "true" {
		return true
	}
	for _, a := range args {
		if a == "--debug" {
			return true
		}
	}
	return false
}

func ParseArgs(raw []string) (Args, error) {
	var out Args
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
	if len(positional) > 0 && FutureSubcommands[positional[0]] {
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
		if p == "." || p == "./" || project.IsDirectory(p) {
			out.Path = p
		} else {
			isAbs := filepath.IsAbs(p) || (len(p) >= 2 && p[1] == ':')
			if isAbs || (strings.Contains(p, string(os.PathSeparator)) && project.IsDirectory(p)) {
				out.Path = p
			} else {
				out.Prompt = p
				out.Path = "."
			}
		}
		return out, nil
	}
	first := positional[0]
	if first == "." || first == "./" || project.IsDirectory(first) || filepath.IsAbs(first) || (len(first) >= 2 && first[1] == ':') {
		if project.IsDirectory(first) || filepath.IsAbs(first) || strings.Contains(first, string(os.PathSeparator)) {
			out.Path = first
			out.Prompt = strings.Join(positional[1:], " ")
			return out, nil
		}
	}
	out.Prompt = strings.Join(positional, " ")
	out.Path = "."
	return out, nil
}

func PrintHelp() {
	useColor := os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
	accent, reset := "", ""
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
  huginn graph                Ver knowledge graph
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

func PrintVersion(version string) {
	fmt.Printf("huginn %s\n", version)
}

func HuginnError(msg string, debug bool, err error) {
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

// ResolveContext is the application use-case: path → project context.
// Delegates to domain/project and domain/vault, no duplication.
func ResolveContext(projectPath string) (abs string, pkgManager string, vaultPath string, vaultOK bool, err error) {
	abs, err = filepath.Abs(projectPath)
	if err != nil {
		abs = projectPath
	}
	if projectPath != "." {
		if _, e := os.Stat(abs); e != nil {
			return "", "", "", false, fmt.Errorf("ruta no encontrada: %s", abs)
		}
		if !project.IsDirectory(abs) {
			return "", "", "", false, fmt.Errorf("la ruta no es un directorio: %s", abs)
		}
	}
	pkgManager = project.DetectPackageManager(abs)
	vaultPath, vaultOK = vault.ResolveVaultPath()
	return abs, pkgManager, vaultPath, vaultOK, nil
}
