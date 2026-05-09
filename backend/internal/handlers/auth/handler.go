package auth

import (
	"candypro/api/internal/config"
	servicesCommon "candypro/api/internal/services/common"
)

type Handler struct {
	cfg          *config.Config
	services     *servicesCommon.AuthScopeServices
	loginTracker *LoginAttemptTracker
}

func NewHandler(cfg *config.Config, svcs *servicesCommon.AuthScopeServices) *Handler {
	return &Handler{
		cfg:          cfg,
		services:     svcs,
		loginTracker: NewLoginAttemptTracker(),
	}
}

// StopLoginTracker stops the login attempt tracker's background cleanup goroutine.
func (h *Handler) StopLoginTracker() {
	if h.loginTracker != nil {
		h.loginTracker.Stop()
	}
}
