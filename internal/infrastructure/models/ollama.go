// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
// Package models agrupa proveedores de modelos que ejecutan inferencia.
//
// Un proveedor no es un agente (ver docs/architecture/audit.md §14):
// aporta tokens; planner, router y agentes deciden su uso.
package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"huginn/internal/domain/agent"
)

// OllamaProvider habla con el daemon local de Ollama por HTTP.
// Usa GET /api/tags para detección y POST /api/generate para inferencia.
// Es el defecto local para research/reasoning/documentación/review; solo stdlib.
type OllamaProvider struct {
	BaseURL string
	Model   string
	// Models sustituye Model por capacidad (ver ModelMatrix); vacíos usan Model.
	Models ModelMatrix
	Client *http.Client
}

// WithModels fija la matriz por capacidad y devuelve o para encadenar.
func (o *OllamaProvider) WithModels(m ModelMatrix) *OllamaProvider {
	o.Models = m
	return o
}

// selectModel resuelve el modelo: Meta["model"] > matriz > Model por defecto.
func (o *OllamaProvider) selectModel(req agent.ProviderRequest) string {
	if m := strings.TrimSpace(req.Meta["model"]); m != "" {
		return m
	}
	if mm := o.Models.For(req.Meta["capability"]); mm != "" {
		return mm
	}
	return o.Model
}

// NewOllamaProvider construye el proveedor con defecto local y modelo llama3.1.
func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = os.Getenv("OLLAMA_HOST")
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
	}
	if model == "" {
		if m := os.Getenv("OLLAMA_MODEL"); m != "" {
			model = m
		} else {
			model = "llama3.1"
		}
	}
	return &OllamaProvider{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Model:   model,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (o *OllamaProvider) Name() string { return "ollama" }

// Available sondea /api/tags con timeout corto sin bloquear.
func (o *OllamaProvider) Available(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.BaseURL+"/api/tags", nil)
	if err != nil {
		return false
	}
	resp, err := o.Client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// Invoke ejecuta una generación sin streaming; el contexto fija el deadline.
func (o *OllamaProvider) Invoke(ctx context.Context, req agent.ProviderRequest) (agent.ProviderResponse, error) {
	start := time.Now()
	prompt := req.Prompt
	if len(req.Context.Memory) > 0 {
		prompt = "Contexto relevante:\n" + strings.Join(req.Context.Memory, "\n---\n") +
			"\n\nTarea: " + req.Prompt
	}
	model := o.selectModel(req)
	body, _ := json.Marshal(generateRequest{Model: model, Prompt: prompt, Stream: false})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.BaseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return agent.ProviderResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := o.Client.Do(httpReq)
	if err != nil {
		return agent.ProviderResponse{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return agent.ProviderResponse{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return agent.ProviderResponse{}, fmt.Errorf("ollama: status %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	var gr generateResponse
	if err := json.Unmarshal(data, &gr); err != nil {
		return agent.ProviderResponse{}, fmt.Errorf("ollama: invalid response: %w", err)
	}
	if gr.Error != "" {
		return agent.ProviderResponse{}, fmt.Errorf("ollama: %s", gr.Error)
	}
	return agent.ProviderResponse{
		Content:  strings.TrimSpace(gr.Response),
		Provider: o.Name(),
		Latency:  time.Since(start),
		Meta:     map[string]string{"model": model},
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
