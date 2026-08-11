# Changelog / Devlog — 2026-08-11 Analytics raw SQL soft-delete filter (G22)

**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** G22 — analytics raw-SQL queries omit `o.deleted_at IS NULL`, so soft-deleted orders are counted in revenue / velocity / RFM / churn / P&L / replenishment.
**Result:** Fixed — all seven order-scoped raw-SQL analytics queries in `internal/repository/order/order.go` now filter `o.deleted_at IS NULL`; a source-guard regression test asserts the filter is present in each query. `go build ./...` clean; `REDIS_URL= go test ./internal/repository/order/... ./internal/handlers/admin/...` pass.

---
## 1. Process

Located the raw SQL by grepping for `Raw(` and `FROM orders` across the backend. The finding attributed the queries to the three admin handler files, but those are thin Gin handlers that delegate to `h.services.Order.*`; the actual SQL lives in `internal/repository/order/order.go` (it was moved there by an earlier "push aggregation into Postgres" refactor, per the M-4 comment on `TopProductsByRevenue`). Confirmed every GORM query-builder analytics method (`RevenueByMonth`, `OrderCountByMonth`, `RevenueByDay`, `SumTotalAmount`, `SumTotalAmountSince`) auto-appends `deleted_at IS NULL` via the model's `gorm.DeletedAt` and needs no change. Fixed the seven `Raw()` queries, then added a regression test.

## 2. Fixes in detail

#### G22 — Add `o.deleted_at IS NULL` to order-scoped raw-SQL analytics queries
- **Problem:** In `backend/internal/repository/order/order.go` the seven analytics `Raw()` queries scoped rows only by `o.status NOT IN ('cancelled','expired')` and a time window, omitting `o.deleted_at IS NULL`. Soft-deleted orders (compliance/retention scrubbing per C-9) were therefore still aggregated into revenue (TopProductsByRevenue), sales velocity (SalesVelocity), RFM (RFMAnalysis), churn (CustomerChurn), inventory-health demand (InventoryHealth), P&L (ProfitLossByPeriod) and replenishment demand (ReplenishmentSuggestions). The outer `p.deleted_at IS NULL` (products) and `u.deleted_at IS NULL` (users) filters present in some queries did not protect the joined/subqueried `orders o` rows.
- **Fix:** Added `o.deleted_at IS NULL` to each orders predicate: `order.go:840` (TopProductsByRevenue), `:1069` (SalesVelocity), `:1109` (RFMAnalysis, inside the join ON clause), `:1140` (CustomerChurn), `:1181` (InventoryHealth), `:1241` (ProfitLossByPeriod), `:1286` (ReplenishmentSuggestions). InventoryHealth and ReplenishmentSuggestions share an identical orders subquery, fixed via a single `replace_all` edit. No change needed in the owned handler files (`admin_analytics.go`, `admin_reports.go`, `admin_financial.go`) — they contain no SQL.
- **Note (finding discrepancy):** the report named "revenue-trends" as affected, but `RevenueByMonth` (the revenue-trends source) is a GORM query-builder call and already gets the soft-delete filter appended automatically. The seven buggy raw-SQL queries are actually top-products, sales-velocity, RFM, churn, inventory-health, profit-loss, replenishment — all seven fixed.

## 3. Verification

| Check | Result |
|-------|--------|
| `go build ./...` (backend) | PASS |
| `go vet ./internal/repository/order/` | PASS |
| `REDIS_URL= go test -count=1 ./internal/repository/order/...` | PASS |
| `REDIS_URL= go test -count=1 ./internal/handlers/admin/...` | PASS |
| `go test -run TestAnalyticsRawSQLFiltersSoftDeletedOrders -v` | PASS (all 7 subtests) |
| Regression test fails on pre-fix path | Confirmed — removing one filter makes the subtest FAIL |
| Manual read of all 7 raw-SQL queries | All contain `o.deleted_at IS NULL` (order.go:840,1069,1109,1140,1181,1241,1286) |

## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.** I attempted to refute and could not construct a credible scenario where soft-deleted orders are still counted in the seven order-scoped analytics queries.

**What I independently verified:**
- Read all seven `Raw()` blocks in `backend/internal/repository/order/order.go` on disk — each contains `o.deleted_at IS NULL` in the orders WHERE/ON predicate (order.go:840, 1069, 1109, 1140, 1181, 1241, 1286). The predicate is placed on the `orders o` rows themselves, not merely the outer products (`p.deleted_at`) or users (`u.deleted_at`) filters — the exact gap G22 named.
- `git diff` confirms exactly 7 additive lines (one per query); the file I restored after an accidental sed corruption matched the fixer's intended state (7 added `o.deleted_at IS NULL` lines, verified against the diff).
- `go build ./...` PASS; `REDIS_URL= go test -count=1 ./internal/repository/order/...` PASS; `./internal/handlers/admin/...` PASS; `TestAnalyticsRawSQLFiltersSoftDeletedOrders` PASS with all 7 subtests.
- Order model uses `gorm.DeletedAt` (models/order/order.go:235), so `deleted_at IS NULL` matches the actual soft-delete storage convention.
- **Fixer's finding correction is sound.** The finding claimed "revenue-trends" was one of the seven buggy queries. I empirically confirmed via a GORM DryRun probe that `Model(&Order{})` + `Scan()` emits `WHERE \`orders\`.\`deleted_at\` IS NULL` automatically — so RevenueByMonth (revenue-trends), OrderCountByMonth, RevenueByDay, SumTotalAmount, SumTotalAmountSince are NOT buggy. The real seventh raw-SQL query was TopProductsByRevenue, which the fixer correctly fixed. Net: all raw-SQL order-analytics aggregation is filtered; all GORM-built analytics were never affected.
- Swept the whole backend for `.Raw(` and `FROM orders` / `JOIN orders` — the only order-scoped raw-SQL analytics are the seven fixed ones. product.go:212 is product-scoped; inquiry ConversionByMonth is GORM query-builder over `inquiries`; migrate_* Raw calls are backfill scripts, not analytics.

**Edge cases / residual concerns (non-blocking):**
1. **Regression test is a source-guard, not an executable SQL test.** It reads `order.go` and asserts the literal string `o.deleted_at IS NULL` appears in each `Raw(` block. This is forced: the queries use PG-only functions (`jsonb_array_elements`, `date_trunc`, `NOW() - INTERVAL`, `EXTRACT`) that cannot execute on the SQLite harness. I manually confirmed each occurrence is semantically placed in the orders predicate, so no false pass. The test will catch removal of the filter but is brittle to future refactors (renaming funcs, splitting Raw blocks, re-alias). Acceptable given the finding's own fallback, but worth noting it guards the string, not the SQL semantics.
2. **Ownership deviation (documented by fixer, corroborated):** the bug's true home is `repository/order/order.go`, outside the finding's owned handler-file list. The three handler files contain no SQL — they are thin delegators (`h.services.Order.*`). The fix landed in the correct layer. `order.go` was not mid-edit by another agent per git status at fix time; as arguer I confirm it is the right and only place the change belongs.
3. **GORM `Joins` to orders in coupon.go:132, fulfillment.go:159, invoice.go:54, order_message.go:65 are record reads, not analytics aggregations** — they join a specific order to fetch a coupon-usage count / message / invoice / fulfillment row, and do not roll up order revenue/velocity/RFM/churn/P&L. Out of scope for G22; correctly flagged by the fixer as a separate concern if the audit intends them.
4. **No regressions on the non-affected path:** added predicates are purely restrictive; column sets, grouping, parameter order, and response shapes are unchanged. All existing order + admin handler tests pass. No concurrency, i18n, decimal/currency, or cross-tenant impact (queries are admin-scoped aggregations; the new predicate only removes rows).

**Bottom line:** I could not produce a concrete failing scenario. The root cause is closed for all seven raw-SQL order-analytics queries, revenue-trends is confirmed never-affected, and the regression guard fails on the pre-fix path.
