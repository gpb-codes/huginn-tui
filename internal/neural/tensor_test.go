// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package neural

import (
	"math/rand"
	"testing"
)

func TestTensorValidate(t *testing.T) {
	ten, err := FromData([]float64{1, 2, 3, 4}, 2, 2)
	if err != nil {
		t.Fatalf("FromData: %v", err)
	}
	if err := ten.validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if ten.Rows != 2 || ten.Cols != 2 {
		t.Fatalf("dims %d x %d", ten.Rows, ten.Cols)
	}
}

func TestTensorForwardBackwardDims(t *testing.T) {
	m := NewModel(4)
	rng := rand.New(rand.NewSource(1))
	if err := m.AddDense(2, rng); err != nil {
		t.Fatalf("AddDense: %v", err)
	}
	m.AddActivation(NewReLU())
	in, _ := FromData([]float64{1, 0, 0, 1}, 1, 4)
	out, err := m.Forward(in)
	if err != nil {
		t.Fatalf("Forward: %v", err)
	}
	if out.Rows != 1 || out.Cols != 2 {
		t.Fatalf("out dims %d x %d", out.Rows, out.Cols)
	}
	grad, _ := FromData([]float64{1, 1}, 1, 2)
	if err := m.Backward(grad); err != nil {
		t.Fatalf("Backward: %v", err)
	}
}
