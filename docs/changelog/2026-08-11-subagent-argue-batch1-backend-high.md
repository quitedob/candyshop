# Changelog / Devlog — 2026-08-11 Subagent-Argue Fix Pass — Batch 1: Backend HIGH (H1–H15)

**Date:** 2026-08-11
**Source report:** [`docs/reports/code-review-audit-2026-08-11.md`](../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go 1.24 / Gin / GORM / PostgreSQL
**Trigger:** The full-audit report listed 24 HIGH findings; the 15 backend HIGH items (H1–H15) were fixed first via a fixer + adversarial-arguer subagent workflow.
**Result:** All 15 backend HIGH findings fixed and adversarially verified (CONFIRMED_FIXED). Two findings (H10, H11) were initially REFUTED because the root cause extended beyond the fixer's owned file; follow-up rounds with expanded ownership closed them. Merged `go build ./...`, `go vet ./...`, `REDIS_URL= go test ./...` all green. 15 devlog drafts written (details in [`drafts/`](drafts/)).

---

## 1. Process

1. **Baseline first** — confirmed `go build` / `go vet` / `go test` (with `REDIS_URL=`) passed on the working tree before any fixer ran.
2. **Per-bug fixer subagent** — one fixer per finding with a disjoint file-ownership contract, told to read-before-edit, match style, add a regression test, and run `go build ./...` + targeted tests.
3. **Adversarial arguer subagent** — one arguer per finding, instructed to *refute* (re-verify the bug, audit edge cases and regressions, re-run tests), appending its verdict to the fixer's devlog draft.
4. **Revise-on-refute** — bounded loop (max 2 revisions); REFUTED fixes were re-dispatched with expanded ownership.
5. **Lead merged verification** — full build/vet/test after each workflow (concurrent agents can't see each other's edits; two transient interface mismatches from H3/H13 were resolved by the agents before completion).

## 2. How subagents were used (parallel, non-conflicting)

Each fixer owned a disjoint file set (full per-file ownership recorded in the drafts). The two refuted findings (H10, H11) went through a second fixer round with the arguer-identified files added to their ownership.

| Agent | Primary ownership | Finding(s) |
|-------|-------------------|------------|
| fix:H1 | `handlers/admin/admin_extensions.go`, `admin_users.go`, `services/auth/auth.go` | H1 |
| fix:H2 | `handlers/customer/customer_return.go` | H2 |
| fix:H3 | `handlers/customer/customer_negotiation.go`, `admin/admin_negotiation.go`, `services/order/negotiation.go`, `repository/order/negotiation.go` | H3 |
| fix:H4 | `handlers/customer/customer_logistics.go`, `repository/trade/{shipment,shipment_event}.go`, `services/trade/logistics.go` | H4 + M3 + M4 |
| fix:H5 | `api/routes/system/register.go`, `handlers/system/ai.go` | H5 |
| fix:H6 | `repository/trade/quotation_review.go` | H6 |
| fix:H7 | `repository/order/payment.go` | H7 |
| fix:H8 | `services/order/gateway_payment.go` | H8 |
| fix:H9 | `services/order/{invoice,invoice_policy}.go` | H9 |
| fix:H10 | `services/orderintake/intake.go` → follow-up `handlers/customer/{customer_orders_write,checkout_pricing}.go` | H10 |
| fix:H11 | `repository/product/product.go` → follow-up `services/product/product.go`, `handlers/admin/{admin_products,admin_extensions,admin_inventory,admin_ai_translate,admin_inventory_xlsx}.go` | H11 |
| fix:H12 | `repository/product/supplier.go` | H12 |
| fix:H13 | `repository/translation/translation.go`, `services/translation/translation.go`, `handlers/admin/admin_translations.go` | H13 + INFO-7 |
| fix:H14 | `database/migrate_product_base_price.go` | H14 |
| fix:H15 | `Makefile`, `cmd/api/main.go` (docs import committed by lead) | H15 |

## 3. Fixes in detail

### HIGH — authorization / security

#### H1 — Plain admin can mint/promote superadmin
- **Problem:** `AdminCreateUser`/`AdminUpdateUserRole`/`AdminUpdateUser` resolved the target role via `Auth.ResolveRoleID` with no caller-role dominance check; any admin could grant `admin`/`superadmin`.
- **Fix:** Added `AuthService.CanGrantRole` (DB-backed caller-role lookup; superadmin-only for admin/superadmin targets, 403 otherwise) wired into all three role-assignment handlers. Regression tests at service + handler level.

#### H2 — Return creation unvalidated (IDOR + refund fraud)
- **Problem:** `customer_return.go` create path blind-inserted a return for any orderID/userID with arbitrary quantities/refunds.
- **Fix:** Validate order ownership, order status, per-item quantity (<= ordered and not already returned), and refund amounts before insert. Regression tests.

#### H3 — Negotiation accept/reject doesn't verify offer ownership
- **Problem:** Status transitions ran `UPDATE ... WHERE id=?` with no inquiry/user scoping; reject path had no ownership check.
- **Fix:** Scoped transitions by offer + inquiry ownership (caller's `inquiry.user_id`), 403/404 on mismatch, for customer and admin paths. Regression tests.

#### H4 + MEDIUM-3 + MEDIUM-4 — Shipment-timeline IDOR + unconditional shipment Save
- **Problem:** Customer timeline endpoint verified trade ownership but never that the shipment belonged to that trade; `Shipment.Update` was an unconditional `Save` so concurrent dispatch double-deducted.
- **Fix:** Handler verifies `shipment.TransactionID == tradeID`; dispatch converted to a conditional `UPDATE ... WHERE id=? AND status='PENDING'` with `RowsAffected==0` guard. Regression tests.

#### H5 — /system/generate-quotation leaks internal cost stack
- **Problem:** Route was auth-only with no role guard; handler injected `costStackJSON` into the response.
- **Fix:** Route now guarded to admin/superadmin (matching admin-portal pattern). Regression test.

#### H6 — SQLite datetime('now') breaks quotation-review decisions on PostgreSQL
- **Problem:** `repository/trade/quotation_review.go` used `gorm.Expr("datetime('now')")`, which errors on PostgreSQL; every approve/reject 409'd.
- **Fix:** Replaced with `time.Now()`; kept the conditional `WHERE id=? AND status='pending'` guard. Existing sqlite tests still pass.

### HIGH — financial integrity

#### H7 — Overpayment guard inverted at fully-paid balance
- **Problem:** `payment.Amount > remaining && remaining > 0` disabled the guard exactly at `remaining==0` (double-submit case).
- **Fix:** Guard now rejects any positive payment above remaining, including `remaining==0`. Regression test for the fully-paid case.

#### H8 — RefundGateway flips local status when gateway unreachable
- **Problem:** On `s.gateway()` error the real gateway refund was skipped and local status still flipped to refunded.
- **Fix:** Return the error before touching local status; only mark refunded after the gateway refund succeeds. Regression test.

#### H9 — Invoice derived from order drops shipping amount
- **Problem:** `CreateInvoice`/`UpdateInvoice` set `TotalAmount = Amount + TaxAmount`, discarding the correctly-computed `ShippingAmount` from `invoice_policy.go`.
- **Fix:** Included `ShippingAmount` in the total so derived invoices bill freight. Regression test.

#### H10 — Negotiated pricing dead data; confirm re-prices from catalog
- **Problem:** Offer unit price applied only to single-line orders; `CustomerConfirmOrder` re-priced every line from catalog and overwrote the negotiated total.
- **Fix:** Intake applies the offer price to every line and sets `Source=inquiry`; the confirm path now rescales lines to the negotiated total (`scaleInquiryLinesToTotal`), restoring `TotalAmount == Subtotal+Tax+Shipping` so the derived invoice bills the negotiated amount. Regression tests (multi-line, OEM BasePrice=0, invoice-billing chain) proven to fail pre-fix.

### HIGH — data integrity

#### H11 — ProductRepository.Update can't write zero values
- **Problem:** GORM struct `Updates` skips zero fields, so admin could never clear `featured`/`halal`, set `stockQuantity:0` or `basePrice:0`.
- **Fix:** Added zero-value-capable repository primitives (`UpdateAll`/`UpdateColumns`); migrated all four admin write paths and the XLSX importer onto them (full-row / column-scoped), including a fix for the importer's int-cell predicate diverging from its writer (decimal/exponent cells silently zeroed stock). Regression tests at repo + handler level proven to fail pre-fix.

#### H12 — PO receive credits arbitrary productIDs, no qty cap or row lock
- **Problem:** `ReceivePO` credited stock/batches for any productID without membership check, qty cap, qty>0 guard, or row lock.
- **Fix:** Validate every line up front (membership, qty>0, cap at remaining ordered), lock the PO row (`FOR UPDATE`), check `RowsAffected`. 6 regression tests.

#### H13 + INFO-7 — Translation repo mutates a shared GORM session; no ctx
- **Problem:** Finishers ran directly on the shared scoped `*gorm.DB`; methods carried no `context.Context`. (Empirically, the clause-leak does not reproduce on pinned GORM v1.30.0, but the shared-statement/ctx gap is real hygiene.)
- **Fix:** Threaded `ctx` through every repo method (`r.db.WithContext(ctx)`), the service interface, and the handler. Regression test added.

#### H14 — EnsureProductBasePrices reverts admin price edits every startup
- **Problem:** Startup migration unconditionally overwrote `base_price` for 18 seeded slugs.
- **Fix:** Added the same `(base_price = 0 OR base_price IS NULL)` guard the in-seed path uses. Regression test.

### HIGH — build / deploy

#### H15 — make prod-start broken; fresh clone fails to build
- **Problem:** `prod-start` ran `docker-compose.prod.yml` standalone (no image/build declared); `cmd/api/main.go` blank-imports the gitignored `backend/docs` so a fresh clone failed `go build`.
- **Fix:** `prod-start`/`prod-stop` now merge `docker-compose.yml`; the generated swagger under `backend/docs` was force-added to git by the lead (`git add -f backend/docs`) so clean clones build.

## 4. Verification

| Check | Result |
|-------|--------|
| `go build ./...` (backend) | exit 0 |
| `go vet ./...` (backend) | exit 0 |
| `REDIS_URL= go test ./...` | all packages pass, 0 FAIL |
| Arguer verdicts | 13 CONFIRMED_FIXED in round 1; H10/H11 CONFIRMED_FIXED after follow-up rounds (all 15 closed) |
| Devlog drafts | 15 written to [`drafts/`](drafts/), each with the arguer's appended adversarial review |

## 5. Deliberately NOT changed (documented follow-ups)

- H10 sub-cent rounding on pathological totals and H11 full-row `Select("*")` lost-update window are documented residuals in the respective drafts.
- The broad `float64` money migration and `interface{}`→`any` style sweep are separate follow-ups (see the MEDIUM/LOW batches).
