package personalization

import (
	"context"
	"fmt"
	"strings"

	"huginn/internal/application/ports"
	"huginn/internal/domain/profile"
)

// Builder assembles Profile + Preferences + Relevant Memories + Project.
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
	// Never send entire vault — limit to 5 memories
	_ = ports.AgentEvent{} // ensure import used
	return sb.String(), nil
}
