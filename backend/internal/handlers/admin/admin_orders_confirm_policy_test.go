package admin

// G20 (follow-up) regression tests for the ADMIN CONFIRM paths the round-3
// arguer refuted:
//
//  1. AdminUpdateOrder / AdminUpdateOrderStatus transition drafts to confirmed
//     with no ValidateMOQ — a below-MOQ draft could be admin-confirmed directly,
//     bypassing the customer confirm path's MOQ re-enforcement.
//  2. The admin confirm paths skipped the CUMULATIVE company credit guard — an
//     admin confirming under-limit orders directly let a company stack several
//     under-limit orders past its credit limit.
//
// The zero-price draft carve-out (G20 r2) is preserved: the checks below only
// fire on a transition TO confirmed and do not touch the unpriced-draft admin
// edit that adds a shipping country.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"candypro/api/internal/config"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	modelsUser "candypro/api/internal/models/user"
	servicesCommon "candypro/api/internal/services/common"
	orderService "candypro/api/internal/services/order"
	productService "candypro/api/internal/services/product"
	userService "candypro/api/internal/services/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// fakeAdminProductRepo is a minimal productRepository for the admin confirm-policy
// tests. Only FindByIDs is exercised (by validateAdminConfirmMOQ); everything else
// is a no-op.
type fakeAdminProductRepo struct {
	products map[string]modelsProduct.Product
}

func (f *fakeAdminProductRepo) FindAll(ctx context.Context, page, limit int, categorySlug string) ([]modelsProduct.Product, int64, error) {
	return nil, 0, nil
}
func (f *fakeAdminProductRepo) FindAllForAdmin(ctx context.Context, page, limit int, categorySlug, status, search string) ([]modelsProduct.Product, int64, error) {
	return nil, 0, nil
}
func (f *fakeAdminProductRepo) FindAllFiltered(ctx context.Context, page, limit int, halal, oemOnly, featuredOnly bool, search, sort string, minMOQ, maxMOQ int, categorySlug ...string) ([]modelsProduct.Product, int64, error) {
	return nil, 0, nil
}
func (f *fakeAdminProductRepo) FindBySlug(ctx context.Context, slug string) (*modelsProduct.Product, error) {
	return nil, nil
}
func (f *fakeAdminProductRepo) FindByID(ctx context.Context, id string) (*modelsProduct.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := p
	return &cp, nil
}
func (f *fakeAdminProductRepo) FindByIDs(ctx context.Context, ids []string) ([]modelsProduct.Product, error) {
	out := make([]modelsProduct.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := f.products[id]; ok {
			cp := p
			out = append(out, cp)
		}
	}
	return out, nil
}
func (f *fakeAdminProductRepo) FindFeatured(ctx context.Context, limit int) ([]modelsProduct.Product, error) {
	return nil, nil
}
func (f *fakeAdminProductRepo) FindRelated(ctx context.Context, slug string, limit int) ([]modelsProduct.Product, error) {
	return nil, nil
}
func (f *fakeAdminProductRepo) Create(ctx context.Context, product *modelsProduct.Product) error {
	return nil
}
func (f *fakeAdminProductRepo) Update(ctx context.Context, product *modelsProduct.Product) error {
	return nil
}
func (f *fakeAdminProductRepo) UpdateStockWithLock(ctx context.Context, productID string, newQty int) (int, error) {
	return newQty, nil
}
func (f *fakeAdminProductRepo) Delete(ctx context.Context, id string) error { return nil }
func (f *fakeAdminProductRepo) FindVariantsByProductID(ctx context.Context, productID string) ([]modelsProduct.ProductVariant, error) {
	return nil, nil
}
func (f *fakeAdminProductRepo) FindMarketProfilesForProducts(ctx context.Context, productIDs []string, marketCode string) ([]modelsProduct.ProductMarketProfile, error) {
	return nil, nil
}
func (f *fakeAdminProductRepo) UpsertProductMarketProfile(ctx context.Context, row *modelsProduct.ProductMarketProfile) error {
	return nil
}
func (f *fakeAdminProductRepo) FindMarketCostStacksForProduct(ctx context.Context, productID string) ([]modelsProduct.ProductMarketCostStack, error) {
	return nil, nil
}
func (f *fakeAdminProductRepo) UpsertProductMarketCostStack(ctx context.Context, row *modelsProduct.ProductMarketCostStack) error {
	return nil
}
func (f *fakeAdminProductRepo) ListWarehouses(ctx context.Context) ([]modelsProduct.Warehouse, error) {
	return nil, nil
}
func (f *fakeAdminProductRepo) GetDefaultWarehouseID(ctx context.Context) (string, error) {
	return "", nil
}
func (f *fakeAdminProductRepo) SaveWarehouse(ctx context.Context, w *modelsProduct.Warehouse) error {
	return nil
}
func (f *fakeAdminProductRepo) UpsertWarehouseStock(ctx context.Context, row *modelsProduct.WarehouseStock) error {
	return nil
}
func (f *fakeAdminProductRepo) SaveOEMProjectInventoryHold(ctx context.Context, row *modelsProduct.OEMProjectInventoryHold) error {
	return nil
}
func (f *fakeAdminProductRepo) ListOEMInventoryHoldsByProject(ctx context.Context, projectID string) ([]modelsProduct.OEMProjectInventoryHold, error) {
	return nil, nil
}
func (f *fakeAdminProductRepo) UpdateOEMHoldStatusIfMatches(ctx context.Context, id uint, expected, target string) (int64, error) {
	return 0, nil
}
func (f *fakeAdminProductRepo) SumActiveOEMHoldsForProduct(ctx context.Context, productID string) (int64, error) {
	return 0, nil
}
func (f *fakeAdminProductRepo) SumActiveOEMHoldsByProductIDs(ctx context.Context, productIDs []string) (map[string]int64, error) {
	return map[string]int64{}, nil
}
func (f *fakeAdminProductRepo) ListChannelInventoriesForProduct(ctx context.Context, productID string) ([]modelsProduct.ChannelInventory, error) {
	return nil, nil
}
func (f *fakeAdminProductRepo) FindChannelInventory(ctx context.Context, productID, channelCode string) (*modelsProduct.ChannelInventory, error) {
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeAdminProductRepo) FindChannelInventoriesByProductIDs(ctx context.Context, productIDs []string, channelCode string) (map[string]*modelsProduct.ChannelInventory, error) {
	return map[string]*modelsProduct.ChannelInventory{}, nil
}
func (f *fakeAdminProductRepo) UpsertChannelInventory(ctx context.Context, row *modelsProduct.ChannelInventory) error {
	return nil
}
func (f *fakeAdminProductRepo) ComputeWeightedAvgCost(ctx context.Context, productID string) float64 {
	return 0
}

