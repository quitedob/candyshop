package customer

// H10 regression tests for the confirm-path guard: a negotiated inquiry order's
// agreed deal-level total must survive CustomerConfirmOrder instead of being
// clobbered by the catalog / contract price list re-pricing, while NORMAL
// (non-inquiry) orders keep the existing re-price behavior.
//
// The intake half (services/orderintake) now persists the accepted offer's unit
// price on every line and marks the draft Source=inquiry. These tests drive the
// full CustomerConfirmOrder flow with an order whose draft lines carry
// offer-derived prices and assert the CONFIRMED order's persisted financials.
//
// fakeFinancialOrderRepo mirrors repository/order/order.go confirmAndReserve so
// ConfirmAndReserveOrderWithFinancials actually persists the financials the
// handler computes (the base fakeOrderRepo falls back to ConfirmAndReserveStock,
// which would drop the totals we need to assert).
//
// Billing-chain note (arguer follow-up): the confirmed order must satisfy
// TotalAmount == Subtotal + Tax + Shipping, the invariant the invoice-derivation
// path (services/order/invoice_policy.go CreateInvoiceFromOrder) relies on. A
// discounted negotiated deal whose intake draft carries a line-price Subtotal
// greater than the negotiated TotalAmount is rescaled at confirm so the persisted
// Subtotal equals the negotiated total — otherwise the auto-derived invoice would
// over-bill by exactly the discount.

import (
	"context"
	"net/http"
	"testing"
	"time"

	"candypro/api/internal/config"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	orderRepo "candypro/api/internal/repository/order"
	servicesCommon "candypro/api/internal/services/common"
	inquiryService "candypro/api/internal/services/inquiry"
	orderService "candypro/api/internal/services/order"
	productService "candypro/api/internal/services/product"
	userService "candypro/api/internal/services/user"

	"github.com/gin-gonic/gin"
)

// fakeFinancialOrderRepo extends fakeOrderRepo so the extended
// ConfirmAndReserveStockWithFinancials path is exercised and the financials the
// handler computes are persisted onto createdOrder (mirroring the real
// repository's confirmAndReserve).
type fakeFinancialOrderRepo struct {
	*fakeOrderRepo
}

