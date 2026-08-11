package customer

// G20 + G24c follow-up regression tests for the ORDER-PLACEMENT paths the round-3
// arguer refuted:
//
//  1. CART CHECKOUT commits the order immediately (ConfirmedAt + stock reserved)
//     yet called only the per-order credit check — a buyer could stack two
//     under-limit cart orders (800 + 800 against a 1000 limit) past the company
//     credit limit. Cart checkout must enforce the cumulative company exposure
//     the confirm path does.
//  2. computeCheckoutPricing returned no error and its shipping/tax blocks
//     guarded err==nil && value!="", so a tax/shipping rate-lookup DB failure was
//     swallowed as "no rate" and the order booked with Tax/Shipping 0 (G24c).
//     Checkout must fail CLOSED.
//  3. The M-19 price-list-min recheck depended on the override map; a per-line
//     split (draft A qty 100 >= min 50 confirmed as A qty 30 + A qty 70) must be
//     caught per-line regardless of override detection.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"candypro/api/internal/config"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	modelsUser "candypro/api/internal/models/user"
	servicesCommon "candypro/api/internal/services/common"
	inquiryService "candypro/api/internal/services/inquiry"
	orderService "candypro/api/internal/services/order"
	productService "candypro/api/internal/services/product"
	userService "candypro/api/internal/services/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// fakeCartRepo is a minimal cartRepository for the cart-checkout tests.
type fakeCartRepo struct {
	items []modelsOrder.CartItem
}

func (f *fakeCartRepo) FindByUserID(ctx context.Context, userID string) ([]modelsOrder.CartItem, error) {
	return f.items, nil
}
func (f *fakeCartRepo) FindByID(ctx context.Context, id uint) (*modelsOrder.CartItem, error) {
	return nil, nil
}
func (f *fakeCartRepo) FindByUserIDAndProductID(ctx context.Context, userID, productID string) (*modelsOrder.CartItem, error) {
	return nil, nil
}
func (f *fakeCartRepo) Create(ctx context.Context, item *modelsOrder.CartItem) error { return nil }
func (f *fakeCartRepo) Update(ctx context.Context, item *modelsOrder.CartItem) error { return nil }
func (f *fakeCartRepo) Delete(ctx context.Context, id uint) error                   { return nil }
func (f *fakeCartRepo) ClearByUserID(ctx context.Context, userID string) error      { return nil }
func (f *fakeCartRepo) CountByUserID(ctx context.Context, userID string) (int64, error) {
	return int64(len(f.items)), nil
}
func (f *fakeCartRepo) UpsertItem(ctx context.Context, item *modelsOrder.CartItem) (*modelsOrder.CartItem, error) {
	return item, nil
}

// fakeShippingRepo drives the G24c fail-closed test: a real (non-RecordNotFound)
// error from FindBestRate must fail checkout instead of being swallowed as 0.
type fakeShippingRepo struct {
	failErr error
}

func (f *fakeShippingRepo) FindAll(ctx context.Context) ([]modelsOrder.ShippingRate, error) {
	return nil, nil
}
func (f *fakeShippingRepo) FindByID(ctx context.Context, id string) (*modelsOrder.ShippingRate, error) {
	return nil, nil
}
func (f *fakeShippingRepo) Create(ctx context.Context, rate *modelsOrder.ShippingRate) error { return nil }
func (f *fakeShippingRepo) Update(ctx context.Context, rate *modelsOrder.ShippingRate) error { return nil }
func (f *fakeShippingRepo) Delete(ctx context.Context, id string) error                     { return nil }
func (f *fakeShippingRepo) FindByDestination(ctx context.Context, destination string) ([]modelsOrder.ShippingRate, error) {
	return nil, nil
}
func (f *fakeShippingRepo) FindBestRate(ctx context.Context, destination string, weightKg float64) (*modelsOrder.ShippingRate, error) {
	if f.failErr != nil {
		return nil, f.failErr
	}
	return nil, gorm.ErrRecordNotFound // no configured rate → legitimate 0 (not an error)
}

