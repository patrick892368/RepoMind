# Release Gate 结果

**语言：** [English](RELEASE_GATE_RESULTS.md) | 简体中文

本文记录组合 release gate 的本地运行结果。release gate 覆盖默认 preflight、safety boundary verification、核心报告内容 smoke、ask evaluation、远程 Git URL analyze smoke、release binary smoke、release version injection、release manifest verification、真实仓库 benchmark 和真实仓库 evaluation quality gate。

## 最新运行

日期：2026-06-24

命令：

```powershell
powershell -ExecutionPolicy Bypass -File scripts\release-gate.ps1 -OutputDir eval\m110-release-gate -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5 -RepoCacheDir eval\release-gate\repo-cache -AskCasesPath docs\examples\ask-cases.example.json
```

状态：PASS

## 步骤摘要

| 步骤 | 状态 | 秒 |
|---|---:|---:|
| Safety boundary | PASS | 0.71 |
| `go test ./...` | PASS | 4.44 |
| `go vet ./...` | PASS | 3.29 |
| 英文 analyze smoke | PASS | 7.18 |
| 中文 analyze smoke | PASS | 2.01 |
| Trace and diagnose smoke | PASS | 0.49 |
| 真实仓库 benchmark | PASS | 1.93 |
| 真实仓库 evaluation | PASS | 5.95 |
| Ask evaluation | PASS | 0.21 |
| Remote repository analyze smoke | PASS | 1.43 |
| Release artifact smoke | PASS | 9.06 |
| Release manifest build and verification | PASS | 12.22 |

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
| Spring REST service | 0.17 | true | 1 | 0 | 0 |
| Gin examples | 0.19 | true | 69 | 0 | 748 |
| FastAPI full-stack template | 0.25 | true | 23 | 2 | 851 |
| Prisma examples | 0.55 | true | 42 | 145 | 1764 |

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

## Release Artifact Smoke 和 Manifest

Release artifact smoke 已通过。

Release manifest build and verification 已通过，覆盖全部 6 个归档：

| Archive | Exists | Size OK | SHA256 OK |
|---|---:|---:|---:|
| `repomind-v0.0.0-release-gate-windows-amd64.zip` | true | true | true |
| `repomind-v0.0.0-release-gate-windows-arm64.zip` | true | true | true |
| `repomind-v0.0.0-release-gate-darwin-amd64.tar.gz` | true | true | true |
| `repomind-v0.0.0-release-gate-darwin-arm64.tar.gz` | true | true | true |
| `repomind-v0.0.0-release-gate-linux-amd64.tar.gz` | true | true | true |
| `repomind-v0.0.0-release-gate-linux-arm64.tar.gz` | true | true | true |

## 备注

- benchmark/evaluation 通过 `RepoCacheDir` 共享 repository cache。
- 最新运行包含 20 个真实仓库 evaluation 样本。
- 最新运行包含 safety boundary verification，检查 ignored 生成物路径和疑似真实密钥模式。
- 最新运行验证英文和中文 HTML 报告包含项目总结、数据库模型、API 路由、调用图和 Mermaid 内容。
- 最新运行验证 release artifact smoke version injection，输出 `repomind dev-smoke`。
- 最新运行包含 offline strict ask evaluation，共 2 个外部示例 case。
- 原始输出位于被 Git 忽略的 `eval/m110-release-gate/`。
