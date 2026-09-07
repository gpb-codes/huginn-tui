# Huginn TUI — Architecture Audit (2026-09-03)

Auditoría completa del repositorio antes de la migración al sistema de agentes.
Todo hallazgo referencia código real (`archivo:línea`). ~9275 líneas Go (119 archivos).

## 1. Current Architecture

Dos entradas conviven (migración "strangler" declarada en `ARCHITECTURE.md:58`):

- **Legacy root** (`main.go:1`, `cli.go`, `model.go`, `update.go:997`, `views.go:1033`,
  `dispatch.go`, `chat.go`, `domain.go`, `servers.go`, `palette.go`): TUI completa
  Bubble Tea v2 funcional. `runTUI` (`cli.go:246`) construye `bootstrap.App` y abre el modelo.
- **Clean entry** (`cmd/huginn/main.go:112`): solo imprime contexto; admite que
  "el TUI vive en la raíz" (`cmd/huginn/main.go:96`) y no abre ninguna UI.

Capas `internal/` (hexagonal, regla documentada en `ARCHITECTURE.md:27`):

| Capa | Contenido real |
|---|---|
| `domain/` | `task` (Task/Status/Graph/Result), `agent` (Agent/Provider factories), `execution` (Record), `memory`, `vault`, `project`, `context`, `session`, `skill`, `profile`, `preference`, `decision`, `onboarding` |
| `application/` | `agents` (Registry + 11 roles), `orchestrator` (Orchestrator/Planner/Scheduler/Dispatcher/Synthesizer), `ports`, `session`, `personalization`, `delegation`, `onboarding`, `cron`, `workflows`, `tools` |
| `infrastructure/` | `agents` (opencode/kilo/hermes adapters), `providers` (10 mocks), `models` (vacío), `runtime` (ProcessRuntime), `security` (Policy), `mcp/lsp/peers`, `memory`, `vault`, `config`, `git`, `filesystem`, `profile`, `skills`, `trace`, `dialog`, `voice` |
| `neural/` + `routing/` | MLP real (Tensor/Dense/ReLU/Softmax/CE/SGD/Trainer) + NeuralRouter (7 features → softmax sobre agentes). **Sin usar: nada los importa** |
| `tui/` | Componentes parciales; la TUI real sigue en root |

Flujo de chat real hoy: `update.go:125` → `callAgentCmd` (`dispatch.go:46`) →
`dispatchViaRegistry` (mock providers) o `dispatchViaExec` (`opencode run -m <model> --format json`,
`dispatch.go:142`) directo desde presentación.

## 2. Problems (verificados)

1. **Mocks como implementación real.** `infrastructure/providers/mock.go:27-66`: 10 providers
   (`chatgpt`, `claude`, `perplexity`, `opencode`, `kilo`, `gmi`, `azure`, `minimax`, `tencent`,
   `muse-spark`) devuelven `"[mock <name>] " + prompt`. `bootstrap/app.go:42-51` los registra
   como providers de producción. `chat.go:174` `mockReply` inventa respuestas por agente.
   `dispatch.go:50` responde ChatGPT con `mockReply` tras `Sleep` — finge inferencia.
2. **Lógica de negocio en la TUI (§19).** `dispatch.go:118-192` hace `exec` de OpenCode,
   parsea JSON y decide workflow desde el paquete `main` (presentación).
3. **CLI duplicada.** `cli.go:112-179` ≡ `internal/cli/cli.go:40-107` (parseo idéntico);
   `cli_test.go` solo cubre la copia root. `domain.go:13-86` ≡ `domain/agent/agent.go`
   (estados + `backendAgents` + `commandAvailable`).
4. **Config duplicada/divergente.** `config/config.go` (persistente `~/.huginn/config.json`,
   4 campos) vs `config/settings.go` (árbol UI en memoria) vs `servers.go:38-58`
   (tercera copia del árbol + valores hardcodeados con nombre de usuario real `"gabriel"`).
5. **Planificadores triviales.** `MockPlanner` (`orchestrator/planner.go:15`) genera pipeline fijo;
   `DirectPlanner.Plan` (`planner_direct.go:29`) ignora el prompt y hardcodea `"opencode"`;
   `NoopDispatcher` (`dispatcher.go:14`) marca tareas completadas sin ejecutar;
   `NoopSynthesizer` (`synthesizer.go:12`) no hace nada. `Orchestrator.Run`
   (`orchestrator.go:22`) es secuencial, sin niveles paralelos, sin evaluator, sin retry.
6. **Estados fragmentados.** Tres enums: `task.Status` (string, 7 estados, `task/status.go:6`),
   `agent.Status` (int, 4 estados, `agent/agent.go:8`), `agentStatus` root (`domain.go:14`).
   Faltan `PLANNING/READY/REVIEW/RETRYING` (§20).
7. **Datos falsos en UI.** `servers.go:80-88` afirma `Connected 12ms/34ms…` hardcodeado;
   `servers.go:130-135` inventa peers con IPs y `Syncing 64%`; `update.go:63-70`
   `simulateServers()` anima esos datos. Nada mide conexiones reales.
