package order

import (
	"context"
	"errors"
	"math"
	"os"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestUpdateStatusGuarded_InvalidatesStaleOrderVersion(t *testing.T) {
	database := setupOrderUpdateTestDB(t)
	repository := NewOrderRepository(database)
	order := &modelsOrder.Order{
		ID: "versioned-order", OrderNumber: "VERSIONED-ORDER", UserID: "buyer",
		Status: modelsOrder.OrderStatusProduction, Version: 4,
		Items: modelsOrder.OrderItemArray{},
	}
	if err := repository.Create(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	shippedAt := time.Now()
	if err := repository.UpdateStatusGuarded(context.Background(), order.ID,
		modelsOrder.OrderStatusProduction, modelsOrder.OrderStatusShipped,
		map[string]interface{}{"shipped_at": shippedAt, "version": order.Version}); err != nil {
		t.Fatal(err)
	}
	if err := repository.UpdateWithVersionGuard(context.Background(), order,
		map[string]interface{}{"status": modelsOrder.OrderStatusProduction}); !errors.Is(err, ErrOptimisticLockConflict) {
		t.Fatalf("stale edit error = %v, want optimistic lock conflict", err)
	}
	if err := repository.UpdateStatusGuarded(context.Background(), order.ID,
		modelsOrder.OrderStatusProduction, modelsOrder.OrderStatusShipped, nil); !errors.Is(err, ErrOrderStateMismatch) {
		t.Fatalf("replayed transition error = %v, want state mismatch", err)
	}
	var persisted modelsOrder.Order
	if err := database.First(&persisted, "id = ?", order.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.Version != order.Version+1 || persisted.Status != modelsOrder.OrderStatusShipped || persisted.ShippedAt == nil {
		t.Fatalf("guarded transition did not retain exactly one version increment and shipment: %+v", persisted)
	}
}

func TestOrderRevenueTotals_ExcludeCancelledExpiredAndDeleted(t *testing.T) {
	database := setupOrderUpdateTestDB(t)
	repository := NewOrderRepository(database)
	since := time.Now().Add(-time.Hour)
	for _, fixture := range []struct {
		id        string
		status    string
		amount    float64
		createdAt time.Time
		deletedAt *time.Time
	}{
		{"active", modelsOrder.OrderStatusConfirmed, 100, since.Add(time.Minute), nil},
		{"older", modelsOrder.OrderStatusDelivered, 50, since.Add(-time.Hour), nil},
		{"cancelled", modelsOrder.OrderStatusCancelled, 1000, since.Add(time.Minute), nil},
		{"expired", modelsOrder.OrderStatusExpired, 2000, since.Add(time.Minute), nil},
		{"deleted", modelsOrder.OrderStatusDelivered, 3000, since.Add(time.Minute), &since},
	} {
		if err := database.Exec(`INSERT INTO orders (id, order_number, user_id, status, items, total_amount, created_at, deleted_at)
			VALUES (?, ?, 'buyer', ?, '[]', ?, ?, ?)`, fixture.id, fixture.id, fixture.status,
			fixture.amount, fixture.createdAt, fixture.deletedAt).Error; err != nil {
			t.Fatal(err)
		}
	}
	amount, err := repository.SumTotalAmount(context.Background())
	if err != nil || amount != 150 {
		t.Fatalf("all-time gross order value = %v, error = %v; want 150", amount, err)
	}
	amount, err = repository.SumTotalAmountSince(context.Background(), since)
	if err != nil || amount != 100 {
		t.Fatalf("recent gross order value = %v, error = %v; want 100", amount, err)
	}
}

// ORDER_ANALYTICS_TEST_DSN must identify a disposable PostgreSQL database.
// Temporary tables keep every fixture private to the test transaction. Running
// this against PostgreSQL exercises interval/date_trunc parsing and bound values,
// which SQLite and assertions on SQL strings cannot validate.
func TestProfitLossByPeriod_Postgres(t *testing.T) {
	databaseDSN := os.Getenv("ORDER_ANALYTICS_TEST_DSN")
	if databaseDSN == "" {
		t.Skip("set ORDER_ANALYTICS_TEST_DSN to run PostgreSQL financial integration tests")
	}
	database, err := gorm.Open(postgres.Open(databaseDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open financial test database failed")
	}
	connection, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	transaction := database.Begin()
	if transaction.Error != nil {
		t.Fatal(transaction.Error)
	}
	t.Cleanup(func() { transaction.Rollback() })
	for _, statement := range []string{
		`CREATE TEMP TABLE orders (id TEXT PRIMARY KEY, status TEXT, total_amount NUMERIC,
			tax_amount NUMERIC, shipping_amount NUMERIC, cogs NUMERIC, created_at TIMESTAMPTZ, deleted_at TIMESTAMPTZ) ON COMMIT DROP`,
		`CREATE TEMP TABLE return_requests (id TEXT PRIMARY KEY, order_id TEXT, status TEXT) ON COMMIT DROP`,
		`CREATE TEMP TABLE return_items (return_id TEXT, refund_amount NUMERIC) ON COMMIT DROP`,
	} {
		if err := transaction.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	repository := NewOrderRepository(transaction)
	rows, err := repository.ProfitLossByPeriod(context.Background(), "month", 6)
	if err != nil || len(rows) != 0 {
		t.Fatalf("empty report = %+v, error = %v", rows, err)
	}
	for _, fixture := range []struct {
		id       string
		status   string
		total    float64
		tax      float64
		shipping float64
		cost     float64
	}{
		{"partial-refund", modelsOrder.OrderStatusDelivered, 132, 12, 20, 40},
		{"full-refund", modelsOrder.OrderStatusReturned, 55, 5, 0, 20},
		{"uncompleted-returns", modelsOrder.OrderStatusConfirmed, 110, 10, 0, 40},
		{"cancelled", modelsOrder.OrderStatusCancelled, 1000, 0, 0, 0},
		{"expired", modelsOrder.OrderStatusExpired, 2000, 0, 0, 0},
	} {
		if err := transaction.Exec(`INSERT INTO orders VALUES (?, ?, ?, ?, ?, ?, NOW(), NULL)`,
			fixture.id, fixture.status, fixture.total, fixture.tax, fixture.shipping, fixture.cost).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, fixture := range []struct{ id, orderID, status string }{
		{"refund-one", "partial-refund", modelsOrder.ReturnStatusRefunded},
		{"refund-two", "partial-refund", modelsOrder.ReturnStatusRefunded},
		{"refund-full", "full-refund", modelsOrder.ReturnStatusRefunded},
		{"pending", "uncompleted-returns", modelsOrder.ReturnStatusPending},
		{"approved", "uncompleted-returns", modelsOrder.ReturnStatusApproved},
		{"received", "uncompleted-returns", modelsOrder.ReturnStatusReceived},
		{"rejected", "uncompleted-returns", modelsOrder.ReturnStatusRejected},
	} {
		if err := transaction.Exec(`INSERT INTO return_requests VALUES (?, ?, ?)`, fixture.id, fixture.orderID, fixture.status).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, statement := range []string{
		`INSERT INTO return_items VALUES ('refund-one', 10), ('refund-one', 15), ('refund-two', 5), ('refund-full', 50),
			('pending', 100), ('approved', 100), ('received', 100), ('rejected', 100)`,
		`INSERT INTO orders VALUES ('deleted', 'delivered', 4000, 0, 0, 0, NOW(), NOW()),
			('old', 'delivered', 10, 0, 0, 0, NOW() - INTERVAL '2 years', NULL)`,
	} {
		if err := transaction.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, groupBy := range []string{"day", "week", "month", "invalid-defaults-to-month"} {
		t.Run(groupBy, func(t *testing.T) {
			rows, err := repository.ProfitLossByPeriod(context.Background(), groupBy, 6)
			if err != nil || len(rows) != 1 {
				t.Fatalf("report rows = %+v, error = %v; want one current period", rows, err)
			}
			period := rows[0]
			// Merchandise refunds total 80. Net revenue is 297 - 27 tax - 80 refunds.
			// Costs remain 100 COGS + 20 shipping; multiple refund lines must not
			// multiply either the cost snapshots or the three included orders.
			if period.Revenue != 190 || period.COGS != 100 || period.ShippingCost != 20 ||
				period.GrossProfit != 70 || period.OrderCount != 3 || math.Abs(period.GrossMarginPct-36.8) > 0.01 {
				t.Fatalf("incorrect P&L after tax/refunds: %+v", period)
			}
		})
	}
	monthly, err := repository.RevenueByMonth(context.Background(), 6)
	if err != nil || len(monthly) != 1 || monthly[0]["revenue"] != float64(297) || monthly[0]["orderCount"] != int64(3) {
		t.Fatalf("monthly gross order value = %+v, error = %v", monthly, err)
	}
	daily, err := repository.RevenueByDay(context.Background(), 7)
	if err != nil || len(daily) != 1 || daily[0]["revenue"] != float64(297) {
		t.Fatalf("daily gross order value = %+v, error = %v", daily, err)
	}
	counts, err := repository.OrderCountByMonth(context.Background(), 6)
	if err != nil || len(counts) != 1 || counts[0]["count"] != int64(3) {
		t.Fatalf("monthly order count = %+v, error = %v", counts, err)
	}
	if err := transaction.Exec(`DELETE FROM orders WHERE id <> 'full-refund'`).Error; err != nil {
		t.Fatal(err)
	}
	rows, err = repository.ProfitLossByPeriod(context.Background(), "month", 6)
	if err != nil || len(rows) != 1 || rows[0].Revenue != 0 || rows[0].GrossProfit != -20 || rows[0].GrossMarginPct != 0 {
		t.Fatalf("fully refunded period should retain cost and avoid division by zero: %+v, error = %v", rows, err)
	}
}
