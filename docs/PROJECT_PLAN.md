# RepoMind Project Plan

RepoMind 的目标是让开发者在 30 秒内理解一个陌生仓库。它不是代码生成工具，也不是通用 AI Agent，而是面向既有代码库的 Repository Understanding 工具。

## 1. Project Goal

RepoMind 要解决的问题：

- 接手老项目时，快速知道项目是什么、怎么启动、核心模块在哪里。
- 阅读大型开源项目时，快速得到架构图、模块关系、数据库结构和 API 地图。
- 使用 AI Coding 工具前，先建立对现有代码库的结构化理解。
- 技术面试或源码阅读时，快速定位业务流程、入口文件和关键调用链。

第一版核心命令：

```bash
repomind analyze .
```

目标输出：

- 技术栈识别
- 项目结构总结
- Mermaid 架构图
- 数据库 ER 图
- API 地图
- 启动流程
- 核心业务流程摘要
- `analysis.json`
- `report.html`

## 2. Product Positioning

项目定位：

```txt
RepoMind
Understand Any Repository in 30 Seconds
```

核心理念：

- 不以生成代码为主。
- 不做另一个聊天 Agent。
- 不做 Admin 框架。
- 优先做“看懂项目”的确定性分析和结构化报告。
- AI 只消费压缩后的结构化上下文，不直接粗暴读取整个仓库。

RepoMind 应该更像：

```txt
Cursor for Existing Codebases
```

而不是：

```txt
Another AI Coding Agent
```

## 3. Scope Boundary

### In Scope

MVP 必须包含：

- 本地仓库扫描。
- 技术栈识别。
- 配置文件识别。
- 入口文件识别。
- 后端框架识别。
- 前端框架识别。
- 数据库模型识别。
- API 路由识别。
- Mermaid 图生成。
- HTML 报告生成。
- JSON 分析结果输出。
- 可导出的 AI Coding 工具上下文文件设计。

### Out of Scope for MVP

MVP 不做：

- 自动修改用户代码。
- 自动生成业务代码。
- 长期运行的 AI Agent。
- IDE 插件。
- 在线 SaaS 平台。
- 权限系统。
- 团队协作功能。
- 复杂增量索引。
- 全语言全框架覆盖。
- 对大型 monorepo 的完整语义理解。

### Hard Constraints

- `repomind analyze .` 必须优先保证 30 秒内可返回初版报告。
- 扫描逻辑必须先走确定性规则，再走 AI 总结。
- AI 输入必须是结构化摘要，不能默认把整个源码仓库塞给模型。
- 模板和规则只能作为第一层扫描，不能成为唯一理解方式。
- AST、静态调用关系和 AI 总结必须围绕统一 IR 逐步增强。
- Codex、Claude Code、Cursor 等工具应该消费 RepoMind 生成的上下文，而不是替代 RepoMind 的扫描器。
- 所有分析结果必须可以落盘，方便后续 `ask`、`trace`、`diagnose` 复用。
- 项目必须保持 CLI-first。
- 核心扫描器必须可跨平台运行。
- 默认不上传源码内容，除非用户明确启用 AI Provider。
- 每完成一个阶段或关键功能后，必须更新本文档的状态、已完成功能和测试记录。

## 4. Architecture Direction

后端语言：

- Go

原因：

- 单文件分发体验好。
- 跨平台构建简单。
- 文件扫描速度快。
- 适合 CLI 工具。

前端报告：

- React
- Next.js 可作为后续交互式 UI 方案
- MVP 可以先输出静态 HTML

图形：

- Mermaid
- ReactFlow 作为后续交互式图谱方案

存储：

- MVP：文件系统输出 `.repomind/analysis.json`
- 后续：SQLite 缓存、索引和问答上下文

AI Provider：

- OpenAI
- Claude
- Gemini

内部必须通过统一 Provider 接口调用。

扫描策略：

```txt
L1: File Scanner
扫描目录、配置、依赖、入口文件、路由文件、模型文件。

L2: Static Parser
用 AST、tree-sitter 或轻量 parser 识别 route、model、service、function、call edge。

L3: AI Understanding
基于 L1/L2 的结构化结果总结项目、推断业务模块、识别核心流程。

L4: AI Coding Tool Integration
导出 Codex、Claude Code、Cursor 可以直接消费的上下文文件。
```

AI Coding 工具集成方向：

```txt
repomind analyze .
repomind export codex
repomind export claude
repomind export cursor
```

目标生成：

```txt
AGENTS.md
CLAUDE.md
.cursor/rules/repomind.md
.repomind/context.md
.repomind/architecture.md
.repomind/api-map.md
.repomind/db-schema.md
.repomind/callgraph.md
```

定位：

- RepoMind 不替代 Codex、Claude Code、Cursor。
- RepoMind 负责把仓库转成结构化理解结果。
- AI Coding 工具基于这些结果继续开发、问答、修改和测试。

## 5. Internal Data Model

RepoMind 内部不要绑定具体框架。所有语言和框架都应该被转换成统一 IR。

核心 IR：

```txt
Repository
Stack
Module
Entrypoint
Config
Dependency
Route
Model
Field
Relation
Service
CallEdge
BusinessFlow
Diagram
Summary
```

推荐 Go 接口：

```go
type Extractor interface {
    Name() string
    Detect(repo RepoFS) DetectionResult
    Extract(repo RepoFS) (*ir.PartialAnalysis, error)
}
```

推荐目录结构：

```txt
cmd/repomind/main.go

internal/analyzer
internal/scanner
internal/detector
internal/ir
internal/extractor/python
internal/extractor/typescript
internal/extractor/php
internal/extractor/java
internal/extractor/golang
internal/graph
internal/ai
internal/report
internal/storage
```

## 6. Milestones

### M1: Project Foundation

Goal:

- 建立 Go CLI 项目骨架。
- 实现 `repomind analyze .` 的最小可运行流程。
- 输出 `.repomind/analysis.json`。

Features:

- CLI 参数解析。
- 仓库路径解析。
- 文件树扫描。
- 忽略规则。
- 基础项目元信息输出。

Unit Tests:

- 路径解析。
- ignore 规则。
- 文件分类。
- JSON 输出结构。

Integration Tests:

- 对一个 fixture 仓库运行 `repomind analyze .`。
- 验证 `.repomind/analysis.json` 被生成。

Status:

- Done

### M2: Stack Detection

Goal:

- 识别主流技术栈和基础设施。

Features:

- Python: Django, FastAPI
- JS/TS: React, Vue, Next.js, Express, NestJS
- Database: Postgres, MySQL, SQLite, MongoDB
- Cache: Redis
- Queue: Celery, BullMQ
- Package managers: npm, pnpm, yarn, pip, poetry
- Config files: `.env.example`, `docker-compose.yml`, `settings.py`, `package.json`

Unit Tests:

- `package.json` 技术栈识别。
- `requirements.txt` 技术栈识别。
- `pyproject.toml` 技术栈识别。
- `docker-compose.yml` 服务识别。
- 多技术栈合并规则。

Integration Tests:

- Django + Vue fixture。
- FastAPI + React fixture。
- Express + Prisma fixture。

Status:

- Done

### M3: Database Analysis

Goal:

- 识别数据库模型并生成 ER 图数据。

Features:

- Django Model 识别。
- SQLAlchemy Model 识别。
- Prisma schema 识别。
- TypeORM entity 识别。
- 模型字段识别。
- 模型关系识别。
- Mermaid ER 图生成。

Unit Tests:

- Django model parser。
- SQLAlchemy parser。
- Prisma parser。
- TypeORM parser。
- 字段类型映射。
- relation 识别。

Integration Tests:

- Django 模型仓库生成 ER 图。
- Prisma 仓库生成 ER 图。
- TypeORM 仓库生成 ER 图。

Status:

- Done

### M4: API Route Analysis

Goal:

- 识别项目 API 路由并生成 API 地图。

Features:

- Django URL 识别。
- FastAPI route 识别。
- Express route 识别。
- NestJS controller 识别。
- HTTP method 识别。
- path 识别。
- handler 文件定位。
- API Mermaid 图生成。

Unit Tests:

- Django URL parser。
- FastAPI decorator parser。
- Express router parser。
- NestJS controller parser。
- method/path/handler 抽取。

Integration Tests:

- Django fixture API 地图。
- FastAPI fixture API 地图。
- Express fixture API 地图。
- NestJS fixture API 地图。

Status:

- Done

### M5: Report Generation

Goal:

- 生成开发者第一次运行时可直接打开的静态报告。

Features:

- `.repomind/report.html`。
- 项目概览。
- 技术栈卡片。
- 目录结构。
- 启动命令。
- 架构图。
- ER 图。
- API 地图。
- 关键模块列表。
- AI 总结占位或可选 AI 总结。

Unit Tests:

- Mermaid 文本生成。
- HTML 模板渲染。
- 空数据降级显示。
- 报告路径生成。

Integration Tests:

- 完整运行 `repomind analyze .` 后生成 report。
- 浏览器打开 report 时 Mermaid 可渲染。

Status:

- Done

### M6: AI Summary

Goal:

- 基于结构化分析结果生成项目总结和业务流程摘要。

Features:

- AI Provider 抽象。
- OpenAI Provider。
- Claude Provider。
- Gemini Provider。
- 离线模式。
- AI 总结 prompt。
- 业务模块推断。
- 启动流程总结。
- 核心流程总结。

Unit Tests:

- Provider config 解析。
- prompt builder。
- response parser。
- token 裁剪策略。
- AI disabled fallback。

Integration Tests:

- 使用 mock provider 生成稳定总结。
- 使用真实 provider 的可选手动测试。

Status:

- Done

### M7: AI Coding Tool Integration

Goal:

- 让 Codex、Claude Code、Cursor 等工具可以直接消费 RepoMind 生成的仓库理解上下文。

Commands:

```bash
repomind export codex
repomind export claude
repomind export cursor
```

Features:

- 生成 `AGENTS.md`。
- 生成 `CLAUDE.md`。
- 生成 `.cursor/rules/repomind.md`。
- 生成 `.repomind/context.md`。
- 生成 `.repomind/architecture.md`。
- 生成 `.repomind/api-map.md`。
- 生成 `.repomind/db-schema.md`。
- 为不同 AI Coding 工具裁剪上下文。

Unit Tests:

- Codex context renderer。
- Claude context renderer。
- Cursor rules renderer。
- 上下文裁剪规则。
- 缺失分析文件时的错误提示。

Integration Tests:

- 从 `.repomind/analysis.json` 生成 Codex 上下文。
- 从 `.repomind/analysis.json` 生成 Claude 上下文。
- 从 `.repomind/analysis.json` 生成 Cursor rules。

Status:

- Done

### M8: Ask Mode

Goal:

- 支持基于分析结果的源码问答。

Command:

```bash
repomind ask .
```

Features:

- 读取 `.repomind/analysis.json`。
- 基于问题定位候选文件。
- 返回关键文件、关键函数、相关模型。
- 支持问题：
  - 订单是怎么派单的？
  - 用户余额在哪里扣减？
  - 风控逻辑在哪？

Unit Tests:

- 问题分类。
- 候选文件排序。
- 上下文构造。
- 输出结构解析。

Integration Tests:

- fixture 项目中定位订单派单逻辑。
- fixture 项目中定位余额扣减逻辑。

Status:

- Done

### M9: Call Chain Analysis

Goal:

- 分析关键业务流程和调用链。

Command:

```bash
repomind trace .
```

Features:

- 函数调用边识别。
- handler 到 service 到 model 的链路识别。
- 支持流程图输出。
- 示例：
  - `pay_callback`
  - `update_order`
  - `update_balance`
  - `send_notify`
  - `write_log`

Unit Tests:

- 函数调用抽取。
- call edge 合并。
- 循环调用处理。
- 最大深度限制。

Integration Tests:

- 支付回调 fixture 调用链。
- 订单创建 fixture 调用链。

Status:

- Done

### M10: Diagnose Mode

Goal:

- 基于代码搜索和调用链生成问题诊断报告。

Command:

```bash
repomind diagnose .
```

Features:

- 状态修改点搜索。
- 数据库写入点搜索。
- 缓存更新点搜索。
- 队列任务搜索。
- 生成诊断报告。

Unit Tests:

- 状态字段识别。
- 写入点识别。
- 缓存操作识别。
- 诊断模板生成。

Integration Tests:

- 订单状态异常 fixture。
- 余额异常 fixture。

Status:

- Done

### M11: Language Expansion

Goal:

- 扩展 PHP、Java、Go 生态支持。

Features:

- PHP:
  - Laravel
  - Symfony
  - ThinkPHP
  - Eloquent Model
  - routes/api.php
  - routes/web.php
- Java:
  - Spring Boot
  - Spring MVC
  - JPA / Hibernate
  - MyBatis
  - `@RestController`
  - `@RequestMapping`
  - `@Entity`
- Go:
  - Gin
  - Echo
  - Fiber
  - GORM
  - sqlc
  - `cmd/*/main.go`

Unit Tests:

- Laravel route/model parser。
- Spring annotation parser。
- MyBatis XML parser。
- Gin route parser。
- GORM model parser。

Integration Tests:

- Laravel fixture。
- Spring Boot fixture。
- Gin + GORM fixture。

Status:

- Done

### M12: Release Hardening and Real Repository Evaluation

Goal:

- 让项目具备开源发布前的基本工程形态，并用真实开源仓库验证扫描质量和速度。

Features:

- README quick start and command reference.
- `.env.example` for AI provider and proxy variable names.
- `.gitignore` protection for `.env`, `.repomind/`, `dist/`, and `eval/`.
- GitHub Actions CI for `go test ./...` and `go vet ./...`.
- Local release artifact builder for Windows, macOS, and Linux.
- Tag-based GitHub release workflow.
- Real repository evaluation script.
- Evaluation script proxy support for restricted networks.
- Real repository evaluation result document.

Unit Tests:

- Grok provider request/response parsing.
- Grok Chat Completions fallback parsing.
- Existing parser and analyzer regression suite.

Integration Tests:

- Local release artifact build.
- Full test suite and `go vet`.
- Real repository evaluation against Laravel, Spring Boot, Gin, FastAPI full-stack template, and Prisma examples.

Status:

- Done

### M13: Evidence, Confidence, and Parser Quality Hardening

Goal:

- 让扫描结果不仅给出结论，还能说明识别来源，降低 AI 总结和人工阅读时的误判成本。

Features:

- DB model IR 增加 `line`、`confidence`、`evidence`。
- API route IR 增加 `line`、`confidence`、`evidence`。
- HTML 报告展示模型和路由的 location、confidence、evidence。
- Codex、Claude Code、Cursor 导出上下文包含模型和路由证据。
- Grok prompt 上下文包含模型和路由位置及可信度。
- Python DB parser 收紧 SQLAlchemy 基类识别，避免把 Pydantic `BaseModel`、`BaseSettings`、普通 `*Base` schema 类误识别为 DB model。
- SQLModel `table=True` 模型识别。
- SQLModel 字段和 Relationship 轻量解析。

Unit Tests:

- API route parser 必须填充 line、confidence、evidence。
- DB model parser 必须填充 line、confidence、evidence。
- SQLModel fixture 覆盖 `table=True` 正例。
- Pydantic `BaseModel` / `BaseSettings` / 非 table schema 类误报回归测试。

Integration Tests:

- 全量 `go test ./...`。
- 真实 FastAPI full-stack template 评估，确认 SQLModel 误报从 10 个模型降到 2 个真实表模型。

Status:

- Done

### M14: Monorepo Package Grouping

Goal:

- 对 monorepo 或多 package 仓库输出局部分组，避免只给出一个过宽的全仓技术栈。

Features:

- 新增 `packages` IR。
- 识别 package roots:
  - `package.json`
  - `pyproject.toml`
  - `requirements.txt`
  - `go.mod`
  - `composer.json`
  - `pom.xml`
  - `build.gradle`
  - `build.gradle.kts`
  - `schema.prisma`
- 每个 package 输出 name、path、type、局部 stack、files、directories、models、routes。
- HTML 报告新增 Packages 表。
- Codex、Claude Code、Cursor 导出上下文新增 Packages 概览。

Unit Tests:

- workspace package root detection。
- package name extraction。
- per-package stack aggregation。
- per-package route count aggregation。

Integration Tests:

