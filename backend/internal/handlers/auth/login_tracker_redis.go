package auth

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// LoginTracker 登录暴力破解防护接口（内存或 Redis）
type LoginTracker interface {
	IsLocked(email string) (bool, int)
	RecordFailure(email string)
	RecordSuccess(email string)
	Stop()
}

// redisLoginTracker 基于 Redis 的 per-account 登录锁定
type redisLoginTracker struct {
	client *redis.Client
}

func newRedisLoginTracker(redisURL string) (LoginTracker, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	log.Printf("auth: login tracker using Redis")
	return &redisLoginTracker{client: client}, nil
}

func (r *redisLoginTracker) failKey(email string) string {
	return "candypro:auth:login:fail:" + strings.ToLower(strings.TrimSpace(email))
}

func (r *redisLoginTracker) lockKey(email string) string {
	return "candypro:auth:login:lock:" + strings.ToLower(strings.TrimSpace(email))
}

func (r *redisLoginTracker) IsLocked(email string) (bool, int) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	ttl, err := r.client.TTL(ctx, r.lockKey(email)).Result()
	if err != nil || ttl <= 0 {
		return false, 0
	}
	return true, int(ttl.Seconds())
}

func (r *redisLoginTracker) RecordFailure(email string) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	failKey := r.failKey(email)
	count, err := r.client.Incr(ctx, failKey).Result()
	if err != nil {
		log.Printf("auth: login tracker INCR failed for %s: %v", email, err)
		return
	}
	if count == 1 {
		if err := r.client.Expire(ctx, failKey, lockoutDuration).Err(); err != nil {
			log.Printf("auth: login tracker EXPIRE failed for %s: %v", email, err)
		}
	}
	if int(count) >= maxLoginAttempts {
		if err := r.client.Set(ctx, r.lockKey(email), "1", lockoutDuration).Err(); err != nil {
			log.Printf("auth: login tracker lock SET failed for %s: %v", email, err)
		}
	}
}

func (r *redisLoginTracker) RecordSuccess(email string) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := r.client.Del(ctx, r.failKey(email), r.lockKey(email)).Err(); err != nil {
		log.Printf("auth: login tracker clear failed for %s: %v", email, err)
	}
}

func (r *redisLoginTracker) Stop() {}

// NewLoginTrackerFromEnv 创建登录锁定 tracker；Redis 不可用时回退内存
func NewLoginTrackerFromEnv() LoginTracker {
	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if redisURL == "" {
		return NewLoginAttemptTracker()
	}
	tracker, err := newRedisLoginTracker(redisURL)
	if err != nil {
		log.Printf("auth: Redis login tracker unavailable (%v), using in-memory", err)
		return NewLoginAttemptTracker()
	}
	return tracker
}
