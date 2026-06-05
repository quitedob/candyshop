package customer

import (
	"candypro/api/internal/config"
	"candypro/api/internal/pkg/realtime"
	"candypro/api/internal/pkg/storage"
	servicesCommon "candypro/api/internal/services/common"
)

type Handler struct {
	cfg      *config.Config
	services *servicesCommon.UserPortalServices
	storage  *storage.Manager
	realtime *realtime.Manager
}

func NewHandler(cfg *config.Config, svcs *servicesCommon.UserPortalServices, st *storage.Manager, rt *realtime.Manager) *Handler {
	return &Handler{
		cfg:      cfg,
		services: svcs,
		storage:  st,
		realtime: rt,
	}
}