- monorepo fixture 完整 analyze。
- 真实 Prisma examples monorepo 评估。
- 真实 FastAPI full-stack template root/backend/frontend 分组评估。

Status:

- Done

### M15: README Preview and Release Presentation

Goal:

- 让开源仓库首页能直观看到 RepoMind 生成报告的形态，降低首次访问理解成本。

Features:

- 生成报告预览图。
- README 新增 Report Preview。
- README 更新 monorepo package grouping 支持说明。
- README 更新 SQLModel 支持说明。

Unit Tests:

- Not required for static README copy and image asset.

Integration Tests:

- Browser verification confirmed generated report renders through local HTTP.
- Screenshot asset manually inspected.

Status:

- Done

### M16: Bilingual Output

Goal:

- 分析结果支持中文和英文，满足中文开发者和国际开源用户两类场景。

Features:

- `repomind analyze` 新增 `--lang en|zh`。
- `analysis.json` 新增 `language` 字段。
- Offline summary 支持英文和简体中文。
- Grok prompt 会按 `language` 要求输出英文或简体中文。
- CLI analyze 输出标签支持英文和简体中文。
- HTML report 固定标签和空状态文案支持英文和简体中文。
- `ask` 读取 `analysis.json` 的语言，并输出对应语言的摘要和区块标题。
- `trace` 读取 `analysis.json` 的语言，并输出对应语言的标题和空状态。
- `diagnose` 读取 `analysis.json` 的语言，并输出对应语言的诊断摘要。

Unit Tests:

- Offline Chinese summary。
- Chinese HTML report labels。
- CLI `analyze --lang zh`。
- `ask` Chinese summary and labels。

Integration Tests:

- 手动运行 `go run ./cmd/repomind analyze --lang zh --output <eval-dir> testdata/fixtures/monorepo`。
- 验证 CLI、`analysis.json.language`、summary 和 report 标签为中文。

Status:

- Done

### M17: Performance Benchmark and Large Repository Guards

Goal:

- 用真实仓库验证 30 秒产品承诺，并为大仓库、生成代码和超大调用图提供默认保护。

Features:

- `repomind analyze` 新增性能限制参数：
  - `--max-files`
  - `--max-file-bytes`
  - `--max-call-edges`
- Analyzer 默认最多扫描 50000 个文件。
- Analyzer 默认跳过超过 512KB 的源码 parser 输入。
- Analyzer 默认最多保留 5000 条调用边。
- `analysis.json.scan.truncated` 标记截断状态。
- CLI 输出截断状态。
- HTML report 显示截断提示。
- 新增 `scripts/benchmark-repos.ps1`。
- 新增 `docs/PERFORMANCE_BENCHMARKS.md`。
- `.gitignore` 忽略 `benchmark/`。

Unit Tests:

- `MaxFiles` 触发截断。
- `MaxCallEdges` 触发截断。
- CLI / analyzer / report 回归测试。

Integration Tests:

- 真实仓库 benchmark，目标 30 秒。

Status:

- Done

### M18: Release Checklist

Goal:

- 固化开源发布前必须执行的检查，避免漏测、漏 benchmark、误提交密钥或生成物。

Features:

- 新增 `docs/RELEASE_CHECKLIST.md`。
- 覆盖工作区安全检查。
- 覆盖 `.env` 和生成目录忽略检查。
- 覆盖 `go test ./...` 和 `go vet ./...`。
- 覆盖英文/中文 CLI smoke test。
- 覆盖真实仓库评估。
- 覆盖性能 benchmark。
- 覆盖可选真实 AI Provider 测试。
- 覆盖本地跨平台 release artifact 构建。
- 覆盖文档更新检查。
- README 增加 release checklist 链接。

Unit Tests:

- Not required for documentation-only milestone.

Integration Tests:

- Release checklist references existing executable test, evaluation, benchmark, and build commands.

Status:

- Done

### M19: Package Hierarchy and Dependency Graph

Goal:

- 让 monorepo 不只是显示 package 列表，还能展示 package 父子层级和本地依赖关系。

Features:

- `PackageInfo` 新增 `parent`。
- `PackageInfo` 新增 `dependencies`。
- `DiagramSet` 新增 `package` Mermaid 图。
- Workspace detector 计算最近父 package。
- Workspace detector 从 manifest 中抽取本地依赖名：
  - `package.json`
  - `pyproject.toml`
  - `requirements.txt`
  - `go.mod`
  - `composer.json`
  - `pom.xml`
- HTML report 的 Packages 表显示 parent 和 dependencies。
- HTML report 新增 Package Graph。
- AI Coding tool export 新增 package parent/dependencies 和 Mermaid package graph。

Unit Tests:

- workspace parent detection。
- workspace local dependency detection。
- Mermaid package graph generation。
- Analyzer package diagram integration。

Integration Tests:

- monorepo fixture 完整 analyze。

Status:

- Done

### M20: Provider Hardening

Goal:

- 把 AI 总结从 Grok 单 Provider 扩展到真实 OpenAI、Claude/Anthropic、Gemini/Google Provider，并保持默认离线和源码不上传边界。

Features:

- `NewProvider` 支持真实 `openai` Provider。
- `NewProvider` 支持真实 `claude` / `anthropic` Provider。
- `NewProvider` 支持真实 `gemini` / `google` Provider。
- OpenAI Provider 调用 Responses API。
- Claude Provider 调用 Messages API。
- Gemini Provider 调用 `generateContent` API。
- 网络 Provider 共享代理、超时、响应大小上限和 HTTP 错误处理。
- Provider API key 支持从系统环境变量或仓库本地 `.env` 读取。
- `.env.example` 增加 OpenAI、Anthropic/Claude、Gemini/Google 变量名。
- README 的 AI Provider 说明从 Grok-only 更新为多 Provider。
- 继续保持默认 offline 模式。
- 继续保持只有显式启用网络 Provider 时才发送结构化分析摘要。

Unit Tests:

- OpenAI Responses API 请求路径、鉴权头、请求体和响应解析。
- Claude Messages API 请求路径、鉴权头、版本头、请求体和响应解析。
- Gemini `generateContent` 请求路径、API key、请求体和响应解析。
- `.env` 中 OpenAI、Claude/Anthropic、Gemini/Google key 的 Provider 构造。
- Grok/xAI 既有 Responses API 和 Chat Completions fallback 回归。

Integration Tests:

- `go test ./internal/ai`。
- `go test ./...`。
- `go vet ./...`。
- 真实网络 Provider 调用仍作为手动验收项，不进入默认 CI。

Status:

- Done

### M21: Go Parser AST Hardening

Goal:

- 用 Go 标准库 AST 强化 Go 生态解析，先获得比正则更稳的 Go route、GORM model 和 callgraph 结果，同时避免引入影响跨平台构建的第三方 parser 依赖。

Features:

- Go API route parser 改为优先使用 `go/parser` AST。
- Go API route parser 识别 selector handler，例如 `controller.Create`。
- Go API route parser 识别 inline handler。
- Go API route parser 保留正则 fallback，用于不完整或不可解析的 Go 文件。
- Go GORM model parser 改为优先使用 `go/parser` AST。
- Go GORM model parser 识别嵌入 `gorm.Model`。
- Go GORM model parser 识别 selector 类型关系，例如 `profiles.Profile`。
- Go GORM model parser 继续避免把普通 DTO 误报为模型。
- Callgraph extractor 增加 `.go` 支持。
- Go callgraph parser 识别函数内的直接调用和 selector 调用。
- Go callgraph parser 忽略路由注册、内置函数和常见格式化/log 调用。

Unit Tests:

- Go route AST parser 覆盖 selector handler、relative path、`Any` 和 inline handler。
- Go GORM AST parser 覆盖嵌入 `gorm.Model`、唯一字段、切片关系和 selector 关系。
- Go GORM AST parser 覆盖普通 DTO 误报回归。
- Go callgraph AST parser 覆盖函数调用和 selector 调用。
- Callgraph extractor 覆盖临时 Go fixture。

Integration Tests:

- `go test ./internal/parser/apiroute ./internal/parser/dbmodel ./internal/parser/callgraph ./internal/analyzer`。
- `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/multilang-repo`。
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M22: AI Provider Smoke Script and Generated Artifact Hygiene

Goal:

- 让真实 AI Provider 调用有可重复的本地 smoke test，同时避免 RepoMind 自分析被 `eval`、`benchmark` 等生成目录污染。

Features:

- 新增 `scripts/smoke-ai-provider.ps1`。
- Smoke 脚本支持 `-Provider`、`-Model`、`-RepoPath`、`-OutputDir`、`-Language`、`-Proxy`、`-TimeoutSeconds`。
- Smoke 脚本默认输出到 ignored `eval/ai-smoke-*`。
- Smoke 脚本生成 `ai-smoke.log`。
- Smoke 脚本生成 `ai-smoke-summary.json`。
- Smoke 脚本不打印 API key。
- README 增加 AI Provider smoke script 用法。
- Release checklist 增加 AI Provider smoke script。
- Scanner 默认忽略 `eval`。
- Scanner 默认忽略 `benchmark`。
- 本仓库自分析不再被真实仓库评估和 benchmark 生成物污染。

Unit Tests:

- Scanner 默认忽略 `.repomind`、`node_modules`、`eval`、`benchmark`。
- Scanner 不扫描 ignored generated files。

Integration Tests:

- `scripts/smoke-ai-provider.ps1 -Provider mock`。
- `scripts/smoke-ai-provider.ps1 -Provider grok -Model grok-4.3 -Proxy http://127.0.0.1:10809`。
- `go test ./internal/scanner ./internal/analyzer`。
- `go run ./cmd/repomind analyze --output <temp> .`。
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M23: Local Preflight Summary

Goal:

- 把发布前最常用的本地检查集中成一条命令，并生成可保存的 JSON/Markdown 检查报告。

Features:

- 新增 `scripts/preflight.ps1`。
- 默认执行 `go test ./...`。
- 默认执行 `go vet ./...`。
- 默认执行英文 `repomind analyze` smoke test。
- 默认执行中文 `repomind analyze --lang zh` smoke test。
- 生成 `summary.json`。
- 生成 `summary.md`。
- 支持 `-IncludeBenchmark` 调用真实仓库 benchmark。
- 支持 `-IncludeEvaluation` 调用真实仓库 evaluation。
- 支持 `-IncludeAISmoke` 调用 AI Provider smoke script。
- 支持 `-Proxy` 透传到 benchmark、evaluation 和 AI smoke。
- README 增加默认 preflight 命令。
- Release checklist 增加 preflight 入口。

Unit Tests:

- Not required for PowerShell orchestration script.

Integration Tests:

- `scripts/preflight.ps1 -TimeoutSeconds 180`。
- `scripts/preflight.ps1 -IncludeAISmoke -AIProvider mock`。

Status:

- Done

### M24: Release Artifact Smoke Test

Goal:

- 验证构建出的本平台二进制可以脱离 `go run` 正常完成核心工作流。

Features:

- 新增 `scripts/smoke-release-artifact.ps1`。
- Smoke 脚本构建当前平台 `repomind` 二进制。
- Smoke 脚本复制 fixture 到临时工作目录，避免污染原始 fixture。
- Smoke 脚本运行 `repomind version`。
- Smoke 脚本运行英文 `repomind analyze`。
- Smoke 脚本验证 `analysis.json` 和 `report.html` 生成。
- Smoke 脚本运行 `repomind export codex`。
- Smoke 脚本验证 `AGENTS.md` 生成。
- Smoke 脚本运行中文 `repomind analyze --lang zh`。
- Smoke 脚本验证中文分析结果写入 `language = zh`。
- Smoke 脚本生成 `summary.json` 和 `summary.md`。
- Preflight 增加 `-IncludeReleaseSmoke` 可选分支。
- README 和 release checklist 增加 release artifact smoke 用法。

Unit Tests:

- Not required for PowerShell orchestration script.

Integration Tests:

- `scripts/smoke-release-artifact.ps1 -TimeoutSeconds 180`。
- `scripts/preflight.ps1 -IncludeReleaseSmoke`。
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M25: Release Workflow Smoke Gate

Goal:

- 在 GitHub tag release artifact 构建流程中增加基础质量门禁，避免发布只构建、不验证二进制可运行。

Features:

- Release workflow 在 tag build 时执行 `go vet ./...`。
- Release workflow 对 linux/amd64 artifact 执行 built binary smoke。
- Built binary smoke 执行 `repomind version`。
- Built binary smoke 使用 fixture 副本执行英文 `analyze`。
- Built binary smoke 验证 `analysis.json` 和 `report.html`。
- Built binary smoke 执行 `export codex`。
- Built binary smoke 验证 `AGENTS.md`。
- Built binary smoke 执行中文 `analyze --lang zh`。
- Built binary smoke 验证 `analysis.json.language = zh`。

Unit Tests:

- Not required for GitHub Actions workflow YAML.

Integration Tests:

- Local `scripts/smoke-release-artifact.ps1` remains the equivalent local binary smoke path.
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M26: CI Analyze Smoke

Goal:

- 让普通 PR / main CI 除了单元测试和 vet，也覆盖 `repomind analyze` 的第一体验。

Features:

- CI workflow 增加英文 analyze smoke。
- CI workflow 验证英文 smoke 生成 `analysis.json`。
- CI workflow 验证英文 smoke 生成 `report.html`。
- CI workflow 增加中文 analyze smoke。
- CI workflow 验证中文 smoke 生成 `analysis.json`。
- CI workflow 验证中文 smoke 生成 `report.html`。
- CI workflow 验证中文 smoke 写入 `language = zh`。

Unit Tests:

- Not required for GitHub Actions workflow YAML.

Integration Tests:

- Local equivalent remains `scripts/preflight.ps1` default run.
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M27: Workflow Documentation

Goal:

- 把本地 preflight、CI、release、benchmark、evaluation、AI smoke 的关系集中说明，降低后续维护成本。

Features:

- 新增 `docs/WORKFLOWS.md`。
- 文档说明默认本地 preflight 覆盖项。
- 文档说明可选 release smoke、AI smoke、benchmark、evaluation。
- 文档说明 CI workflow 的检查项。
- 文档说明 release workflow 的检查项。
- 文档说明本地脚本和 GitHub Actions 的对应关系。
- 文档说明 generated artifacts 的 ignore 和 scanner hygiene 边界。
- README 链接 `docs/WORKFLOWS.md`。
- Release checklist 增加 `docs/WORKFLOWS.md` 更新检查。

Unit Tests:

- Not required for documentation-only milestone.

Integration Tests:

- Documentation references existing runnable scripts and workflows.

Status:

- Done

### M28: README Badges and Evaluation Snapshot

Goal:

- 提升开源首页可信度，让首次访问者立刻看到 CI/Release 状态和真实仓库 benchmark 结果。

Features:

- README 增加 CI workflow badge。
- README 增加 Release workflow badge。
- README 增加 Evaluation Snapshot。
- Evaluation Snapshot 展示真实仓库 benchmark 时间。
- README 链接 `docs/PERFORMANCE_BENCHMARKS.md`。
- README 链接 `docs/REAL_REPO_EVALUATION.md`。

Unit Tests:

- Not required for documentation-only milestone.

Integration Tests:

- Documentation references existing benchmark and evaluation docs.

Status:

- Done

### M29: Installation Documentation

Goal:

- 给用户明确的安装路径，覆盖开发者从源码运行、构建本地二进制、Go install、release binary 和 Windows PATH 配置。

Features:

- 新增 `docs/INSTALL.md`。
- 文档覆盖 `go run` 源码运行。
- 文档覆盖本地 `go build`。
- 文档覆盖未来发布后的 `go install`。
- 文档覆盖 Windows release binary 安装和 PATH 配置。
- 文档覆盖 macOS/Linux release binary 安装。
- 文档覆盖 AI Provider `.env` key 和 proxy 基础配置。
- README 链接安装文档。

Unit Tests:

- Not required for documentation-only milestone.

Integration Tests:

- Installation commands reference existing CLI entrypoint and release artifacts.

Status:

- Done

### M30: Documentation Index

Goal:

- 给文档目录提供一个稳定入口，避免 README、安装、工作流、发布、评估、benchmark 文档分散难找。

Features:

- 新增 `docs/README.md`。
- 文档索引列出 project plan、install、workflows、release checklist。
- 文档索引列出 real repository evaluation 和 performance benchmarks。
- 文档索引列出 report preview asset。
- README 链接 `docs/README.md`。
- Release checklist 增加 docs index 更新检查。

