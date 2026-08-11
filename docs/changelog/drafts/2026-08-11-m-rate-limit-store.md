# Changelog / Devlog — 2026-08-11 G19 Rate-Limit Store Bounded

**Date:** 2026-08-11
**Source report:** [../../reports/code-review-audit-2026-08-11.md](../../reports/code-review-audit-2026-08-11.md)
**Scope:** backend Go / Gin / GORM
**Trigger:** G19 — the in-memory rate-limit store grows a map entry per distinct key and never evicts (memory leak under high distinct-IP traffic), and the Redis rate-limit store fails OPEN on Redis error (an outage silently disables rate limiting).
**Result:** The memory store is now bounded (lazy TTL on access + size cap with oldest-active-first eviction); the Redis store fails BOUNDED on error by routing to a per-instance in-memory fallback that enforces the same limit instead of returning unlimited. Existing happy-path tests pass unchanged; 3 new memory-store regression tests + 1 Redis fail-bounded test added. Package build, full backend build, and targeted tests pass.

---

## 1. Process

Read `backend/internal/pkg/ratelimit/store.go` and `redis.go`, confirmed `memoryStore.Allow` re-assigns every key (never `delete`s) so the map grows unbounded, and `redisStore.Allow` returns `true` on `INCR` error (fail-open). Bounded the memory store with lazy TTL + a `maxKeys` cap (oldest-active-first eviction) and added a bounded in-memory `fallback` to the Redis store for the error path. Added regression tests in `store_test.go` and `redis_test.go`, then verified with `go vet`, `go build ./...`, and `REDIS_URL= go test`.

## 2. Fixes in detail

#### G19a — Unbounded in-memory store
- **Problem:** `backend/internal/pkg/ratelimit/store.go:24-43` — `Allow` filters each key's history to the window but always re-assigns the key (`m.attempts[key] = filtered`), so a key whose window fully elapsed stays in the map forever. One entry per distinct IP, never evicted: unbounded memory under high distinct-IP traffic.
- **Fix:** `store.go` — added `maxKeys` field + `defaultMaxKeys = 100_000`; on access a key with no history inside the window is `delete`d (lazy TTL, line 47-51); after every insert `sweepIfNeeded` (lines 67-98) runs only when over the cap — pass 1 drops keys whose history fully fell out of the window, pass 2 evicts the least-recently-active keys oldest-first. Steady-state map size is bounded by the cap; below the cap it is a no-op.

#### G19b — Redis store fails open on error
- **Problem:** `backend/internal/pkg/ratelimit/redis.go:48-52` — on `INCR` error the store logged and returned `true` (allow), so a Redis outage disabled rate limiting entirely.
- **Fix:** `redis.go` — added a `fallback Store` (a `NewMemoryStore()`, itself now bounded) initialized in `NewRedisStore` (lines 36-41); the error path (lines 56-65) now logs loudly and returns `r.fallback.Allow(key, limit, window)`, enforcing the same limit per-instance rather than unlimited. Comment documents the per-instance (not distributed) caveat.

#### G19 — Tests
- `store_test.go`: `TestMemoryStore_WindowElapseResetsCount` (a key at its limit from before the window is allowed again and records only the fresh attempt), `TestMemoryStore_CapsSize` (map stays ≤ cap under many distinct keys), `TestMemoryStore_CapsSizeOldestEvictedFirst` (least-recently-active key evicted first).
- `redis_test.go`: `TestRedisStore_Allow_FailsBoundedOnRedisError` — with Redis up the limit is enforced; after `miniredis.Close()` the store must still block the 4th attempt (fails on the pre-fix `return true` path).

## 3. Verification

| Check | Result |
|-------|--------|
| `go vet ./internal/pkg/ratelimit/...` | ok |
| `go build ./internal/pkg/ratelimit/...` | ok |
| `go build ./...` (backend) | ok (exit 0, no errors) |
| `REDIS_URL= go test -count=1 ./internal/pkg/ratelimit/...` | ok — all tests pass incl. 4 new cases |
| `REDIS_URL= go test -count=1 ./internal/middleware/...` | ok — dependent package unaffected |
| `-race` | n/a — cgo/gcc unavailable on this Windows box |

## 4. Adversarial review (arguer)

**Verdict: CONFIRMED_FIXED.** I attempted to refute and could not construct a credible scenario where the originally-reported bug (unbounded memory growth; Redis outage = unlimited) still manifests.

**Root-cause audit (re-read `store.go`, `redis.go` on disk):**
- **G19a (memory leak):** `memoryStore` now carries `maxKeys` (set to `defaultMaxKeys=100_000` in the only constructor) and `sweepIfNeeded` runs after every insert, so the map is capped at `maxKeys`; the lazy `delete` on access (store.go:47-51) only shrinks it. Every insert path is followed by `sweepIfNeeded`; the blocked path (store.go:56) and the lazy-delete path never add an entry, so the invariant `len(attempts) <= maxKeys` holds after every `Allow`. Per-entry backing arrays are bounded by `limit` (history ≥ limit is never appended to), so worst-case memory is ~100k × O(limit) `time.Time`s — bounded. Root cause closed. The `history[:0]` in-place filtering and the `range`-with-`delete` in pass 1 are both safe Go (no aliasing bug; the current request's `now` is in `cutoff`-filtered history so the just-inserted key can never be deleted).
- **G19b (Redis fail-open):** the only remaining `return true` on the error path is guarded by `r.client == nil || limit <= 0` (redis.go:44). `client` is never nil (constructor pings), and no caller passes `limit <= 0` (authscope/public routes use limits 2/3/5/10). Every `INCR` error now routes to the bounded `fallback` (redis.go:64), which enforces the same limit per-instance. Not fail-open.

**Verified:** `go build ./...` exit 0; `go vet ./internal/pkg/ratelimit/...` ok; `REDIS_URL= go test -count=1 ./internal/pkg/ratelimit/...` ok (4 new tests + existing); `REDIS_URL= go test -count=1 ./internal/middleware/...` ok (dependent auth-rate-limit middleware unaffected). `-race` cannot run on this box (cgo/gcc missing) — all `memoryStore` access is under one mutex and the fallback is constructor-initialized, so no race is expected.

**Residual edge cases (documented tradeoffs, NOT refutations — all strictly better than the pre-fix state):**
1. **Eviction can reset a live victim counter under botnet-scale distinct-key flood.** Pass 2 evicts oldest-active-first. An attacker flooding >100k distinct source IPs controls their own flood timestamps (always "newest"), which can push a targeted victim's key to be evicted and reset its budget. Inherent to per-IP rate limiting + bounded memory; the finding explicitly requested the security-vs-memory tradeoff. A botnet at that scale defeats per-IP limiting regardless.
2. **Multi-instance Redis outage multiplies budget by instance count.** Each instance's `fallback` is independent, so behind an LB an attacker gets N× the configured limit while Redis is down. The code logs this loudly (redis.go:63). This mitigates fail-open rather than eliminating all abuse during an outage — acceptable and intended.
3. **O(n) sweep cost per request only when at the cap.** Pass 1 scans all entries and pass 2 scans to find the oldest on every insert once the map is at `maxKeys` (100k iterations/request under cap-exceeding load). No-op in the normal case. CPU-DoS-adjacent under extreme flood, but bounded and off the hot path in steady state.
4. **`limit <= 0` semantic inconsistency:** memory returns false (block all), redis returns true (allow all). No current caller hits it; pre-existing; not a regression.
5. **Fallback uses a sliding window while the Redis path uses fixed buckets** — a minor enforcement-semantics mismatch during an outage only; both enforce the limit.
