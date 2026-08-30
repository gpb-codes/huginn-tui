# Huginn — Clean Architecture (Hexagonal)

```
                HUGINN
                  │
        ┌─────────┴─────────┐
        │                   │
      CLI                 UI (TUI)
        │                   │
        └─────────┬─────────┘
                  │
            Orchestrator (application)
                  │
        ┌─────────┼─────────┐
        │         │         │
     Planner    Coder    Researcher  (domain)
        │         │         │
        └─────────┼─────────┘
                  │
             Agent Vault (ports)
                  │
        ┌─────────┼─────────┐
        │         │         │
     Memory     Files    Knowledge  (infrastructure)
```

## Capas y regla de dependencia

```
cmd/huginn       →  internal/cli  →  internal/application  →  internal/domain
internal/tui     →  internal/application  →  internal/domain  ←  internal/infrastructure
```

`domain` nunca importa `lipgloss`, `os` solo donde es necesario, ni `infrastructure`. `application/ports` define interfaces (`VaultPort`, `MemoryPort`, `ToolPort` en `internal/application/ports/ports.go:1`), `infrastructure` las implementa.

## Estructura actual (implementada)

```
cmd/huginn/main.go                  # wiring fino, solo parseArgs → ResolveContext → Run
internal/
  cli/cli.go                        # ParseArgs, PrintHelp, HuginnError, ResolveContext (delega a domain)
  domain/
    project/detector.go             # IsDirectory, DetectPackageManager, DetectProject
    agent/agent.go                  # Agent, Status, BackendAgent, CommandAvailable
    vault/resolver.go               # ResolveVaultPath (env > ~/agent-vault)
    memory/memory.go                # Entry (domain, sin persistencia)
  application/
    ports/ports.go                  # VaultPort, MemoryPort, ToolPort
    usecases/                       # (reservado) AnalyzeProject, Chat
  infrastructure/
    config/settings.go              # Sections, Values (⚙ Settings)
    mcp/servers.go                  # MCP servers (stdio/sse/ws)
    lsp/servers.go                  # LSP servers
    peers/peers.go                  # Peers, LocalPeerID/Hostname
  tui/                              # (en migración) model, views — actualmente en root main.go:460
```

Legacy `main.go:1` (2800 líneas) sigue funcionando (`go run .`) pero es **strangler**: la nueva entrada limpia es `go run ./cmd/huginn` o `go build -o huginn ./cmd/huginn`. Migración incremental sin reescribir todo de golpe.

## Por qué esta estructura

- **CLI fina** (`internal/cli`) — solo parsing y detección de contexto, sin lógica de negocio. Preparada para `huginn chat/agent/memory/vault/config` (`cli.go:22` FutureSubcommands).
- **Domain puro** — `project` y `vault` no dependen de `bubbletea`; testeables (`cli_test.go` usa `project.IsDirectory`).
- **Ports & Adapters** — Huginn no duplica Agent Vault; lo consume vía `ports.VaultPort`. Cambiar de filesystem a API no toca domain.
- **Infraestructura reemplazable** — MCP/LSP/Peers son adaptadores; se pueden mockear en tests.

## Próximos pasos (sin romper)

1. Mover `main.go:460` (model) + `View:1587` + `viewServers:1926` a `internal/tui/` (ya existe el directorio).
2. Extraer `orchestrator` (`advance:1556`) a `internal/application/usecases/orchestrator.go`.
3. Hacer que `cmd/huginn/main.go` llame a `tui.Run(ctx)` en vez de imprimir stub.
4. Añadir `internal/infrastructure/config` con `~/.huginn/config.json` persistente.

## Uso

```bash
go vet ./... && go test ./...
go run ./cmd/huginn --help
go run ./cmd/huginn --version
go run ./cmd/huginn . "analiza este proyecto"
go build -o huginn ./cmd/huginn && ./huginn --help  # Windows: huginn.exe
```
