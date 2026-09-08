param(
  [Parameter(Mandatory = $true)][string]$Archive
)

$ErrorActionPreference = "Stop"
$ArchivePath = [IO.Path]::GetFullPath($Archive)
if (-not $ArchivePath.EndsWith(".tar.gz", [StringComparison]::OrdinalIgnoreCase)) {
  throw "Archive must end in .tar.gz"
}
if (-not (Test-Path -LiteralPath $ArchivePath -PathType Leaf)) {
  throw "Archive does not exist: $ArchivePath"
}

$TaskTempRoot = Join-Path ([IO.Path]::GetTempPath()) ("dlc-linux-verify-" + [Guid]::NewGuid().ToString("N"))
try {
  New-Item -ItemType Directory -Force $TaskTempRoot | Out-Null
  & tar.exe -xzf $ArchivePath -C $TaskTempRoot
  if ($LASTEXITCODE -ne 0) { throw "Unable to extract archive" }

  $Roots = @(Get-ChildItem -Directory $TaskTempRoot)
  if ($Roots.Count -ne 1) { throw "Archive must contain exactly one root directory" }
  $PackageRoot = $Roots[0].FullName
  $Required = @(
    "server", "initdb", "VERSION", "SHA256SUMS", "README.md",
    "public/index.html", "config/dlc.env.example",
    "systemd/dalu-parts.service", "nginx/dalu-parts.conf.template",
    "scripts/install-layout.sh", "scripts/install-mysql-8.0.46.sh",
    "scripts/create-mysql-user.sh", "scripts/configure-site-url.sh",
    "scripts/render-nginx-config.sh", "scripts/migrate.sh",
    "scripts/database-upgrade.sh", "scripts/upgrade.sh", "scripts/start.sh",
    "scripts/stop.sh", "scripts/health-check.sh", "scripts/backup.sh",
    "docs/INSTALL.md", "docs/UPGRADE.md", "docs/RELEASE_NOTES.md",
    "docs/VENDOR_PROMOTION_UPGRADE.md"
  )
  foreach ($Relative in $Required) {
    if (-not (Test-Path -LiteralPath (Join-Path $PackageRoot $Relative))) {
      throw "Missing package entry: $Relative"
    }
  }

  foreach ($BinaryName in @("server", "initdb")) {
    $Bytes = [IO.File]::ReadAllBytes((Join-Path $PackageRoot $BinaryName))
    if ($Bytes.Length -lt 20 -or $Bytes[0] -ne 0x7f -or $Bytes[1] -ne 0x45 -or $Bytes[2] -ne 0x4c -or $Bytes[3] -ne 0x46) {
      throw "$BinaryName is not an ELF binary"
    }
    if ($Bytes[4] -ne 2 -or $Bytes[5] -ne 1 -or [BitConverter]::ToUInt16($Bytes, 18) -ne 62) {
      throw "$BinaryName is not a Linux ELF64 x86_64 binary"
    }
  }

  foreach ($Script in (Get-ChildItem (Join-Path $PackageRoot "scripts") -Filter '*.sh')) {
    $Bytes = [IO.File]::ReadAllBytes($Script.FullName)
    if ($Bytes.Length -lt 2 -or $Bytes[0] -ne 35 -or $Bytes[1] -ne 33 -or $Bytes -contains 13) {
      throw "Shell script must have a shebang and LF line endings without BOM: $($Script.Name)"
    }
  }

  $Forbidden = Get-ChildItem -Recurse -File $PackageRoot | Where-Object {
    $_.Name -eq ".env" -or $_.Extension -in @(".sql", ".db", ".sqlite", ".xls") -or
    ($_.Extension -eq ".xlsx" -and $_.FullName -notlike "*\public\templates\*")
  }
  if ($Forbidden) {
    throw "Package contains a secret, database, or unexpected user document: $($Forbidden[0].FullName)"
  }
  if (Test-Path -LiteralPath (Join-Path $PackageRoot "media_storage")) {
    throw "Package must not contain media_storage"
  }

  foreach ($Line in [IO.File]::ReadAllLines((Join-Path $PackageRoot "SHA256SUMS"))) {
    if ($Line -notmatch '^([0-9a-f]{64})  \./(.+)$') { throw "Invalid SHA256SUMS line: $Line" }
    $Expected = $Matches[1]
    $Relative = $Matches[2].Replace('/', [IO.Path]::DirectorySeparatorChar)
    $FilePath = Join-Path $PackageRoot $Relative
    if (-not (Test-Path -LiteralPath $FilePath -PathType Leaf)) { throw "Checksum file is missing: $Relative" }
    $Actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $FilePath).Hash.ToLowerInvariant()
    if ($Actual -ne $Expected) { throw "Checksum mismatch: $Relative" }
  }

  Write-Output "verified $ArchivePath"
} finally {
  if (Test-Path -LiteralPath $TaskTempRoot) {
    $ResolvedTemp = [IO.Path]::GetFullPath($TaskTempRoot)
    $ExpectedParent = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
    if ($ResolvedTemp.StartsWith($ExpectedParent, [StringComparison]::OrdinalIgnoreCase)) {
      Remove-Item -Recurse -Force -LiteralPath $ResolvedTemp
    }
  }
}