Unit Tests:

- Not required for documentation-only milestone.

Integration Tests:

- Documentation references existing files.

Status:

- Done

### M31: Parser Improvement Backlog

Goal:

- 把后续 parser 质量提升变成可跟踪 backlog，明确 tree-sitter 评估边界和各语言优先级。

Features:

- 新增 `docs/PARSER_BACKLOG.md`。
- 文档定义 parser 工作优先级。
- 文档定义 tree-sitter 引入边界。
- 文档列出 JS/TS parser backlog。
- 文档列出 Python parser backlog。
- 文档列出 PHP parser backlog。
- 文档列出 Java parser backlog。
- 文档列出 Go parser backlog。
- 文档定义 parser quality note 格式。
- README 链接 parser backlog。
- docs index 链接 parser backlog。
- Release checklist 增加 parser backlog 更新检查。

Unit Tests:

- Not required for documentation-only milestone.

Integration Tests:

- Documentation references current parser coverage and future work.

Status:

- Done

### M32: Evaluation Quality Score

Goal:

- 让真实仓库评估不仅展示扫描结果，还能输出轻量质量分数，作为 parser 改进的回归信号。

Features:

- `scripts/evaluate-repos.ps1` 增加 expected stack terms。
- `scripts/evaluate-repos.ps1` 增加 minimum route count checks。
- `scripts/evaluate-repos.ps1` 增加 minimum model count checks。
- `eval/summary.json` 增加 `quality_score`。
- `eval/summary.json` 增加 `quality_checks`。
- `eval/summary.md` 增加 Quality 列。
- `docs/REAL_REPO_EVALUATION.md` 说明 quality score 含义。
- `docs/REAL_REPO_EVALUATION.md` 更新最新 proxied evaluation 结果。
- Parser improvement tasks 更新为当前有效 backlog。

Unit Tests:

- Not required for PowerShell evaluation script.

Integration Tests:

- `scripts/evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809`。

Status:

- Done

### M33: FastAPI Router Prefix Parsing

Goal:

- 改善 FastAPI API map，识别常见 `APIRouter(prefix=...)` 和 `include_router(..., prefix=...)` 前缀组合。

Features:

- FastAPI route parser 捕获 decorator 的 router 变量名。
- FastAPI route parser 识别同文件 `APIRouter(prefix=...)`。
- FastAPI route parser 识别同文件 `include_router(router, prefix=...)`。
- FastAPI route path 合并 include prefix、router prefix 和 decorator path。
- Parser backlog 将 FastAPI prefix propagation 从待办移到已覆盖。

Unit Tests:

- FastAPI router prefix + include prefix + decorator path 合并。
- 既有 Django/FastAPI/Express/NestJS/Laravel/Spring/Go route parser 回归。

Integration Tests:

- `go test ./internal/parser/apiroute`。
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M34: Express Router Prefix Parsing

Goal:

- 改善 Express API map，识别同文件 `app.use("/prefix", router)` 和 router method routes 的前缀组合。

Features:

- Express route parser 捕获 route receiver 变量名。
- Express route parser 识别同文件 `app.use("/prefix", router)`。
- Express route path 合并 app/use prefix 和 router method path。
- Parser backlog 将 same-file Express router prefix propagation 标为已覆盖。

Unit Tests:

- Express router prefix + relative route path 合并。
- 既有 Django/FastAPI/Express/NestJS/Laravel/Spring/Go route parser 回归。

Integration Tests:

- `go test ./internal/parser/apiroute`。
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M35: Known Route and Model Quality Checks

Goal:

- 让真实仓库 evaluation 的质量分数不仅检查数量，也检查关键 route/model 是否被稳定识别。

Features:

- `scripts/evaluate-repos.ps1` 增加 `ExpectedRoutes`。
- `scripts/evaluate-repos.ps1` 增加 `ExpectedModels`。
- Quality checks 增加 known route path 检查。
- Quality checks 增加 known model name 检查。
- FastAPI expected route 更新为 prefix-aware `/users/me`。
- `docs/REAL_REPO_EVALUATION.md` 说明 known route/model checks。
- `docs/REAL_REPO_EVALUATION.md` 更新最新 quality score 结果。

Unit Tests:

- Not required for PowerShell evaluation script.

Integration Tests:

- `scripts/evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809`。

Status:

- Done

### M36: Evaluation Quality Gate

Goal:

- 让真实仓库 evaluation 可以作为 release/preflight 门禁，而不是只生成报告。

Features:

- `scripts/evaluate-repos.ps1` 增加 `-MinimumQualityScore`。
- Evaluation 默认要求 `quality_score >= 1.0`。
- Evaluation 在 clone 失败、analyze 失败或 quality score 低于阈值时返回非零。
- `eval/summary.md` 增加 `Status: PASS/FAIL`。
- `scripts/preflight.ps1` 增加 `-MinimumEvaluationQualityScore`。
- Preflight `-IncludeEvaluation` 透传 quality gate 阈值。
- `docs/REAL_REPO_EVALUATION.md` 说明 quality gate。
- `docs/WORKFLOWS.md` 说明 preflight evaluation quality threshold。

Unit Tests:

- Not required for PowerShell orchestration scripts.

Integration Tests:

- `scripts/evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -MinimumQualityScore 1.0`。
- `scripts/preflight.ps1 -IncludeEvaluation -Proxy http://127.0.0.1:10809 -MinimumEvaluationQualityScore 1.0`。

Status:

- Done

### M37: Combined Release Gate and Repository Cache

Goal:

- 将发布前的核心门禁合并为一个本地 release gate，并降低真实仓库 benchmark/evaluation 因重复 clone 导致的网络失败率。

Features:

- 新增/完善 `scripts/release-gate.ps1`。
- Release gate 调用默认 preflight。
- Release gate 自动启用 release binary smoke。
- Release gate 自动启用 real repository benchmark。
- Release gate 自动启用 real repository evaluation quality gate。
- Release gate 支持 `-IncludeAISmoke` 可选真实 AI Provider smoke。
- Benchmark 脚本新增 `-RepoCacheDir`。
- Evaluation 脚本新增 `-RepoCacheDir`。
- Benchmark/evaluation 在 cache 中存在 `.git` 时复用仓库，不重复 clone。
- Preflight 将 benchmark/evaluation 指向同一个 shared repo cache。
- Release gate 支持显式 `-RepoCacheDir`。
- Workflow 和 release checklist 文档说明 release gate、clone retry 和 repo cache。

Unit Tests:

- Not required for PowerShell orchestration scripts.

Integration Tests:

- `scripts/release-gate.ps1 -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5`。
- Release gate summary reports PASS.
- Benchmark summary reports PASS.
- Evaluation summary reports PASS.

Status:

- Done

### M38: Manual GitHub Release Gate Workflow

Goal:

- 提供一个手动触发的 GitHub Actions release gate，让重型 benchmark/evaluation/release smoke 不进入普通 PR CI，但发布前可以在云端复跑。

Features:

- 新增 `.github/workflows/release-gate.yml`。
- Workflow 支持 `workflow_dispatch` 手动触发。
- Workflow 支持输入 `timeout_seconds`。
- Workflow 支持输入 `clone_retries`。
- Workflow 支持输入 `benchmark_target_seconds`。
- Workflow 支持输入 `minimum_quality_score`。
- Workflow 在 `windows-latest` 上运行本地 `scripts/release-gate.ps1`。
- Workflow 上传 release gate summary artifacts。
- README 增加 Release Gate badge。
- Workflow documentation 增加 manual release gate workflow 说明。
- Release checklist 增加手动 GitHub release gate 步骤。

Unit Tests:

- Not required for GitHub Actions workflow YAML.

Integration Tests:

- Local release gate remains the equivalent validation path.
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M39: Release Gate Results Documentation

Goal:

- 将最近一次本地 release gate 的结果落到文档中，方便发布前审阅和后续回归对比。

Features:

- 新增 `docs/RELEASE_GATE_RESULTS.md`。
- 文档记录 release gate 命令、日期和 PASS 状态。
- 文档记录 preflight step summary。
- 文档记录 benchmark summary。
- 文档记录 evaluation quality summary。
- 文档记录 clone reset 失败原因和 shared repo cache 修复说明。
- README 链接 release gate result。
- docs index 链接 release gate result。
- Release checklist 增加 release gate result 更新检查。

Unit Tests:

- Not required for documentation-only milestone.

Integration Tests:

- Documentation is based on the successful local M37 release gate run.

Status:

- Done

### M40: Django Include Prefix Parsing

Goal:

- 改善 Django API map，识别同文件 `path("prefix/", include(patterns))` 对本地 URL pattern list 的前缀传播。

Features:

- Django URL parser 识别本地 list 赋值。
- Django URL parser 暂存非 `urlpatterns` list 中的 route。
- Django URL parser 对 `urlpatterns` 中直接 route 继续直接输出。
- Django URL parser 识别 `include(local_patterns)`。
- Django URL parser 合并 include prefix 和 child route path。
- Django include prefix 合并保留 child route 尾部 slash。
- Parser backlog 将 same-file Django include prefix propagation 标为已覆盖。

Unit Tests:

- Django same-file include prefix route 合并。
- Django `urlpatterns` direct route 回归。
- Django include child route 不重复输出无前缀路径。
- 既有 route parser 回归。

Integration Tests:

- `go test ./internal/parser/apiroute`。
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M41: Go Chi Route Prefix Parsing

Goal:

- 改善 Go API map，识别 chi 风格 `Route("/prefix", func(...))` 嵌套路由前缀。

Features:

- Go route AST parser 改为递归 route collector。
- Go route parser 识别 `Route("/prefix", func(...))`。
- Go route parser 对嵌套 route calls 合并父前缀。
- Go route parser 支持多层嵌套 route prefix。
- Go route parser 避免重复扫描 `Route` 子树导致无前缀重复输出。
- Parser backlog 将 same-file chi-style `Route` prefix propagation 标为已覆盖。

Unit Tests:

- Go chi-style nested `Route` prefix 合并。
- Go selector handler 回归。
- 既有 route parser 回归。

Integration Tests:

- `go test ./internal/parser/apiroute`。
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M42: Cross-file Route Prefix Strategy

Goal:

- 在继续实现跨文件 route prefix 传播前，先定义边界、IR 方向、解析规则和测试计划，避免 parser 复杂度失控。

Features:

- 新增 `docs/ROUTE_PREFIX_STRATEGY.md`。
- 文档列出现有 same-file prefix 覆盖。
- 文档说明 cross-file prefix 的问题范围。
- 文档定义设计目标。
- 文档提出 internal-only `RouteFragment` / `RouteMount` 方向。
- 文档定义 high-confidence resolution rules。
- 文档列出 Python/Django/FastAPI 计划。
- 文档列出 JS/TS/Express 计划。
- 文档列出 Go 计划。
- 文档定义测试计划和 rollout order。
- docs index 链接 route prefix strategy。
- Parser backlog 链接 route prefix strategy。

Unit Tests:

- Not required for design documentation.

Integration Tests:

- Documentation references current parser coverage and future parser work.

Status:

- Done

### M43: Release Native Smoke Matrix

Goal:

- 增强 tag release workflow 的二进制 smoke 覆盖，在 Windows、macOS、Linux 上都验证本平台 binary workflow。

Features:

- Release workflow 新增 `native-smoke` job。
- Native smoke matrix 覆盖 `ubuntu-latest`。
- Native smoke matrix 覆盖 `macos-latest`。
- Native smoke matrix 覆盖 `windows-latest`。
- Native smoke 调用 `scripts/smoke-release-artifact.ps1`。
- Release build job 依赖 native smoke。
- GitHub Release publish job 依赖 native smoke 和 build。
- Workflow documentation 更新 release workflow 检查项。

Unit Tests:

- Not required for GitHub Actions workflow YAML.

Integration Tests:

- Local `scripts/smoke-release-artifact.ps1` remains the equivalent current-platform smoke path.
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M44: Django Module Include Prefix Parsing

Goal:

- 实现 route prefix strategy 的第一步，识别 Django `include("module.urls")` 跨文件前缀传播。

Features:

- API route extractor 收集 route parser 输入文件内容。
- API route extractor 在普通 parser 后执行 Django module include resolution。
- Django resolver 构建 `urls.py` 文件到模块名的映射。
- Django resolver 支持后缀模块匹配，例如 `orders.urls`。
- Django resolver 识别 `path("prefix/", include("module.urls"))`。
- Django resolver 对 child `urlpatterns` route 合并父前缀。
- Django resolver 在成功解析 include 后过滤 child file 的无前缀 Django routes，减少误报。
- Route prefix strategy 标记 Django module include prefix 已覆盖。
- Parser backlog 标记 Django module include prefix 已覆盖。

Unit Tests:

- Django cross-file `include("orders.urls")` prefix propagation。
- Prefixed child route 输出。
- Unprefixed child route 被过滤。
- 既有 API route parser 回归。

Integration Tests:

- `go test ./internal/parser/apiroute`。
- `go test ./...`。
- `go vet ./...`。

Status:

- Done

### M45: Release Artifact Manifest

Goal:

- 为本地和 GitHub Release 构建产物生成 manifest，记录 archive、size 和 SHA256，方便发布校验和用户校验下载文件。

Features:

- `scripts/build-release.ps1` 生成 `dist/manifest.json`。
- `scripts/build-release.ps1` 生成 `dist/manifest.md`。
- Manifest 记录 version。
- Manifest 记录 GOOS / GOARCH。
- Manifest 记录 archive 文件名。
- Manifest 记录 archive size。
- Manifest 记录 archive SHA256。
- GitHub release workflow 在发布前生成 `manifest.json`。
- GitHub release workflow 在发布前生成 `manifest.md`。
- GitHub release workflow 上传 manifest 到 GitHub Release。
- README 说明 release manifest。
- Release checklist 增加 manifest 检查项。

Unit Tests:

- Not required for release packaging script.

Integration Tests:

- `scripts/build-release.ps1 -Version v0.0.0-manifest-test`。
- Manifest contains 6 platform artifacts.

Status:

- Done

### M46: Release Manifest Verification

Goal:

- 提供本地 release manifest 校验脚本，确认 manifest 中记录的 archive size 和 SHA256 与实际文件一致。

Features:

- 新增 `scripts/verify-release-manifest.ps1`。
- 校验 `manifest.json` 是否存在。
- 校验每个 archive 是否存在。
- 校验每个 archive size 是否一致。
- 校验每个 archive SHA256 是否一致。
- 生成 `manifest-verify.json`。
- 生成 `manifest-verify.md`。
- 校验失败时返回非零。
- README 增加 manifest verification 命令。
- Release checklist 增加 manifest verification 命令和 PASS 检查项。

Unit Tests:

- Not required for PowerShell release script.

Integration Tests:

- `scripts\verify-release-manifest.ps1 -DistDir dist` after M45 build.

Status:

- Done

### M47: Release Gate Manifest Build

Goal:

- 将跨平台 release manifest build/verification 接入本地 release gate，确保发布门禁不仅检查当前平台，也检查完整 release artifact manifest。

Features:

- Preflight 增加 `-IncludeManifestBuild`。
- Preflight 增加 `-ManifestVersion`。
- Manifest build step 调用 `scripts/build-release.ps1`。
- Manifest build step 调用 `scripts/verify-release-manifest.ps1`。
- Release gate 默认启用 manifest build/verification。
- Release gate 增加 `-SkipManifestBuild` 供排查时跳过。
- Workflow documentation 说明 release gate manifest build。
- README 说明 release gate 覆盖 manifest verification。
- Release checklist 增加 release gate manifest build 检查。

Unit Tests:

- Not required for PowerShell orchestration scripts.

Integration Tests:

- `scripts/preflight.ps1 -IncludeManifestBuild -ManifestVersion v0.0.0-preflight-manifest`。

Status:

- Done

### M48: Release Gate Manifest Artifact Upload

Goal:

- 让手动 GitHub Release Gate workflow 上传 release manifest build 和 verification 结果，方便发布前在 Actions artifact 中直接审查跨平台 release 包完整性。

Features:

