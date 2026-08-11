# Changelog / Devlog — 2026-08-11 G23 follow-up: /admin/staff customer PII leak (wire GetStaffUsers)

**Date:** 2026-08-11
**Source report:** Code review finding G23
**Scope:** `backend/internal/handlers/admin/admin_staff.go`, `backend/internal/services/user/user.go`, `backend/internal/repository/user/user.go`
**Trigger:** Adversarial arguer refuted the prior round: `admin_staff.go:21` still called `h.services.User.GetUsers(...)` → `repository/user/user.go:64-80 FindAll` (no role filter), so customer PII still reached the `/admin/staff` response; `GetStaffUsers` (`services/user/user.go:79`) was dead code with zero production callers.
**Result:** CLOSED — `/admin/staff` now sources the staff list from the role-filtered `GetStaffUsers` (admin + superadmin only). Customer accounts and their PII (email, firstName, lastName, phone, company, companyId, status, roleId) never enter the response. Regression tests added at handler level; they FAIL on the pre-fix path and PASS post-fix.

## 1. Process

1. Verified the two named defects against the current code on disk:
   - `backend/internal/handlers/admin/admin_staff.go:21` → `h.services.User.GetUsers(c.Request.Context(), page, limit)` → `FindAll` (unfiltered, Preload Role) → ALL accounts incl. customers with PII.
   - `GetStaffUsers` (`backend/internal/services/user/user.go:79`) existed (added by the prior fixer) and was role-filtered via `FindByRoleNames(ctx, modelsAuth.AdminPortal())`, but had zero production callers.
2. Confirmed the `/staff` route (`backend/internal/api/routes/adminportal/register.go:345`) is already behind `group.Use(middleware.RequireRole(modelsAuth.AdminPortal()...))` — the auth guard is preserved and untouched.
3. Confirmed the generic all-users paths (`AdminGetUsers` at `admin_users.go:25` and the customer XLSX export at `admin_xlsx.go:310`) legitimately still use `GetUsers` and are guarded by `TestGetUsers_StillReturnsAll` — preserved.
4. Wired the handler, added handler-level regression tests, verified fail-pre-fix / pass-post-fix, full build, and targeted suites.

## 2. Fixes in detail

#### G23(b) — Wire /admin/staff to the role-filtered GetStaffUsers

- **Problem:** `GetStaffList` (`backend/internal/handlers/admin/admin_staff.go:21`) called `h.services.User.GetUsers(...)` which delegates to `repository/user/user.go:64-80 FindAll` — no role filter, `Preload("Role")`, returns every account. A customer row (email, firstName, lastName, phone, company, companyId, status, roleId) was serialized into the staff payload. `GetStaffUsers` (`backend/internal/services/user/user.go:79`) was dead code.
- **Fix:** `backend/internal/handlers/admin/admin_staff.go:21` now calls `h.services.User.GetStaffUsers(c.Request.Context(), page, limit)` (same signature: `[]modelsUser.User, int64, error`). `GetStaffUsers` is role-scoped (`FindByRoleNames(ctx, modelsAuth.AdminPortal())` → `["admin","superadmin"]`, join-filtered in `repository/user/user.go:117-128`) and pages in memory on top of that staff set, so customers never enter the response. `GetStaffUsers` now has a production caller (dead-code defect closed). Route guard `RequireRole(AdminPortal()...)` unchanged.

## 3. Verification

| Check | Result |
|---|---|
| `go build ./...` (backend) | PASS |
| `REDIS_URL= go test -count=1 ./internal/handlers/admin/...` | PASS (includes new `TestGetStaffList_ReturnsOnlyStaffRoles`, `TestGetStaffList_PaginationAppliesOnStaffSet`) |
| `REDIS_URL= go test -count=1 ./internal/services/user/...` | PASS (prior G23 service tests preserved: `TestGetStaffUsers_FiltersCustomers`, `TestGetStaffUsers_EmptyStaff`, `TestGetUsers_StillReturnsAll`) |
| Pre-fix regression check (handler reverted to `GetUsers`) | FAILS as required — `total = 3, want 2 (staff only, customer excluded)`; customer PII present |
| Route still behind `RequireRole(AdminPortal()...)` | PASS (`adminportal/register.go:345`) |

