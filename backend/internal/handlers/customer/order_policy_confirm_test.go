package customer

// G20 regression tests for the confirm-path policy enforcement:
//
//  1. Bulk / requisition / reorder drafts are created with Source=OrderSourceBulk
//     (previously Source="", which broke frontend confirm gating and left the
//     source blank on every downstream record).
//  2. CustomerConfirmOrder re-enforces per-product MOQ before stock is committed
//     (draft paths never validated MOQ, and M-19 overrides can drop below it).
//  3. CustomerConfirmOrder validates shipping-country presence for drafts that
//     carry no official compliance evidence (bulk/requisition/reorder), instead
//     of silently skipping the destination-market compliance recheck when the
//     country is empty.
//  4. CustomerConfirmOrder enforces the buyer company's credit limit
//     cumulatively (open order totals + this order's total <= limit), so a buyer
//     cannot stack multiple under-limit orders that collectively exceed credit.
//
// These tests reuse the shared fakes from customer_orders_write_test.go
// (fakeOrderRepo / fakeProductRepo / fakeUserRepo / fakeInquiryRepo) plus the
// new credit-aware fakes below.

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"candypro/api/internal/config"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	modelsUser "candypro/api/internal/models/user"
	orderRepo "candypro/api/internal/repository/order"
	servicesCommon "candypro/api/internal/services/common"
	inquiryService "candypro/api/internal/services/inquiry"
	orderService "candypro/api/internal/services/order"
	productService "candypro/api/internal/services/product"
	userService "candypro/api/internal/services/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// fakeUserRepoWithCompany returns an active user bound to a company so the
// credit-limit path resolves a company and a limit.
type fakeUserRepoWithCompany struct {
	*fakeUserRepo
	companyID string
}

func (f *fakeUserRepoWithCompany) FindByID(ctx context.Context, id string) (*modelsUser.User, error) {
	if id == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &modelsUser.User{ID: id, Status: "active", CompanyID: &f.companyID}, nil
}

// fakeCompanyRepo returns a single configured company (used to expose a credit
// limit to the cumulative-credit path).
type fakeCompanyRepo struct {
	company *modelsUser.Company
}

func (f *fakeCompanyRepo) FindAll(ctx context.Context, page, limit int) ([]modelsUser.Company, int64, error) {
	return nil, 0, nil
}
func (f *fakeCompanyRepo) FindByID(ctx context.Context, id string) (*modelsUser.Company, error) {
	if f.company == nil {
		return nil, nil
	}
	cp := *f.company
	return &cp, nil
}
func (f *fakeCompanyRepo) Create(ctx context.Context, company *modelsUser.Company) error {
	return nil
}
func (f *fakeCompanyRepo) Update(ctx context.Context, company *modelsUser.Company) error {
	return nil
}
func (f *fakeCompanyRepo) Delete(ctx context.Context, id string) error {
	return nil
}
func (f *fakeCompanyRepo) VerifyCompany(ctx context.Context, id, status string) error {
	return nil
}

// creditAwareOrderRepo embeds fakeOrderRepo, returns seeded open orders from
// FindByUserID (for the cumulative credit sum), and persists confirm financials
// the way the real repository's confirmAndReserve does.
type creditAwareOrderRepo struct {
	*fakeOrderRepo
	openOrders []modelsOrder.Order
}

func (f *creditAwareOrderRepo) FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsOrder.Order, int64, error) {
	// Filter by user so the legacy per-user sum path (pre-G20-r3) sees only the
	// confirming user's own open orders — this is what lets the cross-user
	// stacking regression test distinguish per-user exposure from company
	// exposure.
	var filtered []modelsOrder.Order
	for _, o := range f.openOrders {
		if o.UserID == userID {
			filtered = append(filtered, o)
		}
	}
	return filtered, int64(len(filtered)), nil
}

