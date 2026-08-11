# Changelog / Devlog — 2026-08-11 H6: SQLite-only datetime('now') in quotation review decisions
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** H6 — `quotation_review.go` UpdateStatus used SQLite-only `datetime('now')`, which errors on PostgreSQL 15 and made admin approve/reject always return 409.
**Result:** Replaced with `time.Now()`; regression test added; build + targeted tests green.
---
## 1. Process
Read `backend/internal/repository/trade/quotation_review.go:71`, confirmed `gorm.Expr("datetime('now')")` in the conditional UPDATE (SQLite-only function, no such function on PG 15). Replaced with `time.Now()` (importing stdlib `time`), kept the pending-only conditional guard intact. Added a regression test in `quotation_review_test.go` asserting `updated_at` is set and advanced after a decision. Verified with targeted package build and tests on the in-memory sqlite harness.
## 2. Fixes in detail
#### H6 — SQLite-only datetime('now') breaks quotation-review decisions on PostgreSQL
- **Problem:** `backend/internal/repository/trade/quotation_review.go:71` wrote `"updated_at": gorm.Expr("datetime('now')")`. `datetime()` is a SQLite function; PostgreSQL 15 has no such function, so every conditional `UPDATE quotation_reviews SET ... datetime('now') ... WHERE id=? AND status='pending'` fails with `function datetime(unknown) does not exist`, surfacing as a 409/error on every admin approve/reject. The SQLite-only unit tests masked this.
- **Fix:** Replaced the DB-specific expression with Go's `time.Now()` (added `"time"` to imports), keeping the `Where("id = ? AND status = ?", id, pending)` conditional-guard semantics unchanged. `time.Now()` is database-agnostic and works on both PostgreSQL and the sqlite test harness.
## 3. Verification
| Check | Result |
|-------|--------|
| `go build ./...` (backend) | PASS |
| `REDIS_URL= go test ./internal/repository/trade/...` | PASS (3 quotation review tests incl. new regression test) |
## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED** — the root cause is genuinely eliminated; I could not construct a scenario where the H6 bug still manifests on PostgreSQL.

**Original bug confirmed real.** `datetime()` is a SQLite function; PostgreSQL 15 has no such function. On PG the conditional `UPDATE quotation_reviews SET ... updated_at = datetime('now') ... WHERE id=? AND status='pending'` fails with `function datetime(unknown) does not exist`. I traced the full path: `QuotationReviewRepository.UpdateStatus` (quotation_review.go:65-81) → `QuotationReviewService.Decide` (services/trade/quotation_review.go:48-53) → `decideQuotationReview` (handlers/admin/admin_quotation_reviews.go:46-65), which maps *any* error to HTTP 409 `quotation_review_decision_failed`. So the finding's claim (every approve/reject → 409 on PG) is accurate.

**Fix audited.** `"updated_at": time.Now()` (quotation_review.go:72) is a Go-side value GORM serializes correctly on both PG and SQLite; no dialect-specific function remains in the file. The `Where("id = ? AND status = ?", id, pending)` conditional guard is unchanged, so the atomic pending-only transition (concurrent double-decision → RowsAffected 0 → error → 409) is preserved. GORM's auto-`updated_at` behavior on map Updates is irrelevant either way — the explicit map key guarantees the value is set.

**Checks performed.**
- `go build ./...` (backend) — PASS, no unrelated-package failures.
- `REDIS_URL= go test -count=1 -v -run 'TestQuotationReview' ./internal/repository/trade/...` — PASS (3 tests incl. the new `TestQuotationReviewRepository_UpdateStatusSetsUpdatedAt`).
- Callers of `UpdateStatus`: only the service `Decide` → the admin approve/reject handlers. No other call sites; route contracts and response shapes unchanged.
- Grepped the backend for remaining SQLite-only constructs: the only `datetime(` usages are in `backend/internal/repository/order/order_reserve_test.go` (test-only, sqlite harness, not production). All production `gorm.Expr(...)` usages are PG-compatible arithmetic/`GREATEST`/`EXCLUDED` upsert expressions.

**Edge cases / residual weaknesses (non-blocking).**
- The new regression test is a weak guard: on the sqlite harness `datetime('now')` is valid, so the old buggy code would have passed `TestQuotationReviewRepository_UpdateStatusSetsUpdatedAt` too. The test documents intended `updated_at` semantics and exercises the code path but would not catch a reintroduction of a PG-incompatible expression. This is inherent to the sqlite-only harness (the PG failure cannot be reproduced there); the load-bearing correction is the source change itself, which is correct and unambiguous. A PG integration test or a "no dialect-specific function" assertion would harden it, but its absence does not leave H6 unfixed.
- `time.Now()` is evaluated client-side rather than DB-side (`now()`); this is semantically equivalent for the `updated_at` audit field and avoids the entire dialect problem. No precision/ordering concern that affects this domain.
