# Huginn — Target Architecture (Agent Orchestrator)

`Huginn piensa y coordina. Los agentes ejecutan.`

```text
                    USER
                      │
                      ▼
                   HUGINN TUI  (/plan <objetivo>, chat, status)
                      │
                      ▼
               NEURAL ENGINE (application/planner)
                Analyze → Classify → Complexity/Risk → Decompose → DAG
                      │
                      ▼
                  PLAN (domain/plan: Tasks + Levels paralelos)
                      │
                      ▼
                AGENT ROUTER (por Capability, sin nombres hardcodeados)
                      │
        ┌─────────────┼─────────────┐
        ▼             ▼             ▼
     OpenCode       Ollama        Kilo (opt)
     coding/       research/      coding
     testing/      reasoning/     (solo si
     devops/       docs/review    binario)
     git/fs
        │             │             │
        └─────────────┼─────────────┘
                      ▼
               TOOLS / MCP (allowlist mínima por capacidad)
                      │
                      ▼
                  EXECUTION (timeout, cancel, permisos, approval)
                      │
                      ▼
                  EVALUATOR → SUCCESS → DONE
                      │
                      └─────→ FAIL → retry (acotado) → replan (1 ronda) → DONE/FAILED
```

## Capas

- `domain/` — `agent` (Capability, Provider), `task` (Task/Status/Graph/Result),
  `plan` (Plan/Levels/Validate), `execution` (Run/Approval), `memory`, `vault`, …
  Sin dependencias de UI ni infraestructura.
- `application/` — `planner` (NeuralPlanner: heurística + `routing.NeuralRouter`),
  `router` (selección por capacidad), `orchestrator` (Pipeline + legacy),
  `evaluator` (veredicto + RetryPolicy + approval gate), `agents` (Registry y roles
  para el path conversacional).
- `infrastructure/` — `agents` (OpenCode/Kilo adapters + runtimes),
  `models` (Ollama real), `providers` (OpenCode real + Mock solo tests),
  `mcp` (allowlist por capacidad), `security` (Policy + Redact),
  `trace` (executions.jsonl sanitizado), `memory`, `vault`, `config` (v2), …
- `presentation/` — TUI Bubble Tea (root) + `internal/tui` en migración.
  La TUI renderiza estado y envía comandos; nunca hace `exec` de agentes
  (el único `exec` vive en `infrastructure/agents`).

## Dos caminos, un cerebro compartido

| Camino | Entrada | Flujo |
|---|---|---|
| Conversacional | chat `@agent` | Registry → Provider real (opencode/ollama) → traza |
| Planificado | `/plan <objetivo>` | Planner → Plan DAG → Router → runtimes → Evaluator → retry/replan |

Ambos comparten providers reales. Los mocks solo existen con `HUGINN_USE_MOCKS=1`.

## Estados (`task.Status`)

`planning → pending → ready → queued → running → (review) → completed`
`running → retrying → running…` · `failed` · `cancelled` · `waiting`

## Límites anti-loop

- `RetryPolicy.MaxAttempts` (default 2, override por tarea).
- `Pipeline.MaxReplans` (default 1 ronda de recovery).
- `TaskTimeout` por ejecución (default 90s), cancelación por contexto.
- Approval gate: `DenyApprover` por defecto; `AllowApprover` solo con
  `permissions.auto_approve: true`.

## Aprendizaje (honesto)

Hoy: heurística + MLP clasificador (`neural/` + `routing/`), trayectorias
persistidas (`Run` + `executions.jsonl`). Sin RL fingido. La interfaz
`Trajectory → Evaluation → Reward` se añadirá cuando haya datos reales.
