package ratelimit

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisStore 使用 Redis 固定窗口计数实现分布式限流
type redisStore struct {
	client *redis.Client
	prefix string
}

// NewRedisStore 创建 Redis 限流存储；连接失败时返回 error
func NewRedisStore(redisURL string) (Store, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &redisStore{client: client, prefix: "candypro:rl:"}, nil
}

func (r *redisStore) Allow(key string, limit int, window time.Duration) bool {
	if r.client == nil || limit <= 0 {
		return true
	}
	windowSec := int64(window.Seconds())
	if windowSec <= 0 {
		windowSec = 60
	}
	bucket := time.Now().Unix() / windowSec
	redisKey := fmt.Sprintf("%s%s:%d", r.prefix, key, bucket)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	count, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		log.Printf("ratelimit redis INCR failed: %v", err)
		return true
	}
	if count == 1 {
		_ = r.client.Expire(ctx, redisKey, window).Err()
	}
	return int(count) <= limit
}

// Close 关闭 Redis 连接
func (r *redisStore) Close() error {
	if r.client == nil {
		return nil
	}
	return r.client.Close()
}

// NewFromEnv 优先 Redis，失败或未配置时回退内存
//
// R2 C-3: when the in-memory store is selected, log loudly so operators
// running multi-instance deployments notice their per-IP counters aren't
// shared. In-memory limits are correct for a single-process dev run; in
// production they let attackers multiply their effective budget by the
// instance count. Set REDIS_URL to make this go away.
func NewFromEnv(redisURL string) Store {
	if redisURL == "" {
		log.Printf("ratelimit: REDIS_URL not set — using in-memory store (NOT SAFE for multi-instance deployments)")
		return NewMemoryStore()
	}
	store, err := NewRedisStore(redisURL)
	if err != nil {
		log.Printf("ratelimit: Redis unavailable (%v), falling back to in-memory store (NOT SAFE for multi-instance deployments)", err)
		return NewMemoryStore()
	}
	log.Printf("ratelimit: using Redis store")
	return store
}

// WindowKey 生成限流 bucket key（供测试使用）
func WindowKey(window time.Duration) int64 {
	sec := int64(window.Seconds())
	if sec <= 0 {
		sec = 60
	}
	return time.Now().Unix() / sec
}

// FormatBucketKey 格式化 bucket（测试辅助）
func FormatBucketKey(prefix, key string, bucket int64) string {
	return prefix + key + ":" + strconv.FormatInt(bucket, 10)
}
