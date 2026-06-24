# Release Gate Results

**Language:** English | [简体中文](RELEASE_GATE_RESULTS.zh-CN.md)

This document records local release gate runs that combine default preflight, safety boundary verification, core report content smoke, ask evaluation, remote Git URL analyze smoke, release binary smoke, release manifest verification, real repository benchmark, and real repository evaluation quality checks.

## Latest Run

Date: 2026-06-24

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\release-gate.ps1 -OutputDir eval\m106-release-gate -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5 -RepoCacheDir eval\release-gate\repo-cache -AskCasesPath docs\examples\ask-cases.example.json
```

Status: PASS

## Step Summary

| Step | Status | Seconds |
|---|---:|---:|
| Safety boundary | PASS | 0.70 |
| `go test ./...` | PASS | 4.30 |
| `go vet ./...` | PASS | 3.21 |
| English analyze smoke | PASS | 0.31 |
| Chinese analyze smoke | PASS | 0.31 |
| Trace and diagnose smoke | PASS | 0.51 |
| Real repository benchmark | PASS | 1.91 |
| Real repository evaluation | PASS | 6.09 |
| Ask evaluation | PASS | 0.22 |
| Remote repository analyze smoke | PASS | 2.82 |
| Release artifact smoke | PASS | 9.30 |
| Release manifest build and verification | PASS | 12.68 |

## Ask Evaluation Summary

Provider: offline.

Strict: true.

Case source: `docs/examples/ask-cases.example.json`.

Overall score: 1.0.

| Case | Score |
|---|---:|
| api-login-external | 1.0 |
| call-payment-zh-external | 1.0 |

## Benchmark Summary

Target: 30 seconds per repository.

| Repository | Seconds | Under Target | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|---:|
| Laravel | 0.23 | true | 1 | 0 | 0 |
| Spring REST service | 0.16 | true | 1 | 0 | 0 |
| Gin examples | 0.20 | true | 69 | 0 | 748 |
| FastAPI full-stack template | 0.24 | true | 23 | 2 | 851 |
| Prisma examples | 0.58 | true | 42 | 145 | 1764 |

## Evaluation Summary

Minimum quality score: 1.0.

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

## Release Artifact Smoke And Manifest

Release artifact smoke passed.

Release manifest build and verification passed for all six archives:

| Archive | Exists | Size OK | SHA256 OK |
|---|---:|---:|---:|
| `repomind-v0.0.0-release-gate-windows-amd64.zip` | true | true | true |
| `repomind-v0.0.0-release-gate-windows-arm64.zip` | true | true | true |
| `repomind-v0.0.0-release-gate-darwin-amd64.tar.gz` | true | true | true |
| `repomind-v0.0.0-release-gate-darwin-arm64.tar.gz` | true | true | true |
| `repomind-v0.0.0-release-gate-linux-amd64.tar.gz` | true | true | true |
| `repomind-v0.0.0-release-gate-linux-arm64.tar.gz` | true | true | true |

## Notes

- Benchmark/evaluation share a repository cache through `RepoCacheDir`.
- The latest run includes 20 real repository evaluation samples.
- The latest run includes safety boundary verification for ignored generated files and likely secret patterns.
- The latest run verifies English and Chinese HTML reports contain project summary, database model, API route, call graph, and Mermaid content.
- The latest run includes offline strict ask evaluation with 2 external example cases.
- Raw run outputs are under ignored `eval/m106-release-gate/`.
