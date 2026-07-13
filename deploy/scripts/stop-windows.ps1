$ErrorActionPreference = "Stop"

$pidFile = ".\server.pid"
if (-not (Test-Path $pidFile)) {
  Write-Host "server.pid was not found. The service may not have been started by this script."
  exit 0
}

$processId = [int](Get-Content $pidFile -Raw)
$process = Get-Process -Id $processId -ErrorAction SilentlyContinue
if ($process) {
  Stop-Process -Id $processId -Force
  Write-Host "Service stopped. PID: $processId"
} else {
  Write-Host "PID $processId no longer exists."
}
Remove-Item -Force $pidFile
