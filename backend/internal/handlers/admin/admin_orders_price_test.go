package admin

// G20(c) regression tests: AdminUpdateOrder must reject client-supplied prices
// that are never legitimate — a negative line unit price or a negative
// subtotal/tax/shipping/total (manufactures a negative invoice). Pre-fix these
// values were written verbatim onto the order.
//
// Zero values are deliberately ALLOWED: bulk/requisition/reorder drafts are
// created unpriced (UnitPrice=0, TotalAmount=0) by design, and the admin
// order-edit modal echoes those zeros back on the next save. Rejecting them
// deadlocks the draft workflow — the customer cannot confirm an empty-country
// draft (target_country_required) and the admin cannot edit that same draft to
// add a shipping country. Draft prices are never authoritative: the customer
// confirm path re-prices every line server-side.

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"candypro/api/internal/config"
	modelsOrder "candypro/api/internal/models/order"
	orderRepo "candypro/api/internal/repository/order"
	servicesCommon "candypro/api/internal/services/common"
	orderService "candypro/api/internal/services/order"

	"github.com/gin-gonic/gin"
)

// A zero unit price is the legitimate state of an unpriced draft (bulk CSV,
// requisition, reorder) — the admin edit modal echoes it back and must not be
// blocked. G20 refutation: rejecting zero deadlocked the bulk-draft admin edit.
func TestValidateAdminOrderPrices_AllowsZeroUnitPrice(t *testing.T) {
	items := []modelsOrder.OrderItem{{ProductID: "p-1", Quantity: 10, UnitPrice: 0}}
	reason := validateAdminOrderPrices(adminUpdateOrderRequest{Items: &items})
	if reason != "" {
		t.Fatalf("expected zero unit price to be allowed as an unpriced draft, got reason %q", reason)
	}
}

func TestValidateAdminOrderPrices_RejectsNegativeUnitPrice(t *testing.T) {
	items := []modelsOrder.OrderItem{{ProductID: "p-1", Quantity: 10, UnitPrice: -5}}
	reason := validateAdminOrderPrices(adminUpdateOrderRequest{Items: &items})
	if reason == "" {
		t.Fatal("expected negative unit price to be rejected")
	}
}

// A zero total is the legitimate state of an unpriced draft (G20 refutation:
// the admin edit modal sends totalAmount 0 back on every save of such a draft).
func TestValidateAdminOrderPrices_AllowsZeroTotal(t *testing.T) {
	total := 0.0
	reason := validateAdminOrderPrices(adminUpdateOrderRequest{TotalAmount: &total})
	if reason != "" {
		t.Fatalf("expected zero totalAmount to be allowed as an unpriced draft, got reason %q", reason)
	}
}

func TestValidateAdminOrderPrices_RejectsNegativeAmounts(t *testing.T) {
	sub := -1.0
	if reason := validateAdminOrderPrices(adminUpdateOrderRequest{Subtotal: &sub}); reason == "" {
		t.Fatal("expected negative subtotal to be rejected")
	}
	tax := -0.01
	if reason := validateAdminOrderPrices(adminUpdateOrderRequest{TaxAmount: &tax}); reason == "" {
		t.Fatal("expected negative taxAmount to be rejected")
	}
	ship := -5.0
	if reason := validateAdminOrderPrices(adminUpdateOrderRequest{ShippingAmount: &ship}); reason == "" {
		t.Fatal("expected negative shippingAmount to be rejected")
	}
	total := -100.0
	if reason := validateAdminOrderPrices(adminUpdateOrderRequest{TotalAmount: &total}); reason == "" {
		t.Fatal("expected negative totalAmount to be rejected")
	}
}

func TestValidateAdminOrderPrices_IgnoresPlaceholderLines(t *testing.T) {
	// A placeholder line (quantity 0, no price) carries no sellable value and is
	// ignored — it is rejected by the quantity guards elsewhere, not by price.
	items := []modelsOrder.OrderItem{{ProductID: "p-1", Quantity: 0, UnitPrice: 0}}
	reason := validateAdminOrderPrices(adminUpdateOrderRequest{Items: &items})
	if reason != "" {
		t.Fatalf("expected placeholder line to be ignored, got reason %q", reason)
	}
}

func TestValidateAdminOrderPrices_AcceptsValidPrices(t *testing.T) {
	items := []modelsOrder.OrderItem{{ProductID: "p-1", Quantity: 10, UnitPrice: 2.5}}
	sub, tax, ship, total := 25.0, 1.0, 3.0, 29.0
	req := adminUpdateOrderRequest{
		Items:          &items,
		Subtotal:       &sub,
		TaxAmount:      &tax,
		ShippingAmount: &ship,
		TotalAmount:    &total,
	}
	if reason := validateAdminOrderPrices(req); reason != "" {
		t.Fatalf("expected valid prices to pass, got reason %q", reason)
	}
}

// fakeOrderPriceRepo is a minimal orderRepository for the AdminUpdateOrder
// rejection test. Only FindByID is reached before the price guard returns 400.
type fakeOrderPriceRepo struct {
	order *modelsOrder.Order
}