// fakeAdminUserRepo returns an active user bound to a company (for the credit
// helper). The FindAll/analytics surface is a no-op.
type fakeAdminUserRepo struct {
	companyID string
}

func (f *fakeAdminUserRepo) FindByID(ctx context.Context, id string) (*modelsUser.User, error) {
	if id == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &modelsUser.User{ID: id, Status: "active", CompanyID: &f.companyID}, nil
}
func (f *fakeAdminUserRepo) FindByEmail(ctx context.Context, email string) (*modelsUser.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeAdminUserRepo) FindByResetToken(ctx context.Context, token string) (*modelsUser.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeAdminUserRepo) FindAll(ctx context.Context, page, limit int) ([]modelsUser.User, int64, error) {
	return nil, 0, nil
}
func (f *fakeAdminUserRepo) FindRecent(ctx context.Context, limit int) ([]modelsUser.User, error) {
	return nil, nil
}
func (f *fakeAdminUserRepo) Create(ctx context.Context, user *modelsUser.User) error { return nil }
func (f *fakeAdminUserRepo) Update(ctx context.Context, user *modelsUser.User) error { return nil }
func (f *fakeAdminUserRepo) Delete(ctx context.Context, id string) error             { return nil }
func (f *fakeAdminUserRepo) CountAll(ctx context.Context) (int64, error)             { return 0, nil }
func (f *fakeAdminUserRepo) CountCreatedSince(ctx context.Context, since time.Time) (int64, error) {
	return 0, nil
}
func (f *fakeAdminUserRepo) ActivateUsersByCompanyID(ctx context.Context, companyID string) error {
	return nil
}
func (f *fakeAdminUserRepo) FindByRoleNames(ctx context.Context, names []string) ([]modelsUser.User, error) {
	return nil, nil
}

// fakeAdminCompanyRepo returns a single configured company with a credit limit.
type fakeAdminCompanyRepo struct {
	company *modelsUser.Company
}

