# Changelog / Devlog — 2026-08-11 M1 — Auth session hardening (refresh race, open-redirect, remember-me no-op, useApi TypeError)
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** frontend Nuxt 3 / Vue 3 / TypeScript / Tailwind
**Trigger:** Concurrent `/auth/refresh` calls rotate each other out (spurious forced logouts), `?redirect=` is an open redirect, remember-me is a silent no-op, and `useApi` throws a TypeError on non-object error bodies.
**Result:** Fixed in `frontend/composables/useApi.ts`, `frontend/plugins/auth-auto-refresh.client.ts`, `frontend/pages/auth/login.vue`; `npm --prefix frontend run typecheck` passes (exit 0).
---
## 1. Process
1. Read all four owned files plus `frontend/composables/useAuth.ts` to trace the refresh flow (`refreshAccessToken`, `initAuth`, `login`) and the exact 401-retry path in `useApi.fetchApi`.
2. Confirmed against `backend/internal/handlers/auth/auth.go:186-190` that the `Login` handler binds only `email`/`password` and never reads a `remember` field, so the checkbox was provably a no-op.
3. Applied the four fixes (module-scoped shared refresh, redirect sanitizer, honest remember-me removal, robust error normalization) and re-ran the frontend typecheck.

## 2. Fixes in detail
#### M1(a) — Refresh-token race: per-instance refreshPromise + uncoalesced plugin triggers
- **Problem:** `frontend/composables/useApi.ts:277` (old) held `let refreshPromise` inside `useApi()`, so each component's `useApi()` instance had its own in-flight refresh, while `frontend/plugins/auth-auto-refresh.client.ts` called `useAuth().refreshAccessToken()` directly on its interval and on every focus/visibilitychange. Two overlapping `/auth/refresh` calls each carry the same refresh-token cookie; the backend rotates the token on the first success, so the losing call fails and trips `auth.logout()` -> spurious forced logouts.
- **Fix:** `frontend/composables/useApi.ts:267-296` — hoisted a module-scoped `sharedRefreshPromise` and exported `refreshAuthSession()`, which coalesces every caller into a single in-flight refresh (`.finally()` clears it on settle). On `import.meta.server` it bypasses the shared promise because SSR renders are isolated per request and must not share cookies. `frontend/plugins/auth-auto-refresh.client.ts:13-21` now funnels the interval/focus/visibility triggers through `refreshAuthSession()`; `frontend/composables/useApi.ts:346-358` uses it for the 401 retry too. `init-auth.client.ts` is untouched (single boot-time flow, no concurrency).

#### M1(b) — Open redirect via `?redirect=//evil.com`
- **Problem:** `frontend/pages/auth/login.vue:143-150` (old) accepted any value starting with `/`, so `?redirect=//evil.com` (protocol-relative) and `/\evil.com` passed `navigateTo()` and redirected off-site.
- **Fix:** `frontend/pages/auth/login.vue:146-148` — added `isSafeRedirect(target)`: must start with `/`, must NOT start with `//`, and must not contain a backslash. Non-matching values fall through to the normal role-based redirect (admin/superadmin -> `/admin`, else `/customer/dashboard`).

#### M1(c) — remember-me is a silent no-op
- **Problem:** `frontend/pages/auth/login.vue:66-69,115` sent `remember` in the login body, but `backend/internal/handlers/auth/auth.go:186-190` binds only `email`/`password` and never reads it. With HttpOnly cookies the frontend cannot control refresh-cookie lifetime, so the control was pure theater.
- **Fix:** `frontend/pages/auth/login.vue` — removed the checkbox (kept the forgot-password link) and dropped `remember: false` from the form. Full end-to-end wiring would require a backend change (accept `remember`, set the refresh cookie `Max-Age` accordingly) which is outside this frontend-only file scope, so honest removal was chosen per the report's sanctioned option. The `auth.remember_me` i18n keys in the locale files are now unused but left intact (files outside ownership).

#### M1(d) — useApi TypeError on non-object error bodies
- **Problem:** `frontend/composables/useApi.ts:344` (old) ran `'details' in error.data` unguarded; when `error.data` was a non-object (e.g. a proxy error page or plain-text body), `in` throws `TypeError: Cannot use 'in' operator...`, masking the real error.
- **Fix:** `frontend/composables/useApi.ts:365-377` — cast `error.data` to `unknown` and guard with `typeof errorData === 'object' && errorData !== null` before reading `message` / `details`. Non-object bodies now normalize to the default i18n message instead of throwing.

