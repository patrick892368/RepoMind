param(
    [string]$OutputDir = "eval/preflight",
    [int]$TimeoutSeconds = 300,
    [string]$Proxy = "",
    [switch]$IncludeBenchmark,
    [switch]$IncludeEvaluation,
    [switch]$IncludeAskEvaluation,
    [switch]$IncludeAISmoke,
    [switch]$IncludeRemoteAnalyzeSmoke,
    [switch]$IncludeReleaseSmoke,
    [switch]$IncludeManifestBuild,
    [string]$AIProvider = "grok",
    [string]$AIModel = "grok-4.3",
    [string]$AskProvider = "offline",
    [string]$AskModel = "",
    [string]$AskCasesPath = "",
    [switch]$AskStrict,
    [double]$MinimumAskScore = 1.0,
    [int]$BenchmarkTargetSeconds = 30,
    [double]$MinimumEvaluationQualityScore = 1.0,
    [int]$CloneRetries = 3,
    [string]$RepoCacheDir = "",
    [string]$RemoteAnalyzeRepo = "https://github.com/spring-guides/gs-rest-service.git",
    [string]$ManifestVersion = "v0.0.0-preflight"
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
        [int]$TimeoutSeconds = 300,
        [hashtable]$EnvironmentVariables = @{}
    )

    $psi = [System.Diagnostics.ProcessStartInfo]::new()
    $psi.FileName = $FilePath
    $psi.Arguments = ($ArgumentList | ForEach-Object { ConvertTo-ProcessArgument $_ }) -join " "
    $psi.WorkingDirectory = (Resolve-Path $WorkingDirectory).Path
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.UseShellExecute = $false
    $psi.CreateNoWindow = $true
    foreach ($key in $EnvironmentVariables.Keys) {
        $psi.Environment[$key] = [string]$EnvironmentVariables[$key]
    }

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
}

function Invoke-PreflightStep {
    param(
        [string]$Name,
        [scriptblock]$Action
    )

    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $item = [ordered]@{
        name = $Name
        ok = $false
        duration_seconds = $null
        error = ""
    }
    try {
        & $Action
        $item.ok = $true
    } catch {
        $item.error = $_.Exception.Message
    } finally {
        $stopwatch.Stop()
        $item.duration_seconds = [math]::Round($stopwatch.Elapsed.TotalSeconds, 2)
    }
    return [pscustomobject]$item
}

if (-not $Proxy) {
    if ($env:HTTPS_PROXY) {
        $Proxy = $env:HTTPS_PROXY
    } elseif ($env:HTTP_PROXY) {
        $Proxy = $env:HTTP_PROXY
    } elseif ($env:ALL_PROXY) {
        $Proxy = $env:ALL_PROXY
    }
}

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
$outputRoot = (Resolve-Path $OutputDir).Path
$sharedRepoCache = if ($RepoCacheDir) { $RepoCacheDir } else { Join-Path $outputRoot "repo-cache" }

