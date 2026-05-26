// Package authsession 将会话（JWT access + refresh token）存入 Redis，支持分布式校验与吊销。
package authsession

import (
	"context"
	"errors"
	"time"

	modelsAuth "candypro/api/internal/models/auth"
)

// ErrSessionNotFound Redis 中不存在该 access session（已吊销、未写入或 key 被驱逐）
var ErrSessionNotFound = errors.New("session not found")

// ErrStoreUnavailable Redis 基础设施错误（网络/超时），调用方可降级为纯 JWT 校验
var ErrStoreUnavailable = errors.New("store unavailable")

// AccessSession access token 会话元数据（存 Redis，TTL = JWT 过期时间）
type AccessSession struct {
	UserID string
	Email  string
	Role   string
}

// Store 会话持久化接口（Redis 为主，PostgreSQL 仅 refresh 回退）
type Store interface {
	UsesRedis() bool

	SaveRefreshToken(ctx context.Context, token *modelsAuth.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenStr string) (*modelsAuth.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenStr string) error
	DeleteRefreshToken(ctx context.Context, tokenStr string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error

	SaveAccessSession(ctx context.Context, jti string, sess AccessSession, ttl time.Duration) error
	ValidateAccessSession(ctx context.Context, jti string) (*AccessSession, error)
	RevokeAccessSession(ctx context.Context, jti string) error
	RevokeAllUserAccessSessions(ctx context.Context, userID string) error
}
