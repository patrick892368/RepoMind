# 发布检查清单

**语言：** [English](RELEASE_CHECKLIST.md) | 简体中文

发布 RepoMind 前必须完成本清单。

## 1. 工作区安全

确认 `.env` 被忽略：

```powershell
git check-ignore -v .env
```

确认生成目录被忽略：

```powershell
git check-ignore -v .repomind dist eval benchmark
```

确认没有 API key 进入待提交文件：

```powershell
git status --short
```

不要提交 `.env`、`eval/`、`benchmark/`、`.repomind/` 或 `dist/`。

## 2. 测试和集成检查

默认 preflight：

```powershell
.\scripts\preflight.ps1
```

包含 release smoke 的 preflight：

```powershell
.\scripts\preflight.ps1 -IncludeReleaseSmoke
```

完整本地 release gate：

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809
```

GitHub clone 不稳定时复用 cache：

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval/repo-cache -CloneRetries 5
```

核心检查：

```powershell
go test ./...
go vet ./...
```

通过标准：

- 所有 Go tests 通过。
- `go vet ./...` 返回 0。
- preflight summary 为 PASS。
- release gate summary 为 PASS。

## 3. CLI Smoke

英文输出：

```powershell
go run ./cmd/repomind analyze --output .repomind .
```

中文输出：

```powershell
go run ./cmd/repomind analyze --lang zh --output .repomind .
```

通过标准：

- 生成 `analysis.json`。
- 生成 `report.html`。
- 英文输出可读。
- 中文输出可读。

## 4. 真实仓库 Evaluation

```powershell
.\scripts\evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809
```

通过标准：

- 所有配置仓库 clone 成功。
- 所有配置仓库 analyze 成功。
- `eval/summary.md` 生成。
- parser 问题记录到 `docs/REAL_REPO_EVALUATION.md`。

## 5. 性能 Benchmark

```powershell
.\scripts\benchmark-repos.ps1 -TimeoutSeconds 300 -TargetSeconds 30 -Proxy http://127.0.0.1:10809
```

通过标准：

- 所有仓库 analyze 成功。
- 每个仓库低于 30 秒目标。
- `benchmark/summary.md` 生成。
- 性能回归记录到 `docs/PERFORMANCE_BENCHMARKS.md`。

## 6. 可选真实 AI Provider 测试

仅在本地 `.env` 有有效 key 时运行：

```powershell
.\scripts\smoke-ai-provider.ps1 -Provider grok -Model grok-4.3 -Proxy http://127.0.0.1:10809
```

通过标准：

- 网络 provider 返回有效总结。
- `.env` 仍被 ignore，且没有被 stage。

## 7. Release Artifacts

```powershell
.\scripts\smoke-release-artifact.ps1
.\scripts\build-release.ps1 -Version v0.1.0
.\scripts\verify-release-manifest.ps1 -DistDir dist
```

通过标准：

- 当前平台 binary smoke PASS。
- Windows、macOS、Linux amd64/arm64 archive 都存在。
- archive 包含 binary、`README.md`、`README.zh-CN.md` 和 `.env.example`。
- `dist/manifest.json` 和 `dist/manifest.md` 存在。
- manifest verification PASS。

## 8. 文档

- `README.md` 当前有效。
- `README.zh-CN.md` 当前有效。
- `docs/README.md` 和 `docs/README.zh-CN.md` 列出当前文档。
- `docs/PROJECT_PLAN.md` 有最新 milestone。
- 除 `docs/PROJECT_PLAN.md` 外，公开用户文档都有语言切换和对应英文或简体中文版本。
- release gate、parser backlog、evaluation、benchmark 文档已更新。

## 9. Tag Release

可选手动 GitHub release gate：

```txt
Actions -> Release Gate -> Run workflow
```

全部检查通过后：

```powershell
git tag v0.1.0
git push origin v0.1.0
```
