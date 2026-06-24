param(
    [string]$Provider = "grok",
    [string]$Model = "",
    [string]$RepoPath = ".",
    [string]$OutputDir = "",
    [string]$Language = "en",
    [string]$Proxy = "",
    [int]$TimeoutSeconds = 120
)

$ErrorActionPreference = "Stop"

function ConvertTo-ProcessArgument {
    param([string]$Argument)

    if ($null -eq $Argument) {
        return '""'
    }
    if ($Argument -notmatch '[\s"]') {
        return $Argument
    }
    return '"' + ($Argument -replace '"', '\"') + '"'
}

function Invoke-CapturedCommand {
    param(
        [string]$FilePath,
        [string[]]$ArgumentList,
        [string]$LogPath,
        [string]$WorkingDirectory = (Get-Location).Path,
        [int]$TimeoutSeconds = 120
    )

    $psi = [System.Diagnostics.ProcessStartInfo]::new()
    $psi.FileName = $FilePath
    $psi.Arguments = ($ArgumentList | ForEach-Object { ConvertTo-ProcessArgument $_ }) -join " "
    $psi.WorkingDirectory = (Resolve-Path $WorkingDirectory).Path
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.UseShellExecute = $false
    $psi.CreateNoWindow = $true

    $process = [System.Diagnostics.Process]::new()
    $process.StartInfo = $psi
    [void]$process.Start()

    $stdoutTask = $process.StandardOutput.ReadToEndAsync()
    $stderrTask = $process.StandardError.ReadToEndAsync()

    if (-not $process.WaitForExit($TimeoutSeconds * 1000)) {
        try {
            $process.Kill($true)
        } catch {
            $process.Kill()
        }
        throw "$FilePath timed out after $TimeoutSeconds seconds"
    }

    $stdout = $stdoutTask.GetAwaiter().GetResult()
    $stderr = $stderrTask.GetAwaiter().GetResult()
    $log = @()
    if ($stdout) {
        $log += "STDOUT:"
        $log += $stdout.TrimEnd()
    }
    if ($stderr) {
        $log += "STDERR:"
        $log += $stderr.TrimEnd()
    }
    $log | Set-Content -Path $LogPath -Encoding UTF8

    if ($process.ExitCode -ne 0) {
        $message = "$FilePath exited with code $($process.ExitCode)"
        if ($stderr) {
            $message = "$message`: $($stderr.Trim())"
        }
        throw $message
    }

    return @{
        Stdout = $stdout
        Stderr = $stderr
    }
}

if (-not $OutputDir) {
    $OutputDir = Join-Path "eval" "ai-smoke-$Provider"
}

$repoRoot = (Resolve-Path $RepoPath).Path
if ([System.IO.Path]::IsPathRooted($OutputDir)) {
    $analysisOutputDir = $OutputDir
} else {
    $analysisOutputDir = Join-Path $repoRoot $OutputDir
}
New-Item -ItemType Directory -Force -Path $analysisOutputDir | Out-Null

$oldHTTPSProxy = $env:HTTPS_PROXY
$oldHTTPProxy = $env:HTTP_PROXY
$oldALLProxy = $env:ALL_PROXY
try {
    if ($Proxy) {
        $env:HTTPS_PROXY = $Proxy
        $env:HTTP_PROXY = $Proxy
        $env:ALL_PROXY = $Proxy
    }

    $args = @(
        "run", "./cmd/repomind",
        "analyze",
        "--ai", $Provider,
        "--output", $analysisOutputDir,
        "--lang", $Language
    )
    if ($Model) {
        $args += @("--ai-model", $Model)
    }
    $args += $repoRoot

    $logPath = Join-Path $analysisOutputDir "ai-smoke.log"
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    Invoke-CapturedCommand -FilePath "go" -ArgumentList $args -LogPath $logPath -TimeoutSeconds $TimeoutSeconds | Out-Null
    $stopwatch.Stop()

    $analysisPath = Join-Path $analysisOutputDir "analysis.json"
    $analysis = Get-Content $analysisPath -Raw | ConvertFrom-Json
    $summary = [ordered]@{
        ok = $true
        provider = $Provider
        model = $Model
        language = $analysis.language
        repo_path = $repoRoot
        output_dir = $analysisOutputDir
        duration_seconds = [math]::Round($stopwatch.Elapsed.TotalSeconds, 2)
        title = $analysis.summary.title
        overview = $analysis.summary.overview
        modules = $analysis.summary.modules
        stack = $analysis.summary.stack
    }
    $summaryPath = Join-Path $analysisOutputDir "ai-smoke-summary.json"
    [pscustomobject]$summary | ConvertTo-Json -Depth 8 | Set-Content -Path $summaryPath -Encoding UTF8

    Write-Host "AI provider smoke test passed: $Provider"
    Write-Host "Summary written to $summaryPath"
    Write-Host "Log written to $logPath"
} finally {
    $env:HTTPS_PROXY = $oldHTTPSProxy
    $env:HTTP_PROXY = $oldHTTPProxy
    $env:ALL_PROXY = $oldALLProxy
}