// buildCartCheckoutHandler wires a customer handler with a cart, product, order
// and optional company / shipping service for the cart-checkout tests.
func buildCartCheckoutHandler(products map[string]modelsProduct.Product, cartItems []modelsOrder.CartItem, orderSvc *orderService.OrderService, userSvc *userService.UserService, company *modelsUser.Company, shippingSvc *orderService.ShippingService) *Handler {
	cfg := &config.Config{}
	svcs := &servicesCommon.UserPortalServices{
		User:    userSvc,
		Product: productService.NewProductService(&fakeProductRepo{products: products}),
		Inquiry: inquiryService.NewInquiryService(&fakeInquiryRepo{}, cfg),
		Cart:    orderService.NewCartService(&fakeCartRepo{items: cartItems}),
		Order:   orderSvc,
	}
	if company != nil {
		svcs.Company = userService.NewCompanyService(&fakeCompanyRepo{company: company}, userSvc)
	}
	if shippingSvc != nil {
		svcs.Shipping = shippingSvc
	}
	return NewHandler(cfg, svcs, nil, nil)
}

func performCartCheckout(handler *Handler, userID string, body map[string]interface{}) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/customer/cart/checkout", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Set("userID", userID)
	handler.CustomerCheckoutCart(c)
	return rec
}

// G20 (follow-up): cart checkout enforces the CUMULATIVE company credit limit.
// The buyer already has an $800 open order and checks out a second $800 cart
// against a $1000 company limit — the per-order check (800 <= 1000) passes but
// the cumulative exposure (1600) exceeds the limit, so checkout must be blocked
// before stock is reserved. Pre-fix the cart path called only the per-order
// check and committed the second order.
func TestCartCheckout_CumulativeCreditLimitExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderRepo := &creditAwareOrderRepo{
		fakeOrderRepo: &fakeOrderRepo{},
		openOrders: []modelsOrder.Order{
			{ID: "open-1", UserID: "u-cart-credit", Status: modelsOrder.OrderStatusConfirmed, TotalAmount: 800},
		},
	}
	userSvc := userService.NewUserService(&fakeUserRepoWithCompany{fakeUserRepo: &fakeUserRepo{}, companyID: "company-1"})
	handler := buildCartCheckoutHandler(
		map[string]modelsProduct.Product{
			"p-cart": {ID: "p-cart", Status: "active", BasePrice: 1.0, MOQ: 10, StockQuantity: 10000, Ingredients: "Sugar, Pectin", Allergens: "none"},
		},
		[]modelsOrder.CartItem{{ProductID: "p-cart", Quantity: 800, UnitPrice: 1.0, Currency: "USD"}},
		orderService.NewOrderService(orderRepo),
		userSvc,
		&modelsUser.Company{ID: "company-1", CreditLimit: 1000},
		nil,
	)

	rec := performCartCheckout(handler, "u-cart-credit", map[string]interface{}{
		"currency": "USD",
		"shippingAddress": map[string]interface{}{
			"street": "1 Main St", "city": "New York", "state": "NY", "zipCode": "10001", "country": "USA",
		},
	})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for cumulative credit overrun at cart checkout, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "credit_limit_exceeded" {
		t.Fatalf("expected error=credit_limit_exceeded, got %q", got)
	}
	if orderRepo.createdOrder != nil {
		t.Fatalf("expected NO order to be created when cumulative credit is exceeded, got %+v", orderRepo.createdOrder)
	}
}

