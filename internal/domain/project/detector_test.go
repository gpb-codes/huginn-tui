package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsDirectory(t *testing.T) {
	tmp := t.TempDir()
	if !IsDirectory(tmp) {
		t.Fatal("tmp should be dir")
	}
	if IsDirectory(filepath.Join(tmp, "nope")) {
		t.Fatal("nonexistent should be false")
	}
	if IsDirectory("") {
		t.Fatal("empty should be false")
	}
}

func TestDetectPackageManager(t *testing.T) {
	tmp := t.TempDir()
	if pm := DetectPackageManager(tmp); pm != "" {
		t.Fatalf("empty should be empty, got %q", pm)
	}
	os.WriteFile(filepath.Join(tmp, "package.json"), []byte("{}"), 0644)
	if pm := DetectPackageManager(tmp); pm != "npm" {
		t.Fatalf("expected npm got %q", pm)
	}
	os.WriteFile(filepath.Join(tmp, "pnpm-lock.yaml"), []byte(""), 0644)
	if pm := DetectPackageManager(tmp); pm != "pnpm" {
		t.Fatalf("expected pnpm")
	}
	os.Remove(filepath.Join(tmp, "pnpm-lock.yaml"))
	os.WriteFile(filepath.Join(tmp, "yarn.lock"), []byte(""), 0644)
	if pm := DetectPackageManager(tmp); pm != "yarn" {
		t.Fatalf("expected yarn")
	}
	os.Remove(filepath.Join(tmp, "yarn.lock"))
	os.WriteFile(filepath.Join(tmp, "bun.lockb"), []byte(""), 0644)
	if pm := DetectPackageManager(tmp); pm != "bun" {
		t.Fatalf("expected bun")
	}
}

func TestDetectProject(t *testing.T) {
	tmp := t.TempDir()
	ok, _ := DetectProject(tmp)
	if ok {
		t.Fatal("empty should not be project")
	}
	os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module x"), 0644)
	ok, marker := DetectProject(tmp)
	if !ok || marker != "go.mod" {
		t.Fatalf("should detect go.mod")
	}
}
