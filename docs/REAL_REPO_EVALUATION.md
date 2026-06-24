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

The summary includes a `quality_score` from 0 to 1. The current score checks expected stack terms, minimum route/model counts, and selected known route/model names for each sample repository. Known checks currently include representative Gin routes (`/book`, `/bookable`), FastAPI mounted routes (`/api/v1/items`, `/api/v1/users`, `/api/v1/users/{user_id}`, `/api/v1/users/signup`, `/api/v1/utils/health-check`), Express split-router routes (`/api/tags`, `/api/articles`, `/api/articles/feed`, `/api/profiles/:username`, `/api/users/login`), and Prisma route/model names (`/api/filterPosts`, `/api/users`, `Account`, `Comment`, `Location`). It is a lightweight regression signal, not a replacement for manual parser review.

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

## Evaluation Notes

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
