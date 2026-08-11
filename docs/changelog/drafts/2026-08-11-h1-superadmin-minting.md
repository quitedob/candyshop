# Changelog / Devlog — 2026-08-11 H1 Plain Admin Can Mint/Promote Superadmin
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** `AdminCreateUser`, `AdminUpdateUserRole` and `AdminUpdateUser` resolved the target role via `Auth.ResolveRoleID` with no check that the caller's role dominates the target, so any plain admin (admitted by `RequireRole(["admin","superadmin"])`) could mint a superadmin or promote themselves.
**Result:** Fixed — role assignment now goes through a new `AuthService.CanGrantRole` guard (superadmin-only for `admin`/`superadmin` targets, DB-backed caller role lookup, 403 otherwise) wired into all three role-assignment endpoints; regression tests added at service and handler level; `go build ./...`, `go vet`, and targeted `go test` all green.
---
## 1. Process
Read the finding, then traced `ResolveRoleID` (services/auth/auth.go), the middleware identity (`c.Get("userID")` / `c.Get("userRole")` set in middleware/auth.go), and all three role-assignment handlers. Confirmed `ResolveRoleID` never reads the caller's role. Implemented the guard in the AuthService (so it is unit-testable without the full admin services stack) and a thin `Handler.authorizeRoleGrant` wrapper, then added handler-level regression tests backed by an in-memory sqlite DB.
## 2. Fixes in detail
#### H1 — Plain admin can mint/promote superadmin
- **Problem:** `handlers/admin/admin_extensions.go:139,214` and `handlers/admin/admin_users.go:140` called `Auth.ResolveRoleID(ctx, roleID, roleName)` which only does `FindByID`/`FindByName` (services/auth/auth.go:138-160) and never inspects the caller. Any `admin` (JWT role in context, admitted by `RequireRole`) could set a new user's or existing user's role to `admin`/`superadmin`.
- **Fix:**
  1. `services/auth/auth.go` — added `FindByID` to the `userRepository` interface (satisfied by `*user.UserRepository`), split `ResolveRoleID` into `ResolveRole` (returns the full `*modelsAuth.Role`) + a thin `ResolveRoleID` wrapper, and added `ErrForbiddenRoleGrant`, `IsPrivilegedRole`, `CanGrantRole(ctx, callerUserID, target)` (loads the caller from the DB — not the JWT claim, so stale-token demotions are still denied) and a private `roleNameFor` helper.
  2. `handlers/admin/admin_users.go` — new `Handler.authorizeRoleGrant` helper: reads `c.Get("userID")`, calls `CanGrantRole`, writes 403 (`errors.forbidden_insufficient_role`, the existing fully-localized key) on denial.
  3. `handlers/admin/admin_extensions.go` + `admin_users.go` — all three role-assignment paths now resolve the role once via `ResolveRole`, gate it through `authorizeRoleGrant`, then assign `resolvedRole.ID`.
- **Related discovery (fixed in-scope):** GORM `Save` on a user fetched with `Preload("Role")` re-writes the belongs-to foreign key back to the stale snapshot, so a role change via `AdminUpdateUserRole`/`AdminUpdateUser` was silently reverted (reproduced in a temp test; `role_id` stayed `r-customer` after `Save`). Both handlers now set `user.Role = nil` before `UpdateUser` so the new role persists — required for the "superadmin can promote" half of the invariant to actually work.
- **Message key choice:** used the existing, fully-localized `forbidden_insufficient_role` key instead of inventing `forbidden_role_grant`, which would render as a raw dotted key in all 7 locales (the i18n engine falls back to the raw key when absent).
## 3. Verification
| Check | Result |
|-------|--------|
| `go build ./...` (backend) | PASS |
| `go vet ./internal/services/auth ./internal/handlers/admin` | PASS |
| `REDIS_URL= go test ./internal/services/auth/...` | PASS (new `TestIsPrivilegedRole`, `TestCanGrantRole_*`, `TestResolveRole_ByNameAndByID`) |
| `REDIS_URL= go test ./internal/handlers/admin/...` | PASS (new `admin_role_grant_test.go`: admin cannot mint superadmin / promote to admin via role endpoint / promote via profile update — all 403 with role unchanged; superadmin can mint + promote; admin can still create customer; missing caller denied) |
## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED** — I could not construct a credible scenario where the H1 bug still manifests or the fix regresses adjacent behavior.

**Checks performed:**
1. **Bug reproduced conceptually.** `ResolveRoleID` (services/auth/auth.go:165-171) resolves only the *target* role by ID/name and never reads the caller's role; the three role-assignment handlers (`admin_extensions.go:139,218` and `admin_users.go:164`) previously applied it unguarded. Real bug, admitted by `RequireRole(["admin","superadmin"])` + `RequireAdminWritePermission` (a plain admin's `admin:*` permission matches `admin:write`), so the premise holds.
2. **Role-grant surface audit.** Full-backend grep for `RoleID =` / `ResolveRole` / `RoleName` shows the only request-scoped role writes are the three guarded paths. Self-registration (`handlers/auth/auth.go:85-91`) accepts no role field and defaults to customer; seed code is not request-scoped. No bypass path found.
3. **Guard is fail-closed.** `CanGrantRole` (auth.go:186-201) denies on empty/unknown caller, DB-loads the caller (not the JWT `role` claim, so a stale-token demotion is still denied) and requires `roleNameFor(caller)=="superadmin"` for privileged targets. Non-privileged targets (customer/supplier/custom) remain grantable by any admin. Nil target can't reach the guard (authorizeRoleGrant is only called after a successful `ResolveRole`).
4. **Persistence verified.** `TestAdminUpdateUserRole_SuperadminCanPromoteToAdmin` confirms roleId actually persists (r-customer → r-admin) with `user.Role = nil` before `UpdateUser`; the profile path (`admin_users.go:173,176`) uses the identical pattern. The GORM belongs-to FK-rewrite claim is consistent with GORM `Save` semantics and the end-state test proves the fix works.
5. **Regressions checked.** `forbidden_insufficient_role` is localized in all 9 `seed_locales`. Frontend `admin/pages/users/[id].vue` `updateProfile` never sends `roleId`, so the now-omitted `role` field in the `AdminUpdateUser` response is not consumed; `updateRole` uses the dedicated `/users/:id/role` endpoint. `go build ./...` PASS, `go vet` PASS, `REDIS_URL= go test ./internal/services/auth/...` and `./internal/handlers/admin/...` PASS.

**Edge cases / residual notes (non-blocking):**
- The superadmin-promote branch of `AdminUpdateUser` (profile path) has no dedicated handler test; it mirrors the tested `AdminUpdateUserRole` path, so this is a minor coverage gap, not a functional one.
- Role-name comparison is case-sensitive (`role.Name == "admin"/"superadmin"`); a differently-cased name is not treated as privileged, but also confers no admin powers because `RequireRole` and `AttachAdminPermissions` key off the lowercase JWT role claim.
- The guard is enforced at the three handler sites only; a hypothetical future service-layer role write would bypass it (fixer's concern #5). No such caller exists today.
- A narrow race exists between `CanGrantRole` (superadmin check) and the role write; closing it would require wrapping the check and write in a single transaction. Not credible as an H1 exploit vector.
