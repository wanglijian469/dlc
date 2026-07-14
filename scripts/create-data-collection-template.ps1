$ErrorActionPreference = 'Stop'
$script = Join-Path $PSScriptRoot 'generate-data-import-template.mjs'
node $script
