// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package bootstrap

import (
	"os"
	"strings"
	"time"

	"huginn/internal/application/agents"
	"huginn/internal/application/evaluator"
	"huginn/internal/application/orchestrator"
	"huginn/internal/application/personalization"
	"huginn/internal/application/planner"
	"huginn/internal/application/ports"
	"huginn/internal/application/router"
	domainagent "huginn/internal/domain/agent"
	domainctx "huginn/internal/domain/context"
	infraagents "huginn/internal/infrastructure/agents"
	"huginn/internal/infrastructure/config"
	memoryinfra "huginn/internal/infrastructure/memory"
	"huginn/internal/infrastructure/models"
	"huginn/internal/infrastructure/profile"
	"huginn/internal/infrastructure/providers"
	"huginn/internal/infrastructure/security"
	"huginn/internal/infrastructure/trace"
)

// App es la raíz de composición: Config → Storage → Repositories → Ports → casos de uso → Orchestrator → TUI.
type App struct {
	Config          config.Config
	MemoryStore     ports.MemoryPort
	ProfileStore    *profile.Store
	Personalization *personalization.Engine
	Registry        *agents.Registry
	ContextManager  *domainctx.Manager
	Trace           *trace.Store
	// Agent system (Huginn piensa y coordina, los agentes ejecutan).
	Planner  *planner.Planner
	Router   *router.Router
	Pipeline *orchestrator.Pipeline
	// PipeOptions reutiliza la configuración del pipeline para pipelines acotados a una ejecución con su propio manejador de eventos (ver NewPipelineRun).
	PipeOptions orchestrator.Options
}

// NewPipelineRun construye un Pipeline acotado a una ejecución que comparte planner, router y opciones de App,
// pero entrega los eventos de ciclo de vida a h (p. ej., la TUI).
func (a *App) NewPipelineRun(h orchestrator.EventHandler) *orchestrator.Pipeline {
	opts := a.PipeOptions
	opts.OnEvent = h
	return orchestrator.NewPipeline(a.Planner, a.Router, opts)
}

