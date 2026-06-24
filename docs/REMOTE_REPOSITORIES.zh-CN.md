# 远程仓库分析

**语言：** [English](REMOTE_REPOSITORIES.md) | 简体中文

RepoMind 可以分析本地路径，也可以直接分析远程 Git URL。

```powershell
go run ./cmd/repomind analyze .
go run ./cmd/repomind analyze https://github.com/owner/repo.git
```

远程分析依赖本机 Git。RepoMind 不自己实现 GitHub clone 客户端。

## 要求

- `git` 已安装并可在 `PATH` 中访问。
- 当前网络能访问目标 Git remote。
- 私有仓库需要当前 shell 中的 Git 认证已经可用。

## 公共仓库

分析默认分支：

```powershell
go run ./cmd/repomind analyze https://github.com/spring-guides/gs-rest-service.git
```

分析 branch、tag 或可达 commit SHA：

```powershell
go run ./cmd/repomind analyze --ref main https://github.com/spring-guides/gs-rest-service.git
go run ./cmd/repomind analyze --branch main https://github.com/spring-guides/gs-rest-service.git
go run ./cmd/repomind analyze --ref e9efc9dfa0abe8cf8e15cf0e71830b5125322cae https://github.com/spring-guides/gs-rest-service.git
```

`--branch` 是 `--ref` 的别名。两者同时提供时，值必须一致。

## 输出位置

分析本地路径时，相对 `--output` 会相对于目标仓库解析。

分析远程 Git URL 时，相对 `--output` 会相对于当前工作目录解析。这样临时 clone 清理后，报告仍保留。

远程分析会在 `analysis.json` 写入 repository metadata：

```json
{
  "repository": {
    "name": "repo",
    "remote": true,
    "ref": "main"
  }
}
```

未传入 `--ref` / `--branch` 时会省略 `ref`。RepoMind 不会把原始 remote URL 写入 `analysis.json`，避免持久化私有 URL 中可能携带的凭证。

## Clone Cache

反复分析同一批远程仓库时使用：

```powershell
go run ./cmd/repomind analyze --repo-cache .repomind/repo-cache https://github.com/owner/repo.git
```

行为：

- cache 保存 bare Git repository。
- 每次分析前执行 `git fetch --prune` 更新 cache。
- 单次分析使用的临时 checkout 仍会在结束后删除。
- 只有显式传入 `--repo-cache` 时才会创建 cache。

## 私有仓库

RepoMind 使用 Git 的正常认证机制。

推荐方式：

- 配置 SSH key 后使用 SSH URL：

```powershell
go run ./cmd/repomind analyze git@github.com:owner/private-repo.git
```

- 使用 HTTPS URL，并通过 Git Credential Manager 或 credential helper 管理凭据：

```powershell
go run ./cmd/repomind analyze https://github.com/owner/private-repo.git
```

不要把 access token 直接写进命令历史。优先使用 SSH key、Git Credential Manager 或系统 credential helper。

RepoMind 不保存私有仓库凭据，凭据由 `git` 处理。

## 代理

Git 和 AI provider 调用可以使用标准代理环境变量：

```powershell
$env:HTTPS_PROXY="http://127.0.0.1:10809"
$env:HTTP_PROXY="http://127.0.0.1:10809"
go run ./cmd/repomind analyze https://github.com/owner/repo.git
```

SOCKS 代理：

```powershell
$env:ALL_PROXY="socks5://127.0.0.1:10808"
go run ./cmd/repomind analyze https://github.com/owner/repo.git
```

## 安全边界

- RepoMind 默认不上传仓库源码。
- 只有显式选择 `--ai openai`、`--ai claude`、`--ai gemini` 或 `--ai grok` 时才调用网络 AI provider。
- 默认发送给 AI provider 的是结构化分析摘要，不是完整源码。
- `.env`、`.repomind/`、`eval/`、`benchmark/`、`dist/` 不应进入 Git。

## 排查

clone 失败时先运行：

```powershell
git ls-remote https://github.com/owner/repo.git HEAD
```

私有仓库失败时：

```powershell
git ls-remote git@github.com:owner/private-repo.git HEAD
```

`--ref` 失败时确认 ref 存在：

```powershell
git ls-remote https://github.com/owner/repo.git refs/heads/main
git ls-remote https://github.com/owner/repo.git refs/tags/v1.0.0
```
