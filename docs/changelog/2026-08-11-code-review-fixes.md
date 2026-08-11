# Changelog / Devlog — 2026-08-11 Code-Review Fixes

**Date:** 2026-08-11
**Source report:** [`docs/reports/code-review-audit-2026-08-11.md`](../reports/code-review-audit-2026-08-11.md)
**Scope:** `backend/` (Go 1.24 / Gin / GORM), `frontend/` (Nuxt 3 / Vue 3 / TS)
**Result:** 76 files changed, ~+1730 / −263. Backend `go build` + `go vet` + tests green; frontend `nuxt build` green (2×). All report findings below the "later/large-backlog" tier fixed.

---

## 1. Process

1. **Subagent directory scan first** — three parallel read-only Explore agents mapped `/docs`, `/backend`, `/frontend` (structure, route counts, package inventory, recent mtimes) before touching anything.
2. **Read the report** — `docs/reports/code-review-audit-2026-08-11.md` (4 HIGH / 8 MEDIUM / LOWs + frontend UX/i18n/a11y + repair priority §8).
3. **Fix in priority order** — the 4 HIGH defects first (done inline, directly), then the 8 MEDIUM + repair-priority items (delegated to parallel subagents), then final verification.

---

## 2. How subagents were used (parallel, non-conflicting)

The MEDIUM backlog was delegated to **4 parallel general-purpose agents**, each owning a *disjoint file set* so concurrent edits could not collide:

| Agent | Ownership (files only it touched) | Defects |
|-------|-----------------------------------|---------|
| A | `repository/order/coupon.go`, `handlers/customer/customer_coupon.go`, `handlers/customer/cart.go`, `cmd/api/main.go` | M3, M5, M7 |
| B | `repository/order/fulfillment.go`, `handlers/customer/kyb.go`, `handlers/system/kyb_gate.go` | M4, M6 |
| C | `services/order/gateway_payment.go`, `services/order/payment.go`, `handlers/system/sse_handler.go` | M8, #13 |
| D | frontend auth pages + product pages + `useInquiry.ts` + i18n JSON (all 9 locales) | #14, #15 |

Each agent was given the report's exact findings with `file:line` references, told to **read before edit**, **match surrounding style**, and **run `go build ./...`** before reporting. After all four finished, **I ran the final merged verification** (`go build`, `go vet`, full `go test ./...`) since concurrent agents can't see each other's edits. A fifth short agent applied the WhatsApp-button guard pattern across 10 pages.

The H-series and the swap of trade-offs that needed cross-file judgment (re-pricing, PayPal signature) were done inline by me rather than delegated.

---

## 3. Fixes in detail (what was hard + how it was fixed)

### HIGH

#### H1 + H2 — Order confirm re-priced server-side (revenue-loss defects)
- **Problem:** (H1) bulk/requisition/reorder drafts are created with `UnitPrice=0`; confirm trusted `order.Subtotal` (0) → goods shipped at ≈ shipping cost, stock reserved in full. (H2) confirm accepted client-supplied item overrides while `subtotal` stayed at the stale draft value → underpayment with full reservation.
- **Fix:** new helper `repriceOrderItems` (in `checkout_pricing.go`) resolves every line server-side from the catalog / user's contract price list → market-cost stack → channel multiplier, **discarding any client `unitPrice`**. Confirm now re-prices `itemsToConfirm` first, recomputes subtotal/tax/shipping/total/COGS from the re-priced lines, and commits financials in the same transaction as the status flip + stock reserve (`ConfirmAndReserveStockWithFinancials`). Overrides are still allowed for quantity/specs (the AI-draft flow) but can never set price.

#### H3 — PayPal webhook signature verification (payment-integrity hole)
- **Problem:** public `/system/paypal-webhook` confirmed/failed payments on any known `reference_id` with no signature check (Stripe was verified; PayPal was not) → forged webhooks could falsely confirm or fail payments.
- **Fix:** added `paypal.VerifyWebhook(payload, transmissionID, time, sig, certURL, authAlgo)` mirroring Stripe: builds `transmissionID|time|webhookID|body`, verifies RSA-SHA256 against the public cert at `PayPal-Cert-Url`. Guards: cert URL restricted to `api.paypal.com` / `api.sandbox.paypal.com` (SSRF), ±5 min timestamp window, URL-unescape + base64 signature, cert cached 24h, webhookID required. Handler now verifies **before any mutation** and fails closed.

#### H4 — Dark-mode contrast
- **Problem:** sidebars/summary cards used fixed white text on `--color-primary-container`, which flips to light `#e6e1e0` in dark mode → invisible text.
- **Fix:** sidebars keep the dark surface in both themes via a scoped `.dark` token override; `.summary-card` and `.btn-secondary` use `--color-on-primary-container` / `--color-bg`; the scrolled marketing header uses `--color-bg` instead of hardcoded `rgba(255,255,255,…)`.

### MEDIUM

#### M1 — FEFO reserve/release asymmetry
- **Problem:** FEFO reserve permanently decrements `product_batches.quantity`; cancel/release restored only `products.stock_quantity`, so later FEFO orders failed "insufficient batch stock" despite aggregate stock.
- **Fix:** `restoreFEFOBatches` reverses the exact batches recorded in the order's `stock_reserved` stock-transaction rows (grouped per `batch_id`), wired into `releaseStockForOrderLine` when the product has batches.

