# Release Checklist

**Language:** English | [简体中文](RELEASE_CHECKLIST.zh-CN.md)

RepoMind 发布前必须完成本清单。

## 1. Working Tree Safety

- [ ] Confirm `.env` is ignored:

```powershell
git check-ignore -v .env
```

- [ ] Confirm generated outputs are ignored:

```powershell
git check-ignore -v .repomind dist eval benchmark
```

- [ ] Confirm no API keys are present in tracked files:

```powershell
git status --short
```

Do not commit local `.env`, `eval/`, `benchmark/`, `.repomind/`, or `dist/` artifacts.

## 2. Unit and Integration Tests

Run the default preflight:

```powershell
.\scripts\preflight.ps1
```

Optional full local preflight with binary smoke:

```powershell
.\scripts\preflight.ps1 -IncludeReleaseSmoke
```

Full local release gate:

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809
```

When GitHub clone is unstable, reuse a cache:

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval/repo-cache -CloneRetries 5
```

Or run the core checks directly:

Run:

```powershell
go test ./...
go vet ./...
```

Pass criteria:

- [ ] All Go tests pass.
- [ ] `go vet ./...` exits 0.
- [ ] `eval/preflight/summary.md` reports PASS when using the preflight script.
- [ ] `eval/release-gate/summary.md` reports PASS when using the release gate script.
- [ ] Release gate includes ask evaluation unless `-SkipAskEvaluation` is intentionally used for investigation.
- [ ] Release gate includes manifest build unless `-SkipManifestBuild` is intentionally used for investigation.

## 3. Local CLI Smoke Tests

Run English output:

```powershell
go run ./cmd/repomind analyze --output .repomind .
```

Run Chinese output:

```powershell
go run ./cmd/repomind analyze --lang zh --output .repomind .
```

Pass criteria:

- [ ] `analysis.json` is generated.
- [ ] `report.html` is generated.
- [ ] English output is readable.
- [ ] Chinese output is readable.
- [ ] `scan.truncated` is acceptable for the tested repository size.

Ask evaluation:

```powershell
go run ./cmd/repomind eval ask --cases docs/examples/ask-cases.example.json --strict
.\scripts\evaluate-ask.ps1 -Provider offline -Strict
.\scripts\preflight.ps1 -IncludeAskEvaluation -AskProvider mock -AskStrict
.\scripts\evaluate-ask.ps1 -Provider offline -Strict -CasesPath docs\examples\ask-cases.example.json
```

Pass criteria:

- [ ] Expected files, handlers, routes, models, and call-chain edges match the fixed questions.
- [ ] English and Chinese ask cases both pass.
- [ ] Expected evidence types are present.
- [ ] Custom case files load successfully when used.
- [ ] Go CLI `eval ask` path passes for custom case files.
- [ ] Preflight ask evaluation uses the Go CLI runner.
- [ ] PowerShell `evaluate-ask.ps1` compatibility wrapper still passes for at least one smoke.
- [ ] Strict mode returns local evidence for every ask case.
- [ ] `summary.json` and `summary.md` are generated.

## 4. Real Repository Evaluation

Run:

```powershell
.\scripts\evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809
```

Pass criteria:

- [ ] All configured repositories clone successfully.
- [ ] All configured repositories analyze successfully.
- [ ] Results are summarized in `eval/summary.md`.
- [ ] Any parser false positive or false negative is recorded in `docs/REAL_REPO_EVALUATION.md`.

## 5. Performance Benchmark

Run:

```powershell
.\scripts\benchmark-repos.ps1 -TimeoutSeconds 300 -TargetSeconds 30 -Proxy http://127.0.0.1:10809
```

Pass criteria:

- [ ] All configured repositories analyze successfully.
- [ ] Every repository is under the 30-second target.
- [ ] Results are summarized in `benchmark/summary.md`.
- [ ] Any benchmark regression is recorded in `docs/PERFORMANCE_BENCHMARKS.md`.

## 6. Optional Real AI Provider Test

Only run when a local `.env` has a valid key.

Preferred smoke script:

```powershell
.\scripts\smoke-ai-provider.ps1 -Provider grok -Model grok-4.3 -Proxy http://127.0.0.1:10809
```

Direct Grok command:

```powershell
go run ./cmd/repomind analyze --ai grok --ai-model grok-4.3 --output .repomind .
```

Pass criteria:

- [ ] Network provider returns a valid summary.
- [ ] Provider output is written to `analysis.json.summary`.
- [ ] Smoke result is written to `eval/ai-smoke-*/ai-smoke-summary.json` when using the script.
- [ ] `.env` is still ignored and not staged.

## 7. Release Artifacts

Run current-platform binary smoke:

```powershell
.\scripts\smoke-release-artifact.ps1
```

Run:

```powershell
.\scripts\build-release.ps1 -Version v0.1.0
.\scripts\verify-release-manifest.ps1 -DistDir dist
```

Pass criteria:

- [ ] Current-platform binary smoke reports PASS in `dist/release-smoke/summary.md`.
- [ ] Windows amd64 artifact exists.
- [ ] Windows arm64 artifact exists.
- [ ] macOS amd64 artifact exists.
- [ ] macOS arm64 artifact exists.
- [ ] Linux amd64 artifact exists.
- [ ] Linux arm64 artifact exists.
- [ ] Archives include binary, `LICENSE`, `README.md`, `README.zh-CN.md`, and `.env.example`.
- [ ] `dist/manifest.json` exists and includes SHA256 values.
- [ ] `dist/manifest.md` exists for release notes review.
- [ ] `dist/manifest-verify.md` reports PASS.

## 8. Documentation

- [ ] `README.md` is current.
- [ ] `README.zh-CN.md` is current and linked from `README.md`.
- [ ] `docs/README.md` and `docs/README.zh-CN.md` list current bilingual documentation.
- [ ] `docs/PROJECT_PLAN.md` has the latest completed milestone.
- [ ] Public user-facing docs, excluding `docs/PROJECT_PLAN.md`, have language switches and matching English or Simplified Chinese counterparts.
- [ ] `docs/REMOTE_REPOSITORIES.md` matches current remote URL, proxy, ref, cache, and private repository behavior.
- [ ] `docs/WORKFLOWS.md` matches the current local, CI, and release checks.
- [ ] `docs/RELEASE_GATE_RESULTS.md` records the latest release gate run.
- [ ] `docs/PARSER_BACKLOG.md` reflects newly discovered parser issues.
- [ ] `docs/REAL_REPO_EVALUATION.md` has the latest evaluation notes.
- [ ] `docs/PERFORMANCE_BENCHMARKS.md` has the latest benchmark notes.
- [ ] Report preview image still matches current report layout.

## 9. Tag Release

Optional manual GitHub release gate:

```txt
Actions -> Release Gate -> Run workflow
```

Pass criteria:

- [ ] The workflow run passes.
- [ ] The uploaded `release-gate-summary` artifact includes ask evaluation summary, `manifest.json`, `manifest.md`, `manifest-verify.json`, and `manifest-verify.md`.

After all checks pass:

```powershell
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will build release artifacts from the tag.
