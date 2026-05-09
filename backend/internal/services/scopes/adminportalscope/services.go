package adminportalscope

import (
	"candypro/api/internal/config"
	repositoryCommon "candypro/api/internal/repository/common"
	orderRepo "candypro/api/internal/repository/order"
	auth "candypro/api/internal/services/auth"
	activitylogSvc "candypro/api/internal/services/activitylog"
	content "candypro/api/internal/services/content"
	inquiry "candypro/api/internal/services/inquiry"
	notificationsvc "candypro/api/internal/services/notification"
	oem "candypro/api/internal/services/oem"
	order "candypro/api/internal/services/order"
	product "candypro/api/internal/services/product"
	systemsettingSvc "candypro/api/internal/services/systemsetting"
	translationSvc "candypro/api/internal/services/translation"
	trade "candypro/api/internal/services/trade"
	user "candypro/api/internal/services/user"

	"gorm.io/gorm"
)

type Services struct {
	User           *user.UserService
	Company        *user.CompanyService
	Inquiry        *inquiry.InquiryService
	Order          *order.OrderService
	Payment        *order.PaymentService
	Invoice        *order.InvoiceService
	Product        *product.ProductService
	Price          *product.PriceService
	Content        *content.ContentService
	Auth           *auth.AuthService
	Factory        *oem.FactoryService
	Project        *oem.ProjectService
	Trade          *trade.TradeService
	Shipment       *trade.ShipmentService
	Logistics      *trade.LogisticsService
	TradeDocDetail *trade.TradeDocumentDetailService
	ActivityLog     *activitylogSvc.ActivityLogService
	SystemSetting   *systemsettingSvc.SystemSettingService
	StockTransaction *orderRepo.StockTransactionRepository
	Translation      *translationSvc.TranslationService
	Notification     *notificationsvc.NotificationService
}

func New(repos *repositoryCommon.AdminPortalRepositories, cfg *config.Config, authSvc *auth.AuthService, db *gorm.DB) *Services {
	if repos == nil {
		return &Services{Auth: authSvc}
		}

	userSvc := user.NewUserService(repos.User)

	return &Services{
		User:           userSvc,
		Company:        user.NewCompanyService(repos.Company, userSvc),
		Inquiry:        inquiry.NewInquiryService(repos.Inquiry, cfg),
		Order:          order.NewOrderServiceWithConfig(repos.Order, cfg),
		Payment:        order.NewPaymentService(repos.Payment, repos.Order),
		Invoice:        order.NewInvoiceService(repos.Invoice, repos.Order, repos.DocumentAdjustment),
		Product:        product.NewProductService(repos.Product),
		Price:          product.NewPriceService(repos.Price),
		Content:        content.NewContentService(repos.Content),
		Auth:           authSvc,
		Factory:        oem.NewFactoryService(repos.Factory),
		Project:        oem.NewProjectService(repos.Project),
		Trade:          trade.NewTradeService(repos.Trade),
		Shipment:       trade.NewShipmentService(repos.Shipment),
		Logistics:      trade.NewLogisticsService(repos.Shipment, repos.ShipmentEvent, repos.Order, repos.Trade, db),
		TradeDocDetail: trade.NewTradeDocumentDetailService(repos.TradeDocDetail),
		ActivityLog:    activitylogSvc.NewActivityLogService(repos.ActivityLog),
		SystemSetting:    systemsettingSvc.NewSystemSettingService(repos.SystemSetting),
		StockTransaction: repos.StockTransaction,
		Translation:      translationSvc.NewService(repos.Translation),
		Notification:     notificationsvc.NewNotificationService(repos.Notification),
		}
}
