package authscopescope

import (
	auth "candypro/api/internal/repository/auth"
	user "candypro/api/internal/repository/user"

	"gorm.io/gorm"
)

type Repositories struct {
	User         *user.UserRepository
	Role         *auth.RoleRepository
	RefreshToken *auth.RefreshTokenRepository
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		User:         user.NewUserRepository(db),
		Role:         auth.NewRoleRepository(db),
		RefreshToken: auth.NewRefreshTokenRepository(db),
	}
}
