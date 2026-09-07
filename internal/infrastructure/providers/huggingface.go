// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"huginn/internal/domain/agent"
)

// HuggingFaceProvider usa la Inference API remota de HuggingFace.
// No requiere modelos locales: consume HF_TOKEN y un modelo como
// "Qwen/Qwen3-Coder-480B-A35B-Instruct" o "deepseek-ai/DeepSeek-R1".
// Documentación: https://huggingface.co/docs/api-inference
type HuggingFaceProvider struct {
	model   string
	token   string
	baseURL string
	client  *http.Client
}

// NewHuggingFaceProvider crea el provider. token puede venir de
// config.Models.HuggingFaceToken o de HF_TOKEN / HUGGINGFACE_TOKEN.
func NewHuggingFaceProvider(model, token string) *HuggingFaceProvider {
	if strings.TrimSpace(model) == "" {
		model = "Qwen/Qwen2.5-Coder-32B-Instruct"
	}
	return &HuggingFaceProvider{
		model:   strings.TrimSpace(model),
		token:   strings.TrimSpace(token),
		baseURL: "https://api-inference.huggingface.co/models",
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *HuggingFaceProvider) Name() string { return "huggingface:" + p.model }

func (p *HuggingFaceProvider) Available(_ context.Context) bool {
	return p.token != "" && p.model != ""
}

func (p *HuggingFaceProvider) Invoke(ctx context.Context, req agent.ProviderRequest) (agent.ProviderResponse, error) {
	if p.token == "" {
		return agent.ProviderResponse{}, fmt.Errorf("huggingface: falta HF_TOKEN (config models.huggingface_token o env HF_TOKEN)")
	}
	// Permite override del modelo por Meta["model"]
	model := p.model
	if m, ok := req.Meta["model"]; ok && strings.TrimSpace(m) != "" {
		model = strings.TrimSpace(m)
	}
	url := p.baseURL + "/" + model

	payload := map[string]any{
		"inputs": req.Prompt,
		"parameters": map[string]any{
			"max_new_tokens":   1024,
			"temperature":      0.7,
			"return_full_text": false,
		},
	}
	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return agent.ProviderResponse{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.token)
	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return agent.ProviderResponse{}, fmt.Errorf("huggingface: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return agent.ProviderResponse{}, fmt.Errorf("huggingface: no autorizado (revisa HF_TOKEN) — %s", strings.TrimSpace(string(b)))
	}
	if resp.StatusCode == 404 {
		return agent.ProviderResponse{}, fmt.Errorf("huggingface: modelo no encontrado %q — %s", model, strings.TrimSpace(string(b)))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(b))
		if len(msg) > 500 {
			msg = msg[:500] + "…"
		}
		return agent.ProviderResponse{}, fmt.Errorf("huggingface: HTTP %d — %s", resp.StatusCode, msg)
	}

	// La API puede devolver [{"generated_text":"..."}] o {"generated_text":"..."} o {"error":...}
	var content string
	// Intenta array
	var arr []map[string]any
	if err := json.Unmarshal(b, &arr); err == nil && len(arr) > 0 {
		if v, ok := arr[0]["generated_text"].(string); ok {
			content = v
		}
	}
	if content == "" {
		var obj map[string]any
		if err := json.Unmarshal(b, &obj); err == nil {
			if v, ok := obj["generated_text"].(string); ok {
				content = v
			} else if v, ok := obj["error"].(string); ok {
				return agent.ProviderResponse{}, fmt.Errorf("huggingface: %s", v)
			}
		}
	}
	if content == "" {
		// Fallback: usa el body crudo si no hay generated_text
		content = strings.TrimSpace(string(b))
		// Si es JSON con "error", ya se manejó arriba
		if content == "" {
			content = "(respuesta vacía de huggingface)"
		}
	}

	return agent.ProviderResponse{
		Content:  strings.TrimSpace(content),
		Provider: p.Name(),
		Latency:  time.Since(start),
	}, nil
}
