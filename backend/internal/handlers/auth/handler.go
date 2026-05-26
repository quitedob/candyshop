package auth

import (
	"candypro/api/internal/config"
	"candypro/api/internal/pkg/authsession"
	servicesCommon "candypro/api/internal/services/common"
)

type Handler struct {
	cfg          *config.Config
	services     *servicesCommon.AuthScopeServices
	loginTracker LoginTracker
}

// SessionStore 返回 JWT 会话存储（Redis 或 PostgreSQL 回退）
func (h *Handler) SessionStore() authsession.Store {
	if h.services == nil {
		return nil
	}
	return h.services.Sessions
}

func NewHandler(cfg *config.Config, svcs *servicesCommon.AuthScopeServices) *Handler {
	return &Handler{
		cfg:          cfg,
		services:     svcs,
		loginTracker: NewLoginTrackerFromEnv(),
	}
}

// StopLoginTracker stops the login attempt tracker's background cleanup goroutine.
func (h *Handler) StopLoginTracker() {
	if h.loginTracker != nil {
		h.loginTracker.Stop()
	}
}
