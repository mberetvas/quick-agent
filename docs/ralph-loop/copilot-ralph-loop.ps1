<#
.SYNOPSIS
    Implements all issues in a folder sequentially using the Copilot CLI.

.DESCRIPTION
    Iterates over every Markdown issue file (sorted by name, README excluded)
    in the specified folder and invokes `copilot` for each one in non-interactive
    autopilot mode. Each issue is attached to the prompt so the model receives
    the full specification before making changes.

.PARAMETER IssuesFolder
    Path to the folder that contains the issue .md files (no subfolders are scanned).

.PARAMETER Model
    Copilot model to use. Defaults to claude-sonnet-4.6.

.PARAMETER Effort
    Reasoning effort level passed to Copilot. Choices: low, medium, high, xhigh.
    Defaults to high.

.EXAMPLE
    .\Invoke-IssueLoop.ps1 -IssuesFolder "Docs\issues\logging-v2-hardening"
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [string] $IssuesFolder,

    [string] $Model = 'claude-sonnet-4.6',

    [ValidateSet('low', 'medium', 'high', 'xhigh')]
    [string] $Effort = 'medium'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# ── Resolve folder ────────────────────────────────────────────────────────────

$resolvedFolder = Resolve-Path -LiteralPath $IssuesFolder -ErrorAction Stop |
    Select-Object -ExpandProperty Path

Write-Host "[INFO] Issues folder : $resolvedFolder"
Write-Host "[INFO] Model         : $Model"
Write-Host "[INFO] Effort        : $Effort"
Write-Host ""

# ── Collect issue files (shallow, sorted, README excluded) ────────────────────

$issueFiles = Get-ChildItem -LiteralPath $resolvedFolder -File -Filter '*.md' |
    Where-Object { $_.Name -notmatch '^README(\.md)?$' } |
    Sort-Object Name

if ($issueFiles.Count -eq 0) {
    Write-Warning "No issue files found in '$resolvedFolder'. Nothing to do."
    exit 0
}

Write-Host "[INFO] Found $($issueFiles.Count) issue(s) to process."
Write-Host ""

# ── Process each issue ────────────────────────────────────────────────────────

$index = 0
foreach ($file in $issueFiles) {
    $index++
    $label = "$index/$($issueFiles.Count)"

    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    Write-Host "[$label] Starting : $($file.Name)"
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    $issueContent = Get-Content -LiteralPath $file.FullName -Raw

    $prompt = @"
Implement the issue below exactly as specified.
Follow all acceptance criteria. Use TDD: write failing tests first, then make them pass.
Do not change code outside the scope of this issue.

--- ISSUE: $($file.Name) ---
$issueContent
--- END ISSUE ---
"@

    copilot `
        --model             $Model `
        --effort            $Effort `
        --autopilot `
        --allow-all-tools `
        --allow-all-paths `
        --allow-all-urls `
        --no-ask-user `
        --prompt            $prompt

    $exitCode = $LASTEXITCODE

    if ($exitCode -ne 0) {
        Write-Error "[$label] Copilot exited with code $exitCode for '$($file.Name)'. Aborting loop."
        exit $exitCode
    }

    Write-Host ""
    Write-Host "[$label] Completed : $($file.Name)"
    Write-Host ""
}

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
Write-Host "[INFO] All $($issueFiles.Count) issue(s) processed successfully."
