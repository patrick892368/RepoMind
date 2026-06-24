# 性能基准

**语言：** [English](PERFORMANCE_BENCHMARKS.md) | 简体中文

RepoMind 的核心产品目标是：`repomind analyze .` 在 30 秒内产出可用的仓库理解结果。

本文记录发布前性能基准流程。

## 运行 Benchmark

默认运行：

```powershell
.\scripts\benchmark-repos.ps1
```

如果 GitHub 访问需要本地代理：

```powershell
.\scripts\benchmark-repos.ps1 -Proxy http://127.0.0.1:10809
```

配置阈值和安全限制：

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

## Benchmark 仓库

当前基准仓库：

| Repo | Coverage |
|---|---|
| `laravel/laravel` | PHP, Laravel |
| `spring-guides/gs-rest-service` | Java, Spring Boot |
| `gin-gonic/examples` | Go, Gin |
| `fastapi/full-stack-fastapi-template` | FastAPI, React, Postgres, SQLModel |
| `prisma/prisma-examples` | Prisma, TypeScript monorepo |

## 最新有效结果

日期：2026-06-24。

来源：启用 repository cache 的组合 release gate。

目标：每个仓库低于 30 秒。

| 仓库 | 秒 | 低于目标 | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|---:|
| Laravel | 0.23 | true | 1 | 0 | 0 |
| Spring REST service | 0.16 | true | 1 | 0 | 0 |
| Gin examples | 0.20 | true | 69 | 0 | 748 |
| FastAPI full-stack template | 0.24 | true | 23 | 2 | 851 |
| Prisma examples | 0.58 | true | 42 | 145 | 1764 |

## 通过标准

每个仓库必须满足：

- clone 成功。
- analyze 成功。
- analyze 用时不超过 `TargetSeconds`，默认 30 秒。
- 即使 `truncated=true`，也必须生成 `analysis.json` 和 `report.html`。

脚本在任一仓库失败或超过阈值时返回非零退出码。

## 大仓库保护

默认限制：

| Limit | Default |
|---|---:|
| Max scanned files | 50000 |
| Max source file bytes parsed | 524288 |
| Max call graph edges kept | 5000 |

这些限制用于避免第一次运行被巨型仓库、生成代码或大文件卡住。

命中限制时：

- `analysis.json.scan.truncated` 为 `true`。
- CLI 显示 `Truncated: true` 或 `已截断: true`。
