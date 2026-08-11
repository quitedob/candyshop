# Changelog / Devlog — 2026-08-11 Subagent-Argue Fix Pass — Batch 2: Backend MEDIUM (G16–G29)

**Date:** 2026-08-11
**Source report:** [`docs/reports/code-review-audit-2026-08-11.md`](../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go 1.24 / Gin / GORM / PostgreSQL
**Trigger:** After the backend HIGH batch, the consolidated MEDIUM findings (G16–G29) were fixed via the same fixer + adversarial-arguer workflow.
**Result:** All 13 backend MEDIUM groups fixed and adversarially verified (CONFIRMED_FIXED). 6 groups (G20, G21, G23, G24, G27, G29) were initially REFUTED because the root cause lived in files outside the fixer's ownership; follow-up rounds with expanded ownership closed every one. Merged `go build ./...`, `go vet ./...`, `REDIS_URL= go test ./...` all green (note: one test initially failed under `go test ./...` from a stale compiled test binary in Go's build cache — resolved with `go clean -cache`, no source change). 19 devlog drafts written (details in [`drafts/`](drafts/)).

---

## 1. Process

Same per-bug fixer + adversarial-arguer workflow as Batch 1, with one addition: findings whose root cause spanned files outside the fixer's owned scope were re-dispatched in a follow-up workflow with the arguer-identified files added to ownership (G20+G24c merged because they share `cart.go`/`customer_orders_write.go`). Every follow-up passed its arguer on the first or second round.

## 2. How subagents were used (parallel, non-conflicting)

| Agent | Primary ownership | Finding(s) |
|-------|-------------------|------------|
| fix:G16 | `pkg/payment/{stripe,paypal}/...`, `handlers/system/paypal_webhook.go` | G16 |
| fix:G17 | `pkg/i18n/i18n.go`, `services/translation/translation.go`, `repository/translation/translation.go` | G17 |
| fix:G18 | `pkg/authsession/redis.go`, `middleware/auth.go` | G18 |
| fix:G19 | `pkg/ratelimit/{store,redis}.go` | G19 |
| fix:G20+24c → follow-up | `handlers/customer/{cart,customer_orders_write,checkout_pricing}.go`, `handlers/admin/{admin_orders,admin_orders_crud,order_helpers}.go` | G20 + G24c |
| fix:G21 → follow-up | `repository/order/order.go`, `services/trade/logistics_stock.go` | G21 |
| fix:G22 | `handlers/admin/{admin_analytics,admin_reports,admin_financial}.go` | G22 |
| fix:G23 → follow-up | `handlers/admin/admin_staff.go`, `services/user/user.go`, `repository/user/user.go` | G23 |
| fix:G24 → follow-up (a) | `database/database.go`, `models/{user,product}/*.go` | G24a |
| fix:G26 | `pkg/money/money.go`, `pkg/docxgen/docxgen.go` | G26 |
| fix:G27 → follow-up | `cmd/api/main.go`, `database/seed_catalog_i18n.go`, `database/seed.go` | G27 |
| fix:G28 | `pkg/workerlock/`, `pkg/realtime/redis.go` | G28 |
| fix:G29 → follow-up | `pkg/eino/graph/pipeline.go`, `pkg/eino/toolargs.go`, `models/trade/trade.go` | G29 |

## 3. Fixes in detail

### MEDIUM

#### G16 — Payment gateways: Stripe zero-decimal scaling + webhook replay; PayPal APPROVED-as-paid + Capture amount
- **Problem:** Stripe Capture/Refund hardcoded `amount*100` (zero-decimal currencies like JPY/KRW break); webhook event IDs never deduped (replays double-process); PayPal treated `CHECKOUT.ORDER.APPROVED` as payment-confirmed and dropped the amount in Capture.
- **Fix:** Currency-aware scaling (only ×100 for 2-decimal currencies); webhook event-ID dedupe; PayPal requires an actual capture before marking paid, and passes the correct amount through Capture. Regression tests (zero-decimal, replay, APPROVED-vs-captured).

#### G17 — i18n cache never invalidated on delete/import
- **Problem:** The in-memory i18n cache was additive-only; deleted/deactivated keys kept resolving (or went blank) until restart.
- **Fix:** `Delete` now removes the key (`DeleteFromCache`); `Import` rebuilds the cache (`ReplaceCache`). Regression tests (`TestDeleteInvalidatesCache`, `TestImportRebuildsCache`).

#### G18 — Access-token revocation silently defeated by Redis fail-open
- **Problem:** `middleware/auth.go` treated "session not found" (the post-revocation signal) as Redis-down and fell back to trusting the JWT, so logout/revoke/password-rotation had no effect.
- **Fix:** Distinguish revoked/missing session (deny) from a genuine Redis error (fail closed / explicit policy) — never trust the JWT on "not found". Regression test.

#### G19 — Rate-limit store unbounded; Redis rate-limit fails open
- **Problem:** In-memory store never evicted entries (memory leak); the Redis backend failed open on error (unlimited requests).
- **Fix:** Bounded in-memory store (TTL eviction) and a fail-closed / conservative Redis error path. Regression tests.

#### G20 — Compliance/MOQ/credit bypass; bulk Source empty; admin re-pricing unvalidated
- **Problem:** Bulk/requisition/reorder drafts bypassed compliance/MOQ/credit checks and were created with `Source=''`; credit was checked per-order only, never cumulatively at confirm; admin order edits accepted negative prices; MOQ not re-enforced on admin confirm.
- **Fix:** `Source=bulk` set on those flows; cumulative company credit enforced on ALL commit paths (customer confirm, cart checkout, both admin confirm endpoints); per-line price-list-min gate (split-line bypass closed); admin confirm re-enforces MOQ + price-list minimum; admin negative-price validation. Regression tests proven to fail pre-fix.

