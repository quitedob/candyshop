package authscopescope

import (
	"log"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/authsession"
	authrepo "candypro/api/internal/repository/auth"
	repositoryCommon "candypro/api/internal/repository/common"
	auth "candypro/api/internal/services/auth"
	user "candypro/api/internal/services/user"
)

type Services struct {
	User               *user.UserService
	Company            *user.CompanyService
	JWT                *auth.JWTService
	Auth               *auth.AuthService
	Sessions           authsession.Store
	PasswordResetToken *authrepo.PasswordResetTokenRepository
}

func New(repos *repositoryCommon.AuthScopeRepositories, cfg *config.Config) *Services {
	if repos == nil {
		return &Services{}
	}

	sessions, err := authsession.NewFromConfig(cfg.Security.RedisURL, repos.RefreshToken)
	if err != nil {
		log.Fatalf("JWT 会话存储初始化失败: %v", err)
	}
	jwtSvc := auth.NewJWTService(sessions, cfg)

	userSvc := user.NewUserService(repos.User)

	return &Services{
		User:               userSvc,
		Company:            user.NewCompanyService(repos.Company, userSvc),
		JWT:                jwtSvc,
		Auth:               auth.NewAuthService(repos.User, repos.Role, sessions, jwtSvc, cfg),
		Sessions:           sessions,
		PasswordResetToken: repos.PasswordResetToken,
	}
}
