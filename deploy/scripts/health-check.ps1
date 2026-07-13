param([string]$Url = "http://127.0.0.1:8080/api/health")

$ErrorActionPreference = "Stop"
try {
  $response = Invoke-RestMethod -Uri $Url -TimeoutSec 10
  if ($response.code -ne 0 -or $response.data.status -ne "ok") { throw "健康检查返回异常" }
  Write-Host "健康检查通过：$Url"
} catch {
  Write-Error "健康检查失败：$($_.Exception.Message)"
  exit 1
}
