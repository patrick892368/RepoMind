# Release Gate Results

**Language:** English | [简体中文](RELEASE_GATE_RESULTS.zh-CN.md)

This document records local release gate runs that combine default preflight, ask evaluation, release binary smoke, release manifest verification, real repository benchmark, and real repository evaluation quality checks.

## Latest Run

Date: 2026-06-24

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\release-gate.ps1 -OutputDir eval\m98-release-gate -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5 -RepoCacheDir eval\release-gate\repo-cache -AskCasesPath docs\examples\ask-cases.example.json -SkipManifestBuild
```

Status: PASS

## Step Summary

| Step | Status | Seconds |
|---|---:|---:|
| `go test ./...` | PASS | 4.40 |
| `go vet ./...` | PASS | 2.88 |
| English analyze smoke | PASS | 0.25 |
| Chinese analyze smoke | PASS | 0.24 |
| Real repository benchmark | PASS | 2.03 |
| Real repository evaluation | PASS | 5.78 |
| Ask evaluation | PASS | 0.22 |
| Release artifact smoke | PASS | 9.40 |

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
| Laravel | 0.22 | true | 1 | 0 | 0 |
| Spring REST service | 0.15 | true | 1 | 0 | 0 |
| Gin examples | 0.22 | true | 69 | 0 | 748 |
| FastAPI full-stack template | 0.25 | true | 23 | 2 | 851 |
| Prisma examples | 0.57 | true | 55 | 143 | 1764 |

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
| Prisma examples | 1.00 | 29 | 143 | 1764 |
| Symfony demo | 1.00 | 19 | 0 | 26 |
| Spring PetClinic | 1.00 | 18 | 6 | 0 |
| Spring Data JPA | 1.00 | 0 | 1 | 0 |
| Labstack Echo | 1.00 | 237 | 0 | 5000 |
| GoFiber Recipes | 1.00 | 278 | 49 | 5000 |
| Go GORM Playground | 1.00 | 0 | 6 | 24 |
| Django Oscar | 1.00 | 8 | 79 | 5000 |
| NestJS Starter | 1.00 | 1 | 0 | 4 |
| Next SaaS Starter | 1.00 | 0 | 0 | 284 |
| Vue RealWorld | 1.00 | 0 | 0 | 73 |
| React RealWorld | 1.00 | 0 | 0 | 176 |
| TypeORM Sample | 1.00 | 0 | 0 | 15 |
| Cookiecutter Django | 1.00 | 9 | 0 | 571 |

## Release Artifact Smoke

Release artifact smoke passed. Manifest build and verification were intentionally skipped in this run with `-SkipManifestBuild`; the latest full manifest verification remains the M80 run.

## Notes

- Benchmark/evaluation share a repository cache through `RepoCacheDir`.
- The latest run includes 20 real repository evaluation samples.
- The latest run includes offline strict ask evaluation with 2 external example cases.
- Raw run outputs are under ignored `eval/m98-release-gate/`.
