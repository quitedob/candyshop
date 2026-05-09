package database

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	modelsTrade "candypro/api/internal/models/trade"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/roles"
	"candypro/api/internal/utils"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	demoCompanyID   = "seed-demo-company"
	demoBuyerUserID = "seed-demo-buyer"
	demoOpsUserID   = "seed-demo-ops"
	demoInquiryID   = "seed-inquiry-demo-001"
	demoOrder1      = "seed-order-demo-pending"
	demoOrder2      = "seed-order-demo-confirmed"
	demoOrder3      = "seed-order-demo-production"
)

// SeedDemoWorkspace 幂等写入演示 B2B 客户、运营账号、询盘与多状态订单（不扣减库存，便于现有种子数据一致）
func SeedDemoWorkspace(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("SEED_DEMO_DATA")), "false") {
		log.Println("Demo workspace seed skipped (SEED_DEMO_DATA=false)")
		return nil
	}

	roleIDs, err := ensureSystemRoles(db)
	if err != nil {
		return err
	}

	pwd := strings.TrimSpace(os.Getenv("SEED_DEMO_PASSWORD"))
	if pwd == "" {
		pwd = "DemoBuyer123!"
	}
	hash, err := utils.HashPassword(pwd)
	if err != nil {
		return err
	}

	now := time.Now()
	verified := now

	// 演示公司（已认证，便于下单与账期演示）
	company := modelsUser.Company{
		ID:           demoCompanyID,
		Name:         "Demo Global Sourcing Ltd",
		TaxID:        "DEMO-TAX-US-001",
		Status:       "verified",
		VerifiedAt:   &verified,
		PaymentTerms: "NET_30",
		CreditLimit:  500000,
		Phone:        "+1-555-0100",
		Website:      "https://demo.candypro.local",
		Address: modelsUser.CompanyAddress{
			Street:  "100 Trade Center Dr",
			City:    "Newark",
			State:   "NJ",
			ZipCode: "07102",
			Country: "USA",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Where("id = ?", demoCompanyID).FirstOrCreate(&company).Error; err != nil {
		return err
	}

	customerRoleID := roleIDs[roles.User]
	adminRoleID := roleIDs[roles.Admin]

	buyer := modelsUser.User{
		ID:              demoBuyerUserID,
		FirstName:       "Alex",
		LastName:        "Buyer",
		Email:           "buyer.demo@candypro.local",
		PasswordHash:    hash,
		Phone:           "+1-555-0101",
		Company:         company.Name,
		CompanyID:       strPtr(demoCompanyID),
		Status:          "active",
		EmailVerified:   true,
		EmailVerifiedAt: &now,
		RoleID:          customerRoleID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := upsertDemoUser(db, &buyer); err != nil {
		return err
	}

	ops := modelsUser.User{
		ID:              demoOpsUserID,
		FirstName:       "Jamie",
		LastName:        "Operations",
		Email:           "ops.demo@candypro.local",
		PasswordHash:    hash,
		Status:          "active",
		EmailVerified:   true,
		EmailVerifiedAt: &now,
		RoleID:          adminRoleID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := upsertDemoUser(db, &ops); err != nil {
		return err
	}

	var p1, p2 modelsProduct.Product
	if err := db.Where("slug = ?", "4d-fruit-gummy").First(&p1).Error; err != nil {
		return err
	}
	if err := db.Where("slug = ?", "crystal-hard-candy").First(&p2).Error; err != nil {
		return err
	}

	validUntil := now.AddDate(0, 1, 0)
	inquiry := modelsProduct.Inquiry{
		ID:                     demoInquiryID,
		UserID:                 strPtr(demoBuyerUserID),
		CompanyName:            company.Name,
		ContactPerson:          "Alex Buyer",
		Email:                  buyer.Email,
		TargetCountry:          "USA",
		EstimatedQuantity:      "10000 kg",
		InterestedProducts:     modelsCommon.StringArray{p1.ID, p2.ID},
		Products:               modelsCommon.StringArray{p1.ID},
		PackagingRequirements:  "Cartons 10kg, English labels",
		FlavorRequirements:     "Assorted fruit",
		Status:                 "won",
		Priority:               "normal",
		QuotedAmount:           125000,
		QuotedAt:               &now,
		ValidUntil:             &validUntil,
		Incoterms:              "CIF New York",
		NegotiatedPaymentTerms: "30% T/T advance, 70% before shipment",
		Message:                "Demo inquiry for OEM gummy + hard candy mix.",
		CreatedAt:              now.AddDate(0, 0, -14),
		UpdatedAt:              now,
	}
	if err := db.Where("id = ?", demoInquiryID).FirstOrCreate(&inquiry).Error; err != nil {
		return err
	}

	addr := modelsOrder.Address{
		Street:  company.Address.Street,
		City:    company.Address.City,
		State:   company.Address.State,
		ZipCode: company.Address.ZipCode,
		Country: company.Address.Country,
	}

	// 订单1：待确认 / 未付款（无库存预占，纯展示）
	orderPending := modelsOrder.Order{
		ID:              demoOrder1,
		OrderNumber:     "DEMO-ORD-PENDING-01",
		UserID:          demoBuyerUserID,
		InquiryID:       strPtr(demoInquiryID),
		Status:          "pending",
		PaymentStatus:   "unpaid",
		StockReserved:   false,
		Items: modelsOrder.OrderItemArray{
			{ProductID: p1.ID, Quantity: 5000, UnitPrice: 8.5},
		},
		Subtotal:        5000 * 8.5,
		TotalAmount:     5000 * 8.5,
		Currency:        "USD",
		ShippingAddress: addr,
		CreatedAt:       now.AddDate(0, 0, -3),
		UpdatedAt:       now,
	}
	if err := upsertDemoOrder(db, &orderPending); err != nil {
		return err
	}

	confirmedAt := now.AddDate(0, 0, -2)
	orderConfirmed := modelsOrder.Order{
		ID:              demoOrder2,
		OrderNumber:     "DEMO-ORD-CONFIRMED-01",
		UserID:          demoBuyerUserID,
		InquiryID:       strPtr(demoInquiryID),
		Status:          "confirmed",
		PaymentStatus:   "partial",
		StockReserved:   false,
		Items: modelsOrder.OrderItemArray{
			{ProductID: p1.ID, Quantity: 8000, UnitPrice: 8.2},
			{ProductID: p2.ID, Quantity: 4000, UnitPrice: 5.1},
		},
		Subtotal:        8000*8.2 + 4000*5.1,
		TotalAmount:     8000*8.2 + 4000*5.1,
		Currency:        "USD",
		ShippingAddress: addr,
		ConfirmedAt:     &confirmedAt,
		CreatedAt:       now.AddDate(0, 0, -7),
		UpdatedAt:       now,
	}
	if err := upsertDemoOrder(db, &orderConfirmed); err != nil {
		return err
	}

	orderProduction := modelsOrder.Order{
		ID:              demoOrder3,
		OrderNumber:     "DEMO-ORD-PRODUCTION-01",
		UserID:          demoBuyerUserID,
		Status:          "production",
		PaymentStatus:   "paid",
		StockReserved:   false,
		Items: modelsOrder.OrderItemArray{
			{ProductID: p2.ID, Quantity: 12000, UnitPrice: 4.95},
		},
		Subtotal:        12000 * 4.95,
		TotalAmount:     12000 * 4.95,
		Currency:        "USD",
		ShippingAddress: addr,
		ConfirmedAt:     &confirmedAt,
		CreatedAt:       now.AddDate(0, 0, -10),
		UpdatedAt:       now,
	}
	if err := upsertDemoOrder(db, &orderProduction); err != nil {
		return err
	}

	// --- Seed demo payments ---
	seedDemoPayments(db, now)

	// --- Seed demo invoices ---
	seedDemoInvoices(db, now, orderConfirmed, orderProduction)

	// 与已确认订单关联的贸易主单（演示订单→贸易链路）
	var tradeID uint
	var tradeCount int64
	db.Model(&modelsTrade.TradeTransaction{}).Where("order_id = ?", demoOrder2).Count(&tradeCount)
	if tradeCount == 0 {
		var existing modelsTrade.TradeTransaction
		ref := "TRD-DEMO-SEED-01"
		if err := db.Where("reference = ?", ref).First(&existing).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			tr := modelsTrade.TradeTransaction{
				UserID:          demoBuyerUserID,
				OrderID:         strPtr(demoOrder2),
				InquiryID:       strPtr(demoInquiryID),
				Reference:       ref,
				Status:          modelsTrade.TradeStatusPending,
				Currency:        "USD",
				TotalAmount:     orderConfirmed.TotalAmount,
				Terms:           inquiry.Incoterms,
				CommercialNotes: "Packaging: " + inquiry.PackagingRequirements + "; Payment: " + inquiry.NegotiatedPaymentTerms,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := db.Create(&tr).Error; err != nil {
				return err
			}
			tradeID = tr.ID
		} else if err != nil {
			return err
		} else {
			tradeID = existing.ID
		}
	} else {
		var tr modelsTrade.TradeTransaction
		db.Where("order_id = ?", demoOrder2).First(&tr)
		tradeID = tr.ID
	}

	// --- Seed demo trade documents ---
	seedDemoTradeDocuments(db, tradeID, orderConfirmed, now)

	log.Printf("Demo workspace: buyer %s / ops %s (password from SEED_DEMO_PASSWORD or default DemoBuyer123!)", buyer.Email, ops.Email)
	return nil
}

func seedDemoPayments(db *gorm.DB, now time.Time) {
	payments := []modelsOrder.Payment{
		{
			ID:        "seed-payment-advance",
			OrderID:   demoOrder2,
			Amount:    42680, // ~30% advance of 85600
			Currency:  "USD",
			Method:    "wire",
			Status:    "confirmed",
			Reference: "DEMO-WIRE-ADV-001",
			Notes:     "30% T/T advance payment (demo)",
			CreatedAt: now.AddDate(0, 0, -2),
			UpdatedAt: now.AddDate(0, 0, -2),
		},
		{
			ID:        "seed-payment-full",
			OrderID:   demoOrder3,
			Amount:    59400, // full payment for production order
			Currency:  "USD",
			Method:    "wire",
			Status:    "confirmed",
			Reference: "DEMO-WIRE-FULL-001",
			Notes:     "Full T/T payment (demo)",
			CreatedAt: now.AddDate(0, 0, -9),
			UpdatedAt: now.AddDate(0, 0, -8),
		},
	}
	for _, p := range payments {
		db.Where("id = ?", p.ID).FirstOrCreate(&p)
	}
}

func seedDemoInvoices(db *gorm.DB, now time.Time, orderConfirmed, orderProduction modelsOrder.Order) {
	dueDate := now.AddDate(0, 0, 30)
	dueDateProd := now.AddDate(0, 0, 15)
	sentAt := now.AddDate(0, 0, -1)

	invoices := []modelsOrder.Invoice{
		{
			ID:          "seed-invoice-confirmed",
			OrderID:     orderConfirmed.ID,
			Type:        modelsOrder.InvoiceTypeProforma,
			Status:      modelsOrder.InvoiceStatusSent,
			InvoiceNo:   "INV-DEMO-PI-001",
			Amount:      orderConfirmed.TotalAmount,
			TaxAmount:   0,
			TotalAmount: orderConfirmed.TotalAmount,
			Currency:    "USD",
			DueDate:     &dueDate,
			SentAt:      &sentAt,
			Notes:       "Proforma invoice for confirmed order (demo)",
			CreatedAt:   now.AddDate(0, 0, -2),
			UpdatedAt:   now,
		},
		{
			ID:          "seed-invoice-production",
			OrderID:     orderProduction.ID,
			Type:        modelsOrder.InvoiceTypeCommercial,
			Status:      modelsOrder.InvoiceStatusPaid,
			InvoiceNo:   "INV-DEMO-CI-001",
			Amount:      orderProduction.TotalAmount,
			TaxAmount:   0,
			TotalAmount: orderProduction.TotalAmount,
			Currency:    "USD",
			DueDate:     &dueDateProd,
			SentAt:      &sentAt,
			PaidAt:      &now,
			Notes:       "Commercial invoice for production order (demo)",
			CreatedAt:   now.AddDate(0, 0, -9),
			UpdatedAt:   now,
		},
	}
	for _, inv := range invoices {
		db.Where("id = ?", inv.ID).FirstOrCreate(&inv)
	}
}

func seedDemoTradeDocuments(db *gorm.DB, tradeID uint, orderConfirmed modelsOrder.Order, now time.Time) {
	if tradeID == 0 {
		return
	}

	// TradeDocument envelope entries
	docs := []modelsTrade.TradeDocument{
		{
			TransactionID: tradeID,
			Type:          "proforma_invoice",
			DocNumber:     "PI-DEMO-2026-001",
			Status:        modelsTrade.TradeStatusConfirmed,
			Content: datatypes.JSON(`{
				"buyerName": "Demo Global Sourcing Ltd",
				"sellerName": "CandyPro Manufacturing",
				"incoterms": "CIF New York",
				"paymentTerms": "30% T/T advance, 70% before shipment"
			}`),
			CreatedAt: now.AddDate(0, 0, -2),
			UpdatedAt: now,
		},
		{
			TransactionID: tradeID,
			Type:          "sales_contract",
			DocNumber:     "SC-DEMO-2026-001",
			Status:        modelsTrade.TradeStatusConfirmed,
			Content: datatypes.JSON(`{
				"buyerName": "Demo Global Sourcing Ltd",
				"sellerName": "CandyPro Manufacturing",
				"incoterms": "CIF New York",
				"totalAmount": 85600,
				"currency": "USD"
			}`),
			CreatedAt: now.AddDate(0, 0, -2),
			UpdatedAt: now,
		},
		{
			TransactionID: tradeID,
			Type:          "packing_list",
			DocNumber:     "PL-DEMO-2026-001",
			Status:        modelsTrade.TradeStatusDraft,
			Content: datatypes.JSON(`{
				"totalCartons": 400,
				"totalGrossWeight": 3200,
				"totalNetWeight": 3000,
				"totalVolume": 12.5
			}`),
			CreatedAt: now.AddDate(0, 0, -1),
			UpdatedAt: now,
		},
	}
	for _, doc := range docs {
		var existing modelsTrade.TradeDocument
		if err := db.Where("transaction_id = ? AND type = ?", doc.TransactionID, doc.Type).First(&existing).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			db.Create(&doc)
		}
	}
}

func strPtr(s string) *string {
	return &s
}

func upsertDemoUser(db *gorm.DB, u *modelsUser.User) error {
	var existing modelsUser.User
	err := db.Where("email = ?", u.Email).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(u).Error
	}
	if err != nil {
		return err
	}
	needs := false
	if existing.RoleID != u.RoleID {
		existing.RoleID = u.RoleID
		needs = true
	}
	if existing.Status != u.Status {
		existing.Status = u.Status
		needs = true
	}
	if u.CompanyID != nil && (existing.CompanyID == nil || *existing.CompanyID != *u.CompanyID) {
		existing.CompanyID = u.CompanyID
		needs = true
	}
	if !existing.EmailVerified {
		existing.EmailVerified = true
		existing.EmailVerifiedAt = u.EmailVerifiedAt
		needs = true
	}
	if needs {
		existing.UpdatedAt = time.Now()
		return db.Save(&existing).Error
	}
	return nil
}

func upsertDemoOrder(db *gorm.DB, o *modelsOrder.Order) error {
	var cur modelsOrder.Order
	err := db.Where("id = ?", o.ID).First(&cur).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Omit("User", "Inquiry").Create(o).Error
	}
	if err != nil {
		return err
	}
	// 已存在则仅补齐关键演示字段，避免覆盖用户真实修改
	cur.Status = o.Status
	cur.PaymentStatus = o.PaymentStatus
	cur.Items = o.Items
	cur.Subtotal = o.Subtotal
	cur.TotalAmount = o.TotalAmount
	cur.Currency = o.Currency
	cur.ShippingAddress = o.ShippingAddress
	cur.InquiryID = o.InquiryID
	cur.StockReserved = o.StockReserved
	if o.ConfirmedAt != nil {
		cur.ConfirmedAt = o.ConfirmedAt
	}
	cur.UpdatedAt = time.Now()
	return db.Omit("User", "Inquiry").Save(&cur).Error
}
