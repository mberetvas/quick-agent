# install.ps1 — Install or uninstall quick-agent as a Windows Scheduled Task.
#
# Usage:
#   .\install.ps1             Install the service (runs at logon)
#   .\install.ps1 -Uninstall  Remove the scheduled task

#Requires -Version 5.1
param(
    [switch]$Uninstall
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$TaskName   = 'quick-agent'
$BinaryName = 'quick-agent.exe'
$InstallDir = Join-Path $env:LOCALAPPDATA 'quick-agent'
$ScriptRoot = Split-Path -Parent $PSScriptRoot
$BinarySrc  = Join-Path $ScriptRoot $BinaryName

function Write-Info([string]$Msg) { Write-Host "• $Msg" }
function Write-Err([string]$Msg)  { Write-Host "error: $Msg" -ForegroundColor Red; exit 1 }

# ─── uninstall ───────────────────────────────────────────────────────────────

function Uninstall-Service {
    $task = Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
    if ($task) {
        Stop-ScheduledTask  -TaskName $TaskName -ErrorAction SilentlyContinue
        Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
        Write-Info "Scheduled task '$TaskName' removed."
    } else {
        Write-Info "Scheduled task '$TaskName' not found — nothing to remove."
    }
    $installedBin = Join-Path $InstallDir $BinaryName
    if (Test-Path $installedBin) {
        Remove-Item $installedBin -Force
        Write-Info "Removed $installedBin"
    }
    Write-Info 'Uninstall complete.'
}

# ─── install ─────────────────────────────────────────────────────────────────

function Install-Service {
    # Resolve binary
    if (-not (Test-Path $BinarySrc)) {
        Write-Info "Binary not found at $BinarySrc; attempting build..."
        Push-Location $ScriptRoot
        try {
            go build -o $BinarySrc .\cmd\quick-agent\
            if ($LASTEXITCODE -ne 0) { Write-Err "Build failed. Run 'go build .\cmd\quick-agent\' first." }
        } finally {
            Pop-Location
        }
    }

    # Copy binary
    if (-not (Test-Path $InstallDir)) { New-Item -ItemType Directory -Path $InstallDir | Out-Null }
    Copy-Item $BinarySrc (Join-Path $InstallDir $BinaryName) -Force
    Write-Info "Installed binary to $InstallDir\$BinaryName"

    # Remove existing task if present
    $existing = Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
    if ($existing) {
        Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
    }

    # Create AtLogon trigger
    $action  = New-ScheduledTaskAction -Execute (Join-Path $InstallDir $BinaryName) -Argument 'daemon'
    $trigger = New-ScheduledTaskTrigger -AtLogOn
    $settings = New-ScheduledTaskSettingsSet `
        -ExecutionTimeLimit (New-TimeSpan -Hours 0) `
        -RestartCount 3 `
        -RestartInterval (New-TimeSpan -Minutes 1) `
        -StartWhenAvailable

    Register-ScheduledTask `
        -TaskName $TaskName `
        -Action   $action  `
        -Trigger  $trigger `
        -Settings $settings `
        -RunLevel Limited `
        -Force | Out-Null

    # Start it now
    Start-ScheduledTask -TaskName $TaskName
    Write-Info "Scheduled task '$TaskName' registered and started."
    Write-Info 'Install complete. The daemon will start automatically on login.'
}

# ─── entry point ─────────────────────────────────────────────────────────────

if ($Uninstall) {
    Uninstall-Service
} else {
    Install-Service
}
