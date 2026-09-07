// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package personalization

import (
	"context"
	"fmt"
	"strings"

	"huginn/internal/application/ports"
	"huginn/internal/domain/profile"
)

// Builder combina perfil, preferencias, memorias relevantes y proyecto.
type Builder struct {
	profileStore interface {
		Load() (profile.Profile, error)
	}
	retriever Retriever
}

func NewBuilder(profileStore interface {
	Load() (profile.Profile, error)
}, retriever Retriever) *Builder {
	return &Builder{profileStore: profileStore, retriever: retriever}
}

func (b *Builder) Build(ctx context.Context, prompt, project string) (string, error) {
	prof, _ := b.profileStore.Load()
	mems, _ := b.retriever.Retrieve(ctx, prompt, 5)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Profile: %s, style %s\n", prof.Communication.Language, prof.Communication.Style))
	sb.WriteString(fmt.Sprintf("Project: %s\n", project))
	if len(mems) > 0 {
		sb.WriteString("Relevant memories:\n")
		for _, m := range mems {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", m.Title, m.Content))
		}
	}
	// Nunca envía el vault completo: limita a 5 memorias.
	_ = ports.AgentEvent{} // ensure import used
	return sb.String(), nil
}
