<p align="center">
  <img src="./logos/huginn-logo-horizontal-fondo-oscuro.png" width="900" alt="HUGINN — La TUI para Agent Vault" />
</p>

<p align="center">La TUI open source para Agent Vault.</p>

<p align="center">
  <a href="https://github.com/gpb-codes/huginn-tui"><img src="https://img.shields.io/github/stars/gpb-codes/huginn-tui?style=flat-square&label=stars" alt="stars" /></a>
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go 1.25" />
  <img src="https://img.shields.io/badge/Bubble_Tea-v2-9061f9?style=flat-square" alt="Bubble Tea v2" />
  <img src="https://img.shields.io/badge/Licencia-Privativa-red?style=flat-square" alt="Privativa — todos los derechos reservados" />
  <img src="https://img.shields.io/badge/plataforma-Windows%20%7C%20Linux%20%7C%20macOS-black?style=flat-square" alt="plataforma" />
</p>

<p align="center">
  <a href="./landing/index.html"><b>Landing page</b> — ábrela en el navegador</a>
</p>

---

### Qué es HUGINN

**HUGINN — AI Agent Orchestration Environment.**

Huginn es el cerebro que piensa y coordina. Los agentes ejecutan.

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
                           │
                           ▼
                         USER
```

Huginn **no** es otro agente de coding. Su responsabilidad es:

- recibir objetivos, planificar, seleccionar agentes
- proporcionar contexto, coordinar workflows, controlar estado
- administrar sesiones, gestionar memoria, registrar eventos, presentar resultados

Filosofía: **sin éxitos falsos**. Si un provider no está disponible, Huginn lo dice con un error honesto. Nunca simula funcionalidad ni inventa respuestas.

### Instalación

```bash
# curl (Linux / macOS / WSL)
curl -fsSL https://raw.githubusercontent.com/gpb-codes/huginn-tui/main/install/install.sh | bash

# npm / pnpm / bun
npm i -g huginn-tui
pnpm add -g huginn-tui
bun add -g huginn-tui

# brew (requiere crear el tap github.com/gpb-codes/homebrew-tap)
brew tap gpb-codes/tap && brew install huginn

# Arch Linux (requiere subir packaging/aur/PKGBUILD a AUR)
paru -S huginn-tui
```

> Detalle por gestor y prerrequisito del release: [`docs/install.md`](./docs/install.md).

```bash
# compilar desde código fuente
git clone https://github.com/gpb-codes/huginn-tui.git
cd huginn-tui
go build -o huginn .
./huginn --help

# Windows — PowerShell
powershell -ExecutionPolicy Bypass -File install/install.ps1
```

---

### Uso

```bash
huginn                              # abre la TUI en el proyecto actual
huginn ./mi-proyecto                # con contexto de proyecto
huginn "analiza este proyecto"      # prompt directo
huginn --help
huginn --version                    # v0.2.0
```

#### Homescreen

```
                         H U G I N N

                 AI Agent Orchestration

                 ┌─────────────────────┐
                 │ > Ask Huginn...     │
                 └─────────────────────┘

                 Type / for commands
```

Pantalla **extremadamente limpia**: título centrado, input centrado con mucho espacio negativo, tipografía monoespaciada, jerarquía clara. Inspirada en OpenCode pero con identidad propia de HUGINN — oscura, minimalista, profesional.

Dentro de la TUI:

| Tecla / Comando | Acción |
|----------------|--------|
| `Type` + `Enter` | Envía prompt/objetivo al orquestador (flujo `Tú → Huginn → Agentes → Huginn → Tú`) |
| `/help`, `?` | Ayuda |
| `/agents` | Agentes (contrato listo, aún sin conexión si no hay providers) |
| `/tasks` | Tareas (ciclo PENDING→COMPLETED, aún sin tareas) |
| `/projects` | Proyecto actual |
| `/sessions` | Sesiones (preparado) |
| `/memory` | Memoria (capas previstas) |
| `/workflows` | Workflows (DAGs futuros) |
| `/models` | Modelos (Ollama / OpenCode / HuggingFace) |
| `/config` | Config y vault |
| `/clear` | Limpia conversación |
| `/undo` | Revierte último mensaje |
| `/exit`, `/quit` | Salir |
| `Tab` / `Shift+Tab` | Cicla modos `Ask → Architect → Code → Debug → Orchestrator` (se ve en `● Ask · Huginn`) |
| `Ctrl+P` | Paleta de comandos (12 cmds, `↑/↓` + `Enter`) |
| `Esc` | Volver / limpiar input / cerrar paleta |
| `k/j` `↑/↓` `PgUp/PgDn` | Scroll en conversación |
| `Ctrl+C` | Salir |

Logo `HUG` café `#8B5A2B` / `INN` blanco en ASCII de caja, input `60` chars centrado con `> Ask Huginn…` y `● modo · Huginn Zen`. Homescreen con `Type / for commands` y `tab`/`ctrl+p`.

