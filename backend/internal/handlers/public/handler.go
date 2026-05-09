package public

import (
	"candypro/api/internal/config"
	"candypro/api/internal/storage"
	servicesCommon "candypro/api/internal/services/common"
)

type Handler struct {
	cfg      *config.Config
	services *servicesCommon.PublicServices
	storage  storage.StorageService
}

func NewHandler(cfg *config.Config, svcs *servicesCommon.PublicServices, st storage.StorageService) *Handler {
	return &Handler{
		cfg:      cfg,
		services: svcs,
		storage:  st,
	}
}