// SumOpenOrderTotalsByCompany returns the sum of the seeded open orders. The
// fake treats openOrders as the company's open exposure regardless of userID,
// mirroring how the real repository aggregates across a company (G20 r3).
func (f *creditAwareOrderRepo) SumOpenOrderTotalsByCompany(ctx context.Context, companyID, excludeOrderID string) (float64, error) {
	var sum float64
	for _, o := range f.openOrders {
		if o.ID == excludeOrderID {
			continue
		}
		if o.Status == "cancelled" || o.Status == "returned" || o.Status == "expired" {
			continue
		}
		sum += o.TotalAmount
	}
	return sum, nil
}

func (f *creditAwareOrderRepo) ConfirmAndReserveStockWithFinancials(ctx context.Context, id string, stockDeltas map[string]int, confirmedAt time.Time, fin *orderRepo.OrderConfirmFinancials) error {
	if f.createdOrder == nil || f.createdOrder.ID != id {
		return context.Canceled
	}
	f.createdOrder.Status = "pending"
	f.createdOrder.StockReserved = true
	f.createdOrder.ConfirmedAt = &confirmedAt
	f.createdOrder.UpdatedAt = confirmedAt
	if fin != nil {
		if fin.Items != nil {
			f.createdOrder.Items = *fin.Items
		}
		if fin.COGS != nil {
			f.createdOrder.COGS = *fin.COGS
		}
		if fin.Subtotal != nil {
			f.createdOrder.Subtotal = *fin.Subtotal
		}
		if fin.TaxAmount != nil {
			f.createdOrder.TaxAmount = *fin.TaxAmount
		}
		if fin.ShippingAmount != nil {
			f.createdOrder.ShippingAmount = *fin.ShippingAmount
		}
		if fin.TotalAmount != nil {
			f.createdOrder.TotalAmount = *fin.TotalAmount
		}
		if fin.Currency != nil && *fin.Currency != "" {
			f.createdOrder.Currency = *fin.Currency
		}
	}
	return nil
}

// buildPolicyHandler wires a customer handler with the given order repo, an
// optional company (credit limit), and an optional company-bound user repo.
func buildPolicyHandler(products map[string]modelsProduct.Product, userSvc *userService.UserService, company *modelsUser.Company, orderSvc *orderService.OrderService) *Handler {
	cfg := &config.Config{}
	svcs := &servicesCommon.UserPortalServices{
		User:    userSvc,
		Product: productService.NewProductService(&fakeProductRepo{products: products}),
		Inquiry: inquiryService.NewInquiryService(&fakeInquiryRepo{}, cfg),
		Order:   orderSvc,
	}
	if company != nil {
		svcs.Company = userService.NewCompanyService(&fakeCompanyRepo{company: company}, userSvc)
	}
	return NewHandler(cfg, svcs, nil, nil)
}

func performCustomerReorder(handler *Handler, userID, orderID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/customer/orders/"+orderID+"/reorder", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: orderID}}
	c.Set("userID", userID)
	handler.CustomerReorderFromHistory(c)
	return rec
}

func performCustomerBulkUpload(handler *Handler, userID, csv string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", "bulk.csv")
	_, _ = fw.Write([]byte(csv))
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/customer/orders/bulk", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Set("userID", userID)
	handler.CustomerCreateBulkOrder(c)
	return rec
}

// G20(b): the CSV bulk-order draft is created with Source=bulk. Pre-fix the
// order was created with Source="" (OrderSourceCart never set, no source at
// all), which broke frontend confirm gating and downstream source logic.
func TestCustomerCreateBulkOrder_SetsBulkSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{})
	rec := performCustomerBulkUpload(handler, "u-bulk", "product_id,quantity\np-1,100\n")
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder == nil {
		t.Fatalf("expected a draft order to be created")
	}
	if orderRepo.createdOrder.Source != modelsOrder.OrderSourceBulk {
		t.Fatalf("expected Source=%q, got %q", modelsOrder.OrderSourceBulk, orderRepo.createdOrder.Source)
	}
}

