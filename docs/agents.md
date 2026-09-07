# Huginn — Agent System

## Abstracción

```go
// application/router — implementado por infrastructure/*
type Agent interface {
    ID() string
    DisplayName() string
    Capabilities() []agent.Capability // coding, research, reasoning, testing,
                                      // review, devops, documentation,
                                      // browser, filesystem, git
    Available(ctx context.Context) bool
    Execute(ctx context.Context, t task.Task) (task.Result, error)
}
```

El planner declara `task.Capability`; el router elige el runtime disponible con
mejor overlap. Agregar un agente nuevo = implementar esta interfaz y
registrarlo en `bootstrap` — el planner no cambia.

## Runtimes

| Runtime | Capacidades | Estado | Notas |
|---|---|---|---|
| `opencode` (`infrastructure/agents`) | coding, testing, devops, filesystem, git | **PRIMARY** | CLI real: `opencode run --format json --dir --model --agent`. Workspace validado, env mínimo, timeouts, errores tipados (`ErrNotInstalled`, `ErrInvalidWorkspace`, `ErrTimeout`, `ErrProcessCrash`). Subagente de proyecto: `.opencode/agents/huginn-coder.md`. |
| `ollama` (`infrastructure/models`) | research, reasoning, documentation, review | **PRIMARY** | HTTP local (`/api/tags`, `/api/generate`). Config: `models.ollama_host`, `models.ollama_model` (`OLLAMA_HOST`/`OLLAMA_MODEL` como override) + matriz `models.ollama_models` por capacidad (`coding→qwen3`, `research/reasoning/documentation/review/chat→llama3.1`; lo no listado usa `ollama_model`). Precedencia por request: `Meta["model"]` explícito > matriz[capacidad] > default. |
| `kilo` | coding, testing, filesystem, git | OPTIONAL | Solo si el binario está presente. Sin contrato CLI verificado: ejecución deshabilitada con error honesto. |
| `hermes` | — | FUTURE | Solo detección. `Execute` retorna error honesto hasta verificar su CLI. |
| goose / openhands / aider | — | FUTURE | La interfaz `router.Agent` está lista; ver `docs/architecture/audit.md §11`. |

## OpenCode como worker (no como cerebro)

```text
Huginn → NeuralPlanner → Task Graph → AgentRouter → OpenCodeAdapter
  → opencode run --format json [--model …] [--agent …] [--dir workspace]
  → eventos JSON parseados → task.Result → Evaluator → (replan)
```

- Planificación, routing, permisos, retry y memoria viven en Huginn.
- OpenCode nunca decide el workflow; recibe una tarea y un workspace.
- MCP: allowlist mínima por capacidad (`infrastructure/mcp.ForCapability`),
  no todos los servidores a la vez (los servidores MCP consumen contexto).

## Chat conversacional

`@opencode/@chatgpt/@kilo/@mimo/@muse` → `Registry` → provider real
(`opencode`→CLI, resto→ollama). Sin provider en línea: mensaje honesto,
nunca respuesta inventada. `HUGINN_USE_MOCKS=1` habilita doubles para
desarrollo offline.

## Permisos y aprobación

- `security.Policy`: `read` siempre permitido; `write` solo para runtimes en
  `permissions.allowed_write_agents` (default: `opencode`).
- Tasks con `NeedsApproval` o keywords sensibles (`git push`, `rm -rf`,
  `migration`, `production`, …) pausan el pipeline hasta aprobación humana.
  Headless default: denegar. `permissions.auto_approve: true` para permitir.
- Traza (`executions.jsonl`) redacta secretos (`security.Redact`) y trunca inputs.
