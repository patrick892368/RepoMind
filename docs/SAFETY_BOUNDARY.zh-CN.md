# 安全边界

**语言：** [English](SAFETY_BOUNDARY.md) | 简体中文

RepoMind 必须能理解仓库，但不能泄露本地密钥，也不能把生成物提交进仓库。

## 受保护文件

以下路径必须保持 Git ignore 且不能被跟踪：

- `.env`
- `.env.*`，但 `.env.example` 除外
- `.repomind/`
- `dist/`
- `eval/`
- `benchmark/`
- 本地二进制：`repomind`、`repomind.exe`

`.env.example` 可以被跟踪，但只能包含空值或占位符。

## 远程仓库

分析远程 Git URL 时，`analysis.json` 只写入 `repository.remote` 和可选的 `repository.ref`。

RepoMind 不会把原始 remote URL 写入 `analysis.json`，避免将私有仓库 URL 中可能携带的凭证写入报告。

## 验证方式

运行：

```powershell
.\scripts\verify-safety-boundary.ps1
```

该检查会验证：

- 敏感路径和生成物路径已被 Git 忽略；
- 禁止提交的生成物或密钥文件没有被跟踪；
- `.env.example` 仍可被跟踪；
- tracked 文件中没有疑似真实 API key/token 赋值。

默认 preflight、release gate 和 CI 都会运行该安全边界检查。
