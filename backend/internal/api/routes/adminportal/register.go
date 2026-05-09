package adminportalroutes

import (
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/middleware"
	"candypro/api/internal/roles"

	"github.com/gin-gonic/gin"
)

// Register wires admin portal routes under /api/v1/admin.
func Register(group *gin.RouterGroup, h *handlers.Handlers, cfg *config.Config) {
	group.Use(middleware.AuthMiddleware(cfg))
	group.Use(middleware.RequireRole(roles.AdminPortal()...))

	// Dashboard
	group.GET("/dashboard", h.AdminPortal.GetDashboard)
	group.GET("/dashboard/stats", h.AdminPortal.GetDashboardStats)
	group.GET("/reports/sales", h.AdminPortal.GetSalesReport)
	group.GET("/reports/inquiries", h.AdminPortal.GetInquiriesReport)

	// Users
	group.GET("/users", h.AdminPortal.AdminGetUsers)
	group.GET("/users/:id", h.AdminPortal.AdminGetUser)
	group.POST("/users", h.AdminPortal.AdminCreateUser)
	group.PUT("/users/:id", h.AdminPortal.AdminUpdateUser)
	group.DELETE("/users/:id", h.AdminPortal.AdminDeleteUser)
	group.PUT("/users/:id/status", h.AdminPortal.AdminUpdateUserStatus)
	group.PUT("/users/:id/role", h.AdminPortal.AdminUpdateUserRole)

	// Orders
	group.GET("/orders", h.AdminPortal.AdminGetOrders)
	group.POST("/orders", h.AdminPortal.AdminCreateOrder)
	group.GET("/orders/:id", h.AdminPortal.AdminGetOrder)
	group.PUT("/orders/:id", h.AdminPortal.AdminUpdateOrder)
	group.DELETE("/orders/:id", h.AdminPortal.AdminDeleteOrder)
	group.PUT("/orders/:id/status", h.AdminPortal.AdminUpdateOrderStatus)
	group.POST("/orders/:id/create-trade", h.AdminPortal.AdminCreateTradeFromOrder)

	// Order Payments (admin)
	group.GET("/orders/:id/payments", h.AdminPortal.AdminGetOrderPayments)
	group.POST("/orders/:id/payments", h.AdminPortal.AdminCreatePayment)
	group.GET("/orders/:id/payments/:paymentId/file", h.AdminPortal.AdminDownloadPaymentProofFile)
	group.PUT("/orders/:id/payments/:paymentId/confirm", h.AdminPortal.AdminConfirmPayment)
	group.PUT("/orders/:id/payments/:paymentId/refund", h.AdminPortal.AdminRefundPayment)

	// Inquiries
	group.GET("/inquiries", h.AdminPortal.AdminGetInquiries)
	group.POST("/inquiries", h.AdminPortal.AdminCreateInquiry)
	group.GET("/inquiries/:id", h.AdminPortal.AdminGetInquiry)
	group.PUT("/inquiries/:id", h.AdminPortal.AdminUpdateInquiry)
	group.DELETE("/inquiries/:id", h.AdminPortal.AdminDeleteInquiry)
	group.PUT("/inquiries/:id/status", h.AdminPortal.AdminUpdateInquiryStatus)
	group.PUT("/inquiries/:id/assign", h.AdminPortal.AdminAssignInquiry)
	group.POST("/inquiries/:id/analyze", h.AdminPortal.AdminAnalyzeInquiry)
	group.POST("/inquiries/:id/quote", h.AdminPortal.AdminQuoteInquiry)
	group.POST("/inquiries/:id/convert-to-order", h.AdminPortal.AdminConvertInquiryToOrder)

	// Products
	group.GET("/products", h.AdminPortal.AdminGetProducts)
	group.POST("/products", h.AdminPortal.AdminCreateProduct)
	group.GET("/products/:id", h.AdminPortal.AdminGetProduct)
	group.PUT("/products/:id", h.AdminPortal.AdminUpdateProduct)
	group.DELETE("/products/:id", h.AdminPortal.AdminDeleteProduct)
	group.PUT("/products/:id/status", h.AdminPortal.AdminUpdateProductStatus)

	// Content
	group.GET("/content", h.AdminPortal.AdminGetContent)
	group.POST("/content/ai-generate", h.AdminPortal.AdminAIGenerateContent)
	group.GET("/content/:id", h.AdminPortal.AdminGetContentByID)
	group.POST("/content", h.AdminPortal.AdminCreateContent)
	group.PUT("/content/:id", h.AdminPortal.AdminUpdateContent)
	group.DELETE("/content/:id", h.AdminPortal.AdminDeleteContent)

	// Certifications
	group.GET("/certifications", h.AdminPortal.AdminGetCertifications)
	group.GET("/certifications/:id", h.AdminPortal.AdminGetCertification)
	group.POST("/certifications", h.AdminPortal.AdminCreateCertification)
	group.PUT("/certifications/:id", h.AdminPortal.AdminUpdateCertification)
	group.DELETE("/certifications/:id", h.AdminPortal.AdminDeleteCertification)

	// Uploads
	group.POST("/upload/image", h.AdminPortal.AdminUploadImage)
	group.POST("/upload/document", h.AdminPortal.AdminUploadDocument)
	group.POST("/upload/multiple", h.AdminPortal.AdminUploadMultiple)
	group.DELETE("/upload/:filename", h.AdminPortal.AdminDeleteFile)

	// Companies
	group.GET("/companies", h.AdminPortal.AdminGetCompanies)
	group.GET("/companies/:id", h.AdminPortal.AdminGetCompany)
	group.PUT("/companies/:id", h.AdminPortal.AdminUpdateCompany)
	group.PUT("/companies/:id/verify", h.AdminPortal.AdminVerifyCompany)

	// Trade
	group.GET("/trades", h.AdminPortal.AdminGetTradeTransactions)
	group.GET("/trades/:id", h.AdminPortal.AdminGetTradeTransaction)
	group.PUT("/trades/:id/status", h.AdminPortal.AdminUpdateTradeStatus)
	group.GET("/trades/:id/documents", h.AdminPortal.AdminGetTradeDocuments)
	group.POST("/trades/:id/documents", h.AdminPortal.AdminCreateTradeDocument)
	group.PUT("/trades/:id/documents/:docId", h.AdminPortal.AdminUpdateTradeDocument)
	group.DELETE("/trades/:id/documents/:docId", h.AdminPortal.AdminDeleteTradeDocument)

	// Trade rich documents
	group.GET("/trades/:id/sales-contract", h.AdminPortal.AdminGetSalesContract)
	group.POST("/trades/:id/sales-contract", h.AdminPortal.AdminCreateSalesContract)
	group.PUT("/trades/:id/sales-contract", h.AdminPortal.AdminUpdateSalesContract)
	group.GET("/trades/:id/packing-list", h.AdminPortal.AdminGetPackingList)
	group.POST("/trades/:id/packing-list", h.AdminPortal.AdminCreatePackingList)
	group.PUT("/trades/:id/packing-list", h.AdminPortal.AdminUpdatePackingList)
	group.GET("/trades/:id/certificate-of-origin", h.AdminPortal.AdminGetCertificateOfOrigin)
	group.POST("/trades/:id/certificate-of-origin", h.AdminPortal.AdminCreateCertificateOfOrigin)
	group.PUT("/trades/:id/certificate-of-origin", h.AdminPortal.AdminUpdateCertificateOfOrigin)
	group.GET("/trades/:id/health-certificate", h.AdminPortal.AdminGetHealthCertificate)
	group.POST("/trades/:id/health-certificate", h.AdminPortal.AdminCreateHealthCertificate)
	group.PUT("/trades/:id/health-certificate", h.AdminPortal.AdminUpdateHealthCertificate)
	group.GET("/trades/:id/settlements", h.AdminPortal.AdminGetSettlements)
	group.POST("/trades/:id/settlements", h.AdminPortal.AdminCreateSettlement)
	group.PUT("/trades/:id/settlements/:settlementId", h.AdminPortal.AdminUpdateSettlement)
	group.DELETE("/trades/:id/settlements/:settlementId", h.AdminPortal.AdminDeleteSettlement)
	group.GET("/trades/:id/compliance", h.AdminPortal.AdminGetCompliance)
	group.PUT("/trades/:id/compliance/:compId", h.AdminPortal.AdminUpdateCompliance)

	// Trade PI/CI/B/L
	group.GET("/trades/:id/proforma-invoice", h.AdminPortal.AdminGetProformaInvoice)
	group.POST("/trades/:id/proforma-invoice", h.AdminPortal.AdminCreateProformaInvoice)
	group.PUT("/trades/:id/proforma-invoice", h.AdminPortal.AdminUpdateProformaInvoice)
	group.GET("/trades/:id/commercial-invoice", h.AdminPortal.AdminGetCommercialInvoice)
	group.POST("/trades/:id/commercial-invoice", h.AdminPortal.AdminCreateCommercialInvoice)
	group.PUT("/trades/:id/commercial-invoice", h.AdminPortal.AdminUpdateCommercialInvoice)
	group.GET("/trades/:id/bill-of-lading", h.AdminPortal.AdminGetBillOfLading)
	group.POST("/trades/:id/bill-of-lading", h.AdminPortal.AdminCreateBillOfLading)
	group.PUT("/trades/:id/bill-of-lading", h.AdminPortal.AdminUpdateBillOfLading)

	// Shipments
	group.GET("/shipments", h.AdminPortal.AdminGetShipments)
	group.POST("/shipments", h.AdminPortal.AdminCreateShipment)
	group.GET("/shipments/:id", h.AdminPortal.AdminGetShipment)
	group.PUT("/shipments/:id", h.AdminPortal.AdminUpdateShipment)
	group.DELETE("/shipments/:id", h.AdminPortal.AdminDeleteShipment)
	// Shipment logistics orchestration
	group.GET("/shipments/:id/timeline", h.AdminPortal.AdminGetShipmentTimeline)
	group.POST("/shipments/:id/events", h.AdminPortal.AdminAddTrackingEvent)
	group.POST("/shipments/:id/dispatch", h.AdminPortal.AdminDispatchShipment)
	group.POST("/shipments/:id/confirm-delivery", h.AdminPortal.AdminConfirmDelivery)

	// Invoices（from-order 须在 :id 之前注册）
	group.GET("/invoices", h.AdminPortal.AdminGetInvoices)
	group.POST("/invoices/from-order", h.AdminPortal.AdminCreateInvoiceFromOrder)
	group.POST("/invoices", h.AdminPortal.AdminCreateInvoice)
	group.GET("/invoices/stats", h.AdminPortal.AdminGetInvoiceStats)
	group.GET("/invoices/:id", h.AdminPortal.AdminGetInvoice)
	group.PUT("/invoices/:id", h.AdminPortal.AdminUpdateInvoice)
	group.POST("/invoices/:id/send", h.AdminPortal.AdminSendInvoice)
	group.DELETE("/invoices/:id", h.AdminPortal.AdminDeleteInvoice)

	// 跨境：仓库、市场画像、成本栈、OEM 预留、合规 Copilot
	group.GET("/warehouses", h.AdminPortal.AdminListWarehouses)
	group.POST("/warehouses", h.AdminPortal.AdminUpsertWarehouse)
	group.PUT("/warehouses/:id/stock", h.AdminPortal.AdminUpsertWarehouseStock)
	group.GET("/products/:id/market-profile", h.AdminPortal.AdminGetProductMarketProfile)
	group.PUT("/products/:id/market-profile", h.AdminPortal.AdminPutProductMarketProfile)
	group.GET("/products/:id/market-costs", h.AdminPortal.AdminGetProductMarketCosts)
	group.PUT("/products/:id/market-costs", h.AdminPortal.AdminPutProductMarketCost)
	group.GET("/products/:id/channel-inventory", h.AdminPortal.AdminListProductChannelInventory)
	group.PUT("/products/:id/channel-inventory", h.AdminPortal.AdminPutProductChannelInventory)
	group.POST("/oem-projects/:id/inventory-holds", h.AdminPortal.AdminCreateOEMInventoryHold)
	group.GET("/oem-projects/:id/inventory-holds", h.AdminPortal.AdminListOEMInventoryHolds)
	group.POST("/products/:id/compliance-suggest", h.AdminPortal.AdminProductComplianceSuggest)

	// Pricing
	group.GET("/price-lists", h.AdminPortal.AdminGetPriceLists)
	group.POST("/price-lists", h.AdminPortal.AdminCreatePriceList)
	group.PUT("/price-lists/:id", h.AdminPortal.AdminUpdatePriceList)
	group.DELETE("/price-lists/:id", h.AdminPortal.AdminDeletePriceList)
	group.GET("/products/:id/prices", h.AdminPortal.AdminGetProductPrices)
	group.POST("/products/:id/prices", h.AdminPortal.AdminSetProductPrice)
	group.DELETE("/products/:id/prices/:priceId", h.AdminPortal.AdminDeleteProductPrice)

	// OEM Projects
	group.GET("/oem-projects", h.AdminPortal.AdminGetOEMProjects)
	group.GET("/oem-projects/:id", h.AdminPortal.AdminGetOEMProject)
	group.PUT("/oem-projects/:id", h.AdminPortal.AdminUpdateOEMProject)
	group.PUT("/oem-projects/:id/status", h.AdminPortal.AdminUpdateOEMStatus)

	// Reports
	group.GET("/reports/revenue-trends", h.AdminPortal.GetRevenueTrends)
	group.GET("/reports/order-trends", h.AdminPortal.GetOrderTrends)
	group.GET("/reports/top-products", h.AdminPortal.GetTopProducts)
	group.GET("/reports/conversion-trends", h.AdminPortal.GetConversionTrends)

	// Financial
	group.GET("/financial/overview", h.AdminPortal.GetFinancialOverview)
	group.GET("/financial/outstanding-invoices", h.AdminPortal.GetOutstandingInvoices)
	group.GET("/financial/payment-breakdown", h.AdminPortal.GetPaymentBreakdown)

	// Staff & Audit
	group.GET("/staff", h.AdminPortal.GetStaffList)
	group.GET("/staff/:id/activity", h.AdminPortal.GetStaffActivity)
	group.GET("/audit-log", h.AdminPortal.GetAuditLog)

	// Settings
	group.GET("/settings", h.AdminPortal.GetSettings)
	group.PUT("/settings/:key", h.AdminPortal.UpdateSetting)
}
