# Performance Benchmarks

**Language:** English | [简体中文](PERFORMANCE_BENCHMARKS.zh-CN.md)

RepoMind's core product target is: `repomind analyze .` should produce useful repository understanding within 30 seconds.

This document records the performance benchmark process used before release.

## Run Benchmark

Default run:

```powershell
.\scripts\benchmark-repos.ps1
```

When GitHub access requires a local proxy:

```powershell
.\scripts\benchmark-repos.ps1 -Proxy http://127.0.0.1:10809
```

The target and safety limits are configurable:

```powershell
.\scripts\benchmark-repos.ps1 `
  -TargetSeconds 30 `
  -MaxFiles 50000 `
  -MaxFileBytes 524288 `
  -MaxCallEdges 5000
```

Output directories:

```txt
benchmark/summary.json
benchmark/summary.md
benchmark/reports/
benchmark/repos/
```

`benchmark/` is ignored by Git.

## Benchmark Repositories

Current benchmark repositories:

| Repo | Coverage |
|---|---|
| `laravel/laravel` | PHP, Laravel |
| `spring-guides/gs-rest-service` | Java, Spring Boot |
| `gin-gonic/examples` | Go, Gin |
| `fastapi/full-stack-fastapi-template` | FastAPI, React, Postgres, SQLModel |
| `prisma/prisma-examples` | Prisma, TypeScript monorepo |

## Latest Validated Result

Date: 2026-06-24.

Source: combined release gate run with repository cache enabled.

Target: 30 seconds per repository.

| Repository | Seconds | Under Target | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|---:|
| Laravel | 0.22 | true | 1 | 0 | 0 |
| Spring REST service | 0.17 | true | 1 | 0 | 0 |
| Gin examples | 0.19 | true | 69 | 0 | 748 |
| FastAPI full-stack template | 0.25 | true | 23 | 2 | 851 |
| Prisma examples | 0.55 | true | 42 | 145 | 1764 |

## Pass Criteria

Each repository must satisfy:

- Clone succeeds.
- Analyze succeeds.
- Analyze time is no greater than `TargetSeconds`, default 30 seconds.
- `analysis.json` and `report.html` are generated even when `truncated=true`.

`scripts/benchmark-repos.ps1` exits nonzero when any repository fails or exceeds the target.

## Large Repository Guards

RepoMind enables the following safety limits by default:

| Limit | Default |
|---|---:|
| Max scanned files | 50000 |
| Max source file bytes parsed | 524288 |
| Max call graph edges kept | 5000 |

These limits keep first-run analysis from getting stuck on very large repositories, generated code, or large files.

When a limit is reached:

- `analysis.json.scan.truncated` is `true`.
- The CLI shows `Truncated: true` or `已截断: true`.
- The HTML report shows the truncation notice.
- Some large files or call edges may be excluded from model, route, and call graph analysis.
