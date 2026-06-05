package handlers

import (
	"context"
	"fmt"

	"candypro/api/internal/config"
	"candypro/api/internal/handlers/admin"
	"candypro/api/internal/handlers/auth"
	userPortal "candypro/api/internal/handlers/customer"
	"candypro/api/internal/handlers/public"
	"candypro/api/internal/handlers/system"
	"candypro/api/internal/pkg/realtime"
	"candypro/api/internal/pkg/storage"
	repositoryCommon "candypro/api/internal/repository/common"
	servicesCommon "candypro/api/internal/services/common"

	"gorm.io/gorm"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	AdminPortal *admin.Handler
	AuthScope   *auth.Handler
	UserPortal  *userPortal.Handler
	Public      *public.Handler
	System      *system.Handler
	Storage     *storage.Manager
	Realtime    *realtime.Manager
}

// New creates a new Handlers instance
func New(cfg *config.Config, svcs *servicesCommon.Services, db *gorm.DB) (*Handlers, error) {
	if svcs == nil {
		svcs = &servicesCommon.Services{}
	}

	st, err := newStorage(cfg, db)
	if err != nil {
		return nil, err
	}

	broadcaster, err := realtime.NewBroadcaster(cfg.Security.RedisURL)
	if err != nil {
		return nil, err
	}
	rt := realtime.NewManager(
		broadcaster,
		realtime.OriginAllowList(cfg.Security.WSAllowedOrigins),
	)

	return &Handlers{
		AdminPortal: admin.NewHandler(cfg, svcs.AdminPortal, svcs.CountryPaymentPolicy, st, rt),
		AuthScope:   auth.NewHandler(cfg, svcs.AuthScope),
		UserPortal:  userPortal.NewHandler(cfg, svcs.UserPortal, st, rt),
		Public:      public.NewHandler(cfg, svcs.Public, st),
		System:      system.NewHandler(cfg, svcs.System),
		Storage:     st,
		Realtime:    rt,
	}, nil
}

func newStorage(cfg *config.Config, db *gorm.DB) (*storage.Manager, error) {
	var backend storage.StorageService
	switch cfg.Upload.StorageDriver {
	case "s3", "oss": // Alibaba Cloud OSS is S3-compatible
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
			return nil, fmt.Errorf("initialize %s storage: %w", cfg.Upload.StorageDriver, err)
		}
		backend = s3st
	case "", "local":
		backend = storage.NewLocalStorageService(cfg.Upload.UploadPath, cfg.Upload.UploadURL)
	default:
		return nil, fmt.Errorf("unsupported storage driver %q", cfg.Upload.StorageDriver)
	}

	var metadata *repositoryCommon.UploadedFileRepository
	if db != nil {
		metadata = repositoryCommon.NewUploadedFileRepository(db)
	}
	scanner := storage.NewWebhookScanner(cfg.Upload.VirusScanWebhookURL, cfg.Upload.VirusScanWebhookSecret, nil)
	return storage.NewManager(backend, metadata, scanner), nil
}
