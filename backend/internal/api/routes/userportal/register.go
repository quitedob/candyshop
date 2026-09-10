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
func Register(group *gin.RouterGroup, h *handlers.Handlers, cfg *config.Config, db *gorm.DB, afterAuthorization ...gin.HandlerFunc) {
	group.Use(middleware.AuthMiddleware(cfg, h.AuthScope.SessionStore()))
	group.Use(middleware.RequireRole(modelsAuth.UserPortal()...))
	userRoutes := group.Group("")
	userRoutes.Use(afterAuthorization...)

	userRoutes.GET("/dashboard", h.UserPortal.CustomerGetDashboard)

	userRoutes.GET("/cart", h.UserPortal.CustomerGetCart)
	userRoutes.GET("/orders", h.UserPortal.CustomerGetOrders)
	userRoutes.GET("/orders/:id", h.UserPortal.CustomerGetOrder)
	userRoutes.GET("/orders/:id/payments", h.UserPortal.CustomerGetOrderPayments)
	userRoutes.GET("/orders/:id/progress", h.UserPortal.CustomerGetOrderProgress)
	userRoutes.GET("/quotes", h.UserPortal.CustomerGetQuotes)
	userRoutes.GET("/notifications", h.UserPortal.CustomerGetNotifications)
	userRoutes.PUT("/notifications/:id/read", h.UserPortal.CustomerMarkNotificationRead)
	userRoutes.POST("/notifications/mark-all-read", h.UserPortal.CustomerMarkAllNotificationsRead)
	userRoutes.PUT("/profile", h.UserPortal.CustomerUpdateProfile)
	userRoutes.POST("/change-password", h.UserPortal.CustomerChangePassword)
	userRoutes.GET("/inquiries", h.UserPortal.CustomerGetInquiries)
	userRoutes.GET("/inquiries/:id", h.UserPortal.CustomerGetInquiry)
	userRoutes.GET("/trades", h.UserPortal.CustomerListTradeTransactions)
	userRoutes.GET("/trades/:id", h.UserPortal.CustomerGetTradeTransaction)
	userRoutes.GET("/ai/stream", h.System.HandleTradeChat)
	userRoutes.GET("/ai/b2b-coordinator", h.System.HandleB2BCoordinatorChat)
	userRoutes.GET("/ai/order-processing", h.System.HandleOrderProcessingChat)
	userRoutes.POST("/ai/recommend", h.System.CustomerAIRecommendForCart)
	userRoutes.GET("/oem-projects", h.UserPortal.CustomerGetOEMProjects)
	userRoutes.GET("/oem-projects/:id", h.UserPortal.CustomerGetOEMProject)

	// 下单/购物车写入：在 handler 内校验 active 或 KYB_BYPASS_MAX_ORDER_USD 小额免审
	userRoutes.POST("/orders", h.UserPortal.CustomerCreateOrder)
	userRoutes.POST("/orders/ai-assist", h.System.CustomerAIAssistOrder)
	userRoutes.POST("/orders/:id/confirm", h.UserPortal.CustomerConfirmOrder)
	userRoutes.POST("/orders/:id/cancel", h.UserPortal.CustomerCancelOrder)
	userRoutes.POST("/orders/:id/approve", h.UserPortal.CustomerApproveOrder)
	userRoutes.POST("/orders/:id/reject", h.UserPortal.CustomerRejectOrder)
	userRoutes.POST("/orders/:id/nudge", h.UserPortal.CustomerNudgeOrder)
	userRoutes.GET("/orders/messages/unread-count", h.UserPortal.CustomerGetOrderMessageUnreadTotal)
	userRoutes.GET("/orders/:id/messages", h.UserPortal.CustomerGetOrderMessages)
	userRoutes.GET("/orders/:id/messages/unread-count", h.UserPortal.CustomerGetOrderMessageUnreadCount)
	userRoutes.POST("/orders/:id/messages/read", h.UserPortal.CustomerMarkOrderMessagesRead)
	userRoutes.POST("/orders/:id/messages", h.UserPortal.CustomerSendOrderMessage)
	userRoutes.GET("/orders/:id/messages/ws", h.UserPortal.CustomerStreamOrderMessages)
	// Returns
	userRoutes.POST("/orders/:id/returns", h.UserPortal.CustomerCreateReturn)
	userRoutes.GET("/returns", h.UserPortal.CustomerListReturns)
	userRoutes.GET("/returns/:id", h.UserPortal.CustomerGetReturn)
	userRoutes.GET("/orders/:id/payments/:paymentId/file", h.UserPortal.CustomerDownloadPaymentProofFile)
	// Coupons
	userRoutes.POST("/cart/coupon/validate", h.UserPortal.CustomerValidateCartCoupon)
	userRoutes.POST("/cart/coupon", h.UserPortal.CustomerApplyCoupon)
	userRoutes.DELETE("/cart/coupon", h.UserPortal.CustomerRemoveCoupon)

	userRoutes.POST("/cart/items", h.UserPortal.CustomerAddToCart)
	userRoutes.PUT("/cart/items/:itemId", h.UserPortal.CustomerUpdateCartItem)
	userRoutes.DELETE("/cart/items/:itemId", h.UserPortal.CustomerRemoveCartItem)
	userRoutes.DELETE("/cart", h.UserPortal.CustomerClearCart)
	userRoutes.POST("/cart/checkout", h.UserPortal.CustomerCheckoutCart)

	// Invoices (read-only for customer)
	userRoutes.GET("/invoices", h.UserPortal.CustomerGetInvoices)
	userRoutes.GET("/invoices/:id", h.UserPortal.CustomerGetInvoice)

	// Company profile
	userRoutes.GET("/company", h.UserPortal.CustomerGetCompany)
	userRoutes.PUT("/company", h.UserPortal.CustomerUpdateCompany)
	userRoutes.GET("/company/kyb-document-file", h.UserPortal.CustomerDownloadKYBDocumentFile)
	userRoutes.POST("/company/kyb-document", h.UserPortal.CustomerUploadKYBDocument)
	userRoutes.GET("/uploads/:fileId", h.UserPortal.CustomerGetUploadedFile)

	// Pricing
	userRoutes.GET("/price-list", h.UserPortal.CustomerGetMyPriceList)
	userRoutes.GET("/products/:id/price", h.UserPortal.CustomerGetProductPrice)

	// Shipping rates
	userRoutes.GET("/shipping-rates", h.UserPortal.CustomerGetShippingRates)
	userRoutes.GET("/shipping-estimate", h.UserPortal.CustomerGetShippingEstimate)

	// Trade documents (read-only for customer)
	userRoutes.GET("/trades/:id/documents", h.UserPortal.CustomerGetTradeDocuments)
	userRoutes.GET("/trades/:id/sales-contract", h.UserPortal.CustomerGetTradeSalesContract)
	userRoutes.GET("/trades/:id/packing-list", h.UserPortal.CustomerGetTradePackingList)
	userRoutes.GET("/trades/:id/certificate-of-origin", h.UserPortal.CustomerGetTradeCertificateOfOrigin)
	userRoutes.GET("/trades/:id/health-certificate", h.UserPortal.CustomerGetTradeHealthCertificate)
	userRoutes.GET("/trades/:id/proforma-invoice", h.UserPortal.CustomerGetTradeProformaInvoice)
	userRoutes.GET("/trades/:id/commercial-invoice", h.UserPortal.CustomerGetTradeCommercialInvoice)
	userRoutes.GET("/trades/:id/bill-of-lading", h.UserPortal.CustomerGetTradeBillOfLading)

	// Shipment tracking (read-only for customer)
	userRoutes.GET("/trades/:id/shipments", h.UserPortal.CustomerGetTradeShipments)
	userRoutes.GET("/trades/:id/shipments/:shipmentId/timeline", h.UserPortal.CustomerGetShipmentTimeline)
	userRoutes.GET("/trades/:id/timeline", h.UserPortal.CustomerGetTradeTimeline)

	// Requisition lists (read)
	userRoutes.GET("/requisition-lists", h.UserPortal.CustomerListRequisitionLists)
	userRoutes.GET("/requisition-lists/:id", h.UserPortal.CustomerGetRequisitionList)

	// Write operations require active user status (KYB gate)
	activeGuard := group.Group("")
	activeGuard.Use(middleware.RequireActiveUser(db))
	activeGuard.Use(afterAuthorization...)
	{
		activeGuard.POST("/orders/:id/payments", h.UserPortal.CustomerUploadPaymentProof)
		activeGuard.POST("/orders/:id/payments/gateway", h.UserPortal.CustomerCreateGatewayPayment)
		activeGuard.POST("/inquiries", h.UserPortal.CustomerCreateInquiry)
		activeGuard.PUT("/inquiries/:id", h.UserPortal.CustomerUpdateInquiry)
		activeGuard.POST("/inquiries/:id/attachments", h.UserPortal.CustomerUploadInquiryAttachment)
		activeGuard.POST("/inquiries/:id/confirm", h.UserPortal.CustomerConfirmInquiry)
		activeGuard.GET("/inquiries/:id/negotiations", h.UserPortal.CustomerGetNegotiationOffers)
		activeGuard.POST("/inquiries/:id/negotiations", h.UserPortal.CustomerCreateNegotiationOffer)
		activeGuard.POST("/inquiries/:id/negotiations/:offerId/accept", h.UserPortal.CustomerAcceptNegotiationOffer)
		activeGuard.POST("/inquiries/:id/negotiations/:offerId/reject", h.UserPortal.CustomerRejectNegotiationOffer)
		activeGuard.POST("/trades", h.UserPortal.CustomerCreateTradeTransaction)
		activeGuard.POST("/oem-projects", h.UserPortal.CustomerCreateOEMProject)
		activeGuard.POST("/oem-projects/:id/attachments", h.UserPortal.CustomerUploadOEMProjectAttachment)
		activeGuard.POST("/trades/:id/shipments/:shipmentId/nudge", h.UserPortal.CustomerNudgeShipment)
		activeGuard.POST("/trades/:id/shipments/:shipmentId/attachments", h.UserPortal.CustomerUploadShipmentAttachment)

		// Bulk orders & requisition lists
		activeGuard.POST("/orders/bulk", h.UserPortal.CustomerCreateBulkOrder)
		activeGuard.POST("/orders/:id/reorder", h.UserPortal.CustomerReorderFromHistory)
		activeGuard.POST("/requisition-lists", h.UserPortal.CustomerCreateRequisitionList)
		activeGuard.PUT("/requisition-lists/:id", h.UserPortal.CustomerUpdateRequisitionList)
		activeGuard.DELETE("/requisition-lists/:id", h.UserPortal.CustomerDeleteRequisitionList)
		activeGuard.POST("/requisition-lists/:id/convert", h.UserPortal.CustomerConvertRequisitionToOrder)
		activeGuard.POST("/uploads/presign", h.UserPortal.CustomerPresignUpload)
		activeGuard.POST("/uploads/:fileId/complete", h.UserPortal.CustomerCompletePresignedUpload)
	}
}
