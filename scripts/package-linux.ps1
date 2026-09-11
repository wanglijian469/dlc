param(
  [string]$Version = "v2026.09.11-linux.1"
)

$ErrorActionPreference = "Stop"

if ($Version -notmatch '^[A-Za-z0-9._-]+$') {
  throw "Version may contain only letters, numbers, dots, underscores, and hyphens."
}

$Root = Split-Path -Parent $PSScriptRoot
$PackageName = "dlc-deploy-linux-x86_64-$Version"
$ReleaseDir = Join-Path $Root "release"
$Output = Join-Path $ReleaseDir "$PackageName.tar.gz"
$TaskTempRoot = Join-Path ([IO.Path]::GetTempPath()) ("dlc-linux-package-" + [Guid]::NewGuid().ToString("N"))
$PackageRoot = Join-Path $TaskTempRoot $PackageName
$Utf8NoBom = [Text.UTF8Encoding]::new($false)

function Invoke-Checked {
  param(
    [Parameter(Mandatory = $true)][string]$Command,
    [Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments
  )
  & $Command @Arguments
  if ($LASTEXITCODE -ne 0) {
    throw "$Command failed with exit code $LASTEXITCODE"
  }
}

try {
  New-Item -ItemType Directory -Force $ReleaseDir | Out-Null
  @("public", "config", "systemd", "nginx", "scripts", "docs") | ForEach-Object {
    New-Item -ItemType Directory -Force (Join-Path $PackageRoot $_) | Out-Null
  }

  Push-Location (Join-Path $Root "frontend")
  try {
    Invoke-Checked -Command npm -Arguments @("run", "build")
  } finally {
    Pop-Location
  }

  $PreviousGOOS = $env:GOOS
  $PreviousGOARCH = $env:GOARCH
  $PreviousCGOEnabled = $env:CGO_ENABLED
  $env:GOOS = "linux"
  $env:GOARCH = "amd64"
  $env:CGO_ENABLED = "0"
  Push-Location (Join-Path $Root "backend")
  try {
    Invoke-Checked -Command go -Arguments @("build", "-trimpath", "-buildvcs=false", "-ldflags", "-s -w", "-o", (Join-Path $PackageRoot "server"), "./cmd/server")
    Invoke-Checked -Command go -Arguments @("build", "-trimpath", "-buildvcs=false", "-ldflags", "-s -w", "-o", (Join-Path $PackageRoot "initdb"), "./cmd/initdb")
  } finally {
    Pop-Location
    $env:GOOS = $PreviousGOOS
    $env:GOARCH = $PreviousGOARCH
    $env:CGO_ENABLED = $PreviousCGOEnabled
  }

  Copy-Item -Recurse -Force (Join-Path $Root "frontend\dist\*") (Join-Path $PackageRoot "public")
  Copy-Item -Force (Join-Path $Root "deploy\linux\config\dlc.env.example") (Join-Path $PackageRoot "config")
  Copy-Item -Force (Join-Path $Root "deploy\linux\systemd\dalu-parts.service") (Join-Path $PackageRoot "systemd")
  Copy-Item -Force (Join-Path $Root "deploy\linux\nginx\dalu-parts.conf.template") (Join-Path $PackageRoot "nginx")
  Copy-Item -Force (Join-Path $Root "deploy\linux\scripts\*.sh") (Join-Path $PackageRoot "scripts")
  Copy-Item -Force (Join-Path $Root "deploy\linux\README.md") (Join-Path $PackageRoot "README.md")
  Copy-Item -Force (Join-Path $Root "deploy\linux\INSTALL.md") (Join-Path $PackageRoot "docs")
  Copy-Item -Force (Join-Path $Root "deploy\linux\UPGRADE.md") (Join-Path $PackageRoot "docs")
  Copy-Item -Force (Join-Path $Root "deploy\linux\VENDOR_PROMOTION_UPGRADE.md") (Join-Path $PackageRoot "docs")
  Copy-Item -Force (Join-Path $Root "deploy\linux\RELEASE_NOTES.md") (Join-Path $PackageRoot "docs")
  [IO.File]::WriteAllText((Join-Path $PackageRoot "VERSION"), "$Version`n", $Utf8NoBom)

  # Windows checkouts may use CRLF; Linux shell/config files must use LF without BOM.
  Get-ChildItem -Recurse -File $PackageRoot | Where-Object {
    $_.Extension -in @(".sh", ".md", ".example", ".service", ".template")
  } | ForEach-Object {
    $Content = [IO.File]::ReadAllText($_.FullName).Replace("`r`n", "`n")
    [IO.File]::WriteAllText($_.FullName, $Content, $Utf8NoBom)
  }

  $ChecksumLines = Get-ChildItem -Recurse -File $PackageRoot |
    Where-Object Name -ne "SHA256SUMS" |
    ForEach-Object {
      $Relative = $_.FullName.Substring($PackageRoot.Length).TrimStart('\', '/').Replace('\', '/')
      [PSCustomObject]@{ Relative = $Relative; FullName = $_.FullName }
    } |
    Sort-Object Relative |
    ForEach-Object {
      $Hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $_.FullName).Hash.ToLowerInvariant()
      "$Hash  ./$($_.Relative)"
    }
  [IO.File]::WriteAllText((Join-Path $PackageRoot "SHA256SUMS"), (($ChecksumLines -join "`n") + "`n"), $Utf8NoBom)

  if (Test-Path -LiteralPath $Output) {
    Remove-Item -Force -LiteralPath $Output
  }
  Invoke-Checked -Command tar.exe -Arguments @("-czf", $Output, "-C", $TaskTempRoot, $PackageName)
  $ArchiveHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $Output).Hash.ToLowerInvariant()
  [IO.File]::WriteAllText("$Output.sha256", "$ArchiveHash  $PackageName.tar.gz`n", $Utf8NoBom)
  Write-Output $Output
} finally {
  if (Test-Path -LiteralPath $TaskTempRoot) {
    $ResolvedTemp = [IO.Path]::GetFullPath($TaskTempRoot)
    $ExpectedParent = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
    if ($ResolvedTemp.StartsWith($ExpectedParent, [StringComparison]::OrdinalIgnoreCase)) {
      Remove-Item -Recurse -Force -LiteralPath $ResolvedTemp
    }
  }
}