#### G21 — Stock transfer/FEFO/deduction bugs
- **Problem:** Transfer to a target with no `warehouse_stock` row dropped stock; FEFO warehouse-sync failures were log-only; `deductWarehouseStock` never decremented `Product.StockQuantity`; fulfillment + dispatch both deducted (double deduction); `FulfillmentRepository.Ship` had no row lock; draft-cleanup lost audit rows; dispatch FEFO never synced `warehouse_stock`.
- **Fix:** Upsert target stock row; propagate FEFO sync errors; decrement product stock; single-deduction guard (dispatch idempotency); row lock on Ship; propagate `writeStockAuditEntries` errors; dispatch FEFO mirrors `warehouse_stock`; replaced PG-only `GREATEST`. Regression tests.

#### G22 — Analytics raw SQL omits deleted_at IS NULL
- **Problem:** Seven analytics/report/financial queries counted soft-deleted orders.
- **Fix:** Added `deleted_at IS NULL` to every order-scoped raw-SQL query. Regression test over the query builders.

#### G23 — Supplier API key never disclosed; /admin/staff leaks customer PII
- **Problem:** Supplier self-registration never returned the API key (`json:"-"`); `/admin/staff` returned ALL users including customer PII via the unfiltered `GetUsers`/`FindAll`.
- **Fix:** API key returned on registration and retrievable owner-only; `/admin/staff` wired to the role-filtered `GetStaffUsers` (staff roles only, no customer PII). Regression tests.

#### G24 — Soft-delete unique index blocks reuse; product embeddings never written; tax/shipping fail open
- **Problem:** Deleted user emails / product slugs could not be reused (legacy full-table unique index persisted after migration); `product_embeddings` never written so semantic search was inert; the checkout pricing path swallowed tax/shipping DB errors and booked 0.
- **Fix:** Legacy unique indexes dropped on startup + partial index (`deleted_at IS NULL`) so reuse works; embeddings upsert + backfill wired; checkout pricing fails closed on rate-lookup errors (no silent 0 tax/shipping). Regression tests.

#### G26 — FormatMoney off-by-one-cent truncation; lineage hash %f collisions
- **Problem:** `FormatMoney` truncated instead of rounding; the lineage hash used `%f`, which collides across close float values.
- **Fix:** Round to cents with `math.Round`; collision-resistant lineage hash over the exact cents representation. Regression tests.

#### G27 — /ready reports 200 with DB down; make reset-db hangs; seed overwrites admin edits
- **Problem:** `/ready` ignored DB health; `make reset-db` started the server instead of seeding; category/product translation seeding reverted admin zh edits every restart.
- **Fix:** `/ready` pings the DB (503 when down); `reset-db` runs the seed CLI; product + category translation seeding is now fill-only / `alreadyLocalized`-gated so admin zh/scalar edits survive restarts. Regression tests (the category regression test fails on the pre-fix path).

#### G28 — Workerlock never renews; RedisBroadcaster startup can hang
- **Problem:** Workerlock TTL == interval with no renewal, so long cycles lost the lease (mutual exclusion broke); RedisBroadcaster startup blocked indefinitely if Redis was down.
- **Fix:** Lease-renewal heartbeat in workerlock (token CAS so a takeover lock is never deleted); bounded RedisBroadcaster startup (timeout / clear error instead of hang). Regression tests.

#### G29 — Eino HITL tools never complete; Generate concatenates tool content; doc-number collisions; CI pol not a port
- **Problem:** The HITL wrapper unconditionally interrupted read-only tools (track_shipment/translate_content/validate_lc_documents) with no resume path; `Client.Generate` concatenated raw tool-role content into answers; doc numbers used `UnixMilli()%100000` (collisions against a global unique index); the CI `pol` field held an Incoterms term.
- **Fix:** Read-only tools exempted from the interrupt (real completion path); tool-role messages filtered while preserving the ReturnDirectly terminal result (with a separator); doc-number generation now a per-process monotonic atomic counter (collision-free); CI `pol`/`pod` render from real `PortOfLoading`/`PortOfDestination` fields and BL/SLI/insurance agree. Regression tests proven to fail pre-fix.

## 4. Verification

| Check | Result |
|-------|--------|
| `go build ./...` (backend) | exit 0 |
| `go vet ./...` (backend) | exit 0 |
| `REDIS_URL= go test ./...` | all packages pass, 0 FAIL (after `go clean -cache`) |
| Arguer verdicts | 7 CONFIRMED_FIXED in round 1; 6 REFUTED → all closed in the follow-up (13/13) |
| Devlog drafts | 19 written to [`drafts/`](drafts/), each with the arguer's adversarial review |

## 5. Deliberately NOT changed (documented follow-ups)

- `graph/tool.go` (Eino) does not yet advertise `port_of_loading`/`port_of_destination` in the LLM tool schema — the ports apply via defaults until that follow-up.
- `float64` money fields remain end-to-end (only the two concrete defects — FormatMoney, lineage hash — were fixed; a decimal migration is a separate initiative).
- The `interface{}`→`any` and `QF1012`/`minmax` style warnings flagged by the linter are a separate style sweep, not correctness.
