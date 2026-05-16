package systemscope

import (
	content "candypro/api/internal/repository/content"
	inquiry "candypro/api/internal/repository/inquiry"
	order "candypro/api/internal/repository/order"
	product "candypro/api/internal/repository/product"
	trade "candypro/api/internal/repository/trade"
	user "candypro/api/internal/repository/user"

	"gorm.io/gorm"
)

type Repositories struct {
	Inquiry *inquiry.InquiryRepository
	Order   *order.OrderRepository
	Product *product.ProductRepository
	Content *content.ContentRepository
	Trade   trade.TradeRepository
	User    *user.UserRepository
	Payment *order.PaymentRepository
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		Inquiry: inquiry.NewInquiryRepository(db),
		Order:   order.NewOrderRepository(db),
		Product: product.NewProductRepository(db),
		Content: content.NewContentRepository(db),
		Trade:   trade.NewTradeRepository(db),
		User:    user.NewUserRepository(db),
		Payment: order.NewPaymentRepository(db),
	}
}
