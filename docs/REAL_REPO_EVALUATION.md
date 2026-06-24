# Real Repository Evaluation

**Language:** English | [简体中文](REAL_REPO_EVALUATION.zh-CN.md)

This document records real open-source repository evaluation runs for RepoMind.

The goal is to measure whether `repomind analyze` produces useful context within the 30-second product target, and to identify parser false positives, false negatives, and missing framework patterns.

## Evaluation Script

Run:

```powershell
.\scripts\evaluate-repos.ps1
```

When GitHub access requires a local proxy:

```powershell
.\scripts\evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809
```

The script also reads `HTTPS_PROXY`, `HTTP_PROXY`, or `ALL_PROXY` from the process environment when `-Proxy` is not provided.
By default, the script exits nonzero if any repository fails to clone, fails to analyze, or has `quality_score` lower than `1.0`. Use `-MinimumQualityScore <value>` to adjust the gate.

The script clones representative public repositories into ignored `eval/repos/`, runs:

```bash
go run ./cmd/repomind analyze --output <eval-report-dir> <repo-dir>
```

and writes:

```txt
eval/summary.json
eval/summary.md
```

The summary includes a `quality_score` from 0 to 1. The current score checks expected stack terms, minimum route/model counts, and selected known route/model names for each sample repository. Known checks currently include representative Gin/Chi routes, FastAPI mounted routes, Express split-router routes, Prisma route/model names, Spring PetClinic routes and JPA entities, GORM models, Django Oscar models, and Cookiecutter Django URL entries. It is a lightweight regression signal, not a replacement for manual parser review.

`eval/` is intentionally ignored by Git.

## Current Repository Set

| Repo | Expected Coverage |
|---|---|
| `laravel/laravel` | PHP, Laravel routes/config |
| `spring-guides/gs-rest-service` | Java, Spring Boot controllers |
| `gin-gonic/examples` | Go, Gin routes |
| `go-chi/chi` | Go chi examples, mounted route factories |
| `fastapi/full-stack-fastapi-template` | FastAPI, React, Postgres |
| `gothinkster/node-express-realworld-example-app` | Express split routers, Prisma models |
| `prisma/prisma-examples` | Prisma, TypeScript |
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

## Latest Valid Result

The 2026-06-24 M110 release gate kept all 20 fixed samples at `MinimumQualityScore 1.0` while also verifying the core report content, ask evaluation, trace/diagnose smoke, remote repository analyze smoke, safety boundary, release artifact smoke, release version injection, and release manifest verification.

| Repository | Quality | Routes | Models | Call Edges |
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

## Evaluation Notes

### 2026-06-24 SQLAlchemy 2.0 Mapped Column Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m101-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\m95-repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | M100 Models | M101 Models | Finding |
|---|---:|---:|---|
| `django-oscar` | 79 | 70 | Removed 9 false-positive SQLAlchemy models from non-DB `class X(Base)` utility classes while retaining Django model coverage. |

Findings:

- SQLAlchemy parser now supports 2.0 typed `Mapped[] = mapped_column(...)` fields and `Mapped[] = relationship(...)` relations.
- Empty `DeclarativeBase` / `Base` classes without table names, fields, or relations are no longer emitted as SQLAlchemy DB models.
- The 20-sample evaluation gate remains green at quality score 1.0.

### 2026-06-24 Next.js API Route Handler Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m100-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\m95-repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | M99 Routes | M100 Routes | Finding |
|---|---:|---:|---|
| `next-saas-starter` | 0 | 4 | Next.js App Router handlers under `app/api/**/route.ts` are now extracted. |
| `prisma-examples` | 29 | 42 | Next.js examples inside the monorepo now contribute API routes. |

Findings:

- Next.js parser now derives route paths from `app/api/**/route.{js,jsx,ts,tsx}` and method exports such as `GET` and `POST`.
- Pages Router `pages/api/**` files are recognized, including static `req.method` / `request.method` checks and `case "METHOD"` branches.
- Dynamic route segments such as `[id]`, `[...slug]`, and `[[...slug]]` are normalized to Mermaid/API-map-friendly `{id}` / `{slug}` parameters.
- The 20-sample evaluation gate remains green at quality score 1.0.

