# 真实仓库评估

**语言：** [English](REAL_REPO_EVALUATION.md) | 简体中文

本文记录 RepoMind 对真实开源仓库的评估结果和 parser 质量记录。

目标：

- 验证 `repomind analyze` 是否能在 30 秒产品目标内产出有用上下文。
- 发现 parser 误报、漏报和缺失的框架模式。
- 用真实仓库驱动 fixture 和 parser 增强。

## 评估脚本

运行：

```powershell
.\scripts\evaluate-repos.ps1
```

如果 GitHub 访问需要代理：

```powershell
.\scripts\evaluate-repos.ps1 -Proxy http://127.0.0.1:10809
```

使用共享 repo cache：

```powershell
.\scripts\evaluate-repos.ps1 `
  -Proxy http://127.0.0.1:10809 `
  -RepoCacheDir eval\release-gate\repo-cache `
  -MinimumQualityScore 1.0
```

输出：

```txt
eval/summary.json
eval/summary.md
eval/reports/
eval/repos/
```

`eval/` 被 Git 忽略。

## 当前样本

| Repo | Coverage |
|---|---|
| `laravel/laravel` | Laravel, PHP |
| `spring-guides/gs-rest-service` | Spring Boot, Java |
| `gin-gonic/examples` | Gin, Go |
| `go-chi/chi` | Chi, Go |
| `fastapi/full-stack-fastapi-template` | FastAPI, React, Postgres |
| `gothinkster/node-express-realworld-example-app` | Express, Prisma |
| `prisma/prisma-examples` | Prisma, TypeScript monorepo |

## 最新有效结果

M72 release gate 和 M71 evaluation 均通过 `MinimumQualityScore 1.0`。

| Repo | Quality | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|
| Laravel | 1.00 | 1 | 0 | 0 |
| Spring REST service | 1.00 | 1 | 0 | 0 |
| Gin examples | 1.00 | 68 | 0 | 748 |
| Go chi | 1.00 | 210 | 0 | 1805 |
| FastAPI full-stack template | 1.00 | 23 | 2 | 851 |
| Node Express RealWorld | 1.00 | 20 | 4 | 99 |
| Prisma examples | 1.00 | 55 | 143 | 1764 |

## 近期发现

- FastAPI 多行 decorator 支持后，FastAPI full-stack template routes 从 18 提升到 23。
- Express 多行 route 支持后，node-express-realworld routes 从 8 提升到 20。
- Go middleware-wrapped handler 支持后，gin-examples routes 从 66 提升到 68。
- Go chi 样本保持 quality score 1.00，并检测到 210 routes。

## 记录格式

发现 parser 问题时记录：

```txt
Repository:
Language/framework:
Issue type: false positive | false negative | low confidence
Expected:
Actual:
Candidate fixture:
Priority:
```
