package onboarding

import (
	"os"
	"path/filepath"
	"runtime"

	domain "huginn/internal/domain/onboarding"
	"huginn/internal/domain/project"
)

// Propose — analiza respuestas + contexto y propone configuracion (no ejecuta)
func Propose(partial domain.Result, projectPath string) domain.Result {
	proposal := partial
	// distinguir: Preferencia indicada vs Propuesta de HUGINN
	// si faltan datos, inferir con defaults razonables

	// OS
	if proposal.Technical.OS == "" {
		proposal.Technical.OS = runtime.GOOS
	}
	// Editor — detecta
	if proposal.Technical.Editor == "" {
		if _, err := os.Stat(filepath.Join(projectPath, ".vscode")); err == nil {
			proposal.Technical.Editor = "VS Code"
		} else {
			proposal.Technical.Editor = "VS Code"
		}
	}
	// Lenguajes — detecta por archivos
	if len(proposal.Technical.Languages) == 0 {
		if ok, _ := project.DetectProject(projectPath); ok {
			pm := project.DetectPackageManager(projectPath)
			switch pm {
			case "bun", "pnpm", "npm", "yarn":
				proposal.Technical.Languages = []string{"TypeScript", "JavaScript"}
				proposal.Technical.PrimaryLang = "TypeScript"
			default:
				if _, err := os.Stat(filepath.Join(projectPath, "go.mod")); err == nil {
					proposal.Technical.Languages = []string{"Go"}
					proposal.Technical.PrimaryLang = "Go"
				} else if _, err := os.Stat(filepath.Join(projectPath, "pyproject.toml")); err == nil {
					proposal.Technical.Languages = []string{"Python"}
					proposal.Technical.PrimaryLang = "Python"
				} else {
					proposal.Technical.Languages = []string{"Go"}
					proposal.Technical.PrimaryLang = "Go"
				}
			}
		}
	}
	// Stack
	if proposal.Technical.Stack == "" && len(proposal.Technical.Frameworks) == 0 {
		if proposal.Technical.PrimaryLang == "TypeScript" {
			proposal.Technical.Frameworks = []string{"Next.js", "React"}
			proposal.Technical.Stack = "Next.js + React"
		} else if proposal.Technical.PrimaryLang == "Go" {
			proposal.Technical.Stack = "Go + Bubble Tea"
		}
	}
	// Desarrollo defaults
	if proposal.Development.Architecture == "" {
		proposal.Development.Architecture = "Modular y mantenible (Clean/Hexagonal)"
	}
	if proposal.Development.Methodology == "" {
		proposal.Development.Methodology = "Feature branches + PR"
	}
	if proposal.Development.TestingLevel == "" {
		proposal.Development.TestingLevel = "Unit + Integration"
	}
	// Git defaults
	if proposal.Git.Workflow == "" {
		proposal.Git.Workflow = "feature branches"
		proposal.Git.BranchStrategy = "feature/* → develop → main"
		proposal.Git.CommitConvention = "Conventional Commits"
	}
	// AI defaults
	if len(proposal.AI.Providers) == 0 {
		proposal.AI.Providers = []string{"chatgpt", "opencode"}
		proposal.AI.LocalOrCloud = "cloud con fallback local"
		proposal.AI.AutonomyLevel = "conservador — pide confirmacion antes de acciones destructivas"
	}
	// General defaults
	if proposal.General.AnswerStyle == "" {
		proposal.General.AnswerStyle = "conciso con detalle cuando se requiera"
		proposal.General.ExplainDecisions = true
		proposal.General.ProposeAlternatives = true
		proposal.General.AgentStyle = "conservador"
	}
	return proposal
}


