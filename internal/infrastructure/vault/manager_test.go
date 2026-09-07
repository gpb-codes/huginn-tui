// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package vault

import (
	"testing"
)

func TestFilesystemManagerCreateAndDetect(t *testing.T) {
	dir := t.TempDir()
	mgr := NewFilesystemManager()
	v, err := mgr.Create(nil, dir, "test-vault")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if v.Name != "test-vault" {
		t.Fatalf("Name %q", v.Name)
	}
	if !mgr.IsInitialized(v.Path) {
		t.Fatalf("IsInitialized false")
	}
	detected, ok := mgr.Detect(v.Path)
	if !ok || detected.ID != v.ID {
		t.Fatalf("Detect failed %v %v", detected, ok)
	}
	recent, err := mgr.Recent()
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(recent) == 0 {
		t.Fatalf("Recent empty")
	}
}

func TestFilesystemManagerExists(t *testing.T) {
	mgr := NewFilesystemManager()
	if mgr.Exists("/no/such/path/12345") {
		t.Fatalf("Exists should be false")
	}
	dir := t.TempDir()
	if !mgr.Exists(dir) {
		t.Fatalf("Exists should be true for TempDir")
	}
}
