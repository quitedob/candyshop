// Package workerlock 为后台 worker 提供基于 Redis 的分布式互斥（L-6）。
//
// 当多实例部署时，希望"同一个 worker 任务在任意时刻只有一个实例运行"，
// 例如 trade-outbox-relay、shipment-tracking-sync。Worker 在 tick 开始时
// 调用 Acquire，结束时调用 Release。Acquire 失败说明锁被其他实例持有，
// worker 直接 skip 本次 tick。
//
// 锁实现采用 SET NX EX 模式 + 续约 token，确保：
//   - 持锁者 panic / 超时仍会自动过期；
//   - Release 校验 token，防止误删其他实例的锁。
//
// Redis 不可用时退化为 always-acquire（单实例本地行为），以避免阻塞核心
// 业务流；调用方自行决定是否记录降级告警。
package workerlock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrLockNotHeld is returned by Release when the caller did not own the lock.
var ErrLockNotHeld = errors.New("worker lock not held")

// Locker is the public interface; production-bound to RedisLocker, tests can
// substitute a no-op implementation.
type Locker interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (token string, ok bool, err error)
	Release(ctx context.Context, key, token string) error
}

// RedisLocker uses Redis SET NX EX semantics.
type RedisLocker struct {
	client *redis.Client
	prefix string
}

// NewRedisLocker connects to redisURL and returns a RedisLocker.
func NewRedisLocker(redisURL string) (*RedisLocker, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("workerlock: parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("workerlock: redis ping: %w", err)
	}
	return &RedisLocker{client: client, prefix: "candypro:wl:"}, nil
}

// Acquire tries to take the lock identified by key for at most ttl.
// Returns (token, true, nil) on success — keep the token to safely Release later.
// Returns (_, false, nil) if the lock is held by another instance.
func (l *RedisLocker) Acquire(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", false, fmt.Errorf("workerlock: rand: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	ok, err := l.client.SetNX(ctx, l.prefix+key, token, ttl).Result()
	if err != nil {
		return "", false, err
	}
	return token, ok, nil
}

// Release frees the lock only if token matches (CAS via Lua to avoid races).
func (l *RedisLocker) Release(ctx context.Context, key, token string) error {
	const releaseScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	res, err := l.client.Eval(ctx, releaseScript, []string{l.prefix + key}, token).Result()
	if err != nil {
		return err
	}
	if n, ok := res.(int64); !ok || n == 0 {
		return ErrLockNotHeld
	}
	return nil
}

// Close releases the underlying Redis connection.
func (l *RedisLocker) Close() error {
	if l.client == nil {
		return nil
	}
	return l.client.Close()
}

// NoopLocker accepts every Acquire and is a safe fallback when Redis is unavailable.
type NoopLocker struct{}

func (NoopLocker) Acquire(_ context.Context, _ string, _ time.Duration) (string, bool, error) {
	return "noop", true, nil
}
func (NoopLocker) Release(_ context.Context, _, _ string) error { return nil }

// NewFromEnv returns a RedisLocker if redisURL is set and reachable, otherwise
// a NoopLocker so single-instance deployments still work.
func NewFromEnv(redisURL string) Locker {
	if redisURL == "" {
		return NoopLocker{}
	}
	locker, err := NewRedisLocker(redisURL)
	if err != nil {
		return NoopLocker{}
	}
	return locker
}

// WithLock acquires the lock, runs fn, and releases. If the lock is held by
// another instance fn is skipped and (false, nil) returned. Errors from
// Acquire/Release propagate to the caller.
func WithLock(ctx context.Context, l Locker, key string, ttl time.Duration, fn func(ctx context.Context) error) (acquired bool, err error) {
	token, ok, err := l.Acquire(ctx, key, ttl)
	if err != nil || !ok {
		return false, err
	}
	defer func() {
		if relErr := l.Release(ctx, key, token); relErr != nil && err == nil && !errors.Is(relErr, ErrLockNotHeld) {
			err = relErr
		}
	}()
	if fnErr := fn(ctx); fnErr != nil {
		return true, fnErr
	}
	return true, nil
}
