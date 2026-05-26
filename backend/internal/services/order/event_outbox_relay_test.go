package order

import (
	"context"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	orderRepo "candypro/api/internal/repository/order"
)

type atomicUpdateFakeRepo struct {
	updateCalled                         bool
	atomicCalled                         bool
	atomicReserve                        bool
	lastOrder                            *modelsOrder.Order
	UpdateWithOptionalStockReservationFn func(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int, reserve bool, outbox *modelsOrder.EventOutbox) error
}

func (f *atomicUpdateFakeRepo) FindAll(ctx context.Context, page, limit int, status, userID, dateFrom, dateTo string) ([]modelsOrder.Order, int64, error) {
	return nil, 0, nil
}
func (f *atomicUpdateFakeRepo) FindByID(ctx context.Context, id string) (*modelsOrder.Order, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsOrder.Order, int64, error) {
	return nil, 0, nil
}
func (f *atomicUpdateFakeRepo) Create(ctx context.Context, order *modelsOrder.Order) error { return nil }
func (f *atomicUpdateFakeRepo) CreateWithStockReservation(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return nil
}
func (f *atomicUpdateFakeRepo) Update(ctx context.Context, order *modelsOrder.Order) error {
	f.updateCalled = true
	f.lastOrder = order
	return nil
}
func (f *atomicUpdateFakeRepo) UpdateWithOutbox(ctx context.Context, order *modelsOrder.Order, outbox *modelsOrder.EventOutbox) error {
	return nil
}
func (f *atomicUpdateFakeRepo) UpdateWithOptionalStockReservationAndOutbox(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int, reserve bool, outbox *modelsOrder.EventOutbox) error {
	f.atomicCalled = true
	f.atomicReserve = reserve
	f.lastOrder = order
	if f.UpdateWithOptionalStockReservationFn != nil {
		return f.UpdateWithOptionalStockReservationFn(ctx, order, stockDeltas, reserve, outbox)
	}
	order.StockReserved = reserve
	return nil
}
func (f *atomicUpdateFakeRepo) ListPendingOutbox(ctx context.Context, eventType string, limit int) ([]modelsOrder.EventOutbox, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) IncrementOutboxAttempt(ctx context.Context, id uint) error { return nil }
func (f *atomicUpdateFakeRepo) UpdateOutboxResult(ctx context.Context, id uint, status, lastErr string, processedAt *time.Time) error {
	return nil
}
func (f *atomicUpdateFakeRepo) FindInquiryTradeHints(ctx context.Context, inquiryID string) (string, string, error) {
	return "", "", nil
}
func (f *atomicUpdateFakeRepo) UpdateWithStockAdjustment(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return nil
}
func (f *atomicUpdateFakeRepo) Delete(ctx context.Context, id string) error { return nil }
func (f *atomicUpdateFakeRepo) ReleaseStockForOrder(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return nil
}
func (f *atomicUpdateFakeRepo) DeleteWithStockRestore(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return nil
}
func (f *atomicUpdateFakeRepo) ReserveStockForOrder(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return nil
}
func (f *atomicUpdateFakeRepo) ConfirmPendingOrder(ctx context.Context, id string, confirmedAt time.Time) error {
	return nil
}
func (f *atomicUpdateFakeRepo) ConfirmAndReserveStock(ctx context.Context, id string, stockDeltas map[string]int, confirmedAt time.Time) error {
	return nil
}
func (f *atomicUpdateFakeRepo) ReleaseExpiredPendingConfirmationOrders(ctx context.Context, olderThan time.Time, limit int) (int, error) {
	return 0, nil
}
func (f *atomicUpdateFakeRepo) CountAll(ctx context.Context) (int64, error) { return 0, nil }
func (f *atomicUpdateFakeRepo) CountByStatuses(ctx context.Context, statuses []string) (int64, error) {
	return 0, nil
}
func (f *atomicUpdateFakeRepo) SumTotalAmount(ctx context.Context) (float64, error) { return 0, nil }
func (f *atomicUpdateFakeRepo) SumTotalAmountSince(ctx context.Context, since time.Time) (float64, error) {
	return 0, nil
}
func (f *atomicUpdateFakeRepo) FindRecent(ctx context.Context, limit int) ([]modelsOrder.Order, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) RevenueByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) OrderCountByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) TopProductsByRevenue(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) DistinctOrderingUsers(ctx context.Context, since time.Time) (int64, error) {
	return 0, nil
}
func (f *atomicUpdateFakeRepo) RevenueByDay(ctx context.Context, days int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) SalesVelocity(ctx context.Context, months int) ([]orderRepo.SalesVelocityResult, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) RFMAnalysis(ctx context.Context) ([]orderRepo.RFMRecord, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) CustomerChurn(ctx context.Context, dormantDays int) ([]orderRepo.CustomerChurnResult, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) InventoryHealth(ctx context.Context, salesWindowDays int) ([]orderRepo.InventoryHealthResult, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) ProfitLossByPeriod(ctx context.Context, groupBy string, periods int) ([]orderRepo.ProfitLossResult, error) {
	return nil, nil
}
func (f *atomicUpdateFakeRepo) ReplenishmentSuggestions(ctx context.Context, cycleDays, salesWindowDays int) ([]orderRepo.ReplenishmentItem, error) {
	return nil, nil
}

func TestUpdateOrderForAdmin_UsesAtomicReserveOnConfirm(t *testing.T) {
	repo := &atomicUpdateFakeRepo{}
	svc := &OrderService{repo: repo}
	order := &modelsOrder.Order{
		ID:            "ord-1",
		Status:        "confirmed",
		PaymentStatus: "paid",
		Items:         modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 5}},
	}

	if err := svc.UpdateOrderForAdmin(context.Background(), order, "pending", "confirmed"); err != nil {
		t.Fatalf("UpdateOrderForAdmin: %v", err)
	}
	if !repo.atomicCalled {
		t.Fatal("expected atomic reserve+update path")
	}
	if !repo.atomicReserve {
		t.Fatal("expected reserve=true")
	}
	if repo.updateCalled {
		t.Fatal("expected atomic path, not plain Update")
	}
}

func TestUpdateOrderForAdmin_PaymentOnlyWhenStockReserved(t *testing.T) {
	repo := &atomicUpdateFakeRepo{}
	svc := &OrderService{repo: repo}
	order := &modelsOrder.Order{
		ID:            "ord-2",
		Status:        "confirmed",
		PaymentStatus: "paid",
		StockReserved: true,
		Items:         modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 5}},
	}

	if err := svc.UpdateOrderForAdmin(context.Background(), order, "confirmed", "confirmed"); err != nil {
		t.Fatalf("UpdateOrderForAdmin: %v", err)
	}
	if repo.atomicCalled {
		t.Fatal("expected plain Update when stock already reserved and status unchanged")
	}
	if !repo.updateCalled {
		t.Fatal("expected Update to be called")
	}
}
