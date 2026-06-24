param(
    [string]$Version = "dev",
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" },
    @{ GOOS = "windows"; GOARCH = "arm64"; Ext = ".exe" },
    @{ GOOS = "darwin"; GOARCH = "amd64"; Ext = "" },
    @{ GOOS = "darwin"; GOARCH = "arm64"; Ext = "" },
    @{ GOOS = "linux"; GOARCH = "amd64"; Ext = "" },
    @{ GOOS = "linux"; GOARCH = "arm64"; Ext = "" }
)

if (Test-Path $OutputDir) {
    Remove-Item -Recurse -Force $OutputDir
}
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

$manifest = @()

foreach ($target in $targets) {
    $name = "repomind-$Version-$($target.GOOS)-$($target.GOARCH)"
    $binary = "repomind$($target.Ext)"
    $targetDir = Join-Path $OutputDir $name
    New-Item -ItemType Directory -Force -Path $targetDir | Out-Null

    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH
    $env:CGO_ENABLED = "0"

    go build -trimpath -ldflags "-s -w -X main.version=$Version" -o (Join-Path $targetDir $binary) ./cmd/repomind

    Copy-Item LICENSE (Join-Path $targetDir "LICENSE")
    Copy-Item README.md (Join-Path $targetDir "README.md")
    Copy-Item README.zh-CN.md (Join-Path $targetDir "README.zh-CN.md")
    Copy-Item .env.example (Join-Path $targetDir ".env.example")

    $archiveBase = Join-Path $OutputDir $name
    $archivePath = ""
    if ($target.GOOS -eq "windows") {
        $archivePath = "$archiveBase.zip"
        Compress-Archive -Path (Join-Path $targetDir "*") -DestinationPath $archivePath -Force
    } else {
        $archivePath = "$archiveBase.tar.gz"
        tar -czf $archivePath -C $targetDir .
    }

    $archiveItem = Get-Item $archivePath
    $hash = Get-FileHash -Algorithm SHA256 -Path $archivePath
    $manifest += [pscustomobject]@{
        version = $Version
        goos = $target.GOOS
        goarch = $target.GOARCH
        archive = $archiveItem.Name
        size_bytes = $archiveItem.Length
        sha256 = $hash.Hash.ToLowerInvariant()
    }
}

Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

$manifestPath = Join-Path $OutputDir "manifest.json"
$manifest | ConvertTo-Json -Depth 4 | Set-Content -Path $manifestPath -Encoding UTF8

$manifestMarkdownPath = Join-Path $OutputDir "manifest.md"
$lines = @()
$lines += "# RepoMind Release Manifest"
$lines += ""
$lines += "| Version | GOOS | GOARCH | Archive | Size | SHA256 |"
$lines += "|---|---|---|---|---:|---|"
foreach ($item in $manifest) {
    $lines += "| $($item.version) | $($item.goos) | $($item.goarch) | $($item.archive) | $($item.size_bytes) | $($item.sha256) |"
}
$lines | Set-Content -Path $manifestMarkdownPath -Encoding UTF8

Write-Host "Release artifacts written to $OutputDir"
Write-Host "Release manifest written to $manifestPath"
