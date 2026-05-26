package auth

import (
	"candypro/api/internal/config"
	modelsAuth "candypro/api/internal/models/auth"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/pkg/authsession"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/jwtutil"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	sessions authsession.Store
	cfg      *config.Config
}

func NewJWTService(sessions authsession.Store, cfg *config.Config) *JWTService {
	return &JWTService{sessions: sessions, cfg: cfg}
}

// GenerateAccessToken 签发 JWT 并将 jti 会话写入 Redis（启用时）
func (s *JWTService) GenerateAccessToken(ctx context.Context, user *modelsUser.User) (string, error) {
	roleName := modelsAuth.User
	if user.Role != nil && strings.TrimSpace(user.Role.Name) != "" {
		roleName = strings.TrimSpace(user.Role.Name)
	}

	jti := crypto.GenerateID()
	ttl := time.Duration(s.cfg.JWT.AccessTokenDuration) * time.Minute
	now := time.Now()

	claims := jwt.MapClaims{
		"jti":   jti,
		"sub":   user.ID,
		"email": user.Email,
		"role":  roleName,
		"iat":   now.Unix(),
		"exp":   now.Add(ttl).Unix(),
	}
	token, err := jwtutil.GenerateJWT(claims, s.cfg.JWT.Secret)
	if err != nil {
		return "", err
	}

	if s.sessions != nil && s.sessions.UsesRedis() {
		if err := s.sessions.SaveAccessSession(ctx, jti, authsession.AccessSession{
			UserID: user.ID,
			Email:  user.Email,
			Role:   roleName,
		}, ttl); err != nil {
			return "", fmt.Errorf("save access session: %w", err)
		}
	}
	return token, nil
}

func (s *JWTService) GenerateRefreshToken(ctx context.Context, user *modelsUser.User, ipAddress, userAgent string) (string, error) {
	tokenString := crypto.GenerateRandomString(64)
	token := &modelsAuth.RefreshToken{
		ID:        crypto.GenerateID(),
		UserID:    user.ID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.JWT.RefreshTokenDuration) * 24 * time.Hour),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	if err := s.sessions.SaveRefreshToken(ctx, token); err != nil {
		return "", err
	}
	return tokenString, nil
}

// RevokeAccessTokenByString 吊销 access JWT（按 jti）
func (s *JWTService) RevokeAccessTokenByString(ctx context.Context, accessToken string) {
	if s.sessions == nil || !s.sessions.UsesRedis() || accessToken == "" {
		return
	}
	claims, err := jwtutil.ParseClaimsAllowExpired(accessToken, s.cfg.JWT.Secret)
	if err != nil {
		return
	}
	jti, _ := claims["jti"].(string)
	if jti != "" {
		if err := s.sessions.RevokeAccessSession(ctx, jti); err != nil {
			log.Printf("auth: RevokeAccessSession failed for jti=%s: %v", jti, err)
		}
	}
}
