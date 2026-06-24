# Release Gate 结果

**语言：** [English](RELEASE_GATE_RESULTS.md) | 简体中文

本文记录组合 release gate 的本地运行结果。release gate 覆盖默认 preflight、ask evaluation、release binary smoke、release manifest verification、真实仓库 benchmark 和真实仓库 evaluation quality gate。

## 最新运行

日期：2026-06-24

命令：

```powershell
powershell -ExecutionPolicy Bypass -File scripts\release-gate.ps1 -OutputDir eval\m96-release-gate -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5 -RepoCacheDir eval\release-gate\repo-cache -AskCasesPath docs\examples\ask-cases.example.json -SkipManifestBuild
```

状态：PASS

## 步骤摘要

| 步骤 | 状态 | 秒 |
|---|---:|---:|
| `go test ./...` | PASS | 4.42 |
| `go vet ./...` | PASS | 2.81 |
| 英文 analyze smoke | PASS | 0.24 |
| 中文 analyze smoke | PASS | 0.25 |
| 真实仓库 benchmark | PASS | 1.92 |
| 真实仓库 evaluation | PASS | 6.03 |
| Ask evaluation | PASS | 0.22 |
| Release artifact smoke | PASS | 9.45 |

## Ask Evaluation 摘要

Provider：offline。

Strict：true。

Case source：`docs/examples/ask-cases.example.json`。

Overall score：1.0。

| Case | Score |
|---|---:|
| api-login-external | 1.0 |
| call-payment-zh-external | 1.0 |

## Benchmark 摘要

目标：每个仓库低于 30 秒。

| 仓库 | 秒 | 低于目标 | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|---:|
| Laravel | 0.22 | true | 1 | 0 | 0 |
| Spring REST service | 0.15 | true | 1 | 0 | 0 |
| Gin examples | 0.22 | true | 69 | 0 | 748 |
| FastAPI full-stack template | 0.25 | true | 23 | 2 | 851 |
| Prisma examples | 0.57 | true | 55 | 143 | 1764 |

## Evaluation 摘要

最低质量分：1.0。

| 仓库 | Quality | Routes | Models | Call Edges |
|---|---:|---:|---:|---:|
| Laravel | 1.00 | 1 | 0 | 0 |
| Spring REST service | 1.00 | 1 | 0 | 0 |
| Gin examples | 1.00 | 69 | 0 | 748 |
| Go chi | 1.00 | 229 | 0 | 1805 |
| FastAPI full-stack template | 1.00 | 23 | 2 | 851 |
| Node Express RealWorld | 1.00 | 20 | 4 | 99 |
| Prisma examples | 1.00 | 29 | 143 | 1764 |
| Symfony demo | 1.00 | 0 | 0 | 26 |
| Spring PetClinic | 1.00 | 18 | 6 | 0 |
| Spring Data JPA | 1.00 | 0 | 1 | 0 |
| Labstack Echo | 1.00 | 237 | 0 | 5000 |
| GoFiber Recipes | 1.00 | 278 | 49 | 5000 |
| Go GORM Playground | 1.00 | 0 | 6 | 24 |
| Django Oscar | 1.00 | 52 | 79 | 5000 |
| NestJS Starter | 1.00 | 1 | 0 | 4 |
| Next SaaS Starter | 1.00 | 0 | 0 | 284 |
| Vue RealWorld | 1.00 | 0 | 0 | 73 |
| React RealWorld | 1.00 | 0 | 0 | 176 |
| TypeORM Sample | 1.00 | 0 | 0 | 15 |
| Cookiecutter Django | 1.00 | 17 | 0 | 571 |

## Release Artifact Smoke

Release artifact smoke 已通过。本次通过 `-SkipManifestBuild` 明确跳过 manifest build 和 verification；最近一次完整 manifest verification 仍是 M80 运行。

## 备注

- benchmark/evaluation 通过 `RepoCacheDir` 共享 repository cache。
- 最新运行包含 20 个真实仓库 evaluation 样本。
- 最新运行包含 offline strict ask evaluation，共 2 个外部示例 case。
- 原始输出位于被 Git 忽略的 `eval/m96-release-gate/`。
