// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package execution

import (
	"testing"

	"huginn/internal/domain/task"
)

func TestRequiresApproval_Keywords(t *testing.T) {
	cases := []struct {
		title string
		want  bool
	}{
		{"Implement login form", false},
		{"git push to production", true},
		{"Run rm -rf on cache", true},
		{"Apply database migration", true},
		{"Write docs", false},
	}
	for _, c := range cases {
		got := RequiresApproval(task.New("1", c.title, c.title, ""))
		if got != c.want {
			t.Errorf("%q: got %v want %v", c.title, got, c.want)
		}
	}
}

func TestRequiresApproval_Flag(t *testing.T) {
	x := task.New("1", "harmless", "harmless", "")
	x.NeedsApproval = true
	if !RequiresApproval(x) {
		t.Fatal("flag must force approval")
	}
}

func TestNewRun_Planning(t *testing.T) {
	r := NewRun("obj", "plan-1", []string{"a"})
	if r.Status != task.StatusPlanning {
		t.Fatalf("expected planning, got %s", r.Status)
	}
	if r.ID == "" || len(r.TaskIDs) != 1 {
		t.Fatal("bad run init")
	}
}
