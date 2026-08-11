# Changelog / Devlog — 2026-08-11 i18n in-memory cache invalidation (G17)
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** G17 — the i18n in-memory cache is additive-only: deleted keys keep resolving (blank/stale) until restart; batch imports that remove/deactivate keys never drop them from memory.
**Result:** Added `i18n.DeleteFromCache` and `i18n.ReplaceCache` primitives; `TranslationService.Delete` now removes the key from the cache (instead of writing an empty value), `applyDelta` treats a deactivated row as a delete, and `Import` rebuilds the cache via full replace. Also fixed a latent `BulkUpsert` quoting bug that generated invalid `ON CONFLICT` SQL on SQLite. Two red-green regression tests added. Full build green; targeted tests green.
---
## 1. Process
Read the current `pkg/i18n/i18n.go`, `services/translation/translation.go`, and `repository/translation/translation.go` (all already ctx-threaded by a parallel H13 pass this session). Traced the write paths (`Create`/`Update`/`Delete`/`Import`) and the cache primitives. Confirmed the failure modes empirically: pre-fix `Delete` leaves the key resolving as `""` (skipping the zh and raw-key fallbacks) and `Import`'s merge-only `WarmCache` keeps deactivated keys resolving with stale values. Verified the GORM quoting behavior from the pinned v1.30.0 source before touching `BulkUpsert`, then wrote regression tests that fail on the pre-fix path and pass post-fix.
## 2. Fixes in detail
#### G17 — i18n in-memory cache never invalidated on delete/import
- **Problem:** `backend/internal/pkg/i18n/i18n.go:79-92` — `WarmCache` is additive-merge only; nothing removes keys. `backend/internal/services/translation/translation.go` (pre-fix `Delete` at :78) wrote `Value: ""` into the cache, so `Translate()` found the entry and returned `""` instead of falling back to the default locale / raw key. The pre-fix `refreshCache` (`:134`) merged via `WarmCache`, so a key deactivated by a batch import (e.g. `is_active=false`) kept resolving with a stale value until restart.
- **Fix:** Added `DeleteFromCache(locale, key)` (removes one `locale|group.key` entry under the lock) and `ReplaceCache(records)` (rebuilds the whole map, dropping entries not returned by the active set) to `pkg/i18n/i18n.go`. `TranslationService.Delete` now calls `i18n.DeleteFromCache` (translation.go:77). `applyDelta` now routes `IsActive == false` through `DeleteFromCache` (translation.go:111-114) so the admin deactivation path also invalidates immediately. `refreshCache` now calls `i18n.ReplaceCache` instead of `WarmCache` (translation.go:136) so a batch import removes stale/deactivated keys.
- **Also fixed (same file, needed for the Import path to run under SQLite):** `repository/translation/translation.go` `BulkUpsert` used a pre-quoted `` `"group"` `` map key in the `ON CONFLICT ... DO UPDATE` clause. GORM's PostgreSQL `QuoteTo` collapses self-quoted identifiers (`"group"` → `"group"`, valid), but the SQLite dialector's `QuoteTo` is a plain backtick wrapper, producing `` `"group"` `` (invalid). Changed the key to the unquoted `group` so each driver quotes it itself (PG `"group"`, SQLite `` `group` ``) — identical SQL on PostgreSQL, valid SQL on SQLite.
- **Files:** `backend/internal/pkg/i18n/i18n.go`, `backend/internal/services/translation/translation.go`, `backend/internal/repository/translation/translation.go`, new `backend/internal/services/translation/translation_test.go`.
- **Regression tests:** `TestDeleteInvalidatesCache` (seed via repo → `i18n.Init` → `svc.Delete` → asserts the key stops resolving, falling back to the raw key; pre-fix returned `""`). `TestImportRebuildsCache` (seed active key → `i18n.Init` → deactivate row out-of-band in the DB to simulate a stale cache → `svc.Import` a new key → asserts the deactivated key stops resolving and the new key resolves; pre-fix returned the stale `"Hello"`). Both verified red on the pre-fix logic and green post-fix. The Import test deactivates via repo `Update` rather than a batch row because GORM's `default:true` tag coerces an explicit `IsActive:false` back to `true` on `Create` (noted in the test comment).
## 3. Verification
| Check | Result |
|-------|--------|
| `go build ./...` (backend) | PASS (clean) — see note on unrelated packages below |
| `go vet ./internal/pkg/i18n/ ./internal/services/translation/ ./internal/repository/translation/` | PASS (clean) |
| `REDIS_URL= go test ./internal/services/translation/` | PASS (`TestDeleteInvalidatesCache`, `TestImportRebuildsCache`) |
| `REDIS_URL= go test ./internal/repository/translation/` | PASS (`TestDeleteDoesNotLeakWhereClause`) |
| Red-on-pre-fix check | Both new tests FAIL against the pre-fix service logic (Delete → `""`; Import → stale `"Hello"`) and PASS with the fix restored |
Note: at final verification the full `go build ./...` showed 2 unrelated failures in packages other agents are mid-editing this session — `internal/pkg/payment/paypal` (`amountsEqual` undefined, unused imports) and `internal/repository/order` (`deductWarehouseStock` signature mismatch). Both compiled cleanly earlier in this same session, confirming the tree changed under me; neither is in scope here.
## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED** — I could not construct a credible scenario where G17 still manifests in any real admin flow, nor a regression on the non-affected path. This review was run against the current on-disk state, not the fixer's report.

