package common

import (
	dbscopes "candypro/api/internal/database/scopes"
	adminportalscope "candypro/api/internal/repository/scopes/adminportalscope"
	authscopescope "candypro/api/internal/repository/scopes/authscopescope"
	publicscope "candypro/api/internal/repository/scopes/publicscope"
	systemscope "candypro/api/internal/repository/scopes/systemscope"
	userportalscope "candypro/api/internal/repository/scopes/userportalscope"

	"gorm.io/gorm"
)

type PublicRepositories = publicscope.Repositories
type UserPortalRepositories = userportalscope.Repositories
type AdminPortalRepositories = adminportalscope.Repositories
type AuthScopeRepositories = authscopescope.Repositories
type SystemRepositories = systemscope.Repositories

// Repositories holds all repositories
type Repositories struct {
	Public      *PublicRepositories
	UserPortal  *UserPortalRepositories
	AdminPortal *AdminPortalRepositories
	AuthScope   *AuthScopeRepositories
	System      *SystemRepositories
}

// NewRepositories creates all repositories
func NewRepositories(db *gorm.DB) *Repositories {
	scoped := dbscopes.NewConnections(db)

	return &Repositories{
		Public:      publicscope.New(scoped.Public),
		UserPortal:  userportalscope.New(scoped.UserPortal),
		AdminPortal: adminportalscope.New(scoped.AdminPortal),
		AuthScope:   authscopescope.New(scoped.AuthScope),
		System:      systemscope.New(scoped.System),
	}
}
