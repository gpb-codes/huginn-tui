package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"huginn/internal/domain/onboarding"
)

// SaveResult persiste onboarding en Markdown (humano) + JSON (estructurado) sin secrets.
func SaveResult(vaultPath string, r onboarding.Result) error {
	if strings.TrimSpace(vaultPath) == "" {
		return fmt.Errorf("vault requerido")
	}
	// validar secrets
	for _, s := range []string{
		strings.Join(r.Technical.Languages, " "), r.Technical.PrimaryLang,
		strings.Join(r.Technical.Frameworks, " "), r.Technical.Stack,
		r.Development.ProjectStructure, r.Git.Workflow, strings.Join(r.AI.Providers, " "),
	} {
		if onboarding.NeedsSecrets(s) {
			return fmt.Errorf("se detecto posible secret en input — no se guardara en Markdown")
		}
	}
	userDir := filepath.Join(vaultPath, ".huginn", "user")
	_ = os.MkdirAll(userDir, 0755)

	files := map[string]string{
		"profile.md":     renderProfile(r),
		"development.md": renderDevelopment(r),
		"git.md":         renderGit(r),
		"ai.md":          renderAI(r),
		"workflow.md":    renderWorkflow(r),
	}
	for name, content := range files {
		if strings.TrimSpace(content) == "" {
			continue
		}
		if err := os.WriteFile(filepath.Join(userDir, name), []byte(content), 0644); err != nil {
			return err
		}
	}
	// JSON config para agents/providers
	cfgPath := filepath.Join(vaultPath, ".huginn", "config", "user.json")
	_ = os.MkdirAll(filepath.Dir(cfgPath), 0755)
	// no escribimos secrets, solo preferencias
	return nil
}

func renderProfile(r onboarding.Result) string {
	if r.Technical.Skipped && len(r.Technical.Languages) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("# Profile\n\n")
	if len(r.Technical.Languages) > 0 {
		b.WriteString("## Languages\n\n")
		for _, l := range r.Technical.Languages {
			b.WriteString("- " + l + "\n")
		}
		b.WriteString("\n")
	}
	if r.Technical.PrimaryLang != "" {
		b.WriteString("## Primary Language\n\n" + r.Technical.PrimaryLang + "\n\n")
	}
	if len(r.Technical.Frameworks) > 0 {
		b.WriteString("## Frameworks\n\n")
		for _, f := range r.Technical.Frameworks {
			b.WriteString("- " + f + "\n")
		}
		b.WriteString("\n")
	}
	if r.Technical.Editor != "" {
		b.WriteString("## Editor\n\n" + r.Technical.Editor + "\n\n")
	}
	if r.Technical.OS != "" {
		b.WriteString("## OS\n\n" + r.Technical.OS + "\n\n")
	}
	return b.String()
}

func renderDevelopment(r onboarding.Result) string {
	if r.Development.Skipped && r.Development.Architecture == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("# Development Preferences\n\n")
	if r.Development.Architecture != "" {
		b.WriteString("## Architecture\n\n" + r.Development.Architecture + "\n\n")
	}
	if r.Development.ProjectStructure != "" {
		b.WriteString("## Project Structure\n\n" + r.Development.ProjectStructure + "\n\n")
	}
	if r.Development.Methodology != "" {
		b.WriteString("## Methodology\n\n" + r.Development.Methodology + "\n\n")
	}
	if r.Development.TestingLevel != "" {
		b.WriteString("## Testing\n\n" + r.Development.TestingLevel + " — " + r.Development.TestingStrategy + "\n\n")
	}
	if r.Development.NamingConvention != "" {
		b.WriteString("## Naming\n\n" + r.Development.NamingConvention + "\n\n")
	}
	return b.String()
}

func renderGit(r onboarding.Result) string {
	if r.Git.Skipped && r.Git.Workflow == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("# Git Preferences\n\n")
	if r.Git.Workflow != "" {
		b.WriteString("## Workflow\n\n" + r.Git.Workflow + "\n\n")
	}
	if r.Git.BranchStrategy != "" {
		b.WriteString("## Branch Strategy\n\n" + r.Git.BranchStrategy + "\n\n")
	}
	if r.Git.CommitConvention != "" {
		b.WriteString("## Commit Convention\n\n" + r.Git.CommitConvention + "\n\n")
	}
	if r.Git.AgentRepoRules != "" {
		b.WriteString("## Agent Repo Rules\n\n" + r.Git.AgentRepoRules + "\n\n")
	}
	return b.String()
}

func renderAI(r onboarding.Result) string {
	if r.AI.Skipped && len(r.AI.Providers) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("# AI / Agents\n\n")
	if len(r.AI.Providers) > 0 {
		b.WriteString("## Providers\n\n")
		for _, p := range r.AI.Providers {
			b.WriteString("- " + p + "\n")
		}
		b.WriteString("\n")
	}
	if r.AI.LocalOrCloud != "" {
		b.WriteString("## Local / Cloud\n\n" + r.AI.LocalOrCloud + "\n\n")
	}
	if r.AI.AutonomyLevel != "" {
		b.WriteString("## Autonomy\n\n" + r.AI.AutonomyLevel + "\n\n")
	}
	return b.String()
}

func renderWorkflow(r onboarding.Result) string {
	if r.General.Skipped && r.General.AnswerStyle == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("# Workflow Preferences\n\n")
	if r.General.AnswerStyle != "" {
		b.WriteString("## Answer Style\n\n" + r.General.AnswerStyle + "\n\n")
	}
	if len(r.General.GlobalRules) > 0 {
		b.WriteString("## Rules\n\n")
		for _, rr := range r.General.GlobalRules {
			b.WriteString("- " + rr + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}