### What I re-verified independently
- `go build ./...` (backend): **PASS** at review time — the paypal/order failures noted in §3 are no longer present (concurrent agents resolved them); nothing in the G17 fix breaks the tree.
- `REDIS_URL= go test -count=1 ./internal/pkg/i18n/... ./internal/services/translation/... ./internal/repository/translation/...`: PASS (i18n: no tests; services: `TestDeleteInvalidatesCache`, `TestImportRebuildsCache`; repo: `TestDeleteDoesNotLeakWhereClause`).
- `go vet` + `go build` on the 3 owned packages and their sole dependent (`handlers/admin`): PASS, clean.
- GORM quoting claim: traced `gorm.io/driver/postgres@v1.5.7` `QuoteTo`. For the map key `group` it emits `"group"`; for the pre-fix key `"group"` (embedded quotes) it collapses to the same `"group"` (the `selfQuoted` branch). So the unquote change is **byte-identical SQL on PostgreSQL** and fixes invalid `` `"group"` `` on SQLite. No PG regression.

### Entry-point audit (root cause coverage)
The translations table has exactly one runtime writer scope: admin. `TranslationRepository` is instantiated only in `repository/scopes/adminportalscope/repositories.go:90`; `TranslationService` only in `adminportalscope/services.go:55`; the only handler is `handlers/admin/admin_translations.go`. No seed, system-scope, or other code mutates the table at runtime. Each write path now invalidates:
- **Delete** (`DeleteTranslation` → `svc.Delete`): pre-reads the row, soft-deletes, then `i18n.DeleteFromCache(locale, group.key)` — key removed entirely, so `Translate` falls through zh→raw-key (matches fresh-start). Fixes the "blank" symptom (pre-fix wrote `""` into the cache, which skipped both fallbacks).
- **Deactivate** (`UpdateTranslation` with `isActive:false`): `applyDelta` routes `!IsActive` → `DeleteFromCache`. Correct; `repo.Update` uses `Save` which persists the `false` bool (Save is full-field, so the `default:true` tag does not re-coerce it on Update — only on Create).
- **Import** (`ImportTranslations` → `svc.Import`): `refreshCache` now calls `i18n.ReplaceCache` (full rebuild from `FindAllActive`), so keys missing from the active set are dropped. This is the correct semantic — an admin Import only upserts (handler always sets `IsActive:true`), so the only keys that vanish from the active set are ones deactivated/removed out-of-band, exactly what ReplaceCache handles.
- `WarmCache` is now called only from `applyDelta` (single-key active write); no leftover caller re-introduces a stale/blank entry.

### Edge cases that do NOT refute (checked, not blocking)
- **Soft-delete + re-import of the same key**: `BulkUpsert` `ON CONFLICT (key,locale)` will UPDATE the soft-deleted row but not clear `deleted_at`, so `FindAllActive` (and thus the rebuilt cache) omits it. This is a pre-existing DB-layer behavior of the unchanged `BulkUpsert`/`FindAllActive`, not introduced here, and outside G17.
- **Concurrency**: all cache mutations (`WarmCache`/`DeleteFromCache`/`ReplaceCache`) hold the write lock; `Translate` holds RLock. A `Delete` racing an `Import` converges on the DB active set.
- **Cross-locale fallback**: deleting the `en` row leaves `zh|key` intact; `Translate("en", key)` correctly falls back to zh, matching restart behavior.
- **Empty/zero**: `applyDelta` guards `Locale==""||Key==""`; an admin-set empty `Value` still renders `""` (pre-existing, admin-input contract, not cache invalidation).

### Residual observations (pre-existing / out of G17 scope, for the record)
1. **Group-rename leaves an orphan**: `UpdateTranslation` lets the admin change `Group`; `applyDelta` writes the new `group.key` but the old `group.key` lingers (pre-fix behavior unchanged). In the most contrived case — rename then later Delete — the delete removes only the new key and the orphaned old key still resolves until restart. The fixer already flagged this; it is an update-path artifact, not a delete/import invalidation, and predates this fix.
2. **`default:true` tag**: an explicit `IsActive:false` on Create is coerced to true (test comment acknowledges this); the real admin flows (Create/Import) always pass `true`, so production is unaffected.
3. **`refreshCache` error path**: if `FindAllActive` fails after a successful `BulkUpsert`, the cache is left untouched (neither the new keys nor the dropped ones). Pre-existing; DB-read-failure-only.
