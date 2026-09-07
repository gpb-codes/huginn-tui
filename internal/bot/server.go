// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
// Package bot expone un servidor HTTP local de control remoto para Huginn.
//
// Solo stdlib (net/http, encoding/json). Reutiliza el Registry del
// bootstrap.App igual que dispatchViaRegistry en dispatch.go: nunca inventa
// respuestas; sin provider disponible devuelve un error JSON honesto.
// Todos los bucles/llamadas están acotados con timeouts y el token jamás se
// escribe en logs ni en la traza.
package bot

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"huginn/internal/bootstrap"
	agentpkg "huginn/internal/domain/agent"
	"huginn/internal/domain/execution"
)

// Version es la versión reportada por GET /health.
// Mantener sincronizada con palette.go y cmd/huginn VERSION.
const Version = "v0.2.0"

// DefaultAddr es la dirección de escucha por defecto (solo loopback).
const DefaultAddr = "127.0.0.1:8765"

// Cotas obligatorias (AGENTS.md: acotar todo loop).
const (
	// statusTimeout acota la sonda de providers en GET /status.
	statusTimeout = 5 * time.Second
	// chatTimeout acota la invocación de provider en POST /chat (igual que dispatch.go: 90s, con margen hasta 95s).
	chatTimeout = 95 * time.Second
	// maxChatBody acota el cuerpo de POST /chat a 1 MiB.
	maxChatBody = 1 << 20
	// shutdownTimeout acota el apagado graceful.
	shutdownTimeout = 10 * time.Second
)

// Config agrupa la configuración del servidor.
type Config struct {
	// Addr es host:puerto de escucha. Vacío usa DefaultAddr.
	Addr string
	// Token es el secreto Bearer (llega por HUGINN_BOT_TOKEN en cmd/huginn-bot).
	// Jamás se loguea.
	Token string
	// VaultPath es la ruta del vault (se reporta en /status y da contexto).
	VaultPath string
	// ProjectPath es el proyecto activo (workspace para los providers).
	ProjectPath string
	// AllowInsecureLoopback permite arrancar sin token solo en loopback explícito.
	AllowInsecureLoopback bool
}

// Server es el bot de control remoto: mux + http.Server sobre el App.
type Server struct {
	cfg Config
	app *bootstrap.App
	mux *http.ServeMux
	srv *http.Server
}

// New valida la config y compone las rutas. No abre ningún socket.
func New(cfg Config, app *bootstrap.App) (*Server, error) {
	if app == nil {
		return nil, errors.New("bot: app nil, construye con bootstrap.New primero")
	}
	if strings.TrimSpace(cfg.Addr) == "" {
		cfg.Addr = DefaultAddr
	}
	// Sin token no se arranca, salvo loopback explícito con AllowInsecureLoopback.
	// El token llega por HUGINN_BOT_TOKEN (ver cmd/huginn-bot).
	if strings.TrimSpace(cfg.Token) == "" {
		if !isLoopbackAddr(cfg.Addr) {
			return nil, errors.New("bot: Addr no es loopback y falta token: fija HUGINN_BOT_TOKEN")
		}
		if !cfg.AllowInsecureLoopback {
			return nil, errors.New("bot: falta token: fija HUGINN_BOT_TOKEN o usa AllowInsecureLoopback solo en loopback")
		}
	}
	s := &Server{cfg: cfg, app: app}
	mux := http.NewServeMux()
	// /health es público (sonda); /status y /chat exigen Bearer.
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/status", s.requireAuth(s.handleStatus))
	mux.HandleFunc("/chat", s.requireAuth(s.handleChat))
	s.mux = mux
	s.srv = &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s, nil
}

// Handler expone el mux para tests con httptest (sin red real).
func (s *Server) Handler() http.Handler { return s.mux }

// Addr devuelve la dirección configurada.
func (s *Server) Addr() string { return s.cfg.Addr }

// Run escucha hasta que ctx se cancela y apaga con gracia.
// Retorna nil al apagado limpio; cualquier error de escucha es honesto.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = s.srv.Shutdown(shutCtx)
		return nil
	case err := <-errCh:
		return err
	}
}

// Shutdown apaga el servidor con el contexto dado.
func (s *Server) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }

// requireAuth exige `Authorization: Bearer <token>` en tiempo constante.
// En modo inseguro loopback (sin token) deja pasar; en otro caso 401 honesto.
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Token == "" {
			next(w, r)
			return
		}
		const prefix = "Bearer "
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, prefix) {
			writeError(w, http.StatusUnauthorized, "no autorizado: falta header Authorization Bearer")
			return
		}
		got := strings.TrimSpace(strings.TrimPrefix(h, prefix))
		if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(s.cfg.Token)) != 1 {
			writeError(w, http.StatusUnauthorized, "no autorizado: token inválido")
			return
		}
		next(w, r)
	}
}

// handleHealth responde la sonda pública.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "método no permitido, usa GET")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "version": Version})
}

