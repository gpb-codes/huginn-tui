// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package models

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"huginn/internal/domain/agent"
)

func TestMatrixFor(t *testing.T) {
	m := ModelMatrix{"review": "llama3.1", "coding": "qwen3"}
	if m.For("review") != "llama3.1" {
		t.Fatal("exact match failed")
	}
	if m.For(" Review ") != "llama3.1" {
		t.Fatal("lookup must ignore case and whitespace")
	}
	if m.For("unknown") != "" {
		t.Fatal("unknown capability must return empty (caller falls back)")
	}
	if (ModelMatrix)(nil).For("review") != "" {
		t.Fatal("nil matrix must return empty")
	}
}

// captureModel ejecuta Invoke contra un stub y devuelve el modelo pedido.
func captureModel(t *testing.T, o *OllamaProvider, req agent.ProviderRequest) string {
	t.Helper()
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var gr generateRequest
		_ = json.NewDecoder(r.Body).Decode(&gr)
		got = gr.Model
		w.Write([]byte(`{"response":"ok","done":true}`))
	}))
	defer srv.Close()
	o.BaseURL = srv.URL
	resp, err := o.Invoke(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Meta["model"] != got {
		t.Fatalf("response meta model %q != requested %q", resp.Meta["model"], got)
	}
	return got
}

func TestInvoke_SelectsModelByCapability(t *testing.T) {
	o := NewOllamaProvider("http://localhost:1", "llama3.1").
		WithModels(ModelMatrix{"review": "review-model", "coding": "qwen3"})
	req := agent.ProviderRequest{Prompt: "hi", Meta: map[string]string{"capability": "review"}}
	if got := captureModel(t, o, req); got != "review-model" {
		t.Fatalf("capability model = %q, want review-model", got)
	}
}

func TestInvoke_ExplicitModelWins(t *testing.T) {
	o := NewOllamaProvider("http://localhost:1", "llama3.1").
		WithModels(ModelMatrix{"review": "review-model"})
	req := agent.ProviderRequest{Prompt: "hi", Meta: map[string]string{"capability": "review", "model": "custom"}}
	if got := captureModel(t, o, req); got != "custom" {
		t.Fatalf("explicit model = %q, want custom", got)
	}
}

func TestInvoke_FallsBackToDefault(t *testing.T) {
	o := NewOllamaProvider("http://localhost:1", "llama3.1").
		WithModels(ModelMatrix{"review": "review-model"})
	// Sin capacidad conocida ni matriz: aplica el defecto del proveedor.
	for _, req := range []agent.ProviderRequest{
		{Prompt: "hi", Meta: map[string]string{"capability": "unknown-cap"}},
		{Prompt: "hi"},
	} {
		if got := captureModel(t, o, req); got != "llama3.1" {
			t.Fatalf("fallback model = %q, want llama3.1", got)
		}
	}
}
