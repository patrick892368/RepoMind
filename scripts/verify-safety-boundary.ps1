param(
    [string]$RepoRoot = "."
)

$ErrorActionPreference = "Stop"

$root = (Resolve-Path $RepoRoot).Path
Push-Location $root
try {
    function Invoke-Git {
        param([string[]]$Arguments)

        $output = & git @Arguments 2>&1
        return [pscustomobject]@{
            ExitCode = $LASTEXITCODE
            Output = ($output -join "`n")
        }
    }

    $ignoredPaths = @(
        ".env",
        ".env.local",
        ".env.production",
        ".repomind",
        "dist",
        "eval",
        "benchmark",
        "repomind",
        "repomind.exe"
    )

    foreach ($path in $ignoredPaths) {
        $result = Invoke-Git @("check-ignore", "--quiet", "--", $path)
        if ($result.ExitCode -ne 0) {
            throw "Expected path to be ignored by git: $path"
        }
    }

    $allowedEnvExample = Invoke-Git @("check-ignore", "--quiet", "--", ".env.example")
    if ($allowedEnvExample.ExitCode -eq 0) {
        throw ".env.example must remain trackable"
    }

    $trackedRaw = & git ls-files -z
    if ($LASTEXITCODE -ne 0) {
        throw "git ls-files failed"
    }
    $tracked = @($trackedRaw -split "`0" | Where-Object { $_ })

    $forbiddenTracked = @()
    foreach ($path in $tracked) {
        $normalized = $path -replace "\\", "/"
        if ($normalized -eq ".env" -or
            ($normalized -like ".env.*" -and $normalized -ne ".env.example") -or
            $normalized -like ".repomind/*" -or
            $normalized -like "dist/*" -or
            $normalized -like "eval/*" -or
            $normalized -like "benchmark/*" -or
            $normalized -eq "repomind" -or
            $normalized -eq "repomind.exe") {
            $forbiddenTracked += $normalized
        }
    }
    if ($forbiddenTracked.Count -gt 0) {
        throw "Forbidden generated or secret files are tracked: $($forbiddenTracked -join ', ')"
    }

    $secretPattern = '(?i)\b[A-Z0-9_]*(?:API[_-]?KEY|TOKEN|SECRET)[A-Z0-9_]*\s*=\s*["'']?(sk-[A-Za-z0-9_-]{10,}|xai-[A-Za-z0-9_-]{10,}|gsk_[A-Za-z0-9_-]{10,}|AIza[A-Za-z0-9_-]{10,}|ghp_[A-Za-z0-9_]{10,}|glpat-[A-Za-z0-9_-]{10,})'
    $secretHits = @()
    foreach ($path in $tracked) {
        $normalized = $path -replace "\\", "/"
        if ($normalized -like "*.png" -or $normalized -like "*.ico" -or $normalized -like "*.jpg" -or $normalized -like "*.jpeg" -or $normalized -like "*.webp") {
            continue
        }
        try {
            $content = Get-Content -Path $normalized -Raw -ErrorAction Stop
        } catch {
            continue
        }
        if ($content -match $secretPattern) {
            $secretHits += $normalized
        }
    }
    if ($secretHits.Count -gt 0) {
        throw "Tracked files contain likely real API keys or tokens: $($secretHits -join ', ')"
    }

    $summary = [ordered]@{
        ok = $true
        ignored_paths_checked = $ignoredPaths
        tracked_file_count = $tracked.Count
        secret_pattern_checked = $true
    }
    [pscustomobject]$summary | ConvertTo-Json -Depth 4
} finally {
    Pop-Location
}