$steps = @()
$steps += Invoke-PreflightStep -Name "safety boundary" -Action {
    Invoke-CapturedCommand -FilePath "powershell" -ArgumentList @("-ExecutionPolicy", "Bypass", "-File", "scripts\verify-safety-boundary.ps1") -LogPath (Join-Path $outputRoot "safety-boundary.log") -TimeoutSeconds $TimeoutSeconds
}
$steps += Invoke-PreflightStep -Name "go test ./..." -Action {
    Invoke-CapturedCommand -FilePath "go" -ArgumentList @("test", "./...") -LogPath (Join-Path $outputRoot "go-test.log") -TimeoutSeconds $TimeoutSeconds
}
$steps += Invoke-PreflightStep -Name "go vet ./..." -Action {
    Invoke-CapturedCommand -FilePath "go" -ArgumentList @("vet", "./...") -LogPath (Join-Path $outputRoot "go-vet.log") -TimeoutSeconds $TimeoutSeconds
}
$steps += Invoke-PreflightStep -Name "analyze smoke en" -Action {
    Invoke-CapturedCommand -FilePath "go" -ArgumentList @("run", "./cmd/repomind", "analyze", "--output", (Join-Path $outputRoot "analyze-en"), "--lang", "en", ".") -LogPath (Join-Path $outputRoot "analyze-en.log") -TimeoutSeconds $TimeoutSeconds
}
$steps += Invoke-PreflightStep -Name "analyze smoke zh" -Action {
    Invoke-CapturedCommand -FilePath "go" -ArgumentList @("run", "./cmd/repomind", "analyze", "--output", (Join-Path $outputRoot "analyze-zh"), "--lang", "zh", ".") -LogPath (Join-Path $outputRoot "analyze-zh.log") -TimeoutSeconds $TimeoutSeconds
}
$steps += Invoke-PreflightStep -Name "trace and diagnose smoke" -Action {
    $smokeDir = Join-Path $outputRoot "trace-diagnose"
    New-Item -ItemType Directory -Force -Path $smokeDir | Out-Null
    $fixtureRepo = "testdata\fixtures\diagnose-repo"
    $analysisOutput = Join-Path $smokeDir "analysis"
    $analysisPath = Join-Path $analysisOutput "analysis.json"
    $traceLog = Join-Path $smokeDir "trace.log"
    $diagnoseLog = Join-Path $smokeDir "diagnose.log"

    Invoke-CapturedCommand -FilePath "go" -ArgumentList @("run", "./cmd/repomind", "analyze", "--output", $analysisOutput, $fixtureRepo) -LogPath (Join-Path $smokeDir "analyze.log") -TimeoutSeconds $TimeoutSeconds
    Invoke-CapturedCommand -FilePath "go" -ArgumentList @("run", "./cmd/repomind", "trace", $fixtureRepo, "--analysis", $analysisPath, "--symbol", "update_order_status") -LogPath $traceLog -TimeoutSeconds $TimeoutSeconds
    $traceOutput = Get-Content -Path $traceLog -Raw
    if ($traceOutput -notmatch "update_order_status -> save") {
        throw "trace smoke did not find update_order_status -> save"
    }
    if ($traceOutput -notmatch "Mermaid") {
        throw "trace smoke did not print Mermaid diagram"
    }

    Invoke-CapturedCommand -FilePath "go" -ArgumentList @("run", "./cmd/repomind", "diagnose", $fixtureRepo, "--analysis", $analysisPath, "--issue", "order status error") -LogPath $diagnoseLog -TimeoutSeconds $TimeoutSeconds
    $diagnoseOutput = Get-Content -Path $diagnoseLog -Raw
    if ($diagnoseOutput -notmatch "\[state\]") {
        throw "diagnose smoke did not find state finding"
    }
    if ($diagnoseOutput -notmatch "\[database\]") {
        throw "diagnose smoke did not find database finding"
    }
}

if ($IncludeBenchmark) {
    $benchmarkArgs = @("-ExecutionPolicy", "Bypass", "-File", "scripts\benchmark-repos.ps1", "-OutputDir", (Join-Path $outputRoot "benchmark"), "-TimeoutSeconds", "$TimeoutSeconds", "-TargetSeconds", "$BenchmarkTargetSeconds", "-CloneRetries", "$CloneRetries", "-RepoCacheDir", $sharedRepoCache)
    if ($Proxy) {
        $benchmarkArgs += @("-Proxy", $Proxy)
    }
    $steps += Invoke-PreflightStep -Name "benchmark repositories" -Action {
        Invoke-CapturedCommand -FilePath "powershell" -ArgumentList $benchmarkArgs -LogPath (Join-Path $outputRoot "benchmark.log") -TimeoutSeconds ($TimeoutSeconds * 6)
    }
}