// G20(b): reorder-from-history drafts are created with Source=bulk (pre-fix "").
func TestCustomerReorderFromHistory_SetsBulkSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{})
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:     "ord-history",
		UserID: "u-reorder",
		Status: modelsOrder.OrderStatusConfirmed,
		Items: []modelsOrder.OrderItem{
			{ProductID: "p-1", Quantity: 10, UnitPrice: 2.0},
		},
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	rec := performCustomerReorder(handler, "u-reorder", "ord-history")
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder == nil {
		t.Fatalf("expected a draft order to be created")
	}
	if orderRepo.createdOrder.Source != modelsOrder.OrderSourceBulk {
		t.Fatalf("expected Source=%q, got %q", modelsOrder.OrderSourceBulk, orderRepo.createdOrder.Source)
	}
	if orderRepo.createdOrder.UserID != "u-reorder" {
		t.Fatalf("expected reorder to bind to u-reorder, got %s", orderRepo.createdOrder.UserID)
	}
}

// G20(b): MOQ is re-enforced at confirm. A draft line below the product MOQ
// must be rejected before stock is committed. Pre-fix the confirm path never
// validated MOQ, so a qty-5 line against an MOQ-10 product confirmed fine.
func TestCustomerConfirmOrder_MOQBelowEnforced(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-moq": {
			ID: "p-moq", Status: "active", BasePrice: 1.0, MOQ: 10, StockQuantity: 100,
		},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-moq",
		UserID:                     "u-moq",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceCart,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-moq", Quantity: 5, UnitPrice: 0}},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-moq", "ord-moq")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for below-MOQ line, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "min_quantity_not_met" {
		t.Fatalf("expected error=min_quantity_not_met, got %q", got)
	}
	if orderRepo.createdOrder.Status != modelsOrder.OrderStatusPendingConfirm {
		t.Fatalf("expected order to stay pending_confirmation after MOQ rejection, got %s", orderRepo.createdOrder.Status)
	}
}

// G20(b): an MOQ-satisfying confirm still succeeds (control).
func TestCustomerConfirmOrder_MOQMetConfirms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-moq": {
			ID: "p-moq", Status: "active", BasePrice: 1.0, MOQ: 10, StockQuantity: 100,
		},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-moq-ok",
		UserID:                     "u-moq-ok",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceCart,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-moq", Quantity: 50, UnitPrice: 0}},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-moq-ok", "ord-moq-ok")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for MOQ-satisfying confirm, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

// G20(a): a bulk draft (created without an address and carrying no official
// compliance evidence) must declare a shipping country at confirm instead of
// silently skipping the destination-market compliance recheck. Pre-fix the
// confirm skipped the recheck whenever the country was empty and confirmed
// fine. complianceAck:true is sent so the only differentiator is the new
// country guard.
func TestCustomerConfirmOrder_EmptyCountryRequiredForBulkDraft(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-bulk": {
			ID: "p-bulk", Status: "active", BasePrice: 1.0, MOQ: 10, StockQuantity: 1000,
		},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-bulk-nocountry",
		UserID:                     "u-bulk-nocountry",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceBulk,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-bulk", Quantity: 100, UnitPrice: 0}},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: false,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrderWithBody(handler, "u-bulk-nocountry", "ord-bulk-nocountry", map[string]interface{}{
		"complianceAck": true,
	})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for empty-country bulk draft, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "target_country_required" {
		t.Fatalf("expected error=target_country_required, got %q", got)
	}
}

