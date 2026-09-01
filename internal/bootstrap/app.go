package bootstrap

import (
	"huginn/internal/application/agents"
	"huginn/internal/application/orchestrator"
	"huginn/internal/application/personalization"
	"huginn/internal/application/ports"
	"huginn/internal/infrastructure/config"
	memoryinfra "huginn/internal/infrastructure/memory"
	"huginn/internal/infrastructure/profile"
	"huginn/internal/infrastructure/providers"
)

// App is the composition root: Config → Storage → Repositories → Ports → Use cases → Orchestrator → TUI
type App struct {
	Config          config.Config
	MemoryStore     ports.MemoryPort
	ProfileStore    *profile.Store
	Orchestrator    *orchestrator.Orchestrator
	Personalization *personalization.Engine
	Registry        *agents.Registry
}

func New(baseDir string) (*App, error) {
	cfg, err := config.Load(baseDir)
	if err != nil {
		cfg = config.Default()
	}
	memStore := memoryinfra.NewMarkdownStore(baseDir)
	profStore := profile.NewStore(baseDir)
	retriever := personalization.NewSimpleRetriever(memStore)
	learner := personalization.NewConservativeLearner(memStore)
	builder := personalization.NewBuilder(profStore, retriever)
	engine := personalization.NewEngine(retriever, learner, builder, memStore)

	planner := orchestrator.NewMockPlanner()
	scheduler := orchestrator.NewSimpleScheduler()
	dispatcher := orchestrator.NewNoopDispatcher()
	synthesizer := orchestrator.NewNoopSynthesizer()
	orch := orchestrator.New(planner, scheduler, dispatcher, synthesizer)

	// Registry multiagente — providers intercambiables
	reg := agents.NewRegistry()
	_ = reg.RegisterProvider(providers.NewChatGPT())
	_ = reg.RegisterProvider(providers.NewClaude())
	_ = reg.RegisterProvider(providers.NewPerplexity())
	_ = reg.RegisterProvider(providers.NewOpenCode())
	_ = reg.RegisterProvider(providers.NewKiloCode())
	_ = reg.RegisterAgent(agents.NewPlanner())
	_ = reg.RegisterAgent(agents.NewResearchAgent(providers.NewMock("huginn-research"), providers.NewPerplexity()))
	_ = reg.RegisterAgent(agents.NewResearchReviewer())
	_ = reg.RegisterAgent(agents.NewCoding(providers.NewOpenCode(), providers.NewKiloCode()))
	_ = reg.RegisterAgent(agents.NewCodeReviewer())
	_ = reg.RegisterAgent(agents.NewQA())
	_ = reg.RegisterAgent(agents.NewSecurity())
	_ = reg.RegisterAgent(agents.NewGithub())
	_ = reg.RegisterAgent(agents.NewDocs())
	_ = reg.RegisterAgent(agents.NewMemory())
	_ = reg.RegisterAgent(agents.NewKnowledge())
	_ = reg.RegisterAgent(agents.NewContext())

	return &App{
		Config:          cfg,
		MemoryStore:     memStore,
		ProfileStore:    profStore,
		Orchestrator:    orch,
		Personalization: engine,
		Registry:        reg,
	}, nil
}
