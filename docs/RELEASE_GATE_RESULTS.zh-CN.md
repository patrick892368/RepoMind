# Release Gate 结果

**语言：** [English](RELEASE_GATE_RESULTS.md) | 简体中文

本文记录组合 release gate 的本地运行结果。release gate 覆盖默认 preflight、release binary smoke、release manifest verification、真实仓库 benchmark 和真实仓库 evaluation quality gate。

## 最新运行

日期：2026-06-24

命令：

```powershell
powershell -ExecutionPolicy Bypass -File scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5 -RepoCacheDir eval\release-gate\repo-cache
```

状态：PASS

## 步骤摘要

| 步骤 | 状态 | 秒 |
|---|---:|---:|
| `go test ./...` | PASS | 2.30 |
| `go vet ./...` | PASS | 1.35 |
| 英文 analyze smoke | PASS | 0.25 |
| 中文 analyze smoke | PASS | 0.24 |
| 真实仓库 benchmark | PASS | 2.16 |
| 真实仓库 evaluation | PASS | 2.63 |
| Release artifact smoke | PASS | 8.96 |
| Release manifest build and verify | PASS | 12.76 |

## Benchmark 摘要

目标：每个仓库低于 30 秒。

| 仓库 | 秒 | 低于目标 | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|---:|
| Laravel | 0.21 | true | 1 | 0 | 0 |
| Spring REST service | 0.17 | true | 1 | 0 | 0 |
| Gin examples | 0.28 | true | 68 | 0 | 748 |
| FastAPI full-stack template | 0.36 | true | 23 | 2 | 851 |
| Prisma examples | 0.61 | true | 55 | 143 | 1764 |

## Evaluation 摘要

最低质量分：1.0。

| 仓库 | Quality | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|
| Laravel | 1.00 | 1 | 0 | 0 |
| Spring REST service | 1.00 | 1 | 0 | 0 |
| Gin examples | 1.00 | 68 | 0 | 748 |
| Go chi | 1.00 | 210 | 0 | 1805 |
| FastAPI full-stack template | 1.00 | 23 | 2 | 851 |
| Node Express RealWorld | 1.00 | 20 | 4 | 99 |
| Prisma examples | 1.00 | 55 | 143 | 1764 |

## Manifest Verification

所有六个 release archive 的 manifest verification 通过：

- Windows amd64
- Windows arm64
- macOS amd64
- macOS arm64
- Linux amd64
- Linux arm64

## 备注

- 最新运行包含 M71 之前的远程仓库输入、ref/cache 和 Go parser 增强。
- 原始输出位于被 Git 忽略的 `eval/release-gate/`。