// G20(b): a bulk draft WITH a destination country passes the country guard and
// the compliance recheck runs against that destination (control).
func TestCustomerConfirmOrder_CountryPresentBulkDraftConfirms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-bulk": {
			ID: "p-bulk", Status: "active", BasePrice: 1.0, MOQ: 10, StockQuantity: 1000,
			Ingredients: "Sugar, Pectin", Allergens: "none",
		},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-bulk-country",
		UserID:                     "u-bulk-country",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceBulk,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-bulk", Quantity: 100, UnitPrice: 0}},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: false,
		StockReserved:              true,
		ShippingAddress: modelsOrder.Address{
			Street: "1 Main St", City: "New York", State: "NY", ZipCode: "10001", Country: "USA",
		},
		CreatedAt: now.Add(-5 * time.Minute),
		UpdatedAt: now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrderWithBody(handler, "u-bulk-country", "ord-bulk-country", map[string]interface{}{
		"complianceAck": true,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for bulk draft with a destination country, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

// G20(b): cumulative credit is enforced at confirm. The buyer already has an
// open order for 900 and this confirm totals 200 against a 1000 limit; the
// cumulative exposure (1100) exceeds the limit even though this single order is
// under it. Pre-fix the per-order-only check saw 200 <= 1000 and confirmed.
func TestCustomerConfirmOrder_CumulativeCreditLimitExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderRepo := &creditAwareOrderRepo{
		fakeOrderRepo: &fakeOrderRepo{},
		openOrders: []modelsOrder.Order{
			{ID: "other-1", UserID: "u-credit", Status: modelsOrder.OrderStatusConfirmed, TotalAmount: 900},
		},
	}
	userSvc := userService.NewUserService(&fakeUserRepoWithCompany{fakeUserRepo: &fakeUserRepo{}, companyID: "company-1"})
	handler := buildPolicyHandler(
		map[string]modelsProduct.Product{
			"p-c2": {ID: "p-c2", Status: "active", BasePrice: 2.0, MOQ: 10, StockQuantity: 10000},
		},
		userSvc,
		&modelsUser.Company{ID: "company-1", CreditLimit: 1000},
		orderService.NewOrderService(orderRepo),
	)
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-credit",
		UserID:                     "u-credit",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceCart,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-c2", Quantity: 100, UnitPrice: 0}},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-credit", "ord-credit")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for cumulative credit overrun, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "credit_limit_exceeded" {
		t.Fatalf("expected error=credit_limit_exceeded, got %q", got)
	}
}

// G20(b): cumulative credit within the limit still confirms (control). The
// buyer's open exposure (500) plus this confirm (200) stays under 1000.
func TestCustomerConfirmOrder_CumulativeCreditWithinLimitConfirms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderRepo := &creditAwareOrderRepo{
		fakeOrderRepo: &fakeOrderRepo{},
		openOrders: []modelsOrder.Order{
			{ID: "other-1", UserID: "u-credit-ok", Status: modelsOrder.OrderStatusConfirmed, TotalAmount: 500},
		},
	}
	userSvc := userService.NewUserService(&fakeUserRepoWithCompany{fakeUserRepo: &fakeUserRepo{}, companyID: "company-1"})
	handler := buildPolicyHandler(
		map[string]modelsProduct.Product{
			"p-c2": {ID: "p-c2", Status: "active", BasePrice: 2.0, MOQ: 10, StockQuantity: 10000},
		},
		userSvc,
		&modelsUser.Company{ID: "company-1", CreditLimit: 1000},
		orderService.NewOrderService(orderRepo),
	)
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-credit-ok",
		UserID:                     "u-credit-ok",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceCart,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-c2", Quantity: 100, UnitPrice: 0}},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-credit-ok", "ord-credit-ok")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for cumulative credit within limit, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder.Status != "pending" {
		t.Fatalf("expected status=pending after confirm, got %s", orderRepo.createdOrder.Status)
	}
	if got := orderRepo.createdOrder.TotalAmount; got != 200 {
		t.Fatalf("expected confirmed total 200 (100 * 2.0), got %v", got)
	}
}

