param(
    [string]$OutputDir = "benchmark",
    [int]$TimeoutSeconds = 300,
    [int]$TargetSeconds = 30,
    [string]$Proxy = "",
    [int]$MaxFiles = 50000,
    [int]$MaxFileBytes = 524288,
    [int]$MaxCallEdges = 5000,
    [int]$CloneRetries = 3,
    [string]$RepoCacheDir = ""
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
        [int]$TimeoutSeconds = 300
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
}

function Invoke-CloneWithRetry {
    param(
        [string[]]$ArgumentList,
        [string]$LogPath,
        [string]$CleanupPath,
        [int]$TimeoutSeconds = 300,
        [int]$Retries = 3
    )

    $lastError = $null
    for ($attempt = 1; $attempt -le $Retries; $attempt++) {
        try {
            Invoke-CapturedCommand -FilePath "git" -ArgumentList $ArgumentList -LogPath $LogPath -TimeoutSeconds $TimeoutSeconds
            return
        } catch {
            $lastError = $_.Exception.Message
            if ($attempt -ge $Retries) {
                throw $lastError
            }
            if (Test-Path $CleanupPath) {
                Remove-Item -LiteralPath $CleanupPath -Recurse -Force
            }
            Start-Sleep -Seconds ([Math]::Min(30, 5 * $attempt))
        }
    }
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

$repos = @(
    @{ Name = "laravel"; Url = "https://github.com/laravel/laravel.git"; Expected = "Laravel, PHP" },
    @{ Name = "spring-rest-service"; Url = "https://github.com/spring-guides/gs-rest-service.git"; Expected = "Spring Boot, Java" },
    @{ Name = "gin-examples"; Url = "https://github.com/gin-gonic/examples.git"; Expected = "Gin, Go" },
    @{ Name = "fastapi-full-stack-template"; Url = "https://github.com/fastapi/full-stack-fastapi-template.git"; Expected = "FastAPI, React" },
    @{ Name = "prisma-examples"; Url = "https://github.com/prisma/prisma-examples.git"; Expected = "Prisma, TypeScript monorepo" }
)

if (Test-Path $OutputDir) {
    Remove-Item -Recurse -Force $OutputDir
}

$reportsDir = Join-Path $OutputDir "reports"
$reposDir = if ($RepoCacheDir) { $RepoCacheDir } else { Join-Path $OutputDir "repos" }
New-Item -ItemType Directory -Force -Path $reposDir, $reportsDir | Out-Null
$reposDir = (Resolve-Path $reposDir).Path
$reportsDir = (Resolve-Path $reportsDir).Path

$summary = @()

foreach ($repo in $repos) {
    $repoDir = Join-Path $reposDir $repo.Name
    $analysisDir = Join-Path $reportsDir $repo.Name
    $cloneLog = Join-Path $reportsDir "$($repo.Name)-clone.log"
    $analyzeLog = Join-Path $reportsDir "$($repo.Name)-analyze.log"

    $item = [ordered]@{
        name = $repo.Name
        url = $repo.Url
        expected = $repo.Expected
        clone_ok = $false
        analyze_ok = $false
        under_target = $false
        target_seconds = $TargetSeconds
        duration_seconds = $null
        files = $null
        directories = $null
        total_bytes = $null
        truncated = $false
        packages = 0
        models = 0
        routes = 0
        call_edges = 0
        backend = ""
        frontend = ""
        database = ""
        error = ""
    }

    try {
        $gitArgs = @()
        if ($Proxy) {
            $gitArgs += @("-c", "http.proxy=$Proxy", "-c", "https.proxy=$Proxy")
        }
        if (Test-Path (Join-Path $repoDir ".git")) {
            "Reusing cached repository: $repoDir" | Set-Content -Path $cloneLog -Encoding UTF8
        } else {
            if (Test-Path $repoDir) {
                Remove-Item -LiteralPath $repoDir -Recurse -Force
            }
            $gitArgs += @("clone", "--depth", "1", $repo.Url, $repoDir)
            Invoke-CloneWithRetry -ArgumentList $gitArgs -LogPath $cloneLog -CleanupPath $repoDir -TimeoutSeconds $TimeoutSeconds -Retries $CloneRetries
        }
        $item.clone_ok = $true

        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        Invoke-CapturedCommand -FilePath "go" -ArgumentList @(
            "run", "./cmd/repomind",
            "analyze",
            "--output", $analysisDir,
            "--max-files", "$MaxFiles",
            "--max-file-bytes", "$MaxFileBytes",
            "--max-call-edges", "$MaxCallEdges",
            $repoDir
        ) -LogPath $analyzeLog -TimeoutSeconds $TimeoutSeconds
        $stopwatch.Stop()

        $item.duration_seconds = [math]::Round($stopwatch.Elapsed.TotalSeconds, 2)
        $item.analyze_ok = $true
        $item.under_target = $item.duration_seconds -le $TargetSeconds

        $analysisPath = Join-Path $analysisDir "analysis.json"
        $analysis = Get-Content $analysisPath -Raw | ConvertFrom-Json
        $item.files = $analysis.scan.total_files
        $item.directories = $analysis.scan.total_directories
        $item.total_bytes = $analysis.scan.total_bytes
        $item.truncated = [bool]$analysis.scan.truncated
        $item.packages = if ($analysis.packages) { $analysis.packages.Count } else { 0 }
        $item.models = if ($analysis.models) { $analysis.models.Count } else { 0 }
        $item.routes = if ($analysis.routes) { $analysis.routes.Count } else { 0 }
        $item.call_edges = if ($analysis.call_edges) { $analysis.call_edges.Count } else { 0 }
        $item.backend = $analysis.stack.backend
        $item.frontend = $analysis.stack.frontend
        $item.database = $analysis.stack.database
    } catch {
        $item.error = $_.Exception.Message
    }

    $summary += [pscustomobject]$item
}

$summaryPath = Join-Path $OutputDir "summary.json"
$summary | ConvertTo-Json -Depth 8 | Set-Content -Path $summaryPath -Encoding UTF8

$markdownPath = Join-Path $OutputDir "summary.md"
$lines = @()
$lines += "# RepoMind Repository Benchmark"
$lines += ""
$lines += "Target: $TargetSeconds seconds per repository."
$lines += ""
$lines += "| Repo | Analyze | Under Target | Seconds | Files | Bytes | Truncated | Packages | Models | Routes | Call Edges | Backend | Frontend | Database |"
$lines += "|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|---|---|"
foreach ($item in $summary) {
    $lines += "| $($item.name) | $($item.analyze_ok) | $($item.under_target) | $($item.duration_seconds) | $($item.files) | $($item.total_bytes) | $($item.truncated) | $($item.packages) | $($item.models) | $($item.routes) | $($item.call_edges) | $($item.backend) | $($item.frontend) | $($item.database) |"
}
$lines += ""
$failed = @($summary | Where-Object { -not $_.analyze_ok -or -not $_.under_target })
if ($failed.Count -gt 0) {
    $lines += "Status: FAIL"
    $lines += ""
    foreach ($item in $failed) {
        $lines += "- $($item.name): analyze_ok=$($item.analyze_ok), under_target=$($item.under_target), error=$($item.error)"
    }
} else {
    $lines += "Status: PASS"
}
$lines += ""
$lines += 'Raw JSON: `benchmark/summary.json`'
$lines | Set-Content -Path $markdownPath -Encoding UTF8

Write-Host "Benchmark summary written to $summaryPath"
Write-Host "Markdown summary written to $markdownPath"

if ($failed.Count -gt 0) {
    exit 1
}
