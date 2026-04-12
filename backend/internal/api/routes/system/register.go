package systemroutes

import (
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Register wires system routes under /api/v1/system.
func Register(group *gin.RouterGroup, h *handlers.Handlers, cfg *config.Config) {
	group.POST("/chatbot", h.System.Chatbot)
	group.POST("/recommend-products", h.System.RecommendProducts)
	group.POST("/search", h.System.AISearch)

	systemProtected := group.Group("")
	systemProtected.Use(middleware.AuthMiddleware(cfg))
	{
		systemProtected.POST("/analyze-inquiry", h.System.AnalyzeInquiry)
		systemProtected.POST("/generate-quotation", h.System.GenerateQuotation)
		systemProtected.POST("/translate", h.System.Translate)
		systemProtected.GET("/conversations/:id", h.System.GetConversation)
	}
}
