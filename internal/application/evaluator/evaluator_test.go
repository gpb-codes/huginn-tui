// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package evaluator

import (
	"testing"
	"time"

	"huginn/internal/domain/task"
)

func TestEvaluate_Success(t *testing.T) {
	e := New()
	v := e.Evaluate(task.New("1", "T", "", ""), *task.SuccessResult("done"))
	if !v.Success {
		t.Fatalf("expected success: %+v", v)
	}
}

func TestEvaluate_FailureRetryable(t *testing.T) {
	e := New()
	v := e.Evaluate(task.New("1", "T", "", ""), *task.FailedResult("timeout after 90s"))
	if v.Success || !v.Retryable {
		t.Fatalf("expected retryable failure: %+v", v)
	}
}

func TestEvaluate_EmptyOutputFails(t *testing.T) {
	e := New()
	v := e.Evaluate(task.New("1", "T", "", ""), *task.SuccessResult(""))
	if v.Success {
		t.Fatal("empty output must fail")
	}
}

func TestRetryPolicy_Bounds(t *testing.T) {
	p := DefaultRetryPolicy()
	tk := task.New("1", "T", "", "")
	v := Verdict{Success: false, Retryable: true}
	if !p.ShouldRetry(tk, 1, v) {
		t.Fatal("attempt 1 should retry")
	}
	if p.ShouldRetry(tk, 2, v) {
		t.Fatal("attempt 2 must stop with default max 2")
	}
	if p.ShouldRetry(tk, 1, Verdict{Success: false, Retryable: false}) {
		t.Fatal("non-retryable must stop")
	}
	tk.MaxAttempts = 1
	if p.ShouldRetry(tk, 1, v) {
		t.Fatal("task override must win")
	}
	if p.Backoff(1) != 2*time.Second {
		t.Fatal("bad backoff")
	}
}

func TestApprovalGate(t *testing.T) {
	if ApprovalGate(task.New("1", "Write docs", "", "")) != nil {
		t.Fatal("benign task must not gate")
	}
	if ApprovalGate(task.New("1", "git push", "", "")) == nil {
		t.Fatal("push must gate")
	}
}
