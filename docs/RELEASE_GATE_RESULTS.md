# Release Gate Results

**Language:** English | [简体中文](RELEASE_GATE_RESULTS.zh-CN.md)

This document records local release gate runs that combine default preflight, ask evaluation, release binary smoke, release manifest verification, real repository benchmark, and real repository evaluation quality checks.

## Latest Run

Date: 2026-06-24

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\release-gate.ps1 -OutputDir eval\m80-release-gate -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5 -RepoCacheDir eval\release-gate\repo-cache
```

Status: PASS

## Step Summary

| Step | Status | Seconds |
|---|---:|---:|
| `go test ./...` | PASS | 2.40 |
| `go vet ./...` | PASS | 1.39 |
| English analyze smoke | PASS | 0.22 |
| Chinese analyze smoke | PASS | 0.22 |
| Real repository benchmark | PASS | 1.93 |
| Real repository evaluation | PASS | 2.35 |
| Ask evaluation | PASS | 2.81 |
| Release artifact smoke | PASS | 9.19 |
| Release manifest build and verify | PASS | 13.31 |

## Ask Evaluation Summary

Provider: offline.

Strict: true.

Case source: built-in.

Overall score: 1.0.

| Case | Score |
|---|---:|
| api-login | 1.0 |
| api-wallet | 1.0 |
| self-cli-ask | 1.0 |
| db-wallet-model | 1.0 |
| db-models-zh | 1.0 |
| call-payment | 1.0 |
| call-payment-zh | 1.0 |

## Benchmark Summary

Target: 30 seconds per repository.

| Repository | Seconds | Under Target | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|---:|
| Laravel | 0.22 | true | 1 | 0 | 0 |
| Spring REST service | 0.15 | true | 1 | 0 | 0 |
| Gin examples | 0.22 | true | 68 | 0 | 748 |
| FastAPI full-stack template | 0.27 | true | 23 | 2 | 851 |
| Prisma examples | 0.54 | true | 55 | 143 | 1764 |

## Evaluation Summary

Minimum quality score: 1.0.

| Repository | Quality | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|
| Laravel | 1.00 | 1 | 0 | 0 |
| Spring REST service | 1.00 | 1 | 0 | 0 |
| Gin examples | 1.00 | 68 | 0 | 748 |
| Go chi | 1.00 | 210 | 0 | 1805 |
| FastAPI full-stack template | 1.00 | 23 | 2 | 851 |
| Node Express RealWorld | 1.00 | 20 | 4 | 99 |
| Prisma examples | 1.00 | 55 | 143 | 1764 |

## Manifest Verification

Release manifest verification passed for all six release archives:

- Windows amd64
- Windows arm64
- macOS amd64
- macOS arm64
- Linux amd64
- Linux arm64

## Notes

- Benchmark/evaluation share a repository cache through `RepoCacheDir`.
- The latest run includes 7 real repository evaluation samples and release manifest verification.
- The latest run includes built-in offline strict ask evaluation with 7 cases.
- Raw run outputs are under ignored `eval/m80-release-gate/`.
