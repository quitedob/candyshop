# Changelog / Devlog — 2026-08-11 H4 Shipment Timeline Ownership + Dispatch Optimistic Lock
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** Two shipment-logistics findings from the code-review audit: cross-tenant timeline IDOR and a non-atomic dispatch state transition.
**Result:** Both fixed with regression tests; owned packages build and test green. Full `go build ./...` currently blocked only by parallel agents' in-progress edits (services/translation refactor, orderintake), none of which touch the owned files.
---
## 1. Process
Confirmed both bugs by reading the code path: `customer_logistics.go` verified trade ownership but never tied `shipmentId` to the trade; `logistics.go` `DispatchShipment` read PENDING in memory then wrote DISPATCHED via an unconditional `tx.Save` with no `WHERE status` guard. Implemented an ownership-scoped timeline entry point and an atomic conditional status transition, mirroring `repository/trade/quotation_review.go UpdateStatus`. Added regression tests following the glebarez/sqlite in-memory harness used by `quotation_review_test.go`.
## 2. Fixes in detail
#### H4/MEDIUM-3 — Shipment-timeline IDOR (cross-tenant enumeration)
- **Problem:** `backend/internal/handlers/customer/customer_logistics.go:39-49` verified `trans.UserID != userID` but never checked that `shipmentId` belongs to that trade. `ShipmentEventRepository.FindByShipmentID` (shipment_event.go:21) scopes only by `shipment_id`, so any authenticated customer could read another tenant's timeline by guessing shipment IDs.
- **Fix:** Added sentinel `ErrShipmentNotInTransaction` and service method `LogisticsService.GetTransactionShipmentTimeline(ctx, transactionID, shipmentID)` (`services/trade/logistics.go`) that loads the shipment and rejects it unless `shipment.TransactionID == transactionID`. The customer handler now calls this scoped method and maps the sentinel to `403 forbidden`, `gorm.ErrRecordNotFound` to `404 shipment_not_found`, and anything else to `500`. The admin handler keeps the un-scoped `GetShipmentTimeline` (privileged path, unchanged).
#### H4/MEDIUM-4 — Unconditional shipment Save during dispatch (no optimistic lock)
- **Problem:** `repository/trade/shipment.go:62-64` `Update` is a full `Save` with no status guard; `DispatchShipment` (`services/trade/logistics.go:65-132`) read PENDING then `tx.Save(shipment)` wrote DISPATCHED. Two concurrent dispatches both passed the in-memory check, double-deducted stock and emitted duplicate dispatch events.
- **Fix:** Added `ShipmentRepository.Dispatch(ctx, db, id, dispatchedAt)` which runs `UPDATE shipment_trackings SET status='DISPATCHED', updated_at=? WHERE id=? AND status='PENDING'` (optimistic lock, mirroring `QuotationReviewRepository.UpdateStatus`). `DispatchShipment` now claims the transition as the first statement inside its transaction; on `RowsAffected==0` it returns an error so the whole transaction (stock deduction + event) rolls back. Only one concurrent dispatch wins.
## 3. Verification
| Check | Result |
|-------|--------|
| `go build ./...` (backend) | FAIL (only in parallel agents' in-progress files: services/translation, adminportalscope, admin_translations.go, orderintake) — none in owned files |
| `go build ./internal/repository/trade/... ./internal/services/trade/... ./internal/handlers/customer/...` | PASS (handler build confirmed across 3 retries while translation refactor churns) |
| `REDIS_URL= go test ./internal/repository/trade/... ./internal/services/trade/... ./internal/handlers/customer/...` | PASS — ok / ok / ok |
| New tests | `repository/trade/shipment_test.go` (DispatchOnlyWinsOnce, DispatchUnknownID), `services/trade/logistics_test.go` (GetTransactionShipmentTimeline enforces ownership) |
## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.**

Both H4 findings are closed at the root cause, not just symptomatically. I independently re-read the code path, audited edge cases and every caller of the changed functions, and re-ran the owned builds/tests.

### H4/MEDIUM-3 (IDOR) — closed
- The customer handler still does the trade-ownership check (`trans.UserID != userID`), and the new service entry point `GetTransactionShipmentTimeline(tradeID, shipmentID)` additionally rejects unless `shipment.TransactionID == tradeID`. Since the tradeID is itself owner-verified, an attacker cannot supply a foreign tradeID to bypass — the only timelines reachable are for shipments that (a) belong to a transaction the caller owns and (b) are keyed by a shipment ID the caller supplies. Cross-tenant enumeration of `shipment_event` rows is no longer possible via this endpoint.
- Error mapping is correct: `errors.Is(err, gorm.ErrRecordNotFound)` → 404 `shipment_not_found` (matches the sibling `verifyTradeShipmentOwnership` behavior); `ErrShipmentNotInTransaction` → 403; else 500. `FindByID` returns the gorm sentinel unwrapped, so `errors.Is` matches.
- No remaining customer IDOR: `CustomerNudgeShipment` / `CustomerUploadShipmentAttachment` already gate via `verifyTradeShipmentOwnership` (`shipment.TransactionID != tradeID` → 404), and `CustomerGetTradeShipments` / `CustomerGetTradeTimeline` are scoped by transactionID + ownership. The admin timeline (`GetShipmentTimeline`) stays un-scoped but is behind `RequireRole(admin, superadmin)` — privileged, unchanged, acceptable.
- No meaningful TOCTOU: `FindByID`-then-compare runs against the live row and the event read is keyed by the immutable shipmentID; no code path assigns `ShipmentTracking.TransactionID` after creation (grep of `repository/trade` for `.TransactionID =` returned nothing). The only deliberate behavioral change — a nonexistent shipment now returns 404 instead of 200 with an empty list — is a tightening, not a regression.

### H4/MEDIUM-4 (optimistic lock) — closed
- `ShipmentRepository.Dispatch` issues `UPDATE shipment_trackings SET status='DISPATCHED', updated_at=? WHERE id=? AND status='PENDING'` and returns `RowsAffected>0`. In `DispatchShipment` the transition is claimed as the FIRST statement inside the transaction; on RowsAffected==0 the tx returns an error so the stock deduction and dispatch-event insert are never executed and the whole transaction rolls back. Under PostgreSQL READ COMMITTED a second concurrent UPDATE blocks on the row lock, then re-evaluates the predicate against the committed row and matches zero rows — so exactly one dispatch wins, no double stock deduction, no duplicate dispatch events.
- The generic unconditional `Update` (Save) is intentionally left for ConfirmDelivery / AddTrackingEvent / Sync, which do not perform the PENDING→DISPATCHED dispatch transition — matching the finding's scope.
- I checked every `Status = "DISPATCHED"` write in the tree: the only other PENDING→DISPATCHED write is `syncFulfillmentTrackingToOrderAndTrade` (`handlers/admin/order_helpers.go:213-214`), which does NOT deduct stock or emit dispatch events, so it cannot reproduce the finding's impact (double-deduct/duplicate events), and the new lock in `DispatchShipment` also wins/loses correctly against it (either the sync commits DISPATCHED first and Dispatch loses, or Dispatch wins first and the sync's read sees non-PENDING).

### Verification (my own run)
- `go build ./internal/repository/trade/... ./internal/services/trade/... ./internal/handlers/customer/...` → PASS.
- `go build ./...` → PASS in my run (the earlier red state the fixer reported from parallel agents' in-flight edits had cleared; no owned-file errors).
- `REDIS_URL= go test -count=1 -run 'TestShipmentRepository|TestLogisticsService_GetTransactionShipmentTimeline' ./internal/repository/trade/... ./internal/services/trade/...` → ok / ok. Both new test files exist and exercise the intended behavior (first dispatch wins, second loses, status ends DISPATCHED; missing ID loses; ownership enforced).

### Minor non-blocking observations (no refutation)
1. The pre-transaction `shipment.Status != "PENDING"` fast-path in `DispatchShipment` is now redundant with the conditional UPDATE, but harmless.
2. `Dispatch` hardcodes "PENDING"/"DISPATCHED" string literals instead of a shared constant — consistent with the file's existing style, no correctness impact.
3. Cross-endpoint semantic nuance: the timeline endpoint returns 403 for a foreign shipment while `verifyTradeShipmentOwnership` returns 404 — both reject; a UX polish item only.

I could not construct a credible scenario where the H4 bugs still manifest or where the fix regresses an adjacent caller/contract.
