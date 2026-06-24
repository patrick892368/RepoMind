# Performance Benchmarks

RepoMind 的核心产品目标是：`repomind analyze .` 在 30 秒内产出可用的仓库理解结果。

本文件记录发布前性能基准流程。

## Run Benchmark

默认运行：

```powershell
.\scripts\benchmark-repos.ps1
```

如果 GitHub 访问需要本地代理：

```powershell
.\scripts\benchmark-repos.ps1 -Proxy http://127.0.0.1:10809
```

可配置阈值和安全限制：

```powershell
.\scripts\benchmark-repos.ps1 `
  -TargetSeconds 30 `
  -MaxFiles 50000 `
  -MaxFileBytes 524288 `
  -MaxCallEdges 5000
```

输出目录：

```txt
benchmark/summary.json
benchmark/summary.md
benchmark/reports/
benchmark/repos/
```

`benchmark/` 被 Git 忽略。

## Benchmark Repositories

当前基准仓库：

| Repo | Coverage |
|---|---|
| `laravel/laravel` | PHP, Laravel |
| `spring-guides/gs-rest-service` | Java, Spring Boot |
| `gin-gonic/examples` | Go, Gin |
| `fastapi/full-stack-fastapi-template` | FastAPI, React, Postgres, SQLModel |
| `prisma/prisma-examples` | Prisma, TypeScript monorepo |

## Pass Criteria

每个仓库必须满足：

- clone 成功。
- analyze 成功。
- analyze 用时不超过 `TargetSeconds`，默认 30 秒。
- 即使 `truncated=true`，也必须生成 `analysis.json` 和 `report.html`。

`scripts/benchmark-repos.ps1` 在任一仓库失败或超过阈值时返回非零退出码。

## Large Repository Guards

RepoMind 默认启用以下安全限制：

| Limit | Default |
|---|---:|
| Max scanned files | 50000 |
| Max source file bytes parsed | 524288 |
| Max call graph edges kept | 5000 |

这些限制是为了保证第一次运行不会因为巨型仓库、生成代码或大文件卡住。

如果命中限制：

- `analysis.json.scan.truncated` 会为 `true`。
- CLI 会显示 `Truncated: true` 或 `已截断: true`。
- HTML report 会显示截断提示。
- 部分大文件或调用边可能不会进入模型、路由、调用图分析。
