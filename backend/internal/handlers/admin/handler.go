package admin

import (
	"candypro/api/internal/config"
	servicesCommon "candypro/api/internal/services/common"
)

type Handler struct {
	cfg      *config.Config
	services *servicesCommon.AdminPortalServices
}

func NewHandler(cfg *config.Config, svcs *servicesCommon.AdminPortalServices) *Handler {
	return &Handler{
		cfg:      cfg,
		services: svcs,
	}
}
