package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"huginn/internal/domain/memory"
)

func TestMarkdownStore_SaveAndGet(t *testing.T) {
	base := t.TempDir()
	store := NewMarkdownStore(base)
	m := memory.Memory{
		ID:         "test_001",
		Type:       memory.TypePreference,
		Title:      "Test Preference",
		Content:    "prefiero pnpm",
		Importance: 0.8,
		Confidence: 0.9,
		Tags:       []string{"test"},
	}
	if err := store.Save(context.Background(), m); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	got, err := store.Get(context.Background(), "test_001")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Title != "Test Preference" || got.Type != memory.TypePreference {
		t.Fatalf("mismatch %+v", got)
	}
	// check file exists
	if _, err := os.Stat(filepath.Join(base, "memory", "test_001.md")); err != nil {
		t.Fatalf("file not created")
	}
}

func TestMarkdownStore_Search(t *testing.T) {
	base := t.TempDir()
	store := NewMarkdownStore(base)
	m1 := memory.Memory{ID: "a1", Type: memory.TypeFact, Title: "Go is great", Content: "Go for CLI", Importance: 0.9, Tags: []string{"go"}}
	m2 := memory.Memory{ID: "a2", Type: memory.TypeFact, Title: "Python is great", Content: "Python for scripts", Importance: 0.5}
	store.Save(context.Background(), m1)
	store.Save(context.Background(), m2)
	res, _ := store.Search(context.Background(), "go")
	if len(res) == 0 {
		t.Fatal("expected at least 1 result for go")
	}
	if res[0].ID != "a1" {
		t.Fatalf("expected a1 first, got %s", res[0].ID)
	}
}

func TestMarkdownStore_SecretRejection(t *testing.T) {
	base := t.TempDir()
	store := NewMarkdownStore(base)
	m := memory.Memory{
		ID:      "sec_001",
		Type:    memory.TypeFact,
		Title:   "Secret",
		Content: "api_key= sk-1234567890abcdef1234567890",
	}
	err := store.Save(context.Background(), m)
	if err == nil {
		t.Fatal("should reject secret")
	}
}

func TestMarkdownStore_IndexRebuild(t *testing.T) {
	base := t.TempDir()
	store := NewMarkdownStore(base)
	m := memory.Memory{ID: "idx1", Type: memory.TypeLesson, Title: "Lesson", Content: "test", Importance: 0.5}
	store.Save(context.Background(), m)
	if err := store.RebuildIndex(); err != nil {
		t.Fatalf("rebuild failed: %v", err)
	}
	idx, _ := store.ReadIndex()
	if len(idx) == 0 {
		t.Fatal("index empty after rebuild")
	}
}
