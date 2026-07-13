param(
  [string]$EnvFile = ".\.env",
  [string]$MySqlDump = "mysqldump"
)

$ErrorActionPreference = "Stop"
if (-not (Test-Path $EnvFile)) { throw "Missing $EnvFile" }

$settings = @{}
Get-Content $EnvFile | ForEach-Object {
  $line = $_.Trim()
  if ($line -and -not $line.StartsWith("#")) {
    $parts = $line.Split("=", 2)
    if ($parts.Length -eq 2) { $settings[$parts[0]] = $parts[1] }
  }
}

$stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$backupDir = ".\backups\$stamp"
New-Item -ItemType Directory -Force $backupDir | Out-Null
& $MySqlDump "-h$($settings.DB_HOST)" "-P$($settings.DB_PORT)" "-u$($settings.DB_USER)" "-p$($settings.DB_PASSWORD)" "--single-transaction" "--routines" "--events" $settings.DB_NAME "--result-file=$backupDir\database.sql"
if ($LASTEXITCODE -ne 0) { throw "Database export failed" }
if (Test-Path ".\media_storage") { Copy-Item -Recurse -Force ".\media_storage" "$backupDir\media_storage" }
Write-Host "Backup completed: $backupDir"
