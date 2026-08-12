# CandyPro — Future Problems (Backend & Frontend Risk Register)

**Last updated:** 2026-08-12
**Context:** This is a *forward-looking* risk register, not a bug list. It captures the problems that will bite as the project grows — drawn from the 2026-08-11 full code-review audit (`docs/reports/code-review-audit-2026-08-11.md`), the fix pass that closed 42 finding groups (see `docs/changelog/`), the documented follow-ups, and the codebase's underlying design. It answers the question: *"what breaks next, and why?"*

**Relationship to other docs:** `docs/planning/todo.md` / `plan.md` are retrospective fix lists (all closed). This doc is the set of *latent* problems that remain or will recur. Each item names the future failure mode and a mitigation.

---

## Table of contents

- [Backend future problems](#backend-future-problems)
- [Frontend future problems](#frontend-future-problems)
- [Cross-cutting future problems](#cross-cutting-future-problems)
- [Priority order — fix these first](#priority-order--fix-these-first)

---

## Backend future problems

### B1. Money is `float64` end-to-end (the ticking bomb)
**What:** Order, payment, invoice, coupon, trade, price, and credit-limit fields are all `float64`. Only two concrete defects were fixed (FormatMoney rounding, lineage-hash collisions); the type is unchanged.
**Why it bites:** Rounding drift on large B2B amounts; `==` float comparisons silently fail; COGS / credit / partial-payment drift compounds with volume. A real-currency reconciliation or partial-payment feature will surface a "money doesn't add up" bug that is very hard to trace.
**Mitigation:** Migrate money fields to `shopspring/decimal` (or store integer minor-units) before building more financial features. This is a large, deliberate migration — schedule it.

### B2. No row-lock / optimistic-lock *convention* — only spot-fixes
**What:** Locks were added to the worst paths (shipment dispatch, PO receive, payment balance-check, fulfillment ship), but there is no documented pattern.
**Why it bites:** Every future state-transition endpoint (returns, refunds, fulfillments, approvals, order edits) is written fresh and the double-submit / double-deduct race reappears. The audit found 4+ such races already.
**Mitigation:** Document a "state transition must be `UPDATE … WHERE id=? AND status=<expected>` + `RowsAffected==0` guard" convention, and enforce it in code review.

### B3. Tenant isolation lives in handlers, not repositories
**What:** Order/Payment/Invoice/Shipment/Return/Requisition repositories are unscoped by `user_id`/`company_id`; isolation is done by handler ownership checks.
**Why it bites:** Every new handler must remember to scope by user/company — and future ones will forget, exactly as the current ones did (4 IDORs were fixed in the audit pass). One missed check = cross-tenant data exposure.
**Mitigation:** Add company/user scoping to the repository APIs (or a scoped-repo wrapper per portal) so isolation is structural, not a per-handler habit.

### B4. AutoMigrate never drops/renames — schema drift is permanent
**What:** GORM `AutoMigrate` only adds; it never drops removed indexes/columns on an already-migrated DB (this is why the legacy email/slug unique indexes had to be dropped by hand).
**Why it bites:** Every future migration touching an existing table needs a hand-written `database/migrate_*.go`, or dev and prod schemas silently diverge. Soft-delete + unique-index (email/slug/SKU/doc_number) is the recurring trap.
**Mitigation:** Establish a migration convention (hand-written, versioned `migrate_*.go` files) and a startup migration runner; document the soft-delete/unique-index pattern for new models.

### B5. The Eino AI layer is the most fragile code
**What:** The tool-content filtering in `Client.Generate` and the trailing-`Tool`-event predicate are proven correct only for `ChatModelAgent` react loops; the HITL `RequiresReview` override can still re-route a tool to a dead-end interrupt; `graph/tool.go` does not advertise `port_of_loading`/`port_of_destination` so the LLM sends nothing.
**Why it bites:** Wiring a graph-based agent, a new state-changing tool, or a new AI document type can silently reintroduce raw tool output into answers, dead-end a tool, or persist a nonsense port. The recent doc-number/CI-pol fixes are one wrong abstraction away from regressing.
**Mitigation:** Add the missing tool-schema params; document the ChatModelAgent-only invariant; add a graph-agent guard test before any graph-based agent is wired.

### B6. Postgres-only SQL vs the sqlite test harness
**What:** Tests run on `glebarez/sqlite` (in-memory); production is PostgreSQL 15. `GREATEST` was a real portability bug; the analytics/report raw SQL already drifted on `deleted_at IS NULL`.
**Why it bites:** Future raw SQL using PG-only functions (locks, window functions, casts, `GREATEST`/`LEAST`/`DISTINCT ON`) will compile, pass on sqlite, and fail — or behave differently — in production. Tests give false confidence.
**Mitigation:** A lint rule / review gate for PG-only SQL in test-covered paths; prefer GORM-safe constructs; keep a Postgres-backed integration test target.

### B7. Background workers have no systemic idempotency
**What:** The outbox relay, draft-cleanup, payment lock, and workerlock were hardened individually, but there is no shared at-least-once / idempotency convention.
**Why it bites:** Future batch jobs (report generation, syncs, notifications, cache rebuilds) risk double-processing or lost audit rows; the audit already found one worker losing audit rows (fixed).
**Mitigation:** A shared job-dedup / idempotency helper (event key, `processed_at`, conditional update) reused by all workers.

### B8. Observability is essentially absent
**What:** `/ready` now checks the DB, but there are no Prometheus metrics, no structured request tracing, no audit-log completeness verification.
**Why it bites:** When a production incident happens, debugging is logs-only across a 3-tier stack with no way to correlate a single request or measure the system.
**Mitigation:** Add `/metrics` (Prometheus), request-ID/trace propagation (the middleware already sets request-id), and an audit-log completeness check.

### B9. pgvector embeddings freshness
**What:** Product embeddings are now written + backfilled, but there is no invalidation-on-update mechanism.
**Why it bites:** As products change, semantic search drifts from the catalog silently — a slow correctness decay users notice as "search returns wrong stuff."
**Mitigation:** Re-embed on product update (hook into the update path) or a periodic reconciliation job; add a drift check.

### B10. Secrets committed and loose config parsing
**What:** `backend/.env` (live DeepSeek key + superadmin password) is in the repo; booleans are parsed as strings (`== "true"`); `ENABLE_SWAGGER` defaults true in production.
**Why it bites:** Credential leak risk on any repo share/mirror; config values silently misparse (uppercase `TRUE`/`1` → false); swagger exposed in prod.
**Mitigation:** Gitignore `.env` and rotate the leaked keys; centralize typed config parsing; flip prod security defaults.

---

## Frontend future problems

### F1. No shared contract layer — perpetual DTO drift (the #1 frontend problem)
**What:** No OpenAPI codegen; pages hand-write field names against backend responses.
**Why it bites:** Every future backend response-shape change silently breaks a page — no compile error, just blank/0/400 at runtime. The audit pass fixed 8+ mismatches (invoices, analytics, webhooks, OEM, compliance, statuses); more will accumulate.
**Mitigation:** Generate a typed API client from the Swagger spec (`backend/docs/swagger.json` is now committed) and use it across pages.

### F2. SSR sanitization is a two-path system that must never fork
**What:** Server renders via `sanitize-html` (new dependency), client via DOMPurify, sharing an allowlist.
**Why it bites:** They format edge cases slightly differently → hydration mismatches; every new `v-html` sink (AI-content page, CMS block) must remember both paths or the XSS surface (H16) returns.
**Mitigation:** A lint/guard for `v-html` usage (must route through `useSanitizer`); keep the two configs derived from one allowlist.

### F3. The CSP is fragile by construction
**What:** `script-src` without `unsafe-inline` + a SHA-256-pinned inline theme script + `connect-src` derived from the API origin.
**Why it bites:** Any future inline script, or a new absolute API origin, breaks the policy or the build silently. The theme-cookie regex residual can also regress first-paint dark mode unnoticed.
**Mitigation:** Add a CI check that regenerates/validates the CSP hash; fix the theme-cookie regex; document "no inline scripts without updating the hash."

### F4. Status enums and locale keys duplicated in three places
**What:** Backend enums, frontend enum maps, and DB-backed translation keys must all agree.
**Why it bites:** Every new status adds a three-way sync obligation; the failure mode is silent "Unknown"/blank rendering (already true for `pending_confirmation`/`confirmed` inquiry statuses).
**Mitigation:** A single source of truth for enums (generate the frontend enum + locale keys from the backend), plus a missing-key check.

### F5. The list-clamp pattern will keep biting
**What:** `limit:200` silently clamped by `ParsePagination` appeared in blog, cases, and two product pickers; the fix (fetch-all-pages) is fragile for large catalogs (N-parallel requests, page cap).
**Why it bites:** Future list pages will silently truncate again; the fetch-all approach scales poorly.
**Mitigation:** A shared paged-fetch helper (fetch pages up to `totalPages` at the real clamp size) as the standard for list pages; revisit the 5000-SKU cap.

### F6. 100 pages, 42 components, zero component tests
**What:** Typecheck is the only gate; ~500 a11y items (input labels, aria) remain; no component-level tests; the a11y axe gate covers public routes only.
**Why it bites:** Regressions surface at runtime / user reports, not at build; admin/customer contrast and accessibility are unverified.
**Mitigation:** Component tests for high-risk shared components (ProductCard, forms, tables); extend the axe gate to admin/customer routes.

### F7. Auth/session complexity is centralized in one composable
**What:** `useApi` handles SSR/CSR base-URL branching, error normalization, refresh-token coalescing, and the token lifecycle.
**Why it bites:** Future auth changes (remember-me, refresh rotation, multi-tab) all funnel through it; its edge cases are the highest-risk frontend surface.
**Mitigation:** Keep it thin; extract SSR/CSR URL resolution and error normalization into separate tested modules.

### F8. Nuxt version drift
**What:** Running Nuxt 3.21 against a `^3.15` declaration; `nuxt-seo-utils` is incompatible and disabled.
**Why it bites:** A Nuxt 4 upgrade will break plugins, middleware, and config in ways typecheck won't fully catch.
**Mitigation:** Pin the actual version in `package.json`; schedule a deliberate Nuxt 4 upgrade with an e2e smoke pass.

---

## Cross-cutting future problems

### C1. No CI in git — no automated safety net
**What:** `.github/` (CI, CODEOWNERS, PR template) is gitignored and untracked.
**Why it bites:** Regressions ship silently; nothing re-runs `go test ./...` + typecheck + a11y on every change.
**Mitigation:** Commit the CI pipeline (`go build/vet/test`, `npm typecheck`, `nuxt build`, axe gate) — cheap and the single highest-leverage fix.

### C2. No deployment configured
**What:** Docker compose files exist and `make prod-start` is fixed, but there is no host, domain, HTTPS reverse proxy, or deploy pipeline.
**Why it bites:** All the above problems stay invisible until someone tries to run prod, then several surface at once (secrets, PG-only SQL, SSR base URL, missing metrics).
**Mitigation:** A deployment plan: Postgres + backend + frontend behind a reverse proxy with `BACKEND_URL` set; a smoke-test after first deploy.

---

## Priority order — fix these first

1. **C1 — commit the CI pipeline** (cheapest gate against every other regression).
2. **F1 — OpenAPI-typed API client** (stops the perpetual frontend/backend drift).
3. **B1 — money → decimal** (the correctness bomb; schedule the migration).
4. **B3 — structural tenant scoping** (the security invariant is per-handler today).
5. **B5 — Eino tool-schema params + invariant docs** (the most fragile code, most likely to regress).
6. **F4 — single-source-of-truth enums/locale keys** (silent "Unknown" rendering).
7. **B4 — migration convention** (schema drift is permanent once it happens).
8. **B2 — state-transition lock convention** (prevents the next race).
9. **B8 — observability** (needed before you can debug production).
10. **C2 — deployment plan** (to actually serve online).
