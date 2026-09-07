// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package models

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"huginn/internal/domain/agent"
)

func TestAvailable_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Write([]byte(`{"models":[]}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()
	o := NewOllamaProvider(srv.URL, "test")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if !o.Available(ctx) {
		t.Fatal("expected available")
	}
}

func TestAvailable_Down(t *testing.T) {
	o := NewOllamaProvider("http://127.0.0.1:1", "test")
	o.Client.Timeout = time.Second
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if o.Available(ctx) {
		t.Fatal("expected unavailable")
	}
}

func TestInvoke_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"model":"test","response":" hola mundo ","done":true}`))
	}))
	defer srv.Close()
	o := NewOllamaProvider(srv.URL, "test")
	resp, err := o.Invoke(context.Background(), agent.ProviderRequest{Prompt: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "hola mundo" {
		t.Fatalf("bad content: %q", resp.Content)
	}
	if resp.Meta["model"] != "test" {
		t.Fatal("model meta missing")
	}
}

func TestInvoke_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("boom"))
	}))
	defer srv.Close()
	o := NewOllamaProvider(srv.URL, "test")
	if _, err := o.Invoke(context.Background(), agent.ProviderRequest{Prompt: "hi"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestInvoke_ModelError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"error":"model not found"}`))
	}))
	defer srv.Close()
	o := NewOllamaProvider(srv.URL, "missing")
	if _, err := o.Invoke(context.Background(), agent.ProviderRequest{Prompt: "hi"}); err == nil {
		t.Fatal("expected model error")
	}
}
