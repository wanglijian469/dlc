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
$env:GOTMPDIR = Join-Path $env:TEMP "dlc-go-build"
New-Item -ItemType Directory -Force $env:GOTMPDIR | Out-Null
go build -o (Join-Path $PackageRoot "server.exe") ./cmd/server
Pop-Location

Copy-Item -Recurse (Join-Path $Root "frontend\dist") (Join-Path $PackageRoot "frontend\dist")
Copy-Item (Join-Path $Root "deploy\config\.env.example") (Join-Path $PackageRoot ".env.example")
Copy-Item (Join-Path $Root "deploy\config\nginx.conf.example") (Join-Path $PackageRoot "config\nginx.conf.example")
Copy-Item (Join-Path $Root "deploy\scripts\init-db.sql") (Join-Path $PackageRoot "scripts\init-db.sql")
Copy-Item (Join-Path $Root "deploy\scripts\start-windows.ps1") (Join-Path $PackageRoot "start-windows.ps1")
Copy-Item (Join-Path $Root "README.md") (Join-Path $PackageRoot "README.md")

$DeployDoc = @(
  "# Dalu Nongji Parts Windows Deployment",
  "",
  "## Package Contents",
  "",
  "- server.exe: backend API server. It also serves frontend static files when PUBLIC_DIR is set.",
  "- frontend/dist: production frontend static files.",
  "- .env.example: runtime configuration example. Defaults target MySQL 5.7 at 127.0.0.1:13306 with root/root.",
  "- scripts/init-db.sql: optional database initialization script.",
  "- start-windows.ps1: Windows startup script.",
  "- config/nginx.conf.example: optional Nginx reverse proxy example.",
  "",
  "## Deployment Steps",
  "",
  "1. Extract this zip to a server directory, for example D:\apps\dlc.",
  "2. Copy .env.example to .env and update DB_PASSWORD and AUTH_SECRET before production use.",
  "3. Ensure MySQL 5.7 is running. Default connection: 127.0.0.1:13306.",
  "4. Optional database initialization:",
  "   mysql -uroot -proot -P13306 < scripts\init-db.sql",
  "5. Start the system:",
  "   powershell -ExecutionPolicy Bypass -File .\start-windows.ps1",
  "6. Open:",
  "   http://SERVER_IP:8080/",
  "   http://SERVER_IP:8080/admin/login",
  "",
  "Notes: server.exe automatically creates the database, runs migrations, and seeds default data on startup. The frontend production build uses same-origin /api routes by default, so it will not call the visitor machine's 127.0.0.1."
)
$DeployDoc | Set-Content -Encoding UTF8 (Join-Path $PackageRoot "DEPLOY.md")

Compress-Archive -Path (Join-Path $PackageRoot "*") -DestinationPath $ZipPath -Force
Write-Host "Package created: $ZipPath"