// fakePriceRepo backs a PriceService for the price-list minimum-quantity test.
// FindBestPrice errors so the confirm re-price falls back to product.BasePrice —
// the test isolates the price-list min guard, not pricing. The rules feed
// MinQuantityForPriceList (edge #4, G20).
type fakePriceRepo struct {
	rules []modelsProduct.PriceRule
}

func (f *fakePriceRepo) FindAllPriceLists(ctx context.Context, page, limit int) ([]modelsProduct.PriceList, int64, error) {
	return nil, 0, nil
}
func (f *fakePriceRepo) FindPriceListByID(ctx context.Context, id string) (*modelsProduct.PriceList, error) {
	return nil, nil
}
func (f *fakePriceRepo) CreatePriceList(ctx context.Context, list *modelsProduct.PriceList) error { return nil }
func (f *fakePriceRepo) UpdatePriceList(ctx context.Context, list *modelsProduct.PriceList) error { return nil }
func (f *fakePriceRepo) DeletePriceList(ctx context.Context, id string) error                     { return nil }
func (f *fakePriceRepo) FindPriceRulesByProduct(ctx context.Context, productID string) ([]modelsProduct.PriceRule, error) {
	return f.rules, nil
}
func (f *fakePriceRepo) FindPriceRulesByPriceListID(ctx context.Context, priceListID string) ([]modelsProduct.PriceRule, error) {
	return f.rules, nil
}
func (f *fakePriceRepo) FindPriceRule(ctx context.Context, id string) (*modelsProduct.PriceRule, error) {
	return nil, nil
}
func (f *fakePriceRepo) CreatePriceRule(ctx context.Context, rule *modelsProduct.PriceRule) error { return nil }
func (f *fakePriceRepo) UpdatePriceRule(ctx context.Context, rule *modelsProduct.PriceRule) error { return nil }
func (f *fakePriceRepo) DeletePriceRule(ctx context.Context, id string) error                     { return nil }
func (f *fakePriceRepo) FindBestPrice(ctx context.Context, productID, priceListID string, quantity int) (*modelsProduct.PriceRule, error) {
	return nil, context.Canceled
}

// G20 edge #4: a bulk draft carries a quantity below its contract price-list
// minimum. The draft path never validated the price-list min at creation (only
// product MOQ), so confirm must enforce it — a below-contract line cannot be
// confirmed even when the product-level MOQ is satisfied (MOQ=0 here).
func TestCustomerConfirmOrder_BulkDraftEnforcesPriceListMin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderRepo := &fakeOrderRepo{}
	userSvc := userService.NewUserService(&fakeUserRepoWithCompany{fakeUserRepo: &fakeUserRepo{}, companyID: "company-1"})
	plID := "pl-1"
	company := &modelsUser.Company{ID: "company-1", CreditLimit: 100000, PriceListID: &plID}
	priceSvc := productService.NewPriceService(&fakePriceRepo{rules: []modelsProduct.PriceRule{
		{ProductID: "p-min", PriceListID: "pl-1", MinQuantity: 50, UnitPrice: 1.0},
	}})
	svcs := &servicesCommon.UserPortalServices{
		User:    userSvc,
		Product: productService.NewProductService(&fakeProductRepo{products: map[string]modelsProduct.Product{
			"p-min": {ID: "p-min", Status: "active", BasePrice: 1.0, MOQ: 0, StockQuantity: 1000, Ingredients: "Sugar, Pectin", Allergens: "none"},
		}}),
		Inquiry: inquiryService.NewInquiryService(&fakeInquiryRepo{}, &config.Config{}),
		Order:   orderService.NewOrderService(orderRepo),
		Price:   priceSvc,
		Company: userService.NewCompanyService(&fakeCompanyRepo{company: company}, userSvc),
	}
	handler := NewHandler(&config.Config{}, svcs, nil, nil)
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-plmin",
		UserID:                     "u-plmin",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceBulk,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-min", Quantity: 10, UnitPrice: 0}},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: false,
		StockReserved:              true,
		ShippingAddress: modelsOrder.Address{
			Street: "1 Main St", City: "New York", State: "NY", ZipCode: "10001", Country: "USA",
		},
		CreatedAt: now.Add(-5 * time.Minute),
		UpdatedAt: now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrderWithBody(handler, "u-plmin", "ord-plmin", map[string]interface{}{
		"complianceAck": true,
	})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for bulk line below price-list min, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "min_quantity_not_met" {
		t.Fatalf("expected error=min_quantity_not_met, got %q", got)
	}
	if orderRepo.createdOrder.Status != modelsOrder.OrderStatusPendingConfirm {
		t.Fatalf("expected order to stay pending_confirmation after price-list min rejection, got %s", orderRepo.createdOrder.Status)
	}
}

