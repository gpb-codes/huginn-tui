# Huginn CLI — instalador global para Windows
# Ejecuta: powershell -ExecutionPolicy Bypass -File install.ps1

$ErrorActionPreference = "Stop"

Write-Host "Huginn — instalador global" -ForegroundColor Cyan
Write-Host "Construyendo huginn.exe ..."

# requiere Go 1.25+
go version | Out-Null
if ($LASTEXITCODE -ne 0) { Write-Error "Go no encontrado. Instala Go 1.25+"; exit 1 }

go vet ./...
if ($LASTEXITCODE -ne 0) { Write-Error "go vet falló"; exit 1 }

go test ./... -count=1 | Out-Null
if ($LASTEXITCODE -ne 0) { Write-Warning "tests fallaron, continuando..." }

go build -o huginn.exe .
if ($LASTEXITCODE -ne 0) { Write-Error "build falló"; exit 1 }

# destino: %USERPROFILE%\go\bin (ya en PATH si Go está instalado) o %LOCALAPPDATA%\Programs\huginn
$destDir = Join-Path $env:USERPROFILE "go\bin"
if (-not (Test-Path $destDir)) { $destDir = Join-Path $env:LOCALAPPDATA "Programs\huginn"; New-Item -ItemType Directory -Force -Path $destDir | Out-Null }

$dest = Join-Path $destDir "huginn.exe"
Copy-Item -Force huginn.exe $dest
Write-Host "Instalado en: $dest" -ForegroundColor Green

# verifica PATH
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

Write-Host ""
Write-Host "Verifica:" -ForegroundColor Cyan
& $dest --version
& $dest --help | Select-Object -First 20 | Out-String | Write-Host
Write-Host ""
Write-Host "Prueba:" -ForegroundColor Cyan
Write-Host "  huginn" -ForegroundColor White
Write-Host "  huginn . " -ForegroundColor White
Write-Host "  huginn C:\Projects\mi-proyecto" -ForegroundColor White
Write-Host "  huginn `"analiza este proyecto`"" -ForegroundColor White
