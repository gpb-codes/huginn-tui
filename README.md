# Huginn — CLI + TUI para Agent Vault

Huginn es la **interfaz y capa de interacción de Agent Vault** — inspirado en OpenCode: comando global `huginn`, rápido y limpio.

```
                HUGINN
                  │
        ┌─────────┴─────────┐
        │                   │
      CLI                 UI (TUI)
        │                   │
        └─────────┬─────────┘
                  │
            Orchestrator
                  │
        ┌─────────┼─────────┐
        │         │         │
     Planner    Coder    Researcher
        │         │         │
        └─────────┼─────────┘
                  │
             Agent Vault
                  │
        ┌─────────┼─────────┐
        │         │         │
     Memory     Files    Knowledge
```

**Huginn (CLI/TUI)** = interacción, chat, orquestación, agentes, herramientas, contexto.  
**Agent Vault** = memoria persistente, conocimiento, embeddings, knowledge graph.

> Vikingpunk developer aesthetic — Go + Bubble Tea + Lipgloss, oscuro, morado/cyan, 100% terminal.

---

## Objetivo actual

Un ejecutable global `huginn` que use el directorio actual como contexto, detecte el proyecto, conecte con Agent Vault y abra la TUI o ejecute un prompt directo, sin que el usuario conozca la arquitectura interna.

## Funcionalidades implementadas

**CLI oficial** (`main.go:2428`, `internal/cli/cli.go:1`):
- `huginn` / `huginn .` / `huginn <path>` / `huginn "<prompt>"` / `huginn <path> "<prompt>"` / `huginn --help` / `huginn --version` / `huginn --debug`
- Subcomandos preparados: `chat`, `agent`, `agents`, `memory`, `vault`, `graph`, `config` (sin reescribir CLI)
- Detección automática: proyecto (`go.mod/package.json/.git`), package manager (`bun>pnpm>yarn>npm`), Vault (`HUGINN_VAULT` > `~/agent-vault`)
- Errores sin stack trace; `--debug`/`HUGINN_DEBUG=1` muestra detalle; nunca imprime secrets

**TUI** (`main.go:1587`):
- Dashboard `viewChat` fiel al mock: header `HUGINN • AI ORCHESTRATOR • ● CONNECTED`, split `63/22` con `AGENT DISPATCH` (ChatGPT/OpenCode/Kilo/Mimo/Muse) y `BACKGROUND TASKS` vs `AGENTS` + `TASK PIPELINE` + `CONTEXT` + input `> _` + footer `TAB/CTRL+P/CTRL+B/CTRL+G/CTRL+L`
- Chat con `@chatgpt @opencode @kilo @mimo @muse @all` + autocomplete `Tab/Enter` y `parseMentions`
- Orquestación simulada `advance:1556` (4 agentes secuencial 180ms), logs en vivo, `Esc` cancela, `Ctrl+C` sale
- `⚙ Settings` (`viewSettings:2298`): árbol 8 secciones / 36 ajustes (General/Appearance/AI/Agents/Memory/Tools/Vault/Advanced) con `↑↓/←→/enter`, toggle `On↔Off`
- `SERVERS` (`viewServers:2043`): tabs `MCP (7) / LSP (5) / Peers (4)` con `mcpServers:198` (stdio/sse/ws), `lspServers:219` (gopls/tsserver/pyright...), `peerServers:256` (Online/Syncing/Offline/Pairing) + `vault-sync` + `tickServers:1552`
- `KNOWLEDGE GRAPH` (`viewGraph:2298`): comando `graph`/`/graph` (o `huginn graph` CLI + `Ctrl+G`), 12 nodos / 23 edges, `● Agent → ◉ Memory/● Context/○ Project ...`
- `HUGINN` wordmark morado/blanco con cuervos `𓅃`

**Infraestructura**:
- `install.ps1:1` para Windows (PowerShell/CMD/WT/VS Code), `cmd/huginn/main.go:1` entry limpia
- Tests `cli_test.go:1` (12 tests, `go test` OK), `go vet` OK, `go build` OK

## Estado actual del desarrollo

Proyecto funcional en estado **MVP + CLI oficial**. La TUI abre, navega, orquesta (simulado), gestiona settings/servers/graph y el CLI resuelve contexto real. No hay integración real con Agent Vault más allá de `resolveVaultPath` y `openVault` (explorer); embeddings/graph son mock. No hay persistencia de config ni red P2P real (peers simulados).

## Arquitectura actual

Clean + Hexagonal (monolito modular, migración incremental):

```
cmd/huginn/main.go              # wiring fino
internal/cli/cli.go             # ParseArgs, PrintHelp
internal/domain/project/        # IsDirectory, DetectPackageManager, DetectProject
internal/domain/agent/          # Agent, Status, BackendAgent
internal/domain/vault/          # ResolveVaultPath
internal/domain/memory/         # Entry
internal/application/ports/     # VaultPort, MemoryPort, ToolPort
internal/infrastructure/config/ # Sections/Values
internal/infrastructure/mcp/lsp/peers # servidores mock
internal/tui/                   # (en migración) — hoy aún en main.go
main.go                         # 2800 líneas God file (legacy, será partido)
```