## 4. Adversarial review (arguer)

Verdict: **CONFIRMED_FIXED** — I attempted to refute the follow-up claim and could not construct a credible scenario where G23(b) still manifests.

What I verified on disk:
- `backend/internal/handlers/admin/admin_staff.go:24` — `GetStaffList` now calls `h.services.User.GetStaffUsers(c.Request.Context(), page, limit)`. The generic `GetUsers`/`FindAll` path is no longer reachable from `/admin/staff`. The only remaining `GetUsers` callers are `admin_users.go:25` (AdminGetUsers — intentional all-users listing) and `admin_xlsx.go:310` (customer XLSX export), both legitimately out of scope.
- `backend/internal/services/user/user.go:79-100` — `GetStaffUsers` has a production caller (dead-code defect closed) and is role-scoped: `FindByRoleNames(ctx, modelsAuth.AdminPortal())` → `["admin","superadmin"]`.
- `backend/internal/repository/user/user.go:117-128` — `FindByRoleNames` is a `JOIN roles ON roles.id = users.role_id ... WHERE roles.name IN (...)`, so a `customer`/`supplier` row is excluded at the SQL level; even a malformed RoleID cannot inject a customer because the JOIN drops rows without a matching admin/superadmin role.
- Route guard preserved: `backend/internal/api/routes/adminportal/register.go:15` applies `RequireRole(modelsAuth.AdminPortal()...)` to the whole admin group, and `/staff` (`register.go:345`) sits inside that group. Part (a) supplier API key disclosure untouched.
- Frontend contract preserved: `frontend/pages/admin/staff/index.vue:149-151` reads `res.data` + `res.pagination`; handler returns `PaginatedResponse` with those fields; `User.Role` JSON tag (`json:"role,omitempty"`) and `RoleID` (`json:"roleId"`) unchanged, so `roleBadgeClass(role?.name)` still works.

Regression tests non-vacuity: `TestGetStaffList_ReturnsOnlyStaffRoles` seeds 2 staff + 1 customer (carol@acme.com with rich PII) and asserts `total == 2` and no `carol@acme.com`/`Carol`/`333`/`ACME Inc` in the payload. On the pre-fix path (`GetUsers`/`FindAll`) total would be 3 and the customer row would be present → both assertions fail. `TestGetStaffList_PaginationAppliesOnStaffSet` (limit=1) fails pre-fix on `total` (3 vs 2) and `totalPages` (3 vs 2). Verified non-vacuous by construction, not just by the fixer's word.

Runs I reproduced from `backend/`:
- `go build ./...` — PASS.
- `REDIS_URL= go test -count=1 ./internal/handlers/admin/... ./internal/services/user/... ./internal/repository/user/...` — ok (handler + service suites green; repo has no test files).

Edge cases considered and dismissed as non-refutations:
- **In-memory pagination of a large staff set** (`GetStaffUsers` pages after loading all staff via `FindByRoleNames`): not a PII issue and bounded by the staff set size; it is the prior fixer's design, explicitly preserved. Not a defect.
- **No ORDER BY on `FindByRoleNames`** → unstable staff ordering across requests: cosmetic at worst, pre-existing, out of G23 scope.
- **Page-beyond-range** (`start >= len(staff)` returns empty slice + correct total): returns `[]` not `null`, matches `TestGetStaffUsers_EmptyStaff` guard.
- **A customer holding an admin role_id**: that account is by definition an admin (role-based access); promotion is superadmin-gated via `authorizeRoleGrant` (`admin_users.go:101-117`). Not a leak.

No regression on the already-fixed part (a) or the preserved all-users paths. Nothing to refute; the finding is closed.
