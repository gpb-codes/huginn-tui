// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefault_Complete(t *testing.T) {
	c := Default()
	if c.Version != CurrentVersion {
		t.Fatal("bad version")
	}
	if c.Retry.MaxAttempts != 2 || c.Timeouts.TaskSeconds != 90 {
		t.Fatal("bad retry/timeout defaults")
	}
	if c.Permissions.AutoApprove {
		t.Fatal("auto-approve must default off")
	}
	if !c.Logging.RedactSecrets {
		t.Fatal("redaction must default on")
	}
}

func TestLoad_MigratesV1(t *testing.T) {
	dir := t.TempDir()
	v1 := map[string]any{"version": 1, "theme": "custom", "memory": map[string]any{"enabled": true}}
	b, _ := json.Marshal(v1)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), b, 0644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.Version != CurrentVersion {
		t.Fatalf("expected migration to %d, got %d", CurrentVersion, c.Version)
	}
	if c.Theme != "custom" {
		t.Fatal("user value must survive migration")
	}
	if c.Retry.MaxAttempts != 2 || c.Timeouts.TaskSeconds != 90 {
		t.Fatal("migration must fill new defaults")
	}
}

func TestSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := Default()
	c.Models.OllamaModel = "qwen2.5"
	if err := Save(dir, c); err != nil {
		t.Fatal(err)
	}
	back, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if back.Models.OllamaModel != "qwen2.5" {
		t.Fatal("round trip failed")
	}
}

func TestDefault_ModelMatrix(t *testing.T) {
	m := Default().Models.OllamaModels
	for cap, want := range map[string]string{
		"coding": "qwen3", "testing": "qwen3", "devops": "qwen3",
		"research": "llama3.1", "reasoning": "llama3.1",
		"documentation": "llama3.1", "review": "llama3.1",
		ChatModelKey: "llama3.1",
	} {
		if m[cap] != want {
			t.Fatalf("matrix[%q] = %q, want %q", cap, m[cap], want)
		}
	}
}

func TestLoad_MigratesV2FillsMatrix(t *testing.T) {
	dir := t.TempDir()
	// Un fichero v2 sin ollama_models recibe defectos sin perder personalizados.
	v2 := map[string]any{
		"version": 2,
		"models":  map[string]any{"ollama_model": "custom", "ollama_models": map[string]any{"review": "my-review-model"}},
	}
	b, _ := json.Marshal(v2)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), b, 0644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.Version != CurrentVersion {
		t.Fatalf("expected migration to %d, got %d", CurrentVersion, c.Version)
	}
	if c.Models.OllamaModels["review"] != "my-review-model" {
		t.Fatal("custom matrix entry lost in migration")
	}
	if c.Models.OllamaModels["coding"] != "qwen3" {
		t.Fatal("migration must fill missing matrix entries with defaults")
	}
}

func TestSave_MatrixRoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := Default()
	c.Models.OllamaModels["review"] = "my-review-model"
	if err := Save(dir, c); err != nil {
		t.Fatal(err)
	}
	back, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if back.Models.OllamaModels["review"] != "my-review-model" {
		t.Fatal("matrix round trip failed")
	}
}