#### M2 — `ApproveOrderWithModifications` non-atomic
- **Problem:** 3 separate writes outside a transaction; qty changes not matched to stock; reduced line leaked reserved stock / increased line oversold; COGS never recomputed.
- **Fix:** new atomic `ApprovePendingOrderWithModifications` repo method — locks the order, reconciles reserved stock to the modified lines (release reductions / reserve increases), commits financials + status flip + audit row in one tx. Service now recomputes subtotal/total and COGS (product-cost lookup injected into `ApprovalService`).

#### M3 — `RemoveCouponFromCart` IDOR + usage never refunded
- **Problem:** locked the order by ID only (no `user_id`), no status guard, never decremented `coupon.used_count` → cross-tenant total mutation and burnt `MaxUses`.
- **Fix:** scoped by `user_id` (returns a forbidden sentinel indistinguishable from "missing"), status-guarded (mutable draft states only), refunds `used_count` with a conditional decrement inside the same transaction; handler maps to 403/409.

#### M4 — Fulfillment create bypassed the order state machine
- **Problem:** could set `shipped`/`partially_shipped` on any order status.
- **Fix:** within the existing row-locked transaction, a switch enforces the legal path — `confirmed` is first advanced to `production`, then `production`/`partially_shipped` may ship; everything else rolls back. Belt-and-suspenders `ValidateOrderStatusTransition` before the final write.

#### M5 — Abandoned cart orders held reserved stock forever
- **Problem:** cart checkout created `pending` + reserved; the cleanup worker only swept `pending_confirmation`; payment-failure compensation only fired if a payment row existed.
- **Fix:** new `startAbandonedPendingOrderCleanup` worker (mirrors the draft worker, `workerlock` + `safego`, configurable interval/TTL/batch env vars). It enumerates `pending` orders older than the cutoff and, **only while still `pending` and `payment_status = unpaid`**, releases stock and cancels. Conservative: never touches an in-flight/confirmed payment.

#### M6 — KYB tier bypass via non-USD currency
- **Problem:** non-USD totals compared directly against USD caps.
- **Fix:** convert to USD using the config FX table (`cfg.ExchangeRates`) before the cap comparison, in both customer and system gates; unknown currencies fail closed (`+Inf` → no bypass). Mirrored helper in both handler files since `pkg/kyb` was outside agent ownership.

#### M7 — Cart checkout taxed the pre-coupon subtotal
- **Problem:** tax computed before the coupon discount; `ApplyCouponToCart` rebuilt the total without recomputing tax → overstated tax base.
- **Fix:** tax recomputed on the post-discount net (proportional `tax × net/subtotal`) inside the coupon transaction; `RemoveCouponFromCart` restores the full-subtotal base symmetrically so apply→remove round-trips.

#### M8 — Gateway checkout race + orphaned intents
- **Problem:** Stripe/PayPal intent created *before* the payment row was persisted; balance check not row-locked → concurrent checkouts both pass; DB failure after gateway success orphaned a collectible intent.
- **Fix:** reordered to persist the payment row (status `pending`) with the balance check first, then call the gateway; on success back-fill `gatewayTransactionId` via `AttachGatewayResult`, on failure mark the row failed (no orphan). An in-process mutex serializes concurrent balance checks.

### Repair-priority / frontend / placeholders

- **Webhooks out of public-AI middleware** — `/stripe-webhook` + `/paypal-webhook` moved to their own group so the AI disable switch (503) and AI rate limit (429) can't kill payment confirmations.
- **`enrichTradeQuery` IDOR** — ownership check mirroring `resolveOrderForUser`: admins allowed, customers only for their own trades; non-owners get no trade context injected.
- **L6** — `stripeAmount` uses `math.Round` instead of `Sprintf("%.0f", …)` (fractional-cent rounding).
- **Frontend auth middleware** — removed `guest`/`auth` guards that bounced users off `verify-email`, `reset-password`, `forgot-password` links.
- **i18n leaks** — localized sample message, WhatsApp prefill, `$` → `cur()`, packaging/OEM/scenario copy → `t()` keys; Chinese filter matchers (`夹心`/`维生素`) now localized. **33 new keys added to all 9 locales** (form, products, oem namespaces), each verified non-empty.
- **Placeholder data** — fake WhatsApp default `1234567890` removed at source (`nuxt.config`, `.env.example`); all 12 WhatsApp buttons hide when unset; fake JSON-LD (`123 Industrial Zone`, `Sweet City`, `+86-123-456-7890`) replaced with the real contact-page data (`No. 88 Shipin Road, Shanghai`, `+86 21 6731 0088`, `sales@candypro.com`).

---

## 4. Verification

| Check | Result |
|-------|--------|
| `go build ./...` (backend) | exit 0 |
| `go vet ./...` (backend) | exit 0 |
| `go test ./...` (backend) | 26 packages pass, 0 FAIL |
| `npx nuxt build` (frontend) | exit 0, "Build complete" (run twice, incl. after all post-edit changes) |
| i18n key parity | 33 new keys present + non-empty in all 9 locales |
| WhatsApp anchors | all 12 `:href="whatsappUrl"` guarded with `v-if` |

---

## 5. Deliberately NOT changed (documented follow-ups)

- **`UpdateWithVersionGuard` wiring** — `AdminUpdateOrder` already runs under `FOR UPDATE` + RowsAffected status guards (the audit itself rates the confirm path "properly guarded"); forcing optimistic-lock churn into it is high-risk with no proven benefit.
- **Fabricated category product counts** — need API-backed data (product decision), not a code tweak.
- **`track_shipment` AI tool / `quotation_human_review`** — need real carrier API / persistence integration.
- **a11y backlog (~500 items)** — report says docs understate it ~3×; requires a dedicated sweep, not part of a bug-fix pass.
