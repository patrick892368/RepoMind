# Safety Boundary

**Language:** English | [简体中文](SAFETY_BOUNDARY.zh-CN.md)

RepoMind must understand repositories without leaking local secrets or committing generated artifacts.

## Guarded Files

The following paths must stay ignored and untracked:

- `.env`
- `.env.*`, except `.env.example`
- `.repomind/`
- `dist/`
- `eval/`
- `benchmark/`
- local binaries: `repomind`, `repomind.exe`

`.env.example` is intentionally trackable and must contain only empty or placeholder values.

## Remote Repositories

Remote Git URL analysis writes `repository.remote` and optional `repository.ref` to `analysis.json`.

RepoMind does not persist the original remote URL in `analysis.json`, so credentials embedded in private repository URLs are not written into generated reports.

## Verification

Run:

```powershell
.\scripts\verify-safety-boundary.ps1
```

This check verifies:

- sensitive and generated paths are ignored by Git;
- forbidden generated or secret files are not tracked;
- `.env.example` remains trackable;
- tracked files do not contain likely real API key/token assignments.

The default preflight, release gate, and CI run this safety boundary check.
