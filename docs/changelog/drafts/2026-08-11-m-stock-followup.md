# Changelog / Devlog — 2026-08-11 Stock Follow-up: Draft-Cleanup Audit Atomicity + Dispatch FEFO Warehouse Sync (G21)

**Date:** 2026-08-11
**Source report:** G21 — "Draft-cleanup audit rows preserved; dispatch FEFO syncs warehouse stock"
**Scope:** backend (`internal/repository/order`, `internal/services/trade`)
**Trigger:** adversarial review refuted the first-round fix for (f) with repository/order/order.go:653-655 and re-flagged the dispatch FEFO warehouse mirror + PG-only `GREATEST` in logistics_stock.go:251
**Result:** CLOSED — all three remaining sub-defects fixed; the draft-cleanup release is now atomic with its audit rows, the dispatch FEFO path mirrors onto `warehouse_stock`, and the reserved-drain SQL is portable (sqlite/PG). Regression tests pass (each fails on the pre-fix path).

## 1. Process

G21 landed a first round that closed dispatch idempotency (`services/trade/logistics.go`) and the order-stock FEFO/transfer fixes (`repository/order/order_stock_allocate.go`). An adversarial reviewer then refuted that the finding was fully closed, naming three concrete defects still live on disk:

1. **G21-f** — `repository/order/order.go:653-655` still swallowed the `writeStockAuditEntries` error with `log.Printf`, so the draft-cleanup worker committed the stock release + `pending_confirmation -> expired` transition while the release's `draft_expired` audit rows were silently lost.
2. **dispatch FEFO** — `services/trade/logistics_stock.go:156-229` (`logisticsDeductFEFOFromBatches`) decremented `product_batches.quantity` and `Product.StockQuantity` but never synced the `warehouse_stock` mirror. Reachable from `DispatchShipment` whenever `order.StockReserved == false` and the product has batches — the same warehouse-level drift the order-stock side already fixed.
3. **PG-only SQL** — `services/trade/logistics_stock.go:251` used `GREATEST(reserved - ?, 0)`, which does not exist in SQLite and made the reserved-drain expression unportable (untestable on the in-memory harness, and a latent cross-DB break).

Verification: each new regression test was run against a temporarily reverted pre-fix path to confirm it FAILS, then against the fixed path to confirm it PASSES. `go build ./...`, `go vet`, and the full `internal/repository/order` + `internal/services/trade` suites are green.

## 2. Fixes in detail

#### G21-f — Propagate the draft-cleanup audit-write error

- **Problem:** `ReleaseExpiredPendingConfirmationOrders` wrote the `draft_expired` audit rows inside the per-order transaction but, on failure, only printed `log.Printf("order: writeStockAuditEntries failed: %v", ae)` and continued. The transaction then committed the stock release and the `pending_confirmation -> expired` status flip, leaving the stock movement unaccountable in `stock_transactions`.
- **Fix:** `backend/internal/repository/order/order.go` — the audit-write failure is now returned from the transaction closure (`return fmt.Errorf("order %s: write draft-expired stock audit rows failed: %w", order.ID, ae)`), so the stock release + status transition roll back atomically with the audit rows. The batch worker's per-order log-and-continue contract is preserved (one order's failure does not abort the sweep), but a failed audit write now fails that order's release loudly instead of committing a silent audit gap.

#### G21 — Dispatch FEFO mirrors onto warehouse_stock

- **Problem:** `services/trade/logistics_stock.go` `logisticsDeductFEFOFromBatches` decremented `product_batches.quantity` and `Product.StockQuantity` and returned without touching `warehouse_stock`. Because `DispatchShipment` routes to it whenever `order.StockReserved == false` and the product has batches, a batch-tracked SKU's warehouse aggregate drifted by the shipped qty — exactly the drift `deductFEFOFromBatches` (order-stock side) already closed.
- **Fix:** `backend/internal/services/trade/logistics_stock.go` — after the product-level deduction, if `warehouse_stock` rows exist for the product the deduction is mirrored onto the resolved default warehouse via the new `logisticsSyncFEFOWarehouseStock` helper. A missing/insufficient row or DB error is returned (wrapped) so the dispatch transaction rolls back rather than reporting success with a stale warehouse mirror. This mirrors the order-stock `syncFEFOWarehouseStock` behavior.

