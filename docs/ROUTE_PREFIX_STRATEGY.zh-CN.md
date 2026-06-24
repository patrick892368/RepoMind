# Route Prefix 策略

**语言：** [English](ROUTE_PREFIX_STRATEGY.md) | 简体中文

RepoMind 已支持多种同文件 route prefix 模式。跨文件 route prefix 传播必须谨慎增加，因为它很容易把轻量 scanner 变成部分 language server。

## 当前同文件覆盖

- FastAPI `APIRouter(prefix=...)` 和 same-file `include_router(router, prefix=...)`。
- FastAPI direct imported router prefix。
- FastAPI composed router prefix 和静态 settings 常量。
- Express `app.use("/prefix", router)` 和 same-file router methods。
- Express relative router import prefix。
- Express composed router prefix。
- Django `path("prefix/", include(patterns))` 和 local pattern lists。
- Django `path("prefix/", include("module.urls"))` module include prefix。
- Go chi-style `Route("/prefix", func(...))`。
- Go same-block mounted sub-router variable prefix。
- Go same-package route factory prefix。
- Go same-package receiver method route factory prefix。
- NestJS controller prefix 和 method decorators。
- Spring class-level `@RequestMapping` 和 method mappings。

## 问题

很多仓库会把 route 拆到多个文件：

- FastAPI 在 `app/api/routes/users.py` 定义 router，再由 `app/main.py` include。
- Express 在 `routes/order.js` 定义 router，再由 `app.js` mount。
- Django project-level `urls.py` include `orders.urls`。
- Go 在 helper function 中创建 subrouter，再在其他地方 mount。

parser 可以检测 child routes，但可能漏掉 parent prefix。

## 设计目标

- 保持 30 秒目标。
- 避免完整语言类型解析。
- 输出必须有 evidence 和 confidence。
- 优先做高置信 prefix propagation。
- 无法解析 parent prefix 时，仍保留 child route fallback。

## 内部解析结构

建议在最终 `ir.APIRoute` 输出前维护内部结构：

```txt
RouteFragment
  method
  path
  handler
  file
  line
  source
  receiver/router name
  exported symbol/name

RouteMount
  prefix
  target symbol/import/module
  file
  line
  source
  confidence
```

最终解析：

```txt
RouteMount(prefix="/api", target="ordersRouter")
RouteFragment(receiver="ordersRouter", path="/create")
=> /api/create
```

## 解析规则

只解析高置信场景：

- 静态字符串路径。
- 同 package/module 的直接 import。
- 单一 exported router symbol。
- 直接 identifier match。
- Django `include("orders.urls")` 这类明确 module include。
- 不解析动态字符串拼接。
- 不猜测有 runtime arguments 的 router factories。

如果多个目标都可能匹配，保留原始 route，不生成猜测结果。

## 语言计划

Python / Django / FastAPI：

- `include("module.urls")`。
- exported `router = APIRouter(...)`。
- imported router symbols。
- module router references。
- 唯一静态 prefix 常量。

JavaScript / TypeScript / Express：

- `module.exports = router`、`export default router`、`export const router`。
- `require("./routes/orders")` 和 default import。
- 相对 import。
- dotted basename。
- 第一版跳过多层 alias chains。

Go：

- 同 package 内明确返回 router 的函数。
- 同 block sub-router variable。
- 跨 package route assembly 暂缓。

## Rollout

1. FastAPI imported router prefix。Done。
2. Express relative router import prefix。Done。
3. Go same-package route factory prefix。Done。
4. Go same-block mounted sub-router variable prefix。Done。
