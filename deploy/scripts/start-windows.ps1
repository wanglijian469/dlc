param(
  [string]$EnvFile = ".\.env"
)

$ErrorActionPreference = "Stop"

if (Test-Path $EnvFile) {
  Get-Content $EnvFile | ForEach-Object {
    $line = $_.Trim()
    if ($line -and -not $line.StartsWith("#")) {
      $parts = $line.Split("=", 2)
      if ($parts.Length -eq 2) {
        [Environment]::SetEnvironmentVariable($parts[0], $parts[1], "Process")
      }
    }
  }
}

if (-not $env:PUBLIC_DIR) {
  $env:PUBLIC_DIR = ".\frontend\dist"
}

.\server.exe