if ($IncludeEvaluation) {
    $evaluationArgs = @("-ExecutionPolicy", "Bypass", "-File", "scripts\evaluate-repos.ps1", "-OutputDir", (Join-Path $outputRoot "evaluation"), "-TimeoutSeconds", "$TimeoutSeconds", "-MinimumQualityScore", "$MinimumEvaluationQualityScore", "-CloneRetries", "$CloneRetries", "-RepoCacheDir", $sharedRepoCache)
    if ($Proxy) {
        $evaluationArgs += @("-Proxy", $Proxy)
    }
    $steps += Invoke-PreflightStep -Name "evaluate repositories" -Action {
        Invoke-CapturedCommand -FilePath "powershell" -ArgumentList $evaluationArgs -LogPath (Join-Path $outputRoot "evaluation.log") -TimeoutSeconds ($TimeoutSeconds * 6)
    }
}

if ($IncludeAskEvaluation) {
    $askEvaluationArgs = @("run", "./cmd/repomind", "eval", "ask", "--output", (Join-Path $outputRoot "ask-evaluation"), "--ai", $AskProvider, "--minimum-score", "$MinimumAskScore")
    if ($AskModel) {
        $askEvaluationArgs += @("--ai-model", $AskModel)
    }
    if ($AskCasesPath) {
        $askEvaluationArgs += @("--cases", $AskCasesPath)
    }
    if ($AskStrict) {
        $askEvaluationArgs += "--strict"
    }
    $askEvaluationEnv = @{}
    if ($Proxy) {
        $askEvaluationEnv["HTTPS_PROXY"] = $Proxy
        $askEvaluationEnv["HTTP_PROXY"] = $Proxy
        $askEvaluationEnv["ALL_PROXY"] = $Proxy
    }
    $steps += Invoke-PreflightStep -Name "evaluate ask (go cli)" -Action {
        Invoke-CapturedCommand -FilePath "go" -ArgumentList $askEvaluationArgs -LogPath (Join-Path $outputRoot "ask-evaluation.log") -TimeoutSeconds ($TimeoutSeconds * 3) -EnvironmentVariables $askEvaluationEnv
    }
}

if ($IncludeRemoteAnalyzeSmoke) {
    $remoteAnalyzeEnv = @{}
    if ($Proxy) {
        $remoteAnalyzeEnv["HTTPS_PROXY"] = $Proxy
        $remoteAnalyzeEnv["HTTP_PROXY"] = $Proxy
        $remoteAnalyzeEnv["ALL_PROXY"] = $Proxy
    }
    $remoteOutput = Join-Path $outputRoot "remote-analyze"
    $steps += Invoke-PreflightStep -Name "remote repository analyze smoke" -Action {
        Invoke-CapturedCommand -FilePath "go" -ArgumentList @("run", "./cmd/repomind", "analyze", "--output", $remoteOutput, "--repo-cache", $sharedRepoCache, $RemoteAnalyzeRepo) -LogPath (Join-Path $outputRoot "remote-analyze.log") -TimeoutSeconds ($TimeoutSeconds * 3) -EnvironmentVariables $remoteAnalyzeEnv
        $analysisPath = Join-Path $remoteOutput "analysis.json"
        if (-not (Test-Path $analysisPath)) {
            throw "remote analyze did not write analysis.json"
        }
        $analysis = Get-Content -Path $analysisPath -Raw | ConvertFrom-Json
        if (-not $analysis.repository.remote) {
            throw "remote analyze did not mark repository as remote"
        }
        if (@($analysis.routes).Count -lt 1) {
            throw "remote analyze did not extract expected API routes"
        }
    }
}

if ($IncludeAISmoke) {
    $aiArgs = @("-ExecutionPolicy", "Bypass", "-File", "scripts\smoke-ai-provider.ps1", "-Provider", $AIProvider, "-Model", $AIModel, "-OutputDir", (Join-Path $outputRoot "ai-smoke"), "-TimeoutSeconds", "$TimeoutSeconds")
    if ($Proxy) {
        $aiArgs += @("-Proxy", $Proxy)
    }
    $steps += Invoke-PreflightStep -Name "ai provider smoke" -Action {
        Invoke-CapturedCommand -FilePath "powershell" -ArgumentList $aiArgs -LogPath (Join-Path $outputRoot "ai-smoke.log") -TimeoutSeconds $TimeoutSeconds
    }
}

