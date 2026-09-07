// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package agents

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"huginn/internal/domain/agent"
)

func fakeExec(output string, err error) func(context.Context, string, []string, string) ([]byte, error) {
	return func(_ context.Context, _ string, _ []string, _ string) ([]byte, error) {
		return []byte(output), err
	}
}

func TestExecute_SuccessParsesJSON(t *testing.T) {
	a := NewOpenCodeAdapterWithConfig(Config{Bin: "opencode"})
	a.exec = fakeExec("{\"type\":\"text\",\"part\":{\"text\":\"hola\"}}\n{\"type\":\"text\",\"part\":{\"text\":\" mundo\"}}\n", nil)
	res, err := a.Execute(context.Background(), agent.AgentTask{ID: "1", Input: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	out, ok := res.Output.(agent.CodeResult)
	if !ok || out.Summary != "hola mundo" {
		t.Fatalf("bad parse: %+v", res.Output)
	}
	if res.Status != "ok" {
		t.Fatal("expected ok")
	}
}

func TestExecute_RawFallback(t *testing.T) {
	a := NewOpenCodeAdapterWithConfig(Config{Bin: "opencode"})
	a.exec = fakeExec("plain text output", nil)
	res, err := a.Execute(context.Background(), agent.AgentTask{ID: "1", Input: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if out := res.Output.(agent.CodeResult); out.Summary != "plain text output" {
		t.Fatalf("expected raw fallback: %q", out.Summary)
	}
}

func TestExecute_NotInstalled(t *testing.T) {
	a := NewOpenCodeAdapterWithConfig(Config{Bin: ""})
	a.bin = ""
	_, err := a.Execute(context.Background(), agent.AgentTask{ID: "1", Input: "hi"})
	if !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("expected ErrNotInstalled, got %v", err)
	}
}

func TestExecute_InvalidWorkspace(t *testing.T) {
	a := NewOpenCodeAdapterWithConfig(Config{Bin: "opencode"})
	_, err := a.Execute(context.Background(), agent.AgentTask{
		ID: "1", Input: "hi",
		Context: agent.AgentContext{ProjectPath: `Z:\no-existe-huginn-xyz`},
	})
	if !errors.Is(err, ErrInvalidWorkspace) {
		t.Fatalf("expected ErrInvalidWorkspace, got %v", err)
	}
}

func TestExecute_ValidWorkspacePassedAsDir(t *testing.T) {
	dir := t.TempDir()
	var gotDir string
	a := NewOpenCodeAdapterWithConfig(Config{Bin: "opencode"})
	a.exec = func(_ context.Context, _ string, _ []string, dir string) ([]byte, error) {
		gotDir = dir
		return []byte("ok"), nil
	}
	_, err := a.Execute(context.Background(), agent.AgentTask{
		ID: "1", Input: "hi", Context: agent.AgentContext{ProjectPath: dir},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotDir != dir {
		t.Fatalf("workspace not forwarded: %q", gotDir)
	}
}

func TestExecute_Timeout(t *testing.T) {
	a := NewOpenCodeAdapterWithConfig(Config{Bin: "opencode", Timeout: 50 * time.Millisecond})
	a.exec = func(ctx context.Context, _ string, _ []string, _ string) ([]byte, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			return []byte("late"), nil
		}
	}
	_, err := a.Execute(context.Background(), agent.AgentTask{ID: "1", Input: "hi"})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}
}

func TestExecute_ProcessCrash(t *testing.T) {
	a := NewOpenCodeAdapterWithConfig(Config{Bin: "opencode"})
	a.exec = fakeExec("partial output", errors.New("exit status 1"))
	res, err := a.Execute(context.Background(), agent.AgentTask{ID: "1", Input: "hi"})
	if !errors.Is(err, ErrProcessCrash) {
		t.Fatalf("expected ErrProcessCrash, got %v", err)
	}
	if out := res.Output.(agent.CodeResult); out.Summary != "partial output" {
		t.Fatal("crash must keep captured output for the evaluator")
	}
}

func TestExecute_ModelAndAgentFlags(t *testing.T) {
	var gotArgs []string
	a := NewOpenCodeAdapterWithConfig(Config{Bin: "opencode", Model: "m", Agent: "sub"})
	a.exec = func(_ context.Context, _ string, args []string, _ string) ([]byte, error) {
		gotArgs = args
		return []byte("ok"), nil
	}
	_, _ = a.Execute(context.Background(), agent.AgentTask{ID: "1", Input: "hi"})
	joined := ""
	for _, x := range gotArgs {
		joined += x + " "
	}
	for _, want := range []string{"--model m", "--agent sub", "--format json"} {
		found := false
		for i := 0; i+len(want) <= len(joined); i++ {
			if joined[i:i+len(want)] == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing flag %q in %q", want, joined)
		}
	}
}

// TestOpenCodeAdapter_Integration usa el binario real cuando está disponible.
// Ejecutar con: HUGINN_TEST_OPENCODE=1 go test ./internal/infrastructure/agents/ -run Integration -v
func TestOpenCodeAdapter_Integration(t *testing.T) {
	if os.Getenv("HUGINN_TEST_OPENCODE") != "1" {
		t.Skip("set HUGINN_TEST_OPENCODE=1 to run against the real opencode binary")
	}
	a := NewOpenCodeAdapter()
	ok, ver := a.Detect()
	if !ok {
		t.Skip("opencode not installed")
	}
	t.Logf("opencode %s", ver)
	res, err := a.Execute(context.Background(), agent.AgentTask{
		ID: "it-1", Input: "reply with exactly: huginn-integration-ok",
		Context: agent.AgentContext{ProjectPath: t.TempDir()},
	})
	if err != nil {
		t.Fatalf("integration execute: %v", err)
	}
	t.Logf("output: %+v", res.Output)
}
