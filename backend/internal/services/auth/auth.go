package auth

import (
	"candypro/api/internal/config"
	modelsAuth "candypro/api/internal/models/auth"
	modelsUser "candypro/api/internal/models/user"
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

type refreshTokenRepository interface {
	FindByToken(ctx context.Context, token string) (*modelsAuth.RefreshToken, error)
	Update(ctx context.Context, token *modelsAuth.RefreshToken) error
	DeleteByToken(ctx context.Context, token string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}

type AuthService struct {
	userRepo   userRepository
	roleRepo   roleRepository
	tokenRepo  refreshTokenRepository
	jwtService *JWTService
	cfg        *config.Config
}

func NewAuthService(
	userRepo userRepository,
	roleRepo roleRepository,
	tokenRepo refreshTokenRepository,
	jwtService *JWTService,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		tokenRepo:  tokenRepo,
		jwtService: jwtService,
		cfg:        cfg,
	}
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

	// Hash password
	hash, err := pwdutil.HashPassword(user.PasswordHash) // Password string originally passed through model field
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
	token, err := s.tokenRepo.FindByToken(ctx, tokenStr)
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
	token, err := s.tokenRepo.FindByToken(ctx, tokenStr)
	if err != nil {
		return err
	}

	now := time.Now()
	token.RevokedAt = &now
	return s.tokenRepo.Update(ctx, token)
}

// Logout revokes a token
func (s *AuthService) Logout(ctx context.Context, tokenStr string) error {
	return s.tokenRepo.DeleteByToken(ctx, tokenStr)
}

// RevokeAllUserTokens revokes all refresh tokens for a user (e.g., after password change).
func (s *AuthService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return s.tokenRepo.DeleteAllForUser(ctx, userID)
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
