# Huginn — Testing

```bash
go vet ./...
go test ./... -count=1
go build ./...
```

## Cobertura por capa

| Paquete | Qué se prueba |
|---|---|
| `domain/plan` | `Validate` (deps desconocidas, ciclos, duplicados), `Levels` (paralelismo real) |
| `domain/agent` | `Supports`/`HasCapabilities` |
| `domain/execution` | `RequiresApproval` (keywords + flag), `NewRun` |
| `application/planner` | clasificación de intent, DAG válido, fan-out por complejidad, repro en fix, approval en devops, recovery acotado |
| `application/router` | routing por capacidad, preferencia available, fallback honesto, `ErrNoAgent`, reemplazo de duplicados |
| `application/evaluator` | veredictos, retryable vs terminal, output vacío, `RetryPolicy` acotada, backoff, approval gate |
| `application/orchestrator` | pipeline end-to-end con runtimes fake: éxito, retry→éxito, fallo acotado, approval deny, skip por dependencia fallida |
| `infrastructure/agents` | OpenCode: parse JSON, fallback raw, not-installed, workspace inválido/válido (`--dir`), timeout, crash con output parcial, flags `--model/--agent` |
| `infrastructure/models` | Ollama con `httptest`: tags ok/down, generate ok, error HTTP, error de modelo |
| `infrastructure/mcp` | allowlist mínima por capacidad, default-deny |
| `infrastructure/security` | policy read-always/default-deny/allow explícito, `Redact` sin fugas |
| `infrastructure/config` | defaults v2, migración v1→v2 preservando valores, round-trip |

## Integración con OpenCode real

```bash
HUGINN_TEST_OPENCODE=1 go test ./internal/infrastructure/agents/ -run Integration -v
```

Requiere el binario `opencode` en PATH. Hace `t.Skip` en caso contrario:
el resto de la suite nunca depende del binario.

## Reglas

- Sin tests que validen mocks como comportamiento real: `MockProvider` solo
  aparece en tests de Registry y desarrollo offline (`HUGINN_USE_MOCKS=1`).
- Todo retry/replan bajo test debe demostrar su cota (contar llamadas).
- Los errores tipados (`ErrTimeout`, `ErrNotInstalled`, …) se asertan con
  `errors.Is`, no con substrings.
