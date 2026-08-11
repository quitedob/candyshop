# Changelog / Devlog — 2026-08-11 Money FormatMoney rounding + lineage hash cent-stability (G26)

**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** G26 — `docxgen.FormatMoney` truncates instead of rounding (19.999 renders as 19.99, off-by-one-cent), and the order→invoice lineage hash feeds raw float64 through `fmt.Sprintf("%f", ...)`, which keeps 6 decimals of float noise so the same cent value can hash differently (false staleness) and nearby values can collide (wrong dedup).
**Result:** Both concrete defects fixed (not a full float64→decimal migration, per the scope note). `FormatMoney` rounds to cents with `math.Round` before formatting; the lineage hash now hashes money — the aggregate amounts AND each line item's `UnitPrice` — as exact integer cents via a new `money.MoneyToCentsInt` helper and a canonical line-item projection. The round-2 arguer refutation (line items still leaked raw float noise via `json.Marshal(o.Items)`; nil-vs-empty `Items` instability) is closed: the canonical projection normalizes per-item `UnitPrice` to cents and renders nil/empty `Items` identically. Regression tests added, including the non-empty-line-item noise test the arguer required. `go build ./...` clean; targeted `REDIS_URL= go test -count=1` passes; the new tests were proven to fail on the refuted path.

---
## 1. Process

Grep located `FormatMoney` (`internal/pkg/docxgen/docxgen.go:302`, callers in `contract.go`/`invoice.go`) and the lineage hash in `internal/services/order/lineage_hash.go:18` (called from `invoice_policy.go:86` and stored as `Invoice.OrderFinancialHash`). Confirmed `OrderFinancialHash` is only ever written, never compared against a stored value, so changing the hash algorithm is safe for existing rows. Empirically verified with a throwaway program that pre-fix `%f` prints 19.99 vs 19.9900006 as `19.990000` vs `19.990001` (different hashes) while cents-int normalizes both to 1999. After round 2 the arguer refuted: the aggregates were normalized but `json.Marshal(o.Items)` still fed the float64 `OrderItem.UnitPrice` into the hash, and nil vs empty `Items` hashed differently. Probes confirmed both claims (`json.Marshal` of 19.99 vs 19.9900006 yields different bytes; nil → `null` vs empty → `[]`). Implemented a canonical line-item projection in `lineage_hash.go` (per-item `UnitPrice` → integer cents; nil/empty → `[]`), added regression tests that fail on the refuted path, and re-verified build + tests.

## 2. Fixes in detail

#### G26a — `FormatMoney` rounds to the nearest cent instead of truncating (round 1, confirmed by arguer)
- **Problem:** `internal/pkg/docxgen/docxgen.go:302-309` computed `intPart := int(amt)` and `decPart := int((amt - intPart) * 100)`, truncating the fractional part. `FormatMoney(19.999, "USD")` returned `USD 19.99` — off by one cent on contract/invoice renderings.
- **Fix:** `docxgen.go:302` now rounds first: `rounded := math.Round(amt*100)/100`, then splits into integer part and absolute, rounded decimal part. 19.999 → `USD 20.00`; clean values (19.99 → 19.99) and negatives (−19.999 → `USD -20.00`) are unchanged in form. NaN/Inf guard added to match `money.RoundMoney` semantics. docxgen stays Go-stdlib-only (imports `math`), per its package doc.

#### G26b — Lineage hash is collision-resistant / cent-stable (rounds 1–2 aggregates; round 3 completes line items)
- **Problem:** `internal/services/order/lineage_hash.go:25` hashed `o.Subtotal`, `o.TaxAmount`, `o.ShippingAmount` with `%f` (6 decimals). Sub-cent float noise made the same nominal value hash differently (`%f` of 19.99 = `19.990000` vs 19.9900006 = `19.990001`), so a derived invoice could be falsely flagged stale. Conversely, `%f`-truncated neighbours near the 6th decimal collapse, i.e. hashing is not normalized to the money domain.
- **Fix:** added `money.MoneyToCentsInt(amount) int64` (`internal/pkg/money/money.go:45`) — rounds to 2-decimal cents (`math.Round(amount*100)`, NaN/Inf → 0), matching the existing `money.RoundMoney` semantics. `lineage_hash.go` now formats the three aggregates as `%d` integer cents. Hash format change is safe because `OrderFinancialHash` is audit-write-only (no stored comparison, no unique index).

