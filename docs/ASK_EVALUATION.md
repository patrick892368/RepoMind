# Ask Evaluation

**Language:** English | [简体中文](ASK_EVALUATION.zh-CN.md)

`scripts/evaluate-ask.ps1` validates `repomind ask` with repeatable repository questions. It checks expected files, handlers, routes, models, call-chain edges, evidence types, and evidence counts.

Run the built-in cases:

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

`scripts/release-gate.ps1` runs the built-in offline strict ask evaluation by default. Use `-AskCasesPath` for custom release-gate cases or `-SkipAskEvaluation` during investigation:

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
