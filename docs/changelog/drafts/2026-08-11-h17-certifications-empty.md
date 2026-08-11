# Changelog / Devlog — 2026-08-11 H17 — Admin certifications page always empty (bare array response)
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** frontend Nuxt 3 / Vue 3 / TypeScript / Tailwind
**Trigger:** Admin Certifications list rendered empty because the page assumed a `PaginatedResponse` while the backend returns a bare array.
**Result:** Fixed in `frontend/pages/admin/certifications/index.vue`; `npm --prefix frontend run typecheck` passes (exit 0).
---
## 1. Process
1. Read the owned page `frontend/pages/admin/certifications/index.vue` and the `useApi` composable to trace what `adminGetCertifications()` actually returns.
2. Verified the backend response shape in `backend/internal/handlers/admin/admin_certifications.go` (`AdminGetCertifications` serializes `[]modelsProduct.Certification` via `c.JSON(200, ...)` — a bare array, not a paginated envelope).
3. Confirmed the runtime mismatch, patched the data assignment in the page, and re-ran the frontend typecheck.

## 2. Fixes in detail
#### H17 — Admin certifications page always empty (PaginatedResponse vs bare array)
- **Problem:** `frontend/pages/admin/certifications/index.vue` (fetchCertifications) read `res.data || []`. The composable `frontend/composables/useApi.ts:952` types `adminGetCertifications` as `GET<PaginatedResponse<any>>('/admin/certifications', params)`, but at runtime `GET`/`fetchApi` return the raw `$fetch` body — the bare JSON array produced by `backend/internal/handlers/admin/admin_certifications.go:41` (`c.JSON(http.StatusOK, certifications)` where `certifications` is `[]modelsProduct.Certification`). A bare array has no `.data`, so `res.data` was always `undefined` and the table always fell back to `[]`.
- **Fix:** `frontend/pages/admin/certifications/index.vue:174-186` — rewrote the data assignment to accept the real bare-array payload and defensively the mislabeled `PaginatedResponse` shape: `Array.isArray(payload) ? payload : Array.isArray(payload.data) ? payload.data : []`. Rows now render from the seeded HACCP/ISO/Halal/FDA/BRC certifications. Because the composable typing is owned elsewhere, the page guards both shapes so it stays correct even if `useApi.ts` is fixed in parallel.

## 3. Verification
| Check | Result |
|-------|--------|
| `npm --prefix frontend run typecheck` | PASS (exit 0, no TS errors) |

## 4. Adversarial review (arguer)

Verdict: **CONFIRMED_FIXED**. I attempted to refute and could not construct a credible scenario where the bug still manifests or a regression is introduced.

**Root cause re-verified independently (not just from the report):**
- `backend/internal/handlers/admin/admin_certifications.go:41` — `c.JSON(http.StatusOK, certifications)` serializes `[]modelsProduct.Certification` as a bare JSON array. Confirmed against `backend/internal/models/product/product.go:366` (Certification struct).
- `frontend/composables/useApi.ts:641-644` (`GET`) → `fetchApi` (`useApi.ts:304-308`) returns `$fetch<T>` body directly. So `adminGetCertifications()` at runtime resolves to the raw array, never `{data: [...]}`. Old `res.data || []` → `undefined || []` → empty table. Bug reproduction is accurate.
- Fix (`frontend/pages/admin/certifications/index.vue:180-186`) branches on `Array.isArray(payload)` (real bare-array payload) with a defensive `payload.data` fallback (in case the mislabeled composable shape is ever real) and `[]` terminal. Correct for both shapes plus empty/null.

**Edge cases probed (all handled):**
- Empty backend array `[]` → `Array.isArray([])` → `certifications.value = []` → renders the `no_data` row (template line 33). Correct.
- Backend error path (`response.ErrorResp` → non-2xx JSON) → `$fetch` throws → caught at line 187-188 → `error` ref set → error row (template line 30). No crash.
- `null`/`undefined` `res` → `(res as any) ?? {}` → `payload = {}` → not array, no `.data` → `[]`. Safe.
- Only caller: `adminGetCertifications` appears in exactly one page (grep across `frontend/`) — no other consumer can be affected by the page-side change.
- Model JSON tags (`validUntil`, `certificateUrl`, `badgeUrl`, `abbreviation`, `issuer`, `name`) match template bindings exactly — rows render real values once the array is assigned.
- Create/edit/delete modal flow untouched; each mutation calls `fetchCertifications()` again, which now correctly re-renders.
- SSR: fetch is in `onMounted` (client-only), no SSR hydration mismatch.
- Typing: `(res as any)` cast is contained; `certifications` ref remains `any[]` matching template.

**Regressions on non-affected path:** none — only the page changed; `useApi.ts:952` (owned outside the fix) still mislabels the type, but runtime was and remains a bare array, and the page no longer depends on the wrong shape. If `useApi.ts` is later corrected to wrap in a paginated envelope, the defensive branch keeps the page working.

**Typecheck:** `npm --prefix frontend run typecheck` from repo root → exit 0, vue-tsc clean. (One prior run returned exit 2 with `TS6053: .nuxt/nuxt.d.ts not found` — that was a transient artifact of two concurrent `nuxt prepare` invocations racing on the `.nuxt` dir during this review; a clean single run passes.)

No concrete failing scenario found → CONFIRMED_FIXED.