func (f *fakeOrderPriceRepo) FindAll(ctx context.Context, page, limit int, status, userID, dateFrom, dateTo string) ([]modelsOrder.Order, int64, error) {
	return nil, 0, nil
}
func (f *fakeOrderPriceRepo) FindByID(ctx context.Context, id string) (*modelsOrder.Order, error) {
	if f.order == nil || f.order.ID != id {
		return nil, context.Canceled
	}
	cp := *f.order
	return &cp, nil
}
func (f *fakeOrderPriceRepo) FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsOrder.Order, int64, error) {
	return nil, 0, nil
}
func (f *fakeOrderPriceRepo) Create(ctx context.Context, order *modelsOrder.Order) error {
	return nil
}
func (f *fakeOrderPriceRepo) CreateWithStockReservation(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return nil
}
func (f *fakeOrderPriceRepo) Update(ctx context.Context, order *modelsOrder.Order) error {
	return nil
}
func (f *fakeOrderPriceRepo) UpdateWithOutbox(ctx context.Context, order *modelsOrder.Order, outbox *modelsOrder.EventOutbox) error {
	return nil
}
func (f *fakeOrderPriceRepo) UpdateWithOptionalStockReservationAndOutbox(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int, reserve bool, outbox *modelsOrder.EventOutbox) error {
	return nil
}
func (f *fakeOrderPriceRepo) ListPendingOutbox(ctx context.Context, eventType string, limit int) ([]modelsOrder.EventOutbox, error) {
	return nil, nil
}
func (f *fakeOrderPriceRepo) IncrementOutboxAttempt(ctx context.Context, id uint) error {
	return nil
}
func (f *fakeOrderPriceRepo) UpdateOutboxResult(ctx context.Context, id uint, status, lastErr string, processedAt *time.Time) error {
	return nil
}
func (f *fakeOrderPriceRepo) FindInquiryTradeHints(ctx context.Context, inquiryID string) (incoterms, commercialNotes string, err error) {
	return "", "", nil
}
func (f *fakeOrderPriceRepo) UpdateWithStockAdjustment(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return nil
}
func (f *fakeOrderPriceRepo) Delete(ctx context.Context, id string) error {
	return nil
}
func (f *fakeOrderPriceRepo) ReleaseStockForOrder(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return nil
}
func (f *fakeOrderPriceRepo) DeleteWithStockRestore(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return nil
}
func (f *fakeOrderPriceRepo) ReserveStockForOrder(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return nil
}
func (f *fakeOrderPriceRepo) ConfirmPendingOrder(ctx context.Context, id string, confirmedAt time.Time) error {
	return nil
}
func (f *fakeOrderPriceRepo) ConfirmAndReserveStock(ctx context.Context, id string, stockDeltas map[string]int, confirmedAt time.Time) error {
	return nil
}
func (f *fakeOrderPriceRepo) ReleaseExpiredPendingConfirmationOrders(ctx context.Context, olderThan time.Time, limit int) (int, error) {
	return 0, nil
}
func (f *fakeOrderPriceRepo) CountAll(ctx context.Context) (int64, error) {
	return 0, nil
}
func (f *fakeOrderPriceRepo) CountByStatuses(ctx context.Context, statuses []string) (int64, error) {
	return 0, nil
}
func (f *fakeOrderPriceRepo) SumTotalAmount(ctx context.Context) (float64, error) {
	return 0, nil
}
func (f *fakeOrderPriceRepo) SumTotalAmountSince(ctx context.Context, since time.Time) (float64, error) {
	return 0, nil
}
func (f *fakeOrderPriceRepo) FindRecent(ctx context.Context, limit int) ([]modelsOrder.Order, error) {
	return nil, nil
}
func (f *fakeOrderPriceRepo) RevenueByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeOrderPriceRepo) OrderCountByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeOrderPriceRepo) TopProductsByRevenue(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeOrderPriceRepo) DistinctOrderingUsers(ctx context.Context, since time.Time) (int64, error) {
	return 0, nil
}
func (f *fakeOrderPriceRepo) RevenueByDay(ctx context.Context, days int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeOrderPriceRepo) SalesVelocity(ctx context.Context, months int) ([]orderRepo.SalesVelocityResult, error) {
	return nil, nil
}
func (f *fakeOrderPriceRepo) RFMAnalysis(ctx context.Context) ([]orderRepo.RFMRecord, error) {
	return nil, nil
}
func (f *fakeOrderPriceRepo) CustomerChurn(ctx context.Context, dormantDays int) ([]orderRepo.CustomerChurnResult, error) {
	return nil, nil
}
func (f *fakeOrderPriceRepo) InventoryHealth(ctx context.Context, salesWindowDays int) ([]orderRepo.InventoryHealthResult, error) {
	return nil, nil
}
func (f *fakeOrderPriceRepo) ProfitLossByPeriod(ctx context.Context, groupBy string, periods int) ([]orderRepo.ProfitLossResult, error) {
	return nil, nil
}
func (f *fakeOrderPriceRepo) ReplenishmentSuggestions(ctx context.Context, cycleDays int, salesWindowDays int) ([]orderRepo.ReplenishmentItem, error) {
	return nil, nil
}

