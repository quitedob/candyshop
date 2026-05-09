package systemroutes

import (
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/middleware"

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

	systemProtected := group.Group("")
	systemProtected.Use(middleware.AuthMiddleware(cfg))
	{
		systemProtected.POST("/analyze-inquiry", h.System.AnalyzeInquiry)
		systemProtected.POST("/generate-quotation", h.System.GenerateQuotation)
		systemProtected.POST("/translate", h.System.Translate)
		systemProtected.GET("/conversations/:id", h.System.GetConversation)
	}
}
