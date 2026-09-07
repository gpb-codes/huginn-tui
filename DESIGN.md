# HUGINN — Design System

> Premium, fluido, obsesivo con el detalle. Bronce y oro cálido.

## 1. Principios

- **Denso pero respirable:** 120x36 siempre, pero con aire entre grupos. Nada de cajas vacías.
- **Fluido > bonito:** 60fps, transiciones 150ms, sin flicker. Lo premium se siente, no se ve.
- **Bronze + Gold:** paneles cálidos, bordes bronce, acentos ámbar. Calidez real, no flat.
- **Color con propósito:** solo acentos guían la atención. El resto es bronce oscuro.

## 2. Paleta

Bronce oscuro cálido, gradients de oro. Fuente de verdad:
`internal/tui/styles/tokens.go` (`palette.go` y `header.go` solo re-exportan,
jamás redefinen hex).

| Token | Hex | Uso |
|-------|-----|-----|
| `bg` | `#130E0A` | canvas |
| `panel` | `#20160E` | paneles principales (oscuro: el borde bronce se lee) |
| `panel-2` | `#2E2013` | inset / tarjetas secundarias |
| `border` | `#634924` | bordes, muted oscuro |
| `border-2` | `#976629` | foco, gold-dark |
| `accent` | `#E1A451` | ámbar: acentos, highlights, selección |
| `accent-2` | `#CD8D38` | naranja secondary |
| `text` | `#FBE7AE` | texto primario |
| `text-2` | `#C9A86C` | texto secundario cálido |
| `muted` | `#9D8E69` | hints, timestamps |
| `muted-2` | `#7A6950` | detalle tenue |
| `success` | `#9CAF7A` | salvia cálida (hechos, nunca ámbar) |
| `warn` | `#E1A451` | ámbar (avisos) |
| `error` | `#C96A4A` | terracota (errores, nunca rojo plano) |
| `info` | `#8AB4B8` | cian apagado (un agente de chat, links) |

Semánticos siempre distinguibles: `success ≠ warn ≠ error` (hay test).
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

**Chat (pantalla principal, layout estilo opencode con marca HUGINN):**
- Sin header superior: conversación + sidebar + prompt + una sola fila de estado.
- Vacío: banner ASCII `HUGINN` centrado (bicolor ámbar/cream) + guía `/plan`, `@runtimes`, `/help`.
- Mensajes en flujo libre sin cajas, con timestamp tenue (`[15:04]`) y color por agente.
- Sidebar separado por borde `█` bronce: `AGENTS` (roles reales con %), `PIPELINE` (✓/●/○ por estado), `CONTEXT` (proyecto, pkg, vault, logs).
- Prompt estilo opencode: marco `┃` lateral + base `╹──`, sin char `>`, fondo elevado; 2ª línea inline `@agente · modelo · pkg` + `tab ⇆ agente` a la derecha.
- Footer de una línea: `tab agents   ctrl+p commands   ? help` + `cwd   vault   • HUGINN vX`.

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
