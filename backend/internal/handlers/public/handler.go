package public

import (
	"candypro/api/internal/config"
	servicesCommon "candypro/api/internal/services/common"
)

type Handler struct {
	cfg      *config.Config
	services *servicesCommon.PublicServices
}

func NewHandler(cfg *config.Config, svcs *servicesCommon.PublicServices) *Handler {
	return &Handler{
		cfg:      cfg,
		services: svcs,
	}
}
