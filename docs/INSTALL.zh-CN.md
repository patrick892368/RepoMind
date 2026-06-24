# 安装 RepoMind

**语言：** [English](INSTALL.md) | 简体中文

RepoMind 是 CLI-first 工具。安装方式取决于你要从源码运行，还是使用 release binary。

## 从源码运行

开发 RepoMind 时使用：

```powershell
go run ./cmd/repomind analyze .
```

## 本地构建

构建当前平台二进制：

```powershell
go build -o repomind ./cmd/repomind
```

Windows：

```powershell
.\repomind.exe version
.\repomind.exe analyze .
```

macOS 或 Linux：

```bash
./repomind version
./repomind analyze .
```

## 使用 Go 安装

仓库发布后可通过 module path 安装：

```bash
go install github.com/patrick892368/RepoMind/cmd/repomind@latest
```

确保 Go binary 目录在 `PATH` 中。

Windows PowerShell：

```powershell
$env:PATH += ";$(go env GOPATH)\bin"
repomind version
```

macOS 或 Linux：

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
repomind version
```

## 使用 Release Binary

从 GitHub Releases 下载对应平台 archive，解压后把 `repomind` 或 `repomind.exe` 放到 PATH 目录中。

Archive 包含：

```txt
repomind
LICENSE
README.md
README.zh-CN.md
.env.example
```

检查：

```powershell
repomind version
repomind analyze .
```

## AI Provider 配置

默认离线模式不需要 API key。

如果要使用网络 AI Provider，创建本地 `.env`：

```env
OPENAI_API_KEY=
ANTHROPIC_API_KEY=
GEMINI_API_KEY=
GROK_API_KEY=
HTTPS_PROXY=http://127.0.0.1:10809
```

`.env` 已被 Git 忽略。不要提交 API key。
