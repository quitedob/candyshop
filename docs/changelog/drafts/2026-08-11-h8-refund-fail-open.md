# Changelog / Devlog — 2026-08-11 H8 Refund Fail-Open

**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** H8 — `RefundGateway` swallows the `s.gateway()` resolution error and still marks the local payment refunded, so an unconfigured/unknown gateway method leaves money captured at Stripe/PayPal while the DB claims refunded (fail-open).
**Result:** Gateway resolution failure now returns the error before touching local status; a payment with no gateway transaction ID still marks local refunded (manual/offline path preserved); regression tests added; package build + full backend build + targeted tests pass.

---

## 1. Process

Read `backend/internal/services/order/gateway_payment.go` `RefundGateway` and confirmed the `if gw, err := s.gateway(pay.Method); err == nil` guard discards the resolution error and falls through to `s.payments.RefundPayment(ctx, pay.ID)`, which flips the local record to refunded. Replaced the guard with an explicit error return, added a fake `paymentRepository` stub to the existing test file, and verified via `go vet`, `go build ./...`, and `REDIS_URL= go test ./internal/services/order/...`.

## 2. Fixes in detail

#### H8 — RefundGateway flips local status to refunded when gateway cannot be reached

- **Problem:** `backend/internal/services/order/gateway_payment.go:176` gated the gateway refund with `err == nil`. When `s.gateway(pay.Method)` fails (unconfigured Stripe/PayPal, or unsupported method) the error was silently discarded and control fell through to `s.payments.RefundPayment` — the local payment status flipped to `refunded` even though no money was ever refunded at the gateway (fail-open).
- **Fix:** `backend/internal/services/order/gateway_payment.go:176-181` now resolves the gateway first and returns the error immediately when resolution fails, before any local status change. Only after a successful `gw.Refund` does the code call `s.payments.RefundPayment`. A payment whose `GatewayTransactionID` is nil/empty still skips the gateway and marks the local record refunded (intended manual/offline refund path), preserving prior behavior. Comment added referencing H8.
- **Tests:** `backend/internal/services/order/gateway_payment_test.go` adds `paymentRepoStub` (records whether `MarkPaymentRefundedAndRecomputeOrderStatus` ran), `TestRefundGateway_UnresolvedGateway_FailsClosed` (table over unconfigured stripe, unconfigured paypal, and unknown method "bitcoin"; asserts an error is returned AND local refund is NOT marked), and `TestRefundGateway_NoGatewayTransaction_MarksLocalRefunded` (asserts the non-gateway path still marks local refunded).

## 3. Verification

| Check | Result |
|-------|--------|
| `go vet ./internal/services/order/...` | ok |
| `go build ./internal/services/order/...` | ok |
| `go build ./...` (backend) | ok (exit 0, no errors) |
| `REDIS_URL= go test ./internal/services/order/...` | ok — all package tests pass, incl. the 3 new test cases |

## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.**

Reviewed `backend/internal/services/order/gateway_payment.go` (owned file) on disk, the added tests in `gateway_payment_test.go`, both callers, and `PaymentService.RefundPayment`. Re-ran verification from `backend/`: `go build ./...` (exit 0), `go vet ./internal/services/order/...` (ok), `REDIS_URL= go test ./internal/services/order/...` (ok, all package tests pass incl. the 3 new cases).

What was checked:
- **Root cause**: old code gated the gateway refund with `if gw, err := s.gateway(pay.Method); err == nil` and fell through to `s.payments.RefundPayment` on resolution failure, flipping the local record to `refunded` with no money refunded (fail-open). New code (lines 178-185) resolves the gateway first and returns the resolution error before any local status change, and also returns `gw.Refund` errors. Local status flips only after a successful gateway refund, or when `GatewayTransactionID` is nil/empty (manual/offline path).
- **Callers**: `AdminRefundPayment` (admin_payments.go:223-231) and `refundConfirmedPayments` (admin_orders.go:300-306) now fail-closed (error/"refund_failed" instead of a silent local flip). The cancellation loop's abort-on-first-error after possible earlier manual refunds is pre-existing loop semantics, not introduced here.
- **Legitimacy of the behavior change**: `GatewayTransactionID` is only ever set via `AttachGatewayResult` (payment.go:93), reachable only after a successful gateway Authorize in `CreateCheckout` — so only stripe/paypal rows carry a gateway tx ID in the normal flow. A non-gateway-method row with a gateway tx ID (the fixer's flagged concern) is not produced by normal checkout. Bank-transfer/manual rows (nil tx ID) still mark local refunded — covered by `TestRefundGateway_NoGatewayTransaction_MarksLocalRefunded`.
- **Tests are genuine regression guards**: `TestRefundGateway_UnresolvedGateway_FailsClosed` asserts BOTH an error is returned AND the local refund stub (`MarkPaymentRefundedAndRecomputeOrderStatus`) is NOT invoked; on the old code the stub would have been invoked and the test fails. Table covers unconfigured stripe, unconfigured paypal, and unknown method.
- **Edge cases**: nil/empty tx ID → manual path preserved; nil `pay`/`s.payments` → early error; concurrent double-refund is guarded by repo `ErrPaymentStateMismatch` (payment.go:114) and the fix narrows the silent-flip surface. 

Residual risk (not a refutation, out of H8 scope): if a gateway refund succeeds but the subsequent local `RefundPayment` write fails, money is refunded at the gateway while the DB stays `confirmed` — the inverse consistency direction, pre-existing and unchanged by this fix. Also, the success path (resolvable gateway + successful refund → local marked) has no dedicated test; the fix only reordered error handling on that path, so risk is low, but a fake-gateway success test would harden it.

Files relevant: `E:\go\website\backend\internal\services\order\gateway_payment.go` (fix), `E:\go\website\backend\internal\services\order\gateway_payment_test.go` (tests), `E:\go\website\backend\internal\services\order\payment.go` (RefundPayment/state guard).
