package adminportalscope

import (
	activitylog "candypro/api/internal/repository/activitylog"
	auth "candypro/api/internal/repository/auth"
	content "candypro/api/internal/repository/content"
	inquiry "candypro/api/internal/repository/inquiry"
	oem "candypro/api/internal/repository/oem"
	order "candypro/api/internal/repository/order"
	product "candypro/api/internal/repository/product"
	systemsetting "candypro/api/internal/repository/systemsetting"
	trade "candypro/api/internal/repository/trade"
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
	Product        *product.ProductRepository
	Price          *product.PriceRepository
	Content        *content.ContentRepository
	Role           *auth.RoleRepository
	Factory        *oem.FactoryRepository
	Project        *oem.ProjectRepository
	Trade          trade.TradeRepository
	Shipment       *trade.ShipmentRepository
	ShipmentEvent  *trade.ShipmentEventRepository
	TradeDocDetail *trade.TradeDocumentDetailRepository
	ActivityLog    *activitylog.ActivityLogRepository
	SystemSetting  *systemsetting.SystemSettingRepository
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
		Product:        product.NewProductRepository(db),
		Price:          product.NewPriceRepository(db),
		Content:        content.NewContentRepository(db),
		Role:           auth.NewRoleRepository(db),
		Factory:        oem.NewFactoryRepository(db),
		Project:        oem.NewProjectRepository(db),
		Trade:          trade.NewTradeRepository(db),
		Shipment:       trade.NewShipmentRepository(db),
			ShipmentEvent:  trade.NewShipmentEventRepository(db),
		TradeDocDetail: trade.NewTradeDocumentDetailRepository(db),
		ActivityLog:    activitylog.NewActivityLogRepository(db),
		SystemSetting:  systemsetting.NewSystemSettingRepository(db),
	}
}

