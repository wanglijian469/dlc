param(
  [string]$EnvFile = ".\.env",
  [switch]$Foreground
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $EnvFile)) {
  if ($EnvFile -eq ".\.env" -and (Test-Path ".\default.env")) {
    $EnvFile = ".\default.env"
  } else {
    throw "Missing $EnvFile. Copy .env.example to .env, or use default.env for local testing."
  }
}

Get-Content $EnvFile | ForEach-Object {
  $line = $_.Trim()
  if ($line -and -not $line.StartsWith("#")) {
    $parts = $line.Split("=", 2)
    if ($parts.Length -eq 2) {
      [Environment]::SetEnvironmentVariable($parts[0], $parts[1], "Process")
    }
  }
}

if (-not $env:PUBLIC_DIR) {
  $env:PUBLIC_DIR = ".\public"
}
if (-not $env:MEDIA_DIR) {
  $env:MEDIA_DIR = ".\media_storage"
}

New-Item -ItemType Directory -Force $env:MEDIA_DIR, ".\logs" | Out-Null

if ($Foreground) {
  & ".\server.exe"
  exit $LASTEXITCODE
}

$pidFile = ".\server.pid"
if (Test-Path $pidFile) {
  $previousPid = [int](Get-Content $pidFile -Raw)
  if (Get-Process -Id $previousPid -ErrorAction SilentlyContinue) {
    throw "Service is already running. PID: $previousPid."
  }
  Remove-Item -Force $pidFile
}

$process = Start-Process -FilePath ".\server.exe" -WorkingDirectory (Get-Location) -WindowStyle Hidden -PassThru -RedirectStandardOutput ".\logs\server.out.log" -RedirectStandardError ".\logs\server.err.log"
Set-Content -Encoding ascii $pidFile $process.Id
Write-Host "Service started. PID: $($process.Id). Run .\scripts\health-check.ps1 to verify it."