- `.github/workflows/release-gate.yml` 上传 `manifest.json`。
- `.github/workflows/release-gate.yml` 上传 `manifest.md`。
- `.github/workflows/release-gate.yml` 上传 `manifest-verify.json`。
- `.github/workflows/release-gate.yml` 上传 `manifest-verify.md`。
- Workflow documentation 说明手动 release gate artifact 包含 manifest verification 文件。
- Release checklist 增加手动 release gate artifact 检查项。

Unit Tests:

- Not required for workflow artifact path changes.

Integration Tests:

- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M49: FastAPI Imported Router Prefix

Goal:

- 支持 FastAPI 常见的跨文件 router 组织方式，将 `app.include_router(users_router, prefix="/api/v1")` 的父级 prefix 传播到被导入 router 文件中的路由。

Features:

- 新增 FastAPI imported router resolver。
- 支持直接静态 import：`from app.api.routes.users import router as users_router`。
- 支持唯一模块后缀解析，覆盖常见 `app/...` 子目录布局。
- 只在 import、目标文件和 `APIRouter` symbol 都明确时解析，避免猜测。
- 解析成功时移除对应未挂载的原始 child route，避免 API map 出现重复未挂载路径。
- 解析失败时保留 child route fallback。
- Route prefix strategy 标记 FastAPI imported router prefix 已完成。
- Parser backlog 将 FastAPI 跨文件 router inclusion 收窄为更高级 import 模式。

Unit Tests:

- `TestExtractFastAPIImportedRouterPrefix`。
- `TestExtractFastAPIUnresolvedImportKeepsChildRoute`。

Integration Tests:

- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M50: Expanded Evaluation Known Checks

Goal:

- 增强真实仓库 evaluation quality gate，让它不仅检查最低 route/model 数量，也检查更多稳定的已知 route/model 名称。

Features:

- Gin 样本增加 `/book` 和 `/bookable` route 检查。
- FastAPI 样本增加 `/items`、`/users/signup`、`/utils/health-check` route 检查。
- Prisma 样本增加 `/api/filterPosts` 和 `/api/users` route 检查。
- Prisma 样本增加 `Account`、`Comment`、`Location` model 检查。
- Real repository evaluation 文档说明新增 known checks。

Unit Tests:

- Not required for evaluation expectation data changes.

Integration Tests:

- `scripts/evaluate-repos.ps1 -OutputDir eval/m50-evaluation -RepoCacheDir eval/release-gate/repo-cache -MinimumQualityScore 1.0`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M51: Express Relative Router Import Prefix

Goal:

- 支持 Express 常见的跨文件 router 组织方式，将父文件中的 `app.use("/api/orders", orderRouter)` prefix 传播到相对 import/require 的 router 文件。

Features:

- 新增 Express relative router import resolver。
- 支持 CommonJS：`const orderRouter = require("./routes/order")`。
- 支持 ES default import：`import userRouter from "./routes/users"`。
- 支持 named import 和 destructured require 的基础形式。
- 支持 `module.exports = router`、`export default router`、`export const router`、`exports.name = router`、`export { router }`。
- 仅解析相对路径 import，不解析外部 package import。
- 解析成功时移除对应未挂载的原始 child route，避免 API map 出现重复未挂载路径。
- 解析失败时保留 child route fallback。
- Route identity key 增加 line 字段，降低跨文件 prefix 过滤时的误删风险。
- Route prefix strategy 标记 Express relative router import prefix 已完成。
- Parser backlog 将 Express 跨文件 router import 收窄为更高级 alias/dynamic 模式。

Unit Tests:

- `TestExtractExpressRequireRouterPrefix`。
- `TestExtractExpressImportRouterPrefix`。
- `TestExtractExpressUnresolvedImportKeepsChildRoute`。

Integration Tests:

- `go test ./internal/parser/apiroute -v`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M52: Cross-file Route Resolver Real Evaluation

Goal:

- 用真实仓库 evaluation 验证 M49 FastAPI imported router prefix 和 M51 Express relative router import prefix 没有造成 route/model 质量回退。

Features:

- 复用 `eval/release-gate/repo-cache` 跑真实仓库 evaluation。
- 对比 M50 与 M52 的 route/model 数量。
- 确认五个样本仓库 quality score 均保持 1.00。
- Real repository evaluation 文档增加 resolver regression run 记录。
- 明确当前真实样本还没有覆盖新增跨文件 resolver 的 route count 提升，后续需要增加 split FastAPI/Express 样本。

Unit Tests:

- Not required for evaluation record-only milestone.

Integration Tests:

- `scripts/evaluate-repos.ps1 -OutputDir eval/m52-resolver-evaluation -RepoCacheDir eval/release-gate/repo-cache -MinimumQualityScore 1.0`。
- `git diff --check`。

Status:

- Done

### M53: Go Same-package Route Factory Prefix

Goal:

- 支持 Go/chi 常见的同 package route factory 挂载方式，将 `r.Mount("/api", orderRoutes())` 的父级 prefix 传播到 `orderRoutes()` 函数体内的 route。

Features:

- 新增 Go same-package route factory resolver。
- 支持同一目录、同一 package 内跨文件函数解析。
- 支持无参数 factory 调用：`Mount("/prefix", orderRoutes())`。
- 只在函数名在 package 内唯一时解析，避免猜测。
- 解析成功时移除对应未挂载的原始 child route，避免 API map 出现重复未挂载路径。
- 解析失败时保留 child route fallback。
- Route prefix strategy 标记 Go same-package route factory prefix 已完成。
- Parser backlog 将 Go cross-file route prefix 收窄为更高级变量 group、带参数 factory 和 sub-router 模式。

Unit Tests:

- `TestExtractGoSamePackageRouteFactoryPrefix`。
- `TestExtractGoUnresolvedRouteFactoryKeepsChildRoute`。

Integration Tests:

- `go test ./internal/parser/apiroute -v`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M54: Express Composed Router Evaluation Sample

Goal:

- 增加一个真实 split Express router 样本，并补齐 `Router().use("/api", api)` 包装多个 controller router 的 prefix 解析。

Features:

- Express resolver 支持 composed router：`const api = Router().use(controller)` 后 `export default Router().use("/api", api)`。
- Express relative import resolver 支持 dotted basename：`./tag/tag.controller` -> `tag.controller.ts`。
- Evaluation 增加 `gothinkster/node-express-realworld-example-app`。
- Evaluation 对新样本检查 `/api/tags`、`/api/articles`、`/api/users/login`。
- Evaluation 对新样本检查 Prisma models：`Article`、`Comment`、`Tag`、`User`。
- Real repository evaluation 文档增加 split Express router run。
- Route prefix strategy 和 parser backlog 记录 composed router 覆盖范围。

Unit Tests:

- `TestExtractExpressComposedRouterPrefix` 覆盖 composed router prefix 和 dotted basename import。

Integration Tests:

- `go test ./internal/parser/apiroute -v`。
- `scripts/evaluate-repos.ps1 -OutputDir eval/m54-evaluation -RepoCacheDir eval/release-gate/repo-cache -MinimumQualityScore 1.0`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M55: FastAPI Composed Router Static Prefix

Goal:

- 支持 FastAPI full-stack template 这类真实 split router 项目，把 `app.include_router(api_router, prefix=settings.API_V1_STR)` 的静态配置 prefix 传播到子 router 路由。

Features:

- FastAPI resolver 支持 `from app.api.routes import users` 加 `include_router(users.router)`。
- FastAPI resolver 支持递归解析 composed router。
- FastAPI resolver 支持唯一静态字符串常量 prefix，例如 `API_V1_STR: str = "/api/v1"`。
- Evaluation FastAPI known routes 从未挂载路径更新为 `/api/v1/...` 挂载后路径。
- Real repository evaluation 文档增加 FastAPI mounted prefix run。
- Route prefix strategy 和 parser backlog 记录 FastAPI composed router/static prefix 覆盖范围。

Unit Tests:

- `TestExtractFastAPIComposedRouterStaticPrefix`。

Integration Tests:

- `go test ./internal/parser/apiroute -v`。
- `go run ./cmd/repomind analyze --output eval/fastapi-prefix-m55 eval/release-gate/repo-cache/fastapi-full-stack-template`。
- `scripts/evaluate-repos.ps1 -OutputDir eval/m55-evaluation -RepoCacheDir eval/release-gate/repo-cache -MinimumQualityScore 1.0`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M56: Split Go Router Evaluation Sample

Goal:

- 增加真实 Go chi route 样本，验证 mounted route factory 解析在公开仓库中稳定工作。

Features:

- Evaluation 增加 `go-chi/chi`。
- Evaluation 对 Go chi 样本检查 `/admin/accounts`、`/admin/users/{userId}`、`/articles/search`。
- Real repository evaluation 文档增加 split Go router run。
- 记录当前 Go chi route parser 可识别 route，但 stack detector 尚未显式标记 Chi backend。

Unit Tests:

- Not required for evaluation sample data changes.

Integration Tests:

- `scripts/evaluate-repos.ps1 -OutputDir eval/m56-evaluation -RepoCacheDir eval/release-gate/repo-cache -MinimumQualityScore 1.0`。
- `git diff --check`。

Status:

- Done

### M57: Chi Stack Detection

Goal:

- 将 `go-chi/chi` 显式识别为 Go 后端框架 `Chi`，让 stack detection 与 route parser 能力保持一致。

Features:

- Go module detector 增加 `github.com/go-chi/chi`。
- Backend 输出顺序增加 `Chi`。
- Evaluation 中 `go-chi` 的 ExpectedStack 从 `Go` 提升为 `Chi`。
- README Go backend 支持列表增加 `Chi`。
- Real repository evaluation 文档增加 Chi stack detection run。

Unit Tests:

- `TestDetectStackFromChiGoMod`。

Integration Tests:

- `go test ./internal/detector -v`。
- `scripts/evaluate-repos.ps1 -OutputDir eval/m57-evaluation -RepoCacheDir eval/release-gate/repo-cache -MinimumQualityScore 1.0`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M58: Express Multi-line Route Calls

Goal:

- 提升 Express/TypeScript 项目的 route 覆盖率，识别 path 和 middleware 分多行书写的 `router.get(` / `router.post(` 调用。

Features:

- Express parser 增加轻量多行 route call 聚合。
- 支持 route path 和第一个 handler/middleware 跨行解析。
- `node-express-realworld` route 数量从 8 提升到 20。
- Evaluation 中 `node-express-realworld` 的 `MinRoutes` 提升到 10。
- Evaluation 新增 `/api/articles/feed`、`/api/articles/:slug/comments`、`/api/profiles/:username` known route checks。
- Parser backlog 记录 Express multi-line route coverage 和剩余动态路径边界。
- Real repository evaluation 文档增加 Express multi-line route run。

Unit Tests:

- `TestParseExpressMultilineRoutes`。

Integration Tests:

- `go test ./internal/parser/apiroute -v`。
- `go run ./cmd/repomind analyze --output eval/express-multiline-m58 eval/release-gate/repo-cache/node-express-realworld`。
- `scripts/evaluate-repos.ps1 -OutputDir eval/m58-evaluation -RepoCacheDir eval/release-gate/repo-cache -MinimumQualityScore 1.0`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M59: Go Receiver Method Route Factory

Goal:

- 支持 chi 资源对象常见写法 `r.Mount("/users", usersResource{}.Routes())`，将 mount prefix 传播到 receiver method factory 内的 route。

Features:

- Go route factory resolver 支持 composite literal receiver method call。
- 支持同 package method declaration 匹配，例如 `func (rs usersResource) Routes() chi.Router`。
- 只解析无参数 method factory，跳过带运行时参数的 factory。
- `go-chi` todos-resource 输出 `/users`、`/users/{id}`、`/todos`、`/todos/{id}/sync`。
- Evaluation 为 `go-chi` 增加 `/users/{id}` 和 `/todos/{id}/sync` known route checks。
- Parser backlog 记录 receiver method route factory coverage 和剩余边界。
- Real repository evaluation 文档增加 Go receiver method factory run。

Unit Tests:

- `TestExtractGoMethodRouteFactoryPrefix`。

Integration Tests:

- `go test ./internal/parser/apiroute -v`。
- `go run ./cmd/repomind analyze --output eval/go-method-factory-m59 eval/release-gate/repo-cache/go-chi`。
- `scripts/evaluate-repos.ps1 -OutputDir eval/m59-evaluation -RepoCacheDir eval/release-gate/repo-cache -MinimumQualityScore 1.0`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M60: Full Release Gate After Parser Expansion

Goal:

- 在 M48-M59 的 release、evaluation、FastAPI、Express、Go parser 扩展后，运行完整 release gate，确认组合门禁仍然通过。

Features:

- 完整 release gate 使用本地代理和共享 repo cache。
- Default preflight PASS。
- English/Chinese analyze smoke PASS。
- Benchmark PASS，5 个性能基准仓库均低于 30 秒。
- Evaluation PASS，7 个真实仓库 quality score 全部 1.00。
- Release artifact smoke PASS。
- Release manifest build and verification PASS。
- Release gate result 文档更新为最新运行结果。

Unit Tests:

- Covered by release gate `go test ./...`。

Integration Tests:

- `scripts/release-gate.ps1 -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5 -RepoCacheDir eval/release-gate/repo-cache`。
- `git diff --check`。

Status:

- Done

### M61: FastAPI Multi-line Decorators

Goal:

- 提升 FastAPI 项目的 route 覆盖率，识别 `@router.get(` 下一行才写 path 的多行 decorator。

Features:

- FastAPI parser 增加轻量多行 decorator 聚合。
- 支持 decorator path 和 metadata 分多行解析。
- `fastapi-full-stack-template` route 数量从 18 提升到 23。
- Evaluation 中 FastAPI 样本 `MinRoutes` 提升到 20。
- Evaluation 新增 `/api/v1/users` 和 `/api/v1/users/{user_id}` known route checks。
- Parser backlog 记录 FastAPI multi-line decorator coverage 和剩余动态路径边界。
- Real repository evaluation 文档增加 FastAPI multi-line decorator run。

Unit Tests:

- `TestParseFastAPIWithMultilineDecorator`。

Integration Tests:

- `go test ./internal/parser/apiroute -v`。
- `go run ./cmd/repomind analyze --output eval/fastapi-multiline-m61 eval/release-gate/repo-cache/fastapi-full-stack-template`。
- `scripts/evaluate-repos.ps1 -OutputDir eval/m61-evaluation -RepoCacheDir eval/release-gate/repo-cache -MinimumQualityScore 1.0`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M62: Django REST Framework Router Basics

Goal:

- 支持 Django REST Framework 常见的 `router.register(...)` + `include(router.urls)` 写法，让 DRF viewset 项目能在 API map 中出现常规 REST routes。

Features:

- Django URL parser 识别同文件 `router.register(r"users", views.UserViewSet, ...)`。
- Django URL parser 识别同文件 `path("api/", include(router.urls))`。
- 自动生成 collection routes：`GET /api/users/`、`POST /api/users/`。
- 自动生成 detail routes：`GET/PUT/PATCH/DELETE /api/users/{id}/`。
- DRF 推断 routes 使用 `medium` confidence，区别于显式 URL pattern。
- Parser backlog 记录 DRF same-file router coverage 和剩余 custom action/cross-file 边界。

Unit Tests:

- `TestParseDjangoRESTFrameworkRouterPrefix`。

Integration Tests:

- `go test ./internal/parser/apiroute -v`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M63: Git URL Repository Input

Goal:

- 支持 `repomind analyze https://github.com/owner/repo.git`，让用户可以直接分析 GitHub/Git 仓库 URL，而不必先手动 clone。

Features:

- 新增 repository input 准备层，区分本地路径和远程 Git URL。
- 支持 `https://`、`http://`、`ssh://`、`git://`、`git@...` 和 `file://` Git 输入。
- 远程输入使用 `git clone --depth 1` 克隆到临时目录。
- analyze 完成后自动清理临时 clone。
- 本地路径行为保持不变。
- 远程输入且 `--output` 为相对路径时，输出写入调用者当前工作目录，而不是临时 clone 目录。
- CLI help 更新为 `repomind analyze [path|git-url]`。
- README 增加 GitHub URL analyze 示例和代理说明。

Unit Tests:

- `TestIsRemote`。
- `TestPrepareClonesFileRemote`。

Integration Tests:

