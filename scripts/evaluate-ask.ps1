param(
    [string]$OutputDir = "eval/ask",
    [string]$Provider = "offline",
    [string]$Model = "",
    [switch]$Strict,
    [string]$Proxy = "",
    [double]$MinimumScore = 1.0,
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

function Test-Contains {
    param(
        [object[]]$Values,
        [string]$Expected
    )

    foreach ($value in @($Values)) {
        if ([string]$value -eq $Expected) {
            return $true
        }
    }
    return $false
}

function Test-RouteContains {
    param(
        [object[]]$Routes,
        [string]$Expected
    )

    foreach ($route in @($Routes)) {
        $actual = ("{0} {1}" -f $route.method, $route.path).Trim()
        if ($actual -eq $Expected) {
            return $true
        }
    }
    return $false
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

$oldHTTPSProxy = $env:HTTPS_PROXY
$oldHTTPProxy = $env:HTTP_PROXY
$oldALLProxy = $env:ALL_PROXY

try {
    if ($Proxy) {
        $env:HTTPS_PROXY = $Proxy
        $env:HTTP_PROXY = $Proxy
        $env:ALL_PROXY = $Proxy
    }

    if (Test-Path $OutputDir) {
        Remove-Item -LiteralPath $OutputDir -Recurse -Force
    }
    New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
    $outputRoot = (Resolve-Path $OutputDir).Path

    $cases = @(
        [ordered]@{
            Name = "api-login"
            RepoPath = "testdata/fixtures/api-repo"
            Question = "where is login handled?"
            ExpectedFiles = @("fastapi_app/main.py", "django_project/urls.py")
            ExpectedHandlers = @("login", "views.login_view")
            ExpectedRoutes = @("POST /login", "ANY /login/")
            ExpectedModels = @()
            MinimumEvidence = 2
        },
        [ordered]@{
            Name = "api-wallet"
            RepoPath = "testdata/fixtures/api-repo"
            Question = "where is wallet info exposed?"
            ExpectedFiles = @("fastapi_app/main.py")
            ExpectedHandlers = @("wallet_info")
            ExpectedRoutes = @("GET /wallet/info")
            ExpectedModels = @()
            MinimumEvidence = 1
        },
        [ordered]@{
            Name = "self-cli-ask"
            RepoPath = "."
            Question = "where is ask handled in the CLI?"
            ExpectedFiles = @("cmd/repomind/main.go", "internal/query/query.go")
            ExpectedHandlers = @("runAsk")
            ExpectedRoutes = @()
            ExpectedModels = @()
            MinimumEvidence = 2
        }
    )

    $caseResults = @()

    foreach ($case in $cases) {
        $caseDir = Join-Path $outputRoot $case.Name
        $analysisDir = Join-Path $caseDir "analysis"
        $askDir = Join-Path $caseDir "ask"
        New-Item -ItemType Directory -Force -Path $caseDir | Out-Null

        $item = [ordered]@{
            name = $case.Name
            repo_path = $case.RepoPath
            question = $case.Question
            provider = $Provider
            strict = [bool]$Strict
            analyze_ok = $false
            ask_ok = $false
            score = 0
            checks = @()
            answer_summary = ""
            error = ""
        }

        try {
            Invoke-CapturedCommand -FilePath "go" -ArgumentList @("run", "./cmd/repomind", "analyze", "--output", $analysisDir, $case.RepoPath) -LogPath (Join-Path $caseDir "analyze.log") -TimeoutSeconds $TimeoutSeconds
            $item.analyze_ok = $true

            $askArgs = @("run", "./cmd/repomind", "ask", $case.RepoPath, "--analysis", (Join-Path $analysisDir "analysis.json"), "--question", $case.Question, "--ai", $Provider, "--output", $askDir)
            if ($Model) {
                $askArgs += @("--ai-model", $Model)
            }
            if ($Strict) {
                $askArgs += "--strict"
            }
            Invoke-CapturedCommand -FilePath "go" -ArgumentList $askArgs -LogPath (Join-Path $caseDir "ask.log") -TimeoutSeconds $TimeoutSeconds
            $item.ask_ok = $true

            $answerPath = Join-Path $askDir "last-answer.json"
            $answer = Get-Content $answerPath -Raw | ConvertFrom-Json
            $item.answer_summary = $answer.summary

            $checks = @()
            foreach ($expected in $case.ExpectedFiles) {
                $checks += [pscustomobject]@{
                    name = "file:$expected"
                    ok = [bool](Test-Contains -Values @($answer.files) -Expected $expected)
                    expected = $expected
                    actual = (@($answer.files) -join ", ")
                }
            }
            foreach ($expected in $case.ExpectedHandlers) {
                $checks += [pscustomobject]@{
                    name = "handler:$expected"
                    ok = [bool](Test-Contains -Values @($answer.handlers) -Expected $expected)
                    expected = $expected
                    actual = (@($answer.handlers) -join ", ")
                }
            }
            foreach ($expected in $case.ExpectedRoutes) {
                $checks += [pscustomobject]@{
                    name = "route:$expected"
                    ok = [bool](Test-RouteContains -Routes @($answer.routes) -Expected $expected)
                    expected = $expected
                    actual = (@($answer.routes | ForEach-Object { ("{0} {1}" -f $_.method, $_.path).Trim() }) -join ", ")
                }
            }
            foreach ($expected in $case.ExpectedModels) {
                $checks += [pscustomobject]@{
                    name = "model:$expected"
                    ok = [bool](Test-Contains -Values @($answer.models) -Expected $expected)
                    expected = $expected
                    actual = (@($answer.models) -join ", ")
                }
            }
            $evidenceCount = @($answer.evidence).Count
            $checks += [pscustomobject]@{
                name = "evidence:min"
                ok = [bool]($evidenceCount -ge $case.MinimumEvidence)
                expected = $case.MinimumEvidence
                actual = $evidenceCount
            }
            if ($Strict) {
                $checks += [pscustomobject]@{
                    name = "strict:evidence"
                    ok = [bool]($evidenceCount -gt 0 -and $answer.confidence -ne "insufficient_evidence")
                    expected = "evidence-backed answer"
                    actual = "evidence=$evidenceCount confidence=$($answer.confidence)"
                }
            }

            $passed = @($checks | Where-Object { $_.ok }).Count
            $total = @($checks).Count
            $item.checks = $checks
            $item.score = if ($total -gt 0) { [math]::Round($passed / $total, 4) } else { 0 }
        } catch {
            $item.error = $_.Exception.Message
        }

        $caseResults += [pscustomobject]$item
    }

    $allChecks = @($caseResults | ForEach-Object { $_.checks } | Where-Object { $_ })
    $passedChecks = @($allChecks | Where-Object { $_.ok }).Count
    $totalChecks = @($allChecks).Count
    $overallScore = if ($totalChecks -gt 0) { [math]::Round($passedChecks / $totalChecks, 4) } else { 0 }
    $failedCases = @($caseResults | Where-Object { -not $_.analyze_ok -or -not $_.ask_ok -or $_.score -lt $MinimumScore })
    $ok = ($failedCases.Count -eq 0 -and $overallScore -ge $MinimumScore)

    $summary = [ordered]@{
        ok = $ok
        generated_at = (Get-Date).ToUniversalTime().ToString("o")
        output_dir = $outputRoot
        provider = $Provider
        model = $Model
        strict = [bool]$Strict
        minimum_score = $MinimumScore
        overall_score = $overallScore
        passed_checks = $passedChecks
        total_checks = $totalChecks
        cases = $caseResults
    }

    $summaryPath = Join-Path $outputRoot "summary.json"
    [pscustomobject]$summary | ConvertTo-Json -Depth 10 | Set-Content -Path $summaryPath -Encoding UTF8

    $markdownPath = Join-Path $outputRoot "summary.md"
    $lines = @()
    $lines += "# RepoMind Ask Evaluation Summary"
    $lines += ""
    $lines += "Status: $(if ($ok) { 'PASS' } else { 'FAIL' })"
    $lines += ""
    $lines += "Provider: $Provider"
    $lines += "Strict: $([bool]$Strict)"
    $lines += "Minimum score: $MinimumScore"
    $lines += "Overall score: $overallScore"
    $lines += ""
    $lines += "| Case | Analyze | Ask | Score | Error |"
    $lines += "|---|---:|---:|---:|---|"
    foreach ($caseResult in $caseResults) {
        $errorText = ($caseResult.error -replace "\r?\n", " ")
        $lines += "| $($caseResult.name) | $($caseResult.analyze_ok) | $($caseResult.ask_ok) | $($caseResult.score) | $errorText |"
    }
    $lines += ""
    $lines += "## Checks"
    $lines += ""
    $lines += "| Case | Check | OK | Expected | Actual |"
    $lines += "|---|---|---:|---|---|"
    foreach ($caseResult in $caseResults) {
        foreach ($check in @($caseResult.checks)) {
            $actual = ([string]$check.actual) -replace "\r?\n", " "
            $lines += "| $($caseResult.name) | $($check.name) | $($check.ok) | $($check.expected) | $actual |"
        }
    }
    $lines += ""
    $lines += 'Raw JSON: `summary.json`'
    $lines | Set-Content -Path $markdownPath -Encoding UTF8

    Write-Host "Ask evaluation summary written to $summaryPath"
    Write-Host "Markdown summary written to $markdownPath"

    if (-not $ok) {
        exit 1
    }
} finally {
    $env:HTTPS_PROXY = $oldHTTPSProxy
    $env:HTTP_PROXY = $oldHTTPProxy
    $env:ALL_PROXY = $oldALLProxy
}
