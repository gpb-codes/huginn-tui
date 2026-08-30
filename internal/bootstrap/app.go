package bootstrap

import (
	"huginn/internal/application/orchestrator"
	"huginn/internal/application/personalization"
	"huginn/internal/application/ports"
	"huginn/internal/infrastructure/config"
	memoryinfra "huginn/internal/infrastructure/memory"
	"huginn/internal/infrastructure/profile"
)

// App is the composition root: Config → Storage → Repositories → Ports → Use cases → Orchestrator → TUI
type App struct {
	Config          config.Config
	MemoryStore     ports.MemoryPort
	ProfileStore    *profile.Store
	Orchestrator    *orchestrator.Orchestrator
	Personalization *personalization.Engine
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

	return &App{
		Config:          cfg,
		MemoryStore:     memStore,
		ProfileStore:    profStore,
		Orchestrator:    orch,
		Personalization: engine,
	}, nil
}
