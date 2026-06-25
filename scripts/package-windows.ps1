param(
  [string]$Version = (Get-Date -Format "yyyyMMdd-HHmmss")
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$PackageRoot = Join-Path $Root "release\dlc-mvp-$Version"
$ZipPath = "$PackageRoot.zip"

Remove-Item -Recurse -Force $PackageRoot -ErrorAction SilentlyContinue
Remove-Item -Force $ZipPath -ErrorAction SilentlyContinue

New-Item -ItemType Directory -Force $PackageRoot | Out-Null
New-Item -ItemType Directory -Force (Join-Path $PackageRoot "frontend") | Out-Null
New-Item -ItemType Directory -Force (Join-Path $PackageRoot "scripts") | Out-Null
New-Item -ItemType Directory -Force (Join-Path $PackageRoot "config") | Out-Null

Push-Location (Join-Path $Root "frontend")
npm install
npm run build
Pop-Location

Push-Location (Join-Path $Root "backend")
$env:GOTMPDIR = Join-Path (Get-Location) ".gotmp"
New-Item -ItemType Directory -Force $env:GOTMPDIR | Out-Null
go build -o (Join-Path $PackageRoot "server.exe") ./cmd/server
Pop-Location

Copy-Item -Recurse (Join-Path $Root "frontend\dist") (Join-Path $PackageRoot "frontend\dist")
Copy-Item (Join-Path $Root "deploy\config\.env.example") (Join-Path $PackageRoot ".env.example")
Copy-Item (Join-Path $Root "deploy\config\nginx.conf.example") (Join-Path $PackageRoot "config\nginx.conf.example")
Copy-Item (Join-Path $Root "deploy\scripts\init-db.sql") (Join-Path $PackageRoot "scripts\init-db.sql")
Copy-Item (Join-Path $Root "deploy\scripts\start-windows.ps1") (Join-Path $PackageRoot "start-windows.ps1")
Copy-Item (Join-Path $Root "deploy\scripts\start-linux.sh") (Join-Path $PackageRoot "scripts\start-linux.sh")
Copy-Item (Join-Path $Root "README.md") (Join-Path $PackageRoot "README.md")

@"
# Deployment

1. Create database with scripts/init-db.sql, or let server.exe create DB with root credentials.
2. Copy .env.example to .env and update DB credentials and AUTH_SECRET.
3. Windows: powershell -ExecutionPolicy Bypass -File .\start-windows.ps1
4. Linux: chmod +x scripts/start-linux.sh && ./scripts/start-linux.sh

The backend serves API and frontend static files when PUBLIC_DIR=./frontend/dist.
"@ | Set-Content -Encoding UTF8 (Join-Path $PackageRoot "DEPLOY.md")

Compress-Archive -Path (Join-Path $PackageRoot "*") -DestinationPath $ZipPath -Force
Write-Host "Package created: $ZipPath"
