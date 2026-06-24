# Remote Repository Analysis

RepoMind can analyze a local path or a remote Git URL.

```powershell
go run ./cmd/repomind analyze .
go run ./cmd/repomind analyze https://github.com/owner/repo.git
```

Remote analysis uses Git locally. RepoMind does not implement its own GitHub API client for cloning.

## Requirements

- `git` must be installed and available on `PATH`.
- Network access must be available for remote URLs.
- For private repositories, Git authentication must already work in your shell.

## Public Repositories

Analyze the default branch:

```powershell
go run ./cmd/repomind analyze https://github.com/spring-guides/gs-rest-service.git
```

Analyze a branch, tag, or reachable commit SHA:

```powershell
go run ./cmd/repomind analyze --ref main https://github.com/spring-guides/gs-rest-service.git
go run ./cmd/repomind analyze --branch main https://github.com/spring-guides/gs-rest-service.git
go run ./cmd/repomind analyze --ref e9efc9dfa0abe8cf8e15cf0e71830b5125322cae https://github.com/spring-guides/gs-rest-service.git
```

`--branch` is an alias for `--ref`. If both are provided, they must match.

## Output Location

For local paths, relative `--output` paths are resolved inside the target repository.

For remote Git URLs, relative `--output` paths are resolved from the current working directory. This keeps the generated report after the temporary clone is cleaned up.

```powershell
go run ./cmd/repomind analyze --output .repomind https://github.com/owner/repo.git
```

## Clone Cache

Use `--repo-cache` when repeatedly analyzing the same remote repositories.

```powershell
go run ./cmd/repomind analyze --repo-cache .repomind/repo-cache https://github.com/owner/repo.git
```

Behavior:

- The cache stores bare Git repositories.
- The cache is updated with `git fetch --prune` before each analysis.
- The temporary analysis checkout is still deleted after each run.
- The cache is opt-in and is not created unless `--repo-cache` is provided.

## Private Repositories

RepoMind relies on Git's normal authentication.

Recommended options:

- Use SSH URLs after configuring your SSH key:

```powershell
go run ./cmd/repomind analyze git@github.com:owner/private-repo.git
```

- Use HTTPS URLs with Git Credential Manager or another Git credential helper:

```powershell
go run ./cmd/repomind analyze https://github.com/owner/private-repo.git
```

Avoid placing access tokens directly in command history. Prefer SSH keys, Git Credential Manager, or your platform's credential helper.

RepoMind does not store private repository credentials. Credentials are handled by `git`.

## Proxy

Git and AI provider calls can use standard proxy environment variables.

```powershell
$env:HTTPS_PROXY="http://127.0.0.1:10809"
$env:HTTP_PROXY="http://127.0.0.1:10809"
go run ./cmd/repomind analyze https://github.com/owner/repo.git
```

For SOCKS proxies:

```powershell
$env:ALL_PROXY="socks5://127.0.0.1:10808"
go run ./cmd/repomind analyze https://github.com/owner/repo.git
```

## Security Boundary

- RepoMind does not upload repository source by default.
- Network AI providers are only called when `--ai openai`, `--ai claude`, `--ai gemini`, or `--ai grok` is explicitly selected.
- AI providers receive the structured analysis summary, not the full repository source by default.
- Keep `.env`, `.repomind/`, `eval/`, `benchmark/`, and `dist/` out of Git.

## Troubleshooting

If clone fails:

```powershell
git ls-remote https://github.com/owner/repo.git HEAD
```

If a private repository fails:

```powershell
git ls-remote git@github.com:owner/private-repo.git HEAD
```

If `--ref` fails, verify the ref exists:

```powershell
git ls-remote https://github.com/owner/repo.git refs/heads/main
git ls-remote https://github.com/owner/repo.git refs/tags/v1.0.0
```

If network access is restricted, set `HTTPS_PROXY`, `HTTP_PROXY`, or `ALL_PROXY` before running RepoMind.
