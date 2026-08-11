# Changelog / Devlog — 2026-08-11 H5 Cost-Stack Leak Fix
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** Finding H5 — `/system/generate-quotation` exposed the internal cost stack (logistics/duty/label/compliance cost + target gross margin) to any authenticated user.
**Result:** Route-level role guard (admin/superadmin) + handler-level defense-in-depth gating; regression test added; build and targeted tests green.
---
## 1. Process
1. Read `backend/internal/api/routes/system/register.go` and `backend/internal/handlers/system/ai.go`; confirmed the route sat under the auth-only `systemProtected` group (only `AuthMiddleware`) while `GenerateQuotation` echoes raw `ProductMarketCostStack` (`costStackJSON`) into `pricingAssist.summaries`.
2. Verified both frontend callers are admin-portal tools (`frontend/pages/admin/inquiries/[id].vue` and `frontend/pages/admin/ai/index.vue`), so restricting to admin/superadmin breaks nothing.
3. Implemented the guard (route) plus defense-in-depth (handler), added a routing-layer regression test, and verified.
## 2. Fixes in detail
#### H5 — `/system/generate-quotation` leaks internal cost stack to any authenticated user
- **Problem:** `backend/internal/api/routes/system/register.go:38` registered `POST /generate-quotation` under `systemProtected`, which applied only `middleware.AuthMiddleware` (no role check). `backend/internal/handlers/system/ai.go:191-222` fetched `ProductMarketCostStack` (logistics cost, duty rate, label/compliance amortization, target gross margin) via `BuildPricingCostContextJSON` and echoed it in the response `pricingAssist.summaries`. Any customer JWT could therefore read internal margin/cost structure.
- **Fix:**
  - `backend/internal/api/routes/system/register.go:41` — added `middleware.RequireRole(modelsAuth.AdminPortal()...)` to the `/generate-quotation` route (imported `modelsAuth`), matching the admin-portal guard pattern; admin/superadmin behavior is unchanged. Guard placed on the single route only, not the whole `systemProtected` group, so customer-facing protected routes (`translate`, `chatbot/order`, `conversations/:id`, etc.) are unaffected. Also normalized a pre-existing stray-tab indent on `/analyze-inquiry` so the file is gofmt-clean.
  - `backend/internal/handlers/system/ai.go:174-182` — defense-in-depth: cost-stack fetching, AI injection, and `summaries` echo are now gated behind `role == admin || role == superadmin` (via existing `authContext` helper). Non-admin callers receive `pricingAssist.injectedProducts: 0` with no cost data; admins keep identical behavior.
  - `backend/internal/api/routes/system/register_test.go` (new) — regression test minting customer/supplier/admin/superadmin JWTs via `jwtutil.GenerateJWT` against the real `Register` wiring: no token → 401, customer/supplier → 403, admin/superadmin → pass the guard.
## 3. Verification
| Check | Result |
|-------|--------|
| `go build ./...` (backend) | PASS (exit 0) |
| `REDIS_URL= go test ./internal/api/routes/system/... ./internal/handlers/system/...` | PASS (`ok` both; includes new `TestGenerateQuotationRoleGuard` with 5 subcases) |
| `gofmt -l` on owned files | clean |
| Frontend callers audit | both admin-portal pages; neither consumes `pricingAssist.summaries` client-side |
## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.**

**What I checked (adversarially):**
- **Bug reproduction confirmed real.** Pre-fix, `/system/generate-quotation` sat under the auth-only `systemProtected` group (`routes/system/register.go`, only `AuthMiddleware`), and `GenerateQuotation` (`handlers/system/ai.go`) fetched `ProductMarketCostStack` via `BuildPricingCostContextJSON` and echoed it in `pricingAssist.summaries` unconditionally. A customer JWT could read internal margin/logistics/duty data. This matched the finding.
- **Route guard is in the actual chain and fires.** `register.go:41` now registers `middleware.RequireRole(modelsAuth.AdminPortal()...)` inline on the single route. Gin order is group-middleware → route-middleware → handler, so `AuthMiddleware` (sets `userRole` from the JWT `role` claim) runs before `RequireRole`. Confirmed empirically: `TestGenerateQuotationRoleGuard` mints real JWTs via `jwtutil.GenerateJWT` against the real `Register` wiring — no token → 401, customer/supplier → 403, admin/superadmin → pass. Ran it: 5/5 subcases pass.
- **No alternate entry point.** `GenerateQuotation` and the `/generate-quotation` path are registered exactly once (grep across all `routes/**`). Not reachable via the public `pub` group or `webhooks` group.
- **Role-claim trust chain is sound and consistent.** Access tokens minted at login (`services/auth/jwt.go:39-46`) carry `jti/sub/email/role/iat/exp` with no `purpose` claim, so `AuthMiddleware`'s purpose check does not break admin access. `RequireRole`, `AuthMiddleware`, and the handler's `authContext` all read the same JWT `role` claim, and `modelsAuth.AdminPortal()` == the handler's `isAdmin` set (`Admin`, `SuperAdmin`). No divergence.
- **Defense-in-depth is real, not cosmetic.** The cost fetch, prompt injection, and `pricingAssist.summaries` echo are all gated behind `role == Admin || role == SuperAdmin` (`ai.go:177-205, 226-228`). Non-admin callers get `injectedProducts: 0` with no `summaries` key and no cost text fed to the model — so even a future miswired route cannot re-expose the stack.
- **No regressions found.** Both frontend callers (`frontend/pages/admin/ai/index.vue:296`, `frontend/pages/admin/inquiries/[id].vue:558`) live under `pages/admin/` and only read `res.quotation`/`res.text`; neither consumes `pricingAssist.summaries`. Admin behavior is unchanged (same cost-injection logic inside the gate). `go build ./...` exits 0, `go vet` clean, `gofmt -l` clean on all owned files.
- **Edge cases probed (no bypass found):** token with missing/empty `role` claim → `RequireRole` 403 (empty string matches nothing); malformed role with trailing space → 403 (safe-fail); Redis-session fallback in `AuthMiddleware` only fills role when the JWT claim is empty, and the JWT `role` remains authoritative — a customer cannot self-promote; per-request `costInject` has no shared/racy state; nil/empty `ProductIDs` handled by the same skip/dup logic as before, now inside the admin gate.

**Residual (non-blocking) notes:** (1) The defense-in-depth role gate technically duplicates the route guard; it is intentional and harmless. (2) Any *other* endpoint that echoes `ProductMarketCostStack` to non-admins would be a separate finding — none was found in the system scope in this pass (the cost-stack CRUD handlers in `handlers/admin/admin_cross_border.go` sit under the admin-portal `RequireRole(AdminPortal())` group). (3) Role enforcement still trusts the signed JWT `role` claim, which is the platform-wide pattern; not introduced by this fix.