8. **Seguridad sin cablear.** `security/policy.go` (`Can`, `DetectSecret`) no la usa ninguna
   ruta de ejecución. `trace.Append` (`trace/trace.go:21`) persiste `Input` íntegro (hasta 800
   chars) sin sanitizar secretos. `exec.CommandContext` hereda todo el entorno del proceso.
9. **Router neuronal desconectado.** `routing/neural_router.go` entrena sobre
   `dataset.go:17` (`agentIDs` fijos `opencode/researcher/toolmaster` que no existen en el
   Registry) y `bootstrap/app.go` cablea `DirectPlanner` en su lugar.
10. **Puertos huérfanos.** `ports/agent_runtime.go` (`AgentRuntime.Run`) y
    `runtime/process_runtime.go` no los usa ningún dispatcher. `ports/git.go`, `ports/tool.go`,
    `ports/memory_port.go` sin implementaciones cableadas al flujo principal.

## 3. Dead Code (REMOVE)

| Elemento | Evidencia | Por qué |
|---|---|---|
| `application/cron/dream.go` | 13 líneas, 0 imports, 0 lógica | Tipos con solo `Interval`; sin scheduler que los use |
| `application/workflows/compose.go` | 7 líneas, 0 imports | 3 slices de strings; "JS sandbox" inexistente |
| `application/tools/search.go` | 4 líneas | 1 const sin uso |
| `infrastructure/voice/voice.go` | 7 líneas | 1 struct sin uso; "TenVAD/ASR" no existe en el repo |
| `infrastructure/memory/events.go` | 5 líneas, solo comentario | Sin código |
| `tui/components/{agents,chat,tasks}.go`, `tui/views/{chat,memory}.go` | 2 líneas c/u | Stubs vacíos |
| `tui/{app,styles}.go` | 8/7 líneas | Stubs; estilos reales en `tui/styles/tokens.go` |
| `domain.go` (root) | duplicado exacto de `domain/agent` | La TUI debe usar `domain/agent` |
| `mockReply` (`chat.go:174`) | respuestas inventadas | Sustituir por error honesto si no hay provider |
| Providers mock cloud (`mock.go:27-66`) del path de producción | fingen inferencia | Solo `MockProvider` genérico queda, para tests |
| `NoopDispatcher`, `MockPlanner` del cableado prod | fingen ejecución | Reemplazados por Pipeline real |

## 4. Redundant Components

- `cli.go` root vs `internal/cli/cli.go` → root delega en `internal/cli`.
- `huginnSettings`/`settingsValues` (`servers.go`) vs `config.Sections/Values` → la TUI lee de `config`.
- `mcpServer/lspServer/peerServer` (root) vs `infrastructure/{mcp,lsp,peers}` → la TUI usa infra.
- `VERSION` en `palette.go:5` y `cmd/huginn/main.go:11` → una sola fuente en `internal/cli`.
- `MockPlanner` vs `DirectPlanner` vs `NeuralRouter` → un solo `planner.NeuralPlanner`.

## 5. Architecture Risks

- **Doble entrypoint**: `go run .` (TUI) vs `go run ./cmd/huginn` (stub) confunde; `huginn.exe`
  versionado en el repo (binario de ~MB en git — ver `.gitignore`).
- `update.go` + `views.go` (~2000 líneas) concentran toda la UI: cualquier cambio de dominio
  exige tocar presentación. Mitigación: no reescribir; exponer `Pipeline` y consumir por mensajes.
- `CodingAgent.Execute` (`agents/coding.go:17`) reporta éxito con `FilesChanged: ["main.go"]`
  inventado — éxito falso. Se reescribe para delegar en runtime real.
- Hermes `Execute` (`hermes_adapter.go:43`) retorna `"hermes ok"` sin ejecutar nada — éxito falso.
- Model IDs `opencode/mimo-v2.5-free` etc. (`dispatch.go:131-141`) no verificados contra
  catálogo real; se mueven a configuración explícita (`agents.yaml`-like en `Config`).

## 6. Agent System Problems

- Dos dispatchers (`Noop`, `RegistryDispatcher`) + `dispatch.go` en TUI: tres caminos que
  no comparten política, timeouts ni traza homogénea.
- `Registry.ResolveAgent` matchea por string de tipo (`CanHandle`), no por capacidades (§9).
- Role agents con `Providers() nil` (planner, reviewer, qa…) fallan en `SelectProvider`
  ("sin candidates") si se despachan — solo funcionan los que tienen providers mock.
- Adapters reales (`opencode/kilo/hermes`) existen pero **bootstrap no los usa**.
- Sin lifecycle (cancel/timeout por tarea), sin permisos por runtime, sin aprobación humana.

## 7. Neural Engine Problems

- La red (`neural/`) es un MLP real y correcto, pero su única tarea es clasificar
  `researcher/opencode/toolmaster` con `HistoricalSuccess: 0.5` constante
  (`features.go:48`) y dataset sintético de 15 muestras (`dataset.go:18-36`).