func (f *fakeAdminCompanyRepo) FindAll(ctx context.Context, page, limit int) ([]modelsUser.Company, int64, error) {
	return nil, 0, nil
}
func (f *fakeAdminCompanyRepo) FindByID(ctx context.Context, id string) (*modelsUser.Company, error) {
	if f.company == nil {
		return nil, nil
	}
	cp := *f.company
	return &cp, nil
}
func (f *fakeAdminCompanyRepo) Create(ctx context.Context, company *modelsUser.Company) error {
	return nil
}
func (f *fakeAdminCompanyRepo) Update(ctx context.Context, company *modelsUser.Company) error {
	return nil
}
func (f *fakeAdminCompanyRepo) Delete(ctx context.Context, id string) error { return nil }
func (f *fakeAdminCompanyRepo) VerifyCompany(ctx context.Context, id, status string) error {
	return nil
}

// creditAwareAdminOrderRepo embeds the price-test fake (full orderRepository)
// and adds a company-scoped open-order sum for the cumulative credit check.
type creditAwareAdminOrderRepo struct {
	*fakeOrderPriceRepo
	openOrders []modelsOrder.Order
}

func (f *creditAwareAdminOrderRepo) SumOpenOrderTotalsByCompany(ctx context.Context, companyID, excludeOrderID string) (float64, error) {
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

// buildAdminConfirmPolicyHandler wires an admin handler for the confirm-policy
// tests. A nil company means "no credit limit configured".
func buildAdminConfirmPolicyHandler(order *modelsOrder.Order, products map[string]modelsProduct.Product, userSvc *userService.UserService, company *modelsUser.Company, orderSvc *orderService.OrderService) *Handler {
	cfg := &config.Config{}
	svcs := &servicesCommon.AdminPortalServices{
		Order:   orderSvc,
		Product: productService.NewProductService(&fakeAdminProductRepo{products: products}),
		User:    userSvc,
	}
	if company != nil {
		svcs.Company = userService.NewCompanyService(&fakeAdminCompanyRepo{company: company}, userSvc)
	}
	return NewHandler(cfg, svcs, nil, nil, nil)
}

// G20 (follow-up): AdminUpdateOrder transitioning a draft to confirmed re-enforces
// per-product MOQ. A below-MOQ line must be rejected with 422 before stock is
// committed. Pre-fix the admin confirm path had no MOQ check and confirmed.
func TestAdminUpdateOrder_ConfirmRejectsBelowMOQ(t *testing.T) {
	gin.SetMode(gin.TestMode)
	order := &modelsOrder.Order{
		ID:            "ord-admin-moq",
		UserID:        "u-1",
		Status:        modelsOrder.OrderStatusPending,
		PaymentStatus: "paid",
		Currency:      "USD",
		Items:         []modelsOrder.OrderItem{{ProductID: "p-moq", Quantity: 5, UnitPrice: 2.0}},
		Subtotal:      10,
		TaxAmount:     0,
		ShippingAmount: 0,
		TotalAmount:    10,
	}
	orderSvc := orderService.NewOrderService(&fakeOrderPriceRepo{order: order})
	handler := buildAdminConfirmPolicyHandler(
		order,
		map[string]modelsProduct.Product{
			"p-moq": {ID: "p-moq", Status: "active", MOQ: 10, StockQuantity: 1000, BasePrice: 2.0},
		},
		userService.NewUserService(&fakeAdminUserRepo{companyID: "company-1"}),
		nil, // no company → no credit limit in play; the MOQ check is isolated
		orderSvc,
	)

	rec := performAdminUpdateOrder(handler, "ord-admin-moq", `{"status":"confirmed"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for admin-confirm below-MOQ line, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "min_quantity_not_met" {
		t.Fatalf("expected error=min_quantity_not_met, got %q", got)
	}
	if order.Status != modelsOrder.OrderStatusPending {
		t.Fatalf("expected order to stay pending after MOQ rejection, got %s", order.Status)
	}
}

// G20 (follow-up): AdminUpdateOrder confirming an order that would push the buyer
// company over its cumulative credit limit is rejected. The per-order check at
// draft creation cannot catch stacking; the admin confirm path must enforce the
// same company-scoped cumulative sum the customer confirm path does. Pre-fix the
// admin confirm path had no cumulative credit check.
func TestAdminUpdateOrder_ConfirmRejectsCumulativeCreditOverrun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	order := &modelsOrder.Order{
		ID:            "ord-admin-credit",
		UserID:        "u-1",
		Status:        modelsOrder.OrderStatusPending,
		PaymentStatus: "paid",
		Currency:      "USD",
		Items:         []modelsOrder.OrderItem{{ProductID: "p-ok", Quantity: 100, UnitPrice: 8.0}},
		Subtotal:      800,
		TaxAmount:     0,
		ShippingAmount: 0,
		TotalAmount:    800,
	}
	orderSvc := orderService.NewOrderService(&creditAwareAdminOrderRepo{
		fakeOrderPriceRepo: &fakeOrderPriceRepo{order: order},
		openOrders: []modelsOrder.Order{
			{ID: "open-1", UserID: "u-b", Status: modelsOrder.OrderStatusConfirmed, TotalAmount: 800},
		},
	})
	handler := buildAdminConfirmPolicyHandler(
		order,
		map[string]modelsProduct.Product{
			"p-ok": {ID: "p-ok", Status: "active", MOQ: 10, StockQuantity: 10000, BasePrice: 8.0},
		},
		userService.NewUserService(&fakeAdminUserRepo{companyID: "company-1"}),
		&modelsUser.Company{ID: "company-1", CreditLimit: 1000},
		orderSvc,
	)

	rec := performAdminUpdateOrder(handler, "ord-admin-credit", `{"status":"confirmed"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for admin-confirm cumulative credit overrun, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "credit_limit_exceeded" {
		t.Fatalf("expected error=credit_limit_exceeded, got %q", got)
	}
	if order.Status != modelsOrder.OrderStatusPending {
		t.Fatalf("expected order to stay pending after credit rejection, got %s", order.Status)
	}
}

// G20 (follow-up) control: an admin confirm that satisfies MOQ AND stays within
// the cumulative company credit limit still succeeds.
func TestAdminUpdateOrder_ConfirmPolicyMetConfirms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	order := &modelsOrder.Order{
		ID:            "ord-admin-ok",
		UserID:        "u-1",
		Status:        modelsOrder.OrderStatusPending,
		PaymentStatus: "paid",
		Currency:      "USD",
		Items:         []modelsOrder.OrderItem{{ProductID: "p-ok", Quantity: 100, UnitPrice: 8.0}},
		Subtotal:      800,
		TaxAmount:     0,
		ShippingAmount: 0,
		TotalAmount:    800,
	}
	orderSvc := orderService.NewOrderService(&creditAwareAdminOrderRepo{
		fakeOrderPriceRepo: &fakeOrderPriceRepo{order: order},
		openOrders: []modelsOrder.Order{
			{ID: "open-1", UserID: "u-b", Status: modelsOrder.OrderStatusConfirmed, TotalAmount: 100},
		},
	})
	handler := buildAdminConfirmPolicyHandler(
		order,
		map[string]modelsProduct.Product{
			"p-ok": {ID: "p-ok", Status: "active", MOQ: 10, StockQuantity: 10000, BasePrice: 8.0},
		},
		userService.NewUserService(&fakeAdminUserRepo{companyID: "company-1"}),
		&modelsUser.Company{ID: "company-1", CreditLimit: 1000},
		orderSvc,
	)

	rec := performAdminUpdateOrder(handler, "ord-admin-ok", `{"status":"confirmed"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for policy-met admin confirm, got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp.Status != modelsOrder.OrderStatusConfirmed {
		t.Fatalf("expected order to be confirmed, got %q", resp.Status)
	}
}

// G20 (follow-up): AdminUpdateOrderStatus transitioning to confirmed also
// re-enforces MOQ (the status-update endpoint is a second admin confirm path).
func TestAdminUpdateOrderStatus_ConfirmRejectsBelowMOQ(t *testing.T) {
	gin.SetMode(gin.TestMode)
	order := &modelsOrder.Order{
		ID:            "ord-status-moq",
		UserID:        "u-1",
		Status:        modelsOrder.OrderStatusPending,
		PaymentStatus: "paid",
		Currency:      "USD",
		Items:         []modelsOrder.OrderItem{{ProductID: "p-moq", Quantity: 3, UnitPrice: 2.0}},
		Subtotal:      6,
		TotalAmount:   6,
	}
	orderSvc := orderService.NewOrderService(&fakeOrderPriceRepo{order: order})
	handler := buildAdminConfirmPolicyHandler(
		order,
		map[string]modelsProduct.Product{
			"p-moq": {ID: "p-moq", Status: "active", MOQ: 10, StockQuantity: 1000, BasePrice: 2.0},
		},
		userService.NewUserService(&fakeAdminUserRepo{companyID: "company-1"}),
		nil, // no company → no credit limit in play; the MOQ check is isolated
		orderSvc,
	)

	rec := performAdminUpdateOrderStatus(handler, "ord-status-moq", `{"status":"confirmed"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for AdminUpdateOrderStatus below-MOQ confirm, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func performAdminUpdateOrderStatus(handler *Handler, orderID, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/admin/orders/"+orderID+"/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: orderID}}
	c.Set("userID", "admin-1")
	handler.AdminUpdateOrderStatus(c)
	return rec
}
