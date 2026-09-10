# Audit follow-up — 2026-09-11

## Scope and evidence

This record preserves the audit supplied by the user and the coordinator's targeted remediation evidence. It is not a new whole-site audit or a claim that every reported defect is closed.

The source audit was supplied in the conversation on 2026-09-11. The user described five completed subagent reports and independent verification; those five source workers are distinct from the remediation workers listed in the [changelog](../changelog/2026-09-11-audit-remediation.md).

**User-reported evidence** below records the supplied findings without claiming that this documentation writer or the coordinator independently reproduced every measurement. **Coordinator verification** records checks reported by the coordinator or remediation workers during this fix pass.

The user's attribution of the CSP regression to commit `ab04f59` on 2026-08-11, comparison with parent `981c9cd`, and claim of a month-long outage were not independently re-proven in this follow-up. The supplied browser evidence was reported under `.audit-evidence/live-browser/`.

## User-reported findings

| Area | Supplied finding and evidence boundary |
| --- | --- |
| CSP and hydration | A single theme-script hash allowed only one of eight inline scripts. Blocked `window.__NUXT__`, a missing Vue mount, and browser console CSP errors explained non-working controls. Passing typecheck and build had not detected the failure. |
| Profit and loss SQL | Interval multiplication and concatenation produced text instead of an interval. SQL execution and live `PREPARE` demonstrated the failure; a monetary HTTP request was not completed because login rate limiting intervened. |
| COGS | Reported confirmed and production revenue of 146,250 had zero cost. Cart confirmation omitted costing, and the existing startup backfill excluded zero COGS. Historical source costs may be estimates rather than accounting evidence. |
| Financial aggregation | Tax of 971.81 was reported as margin; completed refunds were absent from aggregates; cancelled and expired orders inflated dashboard revenue by 68,000. Refund evidence was static because the development database contained no refund rows. |
| API contract | The frontend P&L table read `cost` and `profit`, while the response returned `cogs` and `grossProfit`. |
| Resume authorization | `/ai/resume` lacked an administrator guard and checkpoint ownership enforcement; concurrent resumes could execute effects more than once. |
| Idempotency | Replay could terminate before authentication/session revocation checks. Concurrent duplicate requests had no atomic pending reservation. |
| Order and trade state | Guarded order updates omitted version increments; trade guards normalized the requested status but could mismatch stored lowercase values; fulfillment delivery bypassed the order transition matrix. |
| Frontend accessibility | Hardcoded white cards and dark-mode tokens caused low contrast; a hidden mobile sidebar remained focusable. |
| Locale and assets | Legal links lost locale prefixes; CSR route rules missed prefixed routes; category images and nine linked PDFs were missing; Chinese category names leaked into English catalog content. |
| Build and documentation | The production build lacked an explicit internal backend URL. README route counts and previous audit memories disagreed with the user's current inventory. |

The user also reported green Go build/tests, operational health/readiness/metrics/swagger endpoints, successful administrator/customer login, expected authorization responses and rate limiting, and a green 41-route Nuxt build/typecheck. These are source-audit observations, not substitutes for the follow-up checks below.

## Implemented remediation and change impact

| Area | Implementation and practical effect |
| --- | --- |
| CSP | Hash inline scripts from the emitted response using Nitro `render:response` and `beforeResponse`, covering dynamic HTML and Node-served static/SWR responses. Route, locale, and payload changes no longer depend on a single fixed hash. |
| Public authentication | Start public-page session initialization after mount and bound session requests to five seconds. Public controls can mount while the backend is unavailable; protected middleware still awaits authentication. |
| Financial query | Cast the interval expression correctly; remove tax from profit calculations; deduct completed refunds in the original order period; exclude cancelled/expired orders from dashboard revenue. |
| Costing | Apply COGS during cart confirmation and fresh demo seeding; repair missing COGS only when every order line has a usable cost source. Incomplete orders are skipped in full and can be retried later. The existing estimated source-cost approach is preserved; no user-database historical backfill was executed. |
| Financial UI | Match P&L cells to the actual `cogs` and `grossProfit` response fields. |
| Resume | Require administrator access and checkpoint ownership; atomically consume a checkpoint once and rotate continuation identifiers. Old checkpoints with no owner fail closed. |
| Idempotency | Run replay handling after authentication, permission and KYB checks; reserve requests in the database; scope keys by concrete path and query; reject ambiguous legacy key reuse. |
| State and concurrency | Restore guarded order version increments, compare trade guards with stored status casing, and route fulfillment delivery through the order matrix with versioning and transaction rollback. |
| Frontend behavior | Extend CSR rules to locale-prefixed routes, localize legal links, repair dark card/hero/count/footer contrast and shared inverse tokens, public cards, CTAs and status badges; make closed mobile navigation inert with focus handling and provide category-image fallbacks. |
| Missing documents | Replace unavailable PDF downloads with localized contact actions. This resolves misleading download links without claiming the documents exist. |
| Deployment inputs and proxy | Declare the backend URL explicitly; production browser proxy and SSR honor runtime `NUXT_INTERNAL_API_BASE`, while the development proxy uses the resolved base. Operators still need a reachable backend. |
| Blog and prerender | Move CMS route discovery from a runtime plugin into the build hook, paginate all records, and make CMS prerender opt-in with `PRERENDER_CMS=true`. The default keeps CMS pages dynamic and supports offline Docker builds. Correct localized blog links/images and preserve real not-found versus unavailable responses (404/503). Three legacy seeded author-avatar filenames now have byte-identical aliases of existing assets in `frontend/public/images/team/`. |
| Route documentation | Correct README to 416 default business endpoints and 421 routes including infrastructure. The count uses distinct HTTP method/path pairs with normal database-backed wiring and both optional portal/warehouse flags disabled; frontend page counts were not reverified. |

