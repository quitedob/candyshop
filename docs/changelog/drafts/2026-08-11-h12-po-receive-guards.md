# Changelog / Devlog — 2026-08-11 H12 PO Receive Guards

**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** H12 — `ReceivePO` credits warehouse/product stock and creates a batch for every productID in the request map with no membership check against PO lines, no cap at ordered qty, no qty>0 guard, and no row lock (concurrent double-receive). A typo'd ID inflates an unrelated SKU and creates a UnitCost=0 batch that drags weighted-average COGS toward zero.
**Result:** Every received productID/qty is now validated against the PO's line items (membership, qty>0, cap at ordered - already received) before any stock is touched; the PO row is locked (SELECT ... FOR UPDATE) so concurrent double-receive is rejected; the received_qty update checks RowsAffected. Regression tests added; package build + full backend build + targeted tests pass.

---

## 1. Process

Read `backend/internal/repository/product/supplier.go` `ReceivePO` and confirmed all four gaps against the code: the loop over `receivedItems` blindly ran `received_qty + qty` and created batches without checking the ID existed on the PO, without capping at ordered qty, without rejecting qty<=0, and with the PO row read unlocked so two concurrent receives could both pass the status check. Followed the existing SQLite harness pattern (`payment_test.go`, `order_reserve_test.go`) to add a `supplier_test.go` regression suite, then verified via `go build ./...` and `REDIS_URL= go test ./internal/repository/product/...`.

## 2. Fixes in detail

#### H12 — PO receive credits arbitrary productIDs, no qty cap or row lock

