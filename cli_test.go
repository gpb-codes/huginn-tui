package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgs_Help(t *testing.T) {
	for _, a := range [][]string{{"--help"}, {"-h"}, {"help"}, {"--help", "extra"}} {
		parsed, err := parseArgs(a)
		if err != nil || !parsed.Help {
			t.Fatalf("expected help for %v, got %v err %v", a, parsed.Help, err)
		}
	}
}

func TestParseArgs_Version(t *testing.T) {
	for _, a := range [][]string{{"--version"}, {"-v"}, {"version"}} {
		parsed, err := parseArgs(a)
		if err != nil || !parsed.Version {
			t.Fatalf("expected version for %v", a)
		}
	}
}

func TestParseArgs_NoArgs(t *testing.T) {
	parsed, _ := parseArgs([]string{})
	if parsed.Path != "." || parsed.Prompt != "" {
		t.Fatalf("no args should be Path='.' got %+v", parsed)
	}
}

func TestParseArgs_Path(t *testing.T) {
	tmp := t.TempDir()
	parsed, _ := parseArgs([]string{tmp})
	if parsed.Path != tmp {
		t.Fatalf("expected path %s got %s", tmp, parsed.Path)
	}
	if parsed.Prompt != "" {
		t.Fatalf("expected empty prompt")
	}
}

func TestParseArgs_Prompt(t *testing.T) {
	parsed, _ := parseArgs([]string{"analiza este proyecto"})
	if parsed.Prompt != "analiza este proyecto" {
		t.Fatalf("prompt mismatch got %q", parsed.Prompt)
	}
	if parsed.Path != "." {
		t.Fatalf("path should be '.'")
	}
}

func TestParseArgs_PathPrompt(t *testing.T) {
	tmp := t.TempDir()
	parsed, _ := parseArgs([]string{tmp, "crea", "auth"})
	if parsed.Path != tmp {
		t.Fatalf("path mismatch")
	}
	if parsed.Prompt != "crea auth" {
		t.Fatalf("prompt mismatch got %q", parsed.Prompt)
	}
}

func TestParseArgs_FutureSubcommand(t *testing.T) {
	parsed, _ := parseArgs([]string{"memory", "extra"})
	if parsed.Subcommand != "memory" {
		t.Fatalf("expected memory subcommand")
	}
	parsed2, _ := parseArgs([]string{"vault"})
	if parsed2.Subcommand != "vault" {
		t.Fatalf("expected vault")
	}
}

func TestParseArgs_UnknownFlag(t *testing.T) {
	_, err := parseArgs([]string{"--unknown"})
	if err == nil {
		t.Fatalf("expected error for unknown flag")
	}
}

func TestIsDirectory(t *testing.T) {
	tmp := t.TempDir()
	if !isDirectory(tmp) {
		t.Fatalf("tmp should be directory")
	}
	if isDirectory(filepath.Join(tmp, "nope")) {
		t.Fatalf("nonexistent should be false")
	}
}

func TestDetectPackageManager(t *testing.T) {
	tmp := t.TempDir()
	// npm
	os.WriteFile(filepath.Join(tmp, "package.json"), []byte("{}"), 0644)
	if pm := detectPackageManager(tmp); pm != "npm" {
		t.Fatalf("expected npm got %q", pm)
	}
	os.WriteFile(filepath.Join(tmp, "pnpm-lock.yaml"), []byte(""), 0644)
	if pm := detectPackageManager(tmp); pm != "pnpm" {
		t.Fatalf("expected pnpm got %q", pm)
	}
	os.Remove(filepath.Join(tmp, "pnpm-lock.yaml"))
	os.WriteFile(filepath.Join(tmp, "yarn.lock"), []byte(""), 0644)
	if pm := detectPackageManager(tmp); pm != "yarn" {
		t.Fatalf("expected yarn got %q", pm)
	}
	os.Remove(filepath.Join(tmp, "yarn.lock"))
	os.WriteFile(filepath.Join(tmp, "bun.lockb"), []byte(""), 0644)
	if pm := detectPackageManager(tmp); pm != "bun" {
		t.Fatalf("expected bun got %q", pm)
	}
}

func TestDetectProject(t *testing.T) {
	tmp := t.TempDir()
	ok, _ := detectProject(tmp)
	if ok {
		t.Fatalf("empty dir should not be project")
	}
	os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module x"), 0644)
	ok, marker := detectProject(tmp)
	if !ok || marker != "go.mod" {
		t.Fatalf("should detect go.mod")
	}
}

func TestResolveVaultPath_Env(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HUGINN_VAULT", tmp)
	path, ok := resolveVaultPath()
	if path != tmp || !ok {
		t.Fatalf("env vault failed got %q %v", path, ok)
	}
}

func TestInitialModelWithContext(t *testing.T) {
	tmp := t.TempDir()
	m := initialModelWithContext(tmp, "hola mundo")
	if m.projectPath == "" {
		t.Fatalf("projectPath empty")
	}
	// prompt injected
	found := false
	for _, h := range m.chatHistory {
		if h.Text == "hola mundo" && h.IsUser {
			found = true
		}
	}
	if !found {
		t.Fatalf("prompt not injected into chatHistory")
	}
	if m.mode != modeRunning {
		t.Fatalf("expected modeRunning when prompt provided")
	}
}
