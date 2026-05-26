package authscopescope

import (
	auth "candypro/api/internal/repository/auth"
	userRepo "candypro/api/internal/repository/user"

	"gorm.io/gorm"
)

type Repositories struct {
	User               *userRepo.UserRepository
	Company            *userRepo.CompanyRepository
	Role               *auth.RoleRepository
	RefreshToken       *auth.RefreshTokenRepository
	PasswordResetToken *auth.PasswordResetTokenRepository
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		User:               userRepo.NewUserRepository(db),
		Company:            userRepo.NewCompanyRepository(db),
		Role:               auth.NewRoleRepository(db),
		RefreshToken:       auth.NewRefreshTokenRepository(db),
		PasswordResetToken: auth.NewPasswordResetTokenRepository(db),
	}
}
