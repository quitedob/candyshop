package publicroutes

import (
	"candypro/api/internal/handlers"

	"github.com/gin-gonic/gin"
)

// Register wires all public API routes under /api/v1/public.
func Register(group *gin.RouterGroup, h *handlers.Handlers) {
	// Products
	group.GET("/products", h.Public.GetProducts)
	group.GET("/products/featured", h.Public.GetFeaturedProducts)
	group.GET("/products/:slug", h.Public.GetProduct)
	group.GET("/products/:slug/related", h.Public.GetRelatedProducts)
	group.GET("/products/:slug/variants", h.Public.GetProductVariants)

	// Categories
	group.GET("/categories", h.Public.GetCategories)
	group.GET("/categories/:slug", h.Public.GetCategory)

	// OEM
	group.GET("/oem/flows", h.Public.GetOEMFlows)
	group.GET("/oem/flows/:id", h.Public.GetOEMFlow)
	group.GET("/oem/solutions", h.Public.GetOEMSolutions)
	group.GET("/oem/solutions/:slug", h.Public.GetOEMSolution)

	// Factory
	group.GET("/factory", h.Public.GetFactoryInfo)
	group.GET("/factory/quality-controls", h.Public.GetProcessControls)
	group.GET("/factory/timeline", h.Public.GetQualityTimeline)
	group.GET("/certifications", h.Public.GetCertifications)
	group.GET("/certifications/:id", h.Public.GetCertification)

	// Content
	group.GET("/posts", h.Public.GetPosts)
	group.GET("/posts/:slug", h.Public.GetPost)
	group.GET("/posts/:slug/related", h.Public.GetRelatedPosts)
	group.GET("/cases", h.Public.GetCases)
	group.GET("/cases/:slug", h.Public.GetCase)

	// Inquiry submission
	group.POST("/inquiry", h.Public.SubmitInquiry)

	// Search
	group.GET("/search", h.Public.Search)
}
