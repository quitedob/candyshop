package auth

import (
	"candypro/api/internal/config"
	modelsAuth "candypro/api/internal/models/auth"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/pkg/authsession"
	"candypro/api/internal/pkg/crypto"
	pwdutil "candypro/api/internal/pkg/password"
	"context"
	"errors"
	"strings"
	"time"
)

type userRepository interface {
	Create(ctx context.Context, user *modelsUser.User) error
	FindByEmail(ctx context.Context, email string) (*modelsUser.User, error)
}

type roleRepository interface {
	FindByName(ctx context.Context, name string) (*modelsAuth.Role, error)
	FindByID(ctx context.Context, id string) (*modelsAuth.Role, error)
}

type AuthService struct {
	userRepo   userRepository
	roleRepo   roleRepository
	sessions   authsession.Store
	jwtService *JWTService
	cfg        *config.Config
}

func NewAuthService(
	userRepo userRepository,
	roleRepo roleRepository,
	sessions authsession.Store,
	jwtService *JWTService,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		sessions:   sessions,
		jwtService: jwtService,
		cfg:        cfg,
	}
}

// Sessions 返回会话存储（供 middleware 校验 Redis access 会话）
func (s *AuthService) Sessions() authsession.Store {
	return s.sessions
}

// Register creates a new user, hashes their password, and sets up a default profile
func (s *AuthService) Register(ctx context.Context, user *modelsUser.User) error {
	if user.RoleID == "" && s.roleRepo != nil {
		if customerRole, roleErr := s.roleRepo.FindByName(ctx, modelsAuth.User); roleErr == nil {
			user.RoleID = customerRole.ID
			user.Role = &modelsUser.RoleSnapshot{
				ID:          customerRole.ID,
				Name:        customerRole.Name,
				Description: customerRole.Description,
				Permissions: customerRole.Permissions,
				IsSystem:    customerRole.IsSystem,
			}
		}
	}

	hash, err := pwdutil.HashPassword(user.PasswordHash)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	user.ID = crypto.GenerateID()

	return s.userRepo.Create(ctx, user)
}

// ValidateCredentials checks an email and password
func (s *AuthService) ValidateCredentials(ctx context.Context, email, password string) (*modelsUser.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !pwdutil.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	if user.Role == nil && user.RoleID != "" && s.roleRepo != nil {
		if role, roleErr := s.roleRepo.FindByID(ctx, user.RoleID); roleErr == nil {
			user.Role = &modelsUser.RoleSnapshot{
				ID:          role.ID,
				Name:        role.Name,
				Description: role.Description,
				Permissions: role.Permissions,
				IsSystem:    role.IsSystem,
			}
		}
	}
	if user.Role == nil {
		user.Role = &modelsUser.RoleSnapshot{Name: modelsAuth.User}
	}

	return user, nil
}

// ValidateRefreshToken checks if a refresh token is valid and not expired/revoked
func (s *AuthService) ValidateRefreshToken(ctx context.Context, tokenStr string) (string, error) {
	token, err := s.sessions.GetRefreshToken(ctx, tokenStr)
	if err != nil {
		return "", errors.New("invalid or expired refresh token")
	}
	if token.RevokedAt != nil || token.ExpiresAt.Before(time.Now()) {
		return "", errors.New("invalid or expired refresh token")
	}
	return token.UserID, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (s *AuthService) RevokeRefreshToken(ctx context.Context, tokenStr string) error {
	return s.sessions.RevokeRefreshToken(ctx, tokenStr)
}

// Logout revokes a token
func (s *AuthService) Logout(ctx context.Context, tokenStr string) error {
	return s.sessions.DeleteRefreshToken(ctx, tokenStr)
}

// RevokeAllUserTokens revokes all refresh + access sessions for a user
func (s *AuthService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	_ = s.sessions.RevokeAllUserAccessSessions(ctx, userID)
	return s.sessions.RevokeAllUserRefreshTokens(ctx, userID)
}

// ResolveRoleID resolves a role ID from explicit roleID or roleName.
func (s *AuthService) ResolveRoleID(ctx context.Context, roleID, roleName string) (string, error) {
	if s.roleRepo == nil {
		return "", errors.New("role service unavailable")
	}

	if strings.TrimSpace(roleID) != "" {
		role, err := s.roleRepo.FindByID(ctx, strings.TrimSpace(roleID))
		if err != nil || role == nil {
			return "", errors.New("invalid role id")
		}
		return role.ID, nil
	}

	if strings.TrimSpace(roleName) != "" {
		role, err := s.roleRepo.FindByName(ctx, strings.TrimSpace(roleName))
		if err != nil || role == nil {
			return "", errors.New("invalid role name")
		}
		return role.ID, nil
	}

	return "", errors.New("roleId or roleName is required")
}
