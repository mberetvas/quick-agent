# Build quick-agent for the current Windows platform.
# Cross-OS builds use GitHub Actions (.github/workflows/release.yml).
param(
    [string]$Version = "dev",
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Pkg = "./cmd/quick-agent"
$Ldflags = "-s -w -X github.com/mberetvas/quick-agent/internal/version.Version=$Version"

$Out = Join-Path $Root $OutputDir
New-Item -ItemType Directory -Force -Path $Out | Out-Null

$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$Name = "quick-agent-windows-$Arch.exe"
$Dest = Join-Path $Out $Name

Write-Host "Building $Name..."
Push-Location $Root
try {
    $env:CGO_ENABLED = "1"
    go build -ldflags $Ldflags -o $Dest $Pkg
} finally {
    Pop-Location
}

Write-Host "Build complete. Binary: $Dest"
