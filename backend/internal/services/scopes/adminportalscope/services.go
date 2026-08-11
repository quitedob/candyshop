package adminportalscope

import (
	"candypro/api/internal/config"
	repositoryCommon "candypro/api/internal/repository/common"
	orderRepo "candypro/api/internal/repository/order"
	activitylogSvc "candypro/api/internal/services/activitylog"
	auth "candypro/api/internal/services/auth"
	content "candypro/api/internal/services/content"
	inquiry "candypro/api/internal/services/inquiry"
	notificationsvc "candypro/api/internal/services/notification"
	oem "candypro/api/internal/services/oem"
	order "candypro/api/internal/services/order"
	orderintake "candypro/api/internal/services/orderintake"
	productRepo "candypro/api/internal/repository/product"
	product "candypro/api/internal/services/product"
	systemsettingSvc "candypro/api/internal/services/systemsetting"
	trade "candypro/api/internal/services/trade"
	translationSvc "candypro/api/internal/services/translation"
	user "candypro/api/internal/services/user"

	"gorm.io/gorm"
)

type Services struct {
	User             *user.UserService
	Company          *user.CompanyService
	Inquiry          *inquiry.InquiryService
	Order            *order.OrderService
	Payment          *order.PaymentService
	GatewayPayment   *order.GatewayPaymentService
	Invoice          *order.InvoiceService
	Product          *product.ProductService
	Category         *product.CategoryService
	Price            *product.PriceService
	Content          *content.ContentService
	Auth             *auth.AuthService
	Factory          *oem.FactoryService
	Project          *oem.ProjectService
	OEM              *oem.OEMService
	Trade            *trade.TradeService
	Shipment         *trade.ShipmentService
	Logistics        *trade.LogisticsService
	TradeDocDetail   *trade.TradeDocumentDetailService
	QuotationReview  *trade.QuotationReviewService
	ActivityLog      *activitylogSvc.ActivityLogService
	SystemSetting    *systemsettingSvc.SystemSettingService
	StockTransaction *orderRepo.StockTransactionRepository
	StockTransfer    *orderRepo.StockTransferRepository
	Fulfillment      *orderRepo.FulfillmentRepository
	Return           *orderRepo.ReturnRepository
	Coupon           *orderRepo.CouponRepository
	Supplier         *productRepo.SupplierRepository
	Channel          *productRepo.ChannelRepository
	Translation      *translationSvc.TranslationService
	OrderMessage     *order.OrderMessageService
	Negotiation      *order.NegotiationService
	OrderIntake      *orderintake.Service
	Shipping         *order.ShippingService
	Tax              *order.TaxService
	Notification     *notificationsvc.NotificationService
	Webhook          *order.WebhookService
	Approval         *order.ApprovalService
	HookConfig       *order.HookConfigService
	EventBus         *order.EventBus
}

func New(repos *repositoryCommon.AdminPortalRepositories, cfg *config.Config, authSvc *auth.AuthService, db *gorm.DB) *Services {
	if repos == nil {
		return &Services{Auth: authSvc}
	}

	userSvc := user.NewUserService(repos.User)
	paymentSvc := order.NewPaymentService(repos.Payment, repos.Order)
	inquirySvc := inquiry.NewInquiryService(repos.Inquiry, cfg)
	orderSvc := order.NewOrderServiceWithConfig(repos.Order, cfg)
	productSvc := product.NewProductService(repos.Product)
	priceSvc := product.NewPriceService(repos.Price)
	tradeSvc := trade.NewTradeService(repos.Trade)

	return &Services{
		User:             userSvc,
		Company:          user.NewCompanyService(repos.Company, userSvc),
		Inquiry:          inquirySvc,
		Order:            orderSvc,
		Payment:          paymentSvc,
		GatewayPayment:   order.NewGatewayPaymentService(cfg, paymentSvc),
		Invoice:          order.NewInvoiceService(repos.Invoice, repos.Order, repos.DocumentAdjustment),
		Product:          productSvc,
		Category:         product.NewCategoryService(repos.Category),
		Price:            priceSvc,
		Content:          content.NewContentService(repos.Content),
		Auth:             authSvc,
		Factory:          oem.NewFactoryService(repos.Factory),
		Project:          oem.NewProjectService(repos.Project),
		OEM:              oem.NewOEMService(repos.OEM),
		Trade:            tradeSvc,
		Shipment:         trade.NewShipmentService(repos.Shipment),
		Logistics:        trade.NewLogisticsService(repos.Shipment, repos.ShipmentEvent, repos.Order, repos.Trade, db),
		TradeDocDetail:   trade.NewTradeDocumentDetailService(repos.TradeDocDetail),
		QuotationReview:  trade.NewQuotationReviewService(repos.QuotationReview),
		ActivityLog:      activitylogSvc.NewActivityLogService(repos.ActivityLog),
		SystemSetting:    systemsettingSvc.NewSystemSettingService(repos.SystemSetting),
		StockTransaction: repos.StockTransaction,
		StockTransfer:    repos.StockTransfer,
		Fulfillment:      repos.Fulfillment,
		Return:           repos.Return,
		Coupon:           repos.Coupon,
		Supplier:         repos.Supplier,
		Channel:          repos.Channel,
		Translation:      translationSvc.NewService(repos.Translation),
		OrderMessage:     order.NewOrderMessageService(repos.OrderMessage),
		Negotiation:      order.NewNegotiationService(repos.Negotiation),
		OrderIntake:      orderintake.NewService(productSvc, priceSvc, orderSvc, tradeSvc, inquirySvc),
		Shipping:         order.NewShippingService(repos.Shipping),
			Tax:              order.NewTaxService(repos.Tax),
		Notification:     notificationsvc.NewNotificationService(repos.Notification),
		Webhook:          order.NewWebhookService(repos.Webhook),
		Approval:         order.NewApprovalService(repos.BuyerOrg, repos.OrgMember, repos.ApprovalAction, repos.Order, productSvc),
		HookConfig:       order.NewHookConfigService(repos.HookConfig, repos.HookExecution),
		EventBus:         order.NewEventBus(repos.Event, repos.HookConfig, repos.HookExecution, order.NewWebhookService(repos.Webhook)),
	}
}
