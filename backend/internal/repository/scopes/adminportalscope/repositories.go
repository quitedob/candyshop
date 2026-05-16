package adminportalscope

import (
	activitylog "candypro/api/internal/repository/activitylog"
	auth "candypro/api/internal/repository/auth"
	content "candypro/api/internal/repository/content"
	inquiry "candypro/api/internal/repository/inquiry"
	notificationrepo "candypro/api/internal/repository/notification"
	oem "candypro/api/internal/repository/oem"
	order "candypro/api/internal/repository/order"
	product "candypro/api/internal/repository/product"
	systemsetting "candypro/api/internal/repository/systemsetting"
	trade "candypro/api/internal/repository/trade"
	translation "candypro/api/internal/repository/translation"
	user "candypro/api/internal/repository/user"

	"gorm.io/gorm"
)

type Repositories struct {
	User               *user.UserRepository
	Company            *user.CompanyRepository
	Inquiry            *inquiry.InquiryRepository
	Order              *order.OrderRepository
	Payment            *order.PaymentRepository
	Invoice            *order.InvoiceRepository
	DocumentAdjustment *order.DocumentAdjustmentRepository
	Product            *product.ProductRepository
	Price              *product.PriceRepository
	Content            *content.ContentRepository
	Role               *auth.RoleRepository
	Factory            *oem.FactoryRepository
	Project            *oem.ProjectRepository
	OEM                *oem.OEMRepository
	Trade              trade.TradeRepository
	Shipment           *trade.ShipmentRepository
	ShipmentEvent      *trade.ShipmentEventRepository
	TradeDocDetail     *trade.TradeDocumentDetailRepository
	ActivityLog        *activitylog.ActivityLogRepository
	SystemSetting      *systemsetting.SystemSettingRepository
	StockTransaction   *order.StockTransactionRepository
	Translation        *translation.TranslationRepository
	OrderMessage       *order.OrderMessageRepository
	Negotiation        *order.NegotiationRepository
	Shipping           *order.ShippingRepository
	Tax                *order.TaxRepository
	Notification       *notificationrepo.NotificationRepository
	StockTransfer      *order.StockTransferRepository
	Fulfillment        *order.FulfillmentRepository
	Return             *order.ReturnRepository
	Coupon             *order.CouponRepository
	Supplier           *product.SupplierRepository
	Channel            *product.ChannelRepository
	Webhook            *order.WebhookRepository
		Event              *order.EventRepository
		HookConfig         *order.HookRepository
		HookExecution      *order.HookExecutionRepository
	BuyerOrg           *order.BuyerOrgRepository
	OrgMember          *order.OrgMemberRepository
	ApprovalAction     *order.ApprovalActionRepository
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		User:               user.NewUserRepository(db),
		Company:            user.NewCompanyRepository(db),
		Inquiry:            inquiry.NewInquiryRepository(db),
		Order:              order.NewOrderRepository(db),
		Payment:            order.NewPaymentRepository(db),
		Invoice:            order.NewInvoiceRepository(db),
		DocumentAdjustment: order.NewDocumentAdjustmentRepository(db),
		Product:            product.NewProductRepository(db),
		Price:              product.NewPriceRepository(db),
		Content:            content.NewContentRepository(db),
		Role:               auth.NewRoleRepository(db),
		Factory:            oem.NewFactoryRepository(db),
		Project:            oem.NewProjectRepository(db),
		OEM:                oem.NewOEMRepository(db),
		Trade:              trade.NewTradeRepository(db),
		Shipment:           trade.NewShipmentRepository(db),
		ShipmentEvent:      trade.NewShipmentEventRepository(db),
		TradeDocDetail:     trade.NewTradeDocumentDetailRepository(db),
		ActivityLog:        activitylog.NewActivityLogRepository(db),
		SystemSetting:      systemsetting.NewSystemSettingRepository(db),
		StockTransaction:   order.NewStockTransactionRepository(db),
		Translation:        translation.New(db),
		OrderMessage:       order.NewOrderMessageRepository(db),
		Negotiation:        order.NewNegotiationRepository(db),
		Shipping:           order.NewShippingRepository(db),
			Tax:                order.NewTaxRepository(db),
		Notification:       notificationrepo.NewNotificationRepository(db),
		StockTransfer:      order.NewStockTransferRepository(db),
		Fulfillment:        order.NewFulfillmentRepository(db),
		Return:             order.NewReturnRepository(db),
		Coupon:             order.NewCouponRepository(db),
		Supplier:           product.NewSupplierRepository(db),
		Channel:            product.NewChannelRepository(db),
		Webhook:            order.NewWebhookRepository(db),
			Event:              order.NewEventRepository(db),
			HookConfig:         order.NewHookRepository(db),
			HookExecution:      order.NewHookExecutionRepository(db),
		BuyerOrg:           order.NewBuyerOrgRepository(db),
		OrgMember:          order.NewOrgMemberRepository(db),
		ApprovalAction:     order.NewApprovalActionRepository(db),
	}
}