// G20 edge #4 control: a bulk draft at/above the contract price-list min confirms.
func TestCustomerConfirmOrder_BulkDraftMeetsPriceListMinConfirms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderRepo := &fakeOrderRepo{}
	userSvc := userService.NewUserService(&fakeUserRepoWithCompany{fakeUserRepo: &fakeUserRepo{}, companyID: "company-1"})
	plID := "pl-1"
	company := &modelsUser.Company{ID: "company-1", CreditLimit: 100000, PriceListID: &plID}
	priceSvc := productService.NewPriceService(&fakePriceRepo{rules: []modelsProduct.PriceRule{
		{ProductID: "p-min", PriceListID: "pl-1", MinQuantity: 50, UnitPrice: 1.0},
	}})
	svcs := &servicesCommon.UserPortalServices{
		User:    userSvc,
		Product: productService.NewProductService(&fakeProductRepo{products: map[string]modelsProduct.Product{
			"p-min": {ID: "p-min", Status: "active", BasePrice: 1.0, MOQ: 0, StockQuantity: 10000, Ingredients: "Sugar, Pectin", Allergens: "none"},
		}}),
		Inquiry: inquiryService.NewInquiryService(&fakeInquiryRepo{}, &config.Config{}),
		Order:   orderService.NewOrderService(orderRepo),
		Price:   priceSvc,
		Company: userService.NewCompanyService(&fakeCompanyRepo{company: company}, userSvc),
	}
	handler := NewHandler(&config.Config{}, svcs, nil, nil)
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-plmin-ok",
		UserID:                     "u-plmin-ok",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceBulk,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-min", Quantity: 60, UnitPrice: 0}},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: false,
		StockReserved:              true,
		ShippingAddress: modelsOrder.Address{
			Street: "1 Main St", City: "New York", State: "NY", ZipCode: "10001", Country: "USA",
		},
		CreatedAt: now.Add(-5 * time.Minute),
		UpdatedAt: now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrderWithBody(handler, "u-plmin-ok", "ord-plmin-ok", map[string]interface{}{
		"complianceAck": true,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for bulk line meeting price-list min, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

// G20 r3: cross-user company credit stacking is blocked. The credit limit is per
// COMPANY but the legacy sum was per USER — with a 1000 limit, buyer u-a (this
// confirm, 800) and buyer u-b (same company, existing open order 800) would each
// pass their own per-user check (open_a=0 → 800 ≤ 1000) while leaving 1600
// outstanding against the 1000 limit. The company-scoped sum aggregates u-b's
// open order into the exposure, so this confirm exceeds the limit and is
// rejected. Pre-fix (per-user sum via FindByUserID) confirmed at 200.
func TestCustomerConfirmOrder_CrossUserCompanyCreditStackingBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderRepo := &creditAwareOrderRepo{
		fakeOrderRepo: &fakeOrderRepo{},
		openOrders: []modelsOrder.Order{
			{ID: "b-open", UserID: "u-b", Status: modelsOrder.OrderStatusConfirmed, TotalAmount: 800},
		},
	}
	userSvc := userService.NewUserService(&fakeUserRepoWithCompany{fakeUserRepo: &fakeUserRepo{}, companyID: "company-1"})
	handler := buildPolicyHandler(
		map[string]modelsProduct.Product{
			"p-c2": {ID: "p-c2", Status: "active", BasePrice: 2.0, MOQ: 10, StockQuantity: 10000},
		},
		userSvc,
		&modelsUser.Company{ID: "company-1", CreditLimit: 1000},
		orderService.NewOrderService(orderRepo),
	)
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-cross",
		UserID:                     "u-a",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceCart,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-c2", Quantity: 400, UnitPrice: 0}},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-a", "ord-cross")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for cross-user company credit stacking, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "credit_limit_exceeded" {
		t.Fatalf("expected error=credit_limit_exceeded, got %q", got)
	}
}

