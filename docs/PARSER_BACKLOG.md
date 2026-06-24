# Parser Improvement Backlog

**Language:** English | [简体中文](PARSER_BACKLOG.zh-CN.md)

RepoMind should keep deterministic scanning first. AI can summarize and infer, but parser quality must improve through explicit fixtures, real repository evaluation, and measured false positive / false negative notes.

## Priority Rules

Use this order when choosing parser work:

1. Fix high-confidence false positives that pollute first-run reports.
2. Fix missing routes/models in popular frameworks.
3. Improve callgraph edges for business workflows.
4. Add evidence and confidence to any new parser output.
5. Add fixture coverage before expanding real repository evaluation.

## Tree-sitter Evaluation Boundary

Tree-sitter can be useful for JS/TS, Python, PHP, and Java, but it should not be added blindly.

Adopt tree-sitter only when it satisfies:

- Cross-platform build remains simple.
- CI stays stable on Windows, macOS, and Linux.
- Parser code is materially simpler or more accurate than existing code.
- Benchmarks stay under the 30-second target.
- New dependencies do not block single-binary release.

For Go, prefer the standard library AST unless tree-sitter provides a clear advantage.

Cross-file route prefix propagation is tracked in `ROUTE_PREFIX_STRATEGY.md`.

## JavaScript / TypeScript

Current coverage:

- Express routes.
- Express multi-line route calls where the path and middleware are split across lines.
- Same-file Express `app.use("/prefix", router)` prefix propagation.
- Cross-file Express relative router imports for direct `require(...)`, default `import`, and named router imports.
- Express composed router prefix for `Router().use("/api", api)` wrapping imported controller routers.
- NestJS controllers.
- TypeORM entities.
- Lightweight callgraph.

Backlog:

- Express nested routers across files beyond direct one-hop relative imports and simple composed router exports.
- Router variables imported through alias chains or dynamic exports.
- Express route calls whose path is built dynamically or spans more than the lightweight parser window.
- NestJS controller prefix plus method decorator edge cases.
- Next.js App Router route handlers.
- Next.js Pages Router API routes.
- Hono routes.
- tRPC routers.
- TypeORM relation decorators with inverse side functions.
- Prisma schema cross-file examples.
- JS/TS callgraph for class methods and imported service calls.

## Python

Current coverage:

- Django URL patterns.
- Same-file Django `include()` prefix propagation for local pattern lists.
- Django `include("module.urls")` module prefix propagation.
- Django REST Framework same-file router registrations through `router.register(...)` and `include(router.urls)`.
- Django REST Framework custom actions on statically registered ViewSets, including `detail`, `methods`, `url_path`, and one-hop `include("module.urls")` parent prefixes.
- FastAPI decorators.
- FastAPI multi-line decorators where the route path is on a following line.
- FastAPI `APIRouter(prefix=...)` and same-file `include_router(..., prefix=...)` prefix propagation.
- FastAPI imported router prefix for direct static `from ... import router as alias` patterns.
- FastAPI composed router prefix for module imports such as `items.router` and static settings constants such as `settings.API_V1_STR`.
- Django models.
- SQLAlchemy models.
- SQLModel table models.
- Lightweight callgraph.

Backlog:

- Advanced Django includes with namespaces, app names, or dynamic include targets.
- Django REST Framework cross-file routers beyond one-hop static module includes and statically registered ViewSets.
- Cross-file FastAPI module imports beyond direct static router imports and unique static prefix constants.
- FastAPI decorators whose path is built dynamically or spans more than the lightweight parser window.
- SQLAlchemy 2.0 `Mapped[]` and `mapped_column`.
- Alembic model hints.
- Celery task discovery.
- Python callgraph for class methods and imported services.

## PHP

Current coverage:

- Laravel routes.
- Laravel static route groups and prefixes for chained `Route::prefix(...)->group(...)` and array `Route::group(["prefix" => ...], ...)` forms.
- Laravel/Symfony/ThinkPHP stack detection.

Backlog:

- Laravel dynamic route groups and prefixes beyond static string prefixes.
- Laravel controller array syntax edge cases beyond direct `[Controller::class, "method"]` handlers.
- Laravel resource routes.
- Laravel Eloquent models.
- Symfony controller attributes.
- ThinkPHP route definitions.
- PHP service callgraph.

## Java

Current coverage:

- Spring controller routes.
- JPA entities.
- Spring Boot stack detection.

Backlog:

- Controller-level `@RequestMapping` arrays.
- `@RequestParam`, `@PathVariable`, and method metadata.
- Spring WebFlux routes.
- MyBatis mapper interfaces and XML.
- JPA relation annotations with join columns.
- Java callgraph for service methods.

## Go

Current coverage:

- Gin/Echo/Fiber-style route calls through Go AST.
- Same-file chi-style `Route("/prefix", func(...))` prefix propagation.
- Same-block mounted sub-router variables such as `api := NewRouter(); api.Get(...); r.Mount("/api", api)`.
- Same-package `Mount("/prefix", routeFactory())` route factory prefix propagation.
- Same-package `Mount("/prefix", resource{}.Routes())` receiver method route factory prefix propagation.
- Common middleware-wrapped handler calls such as `requireAuth(handler)` and `middleware.Require(controller.Action)`.
- Standard library `net/http` mux routes, including `HandleFunc`, `Handle`, and Go 1.22 method patterns such as `GET /path`.
- GORM models through Go AST.
- Go callgraph through Go AST.

Backlog:

- Cross-file route group prefix propagation beyond direct same-package `Mount` factories and receiver method factories.
- Middleware chains or wrapper expressions without a direct handler argument.
- Handler method type resolution from selector variables.
- Advanced chi route factories with runtime arguments.
- sqlc query file detection.
- GORM relation tags and foreign key metadata.
- Go callgraph receiver method names and cross-file linking.

## Quality Notes Format

When a parser issue is found, record it in the relevant evaluation document with:

```txt
Repository:
Language/framework:
Issue type: false positive | false negative | low confidence
Expected:
Actual:
Candidate fixture:
Priority:
```
