package systemscope

import (
	content "candypro/api/internal/repository/content"
	inquiry "candypro/api/internal/repository/inquiry"
	order "candypro/api/internal/repository/order"
	product "candypro/api/internal/repository/product"
	trade "candypro/api/internal/repository/trade"

	"gorm.io/gorm"
)

type Repositories struct {
	Inquiry *inquiry.InquiryRepository
	Order   *order.OrderRepository
	Product *product.ProductRepository
	Content *content.ContentRepository
	Trade   trade.TradeRepository
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		Inquiry: inquiry.NewInquiryRepository(db),
		Order:   order.NewOrderRepository(db),
		Product: product.NewProductRepository(db),
		Content: content.NewContentRepository(db),
		Trade:   trade.NewTradeRepository(db),
	}
}
