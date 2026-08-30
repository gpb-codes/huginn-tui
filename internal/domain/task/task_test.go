package task

import "testing"

func TestTaskStatus_Terminal(t *testing.T) {
	if !StatusCompleted.IsTerminal() || !StatusFailed.IsTerminal() {
		t.Fatal("terminal status failed")
	}
	if StatusRunning.IsTerminal() {
		t.Fatal("running should not be terminal")
	}
}

func TestGraph_Ready(t *testing.T) {
	g := NewGraph()
	a := New("a", "Research", "", "researcher")
	b := New("b", "Architecture", "", "planner", "a")
	c := New("c", "Implementation", "", "coder", "b")
	g.Add(a)
	g.Add(b)
	g.Add(c)
	g.Tasks["a"].Status = StatusCompleted
	ready := g.Ready()
	if len(ready) != 1 || ready[0].ID != "b" {
		t.Fatalf("expected b ready, got %v", ready)
	}
}

func TestGraph_TopologicalOrder(t *testing.T) {
	g := NewGraph()
	t1 := New("1", "Research", "", "researcher")
	t2 := New("2", "Architecture", "", "planner", "1")
	t3 := New("3", "Implementation", "", "coder", "2")
	g.Add(t3)
	g.Add(t1)
	g.Add(t2)
	order := g.TopologicalOrder()
	if len(order) != 3 {
		t.Fatalf("expected 3, got %d", len(order))
	}
	if order[0].ID != "1" || order[1].ID != "2" || order[2].ID != "3" {
		t.Fatalf("wrong order %v %v %v", order[0].ID, order[1].ID, order[2].ID)
	}
}

func TestGraph_TopologicalOrder_Complex(t *testing.T) {
	g := NewGraph()
	// Research -> Architecture -> Implementation -> Tests -> Review
	tasks := []Task{
		New("1", "Research", "", "researcher"),
		New("2", "Architecture", "", "planner", "1"),
		New("3", "Implementation", "", "coder", "2"),
		New("4", "Tests", "", "coder", "3"),
		New("5", "Review", "", "reviewer", "4"),
	}
	for _, tc := range tasks {
		g.Add(tc)
	}
	order := g.TopologicalOrder()
	if len(order) != 5 {
		t.Fatalf("expected 5")
	}
	// ensure dependencies before dependents
	pos := map[string]int{}
	for i, tt := range order {
		pos[tt.ID] = i
	}
	for _, tt := range tasks {
		for _, dep := range tt.Dependencies {
			if pos[dep] > pos[tt.ID] {
				t.Fatalf("dependency %s should be before %s", dep, tt.ID)
			}
		}
	}
}
