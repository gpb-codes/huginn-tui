package providers

import (
	"context"
	"time"

	"huginn/internal/domain/agent"
)

// MockProvider — stub trazable para tests y fallback
type MockProvider struct {
	name string
}

func NewMock(name string) *MockProvider { return &MockProvider{name: name} }
func (m *MockProvider) Name() string { return m.name }
func (m *MockProvider) Available(_ context.Context) bool { return true }
func (m *MockProvider) Invoke(_ context.Context, req agent.ProviderRequest) (agent.ProviderResponse, error) {
	start := time.Now()
	return agent.ProviderResponse{
		Content:  "[mock " + m.name + "] " + req.Prompt,
		Provider: m.name,
		Latency:  time.Since(start),
	}, nil
}

// ChatGPTProvider / ClaudeProvider — por ahora delegan a mock pero separan config
type ChatGPTProvider struct{ MockProvider }
type ClaudeProvider struct{ MockProvider }

func NewChatGPT() *ChatGPTProvider { return &ChatGPTProvider{MockProvider{name: "chatgpt"}} }
func NewClaude() *ClaudeProvider   { return &ClaudeProvider{MockProvider{name: "claude"}} }

func (c *ChatGPTProvider) Name() string { return "chatgpt" }
func (c *ClaudeProvider) Name() string  { return "claude" }

// PerplexityProvider — research profundo
type PerplexityProvider struct{ MockProvider }
func NewPerplexity() *PerplexityProvider { return &PerplexityProvider{MockProvider{name: "perplexity"}} }
func (p *PerplexityProvider) Name() string { return "perplexity" }

// OpenCodeProvider / KiloCodeProvider — coding
type OpenCodeProvider struct{ MockProvider }
type KiloCodeProvider struct{ MockProvider }
func NewOpenCode() *OpenCodeProvider { return &OpenCodeProvider{MockProvider{name: "opencode"}} }
func NewKiloCode() *KiloCodeProvider { return &KiloCodeProvider{MockProvider{name: "kilo"}} }
func (o *OpenCodeProvider) Name() string { return "opencode" }
func (k *KiloCodeProvider) Name() string { return "kilo" }