// handleStatus reporta vault, proyecto y disponibilidad de providers.
// Cada sonda de provider hereda un contexto acotado a 5s.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "método no permitido, usa GET")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), statusTimeout)
	defer cancel()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"version": Version,
		"vault":   s.cfg.VaultPath,
		"project": s.cfg.ProjectPath,
		"providers": map[string]bool{
			"opencode": providerAvailable(ctx, s.app, "opencode"),
			"ollama":   providerAvailable(ctx, s.app, "ollama"),
		},
	})
}

// chatRequest es el cuerpo de POST /chat.
type chatRequest struct {
	Agent  string `json:"agent"`
	Prompt string `json:"prompt"`
}

// handleChat valida el JSON, acota a 95s e invoca vía el Registry del App.
// Sin provider disponible responde error JSON honesto (nunca mock en producción).
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "método no permitido, usa POST")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxChatBody)
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	agent := strings.TrimSpace(req.Agent)
	if agent == "" {
		agent = "OpenCode"
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, http.StatusBadRequest, "prompt vacío")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), chatTimeout)
	defer cancel()

	p := pickProvider(ctx, s.app, agent)
	if p == nil {
		writeError(w, http.StatusServiceUnavailable, offlineNotice(agent))
		return
	}

	// Contexto del vault (honesto: vacío si no hay manager).
	var vaultCtx string
	if s.app.ContextManager != nil {
		vaultCtx, _ = s.app.ContextManager.Load()
	}
	vaultPath := ""
	if s.app.Trace != nil {
		vaultPath = s.app.Trace.VaultPath()
	}
	actx := agentpkg.AgentContext{
		VaultPath:   vaultPath,
		ProjectPath: s.cfg.ProjectPath,
		Memory:      []string{vaultCtx},
	}
	meta := map[string]string{"agent": agent}
	if p.Name() == "ollama" {
		if m := strings.TrimSpace(s.app.Config.Models.OllamaModels["chat"]); m != "" {
			meta["model"] = m
		}
	}

	start := time.Now()
	resp, err := p.Invoke(ctx, agentpkg.ProviderRequest{
		Prompt:  req.Prompt,
		Context: actx,
		Meta:    meta,
	})

	// Traza acotada y saneada (el Store redacta secretos).
	if s.app.Trace != nil {
		status := "ok"
		if err != nil {
			status = "error"
		}
		_ = s.app.Trace.Append(execution.Record{
			ExecutionID: fmt.Sprintf("bot-chat-%d", start.UnixMilli()),
			Agent:       agent,
			Provider:    p.Name(),
			TaskID:      fmt.Sprintf("bot-chat-%d", start.UnixMilli()),
			TaskType:    "chat",
			Input:       req.Prompt,
			Status:      status,
			StartedAt:   start,
		})
	}

	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("Error %s: %v", agent, err))
		return
	}
	text := strings.TrimSpace(resp.Content)
	if text == "" {
		// Sin éxitos falsos: contenido vacío es error honesto, no {ok:true}.
		writeError(w, http.StatusServiceUnavailable, offlineNotice(agent))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"agent": agent, "text": text})
}

// providerAvailable sondea un provider del Registry con el ctx acotado dado.
func providerAvailable(ctx context.Context, app *bootstrap.App, name string) bool {
	if app == nil || app.Registry == nil {
		return false
	}
	p, ok := app.Registry.GetProvider(name)
	if !ok || p == nil {
		return false
	}
	return p.Available(ctx)
}

// pickProvider prueba el provider asignado y luego cualquier runtime real.
// Copia la lógica de dispatch.go: nunca devuelve mock salvo registro explícito.
func pickProvider(ctx context.Context, app *bootstrap.App, agentName string) agentpkg.Provider {
	if app == nil || app.Registry == nil {
		return nil
	}
	for _, name := range []string{mapAgentToProvider(agentName), "opencode", "ollama"} {
		if p, ok := app.Registry.GetProvider(name); ok && p != nil && p.Available(ctx) {
			return p
		}
	}
	return nil
}

// mapAgentToProvider asigna cada agente de la TUI a su provider preferido.
func mapAgentToProvider(agent string) string {
	switch agent {
	case "ChatGPT":
		return "ollama"
	case "OpenCode":
		return "opencode"
	case "Mimo Code":
		return "ollama"
	case "Kilo Code":
		return "opencode"
	case "Muse Code":
		return "ollama"
	default:
		return "opencode"
	}
}

// offlineNotice es el error honesto sin runtime disponible.
func offlineNotice(agent string) string {
	return fmt.Sprintf("%s no disponible: sin provider en línea (¿ollama en marcha? ¿opencode instalado?).", agent)
}

// writeJSON escribe un objeto JSON con su código de estado.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError escribe un error JSON honesto sin filtrar secretos (el mensaje
// nunca incluye el token ni el prompt completo).
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"ok": false, "error": msg})
}

// isLoopbackAddr dice si el host de addr es loopback.
// ":puerto" (todas las interfaces) NO es loopback.
func isLoopbackAddr(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