func buildAdminOrderPriceHandler(order *modelsOrder.Order) *Handler {
	cfg := &config.Config{}
	svcs := &servicesCommon.AdminPortalServices{
		Order: orderService.NewOrderService(&fakeOrderPriceRepo{order: order}),
	}
	return NewHandler(cfg, svcs, nil, nil, nil)
}

func performAdminUpdateOrder(handler *Handler, orderID, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/admin/orders/"+orderID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: orderID}}
	c.Set("userID", "admin-1")
	handler.AdminUpdateOrder(c)
	return rec
}

// The handler returns a clear 400 for a negative unit price instead of writing
// it verbatim onto the order (pre-fix persisted the value).
func TestAdminUpdateOrder_RejectsNegativeUnitPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := buildAdminOrderPriceHandler(&modelsOrder.Order{
		ID: "ord-1", UserID: "u-1", Status: modelsOrder.OrderStatusPending,
	})
	rec := performAdminUpdateOrder(handler, "ord-1", `{"items":[{"productId":"p-1","quantity":10,"unitPrice":-5}]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative unitPrice, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

// The realistic bulk-draft admin-edit path the G20 refutation called untested:
// a bulk draft is created unpriced (items UnitPrice=0, Subtotal/TotalAmount=0,
// empty shipping address, no compliance evidence). The customer cannot confirm
// it without a destination country (target_country_required at confirm), so the
// admin unblocks it by editing the shipping address — and the admin order-edit
// modal echoes the draft's own zero prices back on the save. Pre-fix the
// zero-price guard 400'd this exact save (invalid_request), deadlocking the
// draft between an unconfirmable customer and an uneditable admin. Post-fix the
// guard allows zero and the country edit persists.
func TestAdminUpdateOrder_EditsUnpricedBulkDraft(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := buildAdminOrderPriceHandler(&modelsOrder.Order{
		ID:            "ord-draft",
		UserID:        "u-1",
		Status:        modelsOrder.OrderStatusPendingConfirm,
		Source:        modelsOrder.OrderSourceBulk,
		PaymentStatus: "unpaid",
		Currency:      "USD",
		Items:         []modelsOrder.OrderItem{{ProductID: "p-1", Quantity: 10, UnitPrice: 0}},
		Subtotal:      0,
		TotalAmount:   0,
	})
	rec := performAdminUpdateOrder(handler, "ord-draft",
		`{"items":[{"productId":"p-1","quantity":10,"unitPrice":0}],"subtotal":0,"taxAmount":0,"shippingAmount":0,"totalAmount":0,"status":"pending_confirmation","paymentStatus":"unpaid","shippingAddress":{"street":"1 Main St","city":"New York","state":"NY","zipCode":"10001","country":"USA"}}`)
	if rec.Code == http.StatusBadRequest {
		t.Fatalf("unpriced bulk-draft edit must NOT be blocked by the price guard, got 400, body=%s", rec.Body.String())
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for a valid unpriced bulk-draft edit, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "USA") {
		t.Fatalf("expected the country edit to persist on the draft, body=%s", rec.Body.String())
	}
}

// Negative totals are still rejected through the handler (finding (c) scope).
func TestAdminUpdateOrder_RejectsNegativeTotal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := buildAdminOrderPriceHandler(&modelsOrder.Order{
		ID: "ord-4", UserID: "u-1", Status: modelsOrder.OrderStatusPending,
	})
	rec := performAdminUpdateOrder(handler, "ord-4", `{"totalAmount":-5}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative totalAmount, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

// A valid price payload still passes the guard (control). The seeded order
// already carries the same financials so the update is a no-op and the handler
// never reaches the un-wired Trade/Invoice services — the only thing under test
// is that the price guard does NOT fire.
func TestAdminUpdateOrder_AcceptsValidPrices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := buildAdminOrderPriceHandler(&modelsOrder.Order{
		ID: "ord-3", UserID: "u-1", Status: modelsOrder.OrderStatusPending, PaymentStatus: "unpaid",
		Items: []modelsOrder.OrderItem{{ProductID: "p-1", Quantity: 10, UnitPrice: 2.5}},
		Subtotal: 25, TaxAmount: 1, ShippingAmount: 3, TotalAmount: 29, Currency: "USD",
	})
	rec := performAdminUpdateOrder(handler, "ord-3", `{"items":[{"productId":"p-1","quantity":10,"unitPrice":2.5}],"subtotal":25,"taxAmount":1,"shippingAmount":3,"totalAmount":29}`)
	if rec.Code == http.StatusBadRequest {
		t.Fatalf("valid prices should pass the guard, got 400, body=%s", rec.Body.String())
	}
}
