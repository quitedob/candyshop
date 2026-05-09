package admin

import (
	"log"

	"candypro/api/internal/config"
	"candypro/api/internal/storage"
	servicesCommon "candypro/api/internal/services/common"
	orderSvc "candypro/api/internal/services/order"
	tradeSvc "candypro/api/internal/services/trade"
)

type Handler struct {
	cfg                  *config.Config
	services             *servicesCommon.AdminPortalServices
	countryPaymentPolicy *orderSvc.CountryPaymentPolicyService
	aiService            *tradeSvc.AIService
	storage              storage.StorageService
	Translations         *TranslationHandler
}

func NewHandler(cfg *config.Config, svcs *servicesCommon.AdminPortalServices, policySvc *orderSvc.CountryPaymentPolicyService, st storage.StorageService) *Handler {
	h := &Handler{
		cfg:                  cfg,
		services:             svcs,
		countryPaymentPolicy: policySvc,
		storage:              st,
	}
	if svcs != nil && svcs.Translation != nil {
		h.Translations = NewTranslationHandler(svcs.Translation)
	}
	if cfg != nil {
		aiSvc, err := tradeSvc.NewAIService(cfg.AI)
		if err != nil && err != tradeSvc.ErrAIServiceDisabled {
			log.Printf("admin AI init: %v", err)
		} else {
			h.aiService = aiSvc
		}
	}
	return h
}
