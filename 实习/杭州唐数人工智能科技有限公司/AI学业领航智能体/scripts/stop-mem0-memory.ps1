$ErrorActionPreference = 'Stop'
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$serverRoot = Join-Path $workspaceRoot 'outputs\runtime\mem0\mem0-main\server'

if (-not (Test-Path -LiteralPath $serverRoot)) {
    Write-Host '没有发现本项目启动的 Mem0 运行目录。'
    exit 0
}

Push-Location $serverRoot
try {
    docker compose down
} finally {
    Pop-Location
}
