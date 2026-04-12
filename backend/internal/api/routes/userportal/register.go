package userportalroutes

import (
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/middleware"
	"candypro/api/internal/roles"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register wires user portal routes under /api/v1/user.
func Register(group *gin.RouterGroup, h *handlers.Handlers, cfg *config.Config, db *gorm.DB) {
	group.Use(middleware.AuthMiddleware(cfg))
	group.Use(middleware.RequireRole(roles.UserPortal()...))

	group.GET("/dashboard", h.UserPortal.CustomerGetDashboard)

	group.GET("/cart", h.UserPortal.CustomerGetCart)
	group.GET("/orders", h.UserPortal.CustomerGetOrders)
	group.GET("/orders/:id", h.UserPortal.CustomerGetOrder)
	group.GET("/orders/:id/progress", h.UserPortal.CustomerGetOrderProgress)
	group.GET("/quotes", h.UserPortal.CustomerGetQuotes)
	group.GET("/notifications", h.UserPortal.CustomerGetNotifications)
	group.PUT("/notifications/:id/read", h.UserPortal.CustomerMarkNotificationRead)
	group.POST("/notifications/mark-all-read", h.UserPortal.CustomerMarkAllNotificationsRead)
	group.PUT("/profile", h.UserPortal.CustomerUpdateProfile)
	group.POST("/change-password", h.UserPortal.CustomerChangePassword)
	group.GET("/inquiries", h.UserPortal.CustomerGetInquiries)
	group.GET("/inquiries/:id", h.UserPortal.CustomerGetInquiry)
	group.GET("/trades", h.UserPortal.CustomerListTradeTransactions)
	group.GET("/trades/:id", h.UserPortal.CustomerGetTradeTransaction)
	group.GET("/ai/stream", h.System.HandleTradeChat)
	group.POST("/ai/recommend", h.System.CustomerAIRecommendForCart)
	group.GET("/oem-projects", h.UserPortal.CustomerGetOEMProjects)
	group.GET("/oem-projects/:id", h.UserPortal.CustomerGetOEMProject)

	// Invoices (read-only for customer)
	group.GET("/invoices", h.UserPortal.CustomerGetInvoices)
	group.GET("/invoices/:id", h.UserPortal.CustomerGetInvoice)

	// Company profile
	group.GET("/company", h.UserPortal.CustomerGetCompany)
	group.PUT("/company", h.UserPortal.CustomerUpdateCompany)

	// Pricing
	group.GET("/price-list", h.UserPortal.CustomerGetMyPriceList)
	group.GET("/products/:id/price", h.UserPortal.CustomerGetProductPrice)

	// Trade documents (read-only for customer)
	group.GET("/trades/:id/documents", h.UserPortal.CustomerGetTradeDocuments)
	group.GET("/trades/:id/sales-contract", h.UserPortal.CustomerGetTradeSalesContract)
	group.GET("/trades/:id/packing-list", h.UserPortal.CustomerGetTradePackingList)
	group.GET("/trades/:id/certificate-of-origin", h.UserPortal.CustomerGetTradeCertificateOfOrigin)
	group.GET("/trades/:id/health-certificate", h.UserPortal.CustomerGetTradeHealthCertificate)
	group.GET("/trades/:id/proforma-invoice", h.UserPortal.CustomerGetTradeProformaInvoice)
	group.GET("/trades/:id/commercial-invoice", h.UserPortal.CustomerGetTradeCommercialInvoice)
	group.GET("/trades/:id/bill-of-lading", h.UserPortal.CustomerGetTradeBillOfLading)

	// Shipment tracking (read-only for customer)
	group.GET("/trades/:id/shipments", h.UserPortal.CustomerGetTradeShipments)

	// Write operations require active user status (KYB gate)
	activeGuard := group.Group("")
	activeGuard.Use(middleware.RequireActiveUser(db))
	{
		activeGuard.POST("/orders", h.UserPortal.CustomerCreateOrder)
		activeGuard.POST("/orders/ai-assist", h.System.CustomerAIAssistOrder)
		activeGuard.POST("/orders/:id/confirm", h.UserPortal.CustomerConfirmOrder)
		activeGuard.POST("/orders/:id/cancel", h.UserPortal.CustomerCancelOrder)
		activeGuard.POST("/orders/:id/payments", h.UserPortal.CustomerUploadPaymentProof)
		activeGuard.POST("/inquiries", h.UserPortal.CustomerCreateInquiry)
		activeGuard.PUT("/inquiries/:id", h.UserPortal.CustomerUpdateInquiry)
		activeGuard.POST("/trades", h.UserPortal.CustomerCreateTradeTransaction)
		activeGuard.POST("/oem-projects", h.UserPortal.CustomerCreateOEMProject)

		// Cart write operations
		activeGuard.POST("/cart/items", h.UserPortal.CustomerAddToCart)
		activeGuard.PUT("/cart/items/:itemId", h.UserPortal.CustomerUpdateCartItem)
		activeGuard.DELETE("/cart/items/:itemId", h.UserPortal.CustomerRemoveCartItem)
		activeGuard.DELETE("/cart", h.UserPortal.CustomerClearCart)
		activeGuard.POST("/cart/checkout", h.UserPortal.CustomerCheckoutCart)

		// KYB document upload
		activeGuard.POST("/company/kyb-document", h.UserPortal.CustomerUploadKYBDocument)
	}
}
