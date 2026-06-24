# Route Prefix Strategy

RepoMind currently supports several same-file route prefix patterns. Cross-file route prefix propagation should be added carefully because it can quickly turn a lightweight scanner into a partial language server.

## Current Same-file Coverage

- FastAPI `APIRouter(prefix=...)` plus same-file `include_router(router, prefix=...)`.
- FastAPI imported router prefix for direct static imports such as `from app.api.routes.users import router as users_router`.
- FastAPI composed router prefix for `items.router` module imports and static settings constants such as `settings.API_V1_STR`.
- Express `app.use("/prefix", router)` plus same-file router methods.
- Express relative router import prefix for direct `require("./routes/orders")` and `import ordersRouter from "./routes/orders"` mounts.
- Express composed router prefix for `Router().use("/api", api)` wrapping imported controller routers.
- Django `path("prefix/", include(patterns))` plus same-file local pattern lists.
- Django `path("prefix/", include("module.urls"))` module include prefix propagation.
- Go chi-style `Route("/prefix", func(...))` nested route calls.
- Go same-block mounted sub-router variable prefix propagation.
- Go same-package route factory prefix for direct `Mount("/prefix", orderRoutes())` calls.
- Go same-package receiver method route factory prefix for `Mount("/prefix", resource{}.Routes())` calls.
- NestJS controller prefix plus method decorators.
- Spring class-level `@RequestMapping` plus method mappings.

## Problem

Many repositories split route definitions across files:

- FastAPI defines routers in `app/api/routes/users.py`, then includes them from `app/main.py`.
- Express defines `orderRouter` in `routes/order.js`, then mounts it from `app.js`.
- Django includes `orders.urls` from a project-level `urls.py`.
- Go creates subrouters in helper functions and mounts them elsewhere.

The current parser can detect child routes but often misses the parent prefix when the mount is in another file.

## Design Goals

- Preserve the 30-second target.
- Avoid full language type resolution.
- Keep output evidence and confidence explicit.
- Prefer partial, high-confidence prefix propagation over broad guesses.
- Keep fallback behavior: a child route without a resolved parent prefix should still be reported.

## Proposed IR Extension

Add internal-only route assembly structures before final `ir.APIRoute` output:

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

The final analyzer can resolve:

```txt
RouteMount(prefix="/api", target="ordersRouter")
RouteFragment(receiver="ordersRouter", path="/create")
=> /api/create
```

## Resolution Rules

Only resolve high-confidence cases:

- Same package/module imports with static string paths.
- Single exported router symbol per file.
- Direct identifier matches.
- Direct module include matches, such as Django `include("orders.urls")`.
- No dynamic string concatenation.
- No runtime router factories unless the factory return symbol is obvious.

If more than one target matches, keep the original route and record a low-confidence scan note instead of guessing.

## Language Plan

### Python / Django / FastAPI

- Detect `include("module.urls")`.
- Map module path to local `urls.py`.
- Detect exported `router = APIRouter(...)`.
- Detect imported router symbols: `from app.api.routes.users import router as users_router`.
- Detect module router references: `from app.api.routes import users` plus `include_router(users.router)`.
- Resolve static string constants used as include prefixes when the constant name is unique in the scanned repository.
- Resolve only static imports inside repository root.

### JavaScript / TypeScript / Express

- Detect `module.exports = router`, `export default router`, and `export const router`.
- Detect `require("./routes/orders")` and `import ordersRouter from "./routes/orders"`.
- Resolve relative imports only.
- Resolve dotted basenames such as `./tag/tag.controller` to `.ts` or `.js` files.
- Skip alias chains beyond one hop in the first version.

### Go

- Detect functions returning routers only when the function body is in the same package.
- Resolve route prefixes across files in the same Go package.
- Skip cross-package route assembly until package graph support is stronger.

## Testing Plan

For every cross-file resolver:

- Add a focused fixture for the framework.
- Assert both prefixed and unprefixed fallback behavior.
- Assert ambiguous mounts do not create guessed routes.
- Assert evidence points to the child route line and includes confidence.
- Run real repository evaluation after each language.

## Rollout Order

1. FastAPI imported router prefix. Done.
2. Express relative router import prefix. Done.
3. Go same-package route factory prefix. Done.
4. Go same-block mounted sub-router variable prefix. Done.
