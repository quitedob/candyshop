# Changelog / Devlog — 2026-08-11 H16 — useSanitizer SSR no-op: raw stored HTML in first paint (XSS)
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** frontend Nuxt 3 / Vue 3 / TypeScript / Tailwind
**Trigger:** `useSanitizer` called `DOMPurify()` without a window, so during SSR DOMPurify returned an "unsupported" factory whose `sanitize()` returns the input UNCHANGED — raw stored HTML reached `v-html` on the SSR'd blog page at first paint, before hydration.
**Result:** The composable now detects the SSR environment (`import.meta.server`) and uses the isomorphic, Node-safe `sanitize-html` (new dependency) server-side, mirroring the DOMPurify allowlist; DOMPurify is kept for the client path. `npm --prefix frontend run typecheck` exits 0; the server config was empirically verified to strip every tested XSS vector.
---
## 1. Process
1. Read `frontend/composables/useSanitizer.ts` and `frontend/pages/blog/[slug].vue`; confirmed the SSR no-op in `dompurify/dist/purify.cjs.js` (`if (!DOMPurify.isSupported) { return dirty; }` at line 1374, `isSupported = false` at line 404 when no window is passed).
2. Verified Nuxt 3.21.1 cannot split composables per environment via `.server.ts`/`.client.ts` file overrides (both variants produced `Duplicated imports "useSanitizer" ... ignored` warnings and registered a single file), so the SSR branch is done in-file with `import.meta.server`.
3. Implemented the dual-path sanitizer, installed `sanitize-html`, ran the required typecheck, and empirically verified the server config against a battery of XSS vectors.
## 2. Fixes in detail
#### H16(a) — Server-safe sanitizer + client DOMPurify — `frontend/composables/useSanitizer.ts`
- **Problem:** `useSanitizer.ts` did `const purify = DOMPurify()` at module scope and called `purify.sanitize(...)` unconditionally. On the server there is no `window`, so DOMPurify returns an unsupported factory and its `sanitize()` returns the dirty input unchanged. `blog/[slug].vue` (`pages/blog/[slug].vue:195-199`) feeds `sanitize(raw)` into `v-html` (line 49) and the page is SSR'd (`ssr: true`, route rule `'/blog/**': { swr: 3600 }`), so admin/AI-authored stored HTML with `<img onerror>`/`<script>` executed at first paint. This was the only defense for the stored-HTML (H24) class on this path.
- **Fix:** Added the `sanitize-html` dependency (`package.json` dependencies) and rewrote `useSanitizer.ts` to branch on `import.meta.server`: server returns `sanitizeHtml(html, serverSanitizeConfig)`; client keeps `DOMPurify.sanitize(html, purifyConfig)`. Both configs are derived from the same shared `ALLOWED_TAGS`/`ALLOWED_ATTR` allowlist; the server config maps the DOMPurify URI policy to `allowedSchemes: ['http','https','mailto','tel']` + `allowProtocolRelative: true` (blocks `javascript:`, `data:`, `ftp:`). `stripHtml` gets the same dual path. `pages/blog/[slug].vue` needed no change — it already routes `content` through `sanitize()` before `v-html`, and the composable now sanitizes in both environments.
## 3. Verification
| Check | Result |
|-------|--------|
| `npm --prefix frontend run typecheck` | Exit 0 (only pre-existing nuxt.config env / nuxt-seo-utils warnings; no TS errors) |
| Server config functional probe (`node` against the exact `serverSanitizeConfig`) | `<img onerror>`→onerror stripped; `<script>`→removed; `javascript:`/`data:`/`ftp:` href/src→stripped; `onclick`/`data-*`/`style`→stripped; `https:`/relative URLs→kept; `<iframe>`→removed; `stripHtml`→plain text |
## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.** No credible scenario was constructed where the H16 bug still manifests or where the change regresses the non-affected path.

Independent verification (did not rely on the fixer's probe):
- Root cause confirmed via `git diff`: pre-fix code did `const purify = DOMPurify()` at module scope and called `purify.sanitize()` unconditionally; with no `window`, DOMPurify returns an unsupported factory whose `sanitize()` returns the input unchanged. The current `frontend/composables/useSanitizer.ts` branches on `import.meta.server`, returning `sanitizeHtml(html, serverSanitizeConfig)` server-side and `DOMPurify.sanitize` client-side.
- `import.meta.server`/`import.meta.client` is a proven, statically-replaced mechanism in this codebase (frontend/composables/useApi.ts:287,302,308,330; useAuth.ts:25,45; useDarkMode.ts; layouts/customer.vue; layouts/admin.vue; plugins/auth-auto-refresh.client.ts), so the SSR'd blog page's server render takes the sanitize-html branch and client hydration takes the DOMPurify branch.
- Independent Node probe running the EXACT `serverSanitizeConfig`/`stripAllServerConfig` from useSanitizer.ts against 24 vectors: `<img src=x onerror>`→onerror stripped; `<script>`/`<iframe>`/`<svg onload>`→removed; `javascript:` href (lower, UPPER, entity-encoded `&#58;`, control-char `\x00`, newline, HTML-comment-split)→all stripped; `data:` img+href→stripped; `ftp:`→stripped; `onclick`/`style`/`data-*`→stripped while `class`/`id` kept; `srcset` with `javascript:`→whole attribute dropped; https/mailto/relative/protocol-relative URLs→kept; `stripHtml`→plain text. The `launder` scheme check is case-insensitive and pre-strips control chars (≤0x20) and embedded HTML comments.
- Only SSR'd `v-html` surface is `pages/blog/[slug].vue:49`, whose `renderedContent` computed (lines 195-199) routes content through `sanitize()` — now effective server-side. The other `v-html` consumer (`pages/admin/ai/index.vue:70`) is CSR under route rule `/admin/**` ssr:false and keeps the working DOMPurify path. No other file under `pages/` uses `v-html`/`innerHTML` for stored HTML.
- Re-ran `npm --prefix frontend run typecheck`: exit 0, no TS errors (only the pre-existing nuxt-seo-utils module-compat warning and nuxt.config env warnings).

Edge cases / residual risks — none reproduce XSS:
- Client-bundle dead `sanitize-html` import: confirmed browser-safe at import time — `node_modules/sanitize-html/index.js` requires only htmlparser2, escape-string-regexp, is-plain-object, deepmerge, parse-srcset, postcss, launder (no Node builtins); `new URL()` is used only at call time, never on the client branch. It is dead code client-side (DOMPurify wins).
- Server/client output formatting may differ cosmetically (e.g. sanitize-html emits `<img src="x" />`, DOMPurify emits `<img src="x">`). Vue patches `v-html` during hydration; both outputs are sanitized, so this is at most a minor hydration diff, not a security or correctness regression.
- `stripHtml` server path now returns plain text (previously the no-op returned raw HTML). Only caller is `components/admin/InlineAiField.vue` (admin, CSR) — no SSR caller, so no behavioral regression.
- `allowProtocolRelative: true` keeps `//evil.com/x` — an external load/tracking concern, not script execution, and consistent with the client regex's `[^a-z]` branch. Out of H16 scope.
- The pre-existing `nuxt build` ENOENT `.nuxt/dist/client/manifest.json` in this parallel-agent tree is environmental (parallel `nuxt prepare` wiping `.nuxt`), not caused by this change.