func useMocks() bool {
	return os.Getenv("HUGINN_USE_MOCKS") == "1"
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

	// Proveedores reales: opencode (runtime de código) + ollama (modelos locales).
	// Mocks solo para desarrollo offline de la UI (HUGINN_USE_MOCKS=1).
	opencodeProvider := providers.NewOpenCodeRuntimeWithConfig(infraagents.Config{
		Model: cfg.Models.OpenCodeModel,
		Agent: cfg.Models.OpenCodeAgent,
	})
	ollama := models.NewOllamaProvider(cfg.Models.OllamaHost, cfg.Models.OllamaModel).
		WithModels(models.ModelMatrix(cfg.Models.OllamaModels))

	reg := agents.NewRegistry()
	_ = reg.RegisterProvider(opencodeProvider)
	_ = reg.RegisterProvider(ollama)
	// HuggingFace remoto (si hay token, tiene prioridad sobre Ollama local)
	hfToken := strings.TrimSpace(cfg.Models.HuggingFaceToken)
	if hfToken == "" {
		hfToken = strings.TrimSpace(os.Getenv("HF_TOKEN"))
	}
	if hfToken == "" {
		hfToken = strings.TrimSpace(os.Getenv("HUGGINGFACE_TOKEN"))
	}
	if hfToken == "" {
		hfToken = strings.TrimSpace(os.Getenv("HUGGING_FACE_HUB_TOKEN"))
	}
	var hfProvider *providers.HuggingFaceProvider
	if hfToken != "" {
		hfModel := strings.TrimSpace(cfg.Models.HuggingFaceModel)
		if hfModel == "" {
			hfModel = "Qwen/Qwen2.5-Coder-32B-Instruct"
		}
		hfProvider = providers.NewHuggingFaceProvider(hfModel, hfToken)
		_ = reg.RegisterProvider(hfProvider)
	}
	var researchFallbacks []domainagent.Provider
	if hfProvider != nil {
		researchFallbacks = append(researchFallbacks, hfProvider)
	}
	researchFallbacks = append(researchFallbacks, ollama)
	if useMocks() {
		for _, name := range []string{"mock-chat", "mock-code", "mock-research"} {
			_ = reg.RegisterProvider(providers.NewMock(name))
		}
		researchFallbacks = append(researchFallbacks, providers.NewMock("mock-research"))
	}

	_ = reg.RegisterAgent(agents.NewPlanner())
	_ = reg.RegisterAgent(agents.NewResearchAgent(researchFallbacks...))
	_ = reg.RegisterAgent(agents.NewResearchReviewer())
	_ = reg.RegisterAgent(agents.NewCoding(opencodeProvider))
	// Review/docs/memory usan HF remoto si hay token, si no Ollama local
	textProvider := domainagent.Provider(ollama)
	if hfProvider != nil {
		textProvider = hfProvider
	}
	_ = reg.RegisterAgent(agents.NewCodeReviewer(textProvider))
	_ = reg.RegisterAgent(agents.NewQA(opencodeProvider))
	_ = reg.RegisterAgent(agents.NewSecurity())
	_ = reg.RegisterAgent(agents.NewGithub(textProvider))
	_ = reg.RegisterAgent(agents.NewDocs(textProvider))
	_ = reg.RegisterAgent(agents.NewMemory(textProvider))
	_ = reg.RegisterAgent(agents.NewKnowledge(textProvider))
	_ = reg.RegisterAgent(agents.NewContext(textProvider))

	// Contexto y traza compartidos.
	ctxMgr := domainctx.New(vaultPath)
	traceStore := trace.New(vaultPath)

	// Sistema de agentes: NeuralPlanner -> Router por capacidad -> Pipeline.
	neuralPlanner := planner.New()
	rt := router.New()
	rt.Register(infraagents.NewOpenCodeRuntimeWithConfig(infraagents.Config{
		Model: cfg.Models.OpenCodeModel,
		Agent: cfg.Models.OpenCodeAgent,
	}))
	rt.Register(infraagents.NewModelRuntime("ollama", "Ollama", ollama,
		domainagent.CapabilityResearch,
		domainagent.CapabilityReasoning,
		domainagent.CapabilityDocumentation,
		domainagent.CapabilityReview,
	))
	if hfProvider != nil {
		rt.Register(infraagents.NewModelRuntime("huggingface", "Hugging Face", hfProvider,
			domainagent.CapabilityResearch,
			domainagent.CapabilityReasoning,
			domainagent.CapabilityCoding,
			domainagent.CapabilityDocumentation,
			domainagent.CapabilityReview,
			domainagent.CapabilityTesting,
		))
	}
	rt.Register(infraagents.NewKiloRuntime())

	policy := security.NewPolicy()
	for _, id := range cfg.Permissions.AllowedWriteAgents {
		policy.Allow(id, "*")
	}
	var approver orchestrator.Approver = orchestrator.DenyApprover{}
	if cfg.Permissions.AutoApprove {
		approver = orchestrator.AllowApprover{}
	}
	pipeOpts := orchestrator.Options{
		Retry: evaluator.RetryPolicy{
			MaxAttempts: cfg.Retry.MaxAttempts,
			BackoffBase: time.Duration(cfg.Retry.BackoffSecs) * time.Second,
		},
		MaxReplans:    cfg.Planner.MaxReplans,
		MaxConcurrent: cfg.Planner.MaxParallel,
		TaskTimeout:   time.Duration(cfg.Timeouts.TaskSeconds) * time.Second,
		Tracer:        traceStore,
		Approver:      approver,
		Permissions:   policy,
	}
	pipe := orchestrator.NewPipeline(neuralPlanner, rt, pipeOpts)

	return &App{
		Config:          cfg,
		MemoryStore:     memStore,
		ProfileStore:    profStore,
		Personalization: engine,
		Registry:        reg,
		ContextManager:  ctxMgr,
		Trace:           traceStore,
		Planner:         neuralPlanner,
		Router:          rt,
		Pipeline:        pipe,
		PipeOptions:     pipeOpts,
	}, nil
}
