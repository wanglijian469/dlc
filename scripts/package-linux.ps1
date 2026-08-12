param(
  [string]$Version = "v2026.08.12-linux.2"
)

$ErrorActionPreference = "Stop"
if ($Version -notmatch '^[A-Za-z0-9._-]+$') { throw "Version may contain only letters, numbers, dots, underscores, and hyphens." }

$Root = Split-Path -Parent $PSScriptRoot
$ReleaseDir = Join-Path $Root "release"
$PackageName = "dlc-deploy-linux-x86_64-$Version"
$Archive = Join-Path $ReleaseDir "$PackageName.tar.gz"
$WorkDir = Join-Path ([System.IO.Path]::GetTempPath()) ("dlc-linux-package-" + [guid]::NewGuid().ToString("N"))
$PackageRoot = Join-Path $WorkDir $PackageName

function Copy-PackageFile([string]$Source, [string]$DestinationDirectory) {
  New-Item -ItemType Directory -Force $DestinationDirectory | Out-Null
  Copy-Item -LiteralPath $Source -Destination $DestinationDirectory -Force
}

try {
  New-Item -ItemType Directory -Force $PackageRoot, $ReleaseDir | Out-Null
  @("public", "config", "systemd", "nginx", "scripts", "docs") | ForEach-Object { New-Item -ItemType Directory -Force (Join-Path $PackageRoot $_) | Out-Null }

  Push-Location (Join-Path $Root "frontend")
  try { npm run build; if ($LASTEXITCODE -ne 0) { throw "Frontend build failed with exit code $LASTEXITCODE" } } finally { Pop-Location }

  $oldCGO = $env:CGO_ENABLED; $oldGOOS = $env:GOOS; $oldGOARCH = $env:GOARCH
  try {
    $env:CGO_ENABLED = "0"; $env:GOOS = "linux"; $env:GOARCH = "amd64"
    Push-Location (Join-Path $Root "backend")
    try {
      go build -trimpath -buildvcs=false -ldflags "-s -w" -o (Join-Path $PackageRoot "server") ./cmd/server
      if ($LASTEXITCODE -ne 0) { throw "Linux server build failed with exit code $LASTEXITCODE" }
      go build -trimpath -buildvcs=false -ldflags "-s -w" -o (Join-Path $PackageRoot "initdb") ./cmd/initdb
      if ($LASTEXITCODE -ne 0) { throw "Linux migration build failed with exit code $LASTEXITCODE" }
    } finally { Pop-Location }
  } finally {
    $env:CGO_ENABLED = $oldCGO; $env:GOOS = $oldGOOS; $env:GOARCH = $oldGOARCH
  }

  Copy-Item -Recurse -Force (Join-Path $Root "frontend\dist\*") (Join-Path $PackageRoot "public")
  Copy-PackageFile (Join-Path $Root "deploy\linux\config\dlc.env.example") (Join-Path $PackageRoot "config")
  Copy-PackageFile (Join-Path $Root "deploy\linux\systemd\dalu-parts.service") (Join-Path $PackageRoot "systemd")
  Copy-PackageFile (Join-Path $Root "deploy\linux\nginx\dalu-parts.conf.template") (Join-Path $PackageRoot "nginx")
  Copy-Item -Force (Join-Path $Root "deploy\linux\scripts\*.sh") (Join-Path $PackageRoot "scripts")
  # The source tree can be checked out on Windows. Bash scripts must always
  # have LF line endings in the Linux archive; CRLF can break command parsing.
  Get-ChildItem -LiteralPath (Join-Path $PackageRoot "scripts") -Filter "*.sh" -File | ForEach-Object {
    $scriptText = [System.IO.File]::ReadAllText($_.FullName)
    [System.IO.File]::WriteAllText($_.FullName, ($scriptText -replace "`r`n", "`n" -replace "`r", "`n"), (New-Object System.Text.UTF8Encoding($false)))
  }
  Copy-PackageFile (Join-Path $Root "deploy\linux\README.md") $PackageRoot
  @("INSTALL.md", "UPGRADE.md", "RELEASE_NOTES.md") | ForEach-Object { Copy-PackageFile (Join-Path $Root "deploy\linux\$_") (Join-Path $PackageRoot "docs") }
  Set-Content -NoNewline -Encoding ascii (Join-Path $PackageRoot "VERSION") $Version

  $checksums = Get-ChildItem -LiteralPath $PackageRoot -Recurse -File |
    Where-Object { $_.Name -ne "SHA256SUMS" } |
    Sort-Object FullName |
    ForEach-Object {
      $relative = $_.FullName.Substring($PackageRoot.Length + 1).Replace('\', '/')
      "{0}  {1}" -f (Get-FileHash -LiteralPath $_.FullName -Algorithm SHA256).Hash.ToLowerInvariant(), "./$relative"
    }
  # GNU sha256sum treats CRLF as part of the filename.  Always write this
  # manifest with Unix LF line endings, even when packaging on Windows.
  [System.IO.File]::WriteAllText(
    (Join-Path $PackageRoot "SHA256SUMS"),
    (($checksums -join "`n") + "`n"),
    (New-Object System.Text.UTF8Encoding($false))
  )

  if (Test-Path -LiteralPath $Archive) { Remove-Item -LiteralPath $Archive -Force }
  & tar.exe -czf $Archive -C $WorkDir $PackageName
  if ($LASTEXITCODE -ne 0) { throw "tar packaging failed with exit code $LASTEXITCODE" }
  Write-Host "Linux deployment package created: $Archive"
} finally {
  if (Test-Path -LiteralPath $WorkDir) { Remove-Item -LiteralPath $WorkDir -Recurse -Force }
}