func (f *fakeFinancialOrderRepo) ConfirmAndReserveStockWithFinancials(ctx context.Context, id string, stockDeltas map[string]int, confirmedAt time.Time, fin *orderRepo.OrderConfirmFinancials) error {
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

// fakeInvoiceRepo satisfies services/order invoiceRepository so the H10 invoice
// test can drive the REAL CreateInvoiceFromOrder over the confirmed order and
// prove the billing document bills the negotiated total rather than the line-sum
// subtotal.
type fakeInvoiceRepo struct {
	created *modelsOrder.Invoice
}

func (f *fakeInvoiceRepo) FindAll(ctx context.Context, page, pageSize int, status string) ([]modelsOrder.Invoice, int64, error) {
	return nil, 0, nil
}
func (f *fakeInvoiceRepo) FindByID(ctx context.Context, id string) (*modelsOrder.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) FindByOrderID(ctx context.Context, orderID string) ([]modelsOrder.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) FindByTradeID(ctx context.Context, tradeID uint) ([]modelsOrder.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) Create(ctx context.Context, invoice *modelsOrder.Invoice) error {
	f.created = invoice
	return nil
}
func (f *fakeInvoiceRepo) Update(ctx context.Context, invoice *modelsOrder.Invoice) error {
	return nil
}
func (f *fakeInvoiceRepo) Delete(ctx context.Context, id string) error {
	return nil
}
func (f *fakeInvoiceRepo) Stats(ctx context.Context) (map[string]int64, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) SumByStatus(ctx context.Context, status string) (float64, error) {
	return 0, nil
}
func (f *fakeInvoiceRepo) OverdueCount(ctx context.Context) (int64, error) {
	return 0, nil
}
func (f *fakeInvoiceRepo) OverdueTotal(ctx context.Context) (float64, error) {
	return 0, nil
}
func (f *fakeInvoiceRepo) FindByStatus(ctx context.Context, status string, page, limit int) ([]modelsOrder.Invoice, int64, error) {
	return nil, 0, nil
}
func (f *fakeInvoiceRepo) FindByUserID(ctx context.Context, userID string) ([]modelsOrder.Invoice, error) {
	return nil, nil
}

// buildTestFinancialOrderHandler wires a handler whose order repo persists the
// confirm financials, so the confirmed order's totals can be asserted.
func buildTestFinancialOrderHandler(products map[string]modelsProduct.Product) (*Handler, *fakeFinancialOrderRepo) {
	cfg := &config.Config{}
	orderRepo := &fakeFinancialOrderRepo{fakeOrderRepo: &fakeOrderRepo{}}

	svcs := &servicesCommon.UserPortalServices{
		User:    userService.NewUserService(&fakeUserRepo{}),
		Product: productService.NewProductService(&fakeProductRepo{products: products}),
		Inquiry: inquiryService.NewInquiryService(&fakeInquiryRepo{}, cfg),
		Order:   orderService.NewOrderService(orderRepo),
	}

	return NewHandler(cfg, svcs, nil, nil), orderRepo
}

// Multi-line negotiated inquiry: p1 BasePrice=10, p2 BasePrice=12, agreed
// UnitPrice=2.5/qty=100 each (total 500). Confirm must NOT re-price to the
// catalog (10+12)*100=2200 — the negotiated 500 must survive onto the confirmed
// order.
func TestCustomerConfirmOrder_NegotiatedInquiryPreservesPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestFinancialOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "active", BasePrice: 10.0},
		"p-2": {ID: "p-2", Status: "active", BasePrice: 12.0},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:     "ord-negotiated",
		UserID: "u-neg",
		Status: modelsOrder.OrderStatusPendingConfirm,
		Source: modelsOrder.OrderSourceInquiry,
		Items: []modelsOrder.OrderItem{
			{ProductID: "p-1", Quantity: 100, UnitPrice: 2.5},
			{ProductID: "p-2", Quantity: 100, UnitPrice: 2.5},
		},
		Subtotal:                   500,
		TotalAmount:                500,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-neg", "ord-negotiated")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder.Status != "pending" {
		t.Fatalf("expected status=pending after confirm, got %s", orderRepo.createdOrder.Status)
	}
	if got := orderRepo.createdOrder.TotalAmount; got != 500 {
		t.Fatalf("expected negotiated total 500 to survive confirm, got %v (catalog re-price would give 2200)", got)
	}
	if got := orderRepo.createdOrder.Subtotal; got != 500 {
		t.Fatalf("expected negotiated subtotal 500 to survive confirm, got %v", got)
	}
	if len(orderRepo.createdOrder.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(orderRepo.createdOrder.Items))
	}
	for _, it := range orderRepo.createdOrder.Items {
		if it.UnitPrice != 2.5 {
			t.Fatalf("expected negotiated UnitPrice=2.5 on every confirmed line, got %v", it.UnitPrice)
		}
	}
}

// OEM-only negotiated order: BasePrice=0 (no catalog price) but the agreed offer
// price is valid. Before the guard this hit errRepriceNoPrice → 422 no_price and
// made a validly-priced negotiated order unconfirmable.
func TestCustomerConfirmOrder_NegotiatedOEMBasePriceZeroConfirms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestFinancialOrderHandler(map[string]modelsProduct.Product{
		"p-oem": {ID: "p-oem", Status: "active", BasePrice: 0},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:     "ord-negotiated-oem",
		UserID: "u-neg-oem",
		Status: modelsOrder.OrderStatusPendingConfirm,
		Source: modelsOrder.OrderSourceInquiry,
		Items: []modelsOrder.OrderItem{
			{ProductID: "p-oem", Quantity: 100, UnitPrice: 2.5},
		},
		Subtotal:                   250,
		TotalAmount:                250,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-neg-oem", "ord-negotiated-oem")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for negotiated BasePrice=0 order, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if got := orderRepo.createdOrder.TotalAmount; got != 250 {
		t.Fatalf("expected negotiated total 250 to survive confirm, got %v", got)
	}
	if len(orderRepo.createdOrder.Items) != 1 || orderRepo.createdOrder.Items[0].UnitPrice != 2.5 {
		t.Fatalf("expected confirmed line UnitPrice=2.5, got %+v", orderRepo.createdOrder.Items)
	}
}

