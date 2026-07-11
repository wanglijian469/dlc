$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$release = Join-Path $root "release\dalu-parts-windows"
$public = Join-Path $release "public"
Push-Location (Join-Path $root "frontend")
try { npm run build; if ($LASTEXITCODE -ne 0) { throw "Frontend build failed with exit code $LASTEXITCODE" } } finally { Pop-Location }
if (Test-Path $public) { Remove-Item -LiteralPath $public -Recurse -Force }
New-Item -ItemType Directory -Force $public | Out-Null
Copy-Item (Join-Path $root "frontend\dist\*") $public -Recurse -Force
Push-Location (Join-Path $root "backend")
try { go build -trimpath -ldflags "-s -w" -o (Join-Path $release "dalu-parts.exe") ./cmd/server; if ($LASTEXITCODE -ne 0) { throw "Backend build failed with exit code $LASTEXITCODE" } } finally { Pop-Location }
Write-Host "Release created at $release"
