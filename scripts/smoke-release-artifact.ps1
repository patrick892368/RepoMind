param(
    [string]$OutputDir = "dist/release-smoke",
    [string]$RepoPath = "testdata/fixtures/api-repo",
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
}

function Invoke-SmokeStep {
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

if (Test-Path $OutputDir) {
    Remove-Item -Recurse -Force $OutputDir
}
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
$outputRoot = (Resolve-Path $OutputDir).Path

$binaryName = "repomind"
if ($IsWindows -or $env:OS -eq "Windows_NT") {
    $binaryName = "repomind.exe"
}
$binaryPath = Join-Path $outputRoot $binaryName
$repoCopy = Join-Path $outputRoot "repo"
Copy-Item -Recurse -Force -Path (Resolve-Path $RepoPath).Path -Destination $repoCopy

$steps = @()
$steps += Invoke-SmokeStep -Name "build binary" -Action {
    Invoke-CapturedCommand -FilePath "go" -ArgumentList @("build", "-trimpath", "-o", $binaryPath, "./cmd/repomind") -LogPath (Join-Path $outputRoot "build.log") -TimeoutSeconds $TimeoutSeconds
}
$steps += Invoke-SmokeStep -Name "version" -Action {
    Invoke-CapturedCommand -FilePath $binaryPath -ArgumentList @("version") -LogPath (Join-Path $outputRoot "version.log") -TimeoutSeconds $TimeoutSeconds
}
$steps += Invoke-SmokeStep -Name "analyze en" -Action {
    Invoke-CapturedCommand -FilePath $binaryPath -ArgumentList @("analyze", "--output", (Join-Path $repoCopy ".repomind"), "--lang", "en", $repoCopy) -LogPath (Join-Path $outputRoot "analyze-en.log") -TimeoutSeconds $TimeoutSeconds
    $analysisPath = Join-Path $repoCopy ".repomind/analysis.json"
    $reportPath = Join-Path $repoCopy ".repomind/report.html"
    if (-not (Test-Path $analysisPath)) {
        throw "analysis.json was not generated"
    }
    if (-not (Test-Path $reportPath)) {
        throw "report.html was not generated"
    }
}
$steps += Invoke-SmokeStep -Name "export codex" -Action {
    Invoke-CapturedCommand -FilePath $binaryPath -ArgumentList @("export", "codex", $repoCopy) -LogPath (Join-Path $outputRoot "export-codex.log") -TimeoutSeconds $TimeoutSeconds
    if (-not (Test-Path (Join-Path $repoCopy "AGENTS.md"))) {
        throw "AGENTS.md was not generated"
    }
}
$steps += Invoke-SmokeStep -Name "analyze zh" -Action {
    $zhOutput = Join-Path $repoCopy ".repomind-zh"
    Invoke-CapturedCommand -FilePath $binaryPath -ArgumentList @("analyze", "--output", $zhOutput, "--lang", "zh", $repoCopy) -LogPath (Join-Path $outputRoot "analyze-zh.log") -TimeoutSeconds $TimeoutSeconds
    $analysis = Get-Content (Join-Path $zhOutput "analysis.json") -Raw | ConvertFrom-Json
    if ($analysis.language -ne "zh") {
        throw "Chinese analyze smoke wrote language=$($analysis.language)"
    }
}

$failed = @($steps | Where-Object { -not $_.ok })
$summary = [ordered]@{
    ok = $failed.Count -eq 0
    generated_at = (Get-Date).ToUniversalTime().ToString("o")
    output_dir = $outputRoot
    binary = $binaryPath
    repo_copy = $repoCopy
    steps = $steps
}

$summaryPath = Join-Path $outputRoot "summary.json"
[pscustomobject]$summary | ConvertTo-Json -Depth 8 | Set-Content -Path $summaryPath -Encoding UTF8

$markdownPath = Join-Path $outputRoot "summary.md"
$statusText = if ($failed.Count -eq 0) { "PASS" } else { "FAIL" }
$lines = @()
$lines += "# RepoMind Release Artifact Smoke"
$lines += ""
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

Write-Host "Release artifact smoke summary written to $summaryPath"
Write-Host "Markdown summary written to $markdownPath"

if ($failed.Count -gt 0) {
    exit 1
}