### Change-impact closure

Paths below are repository-relative. **UPDATED** means the affected implementation changed; **COMPATIBLE** means the identified downstream contract remains usable, subject to the stated evidence boundaries.

| Reference chain | Closure and source paths |
| --- | --- |
| CSP → rendering → cache/static response | **UPDATED:** `frontend/nuxt.config.ts`, `frontend/server/utils/content-security-policy.ts`, `frontend/server/plugins/content-security-policy.ts`; response hashes cover renderer output and Node-served prerender/SWR responses. **COMPATIBLE:** Nuxt payload execution and restrictive inline-script enforcement both pass production browser checks. |
| Report repository → service → API → UI | **UPDATED:** `backend/internal/repository/order/order.go`, `frontend/pages/admin/index.vue`. **COMPATIBLE:** `backend/internal/services/order/order.go` → `backend/internal/handlers/admin/admin_analytics.go` → `backend/internal/api/routes/adminportal/register.go` retain the reporting route and `cogs`/`grossProfit` contract; isolated PostgreSQL tests, authenticated HTTP responses and a DevTools P&L click validate the chain without mutating the user database. |
| Checkpoint producers → store → consumer | **UPDATED:** `backend/internal/handlers/system/sse_handler.go`, `backend/internal/handlers/admin/admin_trade_ai.go` → `backend/internal/pkg/eino/checkpoint_store.go` → `backend/internal/handlers/system/resume_handler.go`, plus `backend/internal/api/routes/system/register.go`. **COMPATIBLE:** owned opaque checkpoint identifiers use the existing request/SSE fields; old unowned checkpoints intentionally fail closed. |
| Middleware construction → route registration → reservation | **UPDATED:** `backend/internal/api/router.go` → `backend/internal/api/routes/userportal/register.go` and `backend/internal/api/routes/adminportal/register.go` → `backend/internal/middleware/idempotency.go` → `backend/internal/repository/common/idempotency_key.go`. **COMPATIBLE:** authorized completed requests retain replay behavior; authorization runs first, while pending/legacy conflicts are deliberate restrictions. |
| Order/trade transitions → guards → admin stale-write check | **UPDATED:** `backend/internal/repository/order/order.go`, `backend/internal/repository/order/fulfillment.go`, `backend/internal/services/trade/trade.go`. **COMPATIBLE:** `backend/internal/handlers/admin/admin_orders.go` can again detect version changes; stored-case matching, transition validation and rollback have focused tests, without a production lock-race proof. |

## Coordinator verification

The initial remediation passed a 42-package Go suite and 29 production browser tests. The subsequent checks below extend that baseline through actual handlers on an isolated fixture backend and a broader public-page audit.

