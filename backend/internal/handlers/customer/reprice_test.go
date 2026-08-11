package customer

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"

	"github.com/gin-gonic/gin"
)

// Regression tests for the H1/H2 server-side re-pricing fix:
//
//   - H1: drafts created with UnitPrice=0 (bulk / requisition / reorder) were
//     confirmed at that zero price. repriceOrderItems must recompute every line
//     from the catalog / contract price list and never persist zero.
//   - H2: confirm accepted a client-supplied unitPrice override. repriceOrderItems
//     discards client prices and re-prices every line server-side.
//
// All tests use the shared fake handler (fakeUserRepo returns a user with no
// company, fakeProductRepo holds the catalog, Price/Company/Channel are nil) so
// the contract price-list branch is skipped and pricing resolves from the base
// catalog price — exactly the "no-company user" path exercised at confirm time.

func repriceTestCtx() *gin.Context {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c
}

// H2: a client-supplied UnitPrice must never reach the order. The resolved
// UnitPrice equals the catalog BasePrice and the subtotal is qty * price.
func TestRepriceOrderItems_IgnoresClientUnitPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "active", BasePrice: 2.0},
	})

	items := []modelsOrder.OrderItem{
		{ProductID: "p-1", Quantity: 3, UnitPrice: 999},
	}
	repriced, _, subtotal, err := h.repriceOrderItems(repriceTestCtx(), "u-100", items, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repriced) != 1 {
		t.Fatalf("expected 1 repriced item, got %d", len(repriced))
	}
	if repriced[0].UnitPrice != 2.0 {
		t.Fatalf("expected UnitPrice=2.0 (catalog BasePrice), got %v (client price 999 must be discarded)", repriced[0].UnitPrice)
	}
	if repriced[0].Quantity != 3 {
		t.Fatalf("expected Quantity=3 preserved, got %d", repriced[0].Quantity)
	}
	if subtotal != 3*2.0 {
		t.Fatalf("expected subtotal=6.0, got %v", subtotal)
	}
}

// H1: a draft line persisted at UnitPrice=0 must be re-priced from the base
// catalog price for a user with no contract price list (fake user has no
// CompanyID, so the contract list branch is skipped).
func TestRepriceOrderItems_NoCompanyResolvesBasePrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "active", BasePrice: 5.0},
	})

	items := []modelsOrder.OrderItem{
		{ProductID: "p-1", Quantity: 2, UnitPrice: 0}, // H1: draft confirmed at zero
	}
	repriced, productByID, subtotal, err := h.repriceOrderItems(repriceTestCtx(), "u-100", items, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repriced) != 1 {
		t.Fatalf("expected 1 repriced item, got %d", len(repriced))
	}
	if repriced[0].UnitPrice != 5.0 {
		t.Fatalf("expected UnitPrice=5.0 (base catalog price), got %v (zero draft price must be re-priced)", repriced[0].UnitPrice)
	}
	if subtotal != 2*5.0 {
		t.Fatalf("expected subtotal=10.0, got %v", subtotal)
	}
	if _, ok := productByID["p-1"]; !ok {
		t.Fatalf("expected productByID to include p-1")
	}
}

// An inactive product must be rejected even if it carries a price.
func TestRepriceOrderItems_InactiveProductRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "inactive", BasePrice: 2.0},
	})

	items := []modelsOrder.OrderItem{
		{ProductID: "p-1", Quantity: 3, UnitPrice: 0},
	}
	_, _, _, err := h.repriceOrderItems(repriceTestCtx(), "u-100", items, "")
	if err == nil {
		t.Fatalf("expected error for inactive product, got nil")
	}
	if !errors.Is(err, errRepriceProductNotFound) {
		t.Fatalf("expected errRepriceProductNotFound, got %v", err)
	}
}

// A product with no resolvable price (BasePrice=0 and no contract list) must be
// rejected — the H1 fix must never confirm a line at zero price.
func TestRepriceOrderItems_ZeroPriceRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "active", BasePrice: 0},
	})

	items := []modelsOrder.OrderItem{
		{ProductID: "p-1", Quantity: 3, UnitPrice: 0},
	}
	_, _, _, err := h.repriceOrderItems(repriceTestCtx(), "u-100", items, "")
	if err == nil {
		t.Fatalf("expected error for zero-price product, got nil")
	}
	if !errors.Is(err, errRepriceNoPrice) {
		t.Fatalf("expected errRepriceNoPrice, got %v", err)
	}
}

// An item referencing a product that is not in the catalog must be rejected.
func TestRepriceOrderItems_ProductNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{})

	items := []modelsOrder.OrderItem{
		{ProductID: "missing", Quantity: 3, UnitPrice: 0},
	}
	_, _, _, err := h.repriceOrderItems(repriceTestCtx(), "u-100", items, "")
	if err == nil {
		t.Fatalf("expected error for missing product, got nil")
	}
	if !errors.Is(err, errRepriceProductNotFound) {
		t.Fatalf("expected errRepriceProductNotFound, got %v", err)
	}
}

// Lines with Quantity < 1 are skipped and never priced; valid lines are still
// re-priced and only they contribute to the subtotal.
func TestRepriceOrderItems_SkipsInvalidLineQuantity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-1": {ID: "p-1", Status: "active", BasePrice: 2.0},
	})

	items := []modelsOrder.OrderItem{
		{ProductID: "p-1", Quantity: 0, UnitPrice: 999}, // invalid: skipped, price untouched
		{ProductID: "p-1", Quantity: 2, UnitPrice: 999}, // valid: re-priced
	}
	repriced, _, subtotal, err := h.repriceOrderItems(repriceTestCtx(), "u-100", items, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repriced) != 2 {
		t.Fatalf("expected 2 items preserved in the returned slice, got %d", len(repriced))
	}
	if repriced[1].UnitPrice != 2.0 {
		t.Fatalf("expected valid line UnitPrice=2.0, got %v", repriced[1].UnitPrice)
	}
	if subtotal != 2*2.0 {
		t.Fatalf("expected subtotal=4.0 (only the valid line), got %v", subtotal)
	}
}
