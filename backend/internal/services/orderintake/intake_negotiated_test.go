package orderintake

import (
	"context"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	orderRepo "candypro/api/internal/repository/order"
	orderSvc "candypro/api/internal/services/order"
	productSvc "candypro/api/internal/services/product"
)

// Regression tests for H10 — negotiated pricing must reach the order lines:
// the accepted offer's unit price is authoritative for EVERY matching line, not
// just single-product inquiries, and the offer-derived total stays authoritative
// on the draft order.

func TestEffectiveUnitPrice(t *testing.T) {
	offer := func(unitPrice float64) *modelsOrder.NegotiationOffer {
		return &modelsOrder.NegotiationOffer{UnitPrice: unitPrice}
	}
	tests := []struct {
		name         string
		catalogPrice float64
		offer        *modelsOrder.NegotiationOffer
		want         float64
	}{
		{name: "negotiated price wins over catalog", catalogPrice: 10.0, offer: offer(2.5), want: 2.5},
		{name: "no offer uses catalog price", catalogPrice: 10.0, offer: nil, want: 10.0},
		{name: "offer with zero unit price keeps catalog price", catalogPrice: 10.0, offer: offer(0), want: 10.0},
		{name: "no offer and no catalog price is unpriced", catalogPrice: 0, offer: nil, want: 0},
		{name: "negotiated price still wins when catalog is zero", catalogPrice: 0, offer: offer(2.5), want: 2.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := effectiveUnitPrice(tt.catalogPrice, tt.offer); got != tt.want {
				t.Fatalf("effectiveUnitPrice(%v, offer) = %v, want %v", tt.catalogPrice, got, tt.want)
			}
		})
	}
}

// --- test harness -----------------------------------------------------------

// fakeProductRepo is a minimal product.productRepository that serves a fixed
// product set and returns empty market/cost-stack/channel data so compliance,
// cross-border pricing and sellable-resolution degrade to base values (mirrors
// the nil-safety fallbacks of the real repository).
type fakeProductRepo struct {
	products []modelsProduct.Product
}

