# Huginn — Arquitectura limpia (rebuild 2026-09)

> Rebuild completo desde cero (2026-09): TUI minimalista, domain puro, eventos desacoplados.
> Ver `docs/architecture/overview.md`, `docs/agents.md`.

```
                         USER
                           │
                           ▼
                        HUGINN
                           │
                  ┌────────┴────────┐
                  │                 │
              PLANNER           CONTEXT
                  │                 │
                  └────────┬────────┘
                           │
                     AGENT ROUTER
                           │
          ┌────────────────┼────────────────┐
          │                │                │
       OpenCode         Hermes           Kilo
          │                │                │
          └────────────────┼────────────────┘
                           │
                         RESULT
```

## Capas y regla de dependencia

```
cmd/huginn      →  internal/tui  →  internal/events  →  internal/domain
internal/tui    →  internal/cli  →  internal/domain  ←  internal/infrastructure
```

`domain` nunca importa `lipgloss`/`bubbletea` ni `infrastructure`. `events` es pub/sub tipado sin dependencias. `tui` solo renderiza y envía comandos; todo `exec` vive en `infrastructure/agents`.

## Estructura actual (rebuild)

```
main.go                          # entrada fina: delega a internal/tui
cmd/huginn/main.go               # entrada fina: delega a internal/tui
cmd/huginn-bot/main.go           # bot: 127.0.0.1:8765 /health /status /chat
internal/tui/app.go              # Model minimalista — homescreen centrada
internal/tui/main.go             # ParseArgs → ResolveContext → Run
internal/tui/styles/tokens.go    # Design system: Bg #130E0A, Accent #CD8D38, dark-first
internal/cli/cli.go              # ParseArgs, PrintHelp, ResolveContext
internal/events/bus.go           # Bus tipado: TaskCreated, AgentMessage…
internal/domain/task/status.go   # PENDING PLANNING RUNNING WAITING COMPLETED FAILED CANCELLED
internal/domain/agent/           # Agent, Capability (coding/research…), Provider
internal/domain/{plan,execution,project,vault,memory,session,skill}
internal/application/{planner,router,evaluator,orchestrator/pipeline,agents}
internal/infrastructure/{agents,models,config,memory,vault,security,…}
install/{install.sh,install.ps1}
packaging/{npm,brew,aur}   landing/ (Astro)
```

TUI: homescreen con `H  U  G  I  N  N` centrado, input `> Ask Huginn…` y `Type / for commands`. Comandos `/help /agents /tasks…` responden con mensajes honestos — nunca fingen éxito. Navegación futura (Dashboard/Agents/Tasks…) se añadirá progresivamente.

## Por qué esta estructura

- **CLI fina** — solo parsing. FutureSubcommands preparados sin lógica de negocio.
- **Domain puro** — `task` con 7 estados + `IsTerminal/IsActive`, `project` sin UI, testeable.
- **Events desacoplado** — `internal/events` sin deps, para logs, UI real-time y observabilidad.
- **TUI separada** — `tui/app.go` no importa `infrastructure`; el orquestador es intercambiable.
- **Agent desacoplado** — `Agent {ID, Name, Capabilities, Execute()}` + routing por `Capability`, jamás `if agent == "opencode"`.

## Uso

```bash
go vet ./... && go test ./... -count=1 && go build ./...
go run . --help
go run . --dump-ansi   # dump del homescreen
```
