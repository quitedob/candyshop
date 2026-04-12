package system

import (
	"log"

	"candypro/api/internal/config"
	servicesCommon "candypro/api/internal/services/common"
	tradeService "candypro/api/internal/services/trade"

	"github.com/cloudwego/eino/adk"
)

type Handler struct {
	cfg        *config.Config
	services   *servicesCommon.SystemServices
	aiService  *tradeService.AIService
	tradeAgent adk.Agent
}

func NewHandler(cfg *config.Config, svcs *servicesCommon.SystemServices) *Handler {
	h := &Handler{
		cfg:      cfg,
		services: svcs,
	}

	if cfg != nil {
		aiSvc, err := tradeService.NewAIService(cfg.AI)
		if err != nil && err != tradeService.ErrAIServiceDisabled {
			log.Printf("failed to initialize AI service: %v", err)
		} else {
			h.aiService = aiSvc
		}
	}

	return h
}
