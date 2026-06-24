# RepoMind

**语言：** [English](README.md) | 简体中文

30 秒理解任何代码仓库。

[![CI](https://github.com/patrick892368/RepoMind/actions/workflows/ci.yml/badge.svg)](https://github.com/patrick892368/RepoMind/actions/workflows/ci.yml)
[![Release](https://github.com/patrick892368/RepoMind/actions/workflows/release.yml/badge.svg)](https://github.com/patrick892368/RepoMind/actions/workflows/release.yml)
[![Release Gate](https://github.com/patrick892368/RepoMind/actions/workflows/release-gate.yml/badge.svg)](https://github.com/patrick892368/RepoMind/actions/workflows/release-gate.yml)

RepoMind 是一个 CLI-first 的代码仓库理解工具。它扫描已有代码库，生成结构化报告，包括技术栈识别、数据库模型、API 路由、Mermaid 图、调用图、AI 总结，以及给 Codex、Claude Code、Cursor 等 AI Coding 工具使用的上下文文件。

RepoMind 不是 AI Agent，也不会默认生成业务代码。它的目标是把陌生仓库转换成开发者和 AI Coding 工具都能消费的结构化上下文。

## 报告预览

![RepoMind report preview](docs/assets/report-preview.png)

## 当前状态

项目仍处于早期开发阶段。核心 CLI 和分析管线已经实现，parser 覆盖会继续通过真实仓库评估逐步增强。

已实现命令：

```bash
repomind analyze .
repomind analyze https://github.com/owner/repo.git
repomind ask . --question "where is order created?"
repomind ask . --question "订单在哪里创建？" --ai grok --ai-model grok-4.3
repomind trace . --symbol pay_callback
repomind diagnose . --issue "order status error"
repomind export codex .
repomind export claude .
repomind export cursor .
```

## 评估快照

最新本地 benchmark 结果均低于 30 秒目标：

| 仓库 | 耗时 |
|---|---:|
| Laravel | 7.41s |
| Spring REST service | 1.94s |
| Gin examples | 1.32s |
| FastAPI full-stack template | 2.99s |
| Prisma examples | 9.68s |

详情：

- `docs/PERFORMANCE_BENCHMARKS.md`
- `docs/REAL_REPO_EVALUATION.md`

## 快速开始

从源码运行：

```bash
go run ./cmd/repomind analyze .
```

分析 GitHub 仓库 URL：

```bash
go run ./cmd/repomind analyze https://github.com/owner/repo.git
```

分析指定 branch、tag 或 commit SHA：

```bash
go run ./cmd/repomind analyze --ref main https://github.com/owner/repo.git
```

重复分析远程仓库时复用本地 Git cache：

```bash
go run ./cmd/repomind analyze --repo-cache .repomind/repo-cache https://github.com/owner/repo.git
```

远程 Git 输入会克隆到临时目录。若 `--output` 使用相对路径，报告会写到当前工作目录下，避免临时目录清理后丢失结果。

远程仓库、私有仓库认证、代理、ref 选择和 clone cache 行为见：

```txt
docs/REMOTE_REPOSITORIES.md
```

命令会生成：

```txt
.repomind/analysis.json
.repomind/report.html
```

打开 `.repomind/report.html` 即可查看报告。

生成中文输出：

```bash
go run ./cmd/repomind analyze --lang zh .
```

支持输出语言：

- `en`：英文
- `zh`：简体中文

安装方式见：

```txt
docs/INSTALL.md
```

## AI 总结 Provider

RepoMind 默认使用离线模式，不会调用网络 AI Provider。

支持的网络 Provider：

```bash
go run ./cmd/repomind analyze --ai openai --ai-model gpt-4.1-mini .
go run ./cmd/repomind analyze --ai claude --ai-model claude-sonnet-4-5 .
go run ./cmd/repomind analyze --ai gemini --ai-model gemini-2.5-flash .
go run ./cmd/repomind analyze --ai grok --ai-model grok-4.3 .
```

创建本地 `.env` 文件：

```env
OPENAI_API_KEY=your_openai_key_here
ANTHROPIC_API_KEY=your_anthropic_key_here
GEMINI_API_KEY=your_gemini_key_here
GROK_API_KEY=your_key_here
HTTPS_PROXY=http://127.0.0.1:10809
# or:
# ALL_PROXY=socks5://127.0.0.1:10808
```

`.env` 已被 Git 忽略，不要提交 API key。

RepoMind 只会把结构化分析摘要发送给显式选择的 AI Provider，默认不会上传完整源码。

## AI 仓库问答

运行 `analyze` 后，可以针对仓库提问：

```bash
go run ./cmd/repomind ask . --question "订单在哪里创建？"
go run ./cmd/repomind ask . --question "订单在哪里创建？" --ai grok --ai-model grok-4.3
go run ./cmd/repomind ask . --question "订单在哪里创建？" --ai grok --strict
```

离线 ask 模式会基于 `.repomind/analysis.json` 排序候选文件、处理函数、模型、路由和调用链。

AI ask 模式只会把结构化分析事实和少量候选源码片段发送给选定 Provider。Provider 返回的文件、处理函数、模型、路由、调用链和证据都会经过本地分析结果校验，校验失败的内容会被丢弃。

Ask 结果会包含 Evidence 区块，展示 RepoMind 能本地验证的文件路径和行号范围。`--strict` 要求必须有本地证据；没有证据时，RepoMind 会返回依据不足，而不是依赖 AI 推断。

结果会写入：

```txt
.repomind/ask/last-answer.json
.repomind/ask/last-answer.md
```

## 支持的识别能力

技术栈识别：

- Python：Django、FastAPI、Celery
- JavaScript / TypeScript：React、Vue、Next.js、Express、NestJS、BullMQ
- PHP：Laravel、Symfony、ThinkPHP
- Java：Spring Boot
- Go：Gin、Chi、Echo、Fiber
- 数据库：Postgres、MySQL、SQLite、MongoDB
- 缓存：Redis
- 包管理器：npm、pnpm、yarn、pip、poetry、composer、Maven、Gradle、Go modules
- Monorepo package grouping：基于 `package.json`、`pyproject.toml`、`requirements.txt`、`go.mod`、`composer.json`、`pom.xml`、`build.gradle`、`schema.prisma`

数据库模型抽取：

- Prisma
- Django Models
- SQLAlchemy
- SQLModel
- TypeORM
- Java JPA
- Go GORM

API 路由抽取：

- Django URL patterns
- FastAPI decorators
- Express routes
- NestJS controllers
- Laravel routes
- Spring controllers
- Go router calls

## AI Coding 工具导出

分析仓库后，可以导出 AI Coding 工具上下文：

```bash
go run ./cmd/repomind export codex .
go run ./cmd/repomind export claude .
go run ./cmd/repomind export cursor .
```

生成文件包括：

```txt
AGENTS.md
CLAUDE.md
.cursor/rules/repomind.md
.repomind/context.md
.repomind/architecture.md
.repomind/api-map.md
.repomind/db-schema.md
```

## 开发

运行测试：

```bash
go test ./...
```

运行 vet：

```bash
go vet ./...
```

运行本地分析：

```bash
go run ./cmd/repomind analyze .
```

运行默认本地 preflight：

```powershell
.\scripts\preflight.ps1
```

运行完整本地 release gate：

```powershell
.\scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809
```

release gate 覆盖测试、CLI smoke、release binary smoke、benchmark、evaluation quality checks 和 release manifest verification。

通过本地 HTTP 代理调用网络 AI Provider：

```bash
$env:HTTPS_PROXY="http://127.0.0.1:10809"
go run ./cmd/repomind analyze --ai openai --ai-model gpt-4.1-mini .
```

同样的代理变量也适用于受限网络下的 GitHub URL 分析。

## 构建

构建当前平台二进制：

```bash
go build -o repomind ./cmd/repomind
```

构建 Windows、macOS、Linux release artifacts：

```powershell
.\scripts\build-release.ps1 -Version v0.1.0
```

构建产物输出到：

```txt
dist/
```

## 发布

GitHub Releases 基于版本 tag 构建：

```bash
git tag v0.1.0
git push origin v0.1.0
```

release workflow 会构建：

- Windows amd64 / arm64
- macOS amd64 / arm64
- Linux amd64 / arm64

每个 archive 包含 `repomind` 二进制、`LICENSE`、`README.md`、`README.zh-CN.md` 和 `.env.example`。

## 许可证

RepoMind 使用 [MIT License](LICENSE) 发布。

## 边界

RepoMind 不会：

- 默认修改应用源码。
- 默认生成业务代码。
- 在未显式启用网络 AI Provider 时上传源码。
- 替代 Codex、Claude Code、Cursor 或其他 AI Coding 工具。

RepoMind 负责提供这些工具可以消费的仓库理解层。
