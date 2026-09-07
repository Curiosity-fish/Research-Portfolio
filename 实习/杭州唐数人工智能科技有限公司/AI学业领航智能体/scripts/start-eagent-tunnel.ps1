param(
    [Parameter(Mandatory = $true)]
    [string]$SshTarget,

    [ValidateSet("api", "mysql")]
    [string]$Mode = "api",

    [string]$LocalEagentPort = "3001",
    [string]$RemoteEagentPort = "3001",

    [string]$LocalBackendPort = "8080",
    [string]$RemoteBackendPort = "18080",

    [string]$LocalMysqlPort = "3306",
    [string]$RemoteMysqlPort = "3306"
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command ssh -ErrorAction SilentlyContinue)) {
    throw "ssh command not found. Please install Windows OpenSSH Client."
}

$forwardArgs = @(
    "-N",
    "-L", "${LocalEagentPort}:localhost:${RemoteEagentPort}"
)

if ($Mode -eq "api") {
    $forwardArgs += @(
        "-R", "127.0.0.1:${RemoteBackendPort}:127.0.0.1:${LocalBackendPort}"
    )
    Write-Host "Mode: expose local backend API on remote port ${RemoteBackendPort}"
} else {
    $forwardArgs += @(
        "-R", "127.0.0.1:${RemoteMysqlPort}:127.0.0.1:${LocalMysqlPort}"
    )
    Write-Host "Mode: expose local MySQL on remote port ${RemoteMysqlPort} (temporary only)"
}

$keepAliveArgs = @(
    "-o", "ExitOnForwardFailure=yes",
    "-o", "ServerAliveInterval=30",
    "-o", "ServerAliveCountMax=3"
)

Write-Host "Opening SSH forward to $SshTarget. Press Ctrl+C to stop."
ssh @forwardArgs @keepAliveArgs $SshTarget
