# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# CandyPro OEM B2B Platform

## Stack
- **Backend**: Go 1.24.7, Gin 1.10, GORM, PostgreSQL 15, pgvector. **Tests run on `glebarez/sqlite` (in-memory), not Postgres** — PG-only SQL won't be exercised by tests.
- **Frontend**: Nuxt 3.21 (installed; declared `^3.15.0`), Vue 3.5, TypeScript 5.6, Tailwind, `@nuxtjs/i18n` v10.3 (zh default)
- **AI**: Cloudwego Eino v0.7 + OpenAI-compatible models (default `deepseek-v4-flash`), pgvector
- **Module path**: `candypro/api`

## Commands

### Makefile
| Command | Description |
|---------|-------------|
| `make dev` | Backend dev server (:8080) |
| `make dev-frontend` | Frontend dev server (:3000) |
| `make dev-all` | Both backend + frontend (`-j2`) |
| `make build-all` / `make build` / `make build-frontend` | Build backend / backend / frontend |
| `make test` | Backend tests (`go test -v ./...`) |
| `make test-coverage` | Backend tests + HTML coverage |
| `make test-frontend` | Frontend `npm run test` (Playwright e2e) |
| `make lint` | golangci-lint |
| `make vet` / `make fmt` | `go vet` / `go fmt` |
| `make seed` / `-business` / `-demo` / `-essential` | Seed via `cmd/seed` |
| `make reset-db` | Drop + recreate + reseed (`cmd/seed`, NOT `make migrate` — that blocks on the API server) |
| `make swagger` | `swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal` |
| `make prod-start` / `prod-stop` | Merged `docker-compose.yml` + `docker-compose.prod.yml` |
| `make install` / `install-tools` | `go mod download` + `npm install` / install swag + golangci-lint |

### Frontend scripts (`frontend/package.json`)
- `npm run dev` — dev server; `npm run dev:clean` — clean `.nuxt` then dev
- `npm run build` — `nuxt build` (production)
- `npm run typecheck` — `nuxt prepare && vue-tsc --noEmit` (the only local static gate)
- `npm test` / `npm run test:e2e` — Playwright; `test:e2e:ui` — headed
- `postinstall` runs `scripts/patch-nuxt-og-image.mjs` + `nuxt prepare` — always run `npm ci`/`install` before first dev/build

### Single backend test
```bash
cd backend && go test ./internal/repository/order/ -run TestOrderStateTransition -v
```

## Directory Layout
```
backend/
├── cmd/api/main.go              # Entry point: config, db, migrations, background workers, server
├── cmd/seed/                    # Manual seed CLI (also runs AutoMigrate)
├── internal/
│   ├── api/
│   │   ├── router.go            # Gin router, middleware, CORS, rate limiting, /metrics
│   │   └── routes/              # Route registration per scope (adminportal/, userportal/, authscope/, public/, system/)
│   ├── config/                  # Env loading + config structs (string-typed booleans, see Gotchas)
│   ├── database/
│   │   ├── database.go          # GORM connection, AutoMigrate (Phase1 + Phase2), EnsureStartupData
│   │   ├── seed.go / seed_demo.go / seed_translations.go
│   │   ├── migrate_*.go         # Hand-written migrations for drops/renames/backfills (see Conventions)
│   │   └── seed_locales/        # en.yaml + zh.yaml (initial DB i18n seed data, not embedded)
│   ├── handlers/                # HTTP handlers per scope (handlers.go = entry)
│   │   ├── admin/  customer/  auth/  public/  system/
│   ├── middleware/               # Auth, role, rate-limit, CORS, recovery, request-id, locale, metrics
│   ├── models/                  # GORM models per domain (order/, product/, trade/, user/, auth/, common/)
│   ├── pkg/                     # Shared, reusable packages (no domain logic)
│   │   ├── eino/                # Eino agents (agent.go, deep_b2b.go, plan_order.go, graph/)
│   │   ├── i18n/                # DB-backed translation engine (T(), Translate(), WarmCache())
│   │   ├── workerlock/          # Redis SET NX + Lua CAS distributed lock for workers
│   │   ├── trade/               # AI trade document processing (classifier, extractor, validator, HITL)
│   │   ├── payment/             # Payment gateway interface + stripe adapter
│   │   └── ...                  # crypto, dberror, jwtutil, kyb, money, pagination, password, response, storage, timeutil, uploadpath, valerr, docxgen
│   ├── repository/              # Data access layer
│   │   ├── scopes/<scope>/repositories.go   # Per-scope repo aggregator structs
│   │   └── */                   # Individual repositories (order/, product/, trade/, user/, translation/, common/)
│   └── services/                # Business logic layer
│       ├── scopes/<scope>/services.go       # Per-scope service aggregator structs
│       └── */                   # Individual services (order/, product/, trade/, user/, translation/)
└── docs/                        # Swagger docs (committed)
frontend/
├── nuxt.config.ts               # Nuxt 3, i18n (zh default), Tailwind, sitemap, CSP with pinned SHA-256
├── app.vue                      # Root: hides marketing chrome for admin/customer/auth routes
├── pages/                       # File-based routing (~100 pages: public, auth, admin, customer)
├── components/                  # Auto-imported (pathPrefix: false), 42 components
├── composables/                 # useApi, useAuth, useInquiry, useSanitizer, useDisplay, useSeo, ...
├── middleware/auth.ts           # Global route guard
├── plugins/init-auth.client.ts  # Auth init on app start
├── layouts/                     # default.vue (marketing), admin.vue, customer.vue, auth.vue
├── i18n/<locale>/               # JSON namespaces per locale — 9 locales, 15 namespaces each
├── e2e/                         # Playwright specs (a11y, auth, legal, locale, negotiation, order)
└── assets/css/main.css          # "Industrial Confectionery" design system
```

