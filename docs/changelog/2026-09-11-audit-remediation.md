# 2026-09-11 — Audit remediation

Source: [audit follow-up and supplied-finding record](../reports/audit-followup-2026-09-11.md). This entry covers targeted fixes prompted by the user's audit; it does not certify a complete site audit or repeat the source audit's regression date as independently verified fact.

## Problem and resulting behavior

The fixed CSP hash omitted Nuxt's changing inline payload, preventing browser hydration. Response-derived hashes now cover emitted scripts, and public session initialization no longer holds page mounting open during a backend outage. Chrome DevTools confirms a mounted application and working public controls on both development and production-preview servers.

The P&L query now constructs an interval correctly, excludes tax from profit, applies completed refunds to the original order period, and uses the correct frontend response fields. Dashboard revenue excludes cancelled/expired orders. Cart confirmation and fresh demo seeding write COGS; missing-cost repair skips an entire order if any line lacks a usable source and retries later. No user-database backfill was run; estimated costing remains an explicit limitation.

Resume now checks administrator access and ownership and consumes checkpoints atomically. Idempotency executes after access checks and reserves database keys before effects, including concrete request path/query and fail-closed handling of old ambiguous keys. Guarded order updates increment versions, trade updates preserve stored-case guards, and fulfillment delivery follows the transition matrix with rollback.

Locale-prefixed protected routes receive the intended CSR rules. Legal links preserve locale; closed mobile navigation is inert with focus handling; public dark-mode cards, footer, shared inverse tokens, CTAs/status badges and category-image fallbacks were repaired. Missing PDF links now lead to localized contact actions. Docker/environment configuration supplies a backend URL; the production HTTP proxy and SSR honor runtime `NUXT_INTERNAL_API_BASE`, and the development proxy uses the same resolved base.

Blog route discovery moved from runtime to `frontend/build/prerender-routes.ts`, registered through the build hook in `frontend/nuxt.config.ts`. It paginates CMS results and makes CMS prerender optional (`PRERENDER_CMS=false` by default), preserving dynamic content and offline Docker builds. Localized blog links/images and 404-versus-503 responses were corrected; the completed default build's link check has zero errors or warnings. Three missing legacy author-avatar filenames now alias existing assets byte for byte in `frontend/public/images/team/`.

README now reports 416 default business endpoints (public 25, auth 12, user 94, admin 269, system 16), or 421 including infrastructure. The source inventory counts distinct HTTP method/path pairs with normal database-backed wiring and eight translation handlers: 431 registrar declarations minus 16 disabled routes plus one separately registered system scan hook. Enabling both supplier and multiwarehouse flags gives 436 business endpoints; disabling Swagger changes only the overall count. Historical frontend page totals were not reverified; the supplied 419-live count is not supported by current source.

## Actual collaboration

| Worker | Bounded work |
| --- | --- |
| Coordinator | Integrated fixes; implemented CSP/auth/proxy/prerender and shared-contrast refinements; verified source counts, isolated HTTP/browser flows and cross-cutting checks; used Chrome DevTools following the user's request. |
| `docs_scout` | Read-only review of documentation requirements, reliability and affected feature descriptions. |
| `financial_fixes` | Financial SQL/reporting/costing, isolated HTTP and seed/backfill validation, and read-only inspection of user-data cost-source gaps. |
| `security_state_fixes` | Resume authorization/consumption, idempotency and state/version safeguards with focused validation. |
| `frontend_content_fixes` | Content, locale, accessibility and image/PDF behavior plus fixture-backed browser checks. |
| `remediation_docs` | Wrote this entry and its linked source record, prepended the changelog index entry, and corrected README's endpoint inventory using the coordinator's source verification. |

These are the actual remediation workers; no fixer/adversarial-arguer pairing is claimed. The five source-audit subagents described by the user were not recreated by this documentation pass.

## Verification and limits

