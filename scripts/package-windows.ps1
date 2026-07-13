param(
  [string]$Version = (Get-Date -Format "yyyyMMdd-HHmmss"),
  [switch]$IncludeMedia
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$PackageName = "dlc-deploy-windows-$Version"
$PackageRoot = Join-Path $Root "release\$PackageName"
$ZipPath = Join-Path $Root "release\$PackageName.zip"

Remove-Item -Recurse -Force $PackageRoot -ErrorAction SilentlyContinue
Remove-Item -Force $ZipPath -ErrorAction SilentlyContinue

@("public", "media_storage", "logs", "scripts", "config", "backups") | ForEach-Object {
  New-Item -ItemType Directory -Force (Join-Path $PackageRoot $_) | Out-Null
}

Push-Location (Join-Path $Root "frontend")
try {
  npm run build
  if ($LASTEXITCODE -ne 0) { throw "Frontend build failed with exit code $LASTEXITCODE" }
} finally { Pop-Location }

Push-Location (Join-Path $Root "backend")
try {
  $env:GOTMPDIR = Join-Path $env:TEMP "dlc-go-build"
  New-Item -ItemType Directory -Force $env:GOTMPDIR | Out-Null
  go build -trimpath -ldflags "-s -w" -o (Join-Path $PackageRoot "server.exe") ./cmd/server
  if ($LASTEXITCODE -ne 0) { throw "Backend build failed with exit code $LASTEXITCODE" }
} finally { Pop-Location }

Copy-Item -Recurse -Force (Join-Path $Root "frontend\dist\*") (Join-Path $PackageRoot "public")
Copy-Item -Force (Join-Path $Root "deploy\config\.env.example") (Join-Path $PackageRoot ".env.example")
Copy-Item -Force (Join-Path $Root "deploy\DEPLOY.md") (Join-Path $PackageRoot "DEPLOY.md")
Copy-Item -Force (Join-Path $Root "deploy\README.txt") (Join-Path $PackageRoot "README.txt")
Copy-Item -Force (Join-Path $Root "deploy\config\nginx.conf.example") (Join-Path $PackageRoot "config\nginx.conf.example")
Copy-Item -Recurse -Force (Join-Path $Root "deploy\scripts\*") (Join-Path $PackageRoot "scripts")

if ($IncludeMedia -and (Test-Path (Join-Path $Root "media_storage"))) {
  Copy-Item -Recurse -Force (Join-Path $Root "media_storage\*") (Join-Path $PackageRoot "media_storage")
}

New-Item -ItemType File -Force (Join-Path $PackageRoot "media_storage\.gitkeep") | Out-Null
New-Item -ItemType File -Force (Join-Path $PackageRoot "logs\.gitkeep") | Out-Null
New-Item -ItemType File -Force (Join-Path $PackageRoot "backups\.gitkeep") | Out-Null

Compress-Archive -Path (Join-Path $PackageRoot "*") -DestinationPath $ZipPath -Force
Write-Host "Deployment package created: $ZipPath"
