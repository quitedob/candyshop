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
	agentReady bool
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

// IsAgentReady reports whether the TradeAgent initialized successfully.
func (h *Handler) IsAgentReady() bool { return h.agentReady }

// TradeAgent returns the 13-tool TradeAgent for cross-handler wiring.
func (h *Handler) TradeAgent() adk.Agent { return h.tradeAgent }

// AttachAgentToAIService wires the TradeAgent into the AIService so Generate() calls
// benefit from the full 13-tool agent instead of the single-tool legacy agent.
func (h *Handler) AttachAgentToAIService(agent adk.Agent) {
	if h.aiService != nil {
		h.aiService.AttachAgent(agent)
	}
}