if ($IncludeReleaseSmoke) {
    $releaseSmokeArgs = @("-ExecutionPolicy", "Bypass", "-File", "scripts\smoke-release-artifact.ps1", "-OutputDir", (Join-Path $outputRoot "release-smoke"), "-TimeoutSeconds", "$TimeoutSeconds")
    $steps += Invoke-PreflightStep -Name "release artifact smoke" -Action {
        Invoke-CapturedCommand -FilePath "powershell" -ArgumentList $releaseSmokeArgs -LogPath (Join-Path $outputRoot "release-smoke.log") -TimeoutSeconds $TimeoutSeconds
    }
}

if ($IncludeManifestBuild) {
    $manifestDist = Join-Path $outputRoot "manifest-build"
    $steps += Invoke-PreflightStep -Name "release manifest build" -Action {
        Invoke-CapturedCommand -FilePath "powershell" -ArgumentList @("-ExecutionPolicy", "Bypass", "-File", "scripts\build-release.ps1", "-Version", $ManifestVersion, "-OutputDir", $manifestDist) -LogPath (Join-Path $outputRoot "manifest-build.log") -TimeoutSeconds $TimeoutSeconds
        Invoke-CapturedCommand -FilePath "powershell" -ArgumentList @("-ExecutionPolicy", "Bypass", "-File", "scripts\verify-release-manifest.ps1", "-DistDir", $manifestDist) -LogPath (Join-Path $outputRoot "manifest-verify.log") -TimeoutSeconds $TimeoutSeconds
    }
}

$failed = @($steps | Where-Object { -not $_.ok })
$summary = [ordered]@{
    ok = $failed.Count -eq 0
    generated_at = (Get-Date).ToUniversalTime().ToString("o")
    output_dir = $outputRoot
    include_benchmark = [bool]$IncludeBenchmark
    include_evaluation = [bool]$IncludeEvaluation
    include_ask_evaluation = [bool]$IncludeAskEvaluation
    ask_evaluation_runner = if ($IncludeAskEvaluation) { "go-cli" } else { "" }
    ask_cases_path = $AskCasesPath
    include_remote_analyze_smoke = [bool]$IncludeRemoteAnalyzeSmoke
    remote_analyze_repo = if ($IncludeRemoteAnalyzeSmoke) { $RemoteAnalyzeRepo } else { "" }
    include_ai_smoke = [bool]$IncludeAISmoke
    include_release_smoke = [bool]$IncludeReleaseSmoke
    include_manifest_build = [bool]$IncludeManifestBuild
    clone_retries = $CloneRetries
    repo_cache_dir = if ($IncludeBenchmark -or $IncludeEvaluation) { $sharedRepoCache } else { "" }
    steps = $steps
}

$summaryPath = Join-Path $outputRoot "summary.json"
[pscustomobject]$summary | ConvertTo-Json -Depth 8 | Set-Content -Path $summaryPath -Encoding UTF8

$markdownPath = Join-Path $outputRoot "summary.md"
$lines = @()
$lines += "# RepoMind Preflight Summary"
$lines += ""
$statusText = if ($failed.Count -eq 0) { "PASS" } else { "FAIL" }
$lines += "Status: $statusText"
$lines += ""
$lines += "| Step | OK | Seconds | Error |"
$lines += "|---|---:|---:|---|"
foreach ($step in $steps) {
    $errorText = ($step.error -replace "\r?\n", " ")
    $lines += "| $($step.name) | $($step.ok) | $($step.duration_seconds) | $errorText |"
}
$lines += ""
$lines += 'Raw JSON: `summary.json`'
$lines | Set-Content -Path $markdownPath -Encoding UTF8

Write-Host "Preflight summary written to $summaryPath"
Write-Host "Markdown summary written to $markdownPath"

if ($failed.Count -gt 0) {
    exit 1
}
