# © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
# Huginn CLI — instalador para Windows (descarga el .exe del release).
# Uso remoto:
#   irm https://raw.githubusercontent.com/gpb-codes/huginn-tui/main/install/install.ps1 | iex
# Uso local (desde la raíz del repo):
#   powershell -ExecutionPolicy Bypass -File install/install.ps1
#
# PRERREQUISITO: debe existir el release v0.2.0 en GitHub con el asset
#   huginn-windows-amd64.exe. Si no existe, el script FALLA con mensaje
#   claro (prohibido el éxito falso). El usuario crea los releases después.

$ErrorActionPreference = "Stop"

# Versión pineada (ver palette.go: const VERSION = "v0.2.0").
$Version = "v0.2.0"
$Repo = "gpb-codes/huginn-tui"
$Asset = "huginn-windows-amd64.exe"
$Url = "https://github.com/$Repo/releases/download/$Version/$Asset"

Write-Host "Huginn — instalador Windows ($Version)" -ForegroundColor Cyan
Write-Host "Descargando $Url ..."

# Descarga el .exe del release. Falla honesto si el release/asset no existe (404).
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) $Asset
try {
    Invoke-WebRequest -Uri $Url -OutFile $tmp -UseBasicParsing -TimeoutSec 60
} catch {
    Write-Error "huginn: ERROR: no existe el release $Version o el asset $Asset.`nPrerrequisito: crea el release $Version en https://github.com/$Repo/releases y sube $Asset.`nAlternativa manual: ver docs/install.md (go build, Go 1.25+). Detalle: $($_.Exception.Message)"
    exit 1
}

# Destino: %USERPROFILE%\go\bin (ya en PATH si Go está instalado) o %LOCALAPPDATA%\Programs\huginn.
$destDir = Join-Path $env:USERPROFILE "go\bin"
if (-not (Test-Path $destDir)) { $destDir = Join-Path $env:LOCALAPPDATA "Programs\huginn"; New-Item -ItemType Directory -Force -Path $destDir | Out-Null }

$dest = Join-Path $destDir "huginn.exe"
Copy-Item -Force $tmp $dest
Remove-Item -Force $tmp -ErrorAction SilentlyContinue
Write-Host "Instalado en: $dest" -ForegroundColor Green

# Verifica PATH.
$path = [Environment]::GetEnvironmentVariable("Path", "User") + ";" + [Environment]::GetEnvironmentVariable("Path", "Machine")
if ($path -notlike "*$destDir*") {
    Write-Host "Agregando $destDir al PATH de usuario..." -ForegroundColor Yellow
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$destDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$destDir", "User")
        $env:Path += ";$destDir"
        Write-Host "PATH actualizado. Reinicia la terminal." -ForegroundColor Yellow
    }
}

# Verifica que el binario responde (si falla, no es éxito).
Write-Host ""
Write-Host "Verifica:" -ForegroundColor Cyan
& $dest --version
if ($LASTEXITCODE -ne 0) { Write-Error "huginn: el binario instalado no responde a --version"; exit 1 }
& $dest --help | Select-Object -First 20 | Out-String | Write-Host
Write-Host ""
Write-Host "Prueba:" -ForegroundColor Cyan
Write-Host "  huginn" -ForegroundColor White
Write-Host "  huginn . " -ForegroundColor White
Write-Host "  huginn C:\Projects\mi-proyecto" -ForegroundColor White
Write-Host "  huginn `"analiza este proyecto`"" -ForegroundColor White