Initial verification passed 42 Go packages and 29 production browser tests. Follow-up extended this to real handlers with isolated fixture data and a broader public-page sweep:

- Go build passed; the latest full Go suite passed 43 packages, latest typecheck passed, and the focused seven-package security suite passed with PostgreSQL integration enabled. No user-database mutation was part of the checks.
- The final asset-packaging production build exited 0 with 225 prerender entries including payload assets, comprising 108 HTML pages, and zero link-check errors/warnings. Its application code matches the 61-test pass; only three static avatar aliases were added afterward. Each image returned 200 from the final artifact and matched its source SHA-256 hash byte for byte.
- The full production Playwright suite passed 61 tests in 1.3 minutes with zero skips: the earlier 29 checks, 22 public accessibility cases, administrator login, customer/admin negotiation (offer → counter → acceptance), locale checks and seven blog/discovery tests. Fixture scopes disable SSR payload prefetch so navigation uses the intended HTTP mocks.
- All 22 accessibility cases passed in development and production: 11 public routes including contact, each in light and dark modes. This does not certify untested pages/states.
- Chrome DevTools fixture login returned 200 and opened `/en/admin`; clicked P&L returned 200 and displayed revenue 145,400, COGS 58,160 and profit 87,240. Visitor multipart inquiry returned 200 with the success banner. Backend 8081 and PostgreSQL 15439 were isolated; external email providers were not used.
- The financial worker verified day/week/month HTTP reports and a separate tax/refund fixture (revenue 145,558; COGS 58,168; shipping 2; profit 87,388), plus first-seed COGS of 58,160.
- Chrome DevTools verified mounted Nuxt payload/Vue on development and production preview, theme toggling, a clicked English-to-Korean switch and password visibility. Privacy reached `/en/privacy` without errors; clicking `/en/blog/trends-candy-industry-2026` displayed its real seeded CMS body. Restarting stale development SSR restored verified `/en` legal hrefs.

Evidence logs/raw captures under `.audit-evidence/` remain local and are excluded from the commit: `remediation-push-browser-tests.log` records the 61-test pass and `remediation-delivery-build.log` the final asset-packaging build. Independent process completions confirmed exit 0 for full Go tests, typecheck and the production browser suite. Login/report/inquiry behavior was verified with isolated services; external email delivery, deployment and user-database historical COGS repair were not.

Final production preview 3013 and development 3000 use isolated fixture backend 8081 and PostgreSQL 15439 for local review. Original backend 8080 remains offline because startup would apply migrations/backfill to the user database; original PostgreSQL 5432 was untouched.

Read-only user-data inspection found six eligible orders: five have complete source-cost estimates totaling 66,860; one confirmed order totaling 100 has no items and cannot be costed. No actual backfill ran. A category's verified English translation and authentic PDF source documents remain unavailable; neither was invented. Old unowned checkpoints are rejected, uncertain requests remain claimed and legacy idempotency conflicts persist until expiry. Actual WebSocket upgrades and PostgreSQL lock race-detector proof are unverified; existing polling fallback remains. See the source record for five-chain impact closure and README/invoice **PARTIALLY RELIABLE** and August 11 CSP **CONTRADICTED** labels.

## Documentation impact

The source record distinguishes user-reported findings from coordinator verification and records closure boundaries, final checks, source paths through five affected reference chains, and outstanding operational validation. This entry and the reverse-chronological index link back to it. Existing documentation was preserved; no historical audit entry was rewritten.

Documentation writer identifier inventory: no program identifiers, renames or extracted constants were introduced. Changes are limited to the two named Markdown artifacts, one index entry and README's documented route inventory. Code identifier/reference inventories belong to the implementation workers' reports, not this documentation-only change.

The user explicitly authorized committing and pushing the completed remediation with this devlog in `docs/`; the coordinator performs that final Git operation. Raw `.audit-evidence/` captures are excluded. This statement records authorization, not a claim that a push has already succeeded.