### 2026-06-24 TypeORM Empty Entity Decorator Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m99-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\m95-repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | M98 Models | M99 Models | Finding |
|---|---:|---:|---|
| `typeorm-sample` | 0 | 2 | `@Entity()` without explicit table names now extracts `Post` and `Category`. |
| `prisma-examples` | 143 | 145 | TypeORM sample entities inside the monorepo now contribute models. |

Findings:

- TypeORM parser now keeps an empty `@Entity()` decorator pending until the following class declaration.
- Multiline relation decorator options no longer create fake fields from object literal option keys.
- The 20-sample evaluation gate remains green at quality score 1.0.

### 2026-06-24 Symfony Controller Attribute Route Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m98-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\m95-repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | M97 Routes | M98 Routes | Finding |
|---|---:|---:|---|
| `symfony-demo` | 0 | 19 | Symfony PHP 8 `#[Route(...)]` controller attributes are now extracted. |

Findings:

- Symfony parser now handles class-level route prefixes, method-level attributes, multiple attributes on one method, method filters, and typed route parameters such as `{slug:post}`.
- Known route checks now include `/blog`, `/blog/posts/{slug}`, `/blog/comment/{postSlug}/new`, `/admin/post`, and `/profile/edit`.
- The 20-sample evaluation gate remains green at quality score 1.0.

### 2026-06-24 FastAPI Test Decorator False Positive Guard Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m97-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\m95-repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | M96 Routes | M97 Routes | Finding |
|---|---:|---:|---|
| `django-oscar` | 52 | 8 | Removed FastAPI false positives from Python test patch decorators. |
| `cookiecutter-django` | 17 | 9 | Removed FastAPI false positives from non-FastAPI helper decorators. |

Findings:

- FastAPI route parsing now requires a FastAPI app/router signal before accepting `@router.get/post/...` decorators.
- Django projects with test `@patch(...)` decorators no longer emit FastAPI API routes.
- The 20-sample evaluation gate remains green at quality score 1.0.

### 2026-06-24 Express Client False Positive Guard Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m96-evaluation-final -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\m95-repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | M95 Routes | M96 Routes | Finding |
|---|---:|---:|---|
| `prisma-examples` | 55 | 29 | Removed route checks that depended on client/non-Express signals. |
| `gofiber-recipes` | 279 | 278 | One frontend-client false positive removed. |
| `vue-realworld` | 7 | 0 | Frontend API wrapper calls are no longer reported as Express API routes. |
| `react-realworld` | 4 | 0 | Frontend API wrapper calls are no longer reported as Express API routes. |

Findings:

- Express route parsing now requires an Express app/router signal before accepting `.get/.post/...` calls.
- Frontend-only repositories still contribute stack and callgraph coverage, but no longer pollute the API map with HTTP client wrapper calls.
- The 20-sample evaluation gate remains green at quality score 1.0.

