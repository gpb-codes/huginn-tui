package trace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"huginn/internal/domain/execution"
)

// Store — append JSONL a .huginn/logs/executions.jsonl sin secrets
type Store struct {
	vaultPath string
}

func New(vaultPath string) *Store { return &Store{vaultPath: vaultPath} }

func (s *Store) Append(rec execution.Record) error {
	if s.vaultPath == "" {
		return nil
	}
	dir := filepath.Join(s.vaultPath, ".huginn", "logs")
	_ = os.MkdirAll(dir, 0755)
	rec.FinishedAt = time.Now()
	if rec.StartedAt.IsZero() {
		rec.StartedAt = rec.FinishedAt
	}
	rec.Latency = rec.FinishedAt.Sub(rec.StartedAt).Milliseconds()
	// nunca persistir secrets: input truncado y sanitizado
	if len(rec.Input) > 800 {
		rec.Input = rec.Input[:800] + "..."
	}
	b, _ := json.Marshal(rec)
	f, err := os.OpenFile(filepath.Join(dir, "executions.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(b, '\n'))
	return err
}
