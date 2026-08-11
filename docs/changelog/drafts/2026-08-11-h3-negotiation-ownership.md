# Changelog / Devlog — 2026-08-11 H3 Negotiation Accept/Reject Ownership Scoping
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** Negotiation offer accept/reject transitions ran by offer id only; the `userID` param was unused and the reject path had no ownership check at all.
**Result:** Fixed — transitions now scoped by offer id AND inquiry id, ownership verified in handler (customer) and service (inquiry binding); regression tests added; targeted packages build/vet/test green.
---
## 1. Process
Read the finding, then the full accept/reject paths in the customer handler, admin handler, NegotiationService, NegotiationRepository and the route registrations to confirm the ownership gap. Traced every caller of `AcceptOffer`/`RejectOffer`/`TransitionStatus` (only the four handlers + one test fake) so the interface change was safe. Implemented inquiry-id scoping end to end, updated the affected test harness, and added service- and repository-level regression tests.
## 2. Fixes in detail
#### H3 — Negotiation accept/reject does not verify offer ownership
- **Problem:** `handlers/customer/customer_negotiation.go:128,180` and `handlers/admin/admin_negotiation.go:80,133` passed an offer id plus a `userID` that the service never used; `repository/order/negotiation.go:51` transitioned by `id AND status` only. The customer accept path verified the *path inquiry* belonged to the caller but never checked the offer actually belonged to that inquiry, so a caller on inquiry X could accept/reject an offer on inquiry Y (cross-user state corruption) and the admin accept could build an order from a path inquiry using another offer's terms. Customer reject had no ownership check at all.
- **Fix:** (1) `repository/order/negotiation.go:51` `TransitionStatus` now scopes by `id AND inquiry_id AND status`. (2) `services/order/negotiation.go` `AcceptOffer`/`RejectOffer` now take `inquiryID` (dropping the unused `userID`), pre-read the offer to return `ErrNegotiationOfferNotFound` (404) when the offer's inquiry differs — the write stays an atomic conditional UPDATE, preserving the A-4 TOCTOU guarantee. (3) `handlers/customer/customer_negotiation.go` reject path now reads the `:id` param, verifies `inquiry.UserID == caller` (403) and passes the inquiry id; accept passes the verified path inquiry id. (4) `handlers/admin/admin_negotiation.go` accept/reject pass the path inquiry id so the offer must belong to the inquiry being acted on. Ownership anchor for offers is the *inquiry owner* (an offer's `UserID` is its sender, so it cannot be used for customer accept checks); the handler checks inquiry ownership via `InquiryService.GetInquiry`, the service/repo enforces the offer-inquiry binding.
## 3. Verification
| Check | Result |
|-------|--------|
| `go build ./...` (backend) | PASS |
| `go vet ./internal/repository/order ./internal/services/order ./internal/handlers/customer ./internal/handlers/admin` | PASS |
| `REDIS_URL= go test ./internal/repository/order/...` | PASS (incl. new `TestTransitionStatus_ScopedByInquiry`) |
| `REDIS_URL= go test ./internal/services/order/...` | PASS (incl. new `TestNegotiation{Accept,Reject}Offer_InquiryMismatch`) |
| `REDIS_URL= go test ./internal/handlers/customer/...` | PASS (incl. new cross-inquiry 404 / forbidden 403 regression tests) |
| `REDIS_URL= go test ./internal/handlers/admin/...` | 1 unrelated pre-existing failure: `TestAdminUpdateUserRole_SuperadminCanPromoteToAdmin` (`admin_role_grant_test.go`, untracked file from another agent's in-progress role-grant work; role stays `r-customer`). No negotiation tests exist for admin; all admin negotiation paths compile/vet clean. |
## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED** — I was unable to construct a scenario where the original H3 bug (cross-user/cross-inquiry offer state corruption, or an admin building an order from a path inquiry using another inquiry's offer terms) still manifests, or where the fix regresses adjacent behavior.

**What I re-verified independently (on-disk code, not just the report):**
- Root cause was real: `handlers/customer/customer_negotiation.go` and `handlers/admin/admin_negotiation.go` called `AcceptOffer`/`RejectOffer` with a `userID` the service never used; `repository/order/negotiation.go` transitioned by `id AND status` only. Confirmed the customer reject path previously had no ownership check and admin accept could bind any offer's terms to any path inquiry.
- Defense is now triple-layered and consistent:
  1. Handler (customer): `GetInquiry(pathID)` then `inquiry.UserID == caller` → 403 before any service call (accept AND reject — reject previously unchecked). Nil-UserID inquiry → 403 (safe default).
  2. Service: `AcceptOffer`/`RejectOffer` pre-read the offer and return `ErrNegotiationOfferNotFound` (404) when `offer.InquiryID != inquiryID` — no mutation, and the 404 is indistinguishable from a genuinely missing offer (no cross-inquiry existence leak).
  3. Repository: `TransitionStatus` WHERE now includes `inquiry_id` (defense-in-depth; a mismatched inquiry can never mutate, even if a caller bypassed layers 1-2).
- TOCTOU/A-4 preserved: the write is still the atomic conditional UPDATE `WHERE id AND inquiry_id AND status='pending'`; the pre-read is only for error classification. Concurrent double-accept: one wins (rows=1), the other gets 409. Reject/accept races likewise.
- Order-build cannot mix terms: `CreateOrderFromAcceptedOffer` has exactly three production callers — `admin_inquiry_convert.go:58` (nil offer, catalog pricing), and the two accept handlers, both of which have already bound the offer to the path inquiry by the time the order is built. The "build order from path inquiry using another offer's terms" scenario is closed.
- No stale callers: grepped every reference to `AcceptOffer`/`RejectOffer`/`TransitionStatus` — only the four handlers plus the test fakes; the interface change compiled everywhere.

**Edge cases checked (none refute the fix):**
- Empty/nil `inquiryID` or `offerID`: route segments are non-empty; if somehow empty, service returns 404 before any write.
- Cross-inquiry where caller owns neither inquiry: handler 403; service 404. Cross-inquiry where caller owns the path inquiry but offer is elsewhere: service 404, status untouched (verified by `TestCustomer{Accept,Reject}NegotiationOffer_OfferFromOtherInquiry`).
- Same-inquiry non-pending offer: 409 via re-lookup error classification; no info leak distinguishing cross-inquiry from nonexistent.
- The unscoped `repo.Update` (Save) and `FindPendingByInquiryID` remain unscoped, but `Update` is not reachable from any handler (the service does not expose it; only the unexported interface carries it) and `FindPendingByInquiryID` has no production caller in the accept/reject flow — out of H3 scope, no new exposure.
- The service re-lookup `FindByID(id)` after a successful transition is unscoped but only reached post-success on a row that was just written under the scoped WHERE — benign.

**Regression verification (re-run, not taken from the report):**
- `go build ./...` (backend): PASS.
- `go vet` on the 4 owned packages: PASS.
- `REDIS_URL= go test ./internal/repository/order/... ./internal/services/order/... ./internal/handlers/customer/...`: all PASS (incl. the new H-3 regression tests).
- `REDIS_URL= go test ./internal/handlers/admin/...`: PASS — the previously-flagged `TestAdminUpdateUserRole_SuperadminCanPromoteToAdmin` now also passes in my run (the fixer's note about that untracked failure appears to have been transient / since settled; the file `admin_role_grant_test.go` still exists but the package is green).

No concrete defect found. The fix closes the root cause (inquiry-id scoping end to end), not just a symptom, and preserves the existing A-4 concurrency guarantee.