Regla: `tui/cli → application → domain ← infrastructure`. `ARCHITECTURE.md:1` detalla flujo y próximos pasos. `go.mod:1` solo `bubbletea/v2` + `lipgloss/v2` (sin cobra/viper aún).

## Tecnologías

- **Go 1.25** — `go.mod:1`
- **Charm Bubble Tea v2** + **Lipgloss v2** — TUI, `tea.NewProgram` `main.go:2800`
- **Lipgloss** — bordes, colores `#111317/#33d9f2/#9061f9/#2fd67a`
- **os/exec** — `opencode`/`kilocode` reales si están instalados (`commandAvailable:90`)
- **Pillow (Python)** — generación de `screenshot.png` (solo docs)

## Cómo ejecutar

```bash
cd huginn-tui
go mod tidy
go vet ./... && go test ./... -count=1
go run .                 # = huginn
go run . -- --help
go run . -- --version
go run . -- "analiza este proyecto"
go run . -- ./mi-proyecto "crea auth"
go run ./cmd/huginn --help
go run ./cmd/huginn graph

# build global Windows
go build -o huginn.exe .
Move-Item huginn.exe $env:USERPROFILE\go\bin\huginn.exe
huginn --help
huginn
huginn C:\Projects\mi-proyecto
huginn "analiza este proyecto"

# o script
powershell -ExecutionPolicy Bypass -File install.ps1
```

Dentro de la TUI: `graph`/`/graph` → grafo, `TAB` agente, `Ctrl+P` commands, `Ctrl+G` graph, `Ctrl+L` clear, `Esc` volver, `q` salir.

## Estructura relevante

```
huginn-tui/
├── cmd/huginn/main.go          # CLI limpia
├── internal/
│   ├── cli/cli.go
│   ├── domain/{project,agent,vault,memory}
│   ├── application/ports/
│   ├── infrastructure/{config,mcp,lsp,peers}
│   └── tui/doc.go
├── main.go                     # God file (a partir en tui/*)
├── cli_test.go                 # tests CLI
├── go.mod / go.sum
├── install.ps1
├── screenshot.png              # 1920×1080 Vikingpunk
├── ARCHITECTURE.md
└── .gitignore
```

## Qué se implementó durante la jornada

1. **Restaurado diseño HUGINN** y fix tabs `viewChat:2356` (RoundedBorder → PlaceHorizontal, sin `─────╮`)
2. **Menú ⚙ Settings** (`huginnSettings:167`, `viewSettings:2298`) — 8 secciones/36 items con `↑↓←→/enter`
3. **Sistema MCP/LSP/Peers** (`mcpServers:198`, `lspServers:219`, `peerServers:256`, `viewServers:2043`, `tickServers:1552`) — tabs, toggle, pairing `a`, ping `r`
4. **CLI oficial** (`main.go:2428`, `internal/cli:1`) — `huginn`, `<path>`, `"<prompt>"`, `--help/--version/--debug`, subcomandos preparados, detección Windows/Unix, `resolveVaultPath` sin duplicar Vault
5. **Clean Architecture** (`cmd/`, `internal/`, `ARCHITECTURE.md`) — Ports & Adapters, `go vet/test/build` OK
6. **Screenshot producción** `screenshot.png` (Python/Pillow, 1920×1080, Vikingpunk)
7. **Comando `graph`** (`modeGraph:287`, `viewGraph:2298`, `handleCommand:645`, `runGraphTUI:2815`) — `graph`/`/graph`/`huginn graph` + `Ctrl+G`
8. **Chat dashboard** (`viewChat:2356` nuevo) — layout fiel al mock `HUGINN • AI ORCHESTRATOR • ● CONNECTED` + split 63/22 + `AGENT DISPATCH` + `BACKGROUND TASKS` + `CONTEXT` + input `> _`
9. **Fixes** `colWarn/colError:22`, `agentStatus:2390` (`color.Color`), `isDirectory` Windows, `.gitignore:3` no ignorar `cmd/huginn`

## Qué queda pendiente / deuda técnica

- `main.go` God file (2800 líneas) aún concentra TUI + CLI + domain; terminar migración a `internal/tui/*.go` y `internal/application/usecases/`
- Orquestación real: `advance` simula 20% cada 180ms; falta `os/exec` streaming por agente y contrato tarea→subtareas JSON
- Agent Vault: solo `resolveVaultPath`/`openVault`; falta `VaultPort` real (search/index/embeddings) y `MemoryPort` persistente
- Peers: sync es mock (`Syncing 64% → Synced`); falta mDNS/WebSocket real y `vault-sync` MCP
- Sin `cobra`/`viper` aún; `huginn.json` no persiste; `install.ps1` no firma binario
- `screenshot.png` en repo (79KB) — considerar `docs/` o LFS si crece

## Próximos pasos

