package customer

import (
	"candypro/api/internal/config"
	servicesCommon "candypro/api/internal/services/common"
)

type Handler struct {
	cfg      *config.Config
	services *servicesCommon.UserPortalServices
}

func NewHandler(cfg *config.Config, svcs *servicesCommon.UserPortalServices) *Handler {
	return &Handler{
		cfg:      cfg,
		services: svcs,
	}
}
