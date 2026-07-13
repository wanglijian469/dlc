$ErrorActionPreference = "Stop"

$pidFile = ".\server.pid"
if (-not (Test-Path $pidFile)) {
  Write-Host "未找到 server.pid，服务可能未由本脚本启动。"
  exit 0
}

$processId = [int](Get-Content $pidFile -Raw)
$process = Get-Process -Id $processId -ErrorAction SilentlyContinue
if ($process) {
  Stop-Process -Id $processId -Force
  Write-Host "服务已停止，PID: $processId"
} else {
  Write-Host "PID $processId 已不存在。"
}
Remove-Item -Force $pidFile
