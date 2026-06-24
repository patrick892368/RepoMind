param(
    [string]$OutputDir = "eval/ask",
    [string]$CasesPath = "",
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

function Resolve-ProxyValue {
    param([string]$ExplicitProxy)

    if ($ExplicitProxy) {
        return $ExplicitProxy
    }
    if ($env:HTTPS_PROXY) {
        return $env:HTTPS_PROXY
    }
    if ($env:HTTP_PROXY) {
        return $env:HTTP_PROXY
    }
    if ($env:ALL_PROXY) {
        return $env:ALL_PROXY
    }
    return ""
}

function Write-WrapperLog {
    param(
        [string]$OutputDir,
        [string[]]$Lines
    )

    if (-not (Test-Path -LiteralPath $OutputDir)) {
        return
    }
    $resolvedOutput = (Resolve-Path -LiteralPath $OutputDir).Path
    $logPath = Join-Path $resolvedOutput "wrapper.log"
    $Lines | Set-Content -Path $logPath -Encoding UTF8
}

$arguments = @(
    "run",
    "./cmd/repomind",
    "eval",
    "ask",
    "--output",
    $OutputDir,
    "--ai",
    $Provider,
    "--minimum-score",
    "$MinimumScore"
)

if ($Model) {
    $arguments += @("--ai-model", $Model)
}
if ($CasesPath) {
    $arguments += @("--cases", $CasesPath)
}
if ($Strict) {
    $arguments += "--strict"
}

$proxyValue = Resolve-ProxyValue -ExplicitProxy $Proxy
$psi = [System.Diagnostics.ProcessStartInfo]::new()
$psi.FileName = "go"
$psi.Arguments = ($arguments | ForEach-Object { ConvertTo-ProcessArgument $_ }) -join " "
$psi.WorkingDirectory = (Resolve-Path (Get-Location).Path).Path
$psi.RedirectStandardOutput = $true
$psi.RedirectStandardError = $true
$psi.UseShellExecute = $false
$psi.CreateNoWindow = $true
if ($proxyValue) {
    $psi.Environment["HTTPS_PROXY"] = $proxyValue
    $psi.Environment["HTTP_PROXY"] = $proxyValue
    $psi.Environment["ALL_PROXY"] = $proxyValue
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
    throw "go timed out after $TimeoutSeconds seconds"
}

$stdout = $stdoutTask.GetAwaiter().GetResult()
$stderr = $stderrTask.GetAwaiter().GetResult()

if ($stdout) {
    Write-Host $stdout.TrimEnd()
}
if ($stderr) {
    [Console]::Error.WriteLine($stderr.TrimEnd())
}

$log = @()
$log += "COMMAND: go $($psi.Arguments)"
$log += "PROXY: $(if ($proxyValue) { "configured" } else { "not configured" })"
$log += "EXIT_CODE: $($process.ExitCode)"
if ($stdout) {
    $log += "STDOUT:"
    $log += $stdout.TrimEnd()
}
if ($stderr) {
    $log += "STDERR:"
    $log += $stderr.TrimEnd()
}
Write-WrapperLog -OutputDir $OutputDir -Lines $log

if ($process.ExitCode -ne 0) {
    exit $process.ExitCode
}
