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
| `symfony/demo` | Symfony, PHP |
| `spring-projects/spring-petclinic` | Spring Boot, JPA |
| `spring-guides/gs-accessing-data-jpa` | Spring Boot, JPA |
| `labstack/echo` | Echo, Go |
| `gofiber/recipes` | Fiber, Go |
| `go-gorm/playground` | GORM models, Go |
| `django-oscar/django-oscar` | Django, Python |
| `nestjs/typescript-starter` | NestJS, TypeScript |
| `leerob/next-saas-starter` | Next.js, React, Postgres |
| `gothinkster/vue-realworld-example-app` | Vue |
| `gothinkster/react-redux-realworld-example-app` | React |
| `typeorm/typescript-express-example` | Express, TypeORM sample |
| `cookiecutter/cookiecutter-django` | Django template |

## 最新有效结果

2026-06-24 M106 release gate 中，20 个固定样本仍全部通过 `MinimumQualityScore 1.0`。同一轮 gate 还验证了核心报告内容、ask evaluation、trace/diagnose smoke、远程仓库 analyze smoke、安全边界、release artifact smoke 和 release manifest verification。

| Repo | Quality | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|
| Laravel | 1.00 | 1 | 0 | 0 |
| Spring REST service | 1.00 | 1 | 0 | 0 |
| Gin examples | 1.00 | 69 | 0 | 748 |
| Go chi | 1.00 | 229 | 0 | 1805 |
| FastAPI full-stack template | 1.00 | 23 | 2 | 851 |
| Node Express RealWorld | 1.00 | 20 | 4 | 99 |
| Prisma examples | 1.00 | 42 | 145 | 1764 |
| Symfony demo | 1.00 | 19 | 0 | 26 |
| Spring PetClinic | 1.00 | 18 | 6 | 0 |
| Spring Data JPA | 1.00 | 0 | 1 | 0 |
| Labstack Echo | 1.00 | 237 | 0 | 5000 |
| GoFiber Recipes | 1.00 | 278 | 49 | 5000 |
| Go GORM Playground | 1.00 | 0 | 6 | 24 |
| Django Oscar | 1.00 | 8 | 70 | 5000 |
| NestJS Starter | 1.00 | 1 | 0 | 4 |
| Next SaaS Starter | 1.00 | 4 | 0 | 284 |
| Vue RealWorld | 1.00 | 0 | 0 | 73 |
| React RealWorld | 1.00 | 0 | 0 | 176 |
| TypeORM Sample | 1.00 | 0 | 2 | 15 |
| Cookiecutter Django | 1.00 | 9 | 0 | 571 |

## 近期发现

- M95 后真实仓库 evaluation 固定样本从 7 个扩展到 20 个，覆盖 PHP、Java、Go、Python、JS/TS 和 frontend-only 仓库。
- M96 后 Express parser 需要文件中存在 Express app/router 信号，Vue/React frontend client 的 API wrapper 不再污染 API map。
- M97 后 FastAPI parser 需要文件中存在 FastAPI app/router 信号，Django 项目测试中的 `@patch(...)` 不再污染 API map。
- M98 后 Symfony demo routes 从 0 提升到 19，覆盖 PHP 8 `#[Route(...)]` controller attributes。
- M99 后 TypeORM `@Entity()` 无显式表名模型可识别，TypeORM sample models 从 0 提升到 2。
- M100 后 Next SaaS Starter routes 从 0 提升到 4，覆盖 Next.js `app/api/**/route.ts` route handlers。
- M101 后支持 SQLAlchemy 2.0 typed `Mapped[] = mapped_column(...)`，并移除 Django Oscar 中 9 个非 DB `class X(Base)` SQLAlchemy 误报。
- M106 后 release gate 同时覆盖核心报告内容、远程仓库 smoke、trace/diagnose smoke、安全边界和 release artifact smoke。
- FastAPI 多行 decorator 支持后，FastAPI full-stack template routes 从 18 提升到 23。
- Express 多行 route 支持后，node-express-realworld routes 从 8 提升到 20。
- Go middleware-wrapped handler 支持后，gin-examples routes 从 66 提升到 69。
- Go chi 样本保持 quality score 1.00，并检测到 229 routes。
- 前端 client 样本当前作为 stack 和 callgraph 覆盖；其中 HTTP client 调用已不再作为 Express API routes 输出。

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
