package common

import (
	"candypro/api/internal/config"
	repositoryCommon "candypro/api/internal/repository/common"
	adminportalscope "candypro/api/internal/services/scopes/adminportalscope"
	authscopescope "candypro/api/internal/services/scopes/authscopescope"
	publicscope "candypro/api/internal/services/scopes/publicscope"
	systemscope "candypro/api/internal/services/scopes/systemscope"
	userportalscope "candypro/api/internal/services/scopes/userportalscope"
	order "candypro/api/internal/services/order"
	searchsvc "candypro/api/internal/services/search"

	"gorm.io/gorm"
)

type PublicServices = publicscope.Services
type UserPortalServices = userportalscope.Services
type AdminPortalServices = adminportalscope.Services
type AuthScopeServices = authscopescope.Services
type SystemServices = systemscope.Services
type SearchService = searchsvc.SearchService

// Services aggregates all scoped service instances.
type Services struct {
	Public                *PublicServices
	UserPortal            *UserPortalServices
	AdminPortal           *AdminPortalServices
	AuthScope             *AuthScopeServices
	System                *SystemServices
	CountryPaymentPolicy  *order.CountryPaymentPolicyService
}

// NewServices creates all scoped services.
func NewServices(repos *repositoryCommon.Repositories, cfg *config.Config, db *gorm.DB) *Services {
	if repos == nil {
		return &Services{}
	}

	publicSvcs := publicscope.New(repos.Public, cfg)
	authScopeSvcs := authscopescope.New(repos.AuthScope, cfg)

	return &Services{
		Public:               publicSvcs,
		UserPortal:           userportalscope.New(repos.UserPortal, cfg, db),
		AdminPortal:          adminportalscope.New(repos.AdminPortal, cfg, authScopeSvcs.Auth, db),
		AuthScope:            authScopeSvcs,
		System:               systemscope.New(repos.System, cfg, publicSvcs.Search),
		CountryPaymentPolicy: order.NewCountryPaymentPolicyService(db),
	}
}
