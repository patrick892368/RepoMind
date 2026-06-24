# RepoMind 工作流

**语言：** [English](WORKFLOWS.md) | 简体中文

本文说明本地检查、CI 和 release 检查的对应关系。

## 本地默认 Preflight

普通提交前运行：

```powershell
.\scripts\preflight.ps1
```

包含：

- `go test ./...`
- `go vet ./...`
- 英文 `repomind analyze` smoke
- 中文 `repomind analyze --lang zh` smoke

输出：

```txt
eval/preflight/summary.json
eval/preflight/summary.md
```

## 可选本地检查

当前平台 binary smoke：

```powershell
.\scripts\preflight.ps1 -IncludeReleaseSmoke
```

使用本地 `.env` key 的 AI Provider smoke：

```powershell
.\scripts\preflight.ps1 -IncludeAISmoke -AIProvider grok -AIModel grok-4.3 -Proxy http://127.0.0.1:10809
```

真实仓库 benchmark：

```powershell
.\scripts\preflight.ps1 -IncludeBenchmark -Proxy http://127.0.0.1:10809
```

真实仓库 evaluation：

```powershell
.\scripts\preflight.ps1 -IncludeEvaluation -Proxy http://127.0.0.1:10809
```

固定问题集 ask evaluation：

```bash
go run ./cmd/repomind eval ask --cases docs/examples/ask-cases.example.json --strict
```

preflight 内部使用 Go CLI 评估器：

```powershell
.\scripts\preflight.ps1 -IncludeAskEvaluation -AskProvider offline -AskStrict
.\scripts\preflight.ps1 -IncludeAskEvaluation -AskProvider mock -AskStrict -AskCasesPath docs\examples\ask-cases.example.json
```

ask evaluation 会用英文和中文问题检查预期文件、处理函数、路由、模型、调用链、证据类型和证据数量，并输出：

```txt
eval/preflight/ask-evaluation/summary.json
eval/preflight/ask-evaluation/summary.md
```

默认要求所有样本 `quality_score >= 1.0`。

## 本地 Release Gate

创建 release tag 前运行：

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809
```

它会运行默认 preflight，并额外运行：

- safety boundary verification
- 默认 preflight 中的 trace 和 diagnose CLI smoke
- 通过 Go CLI 运行 offline strict ask evaluation
- 远程 Git URL analyze smoke
- 当前平台 release binary smoke
- 跨平台 release manifest build and verification
- 真实仓库 benchmark
- 真实仓库 evaluation quality gate
- benchmark、evaluation 和 remote analyze smoke 共享 repository cache

网络不稳定时可以增加 clone 重试次数：

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -CloneRetries 5
```

使用自定义 ask evaluation case：

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -AskCasesPath docs\examples\ask-cases.example.json
```

排查时跳过 ask evaluation：

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -SkipAskEvaluation
```

## CI

GitHub Actions CI 运行：

- `go test ./...`
- `go vet ./...`
- safety boundary verification
- 英文 analyze smoke，并检查核心报告内容
- 中文 analyze smoke，并检查核心报告内容
- trace 和 diagnose smoke

## Release Workflow

tag `v*` 会触发 release workflow，执行：

- Windows、macOS、Linux 原生 smoke
- `go test ./...`
- `go vet ./...`
- safety boundary verification
- 跨平台 release binary build
- release tag 注入到 `repomind version`
- linux/amd64 release binary smoke，并检查版本、英文报告和中文报告核心内容
- archive upload
- release manifest 生成
- GitHub Release 发布
