package agents

import (
	"context"
	"testing"

	"huginn/internal/domain/agent"
	"huginn/internal/infrastructure/providers"
)

func TestRegistry_RegisterAndSelect(t *testing.T) {
	r := NewRegistry()
	if err := r.RegisterProvider(providers.NewChatGPT()); err != nil {
		t.Fatalf("register chatgpt: %v", err)
	}
	if err := r.RegisterProvider(providers.NewClaude()); err != nil {
		t.Fatalf("register claude: %v", err)
	}
	if err := r.RegisterProvider(providers.NewChatGPT()); err == nil {
		t.Fatal("esperaba error duplicado")
	}
	// preferencia claude
	r.SetPreferred("orchestrator", "claude", agent.ProviderConfig{Name: "claude", Priority: 0})
	candidates := []agent.Provider{providers.NewChatGPT(), providers.NewClaude()}
	sel, err := r.SelectProvider(context.Background(), "orchestrator", candidates)
	if err != nil || sel.Name() != "claude" {
		t.Fatalf("select claude: %v %v", sel, err)
	}
	// fallback si preferido no disponible — mock siempre disponible, probamos sin preferencia
	r2 := NewRegistry()
	r2.RegisterProvider(providers.NewChatGPT())
	r2.RegisterProvider(providers.NewClaude())
	sel2, _ := r2.SelectProvider(context.Background(), "orchestrator", candidates)
	if sel2 == nil {
		t.Fatal("select sin pref fallo")
	}
}

func TestRegistry_RegisterAgent(t *testing.T) {
	r := NewRegistry()
	// agent stub
	a := &mockAgent{name: "research", types: []string{"research"}}
	if err := r.RegisterAgent(a); err != nil {
		t.Fatal(err)
	}
	if a2, ok := r.ResolveAgent("research"); !ok || a2.Name() != "research" {
		t.Fatalf("resolve research: %v", ok)
	}
	if _, ok := r.ResolveAgent("coding"); ok {
		t.Fatal("no debe resolver coding")
	}
}

type mockAgent struct {
	name  string
	types []string
}

func (m *mockAgent) Name() string { return m.name }
func (m *mockAgent) Role() string { return "test" }
func (m *mockAgent) CanHandle(t string) bool {
	for _, tt := range m.types {
		if tt == t {
			return true
		}
	}
	return false
}
func (m *mockAgent) Execute(_ context.Context, _ agent.AgentTask) (agent.AgentResult, error) {
	return agent.AgentResult{}, nil
}
func (m *mockAgent) Providers() []agent.Provider { return nil }
