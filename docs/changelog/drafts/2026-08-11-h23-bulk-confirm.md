# Changelog / Devlog — 2026-08-11 H23 — Customer order detail confirm/cancel dead-end on bulk/requisition/reorder orders
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** frontend Nuxt 3 / Vue 3 / TypeScript / Tailwind
**Trigger:** The only confirm control on `pages/customer/orders/[id].vue` was gated on a source allowlist, so a `pending_confirmation` order whose `source` fell outside it had no actionable control (dead-end).
**Result:** Confirm control now gates on the backend-authoritative `pending_confirmation` status (not `source`), and the cancel control is extended to `pending_confirmation` (the backend explicitly allows cancelling that status). Owned file typechecks clean; the single remaining whole-repo TS error is in another agent's in-flight file.
---
## 1. Process
1. Read the owned page `frontend/pages/customer/orders/[id].vue` and traced every backend path that creates a `pending_confirmation` order.
2. Confirmed the backend fix in this pass: bulk / requisition / reorder drafts now set `Source: modelsOrder.OrderSourceBulk` (`backend/internal/handlers/customer/customer_bulk_order.go:208,299,343`), and that `CustomerConfirmOrder` (`backend/internal/handlers/customer/customer_orders_write.go:381-412`) gates on **status + ownership only**, never on `source`.
3. Replaced the stale source allowlist gate with a status gate, removed the now-unused `isAIDraftOrder` computed, extended the cancel panel, and re-ran the frontend typecheck.

## 2. Fixes in detail
#### H23 — Confirm/cancel dead-end on bulk/requisition/reorder and cart-approval orders
- **Problem:** `frontend/pages/customer/orders/[id].vue:74` rendered the only confirm button behind `v-if="order.status === 'pending_confirmation' && isAIDraftOrder"`, where `isAIDraftOrder` checked `['ai_assist','bulk','inquiry'].includes(order.source)`. Before the backend fix, bulk/requisition/reorder drafts were created with `Source=""`, so they never matched and the order had no confirm control. The gate was also wrong by design: `pending_confirmation` orders arrive from `ai_assist` (`customer_order_assist.go:180`), `bulk` (bulk/requisition/reorder, `customer_bulk_order.go:208,299,343`), `inquiry` (order intake, `orderintake/intake.go:228`), **and** `cart` — a cart checkout that requires buyer-org approval transitions `pending_approval → pending_confirmation` (`services/order/approval.go:219`, `cart.go:584`). Any allowlist that omits `cart` (or any future source) re-creates the dead-end, while the backend confirm endpoint only enforces `Status == "pending_confirmation"` + `UserID` ownership. The cancel panel (`v-if="order.status === 'pending'"`) had the same problem even though the backend permits cancelling `pending_confirmation` (`customer_orders_write.go:972`, `ValidateOrderStatusTransition`).
- **Fix:** `frontend/pages/customer/orders/[id].vue`
  - Confirm section gate changed from `order.status === 'pending_confirmation' && isAIDraftOrder` to `order.status === 'pending_confirmation'` — matching the backend's authoritative contract so every confirmable `pending_confirmation` order (regardless of source) shows the control. Authorization is unchanged: the confirm API still enforces ownership + status server-side.
  - Removed the now-unused `isAIDraftOrder` computed (was only referenced by the gate).
  - Cancel panel gate changed from `order.status === 'pending'` to `['pending', 'pending_confirmation'].includes(order.status)`, aligned with the backend cancel endpoint.
  - No UTF-8 BOM was present in the file (verified: first bytes are `<template>`), so none needed stripping.

## 3. Verification
| Check | Result |
|-------|--------|
| `npm --prefix frontend run typecheck` | Owned file clean. Whole repo exits 2 due to a single pre-existing in-flight error in `pages/admin/financial/index.vue:37` (`outstandingInvoices` not defined) — a file owned by another parallel agent, unrelated to this change. |
| Backend source-allowlist cross-check | `customer_bulk_order.go` sets `OrderSourceBulk` on requisition, CSV bulk, and reorder paths; `CustomerConfirmOrder` gates on status only. |
| BOM check | None present. |

## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED** — the root-cause dead-end (owner with a `pending_confirmation` order and no actionable control) no longer manifests for any source, and I could not construct a credible scenario where the original bug survives or a new one is introduced on the owned path.