- Llamarlo "neural engine" hoy es excesivo (§22): es un **clasificador heurístico + MLP
  de routing**. Decisión honesta: conservar MLP como *router de capacidades*,
  nombrar el resto por lo que es (`planner` heurístico + inferencia vía provider),
  y exponer `Trajectory/Reward` para aprendizaje futuro sin fingirlo.
- `weightsKey` (`model.go:185`) colisiona con >10 capas (`'0'+i` sale del dígito) — bug real
  si el modelo crece. Se corrige con formato numérico.

## 8. Recommended Architecture

```text
USER → TUI (mensajes) → Pipeline (application)
  → NeuralPlanner (routing.NeuralRouter + heurística) → Plan (domain/plan, DAG con Levels)
  → AgentRouter (por Capability, sin nombres hardcodeados)
  → Runtimes (infrastructure/agents: opencode real; models: ollama real; kilo/hermes opcionales)
  → Tools/MCP (allowlist mínima por tarea) → Ejecución (timeouts, cancel, permisos)
  → Evaluator (éxito? retry? replan? aprobación humana) → Run trazable (trace+memory)
```

- `Huginn piensa y coordina; los agentes ejecutan.` Planner nunca importa adapters.
- Chat conversacional (`Registry`) y ejecución planificada (`Pipeline`) comparten providers
  reales; los mocks solo viven en tests (`HUGINN_USE_MOCKS=1` para desarrollo offline).

## 9. Migration Plan

1. Audit (este doc) → 2. Cleanup (REMOVE tabla §3) → 3. Dominio (Capability, Plan, Status,
   Run/Approval, campos `Capability/NeedsApproval` en Task) → 4. Application (planner, router,
   evaluator, pipeline) → 5. Infra (opencode hardening, ollama real, opencode provider,
   mcp filter, trace redaction, config v2) → 6. Bootstrap rewire → 7. TUI (`/plan`,
   dispatch sin exec directo) → 8. Tests → 9. Docs → 10. Final audit.

## 10. Keep / Refactor / Replace / Remove / New

| Elemento | Decisión |
|---|---|
| `neural/` MLP | KEEP (+fix `weightsKey`) |
| `routing/` router+features+dataset | KEEP (cablear al planner; alinear IDs con capacidades) |
| `domain/task` Graph/TopologicalOrder | KEEP (+estados y campos) |
| `domain/agent` Agent/Provider | KEEP (+Capability) |
| `agents.Registry` | KEEP (+routing por capacidad) |
| Role agents (research/coding/qa…) | REFACTOR (resultados reales, sin éxito inventado) |
| `RegistryDispatcher`/`Orchestrator` | KEEP compat (fuera del path `/plan`) |
| `security.Policy`, `trace.Store`, `ProcessRuntime` | KEEP (cablear) |
| `memory`/`vault`/`config`/`context`/`profile`/`personalization`/`session`/`skills`/`git`/`filesystem`/`peers`/`lsp`/`dialog`/`delegation`/`onboarding` | KEEP |
| Root TUI (`model/update/views/chat/servers/palette`) | KEEP (sin rewrite; delegar ejecución) |
| `mcp.DefaultServers` | KEEP (+allowlist por tarea) |
| `MockPlanner`/`DirectPlanner.Plan`/`NoopDispatcher`/`NoopSynthesizer` en prod | REPLACE (Pipeline) |
| Cloud mock providers en prod | REPLACE (opencode+ollama reales) |
| `dispatchViaExec` en `main` | REPLACE (via `infrastructure/agents`) |
| `hermes.Execute` éxito falso | REPLACE (error honesto hasta implementar) |
| Stubs §3, `domain.go`, `mockReply`, cloud mocks prod | REMOVE |
| Capability, Plan/DAG levels, Run/Approval, NeuralPlanner, AgentRouter, Evaluator+Retry, Pipeline, Ollama provider, OpenCode provider, MCP filter, config v2, tests, docs, `.opencode/agents/huginn-coder.md` | NEW |

## 11. Agent Evaluation (§12)

| Agente | Estado 2026 | Decisión |
|---|---|---|
| OpenCode (CLI `run --format json --dir --model --agent`, verificado en opencode.ai/docs/cli) | Mantenido, CLI scriptable | **PRIMARY** — coding runtime |
| Ollama (`/api/tags`, `/api/generate` HTTP local) | Mantenido, local-first | **PRIMARY** — model provider (research/reasoning/docs/review) |
| Kilo Code | Mantenido (VSCode+Roo lineage); CLI `kilo` no verificada | OPTIONAL — solo si binario presente; mismo CLI que opencode hoy es supuesto sin verificar → se detecta y degrada con honestidad |
| Hermes (Nous Research) | Binario `hermes` nicho; sin CLI verificada | FUTURE — adapter conservado, ejecución retorna error honesto |
| Goose (Block) | Mantenido, `goose run` CLI+MCP | FUTURE — interfaz `router.Agent` lista para agregarlo sin tocar planner |
| OpenHands | Mantenido pero pesado (Docker, cloud-leaning) | FUTURE — no local-first suficiente hoy |
| Aider | Mantenido, CLI `aider` chat-por-terminal | FUTURE — buen coder alternativo |
