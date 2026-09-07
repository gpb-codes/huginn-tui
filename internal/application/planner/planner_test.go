// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package planner

import (
	"testing"

	domainagent "huginn/internal/domain/agent"
)

func TestAnalyze_Kinds(t *testing.T) {
	p := New()
	cases := map[string]string{
		"investiga la documentación de X": "research",
		"fix the login bug":               "fix",
		"revisa este PR":                  "review",
		"documenta el módulo":             "document",
		"despliega a producción":          "devops",
		"implementa autenticación":        "implement",
	}
	for in, want := range cases {
		if got := p.Analyze(in).Kind; got != want {
			t.Errorf("%q: got %s want %s", in, got, want)
		}
	}
}

func TestPlan_ImplementDAG(t *testing.T) {
	p := New()
	pl, err := p.Plan("implementa autenticación con sistema distribuido escalable multi región optimizar refactorizar complejo", ".")
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := pl.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	levels := pl.Levels()
	// Un objetivo complejo debe bifurcarse: algún nivel tiene más de una tarea.
	fanned := false
	for _, l := range levels {
		if len(l) > 1 {
			fanned = true
		}
	}
	if !fanned {
		t.Fatal("expected parallel fan-out for complex objective")
	}
}

func TestPlan_QuestionCollapses(t *testing.T) {
	p := New()
	pl, err := p.Plan("qué es un vault", ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(pl.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(pl.Tasks))
	}
}

func TestPlan_FixAddsRepro(t *testing.T) {
	p := New()
	pl, err := p.Plan("fix: corrige el bug de login que falla", ".")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, task := range pl.Tasks {
		if task.ID == "task-0-repro" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected repro task for fix")
	}
}

func TestPlan_DevopsNeedsApproval(t *testing.T) {
	p := New()
	pl, err := p.Plan("despliega producción con migration de base de datos y push", ".")
	if err != nil {
		t.Fatal(err)
	}
	gated := false
	for _, task := range pl.Tasks {
		if task.Capability == domainagent.CapabilityDevOps && task.NeedsApproval {
			gated = true
		}
	}
	if !gated {
		t.Fatal("expected approval gate on devops task")
	}
}

func TestRecoveryPlan_Bounded(t *testing.T) {
	p := New()
	pl, _ := p.Plan("implementa X simple", ".")
	failed := []struct{}{}
	_ = failed
	rp := p.RecoveryPlan("implementa X simple", pl.Tasks[:1])
	if len(rp.Tasks) != 1 {
		t.Fatalf("expected 1 recovery task, got %d", len(rp.Tasks))
	}
	if rp.Tasks[0].MaxAttempts != 1 {
		t.Fatal("recovery tasks must not loop")
	}
}