#### G21 — Portable reserved-drain expression

- **Problem:** `services/trade/logistics_stock.go:251` `logisticsDeductReservedWarehouseStock` used `GREATEST(reserved - ?, 0)`, a PostgreSQL-only function. On SQLite it failed with `no such function: GREATEST`, so the reserved-drain path could not run under the in-memory test harness.
- **Fix:** `backend/internal/services/trade/logistics_stock.go` — replaced with the portable, atomic `CASE WHEN reserved - ? < 0 THEN 0 ELSE reserved - ? END`, matching the expression already used by `deductWarehouseStock` in `repository/order/order_stock_allocate.go`. Reserved clamps at 0 (never negative); shipping consumes `min(reserved, qty)` reserved units then dips into sellable stock.

## 3. Verification

| Check | Result |
|-------|--------|
| `go build ./...` (backend) | PASS |
| `go vet ./internal/repository/order/... ./internal/services/trade/...` | PASS |
| `REDIS_URL= go test ./internal/repository/order/... ./internal/services/trade/... -count=1` | PASS (all suites) |
| `TestReleaseExpiredOrders_AuditFailureRollsBackRelease` — injected `stock_transactions` insert failure | FAILS pre-fix (`released = 1`, audit lost), PASSES post-fix (`released = 0`, order stays `pending_confirmation`) |
| `TestReleaseExpiredOrders_WritesAuditRows` — happy path writes `draft_expired` audit rows | PASS |
| `TestLogisticsDeductFEFO_SyncsWarehouseStock` — warehouse mirror follows FEFO deduction | FAILS pre-fix (warehouse qty 20), PASSES post-fix (15) |
| `TestLogisticsDeductFEFO_WarehouseSyncErrorPropagated` — insufficient warehouse mirror errors | FAILS pre-fix (nil error), PASSES post-fix |
| `TestDispatchShipment_FEFOSyncsWarehouseStock` — end-to-end dispatch on `StockReserved=false` + batches | FAILS pre-fix (warehouse qty 20), PASSES post-fix (15) |
| `TestLogisticsDeductReservedWarehouseStock_PortableClamp` — reserved-drain SQL on sqlite | FAILS pre-fix (`no such function: GREATEST`), PASSES post-fix (qty 5 / reserved 0) |

## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.** I attempted to refute the fixer's claim and could not construct a credible scenario where any of the three named sub-defects still manifests or regresses. Each is closed at the exact location the prior refutation named, the new regression tests discriminate (fail on the pre-fix path, pass post-fix), and the already-landed G21 fixes (dispatch idempotency, order-stock transfer/FEFO) are preserved.

### Sub-defect (f) — audit-write error no longer swallowed

