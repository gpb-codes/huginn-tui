package memory

import (
	"os"
	"path/filepath"
)

// Layers — memoria 4 capas Mimo: project MEMORY.md, checkpoint.md, scratch notes.md, history FTS5 (via executions.jsonl)
type Layers struct{ base string }

func NewLayers(base string) *Layers { return &Layers{base: base} }

func (l *Layers) Ensure() error {
	for _, p := range []string{
		filepath.Join(l.base, "MEMORY.md"),
		filepath.Join(l.base, "checkpoint.md"),
		filepath.Join(l.base, "notes.md"),
		filepath.Join(l.base, "tasks"),
	} {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			if filepath.Ext(p) == "" {
				_ = os.MkdirAll(p, 0755)
			} else {
				_ = os.MkdirAll(filepath.Dir(p), 0755)
				_ = os.WriteFile(p, []byte("# "+filepath.Base(p)+"\n"), 0644)
			}
		}
	}
	return nil
}
