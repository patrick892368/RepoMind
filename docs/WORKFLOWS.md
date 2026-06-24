# RepoMind Workflows

This document explains how local checks map to CI and release checks.

## Local Default Preflight

Run before ordinary commits:

```powershell
.\scripts\preflight.ps1
```

This runs:

- `go test ./...`
- `go vet ./...`
- English `repomind analyze` smoke
- Chinese `repomind analyze --lang zh` smoke

Output:

```txt
eval/preflight/summary.json
eval/preflight/summary.md
```

## Optional Local Checks

Run current-platform binary smoke:

```powershell
.\scripts\preflight.ps1 -IncludeReleaseSmoke
```

Run AI Provider smoke with a local `.env` key:

```powershell
.\scripts\preflight.ps1 -IncludeAISmoke -AIProvider grok -AIModel grok-4.3 -Proxy http://127.0.0.1:10809
```

Run real repository benchmark:

```powershell
.\scripts\preflight.ps1 -IncludeBenchmark -Proxy http://127.0.0.1:10809
```

Run real repository evaluation:

```powershell
.\scripts\preflight.ps1 -IncludeEvaluation -Proxy http://127.0.0.1:10809
```

By default, evaluation requires every sampled repository to reach `quality_score >= 1.0`. Adjust only when deliberately investigating a parser regression:

```powershell
.\scripts\preflight.ps1 -IncludeEvaluation -MinimumEvaluationQualityScore 0.8
```

## Local Release Gate

Run before creating a release tag:

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809
```

This runs the default preflight plus:

- current-platform release binary smoke
- cross-platform release manifest build and verification
- real repository benchmark
- real repository evaluation with quality gate
- shared repository cache for benchmark and evaluation

Repository clones are retried by default. Use `-CloneRetries` when the network is unstable:

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -CloneRetries 5
```

Use a persistent repo cache to avoid repeated GitHub clones across runs:

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval/repo-cache
```

To include a real AI provider smoke:

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -IncludeAISmoke -AIProvider grok -AIModel grok-4.3
```

To skip the cross-platform manifest build during investigation:

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -SkipManifestBuild
```

## CI Workflow

`.github/workflows/ci.yml` runs on push to `main` / `master` and pull requests.

It runs:

- `go test ./...`
- `go vet ./...`
- English analyze smoke
- Chinese analyze smoke

The local equivalent is:

```powershell
.\scripts\preflight.ps1
```

## Release Workflow

`.github/workflows/release.yml` runs on version tags matching `v*`.

It runs:

- native binary smoke on Ubuntu, macOS, and Windows
- `go test ./...`
- `go vet ./...`
- cross-platform binary builds
- linux/amd64 built binary smoke
- GitHub Release upload

The local binary smoke equivalent is:

```powershell
.\scripts\smoke-release-artifact.ps1
```

## Manual Release Gate Workflow

`.github/workflows/release-gate.yml` can be triggered manually from GitHub Actions.

It runs the local release gate on `windows-latest`:

- default preflight
- release binary smoke
- release manifest build and verification
- real repository benchmark
- real repository evaluation quality gate

It uploads the release gate summary files, benchmark/evaluation summaries, release smoke summary, and release manifest verification files as a workflow artifact.

## Artifact Hygiene

Generated outputs are ignored by Git and skipped by the scanner:

```txt
.repomind/
dist/
eval/
benchmark/
```

Do not commit `.env`, generated reports, benchmark repositories, or release artifacts.
