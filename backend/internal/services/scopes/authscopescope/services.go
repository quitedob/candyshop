package authscopescope

import (
	"candypro/api/internal/config"
	repositoryCommon "candypro/api/internal/repository/common"
	auth "candypro/api/internal/services/auth"
	user "candypro/api/internal/services/user"
)

type Services struct {
	User *user.UserService
	JWT  *auth.JWTService
	Auth *auth.AuthService
}

func New(repos *repositoryCommon.AuthScopeRepositories, cfg *config.Config) *Services {
	if repos == nil {
		return &Services{}
	}

	jwtSvc := auth.NewJWTService(repos.RefreshToken, cfg)

	return &Services{
		User: user.NewUserService(repos.User),
		JWT:  jwtSvc,
		Auth: auth.NewAuthService(repos.User, repos.Role, repos.RefreshToken, jwtSvc, cfg),
	}
}
