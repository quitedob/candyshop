# Changelog / Devlog — 2026-08-11 Subagent-Argue Fix Pass — Batch 3: Frontend (H16–H24, M1–M4, LOW)

**Date:** 2026-08-11
**Source report:** [`docs/reports/code-review-audit-2026-08-11.md`](../reports/code-review-audit-2026-08-11.md)
**Scope:** frontend Nuxt 3 / Vue 3 / TypeScript / Tailwind (one backend defense group)
**Trigger:** The frontend HIGH (H16–H24), frontend MEDIUM (M1–M4) and LOW/INFO findings from the audit were fixed via the same fixer + adversarial-arguer workflow.
**Result:** All 14 frontend groups fixed and adversarially verified (CONFIRMED_FIXED). The LOW sweep was initially REFUTED on its cases/products list-clamp item (the list pages were outside the fixer's owned set); a follow-up with expanded ownership closed it. `npm --prefix frontend run typecheck` exit 0 and `npm run build` (with `BACKEND_URL` set) exit 0. 14 devlog drafts written (details in [`drafts/`](drafts/)).

---

## 1. Process

Same per-bug fixer + adversarial-arguer workflow as the backend batches, with file-ownership de-overlapped per page/dir (invoice status map folded into the invoices group, trade-doc statuses into trades-detail, SSR-config into the config group, OEM into one group). One finding (LOW) required a follow-up round for the list pages outside the original ownership.

## 2. How subagents were used (parallel, non-conflicting)

| Agent | Primary ownership | Finding(s) |
|-------|-------------------|------------|
| fix:H16 | `composables/useSanitizer.ts` (+ `package.json` dep) | H16 |
| fix:H17 | `pages/admin/certifications/index.vue` | H17 |
| fix:H18 | `pages/admin/financial/index.vue` | H18 |
| fix:H19 | `pages/admin/invoices/index.vue` | H19 |
| fix:H20 | `pages/customer/orders/quick.vue`, `pages/customer/products/[id].vue` | H20 |
| fix:H21 | `pages/admin/trades/[id].vue` | H21 |
| fix:H22 | `pages/admin/trades/index.vue` | H22 |
| fix:H23 | `pages/customer/orders/[id].vue` | H23 |
| fix:H24 | `pkg/storage/validate.go`, `handlers/admin/admin_ai_translate.go`, `pkg/eino/graph/pipeline.go` (backend) | H24 |
| fix:M1 | `composables/useApi.ts`, `plugins/auth-auto-refresh.client.ts`, `plugins/init-auth.client.ts`, `pages/auth/login.vue` | M1 |
| fix:M2 | `composables/useToast.js`, `nuxt.config.ts` | M2 |
| fix:M3 | `pages/admin/{oem,oem-projects,inquiries,analytics,webhooks}.vue`, `pages/customer/{quotes,dashboard}.vue` | M3 |
| fix:M4 | `pages/admin/{ai,inventory,xlsx,orders,content}/` | M4 |
| fix:LOW → follow-up | `pages/auth/{verify-email,forgot-password}.vue`, `pages/{contact,faq}.vue`, `pages/blog/index.vue`, `pages/cases-clients/[slug].vue`, `components/product/ProductCard.vue` → follow-up `pages/cases-clients/index.vue`, `pages/customer/{inquiries,oem-projects}/new.vue` | LOW |

## 3. Fixes in detail

### HIGH — security

#### H16 — useSanitizer is a silent no-op during SSR → raw HTML in first paint (XSS)
- **Problem:** `DOMPurify()` without `window` returns an unsupported factory whose `sanitize()` returns input unchanged, so the SSR'd blog page rendered raw stored HTML before hydration.
- **Fix:** `useSanitizer` now branches on `import.meta.server` — server path uses `sanitize-html` (new dependency) with a shared tag/attr allowlist + scheme whitelist (blocks `javascript:`/`data:`/`ftp:`), client path keeps DOMPurify. Verified against a 24-vector XSS battery.

#### H24 — Stored HTML/XSS defense-in-depth (backend)
- **Problem:** AI trade-document HTML persisted raw, AI-translated fields merged unsanitized, and upload validation allowed `.svg` served inline.
- **Fix:** Sanitize AI-generated trade-document HTML on write; sanitize AI-translated product fields before merging into Translations; harden upload validation (reject/neutralize `.svg`, no permissive fallback). Go tests added.

### HIGH — dead flows

#### H17 — Admin certifications page always empty
- **Fix:** Aligned the page with the backend's bare-array response (was typed as `PaginatedResponse`).

#### H18 — Admin financial summary cards always $0
- **Fix:** Mapped the cards to the fields `GetFinancialOverview` actually returns.

#### H19 — Invoice create always 400
- **Fix:** Send the backend-bound `amount` (not `totalAmount`) and corrected the status map rendering.

#### H20 — Quick-order & product-page order requests always fail
- **Fix:** Order submit now includes the required shipping `street`/`city`/`country` on both flows.

#### H21 — Admin trades rich-doc create/edits always 400
- **Fix:** Numeric fields coerced to numbers before submit (no empty-string floats); trade-doc statuses mapped to the backend's uppercase enum.

#### H22 — Admin trades XLSX export always fails (Bearer null)
- **Fix:** Export now goes through `useApi` (session-authenticated) instead of a raw `fetch` with the deprecated `useAuth().token`.

#### H23 — Customer order detail confirm/cancel dead-ends on bulk/requisition/reorder orders
- **Fix:** Confirm gating now matches the source values the backend emits (including `bulk` for bulk/requisition/reorder), so a `pending_confirmation` order always has an actionable control.

### MEDIUM — reliability / security / contracts

#### M1 — Refresh-token race, redirect open-redirect, remember-me no-op, useApi TypeError
- **Fix:** Shared in-flight refresh promise (coalesces concurrent triggers); `redirect` sanitized to relative paths only; remember-me wired or removed honestly; error normalization robust to non-object bodies.

#### M2 — useToast overlapping hide timers; CSP unsafe-inline; SSR config hazard
- **Fix:** Per-toast hide timers; CSP `script-src` drops `unsafe-inline` (theme script pinned by SHA-256) and `connect-src` narrowed to same-origin + the configured API origin; `INTERNAL_API_BASE` derived from `BACKEND_URL` with a build-time side-effect so SSR fetches don't silently hit localhost.

#### M3 — Contract mismatches (OEM slug-keyed route, compliance field, analytics names, webhook secret, enums, dashboard)
- **Fix:** Aligned 8 frontend pages to the verified backend contracts: OEM edit sends slug; removed the dead AI-compliance section; top-products renders `productId`; removed the never-serialized webhook secret; quotes use `inquiry_status`; customer dashboard now calls `/user/dashboard` + `/user/quotes` + `/user/notifications`; OEM adminNotes via `useApi`; `catch{}` blocks surface toast errors.

#### M4 — Admin UI dead features (HITL approve/reject, Stop, filters, template, orders edit, cancel refund, double-submit)
- **Fix:** HITL Approve/Reject call the real persisted quotation-review endpoints; Stop enabled exactly while streaming; inventory filters page the full dataset; download-template produces a real `.xlsx` (verified opens with excelize); order edit preserves `fulfilledQuantity`/`shippedQuantity`; canceling a paid order sends `refund:true`; content save guarded against double-submit.

### LOW / INFO

#### LOW — verify-email token guard, forgot-password, contact/faq pages, blog/cases clamp, ProductCard a11y
- **Fix:** Type-guard the verify-email token; forgot-password only sets success on a real response; FAQ guards fixed (contact off-by-one, faq blank-at-15); blog list pages to the real page size; **follow-up**: cases-clients list and the two customer product pickers (`inquiries/new.vue`, `oem-projects/new.vue`) now page to the backend clamp sizes (50/100) and concatenate to `totalPages`, so >50 cases / >100 products are no longer silently hidden; ProductCard nested-button/focus a11y fixed.

## 4. Verification

| Check | Result |
|-------|--------|
| `npm --prefix frontend run typecheck` | exit 0, 0 `error TS` |
| `npm run build` (with `BACKEND_URL` set) | exit 0 (Client + Server built) |
| Arguer verdicts | 13 CONFIRMED_FIXED in round 1; LOW CONFIRMED_FIXED after the follow-up (14/14) |
| Devlog drafts | 14 written to [`drafts/`](drafts/), each with the arguer's adversarial review |

## 5. Deliberately NOT changed (documented follow-ups)

- The M2 theme-cookie regex residual (the template literal drops `\s`) is a separate a11y-first-paint ticket.
- The inquiry-status locale keys for the backend's `pending_confirmation`/`confirmed` statuses are a locale-content gap (DB-backed translations).
- The inventory full-dataset filter is capped at 50 pages (~5000 SKUs) — a safety bound for very large catalogs.
