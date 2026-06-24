# Ask Evaluation

**Language:** English | [简体中文](ASK_EVALUATION.zh-CN.md)

`repomind eval ask` validates `repomind ask` with repeatable repository questions. It checks expected files, handlers, routes, models, call-chain edges, evidence types, and evidence counts. `scripts/evaluate-ask.ps1` remains available as a compatibility wrapper that delegates to the Go CLI evaluator.

Run the cross-platform Go CLI evaluator:

```bash
repomind eval ask --cases docs/examples/ask-cases.example.json --strict
```

From source:

```bash
go run ./cmd/repomind eval ask --cases docs/examples/ask-cases.example.json --strict
```

Run built-in cases through preflight:

```powershell
.\scripts\preflight.ps1 -IncludeAskEvaluation -AskProvider offline -AskStrict
```

Run the PowerShell compatibility wrapper:

```powershell
.\scripts\evaluate-ask.ps1 -Provider offline -Strict
```

Run a custom case file:

```powershell
.\scripts\evaluate-ask.ps1 -Provider offline -Strict -CasesPath docs\examples\ask-cases.example.json
```

Run custom cases through preflight:

```powershell
.\scripts\preflight.ps1 -IncludeAskEvaluation -AskProvider mock -AskStrict -AskCasesPath docs\examples\ask-cases.example.json
```

`scripts/preflight.ps1` and `scripts/release-gate.ps1` run ask evaluation through the Go CLI by default. `-Proxy` is forwarded to the Go CLI process through `HTTPS_PROXY`, `HTTP_PROXY`, and `ALL_PROXY`. Use `-AskCasesPath` for custom release-gate cases or `-SkipAskEvaluation` during investigation:

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -AskCasesPath docs\examples\ask-cases.example.json
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -SkipAskEvaluation
```

Outputs:

```txt
eval/ask/summary.json
eval/ask/summary.md
```

## Case File

Case files are JSON and may contain either a top-level `cases` array or a top-level array.

Supported fields:

- `name`: stable case name.
- `repo_path`: local repository or fixture path.
- `language`: optional analyze language, such as `en` or `zh`.
- `question`: ask question.
- `expected_files`: expected answer files.
- `expected_handlers`: expected handlers or functions.
- `expected_routes`: expected route labels such as `POST /login`.
- `expected_models`: expected database model names.
- `expected_call_chain`: expected call-chain prefixes such as `pay_callback -> update_order`.
- `expected_evidence_types`: expected evidence types such as `route`, `model`, or `call_edge`.
- `minimum_evidence`: minimum evidence item count.

Private or proprietary cases can be kept under an ignored local path such as `eval/ask-cases.json`.
