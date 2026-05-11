package userportalroutes

import (
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/middleware"
	modelsAuth "candypro/api/internal/models/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register wires user portal routes under /api/v1/user.
func Register(group *gin.RouterGroup, h *handlers.Handlers, cfg *config.Config, db *gorm.DB) {
	group.Use(middleware.AuthMiddleware(cfg))
	group.Use(middleware.RequireRole(modelsAuth.UserPortal()...))

	group.GET("/dashboard", h.UserPortal.CustomerGetDashboard)

	group.GET("/cart", h.UserPortal.CustomerGetCart)
	group.GET("/orders", h.UserPortal.CustomerGetOrders)
	group.GET("/orders/:id", h.UserPortal.CustomerGetOrder)
	group.GET("/orders/:id/payments", h.UserPortal.CustomerGetOrderPayments)
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

	// 下单/购物车写入：在 handler 内校验 active 或 KYB_BYPASS_MAX_ORDER_USD 小额免审
	group.POST("/orders", h.UserPortal.CustomerCreateOrder)
	group.POST("/orders/ai-assist", h.System.CustomerAIAssistOrder)
	group.POST("/orders/:id/confirm", h.UserPortal.CustomerConfirmOrder)
	group.POST("/orders/:id/cancel", h.UserPortal.CustomerCancelOrder)
	group.POST("/orders/:id/nudge", h.UserPortal.CustomerNudgeOrder)
	group.GET("/orders/:id/messages", h.UserPortal.CustomerGetOrderMessages)
	group.POST("/orders/:id/messages", h.UserPortal.CustomerSendOrderMessage)
	group.GET("/orders/:id/payments/:paymentId/file", h.UserPortal.CustomerDownloadPaymentProofFile)
	group.POST("/cart/items", h.UserPortal.CustomerAddToCart)
	group.PUT("/cart/items/:itemId", h.UserPortal.CustomerUpdateCartItem)
	group.DELETE("/cart/items/:itemId", h.UserPortal.CustomerRemoveCartItem)
	group.DELETE("/cart", h.UserPortal.CustomerClearCart)
	group.POST("/cart/checkout", h.UserPortal.CustomerCheckoutCart)

	// Invoices (read-only for customer)
	group.GET("/invoices", h.UserPortal.CustomerGetInvoices)
	group.GET("/invoices/:id", h.UserPortal.CustomerGetInvoice)

	// Company profile
	group.GET("/company", h.UserPortal.CustomerGetCompany)
	group.PUT("/company", h.UserPortal.CustomerUpdateCompany)
	group.GET("/company/kyb-document-file", h.UserPortal.CustomerDownloadKYBDocumentFile)

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
	group.GET("/trades/:id/shipments/:shipmentId/timeline", h.UserPortal.CustomerGetShipmentTimeline)
	group.GET("/trades/:id/timeline", h.UserPortal.CustomerGetTradeTimeline)

	// Write operations require active user status (KYB gate)
	activeGuard := group.Group("")
	activeGuard.Use(middleware.RequireActiveUser(db))
	{
		activeGuard.POST("/orders/:id/payments", h.UserPortal.CustomerUploadPaymentProof)
		activeGuard.POST("/inquiries", h.UserPortal.CustomerCreateInquiry)
		activeGuard.PUT("/inquiries/:id", h.UserPortal.CustomerUpdateInquiry)
		activeGuard.POST("/trades", h.UserPortal.CustomerCreateTradeTransaction)
		activeGuard.POST("/oem-projects", h.UserPortal.CustomerCreateOEMProject)

		// KYB document upload
		activeGuard.POST("/company/kyb-document", h.UserPortal.CustomerUploadKYBDocument)
	}
}