- `TestRunAnalyzeAcceptsGitURL`。
- `go test ./internal/repository -v`。
- `go test ./cmd/repomind -v`。
- `go test ./...`。
- `go vet ./...`。
- 真实 GitHub URL smoke：`go run ./cmd/repomind analyze --output eval/m63-github-url-smoke --max-files 1000 https://github.com/spring-guides/gs-rest-service.git`。
- `git diff --check`。

Status:

- Done

### M64: Git Ignore Source Directory Protection

Goal:

- 避免构建产物 ignore 规则误伤 `cmd/repomind` 源码目录，确保后续开源提交不会漏掉 CLI 入口代码。

Features:

- `.gitignore` 中的 `repomind` 和 `repomind.exe` 改为只匹配仓库根目录构建产物：`/repomind`、`/repomind.exe`。
- 保持 `.env`、`.repomind/`、`dist/`、`eval/`、`benchmark/` 等本地密钥和生成目录继续被忽略。
- 确认 `cmd/repomind/main.go` 不再被 Git ignore。

Unit Tests:

- Not required for ignore rule change.

Integration Tests:

- `git check-ignore -q cmd/repomind/main.go` must not match.
- `git check-ignore -v .env .repomind dist eval benchmark` must still match.
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M65: Remote Git Ref Selection

Goal:

- 支持远程 Git URL 分析时选择 branch 或 tag，让用户能分析非默认分支或固定版本。

Features:

- `repomind analyze` 新增 `--ref <branch-or-tag>`。
- `repomind analyze` 新增 `--branch <branch-or-tag>` 作为 `--ref` alias。
- 当 `--ref` 和 `--branch` 同时出现且值不一致时，CLI 返回明确错误。
- 远程 clone 使用 `git clone --depth 1 --branch <ref> --single-branch`。
- 本地路径传入 ref 时返回错误，避免误以为 RepoMind 会切换本地工作树。
- README 增加远程分支/tag analyze 示例。

Unit Tests:

- `TestPrepareClonesFileRemoteRef`。
- `TestPrepareRejectsRefForLocalPath`。

Integration Tests:

- `TestRunAnalyzeAcceptsGitURLRef`。
- `TestRunAnalyzeRejectsConflictingRefAndBranch`。
- `go test ./internal/repository -v`。
- `go test ./cmd/repomind -v`。
- 真实 GitHub URL ref smoke：`go run ./cmd/repomind analyze --ref main --output eval/m65-github-ref-smoke --max-files 1000 https://github.com/spring-guides/gs-rest-service.git`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M66: Remote Git Commit SHA Ref Support

Goal:

- 让远程 Git URL 的 `--ref` 不只支持 branch/tag，也能固定到可达 commit SHA，方便复现实验、排查历史版本和生成稳定报告。

Features:

- 指定 ref 时从 `git clone --branch` 路径改为 `git init` + `git fetch --depth 1 origin <ref>` + `git checkout --detach FETCH_HEAD`。
- branch、tag、可达 commit SHA 统一走 fetch/checkout 逻辑。
- 未指定 ref 时继续使用 `git clone --depth 1`，保持默认路径简单快速。
- README 将 `--ref` 说明更新为 branch、tag、commit SHA。

Unit Tests:

- `TestPrepareClonesFileRemoteCommitRef`。

Integration Tests:

- `go test ./internal/repository -v`。
- `go test ./cmd/repomind -v`。
- 真实 GitHub SHA smoke：`go run ./cmd/repomind analyze --ref <sha> --output eval/m66-github-sha-smoke --max-files 1000 https://github.com/spring-guides/gs-rest-service.git`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M67: Optional Remote Repository Clone Cache

Goal:

- 为反复分析同一个远程仓库提供可复用 clone cache，降低网络等待和 GitHub 请求压力，同时保持默认 analyze 行为简单。

Features:

- `repomind analyze` 新增 `--repo-cache <dir>`。
- cache 仅在用户显式传入时启用。
- cache 目录保存 bare Git repository。
- cache 不随单次 analyze 清理；临时分析 worktree 仍在 analyze 后清理。
- cache 已存在时执行 `git fetch --prune` 更新。
- analyzer 从 cache 生成临时分析目录。
- 支持与 `--ref` / `--branch` 一起使用。
- README 增加 `--repo-cache` 示例。

Unit Tests:

- `TestPrepareUsesRepositoryCache`。

Integration Tests:

- `TestRunAnalyzeAcceptsGitURLCache`。
- `go test ./internal/repository -v`。
- `go test ./cmd/repomind -v`。
- 真实 GitHub cache smoke：连续两次使用 `--repo-cache eval/m67-repo-cache` 分析 `https://github.com/spring-guides/gs-rest-service.git`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M68: Remote Repository Documentation

Goal:

- 给真实用户补齐远程仓库 analyze 使用说明，尤其是私有仓库认证、代理、ref 选择、clone cache 和安全边界。

Features:

- 新增 `docs/REMOTE_REPOSITORIES.md`。
- 文档覆盖本地路径与远程 Git URL 的使用差异。
- 文档覆盖 public repository、branch/tag/commit SHA、`--repo-cache`、proxy、private repository authentication。
- 文档说明 RepoMind 依赖 Git 的 SSH key、Git Credential Manager 或 credential helper，不自行保存私有仓库凭据。
- 文档明确默认不上传源码，只有显式启用网络 AI provider 时才调用外部模型。
- README 增加远程仓库文档入口。
- `docs/README.md` 增加远程仓库文档索引。
- `docs/RELEASE_CHECKLIST.md` 增加远程仓库文档检查项。

Unit Tests:

- Not required for documentation-only milestone.

Integration Tests:

- `git diff --check`。

Status:

- Done

### M69: Remote Git Failure Hints

Goal:

- 远程仓库 analyze 失败时给出更可执行的错误提示，减少用户面对原始 Git 输出时的排查成本。

Features:

- repository input 层新增 Git 失败分类。
- 对 ref 不存在或不可达的错误追加 `--ref` 检查提示。
- 对 authentication、permission denied、repository not found 等错误追加私有仓库认证提示。
- 对 DNS、连接失败、超时、connection reset、proxyconnect、TLS handshake timeout 等错误追加代理/网络检查提示。
- clone、cache clone、cache fetch、ref checkout 统一使用分类后的错误包装。
- 不改变正常 clone/fetch/analyze 路径。

Unit Tests:

- `TestClassifyGitFailure`。

Integration Tests:

- `go test ./internal/repository -v`。
- `go test ./cmd/repomind -v`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M70: Go Mounted Sub-router Variable Prefix

Goal:

- 支持 Go/Chi 常见的同一作用域 sub-router 变量挂载写法，提升 API map 对 Go 项目的 route prefix 还原能力。

Features:

- Go route parser 识别同一 block 内 `api := NewRouter()` 后的 `api.Get(...)` / `api.Post(...)` 等 child routes。
- Go route parser 识别同一 block 内 `r.Mount("/api", api)`。
- 对已解析 mounted variable 的 child routes 输出带 mount prefix 的 routes。
- 对同一 block 内已解析 mounted variable 的未加前缀 child routes 进行抑制，避免 API map 重复。
- 解析范围限定在 block scope，降低跨函数同名变量误关联风险。
- Parser backlog 将 same-block mounted sub-router variables 移到 current coverage。
- Route prefix strategy 文档记录该 rollout 已完成。
- Real repository evaluation 文档记录 M70 evaluation 结果。

Unit Tests:

- `TestParseGoRoutesWithMountedSubrouterVariable`。

Integration Tests:

- `go test ./internal/parser/apiroute -v`。
- `go run ./cmd/repomind analyze --output eval/m70-go-subrouter-variable eval/release-gate/repo-cache/go-chi`。
- `scripts/evaluate-repos.ps1 -OutputDir eval/m70-evaluation -RepoCacheDir eval/release-gate/repo-cache -MinimumQualityScore 1.0`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M71: Go Middleware-wrapped Handler Names

Goal:

- Go route handler 被 middleware/helper 包装时仍能识别真实 handler，减少 API map 中的漏报。

Features:

- `goHandlerName` 支持 `ast.CallExpr`。
- 对 `requireAuth(adminHandler)`、`middleware.Require(controller.Update)` 等包装调用，优先提取最后一个可识别参数作为 handler。
- 保持 inline handler、selector handler、普通 ident handler 的既有行为。
- Parser backlog 将常见 middleware-wrapped handler calls 移到 current coverage。
- Real repository evaluation 文档记录 M71 evaluation 结果。

Unit Tests:

- `TestParseGoRoutesWithASTHandlers` 增加 wrapper handler assertions。

Integration Tests:

- `go test ./internal/parser/apiroute -v`。
- `scripts/evaluate-repos.ps1 -OutputDir eval/m71-evaluation -RepoCacheDir eval/release-gate/repo-cache -MinimumQualityScore 1.0`。
- `go test ./...`。
- `go vet ./...`。
- `git diff --check`。

Status:

- Done

### M72: Full Release Gate Verification After Remote and Go Parser Updates

Goal:

- 对 M63-M71 累积的远程仓库输入、ref/cache、错误提示和 Go parser 增强执行完整发布门禁验证。

Features:

- 使用 `scripts/release-gate.ps1` 运行完整本地 release gate。
- 覆盖 `go test ./...`。
- 覆盖 `go vet ./...`。
- 覆盖英文和中文 analyze smoke。
- 覆盖真实仓库 benchmark。
- 覆盖 7 仓库真实 evaluation quality gate。
- 覆盖 release artifact smoke。
- 覆盖 release manifest build and verification。
- `docs/RELEASE_GATE_RESULTS.md` 更新为最新 release gate 输出。

Unit Tests:

- Covered by release gate `go test ./...`。

Integration Tests:

- `powershell -ExecutionPolicy Bypass -File scripts/release-gate.ps1 -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5 -RepoCacheDir eval/release-gate/repo-cache`。
- `git diff --check`。

Status:

- Done

### M73: Bilingual README Switch

Goal:

- 让 GitHub 首页支持英文/简体中文 README 切换，降低中文开发者首次理解成本，同时保留英文开源展示入口。

Features:

- `README.md` 顶部新增 language switch：English / 简体中文。
- 新增 `README.zh-CN.md`。
- `README.zh-CN.md` 顶部新增反向 language switch。
- README badges 更新为 `patrick892368/RepoMind` 仓库地址。
- Release artifact 说明更新为包含 `README.zh-CN.md`。
- `scripts/build-release.ps1` 将 `README.zh-CN.md` 打入 release archives。
- `.github/workflows/release.yml` 将 `README.zh-CN.md` 打入 GitHub release archives。
- `docs/RELEASE_CHECKLIST.md` 增加中文 README 和 release archive 检查项。
- 本地 Git remote `origin` 配置为 `https://github.com/patrick892368/RepoMind.git`。

Unit Tests:

- Not required for README and packaging metadata.

Integration Tests:

- `scripts/build-release.ps1 -Version v0.0.0-readme-bilingual -OutputDir dist/bilingual-readme`。
- `scripts/verify-release-manifest.ps1 -DistDir dist/bilingual-readme`。
- Archive content check confirms `README.zh-CN.md` is included.
- `git diff --check`。

Status:

- Done

## 7. Testing Strategy

### Unit Tests Required For

所有纯逻辑、解析器和生成器都必须有单元测试：

- 文件扫描。
- ignore 规则。
- 技术栈 detector。
- 配置文件 parser。
- package parser。
- route parser。
- model parser。
- relation parser。
- Mermaid generator。
- AI prompt builder。
- Provider response parser。
- JSON schema 输出。

### Integration Tests Required For

跨模块流程必须做组合测试：

- `repomind analyze .` 完整流程。
- fixture 仓库扫描。
- 技术栈 + 数据库 + API 合并输出。
- `.repomind/analysis.json` 生成。
- `.repomind/report.html` 生成。
- Mermaid 内容可被前端渲染。
- mock AI summary 注入完整流程。

### Manual Tests Required For

以下功能需要手动或半自动验收：

- 真实开源项目扫描速度。
- 大仓库扫描性能。
- 报告视觉效果。
- AI 总结质量。
- 真实 Provider 调用。
- Windows/macOS/Linux CLI 运行。

## 8. Quality Bar

每个功能完成时必须满足：

- 有确定性输出。
- 有错误降级。
- 不因为某一种框架解析失败导致整个 analyze 失败。
- 结果能落到 `analysis.json`。
- 报告页面能展示该结果。
- 至少覆盖单元测试或组合测试中的一种。
- 若是核心解析器，必须有单元测试。
- 若是端到端流程，必须有组合测试。

## 9. Update Rule

每完成一个功能、阶段或重要修复，都必须更新本文档。

必须更新的区域：

- 对应 milestone 的 `Status`。
- `Completed Features`。
- `Next Steps`。
- `Test Record`。
- 如果范围变化，更新 `Scope Boundary`。

状态只能使用：

```txt
Not started
In progress
Done
Blocked
Deferred
```

## 10. Completed Features

当前已完成：

