# Parser 改进 Backlog

**语言：** [English](PARSER_BACKLOG.md) | 简体中文

RepoMind 必须优先保持确定性扫描。AI 可以做总结和推断，但 parser 质量必须通过明确 fixture、真实仓库 evaluation、误报/漏报记录来提升。

## 优先级规则

选择 parser 工作时按以下顺序：

1. 修复会污染首次报告的高置信误报。
2. 修复主流框架中的 route/model 漏报。
3. 改进业务流程需要的 callgraph edges。
4. 新增 parser 输出必须包含 evidence 和 confidence。
5. 扩展真实仓库 evaluation 前先补 fixture。

## Tree-sitter 引入边界

tree-sitter 对 JS/TS、Python、PHP、Java 可能有价值，但不能盲目引入。

只有满足以下条件时才考虑采用：

- 跨平台构建仍然简单。
- Windows、macOS、Linux CI 稳定。
- parser 代码明显比现有实现更简单或更准确。
- benchmark 仍低于 30 秒目标。
- 新依赖不影响单 binary 发布。

Go 优先使用标准库 AST，除非 tree-sitter 有明确优势。

跨文件 route prefix 传播见 `ROUTE_PREFIX_STRATEGY.md`。

## JavaScript / TypeScript

当前覆盖：

- Express routes。
- Express 多行 route calls。
- 同文件 Express `app.use("/prefix", router)` prefix 传播。
- Cross-file Express relative router imports。
- Express composed router prefix。
- NestJS controllers。
- TypeORM entities。
- 轻量 callgraph。

Backlog：

- 多层 Express nested routers。
- alias chains 或 dynamic exports。
- path 动态拼接或超过轻量窗口的 route calls。
- NestJS prefix 和 method decorator 边界。
- Next.js App Router route handlers。
- Next.js Pages Router API routes。
- Hono routes。
- tRPC routers。
- TypeORM relation decorators。
- JS/TS class methods 和 imported service callgraph。

## Python

当前覆盖：

- Django URL patterns。
- Django same-file `include()` prefix。
- Django `include("module.urls")` module prefix。
- Django REST Framework same-file router registrations。
- Django REST Framework statically registered ViewSet custom actions，覆盖 `detail`、`methods`、`url_path` 和一跳 `include("module.urls")` 父级 prefix。
- FastAPI decorators。
- FastAPI 多行 decorators。
- FastAPI `APIRouter(prefix=...)` 和 same-file `include_router`。
- FastAPI direct imported router prefix。
- FastAPI composed router prefix 和唯一静态 prefix 常量。
- Django models。
- SQLAlchemy models。
- SQLModel table models。
- 轻量 callgraph。

Backlog：

- Django namespace、app name 或动态 include。
- 超出一跳静态 module include 和静态 ViewSet 注册的 DRF cross-file routers。
- 更复杂的 FastAPI module imports。
- 动态 path 或超出轻量窗口的 FastAPI decorators。
- SQLAlchemy 2.0 `Mapped[]` 和 `mapped_column`。
- Alembic model hints。
- Celery task discovery。
- Python class methods 和 imported service callgraph。

## PHP

当前覆盖：

- Laravel routes。
- Laravel/Symfony/ThinkPHP stack detection。

Backlog：

- Laravel route groups 和 prefixes。
- Laravel controller array syntax。
- Laravel resource routes。
- Laravel Eloquent models。
- Symfony controller attributes。
- ThinkPHP route definitions。
- PHP service callgraph。

## Java

当前覆盖：

- Spring controller routes。
- JPA entities。
- Spring Boot stack detection。

Backlog：

- Controller-level `@RequestMapping` arrays。
- `@RequestParam`、`@PathVariable` 和 method metadata。
- Spring WebFlux routes。
- MyBatis mapper interfaces 和 XML。
- JPA relation annotations。
- Java service method callgraph。

## Go

当前覆盖：

- Gin/Echo/Fiber-style route calls through Go AST。
- Chi-style `Route("/prefix", func(...))` prefix。
- Same-block mounted sub-router variables。
- Same-package `Mount("/prefix", routeFactory())`。
- Same-package receiver method route factories。
- 常见 middleware-wrapped handler calls。
- 标准库 `net/http` mux routes，包括 `HandleFunc`、`Handle` 和 Go 1.22 method pattern。
- GORM models through Go AST。
- Go callgraph through Go AST。

Backlog：

- Cross-file route group prefix propagation。
- 没有直接 handler 参数的 middleware chains。
- selector 变量的 handler method type resolution。
- 带 runtime arguments 的 chi route factories。
- sqlc query file detection。
- GORM relation tags 和 foreign key metadata。
- Go receiver method names 和 cross-file callgraph linking。

## 质量记录格式

发现 parser 问题时，在 evaluation 文档中记录：

```txt
Repository:
Language/framework:
Issue type: false positive | false negative | low confidence
Expected:
Actual:
Candidate fixture:
Priority:
```