// G20 r3: M-19 quantity override below the contract price-list min on a CART
// draft is blocked at confirm. The draft line was validated at intake at qty 100
// (>= min 50), but the confirm-time override drops it to 20 (< min 50) — a new
// quantity that never passed intake validation. Pre-fix the price-list-min
// recheck was gated to OrderSourceBulk only, so this overridden cart line
// confirmed below the contract minimum.
func TestCustomerConfirmOrder_CartM19OverrideBelowPriceListMinBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderRepo := &fakeOrderRepo{}
	userSvc := userService.NewUserService(&fakeUserRepoWithCompany{fakeUserRepo: &fakeUserRepo{}, companyID: "company-1"})
	plID := "pl-1"
	company := &modelsUser.Company{ID: "company-1", CreditLimit: 100000, PriceListID: &plID}
	priceSvc := productService.NewPriceService(&fakePriceRepo{rules: []modelsProduct.PriceRule{
		{ProductID: "p-min", PriceListID: "pl-1", MinQuantity: 50, UnitPrice: 1.0},
	}})
	svcs := &servicesCommon.UserPortalServices{
		User:    userSvc,
		Product: productService.NewProductService(&fakeProductRepo{products: map[string]modelsProduct.Product{
			"p-min": {ID: "p-min", Status: "active", BasePrice: 1.0, MOQ: 0, StockQuantity: 10000, Ingredients: "Sugar, Pectin", Allergens: "none"},
		}}),
		Inquiry: inquiryService.NewInquiryService(&fakeInquiryRepo{}, &config.Config{}),
		Order:   orderService.NewOrderService(orderRepo),
		Price:   priceSvc,
		Company: userService.NewCompanyService(&fakeCompanyRepo{company: company}, userSvc),
	}
	handler := NewHandler(&config.Config{}, svcs, nil, nil)
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-cart-m19",
		UserID:                     "u-cart-m19",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceCart,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-min", Quantity: 100, UnitPrice: 0}},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		ShippingAddress: modelsOrder.Address{
			Street: "1 Main St", City: "New York", State: "NY", ZipCode: "10001", Country: "USA",
		},
		CreatedAt: now.Add(-5 * time.Minute),
		UpdatedAt: now.Add(-5 * time.Minute),
	}

	// M-19 override drops the quantity to 20 (< contract min 50); the client
	// price is discarded server-side.
	rec := performCustomerConfirmOrderWithBody(handler, "u-cart-m19", "ord-cart-m19", map[string]interface{}{
		"items": []map[string]interface{}{
			{"productId": "p-min", "quantity": 20, "unitPrice": 0.5},
		},
	})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for cart M-19 override below price-list min, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "min_quantity_not_met" {
		t.Fatalf("expected error=min_quantity_not_met, got %q", got)
	}
	if orderRepo.createdOrder.Status != modelsOrder.OrderStatusPendingConfirm {
		t.Fatalf("expected order to stay pending_confirmation after price-list min rejection, got %s", orderRepo.createdOrder.Status)
	}
}

