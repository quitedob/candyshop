package userportalscope

import (
	"candypro/api/internal/config"
	repositoryCommon "candypro/api/internal/repository/common"
	orderRepo "candypro/api/internal/repository/order"
	inquiry "candypro/api/internal/services/inquiry"
	notificationsvc "candypro/api/internal/services/notification"
	oem "candypro/api/internal/services/oem"
	order "candypro/api/internal/services/order"
	product "candypro/api/internal/services/product"
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
	Cart           *order.CartService
	Product        *product.ProductService
	Price          *product.PriceService
	OEM            *oem.ProjectService
	Trade          *trade.TradeService
	Shipment       *trade.ShipmentService
	Logistics      *trade.LogisticsService
	TradeDocDetail *trade.TradeDocumentDetailService
	OrderMessage   *order.OrderMessageService
	Negotiation    *order.NegotiationService
	Shipping       *order.ShippingService
	Tax            *order.TaxService
	Notification   *notificationsvc.NotificationService
	Return          *orderRepo.ReturnRepository
	Coupon          *orderRepo.CouponRepository
	RequisitionList *orderRepo.RequisitionListRepository
	Webhook         *order.WebhookService
}

func New(repos *repositoryCommon.UserPortalRepositories, cfg *config.Config, db *gorm.DB) *Services {
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
		Invoice:        order.NewInvoiceService(repos.Invoice, repos.Order, nil),
		Cart:           order.NewCartService(repos.Cart),
		Product:        product.NewProductService(repos.Product),
		Price:          product.NewPriceService(repos.Price),
		OEM:            oem.NewProjectService(repos.Project),
		Trade:          trade.NewTradeService(repos.Trade),
		Shipment:       trade.NewShipmentService(repos.Shipment),
		Logistics:      trade.NewLogisticsService(repos.Shipment, repos.ShipmentEvent, repos.Order, repos.Trade, db),
		TradeDocDetail: trade.NewTradeDocumentDetailService(repos.TradeDocDetail),
		OrderMessage:   order.NewOrderMessageService(repos.OrderMessage),
		Negotiation:    order.NewNegotiationService(repos.Negotiation),
		Shipping:       order.NewShippingService(repos.Shipping),
			Tax:            order.NewTaxService(repos.Tax),
		Notification:   notificationsvc.NewNotificationService(repos.Notification),
		Return:          repos.Return,
		Coupon:          repos.Coupon,
		RequisitionList: repos.RequisitionList,
		Webhook:         order.NewWebhookService(repos.Webhook),
	}
}
