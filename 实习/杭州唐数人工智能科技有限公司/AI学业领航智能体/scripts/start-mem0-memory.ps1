param(
    [Parameter(Mandatory = $true)]
    [string]$OpenAiApiKey,

    [Parameter(Mandatory = $true)]
    [string]$Mem0ApiKey,

    [Parameter(Mandatory = $true)]
    [string]$IdentitySecret,

    [string]$ArchivePath = 'C:\Users\BenBen\Downloads\mem0-main.zip'
)

$ErrorActionPreference = 'Stop'

function New-RandomSecret([int]$bytes = 32) {
    $buffer = New-Object byte[] $bytes
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($buffer)
    return [Convert]::ToBase64String($buffer)
}

if (-not (Test-Path -LiteralPath $ArchivePath)) {
    throw "找不到 Mem0 源码包：$ArchivePath"
}

try {
    docker version --format '{{.Server.Version}}' | Out-Null
} catch {
    throw 'Docker Desktop 未运行。请先启动 Docker Desktop，等待其状态显示 Running 后再执行此脚本。'
}

$workspaceRoot = Split-Path -Parent $PSScriptRoot
$runtimeRoot = Join-Path $workspaceRoot 'outputs\runtime\mem0'
$sourceRoot = Join-Path $runtimeRoot 'mem0-main'
$serverRoot = Join-Path $sourceRoot 'server'

if (-not (Test-Path -LiteralPath $serverRoot)) {
    New-Item -ItemType Directory -Force -Path $runtimeRoot | Out-Null
    Expand-Archive -LiteralPath $ArchivePath -DestinationPath $runtimeRoot -Force
}

$envFile = Join-Path $serverRoot '.env'
$postgresPassword = New-RandomSecret
$jwtSecret = New-RandomSecret 48

@"
OPENAI_API_KEY=$OpenAiApiKey
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=postgres
POSTGRES_USER=postgres
POSTGRES_PASSWORD=$postgresPassword
POSTGRES_COLLECTION_NAME=academic_navigation_memories
ADMIN_API_KEY=$Mem0ApiKey
JWT_SECRET=$jwtSecret
AUTH_DISABLED=false
DASHBOARD_URL=http://localhost:3000
APP_DB_NAME=mem0_app
MEM0_DEFAULT_LLM_MODEL=gpt-5-mini
MEM0_DEFAULT_EMBEDDER_MODEL=text-embedding-3-small
MEM0_TELEMETRY=false
REQUEST_LOG_RETENTION_DAYS=30
"@ | Set-Content -LiteralPath $envFile -Encoding ascii

Push-Location $serverRoot
try {
    docker compose up -d --build
    $deadline = (Get-Date).AddMinutes(5)
    do {
        Start-Sleep -Seconds 3
        try {
            $response = Invoke-WebRequest -UseBasicParsing -TimeoutSec 5 -Uri 'http://localhost:8888/docs'
            if ($response.StatusCode -eq 200) { break }
        } catch {
            # Mem0 may still be building or waiting for Postgres.
        }
    } while ((Get-Date) -lt $deadline)

    $health = Invoke-WebRequest -UseBasicParsing -TimeoutSec 10 -Uri 'http://localhost:8888/docs'
    if ($health.StatusCode -ne 200) {
        throw 'Mem0 未在 5 分钟内启动。请执行 docker compose logs mem0 查看日志。'
    }
} finally {
    Pop-Location
}

Write-Host ''
Write-Host 'Mem0 已启动： http://localhost:8888/docs'
Write-Host '请在 Spring Boot 的安全配置来源中设置以下值：'
Write-Host 'MEM0_ENABLED=true'
Write-Host 'MEM0_BASE_URL=http://localhost:8888'
Write-Host "MEM0_API_KEY=$Mem0ApiKey"
Write-Host "MEM0_IDENTITY_SECRET=$IdentitySecret"
Write-Host ''
Write-Host '注意：.env 位于 outputs/runtime/，请勿提交或共享其中的凭据。'
