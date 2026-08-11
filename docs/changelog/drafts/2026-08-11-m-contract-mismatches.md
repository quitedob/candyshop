# Changelog / Devlog — 2026-08-11 M3 — Contract mismatches: OEM slug-keyed route, compliance field, analytics names, webhook secret, enums

**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** frontend Nuxt 3 / Vue 3 / TypeScript / Tailwind
**Trigger:** Frontend reads/writes fields, routes and enums the backend never exposes, and swallows OEM mutation failures.
**Result:** Frontend aligned to the verified backend contracts for OEM solution edit (slug-keyed), inquiry detail (no aiComplianceCheck), top-products report (`productId`/`revenue` only), webhook create (no secret returned), customer dashboard (real `/user/dashboard` + quote/notification counts), quotes status enum (`inquiry_status`); OEM flow/solution `catch{}` now surfaces errors. `npm --prefix frontend run typecheck` passes clean (exit 0).
---
## 1. Process
Read each backend handler/service/repo/model to confirm the true contract before touching the frontend (admin_oem.go, admin_reports.go → order.go:828, admin_webhooks.go → webhook.go:32, customer_extensions.go CustomerGetDashboard/CustomerGetQuotes/CustomerGetNotifications, admin_audit.go inquiryDetailResponse, negotiation.go, inquiry.go model). Then applied fixes only inside the M3 owned frontend files and re-ran `npm --prefix frontend run typecheck`.
## 2. Fixes in detail
#### M3-1 — OEM solution edit sent a UUID to a slug-keyed route
- **Problem:** `frontend/pages/admin/oem.vue:366` set `solutionForm.id = s.id || s.slug`. The backend update route is slug-keyed — `AdminUpdateOEMSolution` resolves via `GetSolutionBySlug(c.Param("id"))` (admin_oem.go:344). Sending the UUID always 404'd, so editing an existing OEM solution silently failed.
- **Fix:** `openSolutionForm` now stores `solutionForm.id = s.slug || s.id` so `adminUpdateOemSolution` sends the slug. Delete stays id-keyed (`s.id`), matching `DeleteSolution(..., "id = ?")` in repository/oem/oem.go:100.

#### M3-2 — Admin inquiry detail expected an AI compliance field the backend never produces
- **Problem:** `frontend/pages/admin/inquiries/[id].vue:121-153` rendered a whole "AI Compliance Check Results" card gated on `inquiry.aiComplianceCheck`. Grep of the backend found no `aiComplianceCheck` anywhere; the Inquiry model (models/product/product.go:261) has no such field. The card was dead code that never rendered.
- **Fix:** Removed the dead compliance section. Kept the `statusHistory` "Activity Log" section — that one IS backed by `inquiryDetailResponse.StatusHistory` (admin_audit.go:86-90, built from ActivityLog).

#### M3-3 — Analytics top-products read name/quantity the backend doesn't return
- **Problem:** `frontend/pages/admin/analytics/index.vue:117-119` rendered `product.name` and `product.quantity`. The report returns only `{ productId, revenue }` (`TopProductsByRevenue`, repository/order/order.go:828-860), so the name always showed `-` and quantity always `0`.
- **Fix:** Product cell now shows `product.productId` (mono, matching the id-only payload) and the quantity column was removed; revenue column unchanged.

#### M3-4 — Webhook create expected a secret the backend never returns
- **Problem:** `frontend/pages/admin/webhooks/index.vue` read `res.secret` after create and showed a `secret_created` notice. `WebhookConfig.Secret` is `json:"-"` (models/order/webhook.go:32) — the secret is generated server-side but never serialized, so `res.secret` was always empty and the notice could never appear.
- **Fix:** Removed the `createdSecret` ref and the notice block; create now posts without expecting a secret, closes the modal and refetches. Backend secret handling is unchanged (deliberately hidden).

