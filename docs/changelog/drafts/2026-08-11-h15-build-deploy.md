# Changelog / Devlog — 2026-08-11 H15 make prod-start build/deploy fix
**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** `make prod-start` runs `docker-compose.prod.yml` standalone although that file declares no `image:`/`build:`, so the deploy fails with "no image/build".
**Result:** Fixed `prod-start` (and aligned `prod-stop`) to merge the base compose file; `go build ./...` verified green locally. The fresh-clone build break from the gitignored `backend/docs` import is handled by the LEAD force-adding `backend/docs` (no code change needed in `cmd/api/main.go`).
---
## 1. Process
Read the finding H15 and inspected `Makefile` (prod targets at lines 156-164), `docker-compose.prod.yml`, and `docker-compose.yml`. Confirmed `docker-compose.prod.yml` only overrides services declared in the base file and its header (line 3) documents the required merge `docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d`. Confirmed `backend/docs/` exists on disk (docs.go, swagger.json, openapi.yaml, swagger.yaml) while `.gitignore` line 36 (`docs/`) keeps it untracked, so a fresh clone would fail `go build` on the `_ "candypro/api/docs"` import. Fixed the Makefile; no Go code changes.
## 2. Fixes in detail
#### H15 — make prod-start broken; fresh clone fails to build
- **Problem:** `Makefile:161` `prod-start` ran `docker-compose -f docker-compose.prod.yml up -d` alone. The prod file declares no `image:`/`build:` for any service (only overrides), so compose has nothing to build/run and the deploy fails with "no image/build". The same misuse existed in `prod-stop` (Makefile:164). Separately, `backend/cmd/api/main.go:4` blank-imports `_ "candypro/api/docs"`, which is gitignored (`docs/` in `.gitignore`) and untracked — a fresh clone would fail `go build`.
- **Fix:** In `Makefile`, `prod-start` now runs `docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d`, merging the base compose file as documented by the prod file header; added a short comment above the target explaining why the merge is required. `prod-stop` was aligned to `docker-compose -f docker-compose.yml -f docker-compose.prod.yml down` for consistency. `cmd/api/main.go` was intentionally left unchanged — the docs import is correct and required for the swagger wiring; the gitignored-docs build break is resolved by the LEAD force-adding `backend/docs` (`git add -f backend/docs`).
## 3. Verification
| Check | Result |
|-------|--------|
| `go build ./...` (backend) | PASS (exit 0; `_ "candypro/api/docs"` resolves because backend/docs exists on disk) |
| `REDIS_URL= go test ./cmd/api/...` | PASS — "no test files", exit 0 |
| `make -n prod-start` | N/A — `make` not installed in this environment (exit 127); target syntax verified by reading (recipe lines tab-indented, comment above target) |
| Fresh-clone build | Not verifiable here — requires the LEAD's `git add -f backend/docs`; pre-fix state would fail, post-fix state requires docs committed |
## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED** (for the fixer's scope) — with one explicit residual dependency on the LEAD.

### Part A (prod-start deploy failure): genuinely fixed, verified against the real docker-compose binary
- The finding's root cause reproduces exactly. With required env set, the pre-fix command `docker-compose -f docker-compose.prod.yml config` fails with `service "api" has neither an image nor a build context specified: invalid compose project` (exit 1) — the "no image/build" failure.
- The post-fix command `docker-compose -f docker-compose.yml -f docker-compose.prod.yml config` validates cleanly (exit 0): `api.build` resolves from the base file, and the prod overrides apply (`ENVIRONMENT=production`, `CORS_ORIGINS`, `REDIS_URL`). This matches the command documented in `docker-compose.prod.yml`'s own header verbatim.
- Makefile syntax is valid: recipe lines are tab-indented (`^I` via cat -A), and the explanatory comment sits above the target (not echoed as a recipe line). `make` is genuinely absent (exit 127) but `docker-compose`/`docker` binaries exist and provided a stronger functional validation than `make -n` would have.
- Edge cases: required env vars (`CORS_ORIGINS`, `JWT_SECRET`, `SEED_SUPERADMIN_PASSWORD`, all `:?`-interpolated in the prod file) now fail-fast at `up` — correct prod behavior, and the standalone command failed on the same interpolation anyway, so no new failure mode. `deploy:` blocks are ignored by docker-compose v1 outside swarm (warning only; `config` passed). `docker-compose` (v1 binary) vs the header's `docker compose` (v2) is a non-issue: the Makefile uses `docker-compose` consistently across all docker targets and the binary is present.
- prod-stop change: project name is directory-derived and identical across both `-f` forms, so `down` targets the same containers; strictly more consistent, no regression. The diff is 5 insertions / 2 deletions confined to prod-start/prod-stop + comment; no other make target is affected.

### Part B (fresh-clone build): NOT resolved by this change set — still pending the LEAD's `git add -f backend/docs`
- Confirmed current state: `git ls-files backend/docs` returns empty (untracked) and `git check-ignore backend/docs` matches `.gitignore:36` (`docs/`). A fresh clone of the current branch still fails `go build` on `_ "candypro/api/docs"`.
- This is correctly out of the fixer's reach: the finding itself assigns `git add -f backend/docs` to the LEAD, the fixer was instructed not to run git commands, and no code change in `cmd/api/main.go` can fix a git-tracking issue. Leaving `main.go` unchanged was the correct call. The residual risk is real until task #5's force-add lands — the parent orchestrator must NOT assume fresh-clone builds are green until then. This is a hand-off dependency, not a fixer defect.

### Regression check
- `go build ./...` (backend): PASS, exit 0, verified twice — the docs import resolves against on-disk `backend/docs`.
- `go test ./cmd/api/...`: the fixer's PASS ("no test files") held for the H15 change (no Go code modified). In my runs the command failed to build, but on `internal/services/order/negotiation.go:102` and `internal/handlers/admin/admin_extensions.go:4` — both outside H15's owned files (Makefile, cmd/api/main.go) and transient artifacts of concurrent edits by the in-flight H3 task. Unrelated to this fix.