- **Problem:** `backend/internal/repository/product/supplier.go` `ReceivePO` (formerly lines 145-236) iterated `receivedItems` and for every entry executed `Update("received_qty", gorm.Expr("received_qty + ?", qty))`, then (if warehouseID set) incremented `warehouse_stocks.quantity` and `products.stock_quantity`, created a `ProductBatch` with `UnitCost = itemCosts[productID]` (0 for any ID not on the PO), and recomputed weighted-average COGS. There was: (1) no membership check — a typo'd productID silently produced 0-row updates on stock but DID create a `UnitCost=0` batch and drag `weighted_avg_cost` toward 0; (2) no cap at ordered qty — over-receipt inflated stock beyond what was ordered; (3) no qty>0 guard — negative/zero qty could decrement or no-op stock; (4) no row lock — the PO was read unlocked, so two concurrent receives both saw a non-received status and double-credited stock.
- **Fix:** `backend/internal/repository/product/supplier.go` now (a) locks the PO row with `tx.Clauses(clause.Locking{Strength: "UPDATE"})` before the status check so a concurrent second receive blocks, then re-reads the committed "received" status and is rejected; (b) builds `ordered`/`received` lookup maps from the PO line items and validates every line up front — reject qty<=0 ("must be > 0"), reject IDs not on the PO ("not on PO"), reject qty above `ordered - received` ("exceeds remaining ordered qty"); (c) checks `res.RowsAffected == 0` after the `received_qty` update so a line disappearing mid-transaction fails loudly. Valid partial/complete receives behave exactly as before (status flips to `partially_received`/`received`).
- **Tests:** new `backend/internal/repository/product/supplier_test.go` adds: `TestReceivePO_RejectsUnknownProduct` (typo'd ID rejected and PO status + warehouse stock unchanged), `TestReceivePO_RejectsZeroQty` (0 and negative rejected), `TestReceivePO_RejectsOverOrderedQty` (11 on ordered 10 rejected), `TestReceivePO_PartialThenCompleteReceive` (6/10 → partial, over-remaining rejected, 4/10 → received), `TestReceivePO_DoubleReceiveRejected` (second receive after full receive rejected with "already"), `TestReceivePO_ValidReceiveCreditsStock` (warehouse + product stock credited, batch carries the PO UnitCost not 0, weighted_avg_cost recomputed, stock transaction recorded).

## 3. Verification

| Check | Result |
|-------|--------|
| `go build ./internal/repository/product/...` | ok |
| `go build ./...` (backend) | ok (exit 0, no errors) |
| `REDIS_URL= go test ./internal/repository/product/...` | ok — all 6 new ReceivePO tests pass |

## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.** I attempted to refute and could not construct a scenario where H12's symptoms still manifest.

**What I checked**
- Read `backend/internal/repository/product/supplier.go` `ReceivePO` and the caller chain (`internal/handlers/admin/admin_suppliers.go:199` `AdminReceivePO` → services → repo; the only production caller, admin-only, gated behind `ENABLE_SUPPLIER_PORTAL`).
- Reproduced the original bug conceptually against the pre-fix description: the loop over `receivedItems` blindly incremented `received_qty`, credited stock, and created batches with no membership/cap/qty guard and no row lock. Real.
- Confirmed all four root causes are closed BEFORE any mutation, so a bad line cannot partially apply:
  1. Membership: `ordered` map built from PO items (supplier.go:174-178); unknown productID → "not on PO" (line 187-189), whole tx rolls back. Test verifies PO status and warehouse stock unchanged.
  2. Qty cap: `qty > ordered - received` rejected (line 190-192); `received` is read from the committed item row after the FOR UPDATE lock, so cross-call remaining qty is enforced.
  3. qty<=0 rejected (line 184-186).
  4. Row lock: `clause.Locking{Strength:"UPDATE"}` on the PO select (line 159-160) serializes concurrent receives; the second tx re-reads the committed `received`/`cancelled` status and is rejected (line 163-165). Correct PostgreSQL semantics.
- RowsAffected guard on `received_qty` (line 207-209) is defensive-only (row cannot vanish mid-tx) but harmless. Batch `UnitCost` now comes from the validated PO line (line 224), so no UnitCost=0 batch for typo'd IDs and COGS is not dragged toward zero.
- Verified: `go build ./internal/repository/product/...` and `go build ./...` exit 0; `go vet` clean; `REDIS_URL= go test -count=1 -v ./internal/repository/product/...` passes all 6 new tests + pre-existing `TestUpdate_PersistsZeroValues`. The new tests are meaningful — the old code returned `nil` for unknown/zero/over-order/double-receive, so every one of those tests would FAIL pre-fix.
- The one `go test ./internal/handlers/admin/...` failure (`TestAdminUpdateUserRole_SuperadminCanPromoteToAdmin`) is a pre-existing test-isolation artifact (passes in isolation; depends on shared local-Postgres seed state) and is unrelated to the product repository change.

**Edge cases / residual (all pre-existing, none reintroduce H12 symptoms, not blockers)**
- Cross-PO concurrency on the SAME product can still lose-update `products.weighted_avg_cost` (two txs on different PO rows each recompute from a snapshot missing the other's uncommitted batch). Pre-existing; H12 scopes to a single PO receive.
- `product_batches` insert errors are swallowed (log-only, line 235-237), as are product `stock_quantity` update errors (line 220-222). If a batch insert fails in production (e.g., NOT NULL `production_date`/`expiry_date` non-pointer `time.Time` with no defaults), stock is credited without a batch; `recomputeWeightedAvgCost` then no-ops (no batches → `NULLIF` → leaves `weighted_avg_cost` stale). Unchanged code path.
- Duplicate product lines on one PO would make the `ordered`/`received` maps last-wins (line 174-178) and could under-count; the model has no unique constraint on `(po_id, product_id)`. Speculative; typical POs have one line per product.
- Handler maps all errors to 500 `po_receive_failed` (admin_suppliers.go:213), so new validation errors surface as a generic 500 rather than a 400 with the message. Pre-existing handler contract; a loud failure still beats the old silent 200 + corrupted stock.
- Empty `receivedItems` map is a no-op at repo level (handler requires non-empty items); draft POs are still receivable. Both pre-existing.
- The concurrent-receive rejection is only directly exercised via the sequential status-guard path (SQLite ignores FOR UPDATE); the lock's correctness rests on standard PG semantics.
