#!/usr/bin/env bash
# © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
# Instalador de Huginn (Linux / macOS / WSL):
#   curl -fsSL https://raw.githubusercontent.com/gpb-codes/huginn-tui/main/install/install.sh | bash
#
# PRERREQUISITO: debe existir el release v0.2.0 en GitHub con los assets
#   huginn-linux-amd64.tar.gz, huginn-linux-arm64.tar.gz,
#   huginn-darwin-amd64.tar.gz, huginn-darwin-arm64.tar.gz.
# Si el release o el asset no existe, este script FALLA con mensaje claro
# (prohibido el éxito falso). El usuario crea los releases después.
set -euo pipefail

REPO="gpb-codes/huginn-tui"
BIN="huginn"
VERSION="v0.2.0"
DEST="${HUGINN_INSTALL_DIR:-$HOME/.local/bin}"

# Detecta SO y arquitectura (nombres estilo Go: linux/darwin + amd64/arm64).
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "huginn: arquitectura no soportada: $arch" >&2; exit 1 ;;
esac
case "$os" in
  linux|darwin) ;;
  *) echo "huginn: SO no soportado: $os (en Windows usa install/install.ps1)" >&2; exit 1 ;;
esac

# Asset pineado a la versión (no `latest`: instalación reproducible).
ASSET="$BIN-$os-$arch.tar.gz"
url="https://github.com/$REPO/releases/download/$VERSION/$ASSET"

mkdir -p "$DEST"

# Descarga el tarball del release. Falla honesto si no existe (404).
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
echo "huginn: descargando $url …"
if ! curl -fsSL --max-time 30 -o "$tmp/$ASSET" "$url"; then
  echo "huginn: ERROR: no existe el release o el asset $ASSET" >&2
  echo "huginn: prerrequisito: crea el release $VERSION en https://github.com/$REPO/releases" >&2
  echo "huginn: y sube el asset $ASSET (o compila manual: ver docs/install.md, Go 1.25+)" >&2
  exit 1
fi

# Extrae el binario del tarball (espera un único `huginn` dentro).
if ! tar -xzf "$tmp/$ASSET" -C "$tmp"; then
  echo "huginn: ERROR: el asset descargado no es un tar.gz válido: $ASSET" >&2
  exit 1
fi
if [ ! -f "$tmp/$BIN" ]; then
  echo "huginn: ERROR: el tarball $ASSET no contiene el binario '$BIN'" >&2
  exit 1
fi
mv "$tmp/$BIN" "$DEST/$BIN"
chmod +x "$DEST/$BIN"
echo "huginn: instalado en $DEST/$BIN ($VERSION $os/$arch)"

case ":$PATH:" in
  *":$DEST:"*) ;;
  *) echo "huginn: añade a tu PATH: export PATH=\"$DEST:\$PATH\"" ;;
esac

# Verifica que el binario responde (si falla, no es éxito).
"$DEST/$BIN" --version