- `go build ./...` passed, and the latest full `go test ./... -count=1` passed across 43 packages. The prior focused seven-package security rerun also passed with PostgreSQL integration enabled. No user-database mutations were used.
- Latest `npm run typecheck` passed. The final asset-packaging production build exited 0 with default CMS settings, producing 225 prerender entries including payload assets, comprising 108 HTML pages; its link checker reported zero errors and zero warnings. Application code matches the passing 61-test suite; the subsequent rebuild added only three static avatar aliases. All three final-artifact image URLs returned 200 and their SHA-256 hashes matched the source assets byte for byte.
- The full production Playwright run passed all 61 tests in 1.3 minutes, with zero skips. Coverage includes the earlier 29 targeted checks, all 22 public light/dark accessibility cases, administrator login, customer/admin inquiry negotiation (buyer offer → administrator counter → acceptance), locale behavior, and seven blog/discovery tests. Transport-fixture scopes explicitly prevent SSR payload prefetch from bypassing their HTTP mocks.
- All 22 accessibility cases passed in development and production: 11 public routes, including contact, each in light and dark modes after shared-contrast repairs. This establishes the tested page/state coverage, not whole-site accessibility compliance.
- Actual fixture services ran on backend port 8081 with separate PostgreSQL on port 15439. In Chrome DevTools, administrator login returned 200 (request 603) and navigated to `/en/admin`; clicking P&L returned 200 (request 669) with revenue 145,400, COGS 58,160 and gross profit 87,240, with Cost and Profit populated in the table.
- A visitor multipart inquiry returned 200 (request 58) and displayed the success banner. The fixture used real HTTP handlers with isolated data and no actual email providers; delivery to an external mailbox was not tested.
- The financial worker verified day/week/month report HTTP responses were 200. A separate tax/refund fixture produced revenue 145,558, COGS 58,168, shipping 2 and profit 87,388; these values describe that fixture rather than the demo shown in DevTools. Fresh demo seeding produced COGS 58,160 on its first run.
- At the user's explicit request, Chrome DevTools verified Nuxt payload availability and Vue mounting on development page 2 and production page 3, plus dark-mode toggling, a clicked English-to-Korean change and password visibility. Clicking the production cookies page's Privacy link reached `/en/privacy` without errors.
- Chrome DevTools clicked the real seeded article `/en/blog/trends-candy-industry-2026` and displayed its CMS body.
- The development server was restarted after generated-cache changes left stale SSR output. DevTools then confirmed mounted `/en` pages and locale-preserving legal hrefs.
- Backend port 8080 became unreachable around a user interruption and was left offline because restarting would apply new migrations/backfill to the user database. Public controls worked during the outage; later login, report and inquiry verification used the isolated fixture service above.

Evidence logs and raw browser captures remain local under `.audit-evidence/` and are excluded from the commit. `remediation-push-browser-tests.log` records the 61-test pass and `remediation-delivery-build.log` records the final asset-packaging build; earlier checks are in `remediation-final-browser-tests.log`, `remediation-final-build.log` and `remediation-go-tests.log`. Independent process completions confirmed exit 0 for the full Go suite, typecheck and production browser suite.

For local review, the final production artifact on port 3013 and development server on port 3000 use isolated fixture backend 8081 and PostgreSQL 15439. Original backend 8080 remains offline and the user's PostgreSQL database on 5432 was untouched. No deployment is claimed.

## Residual risks and documentation reliability

- Login, monetary HTTP responses and inquiry submission were verified against isolated fixture services, not the user's operational database. No external email delivery, deployment or historical accounting correction is claimed.
- Refunds are attributed to the original order period, not the refund payment period. The chosen reporting behavior is not a complete cash-accounting model.
- Read-only inspection found six eligible user-database orders: five have complete cost sources totaling an estimated 66,860 COGS; the remaining confirmed order totals 100 and has no items, so it cannot be costed. No repair was run there. Reconstructing cost from available product data does not prove historical actual cost.
- Old unowned checkpoints cannot resume. Requests whose outcome is uncertain remain claimed, and ambiguous old idempotency keys conflict until expiry. No PostgreSQL lock race-detector proof is claimed.
- One user-data category still lacks a verified English translation; its unknown source meaning was not invented. Authentic PDF source documents were unavailable, so contact actions remain the truthful replacement. The five baseline `/blog` link errors are resolved in the completed default build's zero-error link check.
- HTTP proxy behavior is verified; actual WebSocket upgrades remain unverified. Existing polling fallback remains available, and WebSocket limitations predate this HTTP proxy change.
- The documentation scout reviewed README, changelog/index/acceptance material and feature documentation for authentication, orders, invoices, AI and public pages. **README — PARTIALLY RELIABLE:** its endpoint inventory is now corrected and source-verified, while historical frontend page counts remain unverified. Default business counts are public 25, auth 12, user 94, admin 269 and system 16: total 416, or 421 including uploads/health/ready/metrics/swagger.
- The coordinator's distinct HTTP method/path inventory uses normal database-backed service wiring, including eight translation handlers, with supplier portal and multiwarehouse disabled. The 431 registrar declarations minus 16 disabled routes plus the separately registered system scan hook give 416 business endpoints; enabling both flags gives 436. Disabling Swagger changes only the overall route count. The source audit's 419-live count is not supported by this current source inventory.
- **August 11 CSP documentation — CONTRADICTED:** the assumption that the theme script was the only inline script and that build/typecheck proved runtime correctness conflicts with response/browser evidence. This label does not independently establish the regression date or outage duration.
- **Invoice documentation — PARTIALLY RELIABLE:** `docs/features/09-invoices.md:42` already notes zero COGS and remains useful evidence of a known gap; it does not establish complete historical costing or that repair has run.
- All closure claims are limited to the named changes and evidence above. Untested routes/states, production configuration, external delivery and historical data repair require their own evidence.

See the [dated remediation entry](../changelog/2026-09-11-audit-remediation.md) for the actual worker roles, verification status and documentation change inventory.
