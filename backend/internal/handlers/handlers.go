package handlers

import (
	"context"
	"log"

	"candypro/api/internal/config"
	"candypro/api/internal/handlers/admin"
	"candypro/api/internal/handlers/auth"
	userPortal "candypro/api/internal/handlers/customer"
	"candypro/api/internal/handlers/public"
	"candypro/api/internal/handlers/system"
	"candypro/api/internal/pkg/storage"
	servicesCommon "candypro/api/internal/services/common"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	AdminPortal *admin.Handler
	AuthScope   *auth.Handler
	UserPortal  *userPortal.Handler
	Public      *public.Handler
	System      *system.Handler
	Storage     storage.StorageService
}

// New creates a new Handlers instance
func New(cfg *config.Config, svcs *servicesCommon.Services) *Handlers {
	if svcs == nil {
		svcs = &servicesCommon.Services{}
	}

	st := newStorage(cfg)

	return &Handlers{
		AdminPortal: admin.NewHandler(cfg, svcs.AdminPortal, svcs.CountryPaymentPolicy, st),
		AuthScope:   auth.NewHandler(cfg, svcs.AuthScope),
		UserPortal:  userPortal.NewHandler(cfg, svcs.UserPortal, st),
		Public:      public.NewHandler(cfg, svcs.Public, st),
		System:      system.NewHandler(cfg, svcs.System),
		Storage:     st,
	}
}

func newStorage(cfg *config.Config) storage.StorageService {
	switch cfg.Upload.StorageDriver {
	case "s3":
		s3st, err := storage.NewS3StorageService(
			context.Background(),
			storage.S3Config{
				Bucket:    cfg.Upload.S3Bucket,
				Region:    cfg.Upload.S3Region,
				AccessKey: cfg.Upload.S3AccessKey,
				SecretKey: cfg.Upload.S3SecretKey,
				Endpoint:  cfg.Upload.S3Endpoint,
				CDNDomain: cfg.Upload.S3CDNDomain,
			},
		)
		if err != nil {
			log.Printf("Failed to init S3 storage, falling back to local: %v", err)
			return storage.NewLocalStorageService(cfg.Upload.UploadPath, cfg.Upload.UploadURL)
		}
		return s3st
	default:
		return storage.NewLocalStorageService(cfg.Upload.UploadPath, cfg.Upload.UploadURL)
	}
}