// A NORMAL (non-inquiry) order must keep the existing re-price behavior: a draft
// line persisted at UnitPrice=0 (bulk/requisition/reorder) is re-priced from the
// catalog at confirm, so the confirmed total reflects catalog prices, not the
// stale draft values.
func TestCustomerConfirmOrder_NormalOrderStillRepriced(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestFinancialOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "active", BasePrice: 10.0},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:     "ord-normal",
		UserID: "u-norm",
		Status: modelsOrder.OrderStatusPendingConfirm,
		Source: modelsOrder.OrderSourceCart,
		Items: []modelsOrder.OrderItem{
			{ProductID: "p-1", Quantity: 2, UnitPrice: 0}, // H1: zero-price draft
		},
		Subtotal:                   0,
		TotalAmount:                0,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-norm", "ord-normal")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if got := orderRepo.createdOrder.TotalAmount; got != 20 {
		t.Fatalf("expected normal order re-priced to total 20 (2*10), got %v", got)
	}
	if len(orderRepo.createdOrder.Items) != 1 || orderRepo.createdOrder.Items[0].UnitPrice != 10.0 {
		t.Fatalf("expected normal line re-priced to UnitPrice=10, got %+v", orderRepo.createdOrder.Items)
	}
}

// H2 preserved on the negotiated path: a client-supplied confirm unitPrice must
// never reach the order — the confirmed line keeps the server draft's agreed
// price even when the client payload echoes a different price.
func TestCustomerConfirmOrder_NegotiatedInquiryDiscardsClientPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestFinancialOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "active", BasePrice: 10.0},
		"p-2": {ID: "p-2", Status: "active", BasePrice: 12.0},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:     "ord-negotiated-clobber",
		UserID: "u-neg-clobber",
		Status: modelsOrder.OrderStatusPendingConfirm,
		Source: modelsOrder.OrderSourceInquiry,
		Items: []modelsOrder.OrderItem{
			{ProductID: "p-1", Quantity: 100, UnitPrice: 2.5},
			{ProductID: "p-2", Quantity: 100, UnitPrice: 2.5},
		},
		Subtotal:                   500,
		TotalAmount:                500,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrderWithBody(handler, "u-neg-clobber", "ord-negotiated-clobber", map[string]interface{}{
		"items": []map[string]interface{}{
			{"productId": "p-1", "quantity": 100, "unitPrice": 999},
			{"productId": "p-2", "quantity": 100, "unitPrice": 999},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if got := orderRepo.createdOrder.TotalAmount; got != 500 {
		t.Fatalf("expected negotiated total 500 to survive confirm despite client price override, got %v", got)
	}
	for _, it := range orderRepo.createdOrder.Items {
		if it.UnitPrice != 2.5 {
			t.Fatalf("expected confirmed UnitPrice=2.5 (client 999 discarded), got %v", it.UnitPrice)
		}
	}
}

// H10 follow-up: an accepted offer whose TotalAmount is a deal-level figure that
// does NOT decompose as sum(unit × qty) (round-number / discounted deal, e.g.
// unit=2.5, qty=1000, total=2000) must survive CustomerConfirmOrder as the
// confirmed total. Intake persists Subtotal=2500 (line sum) with TotalAmount=2000
// and builds the trade from order.TotalAmount (trade_from_order.go:25) → 2000.
// Confirm must preserve that contract amount rather than rebuild total =
// subtotal + tax + shipping (which would persist 2500 and disagree with the
// intake trade).
//
// Arguer follow-up: the confirmed order must ALSO keep the
// TotalAmount == Subtotal + Tax + Shipping invariant that the invoice-derivation
// path relies on, so the negotiated lines are rescaled to the negotiated total
// (2.5 → 2.0). Persisting Subtotal=2500 on a 2000 total would make
// CreateInvoiceFromOrder (Amount = Subtotal) over-bill the discounted deal by
// exactly the discount.
func TestCustomerConfirmOrder_NegotiatedInquiryDiscountedTotalSurvives(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestFinancialOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "active", BasePrice: 10.0},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:     "ord-negotiated-discount",
		UserID: "u-neg-discount",
		Status: modelsOrder.OrderStatusPendingConfirm,
		Source: modelsOrder.OrderSourceInquiry,
		Items: []modelsOrder.OrderItem{
			{ProductID: "p-1", Quantity: 1000, UnitPrice: 2.5},
		},
		Subtotal:                   2500, // sum(unit × qty) = 2.5 × 1000
		TotalAmount:                2000, // negotiated round-number / discounted total
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-neg-discount", "ord-negotiated-discount")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if got := orderRepo.createdOrder.TotalAmount; got != 2000 {
		t.Fatalf("expected negotiated total 2000 (≠ line sum 2500) to survive confirm, got %v (rebuilding total from lines would give 2500)", got)
	}
	// Invoice agreement: the confirmed order must satisfy
	// TotalAmount == Subtotal + Tax + Shipping (the invariant CreateInvoiceFromOrder
	// assumes), so the derived invoice bills the negotiated 2000 rather than the
	// 2500 line-price sum. The intake draft's line prices are rescaled to the
	// negotiated total (2.5 → 2.0) so Subtotal, the line sum and the total agree.
	if got := orderRepo.createdOrder.Subtotal; got != 2000 {
		t.Fatalf("expected subtotal to equal the negotiated total 2000 (invoice Amount=Subtotal), got %v (pre-fix kept 2500, over-billing by the discount)", got)
	}
	if got := orderRepo.createdOrder.Subtotal + orderRepo.createdOrder.TaxAmount + orderRepo.createdOrder.ShippingAmount; got != orderRepo.createdOrder.TotalAmount {
		t.Fatalf("expected TotalAmount == Subtotal+Tax+Shipping on the confirmed order, got subtotal=%v tax=%v shipping=%v total=%v", orderRepo.createdOrder.Subtotal, orderRepo.createdOrder.TaxAmount, orderRepo.createdOrder.ShippingAmount, orderRepo.createdOrder.TotalAmount)
	}
	if len(orderRepo.createdOrder.Items) != 1 {
		t.Fatalf("expected 1 confirmed item, got %d", len(orderRepo.createdOrder.Items))
	}
	it := orderRepo.createdOrder.Items[0]
	if it.UnitPrice != 2.0 {
		t.Fatalf("expected confirmed line rescaled to UnitPrice=2.0 (negotiated 2000/1000), got %v", it.UnitPrice)
	}
	if sum := it.UnitPrice * float64(it.Quantity); sum != 2000 {
		t.Fatalf("expected confirmed line sum 2000 to match the invoice line items, got %v", sum)
	}
}

