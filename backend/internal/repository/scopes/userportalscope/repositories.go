package userportalscope

import (
	inquiry "candypro/api/internal/repository/inquiry"
	notificationrepo "candypro/api/internal/repository/notification"
	oem "candypro/api/internal/repository/oem"
	order "candypro/api/internal/repository/order"
	product "candypro/api/internal/repository/product"
	trade "candypro/api/internal/repository/trade"
	user "candypro/api/internal/repository/user"

	"gorm.io/gorm"
)

type Repositories struct {
	User           *user.UserRepository
	Company        *user.CompanyRepository
	Inquiry        *inquiry.InquiryRepository
	Order          *order.OrderRepository
	Payment        *order.PaymentRepository
	Invoice        *order.InvoiceRepository
	Cart           *order.CartRepository
	Product        *product.ProductRepository
	Price          *product.PriceRepository
	Project        *oem.ProjectRepository
	Trade          trade.TradeRepository
	Shipment       *trade.ShipmentRepository
	ShipmentEvent  *trade.ShipmentEventRepository
	TradeDocDetail *trade.TradeDocumentDetailRepository
	OrderMessage   *order.OrderMessageRepository
	Negotiation    *order.NegotiationRepository
	Shipping       *order.ShippingRepository
	Tax            *order.TaxRepository
	Notification   *notificationrepo.NotificationRepository
	Return           *order.ReturnRepository
	Coupon           *order.CouponRepository
	RequisitionList  *order.RequisitionListRepository
	Webhook          *order.WebhookRepository
	BuyerOrg       *order.BuyerOrgRepository
	OrgMember      *order.OrgMemberRepository
	ApprovalAction *order.ApprovalActionRepository
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		User:           user.NewUserRepository(db),
		Company:        user.NewCompanyRepository(db),
		Inquiry:        inquiry.NewInquiryRepository(db),
		Order:          order.NewOrderRepository(db),
		Payment:        order.NewPaymentRepository(db),
		Invoice:        order.NewInvoiceRepository(db),
		Cart:           order.NewCartRepository(db),
		Product:        product.NewProductRepository(db),
		Price:          product.NewPriceRepository(db),
		Project:        oem.NewProjectRepository(db),
		Trade:          trade.NewTradeRepository(db),
		Shipment:       trade.NewShipmentRepository(db),
		ShipmentEvent:  trade.NewShipmentEventRepository(db),
		TradeDocDetail: trade.NewTradeDocumentDetailRepository(db),
		OrderMessage:   order.NewOrderMessageRepository(db),
		Negotiation:    order.NewNegotiationRepository(db),
		Shipping:       order.NewShippingRepository(db),
			Tax:            order.NewTaxRepository(db),
		Notification:   notificationrepo.NewNotificationRepository(db),
		Return:          order.NewReturnRepository(db),
		Coupon:          order.NewCouponRepository(db),
		RequisitionList: order.NewRequisitionListRepository(db),
		Webhook:         order.NewWebhookRepository(db),
		BuyerOrg:       order.NewBuyerOrgRepository(db),
		OrgMember:      order.NewOrgMemberRepository(db),
		ApprovalAction: order.NewApprovalActionRepository(db),
	}
}