### Attempted refutations (all failed)

1. **Source coverage.** Verified every backend path that produces `pending_confirmation`: `ai_assist` (assist handler), `bulk`/requisition/CSV/reorder (`customer_bulk_order.go:208,299,343` set `OrderSourceBulk` + `OrderStatusPendingConfirm`), `inquiry` (`services/orderintake/intake.go:228-229` sets `OrderSourceInquiry` + `OrderStatusPendingConfirm`), and cart via buyer-org approval (`approval.go:219`: `pending_approval → pending_confirmation`). `CustomerConfirmOrder` (`customer_orders_write.go:381-412`) enforces only `Status == "pending_confirmation"` + `UserID == owner` — never `source`. So gating the UI on status alone is exactly the backend contract; the old source allowlist was the bug, and any allowlist (including one that "fixed" by adding `cart`) would re-break on the next new source. The fix is structurally correct.
2. **Legacy data.** A bulk/requisition/reorder order created *before* the backend `Source` fix (i.e. `source=''`) still reaches `pending_confirmation`. The new gate is status-only, so these orders now show the confirm button too — this is precisely the dead-end case the finding described, now resolved even for stale rows.
3. **Regression on the non-affected path.** `isAIDraftOrder` was the only consumer removed; no remaining source reference to it in `frontend/pages`, `frontend/components`, or the owned file (the only hits are stale `.output/` build artifacts, not source). The confirm callout's compliance checkbox + `confirmOrder` payload are unchanged and still function. Drive-by `Number(idx)`/`Boolean(gatewayLoading)` edits are type-safety no-ops (verified type-clean).
4. **Backend confirm/cancel validity.** Transition matrix (`models/order/order.go:40-57`) confirms `pending_confirmation → cancelled` and `pending_confirmation → confirmed` are both allowed; the new `['pending','pending_confirmation']` cancel gate matches `CustomerCancelOrder` (`customer_orders_write.go:972`).
5. **Typecheck.** Re-ran `npm --prefix frontend run typecheck` (with `INTERNAL_API_BASE`/`BACKEND_URL` set). The owned file `orders/[id].vue` is clean in every run. The whole-repo exit-2 is driven by other agents' in-flight files — `pages/customer/dashboard.vue` (`quoteCount`, `notificationUnread`) and `pages/admin/webhooks/index.vue` (`createdSecret`) in my runs (the fixer saw `financial/index.vue:37` in theirs). The specific failing file differs between runs because parallel agents are editing concurrently; none of these are in the fixer's ownership. Fixer's own-file claim holds.
6. **BOM.** First bytes are `3c 74 65 6d 70 6c 61 74` = `<template`; no UTF-8 BOM present. Nothing to strip.

### Residual (non-blocking) edge cases

- **Non-owner approver dead-control.** `CustomerGetOrder` (`customer.go:45-81`) returns the order to a `canApprove` non-owner. For a non-owner approver viewing a `pending_confirmation` order, the confirm button now shows regardless of source and will 403 on click (`order.UserID != userID`). This widens a *pre-existing* cosmetic pattern (the old gate already showed the button to non-owner approvers for `ai_assist`/`bulk`/`inquiry` sources) to cart-sourced orders. It is a dead control, not a security hole — the backend still enforces ownership — and it does not affect the owner, who is the intended actor. Gating on `isOwner` too would be a stricter-but-nicer enhancement, not a fix requirement.
- **i18n copy mismatch.** The confirm callout reuses `customer.orders.ai_draft_notice`/`ai_draft_review` ("This is an AI-generated draft order…") for *all* `pending_confirmation` orders, including cart-sourced ones that reached the state via approval. The copy is now inaccurate for non-AI orders (acknowledged by the fixer). Cosmetic only — the control is functional, the keys exist in all 8 locales, and the finding was about a dead-end, not wording.
- **Compliance recheck can still block confirm.** A bulk/requisition/reorder draft created without an address and without official compliance evidence will fail the G20 recheck at confirm (`customer_orders_write.go:450`). The button exists and is enabled after the ack checkbox; failure surfaces as an error message, so this is a backend-validation gate, not a UI dead-end (and G20 is tracked separately).

**Bottom line:** The fix closes the root cause for all entry points and all data vintages. The residual items are cosmetic or pre-existing and do not recreate the original dead-end.