// Billing-chain regression the confirm guard must not re-introduce: on a
// discounted negotiated order (Subtotal=2500 line sum, TotalAmount=2000
// negotiated), the auto-derived invoice must bill the NEGOTIATED 2000, not the
// 2500 line-price sum. CreateInvoiceFromOrder derives Amount=Subtotal and
// TotalAmount=Subtotal+Tax+Shipping (services/order/invoice_policy.go:90-102) —
// the pre-fix confirmed order kept Subtotal=2500, so the invoice over-billed by
// exactly the discount even though the order and the intake trade were correct.
// This test drives the real CreateInvoiceFromOrder over the confirmed order to
// prove the billing document agrees with the order, the payment amount
// (order.TotalAmount) and the intake trade.
func TestCustomerConfirmOrder_NegotiatedInquiryInvoiceBillsNegotiatedTotal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestFinancialOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "active", BasePrice: 10.0},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:     "ord-negotiated-discount",
		UserID: "u-neg-discount",
		Status: modelsOrder.OrderStatusPendingConfirm,
		Source: modelsOrder.OrderSourceInquiry,
		Items: []modelsOrder.OrderItem{
			{ProductID: "p-1", Quantity: 1000, UnitPrice: 2.5},
		},
		Subtotal:                   2500,
		TotalAmount:                2000,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-neg-discount", "ord-negotiated-discount")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	invoices := &fakeInvoiceRepo{}
	invSvc := orderService.NewInvoiceService(invoices, orderRepo, nil)
	inv, err := invSvc.CreateInvoiceFromOrder(context.Background(), "ord-negotiated-discount", "u-test")
	if err != nil {
		t.Fatalf("CreateInvoiceFromOrder failed: %v", err)
	}
	if got := inv.Amount; got != 2000 {
		t.Fatalf("expected invoice Amount to bill the negotiated total 2000, got %v (pre-fix billed the 2500 line sum)", got)
	}
	if got := inv.TotalAmount; got != 2000 {
		t.Fatalf("expected invoice TotalAmount 2000 (Subtotal+Tax+Shipping on a negotiated deal), got %v", got)
	}
	// The payment amount and intake trade both derive from order.TotalAmount — the
	// invoice must not disagree with them.
	if got := orderRepo.createdOrder.TotalAmount; got != 2000 {
		t.Fatalf("expected confirmed order TotalAmount 2000 (payment amount), got %v", got)
	}
}

