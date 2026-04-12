package userportalscope

import (
	"candypro/api/internal/config"
	repositoryCommon "candypro/api/internal/repository/common"
	inquiry "candypro/api/internal/services/inquiry"
	notificationsvc "candypro/api/internal/services/notification"
	oem "candypro/api/internal/services/oem"
	order "candypro/api/internal/services/order"
	product "candypro/api/internal/services/product"
	trade "candypro/api/internal/services/trade"
	user "candypro/api/internal/services/user"
)

type Services struct {
	User           *user.UserService
	Company        *user.CompanyService
	Inquiry        *inquiry.InquiryService
	Order          *order.OrderService
	Payment        *order.PaymentService
	Invoice        *order.InvoiceService
	Cart           *order.CartService
	Product        *product.ProductService
	Price          *product.PriceService
	OEM            *oem.ProjectService
	Trade          *trade.TradeService
	Shipment       *trade.ShipmentService
	TradeDocDetail *trade.TradeDocumentDetailService
	Notification   *notificationsvc.NotificationService
}

func New(repos *repositoryCommon.UserPortalRepositories, cfg *config.Config) *Services {
	if repos == nil {
		return &Services{}
	}

	userSvc := user.NewUserService(repos.User)

	return &Services{
		User:           userSvc,
		Company:        user.NewCompanyService(repos.Company, userSvc),
		Inquiry:        inquiry.NewInquiryService(repos.Inquiry, cfg),
		Order:          order.NewOrderService(repos.Order),
		Payment:        order.NewPaymentService(repos.Payment, repos.Order),
		Invoice:        order.NewInvoiceService(repos.Invoice),
		Cart:           order.NewCartService(repos.Cart),
		Product:        product.NewProductService(repos.Product),
		Price:          product.NewPriceService(repos.Price),
		OEM:            oem.NewProjectService(repos.Project),
		Trade:          trade.NewTradeService(repos.Trade),
		Shipment:       trade.NewShipmentService(repos.Shipment),
		TradeDocDetail: trade.NewTradeDocumentDetailService(repos.TradeDocDetail),
		Notification:   notificationsvc.NewNotificationService(repos.Notification),
	}
}