`backend/internal/repository/order/order.go:656-658` (the exact range the prior refutation named, now shifted ~3 lines by the fixer's doc-comment):
```go
if ae := writeStockAuditEntries(tx, allAudit); ae != nil {
    return fmt.Errorf("order %s: write draft-expired stock audit rows failed: %w", order.ID, ae)
}
```
This sits inside the per-order transaction closure (order.go:641), so a failed `draft_expired` audit write now returns an error, GORM rolls the transaction back, and the stock release + `pending_confirmation -> expired` flip do NOT commit. The batch worker's per-order log-and-continue contract is preserved (order.go:676-679 still logs and continues), and `released++` (order.go:671-672) is only reached on the success path inside the closure — a rolled-back order is never counted. I verified the regression test's non-vacuousness directly: `TestReleaseExpiredOrders_AuditFailureRollsBackRelease` injects a `BEFORE INSERT ON stock_transactions` trigger that aborts with `injected audit failure`; the verbose run shows the abort firing, the order staying `pending_confirmation` with `stock_reserved` intact, `released == 0`, and the test PASSING post-fix. Pre-fix the same trigger would have let the closure continue (GORM's `Transaction` only rolls back when the callback returns a non-nil error, and the old code swallowed it with `log.Printf`), committing `released = 1` + the expired flip — so the test genuinely discriminates. No other `writeStockAuditEntries` call site in the package swallows its error (grep of all 12 call sites: all return/check).

### Sub-defect (dispatch FEFO) — warehouse mirror synced

`backend/internal/services/trade/logistics_stock.go:229-245` now counts `warehouse_stock` rows for the product and, when present, resolves the default warehouse and calls the new `logisticsSyncFEFOWarehouseStock` (logistics_stock.go:253-264). A missing/insufficient row or DB error is returned wrapped, and because `DispatchShipment` runs the deduction inside its own transaction (logistics.go:100), a sync failure rolls the whole dispatch back. Reachability is confirmed: `DispatchShipment` → `logisticsDeductStockForProductLine(tx, warehouseID, productID, remaining, order.StockReserved, ...)` (logistics.go:160) → with `order.StockReserved == false` and `batchCount > 0` → `logisticsDeductFEFOFromBatches` (logistics_stock.go:326-333). The implementation is line-for-line behaviorally equivalent to the order-stock side (`deductFEFOFromBatches` / `syncFEFOWarehouseStock` in order_stock_allocate.go:552-564 / 573-584), which is exactly the parity the finding demanded. `TestLogisticsDeductFEFO_SyncsWarehouseStock` (warehouse 20→15), `TestLogisticsDeductFEFO_WarehouseSyncErrorPropagated`, and the end-to-end `TestDispatchShipment_FEFOSyncsWarehouseStock` (on `StockReserved=false` + batched product) all PASS post-fix; pre-fix each would have seen the warehouse row stay at 20 (or a nil error), so they discriminate.

### Sub-defect (PG-only GREATEST) — portable clamp

`backend/internal/services/trade/logistics_stock.go:291` now uses `gorm.Expr("CASE WHEN reserved - ? < 0 THEN 0 ELSE reserved - ? END", qty, qty)` — identical to the expression already in `deductWarehouseStock` (order_stock_allocate.go:219). A repo-wide grep for `GREATEST(reserved` finds no remaining matches; the only GREATEST left in the package is in a comment. `TestLogisticsDeductReservedWarehouseStock_PortableClamp` runs the reserved-drain on the sqlite harness and asserts qty 5 / reserved 0 (clamped, not negative), and PASSES. Pre-fix the `GREATEST()` call would error on sqlite with `no such function: GREATEST`, so the test discriminates.

### Regression audit of landed + non-affected paths

- Dispatch idempotency (order-row FOR UPDATE re-read at logistics.go:122-125, `priorDispatch` audit-count guard at 141-171, unfulfilled-remainder math at 156) is untouched and its tests (`TestDispatchShipment_DeductsOnlyUnfulfilledRemainder`, `TestDispatchShipment_SecondShipmentSkipsDeduction`, `TestDispatchShipment_ReReadsOrderInsideTx`) PASS.
- Order-stock transfer/FEFO (order_stock_allocate.go:666-727 upsert, 476-584 FEFO sync, 198-259 deductWarehouseStock reserved-drain) are untouched; the full `internal/repository/order` suite PASSES.
- Build (`go build ./...`), `go vet ./internal/repository/order/... ./internal/services/trade/...`, and `REDIS_URL= go test ./internal/repository/order/... ./internal/services/trade/... -count=1` are all green in my re-run.

### Edge cases (noted, not refuting)

1. The FEFO sync is fail-closed: if a batched product has `warehouse_stock` rows but the resolved default warehouse's row is insufficient (or no default/MAIN warehouse resolves), the dispatch now errors and rolls back instead of silently drifting. This is a deliberate behavior change, but it exactly matches the already-landed order-stock side, so the two paths are now *more* consistent, not less; no new asymmetry is introduced by this fixer.
2. Pre-existing latent limitation (NOT introduced here): both the order-stock and logistics FEFO paths deduct batches without filtering by `warehouse_id` and mirror onto the DEFAULT warehouse. A batch-tracked SKU stocked only in a non-default, non-MAIN warehouse could hit a default-warehouse resolution failure and error. This exists identically on the order-stock side; the fixer only aligned the logistics side with it, so it is not a regression attributable to this change.
3. Concern (2) in the fixer's report is confirmed benign: the `AND o.deleted_at IS NULL` hunks in the analytics SQL (order.go:843, 1073, 1112, 1143, 1184, 1244, 1289) are pre-existing working-tree changes from another agent, unrelated to G21-f, and the build is green with them present.

