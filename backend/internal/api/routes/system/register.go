package systemroutes

import (
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/middleware"
	modelsAuth "candypro/api/internal/models/auth"

	"github.com/gin-gonic/gin"
)

// Register wires system routes under /api/v1/system.
// publicAI 为针对未登录 AI 的额外中间件链（关闭开关 + 独立限流），与全局限流叠加。
func Register(group *gin.RouterGroup, h *handlers.Handlers, cfg *config.Config, publicAI ...gin.HandlerFunc) {
	pub := group.Group("")
	for _, mw := range publicAI {
		pub.Use(mw)
	}
	pub.POST("/chatbot", h.System.Chatbot)
	pub.POST("/recommend-products", h.System.RecommendProducts)
	pub.POST("/search", h.System.AISearch)

	// Payment webhooks live OUTSIDE the public-AI middleware chain: the AI
	// disable switch would 503 every webhook when AI routes are turned off, and
	// the AI rate limiter (15/min/IP) would 429-drop payment confirmations.
	// Each handler performs its own gateway signature verification instead.
	webhooks := group.Group("")
	webhooks.POST("/stripe-webhook", h.System.HandleStripeWebhook)
	webhooks.POST("/paypal-webhook", h.System.HandlePayPalWebhook)

	systemProtected := group.Group("")
	systemProtected.Use(middleware.AuthMiddleware(cfg, h.AuthScope.SessionStore()))
	{
		systemProtected.GET("/ai/trade-assistant", h.System.HandleTradeChat)
		systemProtected.GET("/ai/b2b-coordinator", h.System.HandleB2BCoordinatorChat)
		systemProtected.GET("/ai/order-processing", h.System.HandleOrderProcessingChat)
		systemProtected.GET("/ai/config", h.System.SystemAIConfig)
		systemProtected.POST("/analyze-inquiry", h.System.AnalyzeInquiry)
		// 生成报价会向模型与响应注入内部成本栈（物流/关税/标签摊销/目标毛利），
		// 属内部定价数据，仅限管理员访问；角色守卫防止任何已登录客户读取成本结构。
		systemProtected.POST("/generate-quotation", middleware.RequireRole(modelsAuth.AdminPortal()...), h.System.GenerateQuotation)
		systemProtected.POST("/chatbot/order", h.System.ChatbotOrderContext)
		systemProtected.POST("/translate", h.System.Translate)
		systemProtected.GET("/conversations/:id", h.System.GetConversation)
	}
}
