# Changelog / Devlog — 2026-08-11 G18 Access-Token Revocation Fail-Open

**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** G18 — `AuthMiddleware` treated "Redis access session not found" (the post-revocation signal) the same as "Redis down" and fell back to trusting the raw JWT claims, so logout / admin token-revoke / password rotation had NO effect on an already-issued access token (fail-open).
**Result:** Missing/revoked session now DENIES with 401; genuine Redis outage / unverifiable state now fails closed with 503 (never falls back to trusting the JWT). Regression tests assert a revoked-session token is rejected even with valid claims; package build + full backend build + targeted tests pass.

---

## 1. Process

Read `internal/middleware/auth.go` and `internal/pkg/authsession/redis.go` and confirmed the fail-open: the switch grouped `ErrSessionNotFound` and `ErrStoreUnavailable` into one "log and keep trusting JWT claims" branch. Traced issuance (`services/auth/jwt.go` `GenerateAccessToken` always writes the access session before returning a token, and errors out if the write fails) and revocation (`RevokeAccessTokenByString`, `RevokeAllUserAccessSessions`) to confirm `ErrSessionNotFound` is exactly the revocation/expiry signal. Rewrote the middleware branch to deny vs fail-closed, tightened the redis.go error contract, and added an in-memory regression test for the middleware. Verified with `go vet`, `go build ./...`, and `REDIS_URL= go test` on the affected packages and the auth dependency chain.

## 2. Fixes in detail

#### G18 — revoked/missing access session no longer falls back to raw JWT claims

- **Problem:** `internal/middleware/auth.go:70-76` (pre-fix) — the `case errors.Is(serr, authsession.ErrSessionNotFound), errors.Is(serr, authsession.ErrStoreUnavailable)` branch only logged and continued to `c.Next()`, so a session deleted by logout / `RevokeAccessTokenByString` / `RevokeAllUserAccessSessions` (password rotation) kept the request authorized purely on JWT signature + `exp`. A stolen access token therefore kept working after revocation.
- **Fix:** `internal/middleware/auth.go:80-96` — the branch is now split:
  - `ErrSessionNotFound` (or empty result) → `401 token_invalid`, request aborted. Revocation now takes effect immediately.
  - `ErrStoreUnavailable` and any other unexpected error → `503 service_unavailable`, request aborted (fail closed). The revocation state cannot be verified, so an unverifiable JWT is never trusted. Log lines added for both paths.
  - Valid session path is unchanged (UserID mismatch still → 401; email/role backfilled from session).
  - Doc comment on `AuthMiddleware` updated to state the REQUIRES/fail-closed contract.
- **Fix (supporting):** `internal/pkg/authsession/redis.go:204-232` — `ValidateAccessSession` doc comment now declares the two-sentinel contract (`ErrSessionNotFound` = invalid, deny; `ErrStoreUnavailable` = unverifiable, fail closed), and a corrupt/empty payload unmarshal error is now wrapped as `%w` `ErrStoreUnavailable` instead of being returned as a bare error, so callers always classify non-`ErrSessionNotFound` failures as unverifiable.
- **Tests:** `internal/middleware/auth_test.go` (new) — `fakeSessionStore` implements `authsession.Store`; `TestAuthMiddleware_RevokedSessionDenied` asserts 401 for a validly-signed JWT whose session reports `ErrSessionNotFound` (fails on the pre-fix code, which returned 200), `TestAuthMiddleware_StoreUnavailableFailsClosed` and `TestAuthMiddleware_StoreUnavailableWrappedFailsClosed` assert 503 for the sentinel and a wrapped infra error, plus positive tests: valid session passes, session/UserID mismatch denied, and non-Redis (`UsesRedis()==false`) / nil store keep JWT-only behavior.

## 3. Verification

| Check | Result |
|-------|--------|
| `go vet ./internal/middleware/... ./internal/pkg/authsession/...` | ok |
| `go build ./internal/middleware/... ./internal/pkg/authsession/...` | ok |
| `go build ./...` (backend) | ok (exit 0, no errors) |
| `REDIS_URL= go test -count=1 ./internal/middleware/... ./internal/pkg/authsession/...` | ok — all pass, incl. 7 new AuthMiddleware cases |
| `REDIS_URL= go test -count=1 ./internal/services/auth/... ./internal/handlers/auth/...` | ok — auth dependency chain unaffected |
| `gofmt -l` on edited files | `auth.go` / `auth_test.go` clean; `redis.go` flagged but ONLY for pre-existing issues outside this fix — CRLF line endings (package-wide convention, `store.go` same) and a long one-liner in the untouched `PostgresRefreshAdapter.RevokeAllUserAccessSessions`. The G18-edited `ValidateAccessSession` block is gofmt-clean. |

## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED** for the reported root cause, with residual adjacent gaps documented below that are out of G18's scope and should be filed as follow-ups.

**Root cause is genuinely closed.** I audited `internal/middleware/auth.go:56-98` and `internal/pkg/authsession/redis.go:204-231`. The pre-fix branch that grouped `ErrSessionNotFound` and `ErrStoreUnavailable` into "log + keep trusting JWT claims" is gone. The switch now has four distinct outcomes, all verified:
- valid session + `UserID` match → pass (email/role backfilled only when claim is empty);
- empty result (`sess==nil, serr==nil`) → 401 `token_invalid` (defensive; `RedisStore` never returns this, `PostgresRefreshAdapter` is gated out by `UsesRedis()==false`);
- `ErrSessionNotFound` (the post-revocation signal from logout / `RevokeAccessTokenByString` / `RevokeAllUserAccessSessions`) → 401;
- any other error (incl. the now-wrapped `ErrStoreUnavailable` for corrupt payloads) → 503 `service_unavailable`, never trusting the JWT.

The issuance path (`services/auth/jwt.go:52-59`) still hard-errors if `SaveAccessSession` fails, so every token handed out with Redis enabled has a live session — a missing session therefore unambiguously means revoked, and the 401 is correct.

**Verification reproduced.** I independently re-ran `REDIS_URL= go test -count=1 ./internal/middleware/... ./internal/pkg/authsession/...` → both packages `ok` (7 middleware cases + redis cases). The `TestAuthMiddleware_RevokedSessionDenied` regression does fail on pre-fix code (which returned 200). `go vet` on the two owned packages passes.

**Build note (unrelated break).** `go build ./...` is currently red ONLY in `internal/repository/order` (`fulfillment.go:65` / `order_stock_allocate.go:665` — `deductWarehouseStock` now expects a `bool` arg its callers don't pass; plus `order_stock_allocate.go:11` unused `log` import). That is the G21 work area another agent is mid-edit on (tasks #15-18). It is unrelated to G18; the G18-owned packages built and tested clean before that edit landed.

**Edge cases / residual gaps (not G18's fail-open, but "revocation has no effect" still holds in these paths — recommend follow-up findings):**
- `internal/middleware/supplier_auth.go:26-40` — `SupplierAuthMiddleware` trusts a `supplier`-role JWT on signature + email lookup alone; it never consults the access session. A stolen supplier access token (issued via `GenerateAccessToken` with a jti + Redis session) keeps working on supplier routes after logout/revoke. Adjacent gap, distinct mechanism (absent check, not fail-open fallback).
- `internal/pkg/authsession/redis.go:269-307` — when `REDIS_URL` is unset, `PostgresRefreshAdapter.UsesRedis()==false` and `ValidateAccessSession` returns `(nil,nil)`; access-session revocation is entirely unenforced. Pre-existing, acknowledged in the fixer's concerns, unchanged.
- `internal/middleware/auth.go:57,97` — if the JWT carries no `jti` (or a non-string `jti`), the session check is skipped and claims are trusted. Not attacker-reachable without the signing secret (every `GenerateAccessToken` sets `jti`), but a defensive hardening opportunity.
- Redis key eviction / `FLUSHALL` → `ErrSessionNotFound` → 401 for otherwise-valid tokens: inherent to any session-store design; fail-closed, acceptable.
- Availability tradeoff: a genuine Redis outage now 503s all authenticated traffic (no JWT fallback). This is the intended security posture mandated by the finding, not a regression.
- Concurrency: a request validating before a revocation completes may still pass — inherent to the design, not introduced here.

**No regression on the non-affected path.** Valid-session 200, session/UserID mismatch 401, and no-Redis/nil-store JWT-only 200 are all preserved and covered by tests. Route/response contract (401 `token_invalid`, 503 `service_unavailable`) matches the existing response helper usage.