- 项目目标定义。
- 产品定位定义。
- MVP 边界定义。
- 阶段规划 M1-M11。
- 测试策略定义。
- 文档持续更新规则定义。
- 混合扫描策略定义：规则扫描、静态解析、AI 理解、AI Coding 工具集成。
- Codex、Claude Code、Cursor 集成路线定义。
- M1: Go module initialized with module path `github.com/repomind/repomind`.
- M1: CLI entrypoint created at `cmd/repomind/main.go`.
- M1: `repomind analyze [path]` command implemented.
- M1: Basic repository path resolution implemented.
- M1: Basic file scanner implemented with default ignored directories.
- M1: `.repomind/analysis.json` writer implemented.
- M1: `.gitignore` added for `.repomind/` and Go build/test artifacts.
- M1: Test fixture added under `testdata/fixtures/basic-repo`.
- M1: Unit tests added for scanner and analyzer path/output behavior.
- M1: CLI integration test added for `repomind analyze`.
- M1: Manual CLI verification completed with `go run ./cmd/repomind analyze .`.
- M2: Stack detector implemented in `internal/detector`.
- M2: `analysis.json` stack schema extended with `config_files`.
- M2: package.json detection added for React, Vue, Next.js, Express, NestJS, database/cache/queue dependencies, and JS package managers.
- M2: Python dependency detection added for Django, FastAPI, Celery, Redis, Postgres, MySQL, MongoDB, SQLite, pip, and poetry.
- M2: docker-compose detection added for Postgres, MySQL, SQLite, MongoDB, Redis, and Celery.
- M2: config file detection added for package, Python, docker-compose, `.env.example`, and Django `settings.py` files.
- M2: detector unit tests added for package.json, Python files, Django settings, and docker-compose.
- M2: Stack fixture added under `testdata/fixtures/stack-repo`.
- M2: Analyzer integration test added to verify detected stack is written to `analysis.json`.
- M2: CLI integration test added to verify stack output is printed in the terminal and written to JSON.
- M2: Manual stack fixture verification completed with `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/stack-repo`.
- M3: IR extended with `models`, database fields, database relations, and `diagrams.er`.
- M3: Analyzer wired to database model extraction and Mermaid ER generation.
- M3: Mermaid ER generator added in `internal/graph`.
- M3: `internal/parser/dbmodel` implemented for Prisma, Django Models, SQLAlchemy, and TypeORM.
- M3: Database fixture added under `testdata/fixtures/db-repo`.
- M3: Parser unit tests added for Prisma, Django, SQLAlchemy, and TypeORM model extraction.
- M3: Graph unit test added for Mermaid ER generation.
- M3: Analyzer integration test added to verify `models` and `diagrams.er` are written to `analysis.json`.
- M3: CLI output now prints database model count when models are detected.
- M3: Manual database fixture verification completed with `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/db-repo`.
- M4: IR extended with `routes` and `diagrams.api`.
- M4: `internal/parser/apiroute` implemented for Django URL, FastAPI, Express, and NestJS route extraction.
- M4: Mermaid API graph generator added in `internal/graph`.
- M4: API fixture added under `testdata/fixtures/api-repo`.
- M4: Parser unit tests added for Django, FastAPI, Express, and NestJS route extraction.
- M4: Analyzer integration test added to verify `routes` and `diagrams.api` are written to `analysis.json`.
- M4: CLI output now prints API route count when routes are detected.
- M4: Manual API fixture verification completed with `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/api-repo`.
- M5: `internal/report` implemented for static HTML report generation.
- M5: Analyzer now writes `.repomind/report.html` alongside `.repomind/analysis.json`.
- M5: CLI output now prints the report path.
- M5: HTML report includes overview metrics, stack, languages, config files, directory/file lists, database models, API routes, Mermaid ER diagram, and Mermaid API map.
- M5: Report unit tests added for rendering and file output.
- M5: Analyzer and CLI integration tests updated to verify `report.html`.
- M5: Manual report verification completed with `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/api-repo`.
- M6: IR extended with structured `summary`.
- M6: `internal/ai` Provider interface added.
- M6: Offline summary provider implemented as the default mode.
- M6: Mock provider added for tests.
- M6: Provider config recognizes `offline`, `mock`, `openai`, `claude`, and `gemini`; real network calls for OpenAI/Claude/Gemini are explicitly not implemented yet.
- M6: Analyzer now writes summary into `analysis.json`.
- M6: CLI output now prints the generated summary.
- M6: HTML report now includes Project Summary, inferred modules, key flows, and start hints.
- M6: AI summary unit and integration tests added.
- M6: Manual summary verification completed with `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/api-repo`.
- M7: `internal/exporter` implemented for Codex, Claude Code, and Cursor context export.
- M7: `repomind export codex|claude|cursor` command added.
- M7: Export now generates `AGENTS.md`, `CLAUDE.md`, `.cursor/rules/repomind.md`, `.repomind/context.md`, `.repomind/architecture.md`, `.repomind/api-map.md`, and `.repomind/db-schema.md`.
- M7: Exporter unit tests added for Codex and Cursor output.
- M7: CLI integration test added for `repomind export claude`.
- M7: Manual export workflow verified in a temporary repository for Codex, Claude Code, and Cursor.
- M8: `internal/query` implemented for offline question-based repository lookup.
- M8: `repomind ask [path] --question "..."` command added.
- M8: Ask mode supports Chinese business keywords and English synonyms for order, dispatch, payment, wallet, balance, user, login, risk, cache, and queue.
- M8: Ask mode returns candidate files, handlers, models, and routes from `.repomind/analysis.json`.
- M8: Query unit tests added for Chinese order and balance questions.
- M8: CLI integration test added for `repomind ask`.
- M8: Manual ask workflow verified in a temporary repository after `repomind analyze`.
- M9: IR extended with `call_edges` and `diagrams.call`.
- M9: `internal/parser/callgraph` implemented for lightweight Python and JS/TS call edge extraction.
- M9: Analyzer now writes call edges and Mermaid call graph to `analysis.json`.
- M9: HTML report now includes Call Graph.
- M9: `repomind trace [path] --symbol ...` command added.
- M9: Trace mode performs bounded DFS over analyzed call edges and prints Mermaid output.
- M9: Call graph fixture added under `testdata/fixtures/call-repo`.
- M9: Parser, graph, analyzer, CLI, and trace tests added.
- M9: Manual trace workflow verified in a temporary repository.
- M10: `internal/diagnose` implemented for lightweight diagnostic scanning.
- M10: `repomind diagnose [path] --issue ...` command added.
- M10: Diagnose mode searches state modification points, database writes, cache operations, and queue task calls.
- M10: Diagnose fixture added under `testdata/fixtures/diagnose-repo`.
- M10: Diagnose unit and CLI integration tests added.
- M10: Manual diagnose workflow verified in a temporary repository.
- M11: Stack detector extended for PHP Composer, Java Maven/Gradle, and Go go.mod.
- M11: PHP Laravel, Symfony, ThinkPHP stack detection added.
- M11: Java Spring Boot stack detection added.
- M11: Go Gin, Echo, Fiber stack detection added.
- M11: API route parser extended for Laravel routes, Spring controllers, and Go router calls.
- M11: DB model parser extended for Java JPA entities and Go GORM structs.
- M11: Scanner default ignores `testdata` so repository self-analysis is not polluted by fixtures.
- M11: Go GORM parser tightened to avoid treating ordinary Go structs as database models.
- M11: Multi-language fixture added under `testdata/fixtures/multilang-repo`.
- M11: Detector, DB parser, API parser, and analyzer integration tests added for PHP/Java/Go support.
- M11: Manual multi-language analyze workflow verified with `testdata/fixtures/multilang-repo`.
- AI Provider: Grok/xAI provider implemented using `GROK_API_KEY` or `XAI_API_KEY` from environment variables or local `.env`.
- AI Provider: `.env` added to `.gitignore`; `.env.example` remains allowed.
- AI Provider: Grok Responses API support added with Chat Completions fallback for unsupported Responses API responses.
- AI Provider: Grok provider now reads `HTTPS_PROXY`, `HTTP_PROXY`, or `ALL_PROXY` from process environment or local `.env`.
- AI Provider: Grok summary parser now accepts model responses where list fields such as `stack` are returned as strings.
- AI Provider: Grok provider tests added with local HTTP test server; no real API key is used in tests.
- AI Provider: Real `--ai grok` verification succeeds through local HTTP proxy `127.0.0.1:10809` and local SOCKS proxy `127.0.0.1:10808`.
- Release hardening: `README.md` added with positioning, quick start, command list, Grok setup, supported detection, export workflow, development commands, and boundaries.
- Release hardening: `.env.example` added with Grok/xAI and proxy variable names only.
- Release hardening: GitHub Actions CI added at `.github/workflows/ci.yml` to run `go test ./...` and `go vet ./...`.
- Release hardening: PowerShell release build script added at `scripts/build-release.ps1`.
- Release hardening: GitHub tag-based release workflow added at `.github/workflows/release.yml`.
- Release hardening: `.gitignore` now ignores `dist/` release artifacts.
- Release hardening: README updated with build and release instructions.
- Release hardening: `.gitignore` now ignores `eval/` real-repository evaluation artifacts.
- M12: Real repository evaluation script added at `scripts/evaluate-repos.ps1`.
- M12: Evaluation script now uses explicit process execution, captures stdout/stderr logs, checks exit codes, and supports command timeouts.
- M12: Evaluation script now supports `-Proxy` and environment proxy variables for GitHub clone in restricted networks.
- M12: Evaluation script now passes absolute output directories so analysis artifacts are written outside cloned repositories.
- M12: Real repository evaluation notes added at `docs/REAL_REPO_EVALUATION.md`.
- M12: Real proxied evaluation completed against Laravel, Spring Boot, Gin, FastAPI full-stack template, and Prisma examples.
- M13: Route and model IR now include line number, confidence, and evidence fields.
- M13: Route parsers now populate evidence for Django, FastAPI, Express, NestJS, Laravel, Spring, and Go routes.
- M13: Model parsers now populate evidence for Prisma, Django, SQLAlchemy, SQLModel, TypeORM, JPA, and GORM models.
- M13: HTML report now displays location, confidence, and evidence for database models and API routes.
- M13: AI Coding tool exports now include source line and evidence for database models and API routes.
- M13: Grok prompt context now includes model/route location and confidence.
- M13: Python DB parser no longer treats Pydantic `BaseModel`, `BaseSettings`, or non-table SQLModel schema classes as DB models.
- M13: SQLModel `table=True` parser added with basic field and relationship extraction.
- M13: SQLModel/Pydantic regression fixture added under `testdata/fixtures/db-repo/sqlmodel_app`.
- M14: `packages` IR added to `analysis.json`.
- M14: `internal/workspace` added for package root detection and per-package stack/count aggregation.
- M14: monorepo fixture added under `testdata/fixtures/monorepo`.
- M14: Analyzer now writes package grouping into `analysis.json`.
- M14: HTML report now displays Packages.
- M14: AI Coding tool exports now include a package overview.
- M14: Real Prisma examples evaluation now exposes root package, starter packages, and nested Prisma schema packages.
- M14: Real FastAPI full-stack template evaluation now separates root, backend, and frontend packages.
- M15: Report preview image added at `docs/assets/report-preview.png`.
- M15: README now includes Report Preview.
- M15: README now documents monorepo package grouping and SQLModel support.
- M16: `--lang en|zh` added to `repomind analyze`.
- M16: `analysis.json` now records output language.
- M16: Offline summary, Grok prompt, CLI analyze output, HTML report labels, ask summary, trace title, and diagnose summary now support English or Simplified Chinese.
- M16: README quick start now documents Chinese output.
- M17: `repomind analyze` now supports `--max-files`, `--max-file-bytes`, and `--max-call-edges`.
- M17: Analyzer now marks `scan.truncated` when file or call graph limits are hit.
- M17: HTML report now warns when analysis is truncated.
- M17: Benchmark script added at `scripts/benchmark-repos.ps1`.
- M17: Benchmark documentation added at `docs/PERFORMANCE_BENCHMARKS.md`.
- M17: `.gitignore` now ignores `benchmark/`.
- M18: Release checklist added at `docs/RELEASE_CHECKLIST.md`.
- M18: README now links to release checklist.
- M19: Package IR now includes `parent` and `dependencies`.
- M19: `diagrams.package` added to `analysis.json`.
- M19: Workspace detector now infers package hierarchy and local manifest dependencies.
- M19: HTML report now includes package parent/dependencies and Package Graph.
- M19: AI Coding tool exports now include package hierarchy and dependency graph.
- M20: Real OpenAI Provider implemented through the Responses API.
- M20: Real Claude/Anthropic Provider implemented through the Messages API.
- M20: Real Gemini/Google Provider implemented through the `generateContent` API.
- M20: Network AI providers now share proxy, timeout, response-size limit, and HTTP error handling.
- M20: `.env.example` and README now document OpenAI, Claude/Anthropic, Gemini/Google, and Grok/xAI keys.
- M20: Network AI provider tests use local HTTP test servers and do not require real API keys.
- M21: Go API route parser now uses Go AST first and regex fallback second.
- M21: Go route extraction now handles selector handlers and inline handlers.
- M21: Go GORM model parser now uses Go AST first and regex fallback second.
- M21: Go GORM model extraction now handles embedded `gorm.Model` and selector relation targets.
- M21: Callgraph extraction now supports Go files.
- M22: Optional AI Provider smoke script added at `scripts/smoke-ai-provider.ps1`.
- M22: README and release checklist now document the AI Provider smoke workflow.
- M22: Scanner now ignores generated `eval` and `benchmark` directories by default.
- M22: Real Grok smoke test through `127.0.0.1:10809` now produces a clean RepoMind summary without generated-directory pollution.
- M23: Local preflight summary script added at `scripts/preflight.ps1`.
- M23: Preflight now runs test, vet, English analyze smoke, and Chinese analyze smoke by default.
- M23: Preflight can optionally run benchmark, real repository evaluation, and AI Provider smoke.
- M23: README and release checklist now document the preflight workflow.
- M24: Release artifact smoke script added at `scripts/smoke-release-artifact.ps1`.
- M24: Release smoke validates current-platform binary `version`, analyze, export, and Chinese analyze workflows.
- M24: Preflight now supports `-IncludeReleaseSmoke`.
- M24: README and release checklist now document release artifact smoke.
- M25: Release workflow now runs `go vet ./...` before artifact build.
- M25: Release workflow now smoke-tests the linux/amd64 built binary before upload.
- M25: Release smoke gate validates version, analyze, Codex export, and Chinese output from the built binary.
- M26: CI workflow now runs English and Chinese analyze smoke tests.
- M26: CI analyze smoke verifies `analysis.json`, `report.html`, and Chinese `language = zh`.
- M27: Workflow mapping documentation added at `docs/WORKFLOWS.md`.
- M27: README and release checklist now link the workflow documentation.
- M28: README now includes CI and Release badges.
- M28: README now includes an Evaluation Snapshot with real repository benchmark times.
- M28: README now links performance benchmark and real repository evaluation docs from the snapshot.
- M29: Installation documentation added at `docs/INSTALL.md`.
- M29: README now links installation documentation.
- M30: Documentation index added at `docs/README.md`.
- M30: README and release checklist now link or require the documentation index.
- M31: Parser improvement backlog added at `docs/PARSER_BACKLOG.md`.
- M31: Parser backlog defines tree-sitter adoption boundaries and per-language parser priorities.
- M31: README, docs index, and release checklist now link the parser backlog.
- M32: Real repository evaluation script now emits `quality_score` and `quality_checks`.
- M32: Real repository evaluation Markdown now includes a Quality column.
- M32: Real repository evaluation docs now explain quality score and include the latest proxied run.
- M33: FastAPI route parser now propagates same-file `APIRouter` and `include_router` prefixes.
- M33: Parser backlog now marks FastAPI same-file prefix propagation as covered.
- M34: Express route parser now propagates same-file `app.use("/prefix", router)` prefixes.
- M34: Parser backlog now marks same-file Express router prefix propagation as covered.
- M35: Real repository quality checks now include selected known route paths.
- M35: Real repository quality checks now include selected known model names.
- M35: FastAPI known route expectation now reflects prefix-aware `/users/me`.
- M36: Real repository evaluation now exits nonzero on clone/analyze failure or quality score below threshold.
- M36: Preflight now passes `MinimumEvaluationQualityScore` to evaluation.
- M36: Evaluation and workflow docs now describe the quality gate.
- M37: Local release gate script now combines default preflight, release smoke, benchmark, and evaluation quality gate.
- M37: Benchmark and evaluation scripts now support `RepoCacheDir` to reuse cloned repositories.
- M37: Preflight now shares one repository cache between benchmark and evaluation.
- M37: Release gate documentation now covers clone retry and persistent repo cache usage.
- M38: Manual GitHub Release Gate workflow added at `.github/workflows/release-gate.yml`.
- M38: README now includes a Release Gate badge.
- M38: Workflow docs and release checklist now describe the manual release gate workflow.
- M39: Release gate result documentation added at `docs/RELEASE_GATE_RESULTS.md`.
- M39: README, docs index, and release checklist now link or require the release gate result document.
- M40: Django URL parser now propagates same-file `include()` prefixes for local pattern lists.
- M40: Parser backlog now marks same-file Django include prefix propagation as covered.
- M41: Go route parser now propagates same-file chi-style nested `Route` prefixes.
- M41: Parser backlog now marks same-file chi-style route prefix propagation as covered.
- M42: Cross-file route prefix strategy added at `docs/ROUTE_PREFIX_STRATEGY.md`.
- M42: Docs index and parser backlog now link the route prefix strategy.
- M43: Release workflow now runs native binary smoke on Ubuntu, macOS, and Windows.
- M43: Release build and publish jobs now depend on native smoke.
- M44: Django route parser now resolves `include("module.urls")` cross-file prefixes.
- M44: Route prefix strategy and parser backlog now mark Django module include prefix as covered.
- M45: Local release build now generates `manifest.json` and `manifest.md` with SHA256 hashes.
- M45: GitHub release workflow now uploads release manifest files.
- M45: README and release checklist now document release manifests.
- M46: Release manifest verification script added at `scripts/verify-release-manifest.ps1`.
- M46: README and release checklist now document manifest verification.
- M47: Preflight now supports optional release manifest build and verification.
- M47: Release gate now includes manifest build and verification by default.
- M47: Workflow docs, README, and release checklist now document release gate manifest verification.
- M48: Manual GitHub Release Gate workflow now uploads release manifest and manifest verification files.
- M48: Workflow docs and release checklist now describe manual release gate manifest artifact verification.
- M49: FastAPI route parser now resolves direct imported router prefixes across files.
- M49: FastAPI resolver keeps fallback child routes when an import cannot be resolved.
- M49: Route prefix strategy and parser backlog now mark direct FastAPI imported router prefix as covered.
- M50: Real repository evaluation now checks additional stable Gin, FastAPI, and Prisma routes.
- M50: Real repository evaluation now checks additional Prisma model names.
- M50: Real repository evaluation docs now describe the expanded known checks.
- M51: Express route parser now resolves direct relative router imports across files.
- M51: Express resolver supports CommonJS require, ES default import, and basic named router imports.
- M51: Route prefix strategy and parser backlog now mark direct Express relative router import prefix as covered.
- M52: Real repository evaluation passed after FastAPI and Express cross-file route resolver changes.
- M52: Real repository evaluation docs now record the resolver regression run and route/model comparison.
- M53: Go route parser now resolves direct same-package `Mount("/prefix", routeFactory())` calls.
- M53: Go resolver keeps fallback child routes when a route factory cannot be resolved.
- M53: Route prefix strategy and parser backlog now mark direct Go same-package route factory prefix as covered.
- M54: Express parser now resolves simple composed router exports mounted with `Router().use("/api", api)`.
- M54: Express import resolver now resolves dotted TypeScript basenames such as `./tag/tag.controller`.
- M54: Real repository evaluation now includes `node-express-realworld` with split Express router quality checks.
- M55: FastAPI parser now resolves composed routers using `module.router` references.
- M55: FastAPI parser now resolves unique static string constants used as include_router prefixes.
- M55: Real repository evaluation now checks mounted `/api/v1/...` FastAPI routes.
- M56: Real repository evaluation now includes `go-chi` with mounted Go route factory quality checks.
- M56: Evaluation docs now record that Chi stack detection should be added next.
- M57: Stack detector now identifies `github.com/go-chi/chi` as backend `Chi`.
- M57: Real repository evaluation now expects and confirms `Chi` for the `go-chi` sample.
- M58: Express parser now recognizes common multi-line route calls.
- M58: `node-express-realworld` route coverage increased from 8 to 20 in real repository evaluation.
- M59: Go route parser now resolves same-package receiver method route factories mounted with `Mount`.
- M59: `go-chi` evaluation now checks mounted receiver method routes such as `/users/{id}` and `/todos/{id}/sync`.
- M60: Full release gate passed after parser and evaluation expansion.
- M60: Release gate results now document 7-repository evaluation and manifest verification.
- M61: FastAPI parser now recognizes common multi-line route decorators.
- M61: `fastapi-full-stack-template` route coverage increased from 18 to 23 in real repository evaluation.
- M62: Django parser now recognizes basic same-file Django REST Framework router registrations.
- M63: `repomind analyze` now accepts Git remote URLs and shallow-clones them into a temporary analysis directory.
- M63: Remote analyze output is written under the caller working directory when `--output` is relative.
- M63: Remote input is covered by repository package clone tests, CLI integration tests, and a real GitHub URL smoke through the local proxy.
- M64: `.gitignore` now ignores only root-level RepoMind binaries and no longer ignores the `cmd/repomind` source directory.
- M65: Remote Git URL analysis now supports `--ref` and `--branch` for branch or tag selection.
- M65: CLI rejects conflicting `--ref` / `--branch` values and local-path ref usage is rejected by the repository input layer.
- M66: Remote Git `--ref` now supports branch, tag, and reachable commit SHA through fetch/checkout.
- M67: Remote Git URL analysis now supports an explicit reusable bare clone cache with `--repo-cache`.
- M68: Remote repository usage documentation now covers GitHub/Git URLs, private repository authentication, proxy setup, ref selection, clone cache, and safety boundaries.
- M69: Remote Git clone, cache, fetch, and ref checkout failures now include classified hints for missing refs, authentication/access failures, and network/proxy failures.
- M70: Go route parser now resolves same-block mounted sub-router variables such as `api.Get(...)` mounted through `r.Mount("/api", api)`.
- M71: Go route parser now extracts handler names from common middleware wrapper calls such as `requireAuth(handler)`.
- M72: Full release gate passed after remote repository input and Go parser updates through M71.
- M73: GitHub README now supports English / Simplified Chinese switching, and release archives include `README.zh-CN.md`.

