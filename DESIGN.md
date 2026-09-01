# HUGINN — Design System

> Premium, fluido, obsesivo con el detalle. Inspirado en Linear + Raycast + Obsidian.

## 1. Principios

- **Denso pero respirable:** 120x36 siempre, pero con aire entre grupos. Nada de cajas vacías.
- **Fluido > bonito:** 60fps, transiciones 150ms, sin flicker. Lo premium se siente, no se ve.
- **Glass + Depth:** paneles con blur sutil, bordes 1px, sombras suaves. Profundidad real, no flat.
- **Color con propósito:** solo acentos guían la atención. El resto es gris frío.

## 2. Paleta

Base oscuro frío, no negro puro.

| Token | Hex | Uso |
|-------|-----|-----|
| `bg` | `#080c14` | canvas |
| `panel` | `#111827` | paneles principales |
| `panel-2` | `#151e2f` | panel secundario / hover |
| `border` | `#1f2a3a` | bordes 1px |
| `border-2` | `#263449` | bordes foco |
| `text` | `#e6edf3` | texto primario |
| `text-2` | `#9aa8b8` | secundario |
| `muted` | `#5b6b82` | terciario / placeholder |
| `accent` | `#22d3ee` | cian — acciones, focus, links |
| `accent-2` | `#8b5cf6` | púrpura — agentes, highlights |
| `success` | `#34d399` | conectado, ok |
| `warn` | `#fbbf24` | atención |
| `error` | `#f87171` | error |

Gradientes solo en acentos: `purple→cyan` para HUGINN wordmark, nunca en fondos.

## 3. Tipografía

- **Display:** `JetBrains Mono` 700, tracking -0.02em, para `HUGINN` wordmark y números grandes. `HU` en `accent-2`, `GINN` en `text`.
- **UI:** `JetBrains Mono` 400/500, 13–14px, line-height 1.5. Nada de Inter — todo mono para coherencia TUI.
- **Code:** `JetBrains Mono` 400, 13px, `text-2` sobre `panel`.
- Escala: 12 / 13 / 14 / 16 / 20 / 32 (logo) / 56 (hero solo en web).

## 4. Espaciado & Layout

- Unidad 8px. Padding panel: 16 (y) 20 (x). Gap vertical: 12.
- Grid 120 cols TUI: left 68% + right 32% en chat, 76 cols centrado en wizard/settings.
- Bordes: 1px `border`, radius 10 (TUI rounded 6), sombra `0 8 24 rgba(0,0,0,.4)`.
- Sidebar nav: 28 cols fijo, colapsable a 8 (iconos).

## 5. Motion

- Duraciones: `150ms` micro (hover, focus), `220ms` panel, `300ms` modal (spring 1.2).
- Easing: `cubic-bezier(.2,.8,.2,1)` — sale rápido, entra suave.
- Nada de bounce. Nada de fade largo.

## 6. Componentes TUI

**Header:** 1 línea, `HUGINN • AI ORCHESTRATOR • CONNECTED` con barra fina abajo. Sin logo gigante.

**Chat:**
- Input sticky abajo, con `> ` + caret bloque blanco, hint `escribe para chatear — Enter • @opcional`.
- Mensajes con avatar letra (`H` `You`) y timestamp tenue.
- Agent Dispatch como chips, no tabla: `ChatGPT → Architecture ✓` en línea.
- Background Tasks como lista compacta con barra progreso fina 2px.

**Vault wizard:**
- Progreso `— 3/7 • /vault • Propósito • 42%` con barra 2px cyan.
- Opciones como lista radio, preview `memory/agents/knowledge` a la derecha en card glass.

**Graph:**
- Nodos como pills con `•` color por tipo, edges como `─` tenue. Focus con glow `accent` 1px + bg `panel-2`.

**Settings:**
- Árbol indentado con `▸/▾`, valores a la derecha en `text-2`. Toggle con `●/○`.

**Servers:**
- Tabs `MCP LSP Peers` con underline cyan activo. Status pill `● Connected 12ms` en `success` sobre `panel-2`.

## 7. Estados

- **Empty:** ilustración línea fina + `No vaults — Seleccionar carpeta` botón cyan.
- **Loading:** spinner `⠋` 120ms, no texto saltando.
- **Error:** borde `error` + mensaje 1 línea + acción `Retry`.

## 8. Premium details

- Cursor bloque `#e6edf3` sobre `#0a0c0f`.
- Selección texto con bg `accent` 20%.
- Scrollbar fina 2px `border-2` sobre `bg`, thumb `muted`.
- Atajos siempre visibles en footer: `TAB Agent  CTRL+P Commands  CTRL+G Graph`.

## 9. No hacer

- No gradientes en fondos.
- No emojis.
- No cajas dobles.
- No colores planos 100% saturados fuera de acentos.

## 10. Implementación

Tokens en `internal/tui/styles/tokens.go` (colores, paddings, durations).
Componentes en `internal/tui/components/{header,chat,agents,tasks,input,progress}.go`.
Vistas solo componen, no definen colores.

---

*Fuente de verdad para TUI y web. Cambiar aquí, no en cada view.*
