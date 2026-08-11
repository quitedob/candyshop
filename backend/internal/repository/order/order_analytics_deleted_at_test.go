package order

import (
	"os"
	"strings"
	"testing"
)

// analyticsRawSQLFuncs are the OrderRepository analytics methods that execute
// hand-written SQL against the orders table (G22). GORM query-builder methods
// (RevenueByMonth, OrderCountByMonth, RevenueByDay, SumTotalAmount,
// SumTotalAmountSince) run through the model's gorm.DeletedAt and get the
// soft-delete filter appended automatically; these Raw() queries bypass that
// machinery, so each must carry o.deleted_at IS NULL explicitly or soft-deleted
// orders leak into revenue / velocity / RFM / churn / P&L / replenishment.
var analyticsRawSQLFuncs = []string{
	"TopProductsByRevenue",
	"SalesVelocity",
	"RFMAnalysis",
	"CustomerChurn",
	"InventoryHealth",
	"ProfitLossByPeriod",
	"ReplenishmentSuggestions",
}

// TestAnalyticsRawSQLFiltersSoftDeletedOrders guards the G22 fix: every
// order-scoped raw-SQL analytics query must filter o.deleted_at IS NULL.
//
// The queries use Postgres-only functions (jsonb_array_elements, date_trunc,
// NOW() - INTERVAL, EXTRACT) so they cannot execute against the SQLite
// in-memory test harness; this test reads the production SQL and asserts the
// orders-alias soft-delete filter is present in each Raw( ) block. The product
// (p.deleted_at) and user (u.deleted_at) filters are NOT sufficient — the
// orders rows in the inner subquery / join must be excluded too.
func TestAnalyticsRawSQLFiltersSoftDeletedOrders(t *testing.T) {
	src, err := os.ReadFile("order.go")
	if err != nil {
		t.Fatalf("read order.go: %v", err)
	}
	file := string(src)
	for _, fn := range analyticsRawSQLFuncs {
		fn := fn
		t.Run(fn, func(t *testing.T) {
			sql := sqlBlockForFunc(t, file, fn)
			if !strings.Contains(sql, "o.deleted_at IS NULL") {
				t.Errorf("%s raw SQL omits o.deleted_at IS NULL; soft-deleted orders would be counted.\nSQL:\n%s", fn, sql)
			}
		})
	}
}

// sqlBlockForFunc returns the SQL string literal passed to Raw( ) inside the
// named OrderRepository method, or fails the test if it cannot be located.
func sqlBlockForFunc(t *testing.T, file, funcName string) string {
	t.Helper()
	marker := "func (r *OrderRepository) " + funcName + "("
	idx := strings.Index(file, marker)
	if idx < 0 {
		t.Fatalf("func %s not found in order.go", funcName)
	}
	rest := file[idx:]
	rawIdx := strings.Index(rest, "Raw(`")
	if rawIdx < 0 {
		t.Fatalf("%s has no Raw(` call", funcName)
	}
	start := rawIdx + len("Raw(`")
	end := strings.Index(rest[start:], "`")
	if end < 0 {
		t.Fatalf("%s Raw(` block has no closing backtick", funcName)
	}
	return rest[start : start+end]
}
