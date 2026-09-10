package systemscope

import (
	"candypro/api/internal/config"
	repositoryCommon "candypro/api/internal/repository/common"
	inquiry "candypro/api/internal/services/inquiry"
	order "candypro/api/internal/services/order"
	product "candypro/api/internal/services/product"
	search "candypro/api/internal/services/search"
	trade "candypro/api/internal/services/trade"
	user "candypro/api/internal/services/user"
)

type Services struct {
	Inquiry *inquiry.InquiryService
	Order   *order.OrderService
	Product *product.ProductService
	Search  *search.SearchService
	Trade           *trade.TradeService
	User            *user.UserService
	Payment         *order.PaymentService
	GatewayPayment  *order.GatewayPaymentService
	QuotationReview *trade.QuotationReviewService
}

func New(repos *repositoryCommon.SystemRepositories, cfg *config.Config, searchSvc *search.SearchService) *Services {
	if repos == nil {
		return &Services{Search: searchSvc}
	}

	paymentSvc := order.NewPaymentService(repos.Payment, repos.Order)

	return &Services{
		Inquiry:        inquiry.NewInquiryService(repos.Inquiry, cfg),
		Order:          order.NewOrderService(repos.Order),
		Product:        product.NewProductService(repos.Product),
		Search:         searchSvc,
		Trade:           trade.NewTradeService(repos.Trade),
		User:            user.NewUserService(repos.User),
		Payment:         paymentSvc,
		GatewayPayment:  order.NewGatewayPaymentService(cfg, paymentSvc),
		QuotationReview: trade.NewQuotationReviewService(repos.QuotationReview, nil),
	}
}
