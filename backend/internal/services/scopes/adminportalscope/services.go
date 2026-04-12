package adminportalscope

import (
	"candypro/api/internal/config"
	repositoryCommon "candypro/api/internal/repository/common"
	auth "candypro/api/internal/services/auth"
	content "candypro/api/internal/services/content"
	inquiry "candypro/api/internal/services/inquiry"
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
	Product        *product.ProductService
	Price          *product.PriceService
	Content        *content.ContentService
	Auth           *auth.AuthService
	Factory        *oem.FactoryService
	Project        *oem.ProjectService
	Trade          *trade.TradeService
	Shipment       *trade.ShipmentService
	TradeDocDetail *trade.TradeDocumentDetailService
}

func New(repos *repositoryCommon.AdminPortalRepositories, cfg *config.Config, authSvc *auth.AuthService) *Services {
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
		Invoice:        order.NewInvoiceService(repos.Invoice),
		Product:        product.NewProductService(repos.Product),
		Price:          product.NewPriceService(repos.Price),
		Content:        content.NewContentService(repos.Content),
		Auth:           authSvc,
		Factory:        oem.NewFactoryService(repos.Factory),
		Project:        oem.NewProjectService(repos.Project),
		Trade:          trade.NewTradeService(repos.Trade),
		Shipment:       trade.NewShipmentService(repos.Shipment),
		TradeDocDetail: trade.NewTradeDocumentDetailService(repos.TradeDocDetail),
	}
}
