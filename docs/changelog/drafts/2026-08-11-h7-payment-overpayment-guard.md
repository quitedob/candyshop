# Changelog / Devlog — 2026-08-11 H7 Payment Overpayment Guard

**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** H7 — overpayment guard is inverted at fully-paid balance; a fully-paid order (remaining == 0) accepted unlimited extra payments, enabling double-submit / over-collection.
**Result:** Guard fixed to reject any positive payment once remaining is 0; regression tests added; package build + full backend build + targeted tests pass.

---

## 1. Process

Read `backend/internal/repository/order/payment.go` `CreateWithBalanceCheck` and the existing test harness in `payment_test.go`. Confirmed the `&& remaining > 0` clause short-circuits the overpayment rejection precisely when the balance is exhausted. Removed the clause, added two regression tests (positive overpayment on fully-paid order rejected; zero-amount payment on fully-paid order still allowed), and verified via `go build ./...` and `REDIS_URL= go test ./internal/repository/order/...`.

## 2. Fixes in detail

#### H7 — Overpayment guard is inverted at fully-paid balance

- **Problem:** `backend/internal/repository/order/payment.go:80` guarded with `if payment.Amount > remaining && remaining > 0`. When the order is fully covered by confirmed+pending payments, `remaining` is clamped to `0`, making the `remaining > 0` term false and the whole check inert — so any positive payment persisted, and two concurrent full-balance payments both landed (double-submit over-collection).
- **Fix:** `backend/internal/repository/order/payment.go:84` now rejects whenever `payment.Amount > remaining`, including `remaining == 0`. Legitimate exact-balance (`Amount == remaining`) and partial (`Amount < remaining`) payments are unaffected, and a zero-amount payment on a fully-paid order is still allowed (not an overpayment). Added a comment referencing H7.
- **Tests:** `backend/internal/repository/order/payment_test.go` adds `TestCreateWithBalanceCheck_RejectsOverpaymentWhenFullyPaid` (order brought to 60+40=100 confirmed, `Amount: 1` must error with "exceeds remaining balance") and `TestCreateWithBalanceCheck_AllowsZeroPaymentWhenFullyPaid` (`Amount: 0` must succeed on a fully-paid order).

## 3. Verification

| Check | Result |
|-------|--------|
| `go build ./internal/repository/order/...` | ok |
| `go build ./...` (backend) | ok (exit 0, no errors) |
| `REDIS_URL= go test ./internal/repository/order/...` | ok — all 5 balance-check tests pass, incl. the 2 new ones |

## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.**

I attempted to refute the fix and could not construct a credible scenario where the H7 bug still manifests or where the change regresses adjacent behavior.

**What I checked**
- Reproduced the bug from source: old guard at `backend/internal/repository/order/payment.go:80` was `if payment.Amount > remaining && remaining > 0`. With `remaining` clamped to `>= 0` at lines 77-79, a fully-paid order leaves `remaining == 0`, making `remaining > 0` false and short-circuiting the whole check — so any positive payment persisted and a double-submit of full-balance payments both landed. Bug is real and matches the finding.
- New guard at line 84 (`if payment.Amount > remaining`) fires whenever `remaining == 0 && payment.Amount > 0`. Exact-balance (`Amount == remaining`) and partial payments are unchanged. A zero-amount payment on a fully-paid order still passes (`0 > 0` is false) — consistent, harmless, not over-collection.
- Concurrency/TOCTOU: the balance is recomputed inside the `SELECT ... FOR UPDATE` order-row lock (lines 56-64), so two concurrent full-balance checkouts serialize; the second observes the first's committed pending payment, `remaining` reaches 0, and the fixed guard now rejects it. This closes the exact double-submit case named in the finding. (No concurrency test exists for this path, but the locking + guard logic is sound.)
- Callers: all production payment creation routes through `CreatePaymentWithBalanceCheck` → `CreateWithBalanceCheck` — `backend/internal/handlers/admin/admin_payments.go:111`, `backend/internal/handlers/customer/customer_payments.go:225`, `backend/internal/services/order/gateway_payment.go:104`. The plain bypass `CreatePayment` (payment.go:59) has no callers. The error string still contains `"exceeds remaining balance"`, so handler substring matching (admin_payments.go:112, customer_payments.go:226) is intact.
- Regression on legitimate flows: for `remaining > 0` the comparison is byte-for-byte identical to the old behavior, so no legitimate exact/partial payment is newly rejected. The only newly-rejected case is a positive amount at `remaining == 0`, which is the intended fix.
- Negative/zero inputs: all three callers reject `amount <= 0` (customer_payments.go:201, gateway_payment.go:86) or `amount < 0`/NaN/Inf (admin_payments.go:67), so the raw `Amount > remaining` comparison is never fed a negative from the API layer.
- Verification: `go build ./internal/repository/order/...` ok; `go build ./...` (backend) ok; `REDIS_URL= go test -count=1 -v -run TestCreateWithBalanceCheck ./internal/repository/order/...` — all 5 tests pass fresh (not cached), including the 2 new regression tests.

**Edge cases noted (none refute H7, all pre-existing, not introduced by this fix)**
- Raw float64 comparison: if the confirmed+pending sum float-undershoots the total by a sub-cent residue (e.g. 33.33+33.33+33.34), `remaining` stays a tiny positive and a small top-up would still pass. This is a pre-existing precision gap in the codebase's raw-float style (the epsilon-aware `money.MoneyCoversTotal` is only used for order status, not this balance check). Severity is limited to sub-cent residues, not "unlimited extra payments", so it is out of H7's scope.
- The allocated sum counts only `confirmed + pending`; an `authorized`-but-unconfirmed payment's amount is not reserved, so a future checkout sees an inflated `remaining`. This is a separate, pre-existing status-coverage gap (arguably intended, since an authorized payment is not yet collected) and is not the inverted-guard bug H7 describes.
- The repo does not enforce order ownership (cross-tenant); that is done in handlers (e.g. gateway_payment.go:79 `order.UserID != userID`). Unchanged by the fix.

Recommended (non-blocking) follow-ups if a separate ticket is desired: use `money`-based tolerance for the `remaining` comparison, and include `authorized` amounts in the allocated sum.
