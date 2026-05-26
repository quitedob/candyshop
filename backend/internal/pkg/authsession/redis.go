package authsession

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	modelsAuth "candypro/api/internal/models/auth"
	authRepo "candypro/api/internal/repository/auth"

	"github.com/redis/go-redis/v9"
)

const (
	redisPrefixRefresh = "candypro:auth:refresh:"
	redisPrefixAccess  = "candypro:auth:access:"
	redisUserRefreshs  = "candypro:auth:user:%s:refreshs"
	redisUserAccess    = "candypro:auth:user:%s:access"
)

type refreshPayload struct {
	UserID    string    `json:"userId"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
	ExpiresAt time.Time `json:"expiresAt"`
	Revoked   bool      `json:"revoked"`
}

type accessPayload struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// RedisStore Redis 会话存储
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore 连接 Redis 并验证可用性
func NewRedisStore(redisURL string) (*RedisStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	log.Printf("authsession: using Redis for JWT sessions")
	return &RedisStore{client: client}, nil
}

// Ping 检查 Redis 是否可达（启动前校验，不保留连接）
func Ping(redisURL string) error {
	redisURL = strings.TrimSpace(redisURL)
	if redisURL == "" {
		return nil
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("REDIS_URL 格式无效: %w", err)
	}
	client := redis.NewClient(opts)
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("无法连接 %s: %w", redisURL, err)
	}
	return nil
}

// EnsureAvailable 配置了 REDIS_URL 时必须能连通，否则返回错误（阻止服务启动）
func EnsureAvailable(redisURL string) error {
	redisURL = strings.TrimSpace(redisURL)
	if redisURL == "" {
		return nil
	}
	if err := Ping(redisURL); err != nil {
		return fmt.Errorf("Redis 未启动或不可达 (%w)。请先执行: docker compose up -d redis", err)
	}
	return nil
}

func (r *RedisStore) UsesRedis() bool { return true }

func (r *RedisStore) SaveRefreshToken(ctx context.Context, token *modelsAuth.RefreshToken) error {
	if token == nil || token.Token == "" {
		return fmt.Errorf("invalid refresh token")
	}
	payload, err := json.Marshal(refreshPayload{
		UserID: token.UserID, IPAddress: token.IPAddress, UserAgent: token.UserAgent,
		ExpiresAt: token.ExpiresAt,
	})
	if err != nil {
		return err
	}
	ttl := time.Until(token.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Hour
	}
	key := redisPrefixRefresh + token.Token
	userKey := fmt.Sprintf(redisUserRefreshs, token.UserID)
	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, payload, ttl)
	pipe.SAdd(ctx, userKey, token.Token)
	pipe.Expire(ctx, userKey, ttl+24*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStore) GetRefreshToken(ctx context.Context, tokenStr string) (*modelsAuth.RefreshToken, error) {
	raw, err := r.client.Get(ctx, redisPrefixRefresh+tokenStr).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, err
	}
	var p refreshPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	if p.Revoked || p.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("expired or revoked")
	}
	return &modelsAuth.RefreshToken{
		UserID: p.UserID, Token: tokenStr, ExpiresAt: p.ExpiresAt,
		IPAddress: p.IPAddress, UserAgent: p.UserAgent,
	}, nil
}

func (r *RedisStore) RevokeRefreshToken(ctx context.Context, tokenStr string) error {
	tok, err := r.GetRefreshToken(ctx, tokenStr)
	if err != nil {
		return r.DeleteRefreshToken(ctx, tokenStr)
	}
	key := redisPrefixRefresh + tokenStr
	raw, _ := json.Marshal(refreshPayload{
		UserID: tok.UserID, IPAddress: tok.IPAddress, UserAgent: tok.UserAgent,
		ExpiresAt: tok.ExpiresAt, Revoked: true,
	})
	ttl := time.Until(tok.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Minute
	}
	return r.client.Set(ctx, key, raw, ttl).Err()
}

func (r *RedisStore) DeleteRefreshToken(ctx context.Context, tokenStr string) error {
	tok, _ := r.GetRefreshToken(ctx, tokenStr)
	pipe := r.client.Pipeline()
	pipe.Del(ctx, redisPrefixRefresh+tokenStr)
	if tok != nil {
		pipe.SRem(ctx, fmt.Sprintf(redisUserRefreshs, tok.UserID), tokenStr)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisStore) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	userKey := fmt.Sprintf(redisUserRefreshs, userID)
	tokens, err := r.client.SMembers(ctx, userKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	pipe := r.client.Pipeline()
	for _, t := range tokens {
		pipe.Del(ctx, redisPrefixRefresh+t)
	}
	pipe.Del(ctx, userKey)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStore) SaveAccessSession(ctx context.Context, jti string, sess AccessSession, ttl time.Duration) error {
	if jti == "" {
		return fmt.Errorf("jti required")
	}
	raw, err := json.Marshal(accessPayload{UserID: sess.UserID, Email: sess.Email, Role: sess.Role})
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	userKey := fmt.Sprintf(redisUserAccess, sess.UserID)
	pipe := r.client.Pipeline()
	pipe.Set(ctx, redisPrefixAccess+jti, raw, ttl)
	pipe.SAdd(ctx, userKey, jti)
	pipe.Expire(ctx, userKey, ttl+time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStore) ValidateAccessSession(ctx context.Context, jti string) (*AccessSession, error) {
	if strings.TrimSpace(jti) == "" {
		return nil, fmt.Errorf("missing jti")
	}
	raw, err := r.client.Get(ctx, redisPrefixAccess+jti).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStoreUnavailable, err)
	}
	var p accessPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	return &AccessSession{UserID: p.UserID, Email: p.Email, Role: p.Role}, nil
}

func (r *RedisStore) RevokeAccessSession(ctx context.Context, jti string) error {
	sess, err := r.ValidateAccessSession(ctx, jti)
	if err != nil {
		return r.client.Del(ctx, redisPrefixAccess+jti).Err()
	}
	pipe := r.client.Pipeline()
	pipe.Del(ctx, redisPrefixAccess+jti)
	pipe.SRem(ctx, fmt.Sprintf(redisUserAccess, sess.UserID), jti)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStore) RevokeAllUserAccessSessions(ctx context.Context, userID string) error {
	userKey := fmt.Sprintf(redisUserAccess, userID)
	jtis, err := r.client.SMembers(ctx, userKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	pipe := r.client.Pipeline()
	for _, jti := range jtis {
		pipe.Del(ctx, redisPrefixAccess+jti)
	}
	pipe.Del(ctx, userKey)
	_, err = pipe.Exec(ctx)
	return err
}

// PostgresRefreshAdapter PostgreSQL refresh token 回退（access 会话不校验 Redis）
type PostgresRefreshAdapter struct {
	repo *authRepo.RefreshTokenRepository
}

func NewPostgresRefreshAdapter(repo *authRepo.RefreshTokenRepository) *PostgresRefreshAdapter {
	return &PostgresRefreshAdapter{repo: repo}
}

func (p *PostgresRefreshAdapter) UsesRedis() bool { return false }

func (p *PostgresRefreshAdapter) SaveRefreshToken(ctx context.Context, token *modelsAuth.RefreshToken) error {
	return p.repo.Create(ctx, token)
}

func (p *PostgresRefreshAdapter) GetRefreshToken(ctx context.Context, tokenStr string) (*modelsAuth.RefreshToken, error) {
	return p.repo.FindByToken(ctx, tokenStr)
}

func (p *PostgresRefreshAdapter) RevokeRefreshToken(ctx context.Context, tokenStr string) error {
	token, err := p.repo.FindByToken(ctx, tokenStr)
	if err != nil {
		return err
	}
	now := time.Now()
	token.RevokedAt = &now
	return p.repo.Update(ctx, token)
}

func (p *PostgresRefreshAdapter) DeleteRefreshToken(ctx context.Context, tokenStr string) error {
	return p.repo.DeleteByToken(ctx, tokenStr)
}

func (p *PostgresRefreshAdapter) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	return p.repo.DeleteAllForUser(ctx, userID)
}

func (p *PostgresRefreshAdapter) SaveAccessSession(context.Context, string, AccessSession, time.Duration) error {
	return nil
}

func (p *PostgresRefreshAdapter) ValidateAccessSession(context.Context, string) (*AccessSession, error) {
	return nil, nil
}

func (p *PostgresRefreshAdapter) RevokeAccessSession(context.Context, string) error { return nil }

func (p *PostgresRefreshAdapter) RevokeAllUserAccessSessions(context.Context, string) error { return nil }

// NewFromConfig 使用 Redis（REDIS_URL 已配置时必选，失败返回 error）
func NewFromConfig(redisURL string, pgRepo *authRepo.RefreshTokenRepository) (Store, error) {
	if strings.TrimSpace(redisURL) != "" {
		return NewRedisStore(redisURL)
	}
	return NewPostgresRefreshAdapter(pgRepo), nil
}