// G20 (follow-up) control: cart checkout within the cumulative credit limit
// still commits the order (stock reserved, ConfirmedAt set).
func TestCartCheckout_CumulativeCreditWithinLimitCommits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderRepo := &creditAwareOrderRepo{
		fakeOrderRepo: &fakeOrderRepo{},
		openOrders: []modelsOrder.Order{
			{ID: "open-1", UserID: "u-cart-credit-ok", Status: modelsOrder.OrderStatusConfirmed, TotalAmount: 300},
		},
	}
	userSvc := userService.NewUserService(&fakeUserRepoWithCompany{fakeUserRepo: &fakeUserRepo{}, companyID: "company-1"})
	handler := buildCartCheckoutHandler(
		map[string]modelsProduct.Product{
			"p-cart": {ID: "p-cart", Status: "active", BasePrice: 1.0, MOQ: 10, StockQuantity: 10000, Ingredients: "Sugar, Pectin", Allergens: "none"},
		},
		[]modelsOrder.CartItem{{ProductID: "p-cart", Quantity: 500, UnitPrice: 1.0, Currency: "USD"}},
		orderService.NewOrderService(orderRepo),
		userSvc,
		&modelsUser.Company{ID: "company-1", CreditLimit: 1000},
		nil,
	)

	rec := performCartCheckout(handler, "u-cart-credit-ok", map[string]interface{}{
		"currency": "USD",
		"shippingAddress": map[string]interface{}{
			"street": "1 Main St", "city": "New York", "state": "NY", "zipCode": "10001", "country": "USA",
		},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for cumulative credit within limit, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder == nil {
		t.Fatal("expected an order to be committed")
	}
	if orderRepo.createdOrder.TotalAmount != 500 {
		t.Fatalf("expected committed total 500 (500 * 1.0), got %v", orderRepo.createdOrder.TotalAmount)
	}
}

// G24c: a shipping rate-lookup DB failure must FAIL cart checkout CLOSED instead
// of being swallowed as "no rate" and booking the order with ShippingAmount=0.
// The shipping service already returns an error for a real (non-RecordNotFound)
// DB failure; the handler must propagate it. Pre-fix computeCheckoutPricing
// guarded err==nil && value!="" and the checkout committed with 0 shipping.
func TestCartCheckout_ShippingLookupFailureFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderRepo := &fakeOrderRepo{}
	userSvc := userService.NewUserService(&fakeUserRepo{})
	shippingSvc := orderService.NewShippingService(&fakeShippingRepo{failErr: context.Canceled})
	handler := buildCartCheckoutHandler(
		map[string]modelsProduct.Product{
			"p-freight": {ID: "p-freight", Status: "active", BasePrice: 1.0, MOQ: 10, StockQuantity: 10000, GrossWeightPerCarton: 1.0, Ingredients: "Sugar, Pectin", Allergens: "none"},
		},
		[]modelsOrder.CartItem{{ProductID: "p-freight", Quantity: 100, UnitPrice: 1.0, Currency: "USD"}},
		orderService.NewOrderService(orderRepo),
		userSvc,
		nil,
		shippingSvc,
	)

	rec := performCartCheckout(handler, "u-freight", map[string]interface{}{
		"currency": "USD",
		"shippingAddress": map[string]interface{}{
			"street": "1 Main St", "city": "New York", "state": "NY", "zipCode": "10001", "country": "USA",
		},
	})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for shipping rate-lookup failure, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder != nil {
		t.Fatalf("expected NO order to be created on a shipping lookup failure, got %+v", orderRepo.createdOrder)
	}
}

// G20 (follow-up) / M-19: a draft line A qty 100 (>= price-list min 50) confirmed
// as a per-line SPLIT A qty 30 + A qty 70 keeps the product aggregate at 100 —
// MOQ (aggregate) passes, but the sub-min 30 line must be caught per-line. The
// min gate must not depend on the override map. Pre-fix (aggregate-by-product
// override detection + override-gated min check) this confirmed below the
// contract minimum.
func TestCustomerConfirmOrder_CartSplitSubMinBlocked(t *testing.T) {
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
		ID:                         "ord-split",
		UserID:                     "u-split",
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

	// M-19 split: one draft line A qty 100 becomes two override lines A qty 30 +
	// A qty 70. The 30 line is below the contract min 50.
	rec := performCustomerConfirmOrderWithBody(handler, "u-split", "ord-split", map[string]interface{}{
		"items": []map[string]interface{}{
			{"productId": "p-min", "quantity": 30, "unitPrice": 0.5},
			{"productId": "p-min", "quantity": 70, "unitPrice": 0.5},
		},
	})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for split line below price-list min, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "min_quantity_not_met" {
		t.Fatalf("expected error=min_quantity_not_met, got %q", got)
	}
	if orderRepo.createdOrder.Status != modelsOrder.OrderStatusPendingConfirm {
		t.Fatalf("expected order to stay pending_confirmation after split rejection, got %s", orderRepo.createdOrder.Status)
	}
}
