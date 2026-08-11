package system

import (
	"context"
	"fmt"
	"log"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/eino"
	"candypro/api/internal/pkg/eino/retry"
	einotool "candypro/api/internal/pkg/eino/tool"
	paypalAdapter "candypro/api/internal/pkg/payment/paypal"
	stripeAdapter "candypro/api/internal/pkg/payment/stripe"
	servicesCommon "candypro/api/internal/services/common"
	tradeService "candypro/api/internal/services/trade"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"gorm.io/gorm"
)

type Handler struct {
	cfg                   *config.Config
	services              *servicesCommon.SystemServices
	aiService             *tradeService.AIService
	tradeAgent            adk.Agent
	agentReady            bool
	b2bCoordinatorAgent   adk.Agent
	b2bCoordinatorReady   bool
	orderProcessingAgent  adk.Agent
	orderProcessingReady  bool
	checkPointStore       *eino.PostgresCheckPointStore
	stripeAdapter         *stripeAdapter.Adapter
	paypalAdapter         *paypalAdapter.Adapter
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
		h.stripeAdapter = stripeAdapter.New(cfg.Stripe.SecretKey, cfg.Stripe.WebhookSecret)
		h.paypalAdapter = paypalAdapter.New(cfg.PayPal.ClientID, cfg.PayPal.ClientSecret, cfg.PayPal.WebhookID, cfg.PayPal.Sandbox)
	}

	return h
}

// IsAgentReady reports whether the TradeAgent initialized successfully.
func (h *Handler) IsAgentReady() bool { return h.agentReady }

// TradeAgent returns the TradeAgent for cross-handler wiring.
func (h *Handler) TradeAgent() adk.Agent { return h.tradeAgent }

// IsB2BCoordinatorReady reports whether the B2B DeepAgent initialized successfully.
func (h *Handler) IsB2BCoordinatorReady() bool { return h.b2bCoordinatorReady }

// B2BCoordinatorAgent returns the DeepAgent for cross-handler wiring.
func (h *Handler) B2BCoordinatorAgent() adk.Agent { return h.b2bCoordinatorAgent }

// IsOrderProcessingReady reports whether the P-E-R agent initialized successfully.
func (h *Handler) IsOrderProcessingReady() bool { return h.orderProcessingReady }

// OrderProcessingAgent returns the P-E-R agent for cross-handler wiring.
func (h *Handler) OrderProcessingAgent() adk.Agent { return h.orderProcessingAgent }

// AttachAgentToAIService wires the TradeAgent into the AIService so Generate() calls
// benefit from the full TradeAgent toolset instead of the single-tool legacy agent.
func (h *Handler) AttachAgentToAIService(agent adk.Agent) {
	if h.aiService != nil {
		h.aiService.AttachAgent(agent)
	}
}

// translateBatchFunc 返回 translate_content 工具同源的批量翻译回调。
func (h *Handler) translateBatchFunc() einotool.TranslateBatchFunc {
	if h.aiService == nil {
		return nil
	}
	return einotool.NewTranslateBatchFunc(func(ctx context.Context, sourceData map[string]string, targetLocales []string) (map[string]map[string]string, []string, error) {
		result, err := h.aiService.BatchTranslateFields(ctx, sourceData, targetLocales)
		if err != nil {
			return nil, nil, err
		}
		if result == nil {
			return map[string]map[string]string{}, nil, nil
		}
		return result.Fields, result.Warnings, nil
	})
}

// translateFunc returns a TranslateFunc for wiring into Eino agents, or nil if AI is disabled.
func (h *Handler) translateFunc() einotool.TranslateFunc {
	return einotool.TranslateFuncFromBatch(h.translateBatchFunc())
}

// InitCheckPointStore creates a PostgreSQL-backed checkpoint store and sets it on the AI service.
func (h *Handler) InitCheckPointStore(db *gorm.DB) {
	if db == nil || h.aiService == nil {
		return
	}
	store := eino.NewPostgresCheckPointStore(db)
	h.aiService.SetCheckPointStore(store)
	h.checkPointStore = store
}

// CheckPointStore returns the PostgreSQL checkpoint store for background cleanup.
func (h *Handler) CheckPointStore() *eino.PostgresCheckPointStore {
	return h.checkPointStore
}

// InitDeepAgent initializes the B2B DeepAgent coordinator (ProductExpert, PricingExpert, LogisticsExpert).
func (h *Handler) InitDeepAgent() error {
	ctx := context.Background()

	rawModel, modelErr := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   h.cfg.AI.OpenAIModel,
		APIKey:  h.cfg.AI.OpenAIAPIKey,
		BaseURL: h.cfg.AI.OpenAIBaseURL,
	})
	if modelErr != nil {
		return fmt.Errorf("init deep agent chat model: %w", modelErr)
	}
	chatModel := retry.New(rawModel, h.cfg.AI.RetryMaxAttempts, h.cfg.AI.RetryIntervalSec)

	var persister einotool.DocumentPersister
	if h.services != nil && h.services.Trade != nil {
		persister = h.services.Trade
	}

	a, err := eino.NewB2BCoordinatorAgent(ctx, chatModel, persister, newProductCatalogAdapter(h), newPricingAdapter(h))
	if err != nil {
		return err
	}
	h.b2bCoordinatorAgent = a
	h.b2bCoordinatorReady = true
	return nil
}

// InitOrderProcessingAgent initializes the Plan-Execute-Replan agent for structured order processing.
func (h *Handler) InitOrderProcessingAgent() error {
	ctx := context.Background()

	rawModel, modelErr := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   h.cfg.AI.OpenAIModel,
		APIKey:  h.cfg.AI.OpenAIAPIKey,
		BaseURL: h.cfg.AI.OpenAIBaseURL,
	})
	if modelErr != nil {
		return fmt.Errorf("init order processing agent chat model: %w", modelErr)
	}
	chatModel := retry.New(rawModel, h.cfg.AI.RetryMaxAttempts, h.cfg.AI.RetryIntervalSec)

	a, err := eino.NewOrderProcessingAgent(ctx, chatModel)
	if err != nil {
		return err
	}
	h.orderProcessingAgent = a
	h.orderProcessingReady = true
	return nil
}
