# Changelog / Devlog — 2026-08-11 Payment gateway correctness (G16)

**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** Audit G16 — Stripe zero-decimal scaling + webhook replay; PayPal APPROVED treated as paid + Capture amount dropped.
**Result:** All three sub-issues fixed with regression tests; `go build ./...`, `go vet`, and affected-package tests green.
---

## 1. Process
Read the three owned files plus the gateway interface, `GatewayPaymentService`, the payment repo state machine, and the existing payment/`paypal_webhook` tests. Verified the PayPal Orders v2 capture request schema against the live API docs (the endpoint does **not** accept an `amount` — it captures the full order amount). Implemented fixes, then wrote regression tests that fail on the pre-fix path (verified empirically by temporarily reverting), then ran build/vet/tests.

## 2. Fixes in detail

#### G16a — Stripe Capture/Refund zero-decimal scaling
- **Problem:** `Capture`/`Refund` hardcoded `amount*100` (`stripe.go` pre-fix :71/:82), so a zero-decimal currency (JPY, KRW, VND, …) could never capture/refund a partial amount — Stripe would reject (or over-take) `amount*100` while the correct minor unit is the raw amount. `Authorize` was already currency-aware via `stripeAmount`; only the partial capture/refund paths were not.
- **Fix:** `Capture`/`Refund` now read the PaymentIntent's currency via a new `intentCurrency` (`GET /payment_intents/{id}`) and scale through the existing `stripeAmount(amount, currency)` helper. The currency is read from the intent rather than cached from Authorize because the `PaymentGateway` interface carries no currency and a cache would not survive restarts or intents created outside this adapter. Zero-amount (full capture/refund) paths are unchanged — Stripe does the scaling itself.
- **Files:** `internal/pkg/payment/stripe/stripe.go` (`Capture`, `Refund`, new `get` + `intentCurrency`).

#### G16b — Webhook event-ID dedupe
- **Problem:** The Stripe `stripeEvent` struct dropped the event `id` (`stripe.go` :219-222) so consumers could not distinguish a replay from a first delivery; the PayPal handler had no dedupe at all, so a redelivered `PAYMENT.CAPTURE.COMPLETED` re-attempted the confirm.
- **Fix (Stripe):** added `ID` to `stripeEvent` and a new exported `EventID(payload)` helper so the event id is parseable. The authoritative replay guard is the payment status state machine (`ConfirmPayment`/`AuthorizePayment`/`FailPayment` are all `WHERE … status = <expected>` with a `RowsAffected` check), which already rejects the second transition — verified for every Stripe event type.
- **Fix (PayPal):** the handler now parses `event.id` and skips replays via an in-memory TTL dedupe set (`isPayPalEventProcessed`/`markPayPalEventProcessed`, 24h TTL, mutex-guarded, bounded). Combined with the state machine this makes replays idempotent across restarts too.
- **Files:** `internal/pkg/payment/stripe/stripe.go`, `internal/handlers/system/paypal_webhook.go`.

#### G16c — PayPal APPROVED is not a payment confirmation; Capture amount verified
- **Problem:** `CHECKOUT.ORDER.APPROVED` was grouped with `PAYMENT.CAPTURE.COMPLETED` and, when `GatewayPayment == nil` or the payment had no gateway transaction id, fell through to a bare `ConfirmPayment` — confirming the payment (and letting goods ship) before any money was captured. Separately, the PayPal `Capture` dropped its `amount` argument (`paypal.go:125`), so a capture whose value diverged from the recorded payment amount was silently accepted.
- **Fix (handler):** split the branch. `PAYMENT.CAPTURE.COMPLETED` (emitted only after funds are captured) confirms directly. `CHECKOUT.ORDER.APPROVED` never calls `ConfirmPayment`; it triggers a server-side `CaptureAndConfirm` and only confirms if the gateway capture succeeds. With no gateway path the payment stays pending. The confirm-vs-capture logic is extracted into `handlePayPalPaymentEvent` for direct unit testing.
- **Fix (adapter):** the Orders v2 capture endpoint does **not** accept an amount (verified against `docs.paypal.ai`); the adapter cannot "send" the amount. Instead it now sends the documented empty object `{}` (fixing a latent null-body risk) and **verifies** the capture reported by PayPal against the requested amount (`amountsEqual`, half-cent numeric tolerance, currency-agnostic for zero-decimal values), returning an error on mismatch so a wrong-value capture can never confirm a payment.
- **Files:** `internal/handlers/system/paypal_webhook.go`, `internal/pkg/payment/paypal/paypal.go`.

## 3. Verification
| Check | Result |
|-------|--------|
| `go build ./...` (backend) | PASS |
| `go vet ./internal/pkg/payment/... ./internal/handlers/system/...` | PASS |
| `REDIS_URL= go test ./internal/pkg/payment/... -count=1` | PASS (stripe 6 tests incl. zero-decimal capture/refund + USD no-regression; paypal 3 capture tests + existing webhook tests) |
| `REDIS_URL= go test ./internal/handlers/system/... -count=1` | PASS (incl. APPROVED-does-not-confirm, CAPTURE.COMPLETED-confirms, replay idempotency, dedupe, existing signature tests) |
| `REDIS_URL= go test ./internal/services/order/... ./internal/repository/order/... -count=1` | PASS (dependents) |
| Regression tests fail on pre-fix path | Verified empirically: `TestCaptureZeroDecimalCurrency` fails (sends 123400), `TestHandlePayPalEvent_ApprovedDoesNotConfirm` fails (payment confirmed) — both reproduced by temporarily reverting, then restored. |

## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.** Independent verification: `go build ./...` PASS; `REDIS_URL= go test ./internal/pkg/payment/stripe/... ./internal/pkg/payment/paypal/... ./internal/handlers/system/... ./internal/services/order/... ./internal/repository/order/... -count=1` all PASS (could not run `-race` — cgo disabled on this Windows build env; the dedupe mutex is a plain Lock/Unlock so no race concern). I could not construct a credible scenario where any G16 sub-issue still manifests or regresses the non-affected path. Residual concerns (none blocking) below.

**G16a (Stripe zero-decimal Capture/Refund) — correct.** `intentCurrency` (stripe.go:138) does `GET /payment_intents/{id}` and scales via the existing `stripeAmount` (stripe.go:214), the same helper `Authorize` already used. Units are internally consistent: `CreateCheckout` stores `pay.Amount == req.Amount`, and both Authorize and Capture round the same float through `math.Round`, so `amount_to_capture` exactly equals the authorized amount for two-decimal and zero-decimal alike. Pre-fix, a JPY intent authorized for 1234 got `amount_to_capture=123400` → Stripe rejects → capture was wholly broken; now correct. Verified: the only Capture/Refund callers (`gateway_payment.go:160,183`) pass `pay.Amount` (full amount), so partial-capture precision is theoretical but handled. Operational note: the new GET requires `payment_intents:read` on the secret key; a restricted `rk_*` key lacking read scope would fail every capture/refund (fail-closed, not a money bug).

**G16b (webhook replay) — functionally safe; two gaps to record.**
- *Stripe handler was NOT modified.* `stripe_webhook.go` is unmodified and still never dedupes event IDs; the new exported `stripeAdapter.EventID` (stripe.go:321) has zero callers (dead code). The finding's literal "dedupe event IDs in the handler" requirement is therefore unmet. It is nonetheless functionally covered: I traced all four Stripe event paths and the pre-existing state machine (`ConfirmPayment` `WHERE status=pending`+RowsAffected, `FailPayment` pending|authorized-only, `AuthorizePayment` pending-only) rejects every replay — succeeded/requires_capture replay → status authorized → error; succeeded/succeeded replay → confirmed → error; payment_failed/canceled replay → failed → early return, so `compensateOrderForFailedPayment` cannot double-cancel. Because the handler always returns HTTP 200 regardless of processing outcome, the state machine is the only idempotency guard and it holds. This satisfies the *behavioral* requirement; the *architectural* gap (no handler-level dedupe) is real but benign.
- *PayPal dedupe TOCTOU.* `isPayPalEventProcessed` → `markPayPalEventProcessed` (paypal_webhook.go:119-124) is not an atomic check-and-set: two concurrent deliveries of the same event id can both proceed. Consequence is harmless (state machine backstops confirm/capture; second PayPal capture gets 422 and errors out). Also, `markPayPalEventProcessed` runs BEFORE processing (line 124); since the handler returns 200 even on processing failure, PayPal's own redelivery retry is not triggered anyway — so the mark-before-process ordering does not swallow recoverable retries beyond the pre-existing always-200 design. Both noted, neither manifests the G16 bug.

**G16c (PayPal APPROVED is not payment confirmation; Capture amount verified) — correct.**
- Handler: `PAYMENT.CAPTURE.COMPLETED` confirms directly (paypal_webhook.go:155); `CHECKOUT.ORDER.APPROVED` (line 161) never calls bare `ConfirmPayment` — with `GatewayPayment == nil` or `pay.GatewayTransactionID == nil` it returns leaving the payment pending (fail-closed); with a transaction id it only confirms via `CaptureAndConfirm` after a successful gateway capture. Pre-fix, the nil-gateway/txid branch fell through to `ConfirmPayment` (shipped goods before funds moved) — the exact bug; now removed. Even if `CaptureAndConfirm`'s `ConfirmPayment` failed after a successful capture, the subsequent `PAYMENT.CAPTURE.COMPLETED` webhook (different event id) confirms — self-healing. Manual/admin payment-creation paths (`admin_payments.go:96`, `customer_payments.go:209`) create pending rows without a gateway reference and never reach PayPal, so they are unaffected by the handler change.
- Adapter: Orders v2 capture does not accept an amount in the body (correct per spec), so the fix verifies the captured value (`paypal.go:161-168`) via `amountsEqual` (numeric, half-cent tolerance, currency-agnostic: "1234" JPY == 1234.0). Since PayPal captures exactly the order amount fixed at creation, the captured value matches `formatAmount(pay.Amount)` to the cent; a divergence (e.g. locally edited `pay.Amount`) fails closed so a wrong-value capture can never confirm. Sending `{}` instead of the previous nil body is the documented form and fixes a latent `UNSUPPORTED_MEDIA_TYPE` risk.
- Residual edge (pre-existing, outside G16): `paypal.Refund` still uses `formatAmount` (`%.2f` → "1234.00" for JPY) which PayPal may reject for zero-decimal currencies; G16 only claimed Capture.

**Regression check (non-affected path):** new `"duplicate"` 200 response only for replayed PayPal events — additive, no contract break; empty event ids are never marked/skipped; `go build ./...` green confirms the `verifyTransmissionSignature`/`certFetch` refactor and the `stripe` `get`/`intentCurrency` additions compile against all callers.

**Bottom line:** all three G16 sub-issues are resolved for the entry points that matter; the two recorded gaps (Stripe handler not deduped at handler level; PayPal dedupe TOCTOU) are backstopped by the persistent state machine and cannot double-process money or ship goods prematurely.
