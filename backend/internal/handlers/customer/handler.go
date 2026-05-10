package customer

import (
	"candypro/api/internal/config"
	"candypro/api/internal/pkg/storage"
	servicesCommon "candypro/api/internal/services/common"
)

type Handler struct {
	cfg      *config.Config
	services *servicesCommon.UserPortalServices
	storage  storage.StorageService
}

func NewHandler(cfg *config.Config, svcs *servicesCommon.UserPortalServices, st storage.StorageService) *Handler {
	return &Handler{
		cfg:      cfg,
		services: svcs,
		storage:  st,
	}
}
