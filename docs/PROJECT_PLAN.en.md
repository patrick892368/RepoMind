# RepoMind Project Plan

**Language:** English | [简体中文](PROJECT_PLAN.md)

RepoMind helps developers understand an unfamiliar repository within 30 seconds. It is not a code generator and not a general AI agent. It is a repository understanding layer for existing codebases.

## Goal

RepoMind should help with:

- Taking over legacy projects.
- Reading large open-source repositories.
- Preparing high-quality context for AI coding tools.
- Technical interview or source reading workflows.

Primary command:

```bash
repomind analyze .
```

Expected first-run output:

- Stack detection.
- Project structure summary.
- Mermaid architecture diagram.
- Database ER diagram.
- API map.
- Startup flow.
- Key business flow summary.
- `.repomind/analysis.json`.
- `.repomind/report.html`.

## Positioning

```txt
RepoMind
Understand Any Repository in 30 Seconds
```

RepoMind should feel like:

```txt
Cursor for Existing Codebases
```

It should not become:

- Another AI agent.
- Another admin framework.
- A tool that mainly generates application code.

## Scope

In scope:

- Local repository scanning.
- Remote Git URL analysis.
- Stack detection.
- Config and dependency detection.
- Database model extraction.
- API route extraction.
- Mermaid diagrams.
- Static HTML report.
- JSON analysis output.
- AI summary through explicit providers.
- Context export for Codex, Claude Code, and Cursor.

Out of scope for MVP:

- Automatically modifying user source code.
- Generating business code by default.
- Long-running autonomous agents.
- SaaS team collaboration.
- Full semantic understanding of every language and framework.

Hard constraints:

- `repomind analyze .` should return a useful first report within 30 seconds.
- Deterministic scanning and static parsing come before AI.
- Network AI providers must only receive structured summaries by default, not full source.
- Every completed milestone or important implementation change must update `PROJECT_PLAN.md`.

## Architecture Direction

Backend:

- Go.

Reasons:

- Single-binary distribution.
- Cross-platform builds.
- Fast filesystem scanning.
- Good CLI ergonomics.

Report:

- Static HTML for MVP.
- Mermaid for diagrams.
- React/Next.js and ReactFlow may be used later for richer UI.

Storage:

- MVP writes `.repomind/analysis.json`.
- Later phases may add SQLite cache and indexing.

AI providers:

- OpenAI.
- Claude.
- Gemini.
- Grok.
- All providers go through a shared internal provider interface.

Scanning layers:

```txt
L1: File Scanner
L2: Static Parser
L3: AI Understanding
L4: AI Coding Tool Integration
```

## Milestone Summary

Completed through M73:

- M1-M5: CLI foundation, scanning, stack detection, database/API extraction, static report.
- M6-M10: AI summary, AI coding tool export, ask, trace, diagnose.
- M11-M20: PHP/Java/Go support, release hardening, evidence/confidence, monorepo packages, provider hardening.
- M21-M31: Go AST parser improvements, AI smoke, preflight, release smoke, docs and parser backlog.
- M32-M42: evaluation quality gates and cross-file route prefix strategy.
- M43-M52: release workflow hardening, manifest verification, FastAPI/Express resolver improvements.
- M53-M62: Go route factory, Express and FastAPI multi-line route coverage, Chi detection, DRF router basics.
- M63-M69: remote Git URL input, `--ref`, commit SHA, `--repo-cache`, remote docs, Git failure hints.
- M70-M72: Go mounted sub-router variables, middleware-wrapped handler names, full release gate verification.
- M73: root README bilingual switch and release archives include `README.zh-CN.md`.

The canonical detailed milestone log remains in `PROJECT_PLAN.md`.

## Testing Strategy

Unit tests are required for:

- Scanner and ignore rules.
- Stack detector.
- Route parsers.
- Model parsers.
- Mermaid generators.
- AI prompt/provider logic.
- JSON output schema behavior.

Integration tests are required for:

- `repomind analyze .`.
- Fixture repositories.
- `analysis.json` and `report.html` generation.
- AI mock summary flow.
- Export, ask, trace, and diagnose workflows.
- Remote Git URL analysis.

Manual or semi-automated checks are required for:

- Real repository speed.
- Large repository performance.
- Report visual quality.
- Real AI provider calls.
- Windows/macOS/Linux binary behavior.

## Update Rule

Every completed feature, milestone, or important fix must update:

- Relevant milestone status.
- Completed features.
- Next steps.
- Test record.
- Scope boundary if the product boundary changed.

Allowed statuses:

```txt
Not started
In progress
Done
Blocked
Deferred
```

## Current Next Steps

- Add more real repository samples and quality checks.
- Improve Go route factories with arguments, middleware chains, and cross-package route assembly.
- Improve Python route metadata and DRF custom action or cross-file router parsing.
- Add optional remote repository cache cleanup and cache size documentation.