#### G26c — Line-item float-noise leak + nil/empty instability (round 3, arguer refutation)
- **Problem:** `lineage_hash.go:37` fed `string(itemsJSON)` straight into the hash, where `itemsJSON = json.Marshal(o.Items)` and `OrderItem.UnitPrice` is `float64` (`internal/models/order/order.go:80`). `json.Marshal` serializes 19.99 vs 19.9900006 to different bytes, so two orders with the same nominal cent value but sub-cent float noise in a line item still produced different hashes — the exact float-noise scenario G26 describes, only the aggregates were normalized. Additionally `json.Marshal(nil Items)` → `null` vs `json.Marshal(OrderItemArray{})` → `[]`, so the same order hashed differently depending on preload state.
- **Fix:** `lineage_hash.go` now marshals `canonicalLineageItems(o.Items)` instead of the raw `OrderItemArray`. The canonical `lineageItem` projection carries `UnitPriceCents int64` (via `money.MoneyToCentsInt`), mirroring the rest of `OrderItem`'s JSON shape (optional fields still `omitempty`). `make([]lineageItem, 0, len(items))` guarantees a non-nil slice, so both nil and empty `Items` serialize as `[]` — same order, same hash regardless of preload. Distinct unit prices stay distinct (no over-collapsing). Scope note: this file is outside the strict owned-files list (see §4 response).

## 3. Verification

| Check | Result |
|-------|--------|
| `go build ./...` (backend) | PASS |
| `go vet ./internal/pkg/money/... ./internal/pkg/docxgen/... ./internal/services/order/...` | PASS |
| `REDIS_URL= go test -count=1 ./internal/pkg/money/... ./internal/pkg/docxgen/... ./internal/services/order/...` | PASS |
| `go test -run TestFormatMoney` | PASS (`round_up_off-by-one-cent`: 19.999 → `USD 20.00`; fails on pre-fix truncation) |
| `go test -run TestMoneyToCentsInt` | PASS |
| `go test -run TestOrderFinancialLineageHash` (all 7) | PASS — incl. new `_CentStableAgainstFloatNoise_WithLineItems`, `_DistinctUnitPriceStillDistinct`, `_NilVsEmptyItemsStable` |
| `go test -run TestOrderFinancialLineageHash_CentStableAgainstFloatNoise_WithLineItems` | PASS; proven to FAIL on the refuted `json.Marshal(o.Items)` path (probe: 19.99 vs 19.9900006 in line items → different hashes) |
| `go test -run TestOrderFinancialLineageHash_NilVsEmptyItemsStable` | PASS; proven to FAIL on the refuted path (probe: `null` vs `[]`) |
| `go test -run TestOrderFinancialLineageHash_DistinctUnitPriceStillDistinct` | PASS; guards against cents-normalization over-collapsing 19.99 vs 20.00 line prices |

## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED** — both G26 defects are closed at their root cause for every affected entry point, and I could not construct a credible scenario where either still manifests or a regression is introduced. (This replaces the prior round's REFUTED verdict; the round-3 fix addressed both refutation points.)

