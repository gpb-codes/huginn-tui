# AGENTS.md — Huginn TUI (instrucciones persistentes para agentes de desarrollo)

## Principio

**Huginn piensa y coordina. Los agentes ejecutan.** Nunca conviertas Huginn en
un wrapper de OpenCode/Ollama, ni en un chatbot con respuestas inventadas.

## Reglas duras

1. **Sin éxitos falsos.** Si no hay provider/runtime, retorna error honesto.
   Prohibido: `Tests: 42`, `"github ok"`, `"hermes ok"`, respuestas mock en
   producción. `MockProvider` solo en tests y `HUGINN_USE_MOCKS=1`.
2. **Planner ≠ runtime.** `application/planner` y `application/router` no
   importan `infrastructure/agents` ni `infrastructure/models`. Routing por
   `agent.Capability`, jamás `if agent == "opencode"`.
3. **TUI no ejecuta.** Presentación renderiza y envía comandos. Todo `exec`
   vive en `infrastructure/agents`. Nuevos comandos de chat van por
   `Registry` o `Pipeline`, nunca con `exec` directo en root.
4. **Acotar todo loop.** Retry (`RetryPolicy`), replan (`MaxReplans`),
   timeouts por tarea. Un test debe demostrar cada cota.
5. **Secretos.** Nunca loguear `API_KEY/TOKEN/PASSWORD/SECRET`. La traza usa
   `security.Redact`. Env mínimo a procesos hijo (`minimalEnv`).
6. **No inventar APIs de terceros.** Antes de usar un CLI/API externo,
   verificar su doc actual (OpenCode: `run --format json --dir --model
   --agent`; Ollama: `/api/tags`, `/api/generate`). Si no está verificado,
   error honesto, no flags adivinados.
7. **Migración controlada.** Entender → auditar → diseñar → limpiar →
   refactorizar → implementar → testear → documentar. No rewrites.
8. **Sin complejidad artificial.** Local-first: stdlib antes que dependencias,
   ficheros antes que DBs. Cada dependencia nueva debe justificarse.

## Dónde va cada cosa

| Necesito… | Va en… |
|---|---|
| Qué hacer / descomponer / DAG | `application/planner` (+ `domain/plan`) |
| Quién lo hace | `application/router` (nuevo runtime = implementar `router.Agent`) |
| Ejecutar (CLI/API real) | `infrastructure/agents`, `infrastructure/models` |
| Juzgar / reintentar / replanificar | `application/evaluator`, `orchestrator.Pipeline` |
| Permisos / secretos | `infrastructure/security` |
| MCP | `infrastructure/mcp` (allowlist mínima vía `ForCapability`) |
| Memoria | `infrastructure/memory` tras interfaces de dominio |
| Estado UI | root TUI (`model.go`/`update.go`/`views.go`) solo mensajes |

## Verificación obligatoria

```bash
go vet ./... && go test ./... -count=1 && go build ./...
```

Actualizar `docs/architecture/audit.md` (hallazgos), `docs/agents.md`
(runtimes) y este archivo cuando cambie la arquitectura.