### 2026-06-24 20 Sample Expansion Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m95-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\m95-repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | Expected | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |
|---|---|---:|---:|---:|---:|---|---|---|---:|---:|---:|
| `laravel` | Laravel, PHP | true | 1.00 | 1.58 | 63 | Laravel |  |  | 0 | 1 | 0 |
| `spring-rest-service` | Spring Boot, Java | true | 1.00 | 0.36 | 59 | Spring Boot |  |  | 0 | 1 | 0 |
| `gin-examples` | Gin, Go | true | 1.00 | 2.72 | 124 | Gin |  |  | 0 | 69 | 748 |
| `go-chi` | Go chi | true | 1.00 | 3.54 | 97 | Chi |  |  | 0 | 229 | 1805 |
| `fastapi-full-stack-template` | FastAPI, React | true | 1.00 | 4.98 | 227 | FastAPI | React | Postgres | 2 | 23 | 851 |
| `node-express-realworld` | Express, Prisma | true | 1.00 | 1.41 | 67 | Express |  |  | 4 | 20 | 99 |
| `prisma-examples` | Prisma, TypeScript | true | 1.00 | 19.44 | 1374 | NestJS, Express | Next.js, Vue, React | Postgres | 143 | 55 | 1764 |
| `symfony-demo` | Symfony, PHP | true | 1.00 | 3.17 | 243 | Symfony |  |  | 0 | 0 | 26 |
| `spring-petclinic` | Spring Boot, JPA | true | 1.00 | 1.24 | 127 | Spring Boot |  | Postgres, MySQL | 6 | 18 | 0 |
| `spring-data-jpa` | Spring Boot, JPA | true | 1.00 | 0.41 | 57 | Spring Boot |  |  | 1 | 0 | 0 |
| `labstack-echo` | Echo, Go | true | 1.00 | 4.58 | 131 | Echo |  |  | 0 | 237 | 5000 |
| `gofiber-recipes` | Fiber, Go | true | 1.00 | 11.49 | 1002 | Fiber | React | Postgres, MySQL, MongoDB | 49 | 279 | 5000 |
| `go-gorm-playground` | GORM, Go | true | 1.00 | 0.35 | 15 |  |  | Postgres, MySQL | 6 | 0 | 24 |
| `django-oscar` | Django, Python | true | 1.00 | 14.33 | 1573 | Django |  | Postgres | 79 | 52 | 5000 |
| `nestjs-starter` | NestJS, TypeScript | true | 1.00 | 0.27 | 16 | NestJS |  |  | 0 | 1 | 4 |
| `next-saas-starter` | Next.js, React | true | 1.00 | 1.38 | 55 |  | Next.js, React | Postgres | 0 | 0 | 284 |
| `vue-realworld` | Vue | true | 1.00 | 0.73 | 67 |  | Vue |  | 0 | 7 | 73 |
| `react-realworld` | React | true | 1.00 | 1.52 | 45 |  | React |  | 0 | 4 | 176 |
| `typeorm-sample` | Express, TypeORM | true | 1.00 | 0.27 | 14 | Express |  | MySQL | 0 | 0 | 15 |
| `cookiecutter-django` | Django template | true | 1.00 | 1.62 | 277 | Django |  |  | 0 | 17 | 571 |

Findings:

- The fixed evaluation matrix now covers 20 public repositories across PHP, Java, Go, Python, JavaScript/TypeScript, and frontend-only stacks.
- All sampled repositories completed under the 30-second product target.
- Frontend client samples are used as stack coverage only; their detected routes should not be treated as API map quality checks until client-side API false positives are tightened.
- Symfony route extraction, Symfony models, and TypeORM sample entity coverage remain parser improvement candidates.

