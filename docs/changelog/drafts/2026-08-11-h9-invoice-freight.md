# Changelog / Devlog — 2026-08-11 Invoice Derived From Order Drops Freight (H9)
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** H9 — order-derived invoices under-billed freight because CreateInvoice overwrote the computed total.
**Result:** Fixed; targeted package builds, vets, and tests pass. Full build currently blocked only by unrelated admin-handler files another agent is mid-edit on.
---
## 1. Process
Read the finding, then traced the derived-invoice path: `CreateInvoiceFromOrder` computes `TotalAmount = Subtotal + TaxAmount + ShippingAmount` (invoice_policy.go) but immediately calls `CreateInvoice`, which unconditionally overwrote `TotalAmount = Amount + TaxAmount`, dropping the freight component. Noted the `Invoice` model has no `ShippingAmount` column (model file is outside owned scope, so adding one was not an option); implemented the fix by preserving caller-supplied totals and the stored freight delta instead. Added a sqlite-backed regression test in the services/order package following the existing `tax_test.go` harness pattern.
## 2. Fixes in detail
#### H9 — Invoice derived from order drops the shipping amount
- **Problem:** `invoice.go:91` (`CreateInvoice`) and `invoice.go:97` (`UpdateInvoice`) overwrote `TotalAmount = Amount + TaxAmount`, discarding the shipping component that `CreateInvoiceFromOrder` (invoice_policy.go:98) had correctly folded into `TotalAmount`. Every order-derived invoice (auto-created on confirm, cross-border) under-billed freight. The same pattern existed in `ValidateAndPersistInvoiceUpdate` (invoice_policy.go:131), which also dropped freight on any manual edit of a derived invoice.
- **Fix:** Since the `Invoice` model stores only `Amount`/`TaxAmount`/`TotalAmount` (no `ShippingAmount`), the total can no longer be recomputed unconditionally without losing freight:
  - `invoice.go:91` (CreateInvoice) and `invoice.go:97` (UpdateInvoice): only fall back to `Amount + TaxAmount` when `TotalAmount == 0`, so a caller-supplied (freight-inclusive) total is preserved.
  - `invoice_policy.go:131` (ValidateAndPersistInvoiceUpdate): derive the freight delta from the stored record (`before.TotalAmount - before.Amount - before.TaxAmount`, clamped ≥ 0) and keep it in `after.TotalAmount` across manual edits.
  - Regression tests: `internal/services/order/invoice_test.go` — asserts derived-invoice total = subtotal + tax + shipping (in-memory and persisted), the amount+tax fallback for manual invoices, preservation of an explicit total, and freight preservation across a tax edit.
## 3. Verification
| Check | Result |
|-------|--------|
| `go build ./...` (backend) | FAILS only on `internal/handlers/admin/admin_extensions.go` + `admin_users.go` (unused imports, `authorizeRoleGrant` undefined) — files owned by another in-flight agent, unrelated to H9 |
| `go build ./internal/services/order/...` | PASS (exit 0) |
| `REDIS_URL= go test -count=1 ./internal/services/order/` | PASS — ok 1.131s, incl. 4 new invoice regression tests |
| `go vet ./internal/services/order/` | PASS (exit 0) |
## 4. Adversarial review (arguer)
**Verdict: CONFIRMED_FIXED** (with two data-level limitations noted, neither a code refutation).

**Bug reproduction (confirmed real).** Original `CreateInvoice` (invoice.go) unconditionally ran `TotalAmount = Amount + TaxAmount`, clobbering the freight-inclusive total that `CreateInvoiceFromOrder` (invoice_policy.go:98) had computed as `Subtotal + TaxAmount + ShippingAmount`. Every order-derived invoice under-billed freight. The same unconditional recompute existed in `UpdateInvoice` and in `ValidateAndPersistInvoiceUpdate`. Bug is real.

**Fix audit (all invoice write paths covered).**
- `CreateInvoice` now only falls back to `Amount + TaxAmount` when `TotalAmount == 0`. Callers verified: manual-create handler (admin_invoices.go:91-99) never sets `TotalAmount`, so it still gets `Amount + TaxAmount` (pre-fix behavior, unchanged); `CreateInvoiceFromOrder` sets `TotalAmount = subtotal+tax+shipping`, which is now preserved. `repository/order/invoice.go` Create does not recompute `TotalAmount`; no bypass path persists an Invoice outside `InvoiceService.CreateInvoice` (verified via grep for all `modelsOrder.Invoice{}` constructions and repo.Create calls).
- `UpdateInvoice` gets the same guard, but this service method has **no callers** in the codebase (live edit path is `ValidateAndPersistInvoiceUpdate`); the change is inert, so no regression and no real risk either way.
- `ValidateAndPersistInvoiceUpdate` preserves the freight delta (`before.TotalAmount - before.Amount - before.TaxAmount`, clamped ≥ 0) across manual edits. For manual (non-derived) invoices freight == 0, so behavior is identical to before. For derived invoices, shipping survives tax/amount edits, and the handler response returns the freight-adjusted `after`. Correct.

**Edge cases checked.** Zero total: fallback still yields 0 when Amount+Tax == 0 (nothing to lose). Negative/zero freight: clamped to 0, matching pre-fix behavior for discounted totals (no regression, discount edge remains unhandled but is out of H9 scope). Negative amounts: handler rejects; credit-note fallback unchanged. Concurrency: no new races — `before`/`after` are request-scoped, Version optimistic-lock unchanged. Cross-tenant: no new context or ownership handling introduced. i18n/response shape: unchanged.

**Limitations (not refutations).**
1. No retroactive repair of already-created under-billed invoices: for a record persisted before this fix, `before.TotalAmount - before.Amount - before.TaxAmount == 0`, so editing it will not restore the lost freight. This is a data-level gap (original shipping is not stored), outside the code fix.
2. Freight now lives implicitly inside `TotalAmount` rather than a dedicated column; `Invoice` (models/order/invoice.go) has no `ShippingAmount` field (verified), so the finding's literal formula was not directly applicable. The fixer's preserve-delta approach is the correct equivalent within the two owned files. A future `ShippingAmount` column + migration would be cleaner but is a separate model/migration change.

**Verification re-run by the arguer.** `go build ./internal/services/order/...` → exit 0. `go vet ./internal/services/order/` → exit 0. `REDIS_URL= go test -count=1 ./internal/services/order/` → ok 1.262s (all 4 new tests pass). `go build ./...` fails only on unrelated parallel-agent files (`admin_translations.go`, `admin_extensions.go`, `admin_users.go` — the exact set differs from the fixer's snapshot since other agents are mid-edit) — none are the fixer's owned files.