## 11. Next Steps

当前 M1-M73 已完成。下一步继续 parser 和远程输入体验质量提升：

1. 继续补充更多真实仓库样本和质量检查。
2. 增强 Go 带参数 route factory、middleware chain 和 cross-package route assembly 解析。
3. 增强 Python route metadata、Django REST Framework custom action/cross-file router 解析。
4. 为远程仓库 cache 增加可选清理策略和 cache size 文档。

## 12. Test Record

当前测试记录：

- Go module initialized. No tests yet.
- `go test ./...` passed after initial CLI/scanner implementation. No test files yet.
- `go test ./...` passed with M1 unit and integration tests.
- `go run ./cmd/repomind analyze .` passed and generated `.repomind/analysis.json`.
- `go test ./internal/detector -v` passed for M2 detector unit tests.
- `go test ./...` passed with M2 detector, analyzer, and CLI integration tests.
- `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/stack-repo` passed and generated stack fields in terminal output and `analysis.json`.
- `go test ./...` passed with M3 database parser, ER graph, analyzer, and CLI tests.
- `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/db-repo` passed and generated 9 models plus Mermaid ER output.
- `go test ./...` passed with M4 API route parser, API graph, analyzer, and CLI tests.
- `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/api-repo` passed and generated 8 routes plus Mermaid API output.
- `go test ./...` passed with M5 report rendering, analyzer, and CLI tests.
- `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/api-repo` passed and generated `report.html` containing overview, API routes, API map, and Mermaid markup.
- `go test ./...` passed with M6 offline AI summary, analyzer, report, and CLI tests.
- `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/api-repo` passed and generated summary in terminal output, `analysis.json`, and `report.html`.
- `go test ./...` passed with M7 exporter and CLI tests.
- Temporary repository workflow passed for `repomind analyze`, `repomind export codex`, `repomind export claude`, and `repomind export cursor`.
- `go test ./...` passed with M8 query and CLI ask tests.
- Temporary repository workflow passed for `repomind analyze` followed by `repomind ask --question "订单是怎么创建的？"`.
- `go test ./...` passed with M9 callgraph parser, trace, analyzer, report, and CLI tests.
- Temporary repository workflow passed for `repomind analyze` followed by `repomind trace --symbol pay_callback`.
- `go test ./...` passed with M10 diagnose and CLI tests.
- Temporary repository workflow passed for `repomind analyze` followed by `repomind diagnose --issue "订单状态异常"`.
- `go test ./...` passed with M11 PHP/Java/Go stack, route, model, and analyzer tests.
- `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/multilang-repo` passed and detected Spring Boot, Laravel, Gin, JPA/GORM models, and Laravel/Spring/Go routes.
- `go run ./cmd/repomind analyze .` passed after fixture ignore and GORM filtering; RepoMind self-analysis no longer reports fixture frameworks or ordinary Go structs as app models.
- `go test ./...` passed with Grok/xAI provider unit tests and Chat Completions fallback tests.
- `go run ./cmd/repomind analyze --ai grok --ai-model grok-4.3 .` attempted a real Grok call using local `.env`, but failed before authentication with TCP timeout to `api.x.ai:443`.
- `go test ./...` passed after adding `.env` proxy support for Grok/xAI.
- `go run ./cmd/repomind analyze --ai grok --ai-model grok-4.3 .` passed with `HTTPS_PROXY=http://127.0.0.1:10809`.
- `go run ./cmd/repomind analyze --ai grok --ai-model grok-4.3 .` passed with `ALL_PROXY=socks5://127.0.0.1:10808`.
- `README.md`, `.env.example`, and `.github/workflows/ci.yml` added for release hardening.
- `scripts/build-release.ps1` and `.github/workflows/release.yml` added for local and GitHub release builds.
- `scripts/build-release.ps1 -Version v0.0.0-test` passed and generated Windows, macOS, and Linux artifacts under ignored `dist/`.
- `go test ./...` and `go vet ./...` passed after release build changes.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -TimeoutSeconds 300` failed to clone GitHub repositories without proxy because `github.com:443` was reset or timed out in the current network.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809` passed against 5 public repositories.
- Real evaluation results: Laravel analyzed 63 files in 0.59s with Laravel and 1 route detected.
- Real evaluation results: Spring REST service analyzed 59 files in 0.21s with Spring Boot and 1 route detected.
- Real evaluation results: Gin examples analyzed 124 files in 0.89s with Gin, 67 routes, and 234 call edges detected.
- Real evaluation results: FastAPI full-stack template analyzed 227 files in 1.93s with FastAPI, React, Postgres, 10 models, 18 routes, and 851 call edges detected.
- Real evaluation results: Prisma examples analyzed 1374 files in 4.95s with NestJS, Express, Next.js, Vue, React, Postgres, 143 models, 55 routes, and 1764 call edges detected.
- `go test ./internal/parser/dbmodel -v` passed after SQLModel and Pydantic false-positive regression tests.
- `go test ./...` passed after evidence/confidence fields and SQLModel parser changes.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809` passed after M13 parser changes.
- Updated real evaluation results: FastAPI full-stack template now reports 2 SQLModel table models instead of 10 mixed SQLModel/Pydantic classes.
- `go test ./internal/workspace ./internal/analyzer -v` passed after adding package grouping.
- `go test ./...` passed after M14 package grouping changes.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809` passed after M14 package grouping changes.
- M14 real evaluation: FastAPI full-stack template reports packages for root, `backend`, and `frontend`.
- M14 real evaluation: Prisma examples reports monorepo package entries for root, starters, and nested `prisma` schema directories.
- Browser verification passed for generated `report.html` served from local `127.0.0.1:8765`: title, Packages table, route evidence, and nonblank rendered page were confirmed.
- `docs/assets/report-preview.png` was generated from the browser-verified report and manually inspected.
- `go test ./cmd/repomind ./internal/ai ./internal/report ./internal/query ./internal/diagnose ./internal/trace ./internal/analyzer` passed after bilingual output changes.
- `go run ./cmd/repomind analyze --lang zh --output <eval-dir> testdata/fixtures/monorepo` passed and produced Chinese CLI output plus `analysis.json.language = "zh"`.
- `go test ./internal/ai`, `go test ./...`, and `go vet ./...` passed after adding Grok language fallback.
- `go test ./internal/analyzer ./cmd/repomind ./internal/report` passed after M17 large repository guards.
- `powershell -ExecutionPolicy Bypass -File scripts\benchmark-repos.ps1 -TimeoutSeconds 300 -TargetSeconds 30 -Proxy http://127.0.0.1:10809` passed.
- M17 benchmark results: Laravel 7.41s, Spring REST service 1.94s, Gin examples 1.32s, FastAPI full-stack template 2.99s, Prisma examples 9.68s. All are under the 30s target.
- `docs/RELEASE_CHECKLIST.md` added and linked from README.
- `go test ./internal/workspace ./internal/graph ./internal/analyzer ./internal/report ./internal/exporter` passed after M19 package hierarchy changes.
- `go run ./cmd/repomind analyze --output <eval-dir> testdata/fixtures/monorepo` passed and produced package parent/dependency data plus Mermaid package graph.
- `go test ./internal/ai` passed after adding real OpenAI, Claude/Anthropic, and Gemini/Google Provider implementations with local HTTP test servers.
- `go test ./...` passed after M20 Provider hardening.
- `go vet ./...` passed after M20 Provider hardening.
- `go test ./internal/parser/apiroute` passed after Go route AST parser changes.
- `go test ./internal/parser/dbmodel` passed after Go GORM AST parser changes.
- `go test ./internal/parser/callgraph` passed after adding Go callgraph extraction.
- `go test ./internal/parser/apiroute ./internal/parser/dbmodel ./internal/parser/callgraph ./internal/analyzer` passed after M21 parser hardening.
- `go run ./cmd/repomind analyze --output <temp> testdata/fixtures/multilang-repo` passed after M21 and produced Laravel, Spring, and Go API routes plus Java/GORM models.
- `go test ./...` passed after M21 Go parser AST hardening.
- `go vet ./...` passed after M21 Go parser AST hardening.
- `powershell -ExecutionPolicy Bypass -File scripts\smoke-ai-provider.ps1 -Provider mock -RepoPath testdata\fixtures\api-repo -OutputDir <temp>` passed and wrote `ai-smoke-summary.json`.
- `powershell -ExecutionPolicy Bypass -File scripts\smoke-ai-provider.ps1 -Provider grok -Model grok-4.3 -Proxy http://127.0.0.1:10809` passed with the local `.env` key.
- Real Grok smoke after scanner hygiene reported RepoMind as a Go-based repository analyzer instead of an eval/benchmark aggregate.
- `go test ./internal/scanner ./internal/analyzer` passed after adding generated-directory ignore rules.
- `go run ./cmd/repomind analyze --output <temp> .` passed and ignored `benchmark` and `eval`, producing 71 scanned files.
- `go test ./...` passed after M22 AI smoke and scanner hygiene changes.
- `go vet ./...` passed after M22 AI smoke and scanner hygiene changes.
- `powershell -ExecutionPolicy Bypass -File scripts\preflight.ps1 -TimeoutSeconds 180` passed and wrote `eval/preflight/summary.md` with test, vet, English analyze, and Chinese analyze smoke checks.
- `powershell -ExecutionPolicy Bypass -File scripts\preflight.ps1 -OutputDir <temp> -TimeoutSeconds 180 -IncludeAISmoke -AIProvider mock` passed and verified the optional AI smoke branch without a real API key.
- `go test ./...` passed after M23 preflight script changes.
- `go vet ./...` passed after M23 preflight script changes.
- `powershell -ExecutionPolicy Bypass -File scripts\smoke-release-artifact.ps1 -TimeoutSeconds 180` passed and validated build, version, English analyze, Codex export, and Chinese analyze from the built binary.
- `powershell -ExecutionPolicy Bypass -File scripts\preflight.ps1 -OutputDir <temp> -TimeoutSeconds 180 -IncludeReleaseSmoke` passed and verified the release smoke preflight branch.
- `go test ./...` passed after M24 release artifact smoke changes.
- `go vet ./...` passed after M24 release artifact smoke changes.
- Release workflow updated to run `go vet ./...` and smoke-test the linux/amd64 built binary before artifact upload.
- `go test ./...` passed after M25 release workflow smoke gate changes.
- `go vet ./...` passed after M25 release workflow smoke gate changes.
- CI workflow updated to run English and Chinese analyze smoke tests after `go test` and `go vet`.
- `powershell -ExecutionPolicy Bypass -File scripts\preflight.ps1 -TimeoutSeconds 180` passed after M26 CI analyze smoke changes.
- `docs/WORKFLOWS.md` added and linked from README and release checklist after M27 workflow documentation.
- README CI/Release badges and Evaluation Snapshot added after M28 README presentation update.
- `docs/INSTALL.md` added and linked from README after M29 installation documentation.
- `docs/README.md` added and linked from README and release checklist after M30 documentation index.
- `go test ./...` passed after M30 documentation index changes.
- `go vet ./...` passed after M30 documentation index changes.
- `docs/PARSER_BACKLOG.md` added and linked from README, docs index, and release checklist after M31 parser backlog.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809` passed after M32 quality score changes.
- M32 real evaluation quality scores were 1.00 for Laravel, Spring REST service, Gin examples, FastAPI full-stack template, and Prisma examples.
- `go test ./internal/parser/apiroute` passed after M33 FastAPI router prefix parsing.
- `go test ./...` passed after M33 FastAPI router prefix parsing.
- `go vet ./...` passed after M33 FastAPI router prefix parsing.
- `go test ./internal/parser/apiroute` passed after M34 Express router prefix parsing.
- `go test ./...` passed after M34 Express router prefix parsing.
- `go vet ./...` passed after M34 Express router prefix parsing.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809` passed after M35 known route/model quality checks.
- M35 real evaluation quality scores were 1.00 for Laravel, Spring REST service, Gin examples, FastAPI full-stack template, and Prisma examples.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -MinimumQualityScore 1.0` passed after M36 evaluation quality gate.
- `powershell -ExecutionPolicy Bypass -File scripts\preflight.ps1 -OutputDir <temp> -TimeoutSeconds 300 -IncludeEvaluation -Proxy http://127.0.0.1:10809 -MinimumEvaluationQualityScore 1.0` passed after M36 preflight quality gate integration.
- `go test ./...` passed after M36 evaluation quality gate changes.
- `go vet ./...` passed after M36 evaluation quality gate changes.
- `git diff --check` passed after M36 evaluation quality gate changes.
- `git check-ignore -v .env eval benchmark dist .repomind` confirmed local secrets and generated outputs are ignored.
- `powershell -ExecutionPolicy Bypass -File scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 3` initially failed because repeated GitHub clone requests were reset by the network.
- Benchmark and evaluation scripts were updated to support shared `RepoCacheDir` reuse.
- `powershell -ExecutionPolicy Bypass -File scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5` passed after M37 shared repo cache changes.
- M37 release gate result: default preflight passed, benchmark passed under 30 seconds for every repository, evaluation quality gate passed with all scores 1.00, and release artifact smoke passed.
- `.github/workflows/release-gate.yml` added after M38 manual GitHub Release Gate workflow.
- `go test ./...` passed after M38 manual GitHub Release Gate workflow changes.
- `go vet ./...` passed after M38 manual GitHub Release Gate workflow changes.
- `git diff --check` passed after M38 manual GitHub Release Gate workflow changes.
- `docs/RELEASE_GATE_RESULTS.md` added after M39 using the successful M37 release gate output.
- `go test ./internal/parser/apiroute` passed after M40 Django include prefix parsing.
- `go test ./...` passed after M40 Django include prefix parsing.
- `go vet ./...` passed after M40 Django include prefix parsing.
- `go test ./internal/parser/apiroute` passed after M41 Go chi route prefix parsing.
- `go test ./...` passed after M41 Go chi route prefix parsing.
- `go vet ./...` passed after M41 Go chi route prefix parsing.
- `docs/ROUTE_PREFIX_STRATEGY.md` added and linked from docs index and parser backlog after M42 cross-file route prefix strategy.
- `.github/workflows/release.yml` updated after M43 to run native binary smoke on Ubuntu, macOS, and Windows before release build/publish.
- `go test ./...` passed after M43 release native smoke matrix changes.
- `go vet ./...` passed after M43 release native smoke matrix changes.
- `git diff --check` passed after M43 release native smoke matrix changes.
- `go test ./internal/parser/apiroute` passed after M44 Django module include prefix parsing.
- `go test ./...` passed after M44 Django module include prefix parsing.
- `go vet ./...` passed after M44 Django module include prefix parsing.
- `powershell -ExecutionPolicy Bypass -File scripts\build-release.ps1 -Version v0.0.0-manifest-test` passed after M45 release manifest changes.
- M45 release manifest contains 6 platform artifacts with size and SHA256 values.
- `go test ./...` passed after M45 release manifest changes.
- `go vet ./...` passed after M45 release manifest changes.
- `git diff --check` passed after M45 release manifest changes.
- `powershell -ExecutionPolicy Bypass -File scripts\verify-release-manifest.ps1 -DistDir dist` passed after M46 manifest verification script.
- `powershell -ExecutionPolicy Bypass -File scripts\preflight.ps1 -OutputDir <temp> -TimeoutSeconds 300 -IncludeManifestBuild -ManifestVersion v0.0.0-preflight-manifest` passed after M47 release gate manifest build integration.
- `go test ./...` passed after M47 release gate manifest build integration.
- `go vet ./...` passed after M47 release gate manifest build integration.
- `git diff --check` passed after M47 release gate manifest build integration.
- `go test ./...` passed after M48 release gate manifest artifact upload changes.
- `go vet ./...` passed after M48 release gate manifest artifact upload changes.
- `git diff --check` passed after M48 release gate manifest artifact upload changes.
- `go test ./internal/parser/apiroute -v` passed after M49 FastAPI imported router prefix changes.
- `go test ./...` passed after M49 FastAPI imported router prefix changes.
- `go vet ./...` passed after M49 FastAPI imported router prefix changes.
- `git diff --check` passed after M49 FastAPI imported router prefix changes.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m50-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0` passed after M50 expanded known checks.
- M50 real evaluation quality scores were 1.00 for Laravel, Spring REST service, Gin examples, FastAPI full-stack template, and Prisma examples.
- `go test ./...` passed after M50 expanded known checks.
- `go vet ./...` passed after M50 expanded known checks.
- `git diff --check` passed after M50 expanded known checks.
- `go test ./internal/parser/apiroute -v` passed after M51 Express relative router import prefix changes.
- `go test ./...` passed after M51 Express relative router import prefix changes.
- `go vet ./...` passed after M51 Express relative router import prefix changes.
- `git diff --check` passed after M51 Express relative router import prefix changes.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m52-resolver-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0` passed after M52 resolver regression evaluation.
- M52 route/model comparison showed no regression from M50: Laravel 1/0, Spring 1/0, Gin 66/0, FastAPI 18/2, Prisma 55/143.
- `git diff --check` passed after M52 resolver regression evaluation docs.
- `go test ./internal/parser/apiroute -v` passed after M53 Go same-package route factory prefix changes.
- `go test ./...` passed after M53 Go same-package route factory prefix changes.
- `go vet ./...` passed after M53 Go same-package route factory prefix changes.
- `git diff --check` passed after M53 Go same-package route factory prefix changes.
- `go test ./internal/parser/apiroute -v` passed after M54 Express composed router prefix changes.
- `go run ./cmd/repomind analyze --output eval\candidate-node-express-realworld-report-m54b eval\candidate-node-express-realworld` passed and produced `/api/tags`, `/api/articles`, and `/api/users/login`.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m54-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0` passed after adding `node-express-realworld`.
- M54 real evaluation quality scores were 1.00 for Laravel, Spring REST service, Gin examples, FastAPI full-stack template, node-express-realworld, and Prisma examples.
- `go test ./...` passed after M54 Express composed router evaluation sample changes.
- `go vet ./...` passed after M54 Express composed router evaluation sample changes.
- `git diff --check` passed after M54 Express composed router evaluation sample changes.
- `go test ./internal/parser/apiroute -v` passed after M55 FastAPI composed router static prefix changes.
- `go run ./cmd/repomind analyze --output eval\fastapi-prefix-m55 eval\release-gate\repo-cache\fastapi-full-stack-template` passed and produced `/api/v1/items`, `/api/v1/users/me`, and `/api/v1/utils/health-check`.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m55-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0` passed after M55 FastAPI mounted prefix changes.
- M55 real evaluation quality scores were 1.00 for Laravel, Spring REST service, Gin examples, FastAPI full-stack template, node-express-realworld, and Prisma examples.
- `go test ./...` passed after M55 FastAPI composed router static prefix changes.
- `go vet ./...` passed after M55 FastAPI composed router static prefix changes.
- `git diff --check` passed after M55 FastAPI composed router static prefix changes.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m56-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0` passed after adding `go-chi`.
- M56 real evaluation quality scores were 1.00 for Laravel, Spring REST service, Gin examples, go-chi, FastAPI full-stack template, node-express-realworld, and Prisma examples.
- `git diff --check` passed after M56 split Go router evaluation sample docs.
- `go test ./internal/detector -v` passed after M57 Chi stack detection changes.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m57-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0` passed after M57 Chi stack detection changes.
- M57 real evaluation confirmed `go-chi` backend is `Chi`.
- `go test ./...` passed after M57 Chi stack detection changes.
- `go vet ./...` passed after M57 Chi stack detection changes.
- `git diff --check` passed after M57 Chi stack detection changes.
- `go test ./internal/parser/apiroute -v` passed after M58 Express multi-line route changes.
- `go run ./cmd/repomind analyze --output eval\express-multiline-m58 eval\release-gate\repo-cache\node-express-realworld` passed and increased routes from 8 to 20.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m58-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0` passed after M58 Express multi-line route changes.
- M58 real evaluation quality scores were 1.00 for Laravel, Spring REST service, Gin examples, go-chi, FastAPI full-stack template, node-express-realworld, and Prisma examples.
- `go test ./...` passed after M58 Express multi-line route changes.
- `go vet ./...` passed after M58 Express multi-line route changes.
- `git diff --check` passed after M58 Express multi-line route changes.
- `go test ./internal/parser/apiroute -v` passed after M59 Go receiver method route factory changes.
- `go run ./cmd/repomind analyze --output eval\go-method-factory-m59 eval\release-gate\repo-cache\go-chi` passed and produced `/users/{id}` and `/todos/{id}/sync`.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m59-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0` passed after M59 Go receiver method route factory changes.
- M59 real evaluation quality scores were 1.00 for Laravel, Spring REST service, Gin examples, go-chi, FastAPI full-stack template, node-express-realworld, and Prisma examples.
- `go test ./...` passed after M59 Go receiver method route factory changes.
- `go vet ./...` passed after M59 Go receiver method route factory changes.
- `git diff --check` passed after M59 Go receiver method route factory changes.
- `powershell -ExecutionPolicy Bypass -File scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5 -RepoCacheDir eval\release-gate\repo-cache` passed after M60 full release gate verification.
- M60 release gate result: `go test`, `go vet`, English smoke, Chinese smoke, benchmark, 7-repository evaluation, release smoke, and release manifest verification all passed.
- `git diff --check` passed after M60 release gate result documentation.
- `go test ./internal/parser/apiroute -v` passed after M61 FastAPI multi-line decorator changes.
- `go run ./cmd/repomind analyze --output eval\fastapi-multiline-m61 eval\release-gate\repo-cache\fastapi-full-stack-template` passed and increased routes from 18 to 23.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m61-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0` passed after M61 FastAPI multi-line decorator changes.
- M61 real evaluation quality scores were 1.00 for Laravel, Spring REST service, Gin examples, go-chi, FastAPI full-stack template, node-express-realworld, and Prisma examples.
- `go test ./...` passed after M61 FastAPI multi-line decorator changes.
- `go vet ./...` passed after M61 FastAPI multi-line decorator changes.
- `git diff --check` passed after M61 FastAPI multi-line decorator changes.
- `go test ./internal/parser/apiroute -v` passed after M62 Django REST Framework router basics.
- `go test ./...` passed after M62 Django REST Framework router basics.
- `go vet ./...` passed after M62 Django REST Framework router basics.
- `git diff --check` passed after M62 Django REST Framework router basics.
- `go test ./internal/repository -v` passed after M63 Git URL input changes.
- `go test ./cmd/repomind -v` passed after M63 CLI Git URL integration test.
- `go test ./...` passed after M63 Git URL input changes.
- `go vet ./...` passed after M63 Git URL input changes.
- `go run ./cmd/repomind analyze --output eval\m63-github-url-smoke --max-files 1000 https://github.com/spring-guides/gs-rest-service.git` passed through `HTTPS_PROXY=http://127.0.0.1:10809`; detected Spring Boot, 59 files, and 1 route.
- `git diff --check` passed after M63 Git URL input changes.
- `git check-ignore -q cmd\repomind\main.go` returned no match after M64 `.gitignore` source directory protection.
- `git check-ignore -v .env .repomind dist eval benchmark` confirmed local secrets and generated outputs remain ignored after M64.
- `go test ./...` passed after M64 `.gitignore` source directory protection.
- `go vet ./...` passed after M64 `.gitignore` source directory protection.
- `git diff --check` passed after M64 `.gitignore` source directory protection.
- `go test ./internal/repository -v` passed after M65 remote Git ref selection.
- `go test ./cmd/repomind -v` passed after M65 CLI `--ref` / `--branch` support.
- `git ls-remote --symref https://github.com/spring-guides/gs-rest-service.git HEAD` through `HTTPS_PROXY=http://127.0.0.1:10809` confirmed default branch `main` for the M65 smoke target.
- `go run ./cmd/repomind analyze --ref main --output eval\m65-github-ref-smoke --max-files 1000 https://github.com/spring-guides/gs-rest-service.git` passed through `HTTPS_PROXY=http://127.0.0.1:10809`; detected Spring Boot, 59 files, and 1 route.
- `go test ./...` passed after M65 remote Git ref selection.
- `go vet ./...` passed after M65 remote Git ref selection.
- `git diff --check` passed after M65 remote Git ref selection.
- `go test ./internal/repository -v` passed after M66 remote Git commit SHA ref support.
- `go test ./cmd/repomind -v` passed after M66 remote Git commit SHA ref support.
- `go run ./cmd/repomind analyze --ref e9efc9dfa0abe8cf8e15cf0e71830b5125322cae --output eval\m66-github-sha-smoke --max-files 1000 https://github.com/spring-guides/gs-rest-service.git` passed through `HTTPS_PROXY=http://127.0.0.1:10809`; detected Spring Boot, 59 files, and 1 route.
- `go test ./...` passed after M66 remote Git commit SHA ref support.
- `go vet ./...` passed after M66 remote Git commit SHA ref support.
- `git diff --check` passed after M66 remote Git commit SHA ref support.
- `go test ./internal/repository -v` passed after M67 optional remote repository clone cache.
- `go test ./cmd/repomind -v` passed after M67 CLI `--repo-cache` support.
- Two consecutive `go run ./cmd/repomind analyze --repo-cache eval\m67-repo-cache --output eval\m67-github-cache-smoke-* --max-files 1000 https://github.com/spring-guides/gs-rest-service.git` runs passed through `HTTPS_PROXY=http://127.0.0.1:10809`; both detected Spring Boot, 59 files, and 1 route.
- `go test ./...` passed after M67 optional remote repository clone cache.
- `go vet ./...` passed after M67 optional remote repository clone cache.
- `git diff --check` passed after M67 optional remote repository clone cache.
- `docs/REMOTE_REPOSITORIES.md` added after M68 remote repository documentation.
- README, docs index, and release checklist now link/check the remote repository documentation.
- `git diff --check` passed after M68 remote repository documentation.
- `go test ./internal/repository -v` passed after M69 remote Git failure hints.
- `go test ./cmd/repomind -v` passed after M69 remote Git failure hints.
- `go test ./...` passed after M69 remote Git failure hints.
- `go vet ./...` passed after M69 remote Git failure hints.
- `git diff --check` passed after M69 remote Git failure hints.
- `go test ./internal/parser/apiroute -v` passed after M70 Go mounted sub-router variable prefix propagation.
- `go run ./cmd/repomind analyze --output eval\m70-go-subrouter-variable eval\release-gate\repo-cache\go-chi` passed and produced 210 Go/Chi routes.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m70-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0` passed after M70.
- M70 real evaluation quality scores were 1.00 for Laravel, Spring REST service, Gin examples, go-chi, FastAPI full-stack template, node-express-realworld, and Prisma examples.
- `go test ./...` passed after M70 Go mounted sub-router variable prefix propagation.
- `go vet ./...` passed after M70 Go mounted sub-router variable prefix propagation.
- `git diff --check` passed after M70 Go mounted sub-router variable prefix propagation.
- `go test ./internal/parser/apiroute -v` passed after M71 Go middleware-wrapped handler names.
- `powershell -ExecutionPolicy Bypass -File scripts\evaluate-repos.ps1 -OutputDir eval\m71-evaluation -TimeoutSeconds 300 -Proxy http://127.0.0.1:10809 -RepoCacheDir eval\release-gate\repo-cache -MinimumQualityScore 1.0` passed after M71.
- M71 real evaluation quality scores were 1.00 for Laravel, Spring REST service, Gin examples, go-chi, FastAPI full-stack template, node-express-realworld, and Prisma examples.
- M71 real evaluation increased `gin-examples` route coverage from 66 to 68.
- `go test ./...` passed after M71 Go middleware-wrapped handler names.
- `go vet ./...` passed after M71 Go middleware-wrapped handler names.
- `git diff --check` passed after M71 Go middleware-wrapped handler names.
- `powershell -ExecutionPolicy Bypass -File scripts\release-gate.ps1 -Proxy http://127.0.0.1:10809 -TimeoutSeconds 300 -CloneRetries 5 -RepoCacheDir eval\release-gate\repo-cache` passed after M72 full release gate verification.
- M72 release gate step results: `go test` 2.30s, `go vet` 1.35s, English smoke 0.25s, Chinese smoke 0.24s, benchmark 2.16s, evaluation 2.63s, release smoke 8.96s, manifest build/verify 12.76s.
- M72 benchmark kept all repositories under the 30-second target.
- M72 evaluation quality scores were 1.00 for all 7 configured real repository samples.
- `docs/RELEASE_GATE_RESULTS.md` updated with the M72 release gate output.
- `git diff --check` passed after M72 release gate result documentation.
- `README.zh-CN.md` added after M73 bilingual README switch.
- `README.md` and `README.zh-CN.md` now link to each other at the top.
- `scripts/build-release.ps1` and `.github/workflows/release.yml` now include `README.zh-CN.md` in release archives.
- `docs/RELEASE_CHECKLIST.md` now checks bilingual README release packaging.
- Local git remote `origin` is set to `https://github.com/patrick892368/RepoMind.git`.
- `powershell -ExecutionPolicy Bypass -File scripts\build-release.ps1 -Version v0.0.0-readme-bilingual -OutputDir dist\bilingual-readme` passed after M73.
- `powershell -ExecutionPolicy Bypass -File scripts\verify-release-manifest.ps1 -DistDir dist\bilingual-readme` passed after M73.
- Release archive content check confirmed `README.zh-CN.md` is present.
- `git diff --check` passed after M73 bilingual README switch.
