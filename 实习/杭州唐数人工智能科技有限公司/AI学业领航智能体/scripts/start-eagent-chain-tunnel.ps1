param(
    [ValidateSet("Start", "Stop", "Status")]
    [string]$Action = "Start",

    [string]$JumpHost = "node1-frp",
    [string]$EagentHost = "172.20.20.35",
    [string]$EagentUser = "root",

    [string]$LocalBackendPort = "8080",
    [string]$RemoteBackendPort = "18080"
)

$ErrorActionPreference = "Stop"
$ssh = "C:\WINDOWS\System32\OpenSSH\ssh.exe"
$localPidFile = Join-Path $env:TEMP "eagent-chain-local.pid"
$remotePattern = "ssh -fN.*$EagentHost.*$RemoteBackendPort"

if (-not (Test-Path $ssh)) {
    throw "Windows OpenSSH client not found."
}

function Test-Node1 {
    $result = & $ssh -o BatchMode=yes -o ConnectTimeout=8 $JumpHost "curl -s -m 8 -o /dev/null -w '%{http_code}' http://127.0.0.1:${RemoteBackendPort}/api/eagent/health"
    return "$result"
}

function Test-EagentHost {
    $remote = "ssh -o BatchMode=yes -o ConnectTimeout=5 ${EagentUser}@${EagentHost} curl -s -m 8 -o /dev/null -w '%{http_code}' http://127.0.0.1:${RemoteBackendPort}/api/eagent/health"
    $result = & $ssh -o BatchMode=yes -o ConnectTimeout=8 $JumpHost $remote
    return "$result"
}

function Stop-RemoteForward {
    $remote = "pkill -f 'ssh -fN.*$EagentHost.*$RemoteBackendPort' 2>/dev/null || true"
    & $ssh -o BatchMode=yes -o ConnectTimeout=8 $JumpHost $remote | Out-Null
}

function Stop-LocalForward {
    if (Test-Path $localPidFile) {
        $localPid = Get-Content -LiteralPath $localPidFile -ErrorAction SilentlyContinue
        if ($localPid -and (Get-Process -Id $localPid -ErrorAction SilentlyContinue)) {
            Stop-Process -Id $localPid -Force
        }
        Remove-Item -LiteralPath $localPidFile -Force -ErrorAction SilentlyContinue
    }
}

switch ($Action) {
    "Start" {
        Stop-RemoteForward
        Stop-LocalForward

        $localArgs = @(
            "-N",
            "-R", "127.0.0.1:${RemoteBackendPort}:127.0.0.1:${LocalBackendPort}",
            "-o", "ExitOnForwardFailure=yes",
            "-o", "ServerAliveInterval=30",
            "-o", "ServerAliveCountMax=3",
            $JumpHost
        )

        $localProcess = Start-Process -FilePath $ssh -ArgumentList $localArgs -WindowStyle Hidden -PassThru
        $localProcess.Id | Set-Content -LiteralPath $localPidFile
        Start-Sleep -Seconds 2

        $node1Status = Test-Node1
        Write-Host "node1 -> local backend: HTTP $node1Status"

        $remoteCommand = "ssh -fN -o BatchMode=yes -o ExitOnForwardFailure=yes -o ServerAliveInterval=30 -o ServerAliveCountMax=3 -R 127.0.0.1:${RemoteBackendPort}:127.0.0.1:${RemoteBackendPort} ${EagentUser}@${EagentHost}"
        & $ssh -o BatchMode=yes -o ConnectTimeout=8 $JumpHost $remoteCommand | Out-Null
        Start-Sleep -Seconds 2

        $eagentStatus = Test-EagentHost
        Write-Host "E-Agent host -> local backend: HTTP $eagentStatus"
    }

    "Stop" {
        Stop-RemoteForward
        Stop-LocalForward
        Write-Host "Reverse tunnel chain stopped."
    }

    "Status" {
        $node1Status = Test-Node1
        $eagentStatus = Test-EagentHost
        Write-Host "node1 -> local backend: HTTP $node1Status"
        Write-Host "E-Agent host -> local backend: HTTP $eagentStatus"
    }
}