### A. Migración/refactorización hacia arquitectura hexagonal

Analizar `main.go` y preparar migración progresiva hacia hexagonal. Separar claramente `Domain` (project/agent/memory/vault), `Application` (usecases `AnalyzeProject`, `Chat`), `Infrastructure` (filesystem, mcp, lsp, peers, config), `Ports` (`VaultPort`, `MemoryPort`, `ToolPort`), `Adapters` (stdio/sse/ws, `gopls`). Identificar lógica acoplada a infraestructura (`os.Stat` en `detectProject`, `exec.LookPath` en `agent.go`), invertir dependencias via interfaces (`application/ports`), convertir servicios en casos de uso, aislar `os/exec` y `Vault` externo. Migración incremental, no romper `go test`/`go vet`.

### B. Refactorización general

Revisar nombres (`mcpServers` vs `MCPRepository`), responsabilidades (model 500 líneas), duplicación (`parseArgs` en `main.go` y `internal/cli`), funciones largas (`viewChat` 120 líneas, `Update` 400 líneas), acoplamiento `viewSettings`↔`settingsValues` global, manejo de errores (`huginnError` solo para Vault), validaciones (`isDirectory` Windows `C:\`), config/env/logging.

### C. Implementar scroll y navegación

Revisar scroll vertical/horizontal en `viewChat` (26 filas fijas), `viewSettings` (18 visibles, `settingsOffset`), `viewServers` (12), `viewGraph` (16), `chatBox` (12). Contenedores dinámicos (`chatHistory`, `logs`), autoscroll al recibir `agentReplyMsg`, `pgup/pgdn`, evitar que `input` quede oculto cuando crece `dispatchBox`, posicionamiento al inyectar `prompt` inicial.

### D. Mejorar la experiencia de usuario

Detectar estados de carga (`Connecting` → spinner), vacíos (`No sessions` en `/sessions`), error (`Needs auth` con `Missing NOTION_API_KEY`), feedback (`serversMsg`), progreso (`pct` → barra), mensajes (`modeMessage` genérico), navegación (`TAB` vs `1-6` inconsistente), legibilidad (contraste `colMuted`/`colPanel`), consistencia (header `HUGINN` vs `AGENT VAULT`).

### E. Manejo de errores y resiliencia

Qué pasa si: `opencode` no instalado (`commandAvailable` → `not installed` pero `callAgentCmd` sigue intentando), Vault no responde (`resolveVaultPath` → `not found` pero TUI sigue), `HUGINN_VAULT` inválido, `gopls` `Error`, `peer` `Offline`, prompt vacío, path con espacios `C:\My Project`, `tick` timeout 90s. Implementar retries, timeouts, mensajes `huginnError` consistentes, no panic en `Update`.

### F. Tests

Aumentar cobertura (hoy solo `cli_test.go` 12 tests). Priorizar: 1) `domain/project` (IsDirectory/Detect*), 2) `application/usecases` (orquestador), 3) `infrastructure/mcp` (Restart), 4) `tui` (View render), 5) integración `huginn --help` exit codes. Los tests deben permitir migrar sin romper `huginn "analiza"`.

### G. Documentación técnica

Actualizar `ARCHITECTURE.md` con flujo real `os.Args → cli.ParseArgs → cli.ResolveContext → tui.Run → orchestrator → VaultPort → Agent`, detallar `Domain` (agentes, project), `Application` (usecases), `Infrastructure` (mcp/sse, lsp/stdio, peers/ws), `Ports/Adapters`, dependencias, cómo añadir `huginn memory search` (nuevo port + adapter + usecase + view).

### H. Limpieza del proyecto

Revisar: `assets` no usado (¿mover a `internal/tui/assets`?), `cuervos.svg` (solo ref), `screenshot*.png` (¿a `docs/`?), `install.ps1` duplicado con `go install`, `cli_test.go` root (¿mover a `internal/cli/cli_test.go`?), imports no usados, comentarios desactualizados (`// chat con todos los agentes`), TODOs, `huginn.exe` ignorado pero `huginn-clean.exe` no.

### I. Preparación para futuras funcionalidades

Dejar preparada la estructura para: nuevos módulos (`internal/domain/task`), agentes (`Researcher` real), proveedores IA (`opencode/mimo` ya, añadir `claude`), interfaces (`huginn server` HTTP), adaptadores (`sqlite` para Vault), persistencia intercambiable (`MemoryPort` → `badger`/`postgres`), APIs externas (`github` MCP real), automatizaciones, tests independientes, ejecución local y prod (`HUGINN_ENV`).

No implementar estas funcionalidades mañana salvo que sean necesarias para el refactor. Objetivo: arquitectura preparada.

---

## Tests / Validación

```bash
go vet ./... && go test ./... -count=1 && go build -o huginn.exe . && go build -o huginn-clean.exe ./cmd/huginn
```

Todo OK al cierre: `vet OK`, `12 tests PASS`, `build OK` (6.7MB).

## Licencia

No definida — Agent Vault interno.