**Independently verified (from backend/, on the current on-disk state)**
- `go build ./...` PASS; `go vet` on the three packages PASS; `REDIS_URL= go test -count=1 ./internal/pkg/money/... ./internal/pkg/docxgen/... ./internal/services/order/...` PASS.
- All 7 lineage-hash tests PASS, including the two that were proven to fail on the refuted `json.Marshal(o.Items)` path (`_CentStableAgainstFloatNoise_WithLineItems`, `_NilVsEmptyItemsStable`) and the over-collapse guard (`_DistinctUnitPriceStillDistinct`).
- Independent probe (throwaway program, removed after run) confirmed:
  - `FormatMoney(19.999, "USD") == "USD 20.00"` — the finding's exact example.
  - Hash noise-stability holds for aggregates AND non-empty line items: same order ID + `Subtotal 19.99` vs `19.9900006` and `UnitPrice 19.99` vs `19.9900006`/`1.25` vs `1.2500001` → identical hash.
  - Distinct financials stay distinct (19.99 vs 20.00 line price → different hashes); item order is still significant (real line-up changes → different hash).
  - nil `Items` vs `OrderItemArray{}` → identical hash.
  - `FormatMoney` decPart is always in `[0,99]` (no `19.100`-style overflow) across a battery including `199999.9999` and `-199999.9999`.

**Why CONFIRMED_FIXED (root-cause audit across all entry points)**
- **defect (a) — off-by-one-cent truncation:** the only money renderer is `docxgen.FormatMoney` (all callers `invoice.go`/`contract.go` route through it; no other truncating money-format site exists in the repo — all other `%.2f` sites round). Now rounds via `math.Round(amt*100)/100` before splitting, NaN/Inf guarded. Only entry point, fully fixed.
- **defect (b) — `%f` lineage collisions:** the only lineage hash in the repo is `services/order.OrderFinancialLineageHash` (grep: no other `%f`-based or SHA256 lineage hash; the trade doc "lineage" is an enum `manual|order_derived|ai_draft`, not a hash). Its money inputs are now all integer cents: the three aggregates via `MoneyToCentsInt` and every line item's `UnitPrice` via the canonical `lineageItem.UnitPriceCents` projection — the only float field on `OrderItem` (`models/order/order.go:80`). `OrderFinancialHash` is audit-write-only (grep: written only at `invoice_policy.go:101`, never read/compared, no unique index), so changing the hash algorithm is safe for existing rows. Line items were the last leak; closed.

**Edge cases found (none refute; all pre-existing or out of the finding's explicit scope)**
1. **`FormatMoney` negative sub-dollar sign loss:** `-0.99 → "USD 0.99"`, `-0.006 → "USD 0.01"` (intPart truncates toward zero, decPart uses `math.Abs`). Pre-fix behavior was the same broken class (`-0.99 → "USD 0.99"`), so this is NOT a regression and is outside G26's two named defects. Worth a future follow-up: split on `cents := int(math.Round(amt*100))` sign-aware.
2. **Float64 binary half-value display limits:** `1.005 → "USD 1.00"` (not 1.01) and `2.675 → "USD 2.68"`. Inherent to float64 representation, explicitly out of scope per the finding note ("NOT a full float64→decimal migration"). The fix's rounding still satisfies the named contract (19.999 → 20.00, no truncation).
3. **Hash separator ambiguity:** the pipe-delimited `%s|%s|%d|%d|%d|%s` format without length-prefix/length-prefix can in theory collide if a stored `ID`/`Currency`/`Specifications` value contains `|` (e.g., `ID="A|B", Currency="USD"` vs `ID="A", Currency="B|USD"`). Pre-existing structure (identical in the old `%f` code), not introduced by this fix, and requires adversarial string data in non-numeric fields — not the float-noise collision G26 names.
4. **`MoneyToCentsInt` int64 overflow** above `|amount| ~9.2e16` — not a realistic money value.

**Regression check on non-affected paths:** no caller contract or response shape changes — `FormatMoney` keeps its `"USD 1,234.56"` output shape, `OrderFinancialLineageHash` keeps returning a 64-char hex string, and no existing test or stored-value comparison depends on the old `%f` bytes. Scope note for the orchestrator: the canonical serializer + its tests live in `services/order/lineage_hash.go`/`lineage_hash_test.go` (outside the strict owned-files list), but this is the file G26's fix-approach mandates; both owned files (`money.go`, `docxgen.go`) are unmodified this round and re-verified by the full build + test run.
