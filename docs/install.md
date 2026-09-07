# Instalación de Huginn (`huginn`, v0.2.0)

> **Prerrequisito (importante):** todos los métodos salvo la compilación manual
> descargan el binario del **release `v0.2.0`** en
> `https://github.com/gpb-codes/huginn-tui/releases`.
> Ese release lo crea el usuario después e incluirá estos assets:
>
> | Asset | Plataforma |
> |---|---|
> | `huginn-linux-amd64.tar.gz` | Linux x86_64 |
> | `huginn-linux-arm64.tar.gz` | Linux ARM64 |
> | `huginn-darwin-amd64.tar.gz` | macOS Intel |
> | `huginn-darwin-arm64.tar.gz` | macOS Apple Silicon |
> | `huginn-windows-amd64.exe` | Windows x86_64 |
>
> **Regla de oro:** si el release o el asset de tu plataforma no existe,
> el instalador **falla con un mensaje claro de error** (prohibido el éxito
> falso). No finge una instalación correcta.
>
> Versión actual verificada en `palette.go`: `const VERSION = "v0.2.0"`.
> Compilación manual requiere **Go 1.25+** (ver `go.mod`: `go 1.25.0`).

## Matriz de instalación

| Gestor | Comando exacto |
|---|---|
| curl (Linux/macOS/WSL) | `curl -fsSL https://raw.githubusercontent.com/gpb-codes/huginn-tui/main/install/install.sh \| bash` |
| Windows (PowerShell) | `irm https://raw.githubusercontent.com/gpb-codes/huginn-tui/main/install/install.ps1 \| iex` |
| npm | `npm install -g huginn-tui` |
| pnpm | `pnpm add -g huginn-tui` |
| bun | `bun add -g huginn-tui` |
| Homebrew (tras crear el tap) | `brew tap gpb-codes/tap && brew install huginn` |
| AUR (tras subir el PKGBUILD) | `paru -S huginn-tui` |
| Manual (`go build`) | `go build -o huginn . && ./huginn --version` |

npm, pnpm y bun usan el **mismo paquete sin cambios**
(`packaging/npm/`, solo stdlib de Node, sin dependencias):
los tres ejecutan el `postinstall` que descarga el asset de tu plataforma.

## Detalles por método

### curl (bash)

```bash
curl -fsSL https://raw.githubusercontent.com/gpb-codes/huginn-tui/main/install/install.sh | bash
```

- Script: `install/install.sh`. Instala en `$HUGINN_INSTALL_DIR` o
  `~/.local/bin` por defecto.
- Si `~/.local/bin` no está en tu `PATH`, el script te dice cómo añadirlo:
  `export PATH="$HOME/.local/bin:$PATH"`.
- Si el release/asset no existe, falla con el prerrequisito y sale `1`.

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/gpb-codes/huginn-tui/main/install/install.ps1 | iex
```

- Script: `install/install.ps1`. Descarga `huginn-windows-amd64.exe` del
  release `v0.2.0` e instala en `%USERPROFILE%\go\bin` (o
  `%LOCALAPPDATA%\Programs\huginn`), añadiéndolo al `PATH` de usuario.
- Ejecución local desde la raíz del repo:
  `powershell -ExecutionPolicy Bypass -File install/install.ps1`.
- Si el release/asset no existe, falla con el prerrequisito y sale `1`.

### npm / pnpm / bun

```bash
npm install -g huginn-tui
pnpm add -g huginn-tui
bun add -g huginn-tui
```

- Paquete `huginn-tui@0.2.0` (`packaging/npm/`): el `postinstall`
  (`install.js`, solo stdlib) descarga el asset según tu plataforma
  (`win32-x64`, `darwin-arm64`, `darwin-x64`, `linux-x64`, `linux-arm64`)
  y el lanzador `bin/huginn.js` ejecuta el binario propagando su exit code.
- Si el release/asset no existe (404), el `postinstall` falla con mensaje
  claro y la instalación no se marca como correcta.

### Homebrew

```bash
brew tap gpb-codes/tap
brew install huginn
```

- Fórmula: `packaging/brew/huginn.rb`.
- **Requiere crear antes el tap** `github.com/gpb-codes/homebrew-tap`
  y copiar la fórmula como `Formula/huginn.rb`.
- Sustituye cada `PON_TU_SHA256_AQUI` por el hash real del tarball:
  `shasum -a 256 huginn-<os>-<arch>.tar.gz`.
- Verifica con `huginn --version`.

### AUR (`paru`)

```bash
paru -S huginn-tui
```

- PKGBUILD: `packaging/aur/PKGBUILD` (`pkgname=huginn-tui`, `pkgver=0.2.0`,
  `arch=('x86_64' 'aarch64')`, instala en `/usr/bin/huginn`).
- **Requiere subir antes el PKGBUILD a AUR.**
- Rellena los hashes ejecutando `updpkgsums` en el directorio del PKGBUILD
  (sustituye los `PON_TU_SHA256_AQUI`).

### Manual (`go build`, Go 1.25+)

```bash
go build -o huginn .
./huginn --version
# esperado: huginn v0.2.0
```

- No necesita el release de GitHub; es la alternativa cuando el release
  aún no existe.
- Requiere **Go 1.25+** (`go version` para comprobarlo).

## Verificación

```bash
huginn --version
# esperado: huginn v0.2.0
```