## Architecture Pattern

**Handler → Service → Repository** (three-layer), scoped by portal:

1. `repository/scopes/<scope>/repositories.go` — Aggregates all repositories into one struct
2. `services/scopes/<scope>/services.go` — Aggregates all services into one struct
3. `handlers/<scope>/` — Handlers hold a `services` field (the scope's services struct)

Each scope has its own Repositories and Services aggregator. The root `handlers.go` wires everything:
```go
repos := repositoryCommon.NewRepositories(db)    // creates all scope repo structs
svcs  := servicesCommon.NewServices(repos, ...)  // creates all scope service structs
h     := handlers.New(cfg, svcs)                  // creates all scope handler structs
```

Services typically define a local unexported interface for their repository dependency (e.g. `type cartRepository interface{...}`).

~430 API endpoints total across 5 scopes; ~100 frontend pages.

## Key Scopes & Route Prefixes

| Scope | Route Prefix | Auth | Middleware |
|-------|-------------|------|------------|
| Public | `/api/v1/public` | No | Rate limit on inquiry |
| Auth | `/api/v1/auth` | Mixed | Login rate limiter |
| User | `/api/v1/user` | JWT | ActiveGuard for write ops |
| Admin | `/api/v1/admin` | JWT | RequireRole(admin, superadmin) |
| System | `/api/v1/system` | Mixed | Public AI rate limit |

## Database

- GORM AutoMigrate runs in two phases:
  - **Phase 1**: FK-independent tables (roles, categories, price_lists, translations, compliance, etc.)
  - **Phase 2**: FK-dependent tables (users, products, orders, payments, trades, etc.)
- **AutoMigrate only adds — it never drops/renames.** Any migration that must drop or rename a column/index on an already-migrated DB goes in a hand-written `database/migrate_*.go` file wired into `EnsureStartupData` (`database.go`). There is no versioned migration runner.
- pgvector extension enabled for semantic search (falls back to keyword-only if unavailable)
- `SeedEssential` always runs on startup (roles + superadmin)
- `SeedTranslations` runs on first startup — reads YAML files from `database/seed_locales/` into the translations table
- i18n engine (`pkg/i18n`) loads all translations from DB into memory after seeding; admin edits refresh the in-memory cache immediately
- `AUTO_SEED_DATA=true` env var triggers full business data seeding

## Background Workers

All workers are wrapped in `pkg/workerlock.WithLock` (Redis SET NX EX + Lua CAS; `NoopLocker` fallback when Redis is off) for cross-instance mutual exclusion. They are started in `cmd/api/main.go`:

1. **Order Draft Cleanup** — Cancels expired `pending_confirmation` orders, releases stock. Config: `ORDER_DRAFT_CLEANUP_INTERVAL_MINUTES`, `ORDER_DRAFT_EXPIRE_MINUTES`, `ORDER_DRAFT_CLEANUP_BATCH_SIZE`
2. **Event Outbox Relay** — Polls event outbox for confirmed orders, creates Trade transactions async. Config: `EVENT_OUTBOX_RELAY_INTERVAL_SECONDS`
3. **Abandoned-pending cleanup** and **checkpoint cleanup** — similar periodic loops

## Conventions (non-obvious, enforced in review)

- **State transitions must be guarded**: `UPDATE … WHERE id=? AND status=<expected>` + check `RowsAffected == 0` → return a state-mismatch error (20 files across `repository/` and `services/` use a `RowsAffected` guard: payment `UpdateStatus`, shipment `Dispatch`, fulfillment, etc.). Do not write a new state-changing endpoint without this guard.
- **Tenant isolation is per-handler today, not structural.** Repositories for order/payment/invoice/shipment/return/requisition are *not* scoped by `user_id`/`company_id`; handlers must call `contextUserID(c)` / `c.GetString("userID")` and ownership-check. New handlers must keep doing this.
- **Money is `float64`** throughout (order, payment, invoice, coupon, trade, price, credit-limit). No decimal type yet — never compare money with `==`; always round via `pkg/money` formatting.
- **Eino tool-output filtering is proven only for `ChatModelAgent` react loops** (`Client.Generate`); graph-based agents (B2B coordinator, plan-execute-replan) are a separate path. New state-changing tools must advertise full param schemas (e.g. `port_of_loading`/`port_of_destination` are still missing from `graph/tool.go`) or the LLM sends nothing.
- **Raw SQL in analytics/reports uses PG-only constructs** (`GREATEST`, `date_trunc`, `EXTRACT`, `::int`/`::numeric`/`::jsonb` casts). These cannot run on the sqlite test harness — prefer GORM-safe constructs or add a source-text regression test.
- **Embeddings are write + backfill only** (`FindProductsMissingEmbeddings`); there is no invalidation-on-update, so semantic search drifts on product edits.
- **`v-html` must route through `useSanitizer`** — one shared allowlist feeds both server `sanitize-html` and client DOMPurify; bypassing it reintroduces XSS and hydration mismatches.

## Common Patterns (Backend)

### Error responses
```go
response.ErrorResp(c, http.StatusNotFound, "not_found")
response.InvalidResp(c, "invalid_request")
response.BindJSONOrInvalid(c, &req) // returns false + writes 400 if bind fails
```

### i18n translation
```go
i18n.T(c, "errors.not_found")              // uses locale from gin.Context
i18n.TWithVars(c, "errors.account_status", map[string]string{"status": user.Status})
```

### Pagination
```go
page, limit := pagination.ParsePagination(c, 20, 100)  // ParsePagination clamps limit (silently) — see list-clamp risk
pagination.BuildPagination(total, page, limit)
```

### User ID extraction
```go
// Admin handlers: c.GetString("userID")
// Customer handlers: userID, ok := contextUserID(c)
```

### Role constants
```go
modelsAuth.User       // "customer"
modelsAuth.Admin      // "admin"
modelsAuth.SuperAdmin // "superadmin"
modelsAuth.UserPortal()   // returns []string{"customer"}
modelsAuth.AdminPortal()  // returns []string{"admin", "superadmin"}
```

### Path param as uint
```go
id, err := parseUintParam(c, "id")
```

## Frontend Conventions

- Nuxt 3 file-based routing; components auto-imported with `pathPrefix: false`
- i18n: Chinese (`zh`) is default locale, strategy is `prefix_except_default`; frontend JSON lives in `frontend/i18n/<locale>/` (9 locales, 15 namespaces)
- Auth: JWT in cookies (`auth_token`, `refresh_token`), auto-refresh via `initAuth()` plugin
- All API calls go through `composables/useApi.ts` — a single monolith that owns SSR/CSR base-URL branching, error normalization, refresh-token coalescing, and idempotency keys. Keep changes there centralized and tested.
- Layouts: `default.vue` (marketing chrome), `admin.vue`, `customer.vue`, `auth.vue`; `app.vue` hides marketing header/footer for admin/customer/auth routes
- CSP is pinned: `script-src` has a SHA-256 hash of the inline theme script (`nuxt.config.ts`). Any new inline script must regenerate that hash, or the build/policy breaks.
- List pages hit the backend's `limit` clamp (default 200) — a page that needs more must page through the API, not assume the clamp doesn't apply.

## Key Env Variables

```
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE        # Postgres
REDIS_URL                                                          # empty => Postgres-backed fallback (dev); set + Redis down => startup fails
JWT_SECRET, JWT_ACCESS_MINUTES, JWT_REFRESH_DAYS
OPENAI_API_KEY, OPENAI_MODEL
AUTO_SEED_DATA, SEED_SUPERADMIN_EMAIL, SEED_SUPERADMIN_PASSWORD
ENABLE_SWAGGER (default "true" — disable in prod), TRUSTED_PROXIES, CORS_ALLOWED_ORIGINS
UPLOAD_DRIVER (local/s3), FRONTEND_URL
```

## Operational Gotchas

- **`REDIS_URL` set + Redis down ⇒ backend refuses to start.** Run with `REDIS_URL=` (empty) to use the Postgres fallback for dev; `pkg/workerlock` degrades to `NoopLocker`.
- **Booleans are parsed as strings** (`getEnv(...) == "true"` in `internal/config/`) — `TRUE`, `1`, `yes` all silently parse as false. Use lowercase `true`.
- **`go test ./...` can run stale compiled test binaries.** If a test passes alone but fails under `./...` with a pre-fix symptom, run `go clean -cache` and re-run.
- **`backend/.env` and `frontend/.env` are gitignored** — don't commit secrets. `docs/` is also gitignored; tracked changelog entries are force-added (`git add -f`).
- **`make reset-db`**: do NOT follow with `make migrate` — it starts the API server and blocks. Use `cmd/seed` (which runs AutoMigrate first).
- Swagger UI at `/swagger/*` (on by default); `/health` and `/ready` for health checks; `/metrics` (Prometheus) is behind `MetricsAuth`.
