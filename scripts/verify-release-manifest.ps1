param(
    [string]$DistDir = "dist"
)

$ErrorActionPreference = "Stop"

$manifestPath = Join-Path $DistDir "manifest.json"
if (-not (Test-Path $manifestPath)) {
    throw "manifest not found: $manifestPath"
}

$manifest = Get-Content $manifestPath -Raw | ConvertFrom-Json
$results = @()

foreach ($item in $manifest) {
    $archivePath = Join-Path $DistDir $item.archive
    $result = [ordered]@{
        archive = $item.archive
        exists = $false
        size_ok = $false
        sha256_ok = $false
        expected_size = [int64]$item.size_bytes
        actual_size = $null
        expected_sha256 = $item.sha256
        actual_sha256 = ""
    }

    if (Test-Path $archivePath) {
        $archive = Get-Item $archivePath
        $hash = Get-FileHash -Algorithm SHA256 -Path $archivePath
        $result.exists = $true
        $result.actual_size = $archive.Length
        $result.actual_sha256 = $hash.Hash.ToLowerInvariant()
        $result.size_ok = $archive.Length -eq [int64]$item.size_bytes
        $result.sha256_ok = $result.actual_sha256 -eq "$($item.sha256)".ToLowerInvariant()
    }

    $results += [pscustomobject]$result
}

$failed = @($results | Where-Object { -not $_.exists -or -not $_.size_ok -or -not $_.sha256_ok })
$summary = [ordered]@{
    ok = $failed.Count -eq 0
    manifest = (Resolve-Path $manifestPath).Path
    artifacts = $results.Count
    results = $results
}

$summaryPath = Join-Path $DistDir "manifest-verify.json"
[pscustomobject]$summary | ConvertTo-Json -Depth 8 | Set-Content -Path $summaryPath -Encoding UTF8

$markdownPath = Join-Path $DistDir "manifest-verify.md"
$statusText = if ($failed.Count -eq 0) { "PASS" } else { "FAIL" }
$lines = @()
$lines += "# RepoMind Release Manifest Verification"
$lines += ""
$lines += "Status: $statusText"
$lines += ""
$lines += "| Archive | Exists | Size OK | SHA256 OK |"
$lines += "|---|---:|---:|---:|"
foreach ($result in $results) {
    $lines += "| $($result.archive) | $($result.exists) | $($result.size_ok) | $($result.sha256_ok) |"
}
$lines += ""
$lines += 'Raw JSON: `manifest-verify.json`'
$lines | Set-Content -Path $markdownPath -Encoding UTF8

Write-Host "Manifest verification written to $summaryPath"
Write-Host "Markdown summary written to $markdownPath"

if ($failed.Count -gt 0) {
    exit 1
}
