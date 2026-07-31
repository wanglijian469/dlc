param(
  [Parameter(Mandatory = $true)]
  [string]$Archive
)

$ErrorActionPreference = "Stop"
$ArchivePath = (Resolve-Path $Archive).Path
if (-not $ArchivePath.EndsWith(".tar.gz")) { throw "Archive must end in .tar.gz" }
$WorkDir = Join-Path ([System.IO.Path]::GetTempPath()) ("dlc-linux-verify-" + [guid]::NewGuid().ToString("N"))
try {
  New-Item -ItemType Directory -Force $WorkDir | Out-Null
  & tar.exe -xzf $ArchivePath -C $WorkDir
  if ($LASTEXITCODE -ne 0) { throw "Archive extraction failed" }
  $roots = @(Get-ChildItem -LiteralPath $WorkDir -Directory)
  if ($roots.Count -ne 1) { throw "Archive must contain exactly one root directory" }
  $root = $roots[0].FullName
  @("server", "initdb", "VERSION", "SHA256SUMS", "README.md", "public\index.html", "public\favicon.png", "config\dlc.env.example", "systemd\dalu-parts.service", "nginx\dalu-parts.conf.template", "scripts\upgrade.sh", "docs\INSTALL.md", "docs\UPGRADE.md", "docs\RELEASE_NOTES.md") | ForEach-Object {
    if (-not (Test-Path -LiteralPath (Join-Path $root $_))) { throw "Missing package entry: $_" }
  }
  if ((Get-Item -LiteralPath (Join-Path $root "public\favicon.png")).Length -gt 32768) { throw "favicon.png must be 32 KiB or smaller" }
  $serverBytes = [System.IO.File]::ReadAllBytes((Join-Path $root "server"))
  $initdbBytes = [System.IO.File]::ReadAllBytes((Join-Path $root "initdb"))
  if ($serverBytes.Length -lt 4 -or $serverBytes[0] -ne 0x7f -or $serverBytes[1] -ne 0x45 -or $initdbBytes[0] -ne 0x7f -or $initdbBytes[1] -ne 0x45) { throw "server or initdb is not an ELF binary" }
  if ([System.IO.File]::ReadAllBytes((Join-Path $root "SHA256SUMS")) -contains 0x0d) { throw "SHA256SUMS must use Unix LF line endings" }
  Get-ChildItem -LiteralPath (Join-Path $root "scripts") -Filter "*.sh" -File | ForEach-Object {
    if ([System.IO.File]::ReadAllBytes($_.FullName) -contains 0x0d) { throw "Shell script must use Unix LF line endings: $($_.Name)" }
  }
  Get-Content -Encoding UTF8 -LiteralPath (Join-Path $root "SHA256SUMS") | ForEach-Object {
    if ($_ -notmatch '^([0-9a-f]{64})  \./(.+)$') { throw "Invalid checksum record: $_" }
    $expected = $Matches[1]
    $relative = $Matches[2]
    $actual = (Get-FileHash -LiteralPath (Join-Path $root $relative) -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actual -ne $expected) { throw "Checksum mismatch: $relative" }
  }
  Write-Host "Verified Linux deployment package: $ArchivePath"
} finally {
  if (Test-Path -LiteralPath $WorkDir) { Remove-Item -LiteralPath $WorkDir -Recurse -Force }
}