func (f *fakeProductRepo) FindByIDs(_ context.Context, ids []string) ([]modelsProduct.Product, error) {
	byID := make(map[string]modelsProduct.Product, len(f.products))
	for _, p := range f.products {
		byID[p.ID] = p
	}
	var out []modelsProduct.Product
	for _, id := range ids {
		if p, ok := byID[id]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakeProductRepo) FindByID(_ context.Context, id string) (*modelsProduct.Product, error) {
	for i := range f.products {
		if f.products[i].ID == id {
			cp := f.products[i]
			return &cp, nil
		}
	}
	return nil, context.Canceled
}

func (f *fakeProductRepo) FindMarketProfilesForProducts(_ context.Context, _ []string, _ string) ([]modelsProduct.ProductMarketProfile, error) {
	return nil, nil
}

func (f *fakeProductRepo) FindMarketCostStacksForProduct(_ context.Context, _ string) ([]modelsProduct.ProductMarketCostStack, error) {
	return nil, nil
}

func (f *fakeProductRepo) SumActiveOEMHoldsByProductIDs(_ context.Context, _ []string) (map[string]int64, error) {
	return map[string]int64{}, nil
}

func (f *fakeProductRepo) FindChannelInventoriesByProductIDs(_ context.Context, _ []string, _ string) (map[string]*modelsProduct.ChannelInventory, error) {
	return map[string]*modelsProduct.ChannelInventory{}, nil
}

func (f *fakeProductRepo) GetDefaultWarehouseID(_ context.Context) (string, error) { return "", nil }

// Remaining productRepository methods are unused by the intake path — stub.
func (f *fakeProductRepo) FindAll(context.Context, int, int, string) ([]modelsProduct.Product, int64, error) {
	return nil, 0, nil
}
func (f *fakeProductRepo) FindAllForAdmin(context.Context, int, int, string, string, string) ([]modelsProduct.Product, int64, error) {
	return nil, 0, nil
}
func (f *fakeProductRepo) FindAllFiltered(context.Context, int, int, bool, bool, bool, string, string, int, int, ...string) ([]modelsProduct.Product, int64, error) {
	return nil, 0, nil
}
func (f *fakeProductRepo) FindBySlug(context.Context, string) (*modelsProduct.Product, error) {
	return nil, context.Canceled
}
func (f *fakeProductRepo) FindFeatured(context.Context, int) ([]modelsProduct.Product, error) {
	return nil, nil
}
func (f *fakeProductRepo) FindRelated(context.Context, string, int) ([]modelsProduct.Product, error) {
	return nil, nil
}
func (f *fakeProductRepo) Create(context.Context, *modelsProduct.Product) error { return nil }
func (f *fakeProductRepo) Update(context.Context, *modelsProduct.Product) error { return nil }
func (f *fakeProductRepo) UpdateStockWithLock(context.Context, string, int) (int, error) {
	return 0, nil
}
func (f *fakeProductRepo) Delete(context.Context, string) error { return nil }
func (f *fakeProductRepo) FindVariantsByProductID(context.Context, string) ([]modelsProduct.ProductVariant, error) {
	return nil, nil
}
func (f *fakeProductRepo) UpsertProductMarketProfile(context.Context, *modelsProduct.ProductMarketProfile) error {
	return nil
}
func (f *fakeProductRepo) UpsertProductMarketCostStack(context.Context, *modelsProduct.ProductMarketCostStack) error {
	return nil
}
func (f *fakeProductRepo) ListWarehouses(context.Context) ([]modelsProduct.Warehouse, error) {
	return nil, nil
}
func (f *fakeProductRepo) SaveWarehouse(context.Context, *modelsProduct.Warehouse) error { return nil }
func (f *fakeProductRepo) UpsertWarehouseStock(context.Context, *modelsProduct.WarehouseStock) error {
	return nil
}
func (f *fakeProductRepo) SaveOEMProjectInventoryHold(context.Context, *modelsProduct.OEMProjectInventoryHold) error {
	return nil
}
func (f *fakeProductRepo) ListOEMInventoryHoldsByProject(context.Context, string) ([]modelsProduct.OEMProjectInventoryHold, error) {
	return nil, nil
}
func (f *fakeProductRepo) UpdateOEMHoldStatusIfMatches(context.Context, uint, string, string) (int64, error) {
	return 0, nil
}
func (f *fakeProductRepo) SumActiveOEMHoldsForProduct(context.Context, string) (int64, error) {
	return 0, nil
}
func (f *fakeProductRepo) ListChannelInventoriesForProduct(context.Context, string) ([]modelsProduct.ChannelInventory, error) {
	return nil, nil
}
func (f *fakeProductRepo) FindChannelInventory(context.Context, string, string) (*modelsProduct.ChannelInventory, error) {
	return nil, nil
}
func (f *fakeProductRepo) UpsertChannelInventory(context.Context, *modelsProduct.ChannelInventory) error {
	return nil
}
func (f *fakeProductRepo) ComputeWeightedAvgCost(context.Context, string) float64 { return 0 }

// fakePriceRepo is a minimal product.priceRepository with no configured price
// rules, so catalog resolution falls back to the product BasePrice.
type fakePriceRepo struct{}

func (f *fakePriceRepo) FindBestPrice(context.Context, string, string, int) (*modelsProduct.PriceRule, error) {
	return nil, context.Canceled
}
func (f *fakePriceRepo) FindAllPriceLists(context.Context, int, int) ([]modelsProduct.PriceList, int64, error) {
	return nil, 0, nil
}
func (f *fakePriceRepo) FindPriceListByID(context.Context, string) (*modelsProduct.PriceList, error) {
	return nil, context.Canceled
}
func (f *fakePriceRepo) CreatePriceList(context.Context, *modelsProduct.PriceList) error { return nil }
func (f *fakePriceRepo) UpdatePriceList(context.Context, *modelsProduct.PriceList) error { return nil }
func (f *fakePriceRepo) DeletePriceList(context.Context, string) error                   { return nil }
func (f *fakePriceRepo) FindPriceRulesByProduct(context.Context, string) ([]modelsProduct.PriceRule, error) {
	return nil, nil
}
func (f *fakePriceRepo) FindPriceRulesByPriceListID(context.Context, string) ([]modelsProduct.PriceRule, error) {
	return nil, nil
}
func (f *fakePriceRepo) FindPriceRule(context.Context, string) (*modelsProduct.PriceRule, error) {
	return nil, context.Canceled
}
func (f *fakePriceRepo) CreatePriceRule(context.Context, *modelsProduct.PriceRule) error { return nil }
func (f *fakePriceRepo) UpdatePriceRule(context.Context, *modelsProduct.PriceRule) error { return nil }
func (f *fakePriceRepo) DeletePriceRule(context.Context, string) error                   { return nil }

// fakeOrderRepo is a minimal order.orderRepository that captures created orders
// for assertions; everything else is stubbed out.
type fakeOrderRepo struct {
	created []*modelsOrder.Order
}

func (f *fakeOrderRepo) Create(_ context.Context, order *modelsOrder.Order) error {
	f.created = append(f.created, order)
	return nil
}

func (f *fakeOrderRepo) FindAll(context.Context, int, int, string, string, string, string) ([]modelsOrder.Order, int64, error) {
	return nil, 0, nil
}
func (f *fakeOrderRepo) FindByID(context.Context, string) (*modelsOrder.Order, error) {
	return nil, context.Canceled
}
func (f *fakeOrderRepo) FindByUserID(context.Context, string, int, int) ([]modelsOrder.Order, int64, error) {
	return nil, 0, nil
}
func (f *fakeOrderRepo) CreateWithStockReservation(context.Context, *modelsOrder.Order, map[string]int) error {
	return nil
}
func (f *fakeOrderRepo) Update(context.Context, *modelsOrder.Order) error { return nil }
func (f *fakeOrderRepo) UpdateWithOutbox(context.Context, *modelsOrder.Order, *modelsOrder.EventOutbox) error {
	return nil
}
func (f *fakeOrderRepo) UpdateWithOptionalStockReservationAndOutbox(context.Context, *modelsOrder.Order, map[string]int, bool, *modelsOrder.EventOutbox) error {
	return nil
}
func (f *fakeOrderRepo) ListPendingOutbox(context.Context, string, int) ([]modelsOrder.EventOutbox, error) {
	return nil, nil
}
func (f *fakeOrderRepo) IncrementOutboxAttempt(context.Context, uint) error { return nil }
func (f *fakeOrderRepo) UpdateOutboxResult(context.Context, uint, string, string, *time.Time) error {
	return nil
}
func (f *fakeOrderRepo) FindInquiryTradeHints(context.Context, string) (string, string, error) {
	return "", "", nil
}
func (f *fakeOrderRepo) UpdateWithStockAdjustment(context.Context, *modelsOrder.Order, map[string]int) error {
	return nil
}
func (f *fakeOrderRepo) Delete(context.Context, string) error { return nil }
func (f *fakeOrderRepo) ReleaseStockForOrder(context.Context, *modelsOrder.Order, map[string]int) error {
	return nil
}
func (f *fakeOrderRepo) DeleteWithStockRestore(context.Context, *modelsOrder.Order, map[string]int) error {
	return nil
}
func (f *fakeOrderRepo) ReserveStockForOrder(context.Context, *modelsOrder.Order, map[string]int) error {
	return nil
}
func (f *fakeOrderRepo) ConfirmPendingOrder(context.Context, string, time.Time) error { return nil }
func (f *fakeOrderRepo) ConfirmAndReserveStock(context.Context, string, map[string]int, time.Time) error {
	return nil
}
func (f *fakeOrderRepo) ReleaseExpiredPendingConfirmationOrders(context.Context, time.Time, int) (int, error) {
	return 0, nil
}
func (f *fakeOrderRepo) CountAll(context.Context) (int64, error)                  { return 0, nil }
func (f *fakeOrderRepo) CountByStatuses(context.Context, []string) (int64, error) { return 0, nil }
func (f *fakeOrderRepo) SumTotalAmount(context.Context) (float64, error)          { return 0, nil }
func (f *fakeOrderRepo) SumTotalAmountSince(context.Context, time.Time) (float64, error) {
	return 0, nil
}
func (f *fakeOrderRepo) FindRecent(context.Context, int) ([]modelsOrder.Order, error) {
	return nil, nil
}
func (f *fakeOrderRepo) RevenueByMonth(context.Context, int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeOrderRepo) OrderCountByMonth(context.Context, int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeOrderRepo) TopProductsByRevenue(context.Context, int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeOrderRepo) DistinctOrderingUsers(context.Context, time.Time) (int64, error) {
	return 0, nil
}
func (f *fakeOrderRepo) RevenueByDay(context.Context, int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeOrderRepo) SalesVelocity(context.Context, int) ([]orderRepo.SalesVelocityResult, error) {
	return nil, nil
}
func (f *fakeOrderRepo) RFMAnalysis(context.Context) ([]orderRepo.RFMRecord, error) { return nil, nil }
func (f *fakeOrderRepo) CustomerChurn(context.Context, int) ([]orderRepo.CustomerChurnResult, error) {
	return nil, nil
}
func (f *fakeOrderRepo) InventoryHealth(context.Context, int) ([]orderRepo.InventoryHealthResult, error) {
	return nil, nil
}
func (f *fakeOrderRepo) ProfitLossByPeriod(context.Context, string, int) ([]orderRepo.ProfitLossResult, error) {
	return nil, nil
}
func (f *fakeOrderRepo) ReplenishmentSuggestions(context.Context, int, int) ([]orderRepo.ReplenishmentItem, error) {
	return nil, nil
}

// --- integration tests ------------------------------------------------------

func newTestIntakeService(products []modelsProduct.Product) (*Service, *fakeOrderRepo) {
	prodService := productSvc.NewProductService(&fakeProductRepo{products: products})
	priceService := productSvc.NewPriceService(&fakePriceRepo{})
	ordRepo := &fakeOrderRepo{}
	ordService := orderSvc.NewOrderService(ordRepo)
	// trade + inquiry services are intentionally nil — trade linking is skipped.
	return NewService(prodService, priceService, ordService, nil, nil), ordRepo
}

func newTestProduct(id string, basePrice float64) modelsProduct.Product {
	return modelsProduct.Product{
		ID:            id,
		Name:          "product-" + id,
		Status:        modelsProduct.ProductStatusActive,
		BasePrice:     basePrice,
		StockQuantity: 1000,
		Ingredients:   "sugar, corn syrup",
		Allergens:     "none",
	}
}

func newTestInquiry(userID string, productIDs []string) *modelsProduct.Inquiry {
	return &modelsProduct.Inquiry{
		ID:                "inq-h10",
		UserID:            &userID,
		TargetCountry:     "US",
		EstimatedQuantity: "100 units",
		Products:          productIDs,
	}
}

func TestCreateOrderFromAcceptedOffer_MultiLineNegotiatedPricing(t *testing.T) {
	products := []modelsProduct.Product{
		newTestProduct("p1", 10.0),
		newTestProduct("p2", 12.0),
	}
	svc, orderRepo := newTestIntakeService(products)

	userID := "u-h10"
	offer := &modelsOrder.NegotiationOffer{
		ID:          "offer-h10",
		UnitPrice:   2.5,
		Quantity:    100,
		TotalAmount: 500.0, // 2.5 * 100 * 2 lines
		Currency:    "USD",
	}
	res, err := svc.CreateOrderFromAcceptedOffer(context.Background(), newTestInquiry(userID, []string{"p1", "p2"}), offer, Options{})
	if err != nil {
		t.Fatalf("CreateOrderFromAcceptedOffer error: %v", err)
	}
	if res == nil || res.Order == nil {
		t.Fatal("expected an order to be created")
	}
	if len(orderRepo.created) != 1 {
		t.Fatalf("expected 1 created order, got %d", len(orderRepo.created))
	}
	order := orderRepo.created[0]
	if len(order.Items) != 2 {
		t.Fatalf("expected 2 order lines, got %d", len(order.Items))
	}
	// H10: the negotiated unit price must reach EVERY line, not just single-line orders.
	for i, item := range order.Items {
		if item.UnitPrice != 2.5 {
			t.Fatalf("item[%d] UnitPrice = %v, want negotiated 2.5 (multi-line must not fall back to catalog)", i, item.UnitPrice)
		}
	}
	if order.TotalAmount != 500.0 {
		t.Fatalf("TotalAmount = %v, want offer-derived total 500.0", order.TotalAmount)
	}
	if order.Subtotal != 500.0 {
		t.Fatalf("Subtotal = %v, want 500.0 (2.5*100 per line)", order.Subtotal)
	}
	if order.Source != modelsOrder.OrderSourceInquiry {
		t.Fatalf("Source = %q, want %q", order.Source, modelsOrder.OrderSourceInquiry)
	}
}

func TestCreateOrderFromAcceptedOffer_SingleLineNegotiatedPricing(t *testing.T) {
	products := []modelsProduct.Product{newTestProduct("p1", 10.0)}
	svc, orderRepo := newTestIntakeService(products)

	userID := "u-h10-single"
	offer := &modelsOrder.NegotiationOffer{
		ID:          "offer-h10-single",
		UnitPrice:   2.5,
		Quantity:    100,
		TotalAmount: 250.0,
		Currency:    "USD",
	}
	res, err := svc.CreateOrderFromAcceptedOffer(context.Background(), newTestInquiry(userID, []string{"p1"}), offer, Options{})
	if err != nil {
		t.Fatalf("CreateOrderFromAcceptedOffer error: %v", err)
	}
	if res == nil || res.Order == nil {
		t.Fatal("expected an order to be created")
	}
	order := orderRepo.created[0]
	if len(order.Items) != 1 {
		t.Fatalf("expected 1 order line, got %d", len(order.Items))
	}
	if order.Items[0].UnitPrice != 2.5 {
		t.Fatalf("UnitPrice = %v, want negotiated 2.5", order.Items[0].UnitPrice)
	}
	if order.TotalAmount != 250.0 {
		t.Fatalf("TotalAmount = %v, want offer-derived total 250.0", order.TotalAmount)
	}
}

func TestCreateOrderFromAcceptedOffer_NoOfferUsesCatalogPrice(t *testing.T) {
	products := []modelsProduct.Product{
		newTestProduct("p1", 10.0),
		newTestProduct("p2", 12.0),
	}
	svc, orderRepo := newTestIntakeService(products)

	userID := "u-h10-nooffer"
	res, err := svc.CreateOrderFromAcceptedOffer(context.Background(), newTestInquiry(userID, []string{"p1", "p2"}), nil, Options{})
	if err != nil {
		t.Fatalf("CreateOrderFromAcceptedOffer error: %v", err)
	}
	if res == nil || res.Order == nil {
		t.Fatal("expected an order to be created")
	}
	order := orderRepo.created[0]
	if len(order.Items) != 2 {
		t.Fatalf("expected 2 order lines, got %d", len(order.Items))
	}
	// Without an offer the catalog BasePrice must stand.
	if order.Items[0].UnitPrice != 10.0 || order.Items[1].UnitPrice != 12.0 {
		t.Fatalf("catalog prices not preserved: got %v, %v (want 10.0, 12.0)", order.Items[0].UnitPrice, order.Items[1].UnitPrice)
	}
}
