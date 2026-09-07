// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package trace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"huginn/internal/domain/execution"
	"huginn/internal/infrastructure/security"
)

// Store añade registros JSONL a executions.jsonl sin persistir secretos.
type Store struct {
	vaultPath string
}

func New(vaultPath string) *Store { return &Store{vaultPath: vaultPath} }

func (s *Store) VaultPath() string { return s.vaultPath }

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
	// Trunca y sanea la entrada para no persistir secretos.
	if len(rec.Input) > 800 {
		rec.Input = rec.Input[:800] + "..."
	}
	rec.Input = security.Redact(rec.Input)
	for i, e := range rec.Errors {
		if len(e) > 500 {
			rec.Errors[i] = e[:500] + "..."
		}
		rec.Errors[i] = security.Redact(rec.Errors[i])
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
