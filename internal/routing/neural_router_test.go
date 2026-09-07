// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package routing

import (
	"context"
	"testing"

	"huginn/internal/neural"
)

func TestNeuralRouterNotTrained(t *testing.T) {
	nr := NewNeuralRouter([]string{"a", "b"}, nil)
	_, err := nr.Route(context.Background(), Task{Title: "hello", Description: "world"})
	if err != ErrModelNotTrained {
		t.Fatalf("expected ErrModelNotTrained, got %v", err)
	}
}

func TestNeuralRouterTrainAndRoute(t *testing.T) {
	nr := NewNeuralRouter([]string{"a", "b"}, nil)
	ds := NewSyntheticDatasetFor([]string{"a", "b"})
	cfg := neural.DefaultTrainConfig()
	cfg.Epochs = 5
	cfg.PrintEvery = 0
	_, err := nr.Train(context.Background(), ds, cfg)
	if err != nil {
		t.Fatalf("Train: %v", err)
	}
	if !nr.IsTrained() {
		t.Fatalf("expected trained")
	}
	dec, err := nr.Route(context.Background(), Task{Title: "code task", Description: "implement feature"})
	if err != nil {
		t.Fatalf("Route: %v", err)
	}
	if dec.AgentID != "a" && dec.AgentID != "b" {
		t.Fatalf("unexpected AgentID %q", dec.AgentID)
	}
	if dec.Confidence != dec.Confidence {
		t.Fatalf("NaN confidence")
	}
	m := nr.GetMetrics()
	if m.InferenceCount != 1 {
		t.Fatalf("InferenceCount %d", m.InferenceCount)
	}
}