## 3. Verification
| Check | Result |
|-------|--------|
| `npm --prefix frontend run typecheck` | PASS (exit 0, no TS errors) |
| `grep remember backend/internal` | Only `paypal_webhook.go` comment — no login handler reads `remember` (backend binding confirmed at `auth.go:187-188`) |
| `grep remember frontend/pages/auth/login.vue` | No matches (checkbox + field removed) |

## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.** Independent audit of `useApi.ts`, `auth-auto-refresh.client.ts`, `init-auth.client.ts`, and `login.vue`, plus re-run of `npm --prefix frontend run typecheck` (exit 0, no TS errors) from repo root.

### M1(a) refresh race — fixed
All client-side `/auth/refresh` triggers now funnel through the single module-scoped `sharedRefreshPromise` (`useApi.ts:267-296`): the plugin's interval/focus/visibility handlers (`auth-auto-refresh.client.ts:20`), the 401-retry path (`useApi.ts:347`). `refreshAccessToken` (`useAuth.ts:81-89`) never rejects — it catches and returns `false` — so `.finally()` reliably clears the promise and no unhandled-rejection path exists. Two concrete non-regressions I probed: (1) the boot-time `initAuth` refresh at `useAuth.ts:107` still calls `refreshAccessToken()` directly and bypasses the shared promise, but it runs once at startup while the plugin interval is `>= 60s` (`refreshMs` floor) and focus events at boot merely cause a second refresh whose `false` result is ignored by the plugin (no `auth.logout()`), so no spurious forced logout can arise from it; (2) SSR (`import.meta.server`) short-circuits before touching `sharedRefreshPromise`, so no cross-request cookie leakage via a module-level promise.

### M1(b) open redirect — fixed
`isSafeRedirect` (`login.vue:146-148`) requires a leading `/`, rejects a leading `//`, and rejects any `\`. Probed bypasses: `/\evil.com` (contains backslash -> rejected), `?redirect=` empty / `?redirect=/admin&redirect=//evil.com` (array -> `typeof !== 'string'`, falls through to role-based redirect), `javascript:...` and `/ /` variants (no leading `/` -> rejected). Percent-encoded variants (`/%2f%2fevil.com`, `/%5cevil.com`) pass the filter but are handed to Vue Router as a single same-origin path segment — the router never interprets them as an external host, so no off-site navigation. No other frontend file reads `route.query.redirect`.

### M1(c) remember-me — fixed (removal option)
Checkbox and `form.remember` removed from `login.vue`; grep confirms zero `remember` references in the page. Backend `Login` handler (`backend/internal/handlers/auth/auth.go:186-189`) binds only `email`/`password` and never reads `remember`, so removal is honest. Residual cosmetic noise (unused `auth.remember_me` keys in 9 locales, optional `remember?: boolean` on `useAuth.login` at `useAuth.ts:120`) does not re-manifest the bug — it is dead, not a silent no-op control.

### M1(d) useApi TypeError — fixed
`useApi.ts:368-377` guards `typeof errorData === 'object' && errorData !== null` before `message`/`details` access; non-object bodies (plain-text, proxy error page, `null`, `undefined`) fall through to the default i18n message. Verified safe shapes: string body, `null`, array body (`'details' in []` -> false), object body with/without `message`. The optional-chaining `error?.statusCode` / `error?.data` also tolerate primitive/non-Error throws without throwing. `submitInquiry`'s separate catch (`useApi.ts:560-568`) uses `error?.data?.message` (already safe) and is untouched.

### Edge cases / regressions
- `sharedRefreshPromise` is cleared in `.finally()` on both resolve and reject, so a settled promise never leaves stale state for a later 401.
- 401-retry double-refresh loop on a genuinely-401 resource is pre-existing behavior (old code also re-ran refresh on retry-401); not introduced here.
- Auto-import of the new `refreshAuthSession` export does not collide (single definition, explicit import in plugin).
- No other file calls `/auth/refresh` or reads `redirect`; no API response-shape or prop changes ripple outside the four owned files.

**Blocker for `CONFIRMED_FIXED`:** none found.
