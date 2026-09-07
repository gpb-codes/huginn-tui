// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package bot

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"huginn/internal/application/agents"
	"huginn/internal/bootstrap"
	"huginn/internal/infrastructure/providers"
)

// newTestApp devuelve un App mínimo sin providers (sin red real).
func newTestApp() *bootstrap.App {
	return &bootstrap.App{}
}

// newTestServer compone un servidor de prueba con el token dado.
func newTestServer(t *testing.T, token string, app *bootstrap.App) *Server {
	t.Helper()
	if app == nil {
		app = newTestApp()
	}
	s, err := New(Config{Addr: "127.0.0.1:0", Token: token}, app)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestHealth(t *testing.T) {
	s := newTestServer(t, "test-token", nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if body["ok"] != true || body["version"] != Version {
		t.Fatalf("cuerpo inesperado: %v", body)
	}
}

func TestHealthRejectsWrongMethod(t *testing.T) {
	s := newTestServer(t, "test-token", nil)
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("esperaba 405, got %d", rec.Code)
	}
}

func TestStatusRejectsWithoutToken(t *testing.T) {
	s := newTestServer(t, "test-token", nil)
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401, got %d", rec.Code)
	}
}

func TestStatusRejectsWrongToken(t *testing.T) {
	s := newTestServer(t, "test-token", nil)
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401, got %d", rec.Code)
	}
}

func TestStatusOKWithToken(t *testing.T) {
	// El token viaja por env como en cmd/huginn-bot (HUGINN_BOT_TOKEN).
	t.Setenv("HUGINN_BOT_TOKEN", "env-token-123")
	s := newTestServer(t, os.Getenv("HUGINN_BOT_TOKEN"), nil)
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	req.Header.Set("Authorization", "Bearer env-token-123")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestChatRejectsWithoutToken(t *testing.T) {
	s := newTestServer(t, "test-token", nil)
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"agent":"OpenCode","prompt":"hola"}`))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401, got %d", rec.Code)
	}
}

func TestChatRejectsWrongToken(t *testing.T) {
	s := newTestServer(t, "test-token", nil)
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"agent":"OpenCode","prompt":"hola"}`))
	req.Header.Set("Authorization", "Bearer wrong")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401, got %d", rec.Code)
	}
}

func TestChatEmptyPrompt(t *testing.T) {
	s := newTestServer(t, "test-token", nil)
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"agent":"OpenCode","prompt":"   "}`))
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestChatBadJSON(t *testing.T) {
	s := newTestServer(t, "test-token", nil)
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{no json`))
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, got %d", rec.Code)
	}
}

func TestChatHonestErrorWithoutProviders(t *testing.T) {
	// Sin providers registrados: error honesto 503, sin red real.
	s := newTestServer(t, "test-token", newTestApp())
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"agent":"OpenCode","prompt":"hola"}`))
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("esperaba 503, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if body["ok"] != false {
		t.Fatalf("esperaba ok:false, got %v", body)
	}
}

func TestChatWithMockProvider(t *testing.T) {
	// MockProvider solo en tests: responde sin red real.
	reg := agents.NewRegistry()
	if err := reg.RegisterProvider(providers.NewMock("opencode")); err != nil {
		t.Fatalf("RegisterProvider: %v", err)
	}
	app := &bootstrap.App{Registry: reg}
	s := newTestServer(t, "test-token", app)
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"agent":"OpenCode","prompt":"hola"}`))
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if body["agent"] != "OpenCode" {
		t.Fatalf("agente inesperado: %v", body)
	}
	if !strings.Contains(body["text"].(string), "hola") {
		t.Fatalf("texto inesperado: %v", body)
	}
}

func TestNewRequiresToken(t *testing.T) {
	// Sin token y sin modo inseguro: el servidor no arranca.
	if _, err := New(Config{Addr: "127.0.0.1:8765"}, newTestApp()); err == nil {
		t.Fatal("esperaba error sin token")
	} else if !strings.Contains(err.Error(), "HUGINN_BOT_TOKEN") {
		t.Fatalf("error debe pedir HUGINN_BOT_TOKEN, got: %v", err)
	}
}

func TestNewNonLoopbackRequiresToken(t *testing.T) {
	if _, err := New(Config{Addr: "0.0.0.0:8765"}, newTestApp()); err == nil {
		t.Fatal("esperaba error en Addr no loopback sin token")
	}
}

func TestNewInsecureLoopbackAllowed(t *testing.T) {
	s, err := New(Config{Addr: "127.0.0.1:8765", AllowInsecureLoopback: true}, newTestApp())
	if err != nil {
		t.Fatalf("loopback explícito con AllowInsecureLoopback debe arrancar: %v", err)
	}
	if s.Addr() == "" {
		t.Fatal("Addr vacío")
	}
}

func TestNewDefaultsAddr(t *testing.T) {
	s, err := New(Config{Token: "x"}, newTestApp())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if s.Addr() != DefaultAddr {
		t.Fatalf("esperaba %s, got %s", DefaultAddr, s.Addr())
	}
}
