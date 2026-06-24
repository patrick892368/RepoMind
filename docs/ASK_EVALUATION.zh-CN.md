# Ask Evaluation

**语言：** [English](ASK_EVALUATION.md) | 简体中文

`repomind eval ask` 用可重复的仓库问题验证 `repomind ask`。它会检查预期文件、处理函数、路由、模型、调用链、证据类型和证据数量。`scripts/evaluate-ask.ps1` 保留为兼容入口。

运行跨平台 Go CLI 评估器：

```bash
repomind eval ask --cases docs/examples/ask-cases.example.json --strict
```

从源码运行：

```bash
go run ./cmd/repomind eval ask --cases docs/examples/ask-cases.example.json --strict
```

通过 preflight 运行内置 case：

```powershell
.\scripts\preflight.ps1 -IncludeAskEvaluation -AskProvider offline -AskStrict
```

运行旧 PowerShell 兼容评估器：

```powershell
.\scripts\evaluate-ask.ps1 -Provider offline -Strict
```

运行自定义 case 文件：

```powershell
.\scripts\evaluate-ask.ps1 -Provider offline -Strict -CasesPath docs\examples\ask-cases.example.json
```

通过 preflight 运行自定义 case：

```powershell
.\scripts\preflight.ps1 -IncludeAskEvaluation -AskProvider mock -AskStrict -AskCasesPath docs\examples\ask-cases.example.json
```

`scripts/preflight.ps1` 和 `scripts/release-gate.ps1` 默认通过 Go CLI 运行 ask evaluation。`-Proxy` 会通过 `HTTPS_PROXY`、`HTTP_PROXY` 和 `ALL_PROXY` 转发给 Go CLI 子进程。可以用 `-AskCasesPath` 指定 release gate 自定义 case，或在排查时用 `-SkipAskEvaluation` 跳过：

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -AskCasesPath docs\examples\ask-cases.example.json
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -SkipAskEvaluation
```

输出：

```txt
eval/ask/summary.json
eval/ask/summary.md
```

## Case 文件

Case 文件使用 JSON，可以是顶层 `cases` 数组，也可以直接是顶层数组。

支持字段：

- `name`：稳定的 case 名称。
- `repo_path`：本地仓库或 fixture 路径。
- `language`：可选 analyze 语言，例如 `en` 或 `zh`。
- `question`：ask 问题。
- `expected_files`：预期文件。
- `expected_handlers`：预期处理函数或函数名。
- `expected_routes`：预期路由标签，例如 `POST /login`。
- `expected_models`：预期数据库模型名。
- `expected_call_chain`：预期调用链前缀，例如 `pay_callback -> update_order`。
- `expected_evidence_types`：预期证据类型，例如 `route`、`model` 或 `call_edge`。
- `minimum_evidence`：最少证据数量。

私有或业务相关的问题集可以放在 ignored 的本地路径，例如 `eval/ask-cases.json`。