// M-19 quantity override on an inquiry order invalidates the negotiated total:
// the confirmed total is recomputed from the preserved unit price at the new
// quantity (negotiated price kept, client price discarded) instead of freezing at
// the draft's negotiated TotalAmount. This pins the "internally consistent" guard
// on total preservation — the negotiated total is authoritative only for the
// exact negotiated line set.
func TestCustomerConfirmOrder_NegotiatedInquiryQuantityOverrideRecomputes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestFinancialOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "active", BasePrice: 10.0},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:     "ord-negotiated-qty",
		UserID: "u-neg-qty",
		Status: modelsOrder.OrderStatusPendingConfirm,
		Source: modelsOrder.OrderSourceInquiry,
		Items: []modelsOrder.OrderItem{
			{ProductID: "p-1", Quantity: 1000, UnitPrice: 2.5},
		},
		Subtotal:                   2500,
		TotalAmount:                2000,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	// Client overrides quantity to 2000 (client price 999 must be discarded).
	rec := performCustomerConfirmOrderWithBody(handler, "u-neg-qty", "ord-negotiated-qty", map[string]interface{}{
		"items": []map[string]interface{}{
			{"productId": "p-1", "quantity": 2000, "unitPrice": 999},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if got := orderRepo.createdOrder.TotalAmount; got != 5000 {
		t.Fatalf("expected recomputed total 5000 (2000 × 2.5) for overridden quantity, got %v (negotiated 2000 must NOT be frozen when quantity changes)", got)
	}
	if len(orderRepo.createdOrder.Items) != 1 || orderRepo.createdOrder.Items[0].UnitPrice != 2.5 || orderRepo.createdOrder.Items[0].Quantity != 2000 {
		t.Fatalf("expected confirmed line UnitPrice=2.5 Quantity=2000, got %+v", orderRepo.createdOrder.Items)
	}
}

// M-19 client-added line on an inquiry confirm: a product absent from the draft
// is priced server-side with the same catalog cascade as a normal re-price
// (contract list → base price → cost-stack → channel), never from the client's
// price. Here the fake user has no contract list, so p-2 resolves to its
// BasePrice=12 while the negotiated p-1 line keeps 2.5. The added line changes
// the line set, so the negotiated total is not preserved and the total is
// recomputed.
func TestCustomerConfirmOrder_NegotiatedInquiryAddedLinePriced(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, orderRepo := buildTestFinancialOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "active", BasePrice: 10.0},
		"p-2": {ID: "p-2", Status: "active", BasePrice: 12.0},
	})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:     "ord-negotiated-added",
		UserID: "u-neg-added",
		Status: modelsOrder.OrderStatusPendingConfirm,
		Source: modelsOrder.OrderSourceInquiry,
		Items: []modelsOrder.OrderItem{
			{ProductID: "p-1", Quantity: 100, UnitPrice: 2.5},
		},
		Subtotal:                   250,
		TotalAmount:                250,
		ComplianceOfficialEvidence: true,
		StockReserved:              true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	// Client adds p-2 at unitPrice=1 (discarded). Lines differ from the draft, so
	// the negotiated total is NOT preserved — the total is recomputed.
	rec := performCustomerConfirmOrderWithBody(handler, "u-neg-added", "ord-negotiated-added", map[string]interface{}{
		"items": []map[string]interface{}{
			{"productId": "p-1", "quantity": 100, "unitPrice": 2.5},
			{"productId": "p-2", "quantity": 100, "unitPrice": 1},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if got := orderRepo.createdOrder.TotalAmount; got != 250+1200 {
		t.Fatalf("expected total 1450 (p-1 100×2.5 + p-2 100×12), got %v", got)
	}
	if len(orderRepo.createdOrder.Items) != 2 {
		t.Fatalf("expected 2 confirmed items, got %d", len(orderRepo.createdOrder.Items))
	}
	for _, it := range orderRepo.createdOrder.Items {
		if it.ProductID == "p-1" && it.UnitPrice != 2.5 {
			t.Fatalf("expected p-1 UnitPrice=2.5 (negotiated), got %v", it.UnitPrice)
		}
		if it.ProductID == "p-2" && it.UnitPrice != 12.0 {
			t.Fatalf("expected p-2 UnitPrice=12 (catalog BasePrice, client 1 discarded), got %v", it.UnitPrice)
		}
	}
}
