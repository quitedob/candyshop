package adminportalroutes

import (
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/middleware"
	modelsAuth "candypro/api/internal/models/auth"

	"github.com/gin-gonic/gin"
)

// Register wires admin portal routes under /api/v1/admin.
func Register(group *gin.RouterGroup, h *handlers.Handlers, cfg *config.Config) {
	group.Use(middleware.AuthMiddleware(cfg))
	group.Use(middleware.RequireRole(modelsAuth.AdminPortal()...))

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
		// Order Approval
		group.POST("/orders/:id/approve", h.AdminPortal.AdminApproveOrder)
		group.POST("/orders/:id/reject", h.AdminPortal.AdminRejectOrder)
		group.GET("/orders/:id/approval-history", h.AdminPortal.AdminGetApprovalHistory)

	group.POST("/orders/:id/create-trade", h.AdminPortal.AdminCreateTradeFromOrder)
		group.POST("/orders/:id/fulfillments", h.AdminPortal.AdminCreateFulfillment)
		group.GET("/orders/:id/fulfillments", h.AdminPortal.AdminListFulfillments)
		group.PUT("/fulfillments/:id/ship", h.AdminPortal.AdminShipFulfillment)
		group.PUT("/fulfillments/:id/deliver", h.AdminPortal.AdminDeliverFulfillment)
		group.PUT("/fulfillments/:id/cancel", h.AdminPortal.AdminCancelFulfillment)
		// Returns (admin)
		group.GET("/returns", h.AdminPortal.AdminListReturns)
		group.GET("/returns/:id", h.AdminPortal.AdminGetReturn)
		group.PUT("/returns/:id/approve", h.AdminPortal.AdminApproveReturn)
		group.PUT("/returns/:id/receive", h.AdminPortal.AdminReceiveReturn)
		group.PUT("/returns/:id/refund", h.AdminPortal.AdminRefundReturn)
		group.PUT("/returns/:id/reject", h.AdminPortal.AdminRejectReturn)

	// Order Payments (admin)
	group.GET("/orders/:id/payments", h.AdminPortal.AdminGetOrderPayments)
	group.POST("/orders/:id/payments", h.AdminPortal.AdminCreatePayment)
	group.GET("/orders/:id/payments/:paymentId/file", h.AdminPortal.AdminDownloadPaymentProofFile)
	group.PUT("/orders/:id/payments/:paymentId/confirm", h.AdminPortal.AdminConfirmPayment)
	group.PUT("/orders/:id/payments/:paymentId/refund", h.AdminPortal.AdminRefundPayment)
	group.GET("/orders/:id/messages", h.AdminPortal.AdminGetOrderMessages)
	group.POST("/orders/:id/messages", h.AdminPortal.AdminSendOrderMessage)

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
	group.PUT("/inquiries/:id/confirm", h.AdminPortal.AdminConfirmInquiry)

	// Negotiation
	group.GET("/inquiries/:id/negotiations", h.AdminPortal.AdminGetNegotiationOffers)
	group.POST("/inquiries/:id/negotiations", h.AdminPortal.AdminCreateNegotiationOffer)
	group.POST("/inquiries/:id/negotiations/:offerId/accept", h.AdminPortal.AdminAcceptNegotiationOffer)
	group.POST("/inquiries/:id/negotiations/:offerId/reject", h.AdminPortal.AdminRejectNegotiationOffer)

	// Products
	group.GET("/products", h.AdminPortal.AdminGetProducts)
	group.POST("/products", h.AdminPortal.AdminCreateProduct)
	group.GET("/products/:id", h.AdminPortal.AdminGetProduct)
	group.PUT("/products/:id", h.AdminPortal.AdminUpdateProduct)
	group.DELETE("/products/:id", h.AdminPortal.AdminDeleteProduct)
	group.PUT("/products/:id/status", h.AdminPortal.AdminUpdateProductStatus)
	group.POST("/products/:id/ai-translate", h.AdminPortal.AdminAITranslateProduct)

	// Content
	group.GET("/content", h.AdminPortal.AdminGetContent)
	group.POST("/content/ai-generate", h.AdminPortal.AdminAIGenerateContent)
	group.GET("/content/:id", h.AdminPortal.AdminGetContentByID)
	group.POST("/content", h.AdminPortal.AdminCreateContent)
	group.PUT("/content/:id", h.AdminPortal.AdminUpdateContent)
	group.DELETE("/content/:id", h.AdminPortal.AdminDeleteContent)
	group.POST("/content/:id/ai-translate", h.AdminPortal.AdminAITranslateContent)

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
		// Buyer Organizations (B2B Approval)
		group.GET("/organizations", h.AdminPortal.AdminGetOrganizations)
		group.POST("/organizations", h.AdminPortal.AdminCreateOrganization)
		group.PUT("/organizations/:id", h.AdminPortal.AdminUpdateOrganization)
		group.GET("/organizations/members", h.AdminPortal.AdminGetOrgMembers)
		group.POST("/organizations/members", h.AdminPortal.AdminAddOrgMember)
		group.DELETE("/organizations/members/:id", h.AdminPortal.AdminRemoveOrgMember)


	// Trade
	group.GET("/trades", h.AdminPortal.AdminGetTradeTransactions)
	group.GET("/trades/:id", h.AdminPortal.AdminGetTradeTransaction)
	group.PUT("/trades/:id/status", h.AdminPortal.AdminUpdateTradeStatus)
	group.GET("/trades/:id/export/contract", h.AdminPortal.AdminExportTradeContract)
	group.GET("/trades/:id/export/sales-contract", h.AdminPortal.AdminExportTradeSalesContract)
	group.GET("/trades/:id/documents", h.AdminPortal.AdminGetTradeDocuments)
	group.POST("/trades/:id/documents", h.AdminPortal.AdminCreateTradeDocument)
	group.POST("/trades/:id/documents/ai-generate", h.AdminPortal.AdminAIGenerateTradeDocument)
	group.GET("/trades/:id/ai-chat", h.AdminPortal.AdminAITradeChat)
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

	// Shipping Rates
	group.GET("/shipping-rates", h.AdminPortal.AdminGetShippingRates)
	group.POST("/shipping-rates", h.AdminPortal.AdminCreateShippingRate)
	group.PUT("/shipping-rates/:id", h.AdminPortal.AdminUpdateShippingRate)
	group.DELETE("/shipping-rates/:id", h.AdminPortal.AdminDeleteShippingRate)

	// Invoices（from-order 须在 :id 之前注册）
	group.GET("/invoices", h.AdminPortal.AdminGetInvoices)
	group.POST("/invoices/from-order", h.AdminPortal.AdminCreateInvoiceFromOrder)
	group.POST("/invoices", h.AdminPortal.AdminCreateInvoice)
	group.GET("/invoices/stats", h.AdminPortal.AdminGetInvoiceStats)
	group.GET("/invoices/:id", h.AdminPortal.AdminGetInvoice)
	group.PUT("/invoices/:id", h.AdminPortal.AdminUpdateInvoice)
	group.POST("/invoices/:id/send", h.AdminPortal.AdminSendInvoice)
	group.GET("/invoices/:id/export", h.AdminPortal.AdminExportInvoice)
	group.GET("/invoices/:id/export/proforma-invoice", h.AdminPortal.AdminExportInvoiceProforma)
	group.DELETE("/invoices/:id", h.AdminPortal.AdminDeleteInvoice)

	// Inventory
	group.GET("/inventory", h.AdminPortal.AdminGetInventory)
	group.PUT("/inventory/:productId", h.AdminPortal.AdminUpdateInventory)
	group.GET("/inventory/:productId/history", h.AdminPortal.AdminGetInventoryHistory)
	group.POST("/inventory/export-xlsx", h.AdminPortal.AdminExportInventoryXLSX)
	group.POST("/inventory/import-xlsx", h.AdminPortal.AdminImportInventoryXLSX)
	group.POST("/inventory/import-xlsx/apply", h.AdminPortal.AdminApplyInventoryImport)
	group.POST("/inventory/batch-update", h.AdminPortal.AdminBatchUpdateInventory)
	group.POST("/inventory/batch-delete", h.AdminPortal.AdminBatchDeleteInventory)
		group.POST("/inventory/transfer", h.AdminPortal.AdminCreateStockTransfer)
		group.GET("/inventory/transfers", h.AdminPortal.AdminListStockTransfers)
		group.GET("/inventory/transfers/:id", h.AdminPortal.AdminGetStockTransfer)

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

			// Coupons & Gift Cards
		group.POST("/coupons", h.AdminPortal.AdminCreateCoupon)
		group.GET("/coupons", h.AdminPortal.AdminListCoupons)
		group.DELETE("/coupons/:id", h.AdminPortal.AdminDeleteCoupon)
		group.POST("/gift-cards", h.AdminPortal.AdminCreateGiftCard)
		group.GET("/gift-cards", h.AdminPortal.AdminListGiftCards)

			// Suppliers & Purchase Orders
		group.GET("/suppliers", h.AdminPortal.AdminListSuppliers)
		group.POST("/suppliers", h.AdminPortal.AdminCreateSupplier)
		group.GET("/suppliers/:id", h.AdminPortal.AdminGetSupplier)
		group.PUT("/suppliers/:id", h.AdminPortal.AdminUpdateSupplier)
		group.DELETE("/suppliers/:id", h.AdminPortal.AdminDeleteSupplier)
		group.GET("/purchase-orders", h.AdminPortal.AdminListPOs)
		group.POST("/purchase-orders", h.AdminPortal.AdminCreatePO)
		group.GET("/purchase-orders/:id", h.AdminPortal.AdminGetPO)
		group.PUT("/purchase-orders/:id/receive", h.AdminPortal.AdminReceivePO)

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

		// OEM Flows
		group.GET("/oem-flows", h.AdminPortal.AdminGetOEMFlows)
		group.GET("/oem-flows/:id", h.AdminPortal.AdminGetOEMFlow)
		group.POST("/oem-flows", h.AdminPortal.AdminCreateOEMFlow)
		group.PUT("/oem-flows/:id", h.AdminPortal.AdminUpdateOEMFlow)
		group.DELETE("/oem-flows/:id", h.AdminPortal.AdminDeleteOEMFlow)

		// OEM Solutions
		group.GET("/oem-solutions", h.AdminPortal.AdminGetOEMSolutions)
		group.GET("/oem-solutions/:id", h.AdminPortal.AdminGetOEMSolution)
		group.POST("/oem-solutions", h.AdminPortal.AdminCreateOEMSolution)
		group.PUT("/oem-solutions/:id", h.AdminPortal.AdminUpdateOEMSolution)
		group.DELETE("/oem-solutions/:id", h.AdminPortal.AdminDeleteOEMSolution)

	// Reports
	group.GET("/reports/revenue-trends", h.AdminPortal.GetRevenueTrends)
	group.GET("/reports/order-trends", h.AdminPortal.GetOrderTrends)
	group.GET("/reports/top-products", h.AdminPortal.GetTopProducts)
	group.GET("/reports/conversion-trends", h.AdminPortal.GetConversionTrends)
		group.GET("/reports/sales-velocity", h.AdminPortal.GetSalesVelocity)
		group.GET("/reports/rfm", h.AdminPortal.GetRFMAnalysis)
		group.GET("/reports/churn", h.AdminPortal.GetCustomerChurn)
		group.GET("/reports/inventory-health", h.AdminPortal.GetInventoryHealth)
		group.GET("/reports/profit-loss", h.AdminPortal.GetProfitLoss)
		group.GET("/reports/replenishment", h.AdminPortal.GetReplenishmentSuggestions)

	// Financial
	group.GET("/financial/overview", h.AdminPortal.GetFinancialOverview)
	group.GET("/financial/outstanding-invoices", h.AdminPortal.GetOutstandingInvoices)
	group.GET("/financial/payment-breakdown", h.AdminPortal.GetPaymentBreakdown)

	// Staff & Audit
	group.GET("/staff", h.AdminPortal.GetStaffList)
	group.GET("/staff/:id/activity", h.AdminPortal.GetStaffActivity)
	group.GET("/audit-log", h.AdminPortal.GetAuditLog)

	// XLSX Export & Translation
	group.GET("/xlsx/export/products", h.AdminPortal.AdminExportProductsXLSX)
	group.GET("/xlsx/export/orders", h.AdminPortal.AdminExportOrdersXLSX)
	group.GET("/xlsx/export/revenue", h.AdminPortal.AdminExportRevenueXLSX)
	group.GET("/xlsx/export/trades", h.AdminPortal.AdminExportTradesXLSX)
	group.GET("/xlsx/export/customers", h.AdminPortal.AdminExportCustomersXLSX)
	group.POST("/xlsx/translate", h.AdminPortal.AdminTranslateXLSX)
	group.POST("/xlsx/translate-batch", h.AdminPortal.AdminBatchTranslateXLSX)

		// Hooks (Plugin/Event System)
		group.GET("/hooks", h.AdminPortal.AdminListHooks)
		group.GET("/hooks/:id", h.AdminPortal.AdminGetHook)
		group.POST("/hooks", h.AdminPortal.AdminCreateHook)
		group.PUT("/hooks/:id", h.AdminPortal.AdminUpdateHook)
		group.DELETE("/hooks/:id", h.AdminPortal.AdminDeleteHook)
		group.GET("/hooks/executions", h.AdminPortal.AdminListHookExecutions)
		group.GET("/events", h.AdminPortal.AdminListEvents)

		// Sales Channels
		group.GET("/channels", h.AdminPortal.AdminListChannels)
		group.POST("/channels", h.AdminPortal.AdminCreateChannel)
		group.GET("/channels/:id", h.AdminPortal.AdminGetChannel)
		group.PUT("/channels/:id", h.AdminPortal.AdminUpdateChannel)
		group.DELETE("/channels/:id", h.AdminPortal.AdminDeleteChannel)

		// Webhooks
		group.GET("/webhooks", h.AdminPortal.AdminListWebhooks)
		group.POST("/webhooks", h.AdminPortal.AdminCreateWebhook)
		group.PUT("/webhooks/:id", h.AdminPortal.AdminUpdateWebhook)
		group.DELETE("/webhooks/:id", h.AdminPortal.AdminDeleteWebhook)
		group.GET("/webhook-deliveries", h.AdminPortal.AdminListWebhookDeliveries)

	// Settings
	group.GET("/settings", h.AdminPortal.GetSettings)
	group.PUT("/settings/:key", h.AdminPortal.UpdateSetting)

	// Translations
	if h.AdminPortal.Translations != nil {
		t := h.AdminPortal.Translations
		group.GET("/translations", t.ListTranslations)
		group.GET("/translations/groups", t.ListGroups)
		group.GET("/translations/:id", t.GetTranslation)
		group.POST("/translations", t.CreateTranslation)
		group.PUT("/translations/:id", t.UpdateTranslation)
		group.DELETE("/translations/:id", t.DeleteTranslation)
		group.POST("/translations/import", t.ImportTranslations)
		group.GET("/translations/export", t.ExportTranslations)
	}
}
