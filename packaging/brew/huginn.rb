# © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
# Fórmula Homebrew para `huginn` (Huginn — Agent Vault Intelligence).
#
# PRERREQUISITO: debe existir el release v0.2.0 en
# https://github.com/gpb-codes/huginn-tui/releases con los assets
# huginn-darwin-amd64.tar.gz, huginn-darwin-arm64.tar.gz,
# huginn-linux-amd64.tar.gz y huginn-linux-arm64.tar.gz.
# Sin ese release, `brew install` falla con error honesto (prohibido el éxito falso).
#
# CÓMO OBTENER CADA SHA256 (uno por asset, sustituye PON_TU_SHA256_AQUI):
#   curl -fsSLO https://github.com/gpb-codes/huginn-tui/releases/download/v0.2.0/huginn-darwin-arm64.tar.gz
#   shasum -a 256 huginn-darwin-arm64.tar.gz
# Repite para cada tarball y pega el hash en su bloque on_macos/on_linux.
#
# NOTA: esta fórmula NO funciona desde este repo directamente. Requiere crear
# el tap github.com/gpb-codes/homebrew-tap y copiar este fichero como
# Formula/huginn.rb dentro del tap. Instalación entonces:
#   brew tap gpb-codes/tap
#   brew install huginn

class Huginn < Formula
  desc "Huginn — Agent Vault Intelligence (TUI)"
  homepage "https://github.com/gpb-codes/huginn-tui"
  version "0.2.0"
  # Licencia privativa del repo (ver LICENSE); sin identificador SPDX.
  license :cannot_represent

  on_macos do
    on_arm do
      url "https://github.com/gpb-codes/huginn-tui/releases/download/v0.2.0/huginn-darwin-arm64.tar.gz"
      # TODO: pega aquí el sha256 real (ver `shasum -a 256` arriba).
      sha256 "PON_TU_SHA256_AQUI"
    end
    on_intel do
      url "https://github.com/gpb-codes/huginn-tui/releases/download/v0.2.0/huginn-darwin-amd64.tar.gz"
      # TODO: pega aquí el sha256 real (ver `shasum -a 256` arriba).
      sha256 "PON_TU_SHA256_AQUI"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/gpb-codes/huginn-tui/releases/download/v0.2.0/huginn-linux-amd64.tar.gz"
      # TODO: pega aquí el sha256 real (ver `shasum -a 256` arriba).
      sha256 "PON_TU_SHA256_AQUI"
    end
    on_arm do
      url "https://github.com/gpb-codes/huginn-tui/releases/download/v0.2.0/huginn-linux-arm64.tar.gz"
      # TODO: pega aquí el sha256 real (ver `shasum -a 256` arriba).
      sha256 "PON_TU_SHA256_AQUI"
    end
  end

  def install
    bin.install "huginn"
  end

  test do
    assert_match "huginn", shell_output("#{bin}/huginn --version")
  end
end