// G20 r3 control: a cart M-19 override at/above the contract price-list min
// confirms (the recheck must not reject legitimate overrides).
func TestCustomerConfirmOrder_CartM19OverrideMeetsPriceListMinConfirms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderRepo := &fakeOrderRepo{}
	userSvc := userService.NewUserService(&fakeUserRepoWithCompany{fakeUserRepo: &fakeUserRepo{}, companyID: "company-1"})
	plID := "pl-1"
	company := &modelsUser.Company{ID: "company-1", CreditLimit: 100000, PriceListID: &plID}
	priceSvc := productService.NewPriceService(&fakePriceRepo{rules: []modelsProduct.PriceRule{
		{ProductID: "p-min", PriceListID: "pl-1", MinQuantity: 50, UnitPrice: 1.0},
	}})
	svcs := &servicesCommon.UserPortalServices{
		User:    userSvc,
		Product: productService.NewProductService(&fakeProductRepo{products: map[string]modelsProduct.Product{
			"p-min": {ID: "p-min", Status: "active", BasePrice: 1.0, MOQ: 0, StockQuantity: 10000, Ingredients: "Sugar, Pectin", Allergens: "none"},
		}}),
		Inquiry: inquiryService.NewInquiryService(&fakeInquiryRepo{}, &config.Config{}),
		Order:   orderService.NewOrderService(orderRepo),
		Price:   priceSvc,
		Company: userService.NewCompanyService(&fakeCompanyRepo{company: company}, userSvc),
	}
	handler := NewHandler(&config.Config{}, svcs, nil, nil)
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-cart-m19-ok",
		UserID:                     "u-cart-m19-ok",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceCart,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-min", Quantity: 100, UnitPrice: 0}},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		ShippingAddress: modelsOrder.Address{
			Street: "1 Main St", City: "New York", State: "NY", ZipCode: "10001", Country: "USA",
		},
		CreatedAt: now.Add(-5 * time.Minute),
		UpdatedAt: now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrderWithBody(handler, "u-cart-m19-ok", "ord-cart-m19-ok", map[string]interface{}{
		"items": []map[string]interface{}{
			{"productId": "p-min", "quantity": 60, "unitPrice": 0.5},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for cart M-19 override meeting price-list min, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder.Status != "pending" {
		t.Fatalf("expected status=pending after confirm, got %s", orderRepo.createdOrder.Status)
	}
}

// G20 r3: an inquiry (H10) draft whose NEGOTIATED quantity sits below the product
// MOQ still confirms when the customer does NOT override the line set. The
// confirm-time MOQ recheck must not re-judge an unchanged negotiated deal against
// the catalog MOQ — a below-MOQ sample order the agent accepted is authoritative.
// Pre-fix the un-gated MOQ recheck 422'd this deal whenever the product carried a
// non-zero MOQ.
func TestCustomerConfirmOrder_InquiryBelowMOQUnchangedConfirms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestFinancialOrderHandler(map[string]modelsProduct.Product{
		"p-inq": {ID: "p-inq", Status: "active", BasePrice: 10.0, MOQ: 10, StockQuantity: 10000},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-inq-moq",
		UserID:                     "u-inq-moq",
		Status:                     modelsOrder.OrderStatusPendingConfirm,
		Source:                     modelsOrder.OrderSourceInquiry,
		Items:                      []modelsOrder.OrderItem{{ProductID: "p-inq", Quantity: 5, UnitPrice: 2.5}},
		Subtotal:                   12.5,
		TotalAmount:                12.5,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-inq-moq", "ord-inq-moq")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for below-MOQ negotiated inquiry draft (H10), got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder.Status != "pending" {
		t.Fatalf("expected status=pending after confirm, got %s", orderRepo.createdOrder.Status)
	}
	if got := orderRepo.createdOrder.TotalAmount; got != 12.5 {
		t.Fatalf("expected negotiated total 12.5 to survive confirm, got %v", got)
	}
}