#### M3-5 — Customer dashboard never called /user/dashboard; quotes/notifications hardcoded '—'
- **Problem:** `frontend/pages/customer/dashboard.vue` called `/user/orders?limit=5` and rendered hardcoded `'—'` for the quotes and notifications KPI cards, even though the backend exposes `GET /user/dashboard` (customer_extensions.go:32, returns `{orders, inquiries, summary}`), `GET /user/quotes` and `GET /user/notifications` (returns `unreadCount`).
- **Fix:** `loadData` now resolves `/user/dashboard`, `/user/quotes?page=1&limit=50`, `/user/notifications` via `Promise.allSettled`; orders come from `dash.orders.data`, quote KPI from the quote list length (the endpoint's `pagination.total` is the raw inquiry count, so the quote rows length is the honest count), notification KPI from `unreadCount`. Only a rejected `/user/dashboard` fails the whole page.

#### M3-6 — Quotes page labelled inquiry statuses with the negotiation_status enum
- **Problem:** `frontend/pages/customer/quotes.vue:32` used `enumLabel('negotiation_status', quote.status)` but `/user/quotes` returns Inquiry records whose statuses are inquiry states (pending/quoted/negotiating/won/lost). NegotiationOffer statuses (pending/accepted/rejected/expired) never matched, so labels fell back to the unknown/raw value.
- **Fix:** Changed to `enumLabel('inquiry_status', quote.status)`.

#### M3-7 — OEM adminNotes update bypassed useApi via a raw $fetch
- **Problem:** `frontend/pages/admin/oem-projects/[id].vue` issued a raw `$fetch` PUT for `adminNotes`, skipping the useApi composable's 401-refresh and error normalisation.
- **Fix:** Replaced with `adminUpdateOemProject(id, { adminNotes })` from useApi (backend `AdminUpdateOEMProject` already accepts `adminNotes`); removed the now-unused `useRuntimeConfig`/`baseURL`.

#### M3-8 — OEM flow/solution mutations swallowed errors
- **Problem:** `frontend/pages/admin/oem.vue` had `catch {}` on `saveFlow`, `deleteFlow`, `saveSolution`, `deleteSolution` — failures were invisible to the admin.
- **Fix:** Each catch now surfaces `toast.error(e?.message || t('errors.api.save_failed'|'delete_failed'))`; success paths still reload the list.
## 3. Verification
| Check | Result |
|-------|--------|
| `npm --prefix frontend run typecheck` | Pass (exit 0; no `error TS` in any owned file) |
| Backend contract re-read for each fix | Done (handlers/services/repos/models cited above) |
| Leftover symbol scan (`createdSecret`, `baseURL`, `aiComplianceCheck`, `product.name/quantity`, `negotiation_status`) | None remaining in owned files |
## 4. Adversarial review (arguer)
Verdict: **CONFIRMED_FIXED**. I attempted to refute each sub-fix by re-deriving the backend contract independently (handlers/services/repos/models) and by running `npm --prefix frontend run typecheck`. No owned file fails typecheck; every claimed contract was verified against the on-disk backend.

Independently re-verified contracts:
- **M3-1** — `admin_oem.go:297` (GET) and `:344` (PUT) resolve via `GetSolutionBySlug`; `AdminDeleteOEMSolution` (`:404`) and `repository/oem/oem.go:100` delete by `id = ?`. The list endpoint returns full `OEMSolution` models (product.go:328) so both `id` and `slug` are present. `openSolutionForm` stores `s.slug` → PUT hits the slug route; delete passes `s.id` → id route. Both directions correct.
- **M3-2** — no `aiComplianceCheck` exists anywhere in the backend; `admin_audit.go:87-90` `inquiryDetailResponse` embeds the Inquiry model and adds `StatusHistory[]` with `{status,note,timestamp}`, matching the retained Activity Log template. `adminGetInquiry` populates it (admin_extensions.go:259-264).
- **M3-3** — `repository/order/order.go:838-856` returns only `{productId, revenue}`; analytics is the only `top-products` consumer (grep), so no other page still reads name/quantity.
- **M3-4** — `models/order/webhook.go:32` `Secret json:"-"`; `AdminCreateWebhook` returns `c.JSON(http.StatusCreated, cfg)` (admin_webhooks.go:91) so the secret is never serialized. The removed `createdSecret` notice was provably dead code; behavior unchanged (secret was never obtainable before or after — pre-existing backend gap).
- **M3-5** — `CustomerGetDashboard` returns `{orders: PaginatedResponse{data,pagination}, inquiries, summary}`; `/user/quotes` returns `{data,pagination}`; `/user/notifications` returns `{data,total,unreadCount}`. Frontend reads `dash.orders?.data`, `q?.data?.length`, `notifs?.unreadCount` — all match; `fetchApi` returns the raw body with no wrapper, so no unwrap mismatch.
- **M3-6** — `/user/quotes` filters `inquiry.Status == "quoted" || QuotedAmount > 0`; statuses are Inquiry values, so `inquiry_status` is the correct group (the old `negotiation_status` never matched any Inquiry status).
- **M3-7** — `AdminUpdateOEMProject` accepts `adminNotes` (admin_oem.go:64, 122-123); `adminUpdateOemProject` (useApi.ts:740) now used; `useRuntimeConfig`/`baseURL` removed from the file.
- **M3-8** — all four OEM mutation paths now surface `toast.error`; success paths still reload.

Residual edge cases (none re-open the original mismatch, none regress):
1. **Locale enum gap (out of scope, pre-existing)**: `inquiry_status` in all 9 locale files (e.g. `frontend/i18n/en/common.json:174`) lacks keys for `confirmed` and `pending_confirmation`, which ARE valid Inquiry statuses (`services/inquiry/inquiry.go:172-182`). An inquiry in either status renders "Unknown" on quotes.vue, the admin inquiry detail, and the statusHistory activity log. The quotes.vue fix is still strictly more correct than before (was guaranteed-Unknown for every status via `negotiation_status`); the gap predates M3 (inquiry detail already used `inquiry_status`). Fixer's concern #1 is accurate but out of ownership (locale files not in the M3 file set).
2. **Dashboard quote-count undercount for >50 quotes** (fixer concern #2): the KPI uses `/user/quotes` `data.length` (cap 50) because the endpoint's `pagination.total` is the raw inquiry count, not the quote count. Minor undercount only for heavy users; honest count is 0 when the endpoint fails (graceful, vs. the old hardcoded `—`).
3. **`customer/oem-projects/new.vue` fallback** (fixer concern #3, left unchanged): `applyQueryToForm` calls `getProduct(productId)` (a UUID) against the slug-keyed public `/products/:slug` route when the product isn't in the 200-item prefetch list. It fails silently and only degrades the prefill preview — the create payload (`customerCreateOemProject`) never sends `productId`, so no dangling-id is persisted. Pre-existing, outside the M3 admin-edit scenario.

Typecheck note: one run produced 3 `Cannot find name 'useSanitizer'` errors in `pages/admin/ai/index.vue`, `pages/admin/content/index.vue`, `pages/blog/[slug].vue` — all from the in-flight H24 task's concurrent `nuxt prepare` (stale `.nuxt` auto-import cache), none in M3-owned files. Two subsequent clean runs exit 0 with zero `error TS` lines.
