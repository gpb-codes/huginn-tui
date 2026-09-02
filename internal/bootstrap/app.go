package bootstrap

import (
	"huginn/internal/application/agents"
	"huginn/internal/application/orchestrator"
	"huginn/internal/application/personalization"
	"huginn/internal/application/ports"
	domainctx "huginn/internal/domain/context"
	"huginn/internal/infrastructure/config"
	memoryinfra "huginn/internal/infrastructure/memory"
	"huginn/internal/infrastructure/profile"
	"huginn/internal/infrastructure/providers"
	"huginn/internal/infrastructure/trace"
)

// App is the composition root: Config → Storage → Repositories → Ports → Use cases → Orchestrator → TUI
type App struct {
	Config          config.Config
	MemoryStore     ports.MemoryPort
	ProfileStore    *profile.Store
	Orchestrator    *orchestrator.Orchestrator
	Personalization *personalization.Engine
	Registry        *agents.Registry
	ContextManager  *domainctx.Manager
	Trace           *trace.Store
}

func New(baseDir, vaultPath string) (*App, error) {
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

	// Registry multiagente — providers intercambiables
	reg := agents.NewRegistry()
	_ = reg.RegisterProvider(providers.NewChatGPT())
	_ = reg.RegisterProvider(providers.NewClaude())
	_ = reg.RegisterProvider(providers.NewPerplexity())
	_ = reg.RegisterProvider(providers.NewOpenCode())
	_ = reg.RegisterProvider(providers.NewKiloCode())
	_ = reg.RegisterProvider(providers.NewGMI())
	_ = reg.RegisterProvider(providers.NewAzure())
	_ = reg.RegisterProvider(providers.NewMiniMax())
	_ = reg.RegisterProvider(providers.NewTencent())
	_ = reg.RegisterProvider(providers.NewMuseSpark())
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

	// Context + Trace
	ctxMgr := domainctx.New(vaultPath)
	traceStore := trace.New(vaultPath)

	// Orchestrator con dispatch real via Registry
	planner := orchestrator.NewDirectPlanner()
	scheduler := orchestrator.NewSimpleScheduler()
	dispatcher := orchestrator.NewRegistryDispatcher(reg, ctxMgr, traceStore)
	synthesizer := orchestrator.NewNoopSynthesizer()
	orch := orchestrator.New(planner, scheduler, dispatcher, synthesizer)

	return &App{
		Config:          cfg,
		MemoryStore:     memStore,
		ProfileStore:    profStore,
		Orchestrator:    orch,
		Personalization: engine,
		Registry:        reg,
		ContextManager:  ctxMgr,
		Trace:           traceStore,
	}, nil
}
