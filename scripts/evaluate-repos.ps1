param(
    [string]$OutputDir = "eval",
    [int]$TimeoutSeconds = 120,
    [string]$Proxy = "",
    [double]$MinimumQualityScore = 1.0,
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

function Invoke-CloneWithRetry {
    param(
        [string[]]$ArgumentList,
        [string]$LogPath,
        [string]$CleanupPath,
        [int]$TimeoutSeconds = 120,
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
    @{ Name = "laravel"; Url = "https://github.com/laravel/laravel.git"; Expected = "Laravel, PHP"; ExpectedStack = @("Laravel"); MinRoutes = 1; MinModels = 0; ExpectedRoutes = @("/"); ExpectedModels = @() },
    @{ Name = "spring-rest-service"; Url = "https://github.com/spring-guides/gs-rest-service.git"; Expected = "Spring Boot, Java"; ExpectedStack = @("Spring Boot"); MinRoutes = 1; MinModels = 0; ExpectedRoutes = @("/greeting"); ExpectedModels = @() },
    @{ Name = "gin-examples"; Url = "https://github.com/gin-gonic/examples.git"; Expected = "Gin, Go"; ExpectedStack = @("Gin"); MinRoutes = 1; MinModels = 0; ExpectedRoutes = @("/", "/book", "/bookable"); ExpectedModels = @() },
    @{ Name = "go-chi"; Url = "https://github.com/go-chi/chi.git"; Expected = "Go chi"; ExpectedStack = @("Chi"); MinRoutes = 1; MinModels = 0; ExpectedRoutes = @("/admin/accounts", "/admin/users/{userId}", "/articles/search", "/users/{id}", "/todos/{id}/sync"); ExpectedModels = @() },
    @{ Name = "fastapi-full-stack-template"; Url = "https://github.com/fastapi/full-stack-fastapi-template.git"; Expected = "FastAPI, React"; ExpectedStack = @("FastAPI", "React"); MinRoutes = 20; MinModels = 1; ExpectedRoutes = @("/api/v1/login/access-token", "/api/v1/users", "/api/v1/users/me", "/api/v1/users/{user_id}", "/api/v1/items", "/api/v1/users/signup", "/api/v1/utils/health-check"); ExpectedModels = @("User", "Item") },
    @{ Name = "node-express-realworld"; Url = "https://github.com/gothinkster/node-express-realworld-example-app.git"; Expected = "Express, Prisma"; ExpectedStack = @("Express"); MinRoutes = 10; MinModels = 1; ExpectedRoutes = @("/api/tags", "/api/articles", "/api/articles/feed", "/api/articles/:slug/comments", "/api/profiles/:username", "/api/users/login"); ExpectedModels = @("Article", "Comment", "Tag", "User") },
    @{ Name = "prisma-examples"; Url = "https://github.com/prisma/prisma-examples.git"; Expected = "Prisma, TypeScript"; ExpectedStack = @("NestJS", "Express", "Next.js", "Vue", "React"); MinRoutes = 1; MinModels = 1; ExpectedRoutes = @("/api/feed", "/api/filterPosts", "/api/users"); ExpectedModels = @("Post", "Account", "Comment", "Location") },
    @{ Name = "symfony-demo"; Url = "https://github.com/symfony/demo.git"; Expected = "Symfony, PHP"; ExpectedStack = @("Symfony"); MinRoutes = 0; MinModels = 0; ExpectedRoutes = @(); ExpectedModels = @() },
    @{ Name = "spring-petclinic"; Url = "https://github.com/spring-projects/spring-petclinic.git"; Expected = "Spring Boot, JPA"; ExpectedStack = @("Spring Boot"); MinRoutes = 10; MinModels = 6; ExpectedRoutes = @("/owners", "/owners/{ownerId}", "/owners/{ownerId}/pets/new"); ExpectedModels = @("Owner", "Pet", "Visit") },
    @{ Name = "spring-data-jpa"; Url = "https://github.com/spring-guides/gs-accessing-data-jpa.git"; Expected = "Spring Boot, JPA"; ExpectedStack = @("Spring Boot"); MinRoutes = 0; MinModels = 1; ExpectedRoutes = @(); ExpectedModels = @("Customer") },
    @{ Name = "labstack-echo"; Url = "https://github.com/labstack/echo.git"; Expected = "Echo, Go"; ExpectedStack = @("Echo"); MinRoutes = 1; MinModels = 0; ExpectedRoutes = @(); ExpectedModels = @() },
    @{ Name = "gofiber-recipes"; Url = "https://github.com/gofiber/recipes.git"; Expected = "Fiber, Go"; ExpectedStack = @("Fiber"); MinRoutes = 1; MinModels = 1; ExpectedRoutes = @(); ExpectedModels = @("Book") },
    @{ Name = "go-gorm-playground"; Url = "https://github.com/go-gorm/playground.git"; Expected = "GORM, Go"; ExpectedStack = @("Go", "Postgres", "MySQL"); MinRoutes = 0; MinModels = 5; ExpectedRoutes = @(); ExpectedModels = @("User", "Account") },
    @{ Name = "django-oscar"; Url = "https://github.com/django-oscar/django-oscar.git"; Expected = "Django, Python"; ExpectedStack = @("Django"); MinRoutes = 1; MinModels = 10; ExpectedRoutes = @("/admin/"); ExpectedModels = @("AbstractBasket") },
    @{ Name = "nestjs-starter"; Url = "https://github.com/nestjs/typescript-starter.git"; Expected = "NestJS, TypeScript"; ExpectedStack = @("NestJS"); MinRoutes = 1; MinModels = 0; ExpectedRoutes = @("/"); ExpectedModels = @() },
    @{ Name = "next-saas-starter"; Url = "https://github.com/leerob/next-saas-starter.git"; Expected = "Next.js, React"; ExpectedStack = @("Next.js", "React", "Postgres"); MinRoutes = 0; MinModels = 0; ExpectedRoutes = @(); ExpectedModels = @() },
    @{ Name = "vue-realworld"; Url = "https://github.com/gothinkster/vue-realworld-example-app.git"; Expected = "Vue"; ExpectedStack = @("Vue"); MinRoutes = 0; MinModels = 0; ExpectedRoutes = @(); ExpectedModels = @() },
    @{ Name = "react-realworld"; Url = "https://github.com/gothinkster/react-redux-realworld-example-app.git"; Expected = "React"; ExpectedStack = @("React"); MinRoutes = 0; MinModels = 0; ExpectedRoutes = @(); ExpectedModels = @() },
    @{ Name = "typeorm-sample"; Url = "https://github.com/typeorm/typescript-express-example.git"; Expected = "Express, TypeORM"; ExpectedStack = @("Express", "MySQL"); MinRoutes = 0; MinModels = 0; ExpectedRoutes = @(); ExpectedModels = @() },
    @{ Name = "cookiecutter-django"; Url = "https://github.com/cookiecutter/cookiecutter-django.git"; Expected = "Django template"; ExpectedStack = @("Django"); MinRoutes = 1; MinModels = 0; ExpectedRoutes = @("/api/"); ExpectedModels = @() }
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
        duration_seconds = $null
        files = $null
        directories = $null
        backend = ""
        frontend = ""
        database = ""
        cache = ""
        queue = ""
        models = 0
        routes = 0
        call_edges = 0
        quality_score = 0
        quality_checks = @()
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
        Invoke-CapturedCommand -FilePath "go" -ArgumentList @("run", "./cmd/repomind", "analyze", "--output", $analysisDir, $repoDir) -LogPath $analyzeLog -TimeoutSeconds $TimeoutSeconds
        $stopwatch.Stop()

        $item.duration_seconds = [math]::Round($stopwatch.Elapsed.TotalSeconds, 2)
        $item.analyze_ok = $true

        $analysisPath = Join-Path $analysisDir "analysis.json"
        $analysis = Get-Content $analysisPath -Raw | ConvertFrom-Json
        $item.files = $analysis.scan.total_files
        $item.directories = $analysis.scan.total_directories
        $item.backend = $analysis.stack.backend
        $item.frontend = $analysis.stack.frontend
        $item.database = $analysis.stack.database
        $item.cache = $analysis.stack.cache
        $item.queue = $analysis.stack.queue
        $item.models = if ($analysis.models) { $analysis.models.Count } else { 0 }
        $item.routes = if ($analysis.routes) { $analysis.routes.Count } else { 0 }
        $item.call_edges = if ($analysis.call_edges) { $analysis.call_edges.Count } else { 0 }

        $stackParts = @(
            $analysis.stack.backend,
            $analysis.stack.frontend,
            $analysis.stack.database,
            $analysis.stack.cache,
            $analysis.stack.queue,
            ($analysis.stack.languages -join ","),
            ($analysis.stack.package_manager -join ",")
        )
        $stackText = ($stackParts | Where-Object { $_ }) -join ", "
        $checks = @()
        foreach ($term in $repo.ExpectedStack) {
            $passed = $stackText -like "*$term*"
            $checks += [pscustomobject]@{
                name = "stack:$term"
                ok = [bool]$passed
                expected = $term
                actual = $stackText
            }
        }
        $checks += [pscustomobject]@{
            name = "routes:min"
            ok = [bool]($item.routes -ge $repo.MinRoutes)
            expected = $repo.MinRoutes
            actual = $item.routes
        }
        $checks += [pscustomobject]@{
            name = "models:min"
            ok = [bool]($item.models -ge $repo.MinModels)
            expected = $repo.MinModels
            actual = $item.models
        }
        $routePaths = @()
        if ($analysis.routes) {
            $routePaths = @($analysis.routes | ForEach-Object { $_.path })
        }
        foreach ($routePath in $repo.ExpectedRoutes) {
            $checks += [pscustomobject]@{
                name = "route:$routePath"
                ok = [bool]($routePaths -contains $routePath)
                expected = $routePath
                actual = ($routePaths | Select-Object -First 20) -join ", "
            }
        }
        $modelNames = @()
        if ($analysis.models) {
            $modelNames = @($analysis.models | ForEach-Object { $_.name })
        }
        foreach ($modelName in $repo.ExpectedModels) {
            $checks += [pscustomobject]@{
                name = "model:$modelName"
                ok = [bool]($modelNames -contains $modelName)
                expected = $modelName
                actual = ($modelNames | Select-Object -First 20) -join ", "
            }
        }
        $passedChecks = @($checks | Where-Object { $_.ok })
        $item.quality_checks = $checks
        $item.quality_score = if ($checks.Count -gt 0) { [math]::Round($passedChecks.Count / $checks.Count, 2) } else { 0 }
    } catch {
        $item.error = $_.Exception.Message
    }

    $summary += [pscustomobject]$item
}

$summaryPath = Join-Path $OutputDir "summary.json"
$summary | ConvertTo-Json -Depth 8 | Set-Content -Path $summaryPath -Encoding UTF8

$markdownPath = Join-Path $OutputDir "summary.md"
$lines = @()
$lines += "# RepoMind Real Repository Evaluation"
$lines += ""
$lines += "| Repo | Expected | Clone | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |"
$lines += "|---|---|---:|---:|---:|---:|---:|---|---|---|---:|---:|---:|"
foreach ($item in $summary) {
    $lines += "| $($item.name) | $($item.expected) | $($item.clone_ok) | $($item.analyze_ok) | $($item.quality_score) | $($item.duration_seconds) | $($item.files) | $($item.backend) | $($item.frontend) | $($item.database) | $($item.models) | $($item.routes) | $($item.call_edges) |"
}
$lines += ""
$failed = @($summary | Where-Object { -not $_.clone_ok -or -not $_.analyze_ok -or $_.quality_score -lt $MinimumQualityScore })
if ($failed.Count -gt 0) {
    $lines += "Status: FAIL"
    $lines += ""
    $lines += "Minimum quality score: $MinimumQualityScore"
    $lines += ""
    foreach ($item in $failed) {
        $lines += "- $($item.name): clone_ok=$($item.clone_ok), analyze_ok=$($item.analyze_ok), quality_score=$($item.quality_score), error=$($item.error)"
    }
} else {
    $lines += "Status: PASS"
}
$lines += ""
$lines += 'Raw JSON: `eval/summary.json`'
$lines | Set-Content -Path $markdownPath -Encoding UTF8

Write-Host "Evaluation summary written to $summaryPath"
Write-Host "Markdown summary written to $markdownPath"

if ($failed.Count -gt 0) {
    exit 1
}
