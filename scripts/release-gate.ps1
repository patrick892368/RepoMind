param(
    [string]$OutputDir = "eval/release-gate",
    [int]$TimeoutSeconds = 600,
    [string]$Proxy = "",
    [int]$BenchmarkTargetSeconds = 30,
    [double]$MinimumEvaluationQualityScore = 1.0,
    [int]$CloneRetries = 3,
    [string]$RepoCacheDir = "",
    [switch]$SkipManifestBuild,
    [switch]$IncludeAISmoke,
    [string]$AIProvider = "grok",
    [string]$AIModel = "grok-4.3"
)

$ErrorActionPreference = "Stop"

$args = @(
    "-ExecutionPolicy", "Bypass",
    "-File", "scripts\preflight.ps1",
    "-OutputDir", $OutputDir,
    "-TimeoutSeconds", "$TimeoutSeconds",
    "-IncludeReleaseSmoke",
    "-IncludeBenchmark",
    "-IncludeEvaluation",
    "-BenchmarkTargetSeconds", "$BenchmarkTargetSeconds",
    "-MinimumEvaluationQualityScore", "$MinimumEvaluationQualityScore",
    "-CloneRetries", "$CloneRetries"
)

if (-not $SkipManifestBuild) {
    $args += @("-IncludeManifestBuild", "-ManifestVersion", "v0.0.0-release-gate")
}

if ($RepoCacheDir) {
    $args += @("-RepoCacheDir", $RepoCacheDir)
}

if ($Proxy) {
    $args += @("-Proxy", $Proxy)
}

if ($IncludeAISmoke) {
    $args += @(
        "-IncludeAISmoke",
        "-AIProvider", $AIProvider,
        "-AIModel", $AIModel
    )
}

Write-Host "Running RepoMind release gate..."
powershell @args
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
