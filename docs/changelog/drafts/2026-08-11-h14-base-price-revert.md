# Changelog / Devlog — 2026-08-11 H14 Base-Price Revert
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** Startup migration `EnsureProductBasePrices` unconditionally reset admin-edited `base_price` on every restart.
**Result:** Guard added; existing nonzero prices preserved. Regression test added; targeted package build + tests pass.
---
## 1. Process
Read the finding and both owned files (`migrate_product_base_price.go`, `database.go`) plus the in-seed path (`seed.go`) and catalog defaults. Confirmed the migrate path lacked the `base_price = 0 OR base_price IS NULL` guard that the seed path has. Applied the matching guard, added a sqlite-based regression test, then built and tested the package.
## 2. Fixes in detail
#### H14 — EnsureProductBasePrices reverts admin price edits every startup
- **Problem:** `backend/internal/database/migrate_product_base_price.go:18-20` unconditionally ran `UPDATE products SET base_price = <catalog default> WHERE slug = ?` for all 18 seeded slugs on every startup (invoked from `database.go:262`). Any admin edit to a seeded product's `base_price` was silently reverted on restart, unlike the in-seed path (`seed.go:745`) which guards `(base_price = 0 OR base_price IS NULL)`.
- **Fix:** Added the identical guard `AND (base_price = 0 OR base_price IS NULL)` to the startup migration at `backend/internal/database/migrate_product_base_price.go:21`, so only unset prices are backfilled. Updated the doc comment to state the idempotent/no-overwrite intent. No change needed in `database.go:262` — it still invokes `EnsureProductBasePrices` within `EnsureStartupData`.
- **Test:** `backend/internal/database/migrate_product_base_price_test.go` — in-memory sqlite `products` table with three rows: admin-edited nonzero price on a catalog slug (must stay), zero price on a catalog slug (must be backfilled), zero price on a non-catalog slug (must stay 0). GORM soft-delete (`deleted_at`) and `updated_at` stamping columns included in the test schema.
## 3. Verification
| Check | Result |
|-------|--------|
| `go build ./...` (backend) | PASS |
| `REDIS_URL= go test ./internal/database/...` | PASS |
| `go vet ./internal/database/` | PASS |
## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED**

I attempted to refute this fix and could not construct a credible scenario where the bug still manifests or where the fix regresses adjacent behavior.

**What I checked**
- Root cause confirmed real: the original `migrate_product_base_price.go:18-20` ran `UPDATE products SET base_price = <catalog default> WHERE slug = ?` for all 18 catalog slugs on every startup (invoked from `database.go:262` inside `EnsureStartupData`), with no zero/NULL guard — unlike the in-seed path `seed.go:745`. Any nonzero admin edit to a seeded product's `base_price` was indeed silently reverted on restart.
- The fix (`migrate_product_base_price.go:21`) adds `AND (base_price = 0 OR base_price IS NULL)`, byte-for-byte identical to the seed path's guard. Behavior between the two paths is now identical — no divergence.
- Catalog defaults verified against `pkg/catalog/defaults.go` (map keys/values match the test's expectations: `4d-fruit-gummy`=8.50, `crystal-hard-candy`=5.10).
- Product model (`models/product/product.go:117,120-122`) confirms `base_price` is `float64 gorm:"default:0"` (NOT NULL, so the `IS NULL` clause is dead on Postgres but harmless/parity), plus `updated_at` and `deleted_at` — the test schema mirrors exactly the columns GORM appends to the UPDATE.
- Only caller of `EnsureProductBasePrices` is `database.go:262`; no other call sites, so no regression surface.
- `go build ./...` PASS, `REDIS_URL= go test ./internal/database/...` PASS (new test + existing catalog/COGS tests), `go vet ./internal/database/` clean.

**Edge cases considered (all closed or pre-existing/out-of-scope)**
- Nonzero admin edit on a catalog slug → WHERE no longer matches → row untouched → preserved. This is the exact reported bug; closed.
- Zero base_price on a catalog slug → still backfilled (intended "unset" semantics, matching seed path).
- Non-catalog slug → never matched by either old or new code; unaffected.
- Admin deliberately setting base_price to 0 to mean "no reference": would be backfilled on restart — but this matches the pre-existing in-seed behavior exactly and, for catalog slugs, COGS scaling (`pkg/catalog/defaults.go:58` `COGSReferencePrice`) consults the static `CatalogBasePrice` map first, so base_price=0 on a catalog slug has no COGS effect anyway. No functional harm.
- Duplicate slug across tenants/workspaces: pre-existing slug-keyed design; the guard strictly reduces overwrite harm (now only the zero-price row is touched). Not introduced by this fix.
- Concurrency: `EnsureStartupData` runs once at startup before serving; no race.
- Test robustness: float round-trip is exact on sqlite REAL (same double stored/read), and the nonzero row is never matched by the UPDATE, so `!= 9.99` comparison is sound. The `p-unset` backfill correctly depends on the catalog default matching the map.

**Non-blocking notes for the record**
- The migration swallows per-slug errors (logs, returns nil) — pre-existing pattern, not part of H14; if a future column change broke the UPDATE it would be silent, but the regression test would catch the wrong-value case.
- `gorm.DeletedAt` on Product means soft-deleted products are excluded from backfill; consistent with prior behavior.

