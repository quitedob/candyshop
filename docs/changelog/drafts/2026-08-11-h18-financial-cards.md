# Changelog / Devlog — 2026-08-11 H18 Admin financial summary cards always $0
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** frontend Nuxt 3 / Vue 3 / TypeScript / Tailwind
**Trigger:** Admin financial summary cards read `accountsReceivable` / `revenueThisMonth` / `overdueCount`, which the backend `GetFinancialOverview` never returns, so the cards always render $0 / 0.
**Result:** Cards remapped to the fields the backend actually returns; two unsupported cards removed; `vue-tsc --noEmit` passes.
---
## 1. Process
- Read `frontend/pages/admin/financial/index.vue` and traced the backend `GetFinancialOverview` handler (`backend/internal/handlers/admin/admin_financial.go`) plus the underlying service/repo methods to enumerate the exact response shape.
- Confirmed the response only contains `invoiceStats` (status→count map), `overdueAmount`, `confirmedPayments`, `totalRevenue`.
- Remapped the summary cards to real fields, reusing existing i18n keys (no locale files touched, per ownership).
## 2. Fixes in detail
#### H18 — Financial overview cards read fields backend never returns
- **Problem:** `frontend/pages/admin/financial/index.vue:29-41` read `overview.accountsReceivable`, `overview.revenueThisMonth`, and `overview.overdueCount`; `GetFinancialOverview` (backend/internal/handlers/admin/admin_financial.go:48-53) only returns `invoiceStats`, `overdueAmount`, `confirmedPayments`, `totalRevenue`. `overdueAmount` was the only field that matched, so 3 of 4 cards were permanently $0 / 0.
- **Fix:**
  - Card 1 relabeled to "Total Revenue" (`admin.dashboard.total_revenue`, present in all 10 locales) reading `overview?.totalRevenue`.
  - Card 2 "Overdue Amount" keeps reading `overview?.overdueAmount` (unchanged, already correct).
  - Card 3 "Outstanding Invoices" now reads a `outstandingInvoices` computed derived from `overview.invoiceStats.sent + overdue` (unpaid billed invoices), instead of the nonexistent `overview?.overdueCount`.
  - Removed the unsupported "Accounts Receivable" and "Revenue This Month" cards (no backing field) and tightened the grid to `lg:grid-cols-3`.
  - `frontend/pages/admin/financial/index.vue` (template lines 25-39, script `outstandingInvoices` computed).
## 3. Verification
| Check | Result |
|-------|--------|
| `npm --prefix frontend run typecheck` | PASS (exit 0, no `error TS`) |
| i18n keys used exist in all 10 locale files | PASS (`admin.dashboard.total_revenue`, `admin.financial.overdue_amount`, `admin.financial.outstanding_invoices`) |
## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.** I attempted to refute the fix and could not construct a scenario where the H18 symptom (summary cards rendering permanent $0/0 from fields the backend never returns) still manifests, nor a regression on the non-affected path.

**Verification performed:**
- Confirmed the backend response shape end-to-end: `GetFinancialOverview` (`backend/internal/handlers/admin/admin_financial.go:48-53`) returns only `invoiceStats`, `overdueAmount`, `confirmedPayments`, `totalRevenue`. `invoiceStats` is `map[status]count` built by `InvoiceRepository.Stats` (`backend/internal/repository/order/invoice.go:104-121`), keyed by stored invoice status; `models/order/invoice.go:18-22` defines statuses including `sent` and `overdue`. `totalRevenue` = `Order.SumSales` (all-time `SumTotalAmount`), `overdueAmount` = `Invoice.OverdueTotal`.
- Confirmed the frontend now reads only real fields: Card 1 `overview?.totalRevenue`, Card 2 `overview?.overdueAmount`, Card 3 `outstandingInvoices` computed = `invoiceStats.sent + invoiceStats.overdue` (guarded with `if (!stats) return 0` and per-key `|| 0`). `api.get` (`frontend/composables/useApi.ts`) returns the parsed JSON body directly, so the field names line up 1:1 with the handler keys.
- Ran `INTERNAL_API_BASE=http://localhost:8080 npm --prefix frontend run typecheck` from repo root: EXIT=0, no `error TS`. (Bare `npm --prefix frontend run typecheck` fails earlier at `nuxt.config.ts:29` because `INTERNAL_API_BASE`/`BACKEND_URL` is unset in this shell — an environment/config prerequisite, not a code error; with the env var set, as in normal `make dev`/`.env` runs, it passes.)
- Grepped the frontend for the removed fields: `accountsReceivable` and `overdueCount` have zero remaining references. `revenueThisMonth` survives only in `pages/admin/index.vue:46`, which is backed by a different endpoint (`dashboard.go:117` returns `revenueThisMonth`) — not the same class of bug, not in H18 scope.
- i18n keys used are present in all 9 real locale admin.json files (`frontend/i18n/{ar,en,id,ja,ko,ms,th,vi,zh}/admin.json`), under the correct namespaces: `dashboard.total_revenue`, `financial.overdue_amount`, `financial.outstanding_invoices`.

**Edge cases audited (all safe):**
- `overview` null/undefined → `?.` + `|| 0` render 0; page also gates on `loading`/`error` states.
- `invoiceStats` absent → computed returns 0.
- `stats.sent`/`stats.overdue` keys absent from the map → `|| 0`.
- Label honesty: Card 1 is labeled "Total Revenue" (all-time), deliberately NOT "this month", since `SumSales` is all-time — avoids a new misleading label.

**Non-blocking observations (do not refute):**
1. Report says "all 10 locale files"; there are 9 locale dirs plus a `composables` dir that is not a locale. Factual nit in the report only; all real locales have the keys.
2. `outstandingInvoices` counts `sent + overdue` statuses, while the table below defaults the `/admin/financial/outstanding-invoices` call to `status=sent` (backend default). If `overdue` is never written as a status, the card equals the `sent` count and may understate the visually-red rows computed client-side by `isOverdue()` (which flags sent invoices with past due dates). Harmless and robust either way — the card now shows real data, not permanent $0.
3. Backend `confirmedPayments` field remains unused by this page (the payment breakdown chart uses a separate endpoint) — pre-existing, not a regression.
4. Locale keys `accounts_receivable` and `revenue_this_month` are now dead but remain defined — harmless.
