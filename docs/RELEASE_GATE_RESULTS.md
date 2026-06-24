# Release Gate Results

**Language:** English | [简体中文](RELEASE_GATE_RESULTS.zh-CN.md)

This document records local release gate runs that combine default preflight, release binary smoke, release manifest verification, real repository benchmark, and real repository evaluation quality checks.

## Latest Run

Date: 2026-06-24

Command:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5 -RepoCacheDir eval\release-gate\repo-cache
```

Status: PASS

## Step Summary

| Step | Status | Seconds |
|---|---:|---:|
| `go test ./...` | PASS | 2.30 |
| `go vet ./...` | PASS | 1.35 |
| English analyze smoke | PASS | 0.25 |
| Chinese analyze smoke | PASS | 0.24 |
| Real repository benchmark | PASS | 2.16 |
| Real repository evaluation | PASS | 2.63 |
| Release artifact smoke | PASS | 8.96 |
| Release manifest build and verify | PASS | 12.76 |

## Benchmark Summary

Target: 30 seconds per repository.

| Repository | Seconds | Under Target | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|---:|
| Laravel | 0.21 | true | 1 | 0 | 0 |
| Spring REST service | 0.17 | true | 1 | 0 | 0 |
| Gin examples | 0.28 | true | 68 | 0 | 748 |
| FastAPI full-stack template | 0.36 | true | 23 | 2 | 851 |
| Prisma examples | 0.61 | true | 55 | 143 | 1764 |

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

- The first release gate attempt failed due to GitHub clone connection resets.
- Benchmark/evaluation now share a repository cache through `RepoCacheDir`.
- Re-running with `-CloneRetries 5` passed.
- The latest run includes 7 real repository evaluation samples and release manifest verification.
- The latest run includes remote Git URL/ref/cache support and Go route parser improvements through M71.
- Raw run outputs are under ignored `eval/release-gate/`.
