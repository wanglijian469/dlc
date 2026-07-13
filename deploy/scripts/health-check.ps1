param([string]$Url = "http://127.0.0.1:8080/api/health")

$ErrorActionPreference = "Stop"
try {
  $response = Invoke-RestMethod -Uri $Url -TimeoutSec 10
  if ($response.code -ne 0 -or $response.data.status -ne "ok") { throw "Unexpected health response" }
  Write-Host "Health check passed: $Url"
} catch {
  Write-Error "Health check failed: $($_.Exception.Message)"
  exit 1
}