### 2026-06-24 Proxied Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809
```

Direct GitHub clone without proxy failed in the current network with `github.com:443` connection reset or timeout. The proxied run completed successfully.

| Repo | Expected | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |
|---|---|---:|---:|---:|---:|---|---|---|---:|---:|---:|
| `laravel` | Laravel, PHP | true | 1.00 | 0.49 | 63 | Laravel |  |  | 0 | 1 | 0 |
| `spring-rest-service` | Spring Boot, Java | true | 1.00 | 0.23 | 59 | Spring Boot |  |  | 0 | 1 | 0 |
| `gin-examples` | Gin, Go | true | 1.00 | 0.94 | 124 | Gin |  |  | 0 | 66 | 748 |
| `fastapi-full-stack-template` | FastAPI, React | true | 1.00 | 1.84 | 227 | FastAPI | React | Postgres | 2 | 18 | 851 |
| `prisma-examples` | Prisma, TypeScript | true | 1.00 | 4.07 | 1374 | NestJS, Express | Next.js, Vue, React | Postgres | 143 | 55 | 1764 |

Status: PASS with `-MinimumQualityScore 1.0`.

Findings:

- All sampled repositories completed well under the 30-second product target.
- Laravel and Spring sample projects are small skeletons, so the lack of DB models is expected.
- Gin route extraction works on a multi-example Go repository and produces a useful route map.
- FastAPI full-stack template produces the strongest MVP output: backend, frontend, database, models, routes, and call edges are all populated.
- The first M13 evidence run exposed a Python DB parser false positive: Pydantic `BaseModel`, `BaseSettings`, and non-table SQLModel schema classes were being treated as SQLAlchemy models. The parser was tightened so only SQLModel `table=True` classes are counted as SQLModel DB models. The FastAPI template now reports 2 DB models (`User`, `Item`) instead of 10 mixed classes.
- Package-level grouping is now available in `analysis.json`.
- FastAPI full-stack template now reports package entries for root, `backend`, and `frontend`.
- Prisma examples now reports package entries for the root package, individual starter packages, and nested `prisma` schema directories. The global stack remains broad by design, but package rows provide local context.
- Quality score is now written to `eval/summary.json` and `eval/summary.md`.
- Quality checks now include selected known route/model names such as `/greeting`, `/book`, `/bookable`, `/api/v1/login/access-token`, `/api/v1/users/me`, `/api/v1/items`, `/api/v1/users/signup`, `/api/v1/utils/health-check`, `/api/feed`, `/api/filterPosts`, `/api/users`, `User`, `Item`, `Post`, `Account`, `Comment`, and `Location`.

Parser improvement tasks:

- Track parser false positives and false negatives using `docs/PARSER_BACKLOG.md`.
- Add more quality checks for expected package counts and cross-file route prefixes.
- Keep this evaluation script as a manual regression gate before releases.

### 2026-06-24 Resolver Regression Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m52-resolver-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

Comparison against the M50 expanded-check run:

| Repo | M50 Routes | M52 Routes | M50 Models | M52 Models | M52 Quality |
|---|---:|---:|---:|---:|---:|
| `laravel` | 1 | 1 | 0 | 0 | 1.00 |
| `spring-rest-service` | 1 | 1 | 0 | 0 | 1.00 |
| `gin-examples` | 66 | 66 | 0 | 0 | 1.00 |
| `fastapi-full-stack-template` | 18 | 18 | 2 | 2 | 1.00 |
| `prisma-examples` | 55 | 55 | 143 | 143 | 1.00 |

Findings:

- FastAPI imported router prefix and Express relative router prefix did not regress the current real repository sample set.
- The sampled repositories do not yet expose a measurable route count increase for the new cross-file prefix resolvers, so focused fixtures remain the primary coverage for these patterns.
- Future evaluation expansion should add real repositories that use split FastAPI routers and split Express routers.

### 2026-06-24 Split Express Router Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m54-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | Expected | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |
|---|---|---:|---:|---:|---:|---|---|---|---:|---:|---:|
| `laravel` | Laravel, PHP | true | 1.00 | 1.85 | 63 | Laravel |  |  | 0 | 1 | 0 |
| `spring-rest-service` | Spring Boot, Java | true | 1.00 | 0.17 | 59 | Spring Boot |  |  | 0 | 1 | 0 |
| `gin-examples` | Gin, Go | true | 1.00 | 0.19 | 124 | Gin |  |  | 0 | 66 | 748 |
| `fastapi-full-stack-template` | FastAPI, React | true | 1.00 | 0.24 | 227 | FastAPI | React | Postgres | 2 | 18 | 851 |
| `node-express-realworld` | Express, Prisma | true | 1.00 | 0.29 | 67 | Express |  |  | 4 | 8 | 99 |
| `prisma-examples` | Prisma, TypeScript | true | 1.00 | 0.53 | 1374 | NestJS, Express | Next.js, Vue, React | Postgres | 143 | 55 | 1764 |

Findings:

- The new `node-express-realworld` sample covers split Express routers with controller files imported into a composed router and mounted under `/api`.
- The Express parser now resolves dotted TypeScript import basenames such as `./tag/tag.controller`.
- Known checks for this sample include `/api/tags`, `/api/articles`, `/api/users/login`, and Prisma models `Article`, `Comment`, `Tag`, and `User`.

### 2026-06-24 FastAPI Mounted Prefix Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m55-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | Expected | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |
|---|---|---:|---:|---:|---:|---|---|---|---:|---:|---:|
| `laravel` | Laravel, PHP | true | 1.00 | 1.86 | 63 | Laravel |  |  | 0 | 1 | 0 |
| `spring-rest-service` | Spring Boot, Java | true | 1.00 | 0.17 | 59 | Spring Boot |  |  | 0 | 1 | 0 |
| `gin-examples` | Gin, Go | true | 1.00 | 0.22 | 124 | Gin |  |  | 0 | 66 | 748 |
| `fastapi-full-stack-template` | FastAPI, React | true | 1.00 | 0.24 | 227 | FastAPI | React | Postgres | 2 | 18 | 851 |
| `node-express-realworld` | Express, Prisma | true | 1.00 | 0.18 | 67 | Express |  |  | 4 | 8 | 99 |
| `prisma-examples` | Prisma, TypeScript | true | 1.00 | 0.50 | 1374 | NestJS, Express | Next.js, Vue, React | Postgres | 143 | 55 | 1764 |

Findings:

- FastAPI full-stack template routes now include the mounted `/api/v1` prefix from `settings.API_V1_STR`.
- Known checks for this sample were updated from unmounted paths to mounted paths such as `/api/v1/users/me`.

### 2026-06-24 Split Go Router Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m56-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | Expected | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |
|---|---|---:|---:|---:|---:|---|---|---|---:|---:|---:|
| `laravel` | Laravel, PHP | true | 1.00 | 0.19 | 63 | Laravel |  |  | 0 | 1 | 0 |
| `spring-rest-service` | Spring Boot, Java | true | 1.00 | 0.16 | 59 | Spring Boot |  |  | 0 | 1 | 0 |
| `gin-examples` | Gin, Go | true | 1.00 | 0.19 | 124 | Gin |  |  | 0 | 66 | 748 |
| `go-chi` | Go chi | true | 1.00 | 1.02 | 97 |  |  |  | 0 | 210 | 1805 |
| `fastapi-full-stack-template` | FastAPI, React | true | 1.00 | 0.25 | 227 | FastAPI | React | Postgres | 2 | 18 | 851 |
| `node-express-realworld` | Express, Prisma | true | 1.00 | 0.20 | 67 | Express |  |  | 4 | 8 | 99 |
| `prisma-examples` | Prisma, TypeScript | true | 1.00 | 0.54 | 1374 | NestJS, Express | Next.js, Vue, React | Postgres | 143 | 55 | 1764 |

Findings:

- The new `go-chi` sample covers mounted Go route factories such as `/admin/accounts` and `/admin/users/{userId}` from chi examples.
- The route parser handles this sample. Chi backend stack detection was added in the follow-up M57 run.

### 2026-06-24 Chi Stack Detection Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m57-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | Expected | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |
|---|---|---:|---:|---:|---:|---|---|---|---:|---:|---:|
| `laravel` | Laravel, PHP | true | 1.00 | 7.39 | 63 | Laravel |  |  | 0 | 1 | 0 |
| `spring-rest-service` | Spring Boot, Java | true | 1.00 | 1.80 | 59 | Spring Boot |  |  | 0 | 1 | 0 |
| `gin-examples` | Gin, Go | true | 1.00 | 0.20 | 124 | Gin |  |  | 0 | 66 | 748 |
| `go-chi` | Go chi | true | 1.00 | 0.22 | 97 | Chi |  |  | 0 | 210 | 1805 |
| `fastapi-full-stack-template` | FastAPI, React | true | 1.00 | 0.29 | 227 | FastAPI | React | Postgres | 2 | 18 | 851 |
| `node-express-realworld` | Express, Prisma | true | 1.00 | 0.19 | 67 | Express |  |  | 4 | 8 | 99 |
| `prisma-examples` | Prisma, TypeScript | true | 1.00 | 0.53 | 1374 | NestJS, Express | Next.js, Vue, React | Postgres | 143 | 55 | 1764 |

Findings:

- `go-chi` now reports backend `Chi`.
- The evaluation quality gate now checks `Chi` as the expected stack term for this sample.

### 2026-06-24 Express Multi-line Route Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m58-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | Expected | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |
|---|---|---:|---:|---:|---:|---|---|---|---:|---:|---:|
| `laravel` | Laravel, PHP | true | 1.00 | 1.91 | 63 | Laravel |  |  | 0 | 1 | 0 |
| `spring-rest-service` | Spring Boot, Java | true | 1.00 | 0.16 | 59 | Spring Boot |  |  | 0 | 1 | 0 |
| `gin-examples` | Gin, Go | true | 1.00 | 0.19 | 124 | Gin |  |  | 0 | 66 | 748 |
| `go-chi` | Go chi | true | 1.00 | 0.20 | 97 | Chi |  |  | 0 | 210 | 1805 |
| `fastapi-full-stack-template` | FastAPI, React | true | 1.00 | 0.26 | 227 | FastAPI | React | Postgres | 2 | 18 | 851 |
| `node-express-realworld` | Express, Prisma | true | 1.00 | 0.20 | 67 | Express |  |  | 4 | 20 | 99 |
| `prisma-examples` | Prisma, TypeScript | true | 1.00 | 0.54 | 1374 | NestJS, Express | Next.js, Vue, React | Postgres | 143 | 55 | 1764 |

Findings:

- Express multi-line route parsing increased `node-express-realworld` route coverage from 8 to 20.
- Known checks for this sample now include `/api/articles/feed`, `/api/articles/:slug/comments`, and `/api/profiles/:username`.

### 2026-06-24 Go Receiver Method Factory Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m59-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | Expected | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |
|---|---|---:|---:|---:|---:|---|---|---|---:|---:|---:|
| `laravel` | Laravel, PHP | true | 1.00 | 1.91 | 63 | Laravel |  |  | 0 | 1 | 0 |
| `spring-rest-service` | Spring Boot, Java | true | 1.00 | 0.16 | 59 | Spring Boot |  |  | 0 | 1 | 0 |
| `gin-examples` | Gin, Go | true | 1.00 | 0.19 | 124 | Gin |  |  | 0 | 66 | 748 |
| `go-chi` | Go chi | true | 1.00 | 0.24 | 97 | Chi |  |  | 0 | 210 | 1805 |
| `fastapi-full-stack-template` | FastAPI, React | true | 1.00 | 0.26 | 227 | FastAPI | React | Postgres | 2 | 18 | 851 |
| `node-express-realworld` | Express, Prisma | true | 1.00 | 0.21 | 67 | Express |  |  | 4 | 20 | 99 |
| `prisma-examples` | Prisma, TypeScript | true | 1.00 | 0.54 | 1374 | NestJS, Express | Next.js, Vue, React | Postgres | 143 | 55 | 1764 |

Findings:

- Go route factory resolution now covers receiver methods such as `usersResource{}.Routes()`.
- Known checks for `go-chi` now include `/users/{id}` and `/todos/{id}/sync`.

### 2026-06-24 FastAPI Multi-line Decorator Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m61-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | Expected | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |
|---|---|---:|---:|---:|---:|---|---|---|---:|---:|---:|
| `laravel` | Laravel, PHP | true | 1.00 | 3.12 | 63 | Laravel |  |  | 0 | 1 | 0 |
| `spring-rest-service` | Spring Boot, Java | true | 1.00 | 0.36 | 59 | Spring Boot |  |  | 0 | 1 | 0 |
| `gin-examples` | Gin, Go | true | 1.00 | 2.37 | 124 | Gin |  |  | 0 | 66 | 748 |
| `go-chi` | Go chi | true | 1.00 | 3.15 | 97 | Chi |  |  | 0 | 210 | 1805 |
| `fastapi-full-stack-template` | FastAPI, React | true | 1.00 | 0.29 | 227 | FastAPI | React | Postgres | 2 | 23 | 851 |
| `node-express-realworld` | Express, Prisma | true | 1.00 | 0.19 | 67 | Express |  |  | 4 | 20 | 99 |
| `prisma-examples` | Prisma, TypeScript | true | 1.00 | 0.53 | 1374 | NestJS, Express | Next.js, Vue, React | Postgres | 143 | 55 | 1764 |

Findings:

- FastAPI multi-line decorator parsing increased `fastapi-full-stack-template` route coverage from 18 to 23.
- Known checks for this sample now include `/api/v1/users` and `/api/v1/users/{user_id}`.

### 2026-06-24 Go Mounted Sub-router Variable Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m70-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | Expected | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |
|---|---|---:|---:|---:|---:|---|---|---|---:|---:|---:|
| `laravel` | Laravel, PHP | true | 1.00 | 1.93 | 63 | Laravel |  |  | 0 | 1 | 0 |
| `spring-rest-service` | Spring Boot, Java | true | 1.00 | 0.21 | 59 | Spring Boot |  |  | 0 | 1 | 0 |
| `gin-examples` | Gin, Go | true | 1.00 | 0.24 | 124 | Gin |  |  | 0 | 66 | 748 |
| `go-chi` | Go chi | true | 1.00 | 0.22 | 97 | Chi |  |  | 0 | 210 | 1805 |
| `fastapi-full-stack-template` | FastAPI, React | true | 1.00 | 0.27 | 227 | FastAPI | React | Postgres | 2 | 23 | 851 |
| `node-express-realworld` | Express, Prisma | true | 1.00 | 0.21 | 67 | Express |  |  | 4 | 20 | 99 |
| `prisma-examples` | Prisma, TypeScript | true | 1.00 | 0.59 | 1374 | NestJS, Express | Next.js, Vue, React | Postgres | 143 | 55 | 1764 |

Findings:

- Go parser now resolves same-block mounted sub-router variables such as `api := NewRouter(); api.Get(...); r.Mount("/api", api)`.
- `go-chi` remains at quality score 1.00 with 210 detected routes.

### 2026-06-24 Go Middleware-wrapped Handler Run

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m71-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0
```