Todos los comandos no implementados responden con **mensaje honesto** y el contrato que los soportará. Nunca simulan éxito.

Comandos de vault (CLI, sí implementados):

```bash
huginn vault              # muestra el vault actual
huginn vault open <path>  # abre un vault existente
huginn vault create "Mi Vault" ./projects
huginn vault list
```

---

### Arquitectura

```
cmd/huginn        →  internal/tui  →  internal/events  →  internal/domain
cmd/huginn-bot    →  internal/bot  →  internal/bootstrap
```

`domain` nunca importa `lipgloss` ni `bubbletea`. `tui` renderiza y envía comandos. Todo `exec` vive en `infrastructure/agents`.

```
main.go                                    # entrada fina: delega a internal/tui
cmd/huginn/main.go                         # entrada fina: delega a internal/tui
cmd/huginn-bot/main.go                     # bot: servidor local 127.0.0.1:8765
internal/tui/app.go                        # Model minimalista (homescreen + comandos)
internal/tui/main.go                       # ParseArgs → ResolveContext → Run
internal/tui/styles/tokens.go              # Design system (dark, bronce/dorado)
internal/cli/cli.go                        # ParseArgs honesto
internal/events/bus.go                     # Pub/Sub tipado (TaskCreated, AgentMessage…)
internal/domain/task/status.go             # PENDINGPLANNINGRUNNINGWAITINGCOMPLETEDFAILED/CANCELLED
internal/domain/agent/{agent,capability}.go # Agent + Capability (routing por capacidad)
internal/domain/{plan,execution,workspace}  # DAG, Run, Project detector
internal/application/{planner,router,evaluator,orchestrator}
internal/infrastructure/{agents,models,config,memory,vault,…}
```

Ver detalle en [`ARCHITECTURE.md`](./ARCHITECTURE.md) y [`DESIGN.md`](./DESIGN.md).

### Vault

Agent Vault es la capa persistente. Huginn no la duplica — la consume vía ports.

```
.huginn/
  config.json   vault.json   agents.json   memory.jsonl
  agents/  memory/  plugins/  cache/ logs/ runtime/ (ignorados)
```

Orden: `HUGINN_VAULT` > `AGENT_VAULT` > `~/agent-vault` > `~/huginn-vault`.

### Agentes

Abstracción desacoplada:

```
Agent
 ├── ID, Name, Description
 ├── Capabilities []
 ├── Status
 ├── Provider / Configuration
 └── Execute(ctx, task) (Result, error)
```

Runtimes actuales: **OpenCode** (coding, CLI real) y **Ollama** (research, local). Kilo/Hermes/Goose son futuras integraciones vía `router.Agent`. Routing **por capacidad**, nunca `if agent == "opencode"` en el core. Ver [`docs/agents.md`](./docs/agents.md).

### Verificaciones

```bash
go vet ./... && go test ./... -count=1 && go build ./...
go run . --help
go run . --dump-ansi   # dump del homescreen para la landing
```

### Roadmap

Fases completadas: auditoría → reset TUI → foundation (domain/events/config) → homescreen minimal → comandos honestos → agent/task contracts → integrations stub → memory interfaces.

Pendiente (sin inventar): conectar planner/router real a la TUI, motor de tareas con scheduler, memoria persistente con embeddings, neural engine para ranking de contexto.

### Contribuir

```bash
git checkout -b feat/mi-feature
go vet ./... && go test ./... -count=1 && go build ./...
git commit -m "feat: mi feature"
```

Mantén `domain` libre de UI y añade tests para lógica nueva.

### Bot local

```bash
go build -o huginn-bot ./cmd/huginn-bot
HUGINN_BOT_TOKEN=secreto ./huginn-bot --project .   # 127.0.0.1:8765
curl http://127.0.0.1:8765/health
```

`GET /health /status` y `POST /chat` con token. Sin token no arranca.

### Landing

```bash
cd landing && npm install && npm run build
```

---

<p align="center">
  <a href="https://github.com/gpb-codes/huginn-tui">github.com/gpb-codes/huginn-tui</a><br/>
  <sub>Go 1.25 · Bubble Tea v2 · Lipgloss v2</sub>
</p>
