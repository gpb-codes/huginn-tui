package git

import (
	"context"
	"os/exec"
	"strings"
)

// Adapter implements ports.GitPort via os/exec git.
type Adapter struct {
	Dir string
}

func New(dir string) *Adapter { return &Adapter{Dir: dir} }

func (a *Adapter) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = a.Dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (a *Adapter) Status(ctx context.Context) (string, error) {
	return a.run(ctx, "status", "--porcelain")
}
func (a *Adapter) Diff(ctx context.Context) (string, error) { return a.run(ctx, "diff") }
func (a *Adapter) Branch(ctx context.Context) (string, error) {
	return a.run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
}
func (a *Adapter) Commit(ctx context.Context, message string) (string, error) {
	return a.run(ctx, "commit", "-m", message)
}
