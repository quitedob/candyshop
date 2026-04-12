package handlers

import (
	"candypro/api/internal/config"
	"candypro/api/internal/handlers/admin"
	"candypro/api/internal/handlers/auth"
	userPortal "candypro/api/internal/handlers/customer"
	"candypro/api/internal/handlers/public"
	"candypro/api/internal/handlers/system"
	servicesCommon "candypro/api/internal/services/common"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	AdminPortal *admin.Handler
	AuthScope   *auth.Handler
	UserPortal  *userPortal.Handler
	Public      *public.Handler
	System      *system.Handler
}

// New creates a new Handlers instance
func New(cfg *config.Config, svcs *servicesCommon.Services) *Handlers {
	if svcs == nil {
		svcs = &servicesCommon.Services{}
	}

	return &Handlers{
		AdminPortal: admin.NewHandler(cfg, svcs.AdminPortal),
		AuthScope:   auth.NewHandler(cfg, svcs.AuthScope),
		UserPortal:  userPortal.NewHandler(cfg, svcs.UserPortal),
		Public:      public.NewHandler(cfg, svcs.Public),
		System:      system.NewHandler(cfg, svcs.System),
	}
}