Status: PASS with `-MinimumQualityScore 1.0`.

| Repo | Expected | Analyze | Quality | Seconds | Files | Backend | Frontend | Database | Models | Routes | Call Edges |
|---|---|---:|---:|---:|---:|---|---|---|---:|---:|---:|
| `laravel` | Laravel, PHP | true | 1.00 | 5.62 | 63 | Laravel |  |  | 0 | 1 | 0 |
| `spring-rest-service` | Spring Boot, Java | true | 1.00 | 1.88 | 59 | Spring Boot |  |  | 0 | 1 | 0 |
| `gin-examples` | Gin, Go | true | 1.00 | 0.20 | 124 | Gin |  |  | 0 | 68 | 748 |
| `go-chi` | Go chi | true | 1.00 | 0.21 | 97 | Chi |  |  | 0 | 210 | 1805 |
| `fastapi-full-stack-template` | FastAPI, React | true | 1.00 | 0.25 | 227 | FastAPI | React | Postgres | 2 | 23 | 851 |
| `node-express-realworld` | Express, Prisma | true | 1.00 | 0.19 | 67 | Express |  |  | 4 | 20 | 99 |
| `prisma-examples` | Prisma, TypeScript | true | 1.00 | 0.60 | 1374 | NestJS, Express | Next.js, Vue, React | Postgres | 143 | 55 | 1764 |

Findings:

- Go parser now extracts handler names from common wrapper calls such as `requireAuth(handler)` and `middleware.Require(controller.Action)`.
- `gin-examples` route coverage increased from 66 to 68 without failing the quality gate.
