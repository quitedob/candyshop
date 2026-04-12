package auth

import (
	modelsAuth "candypro/api/internal/models/auth"
	modelsUser "candypro/api/internal/models/user"
)

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"candypro/api/internal/config"
	"candypro/api/internal/roles"
	"candypro/api/internal/utils"
)

type JWTService struct {
	repo refreshTokenCreator
	cfg  *config.Config
}

type refreshTokenCreator interface {
	Create(ctx context.Context, token *modelsAuth.RefreshToken) error
}

func NewJWTService(repo refreshTokenCreator, cfg *config.Config) *JWTService {
	return &JWTService{repo: repo, cfg: cfg}
}

func (s *JWTService) GenerateAccessToken(user *modelsUser.User) (string, error) {
	roleName := roles.User
	if user.Role != nil && strings.TrimSpace(user.Role.Name) != "" {
		roleName = strings.TrimSpace(user.Role.Name)
	}

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  roleName,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Duration(s.cfg.JWT.AccessTokenDuration) * time.Minute).Unix(),
	}
	return utils.GenerateJWT(claims, s.cfg.JWT.Secret)
}

func (s *JWTService) GenerateRefreshToken(ctx context.Context, user *modelsUser.User, ipAddress, userAgent string) (string, error) {
	tokenString := utils.GenerateRandomString(64) // 64 chars random string for custom token

	token := &modelsAuth.RefreshToken{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.JWT.RefreshTokenDuration) * 24 * time.Hour),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := s.repo.Create(ctx, token); err != nil {
		return "", err
	}

	return tokenString, nil
}
