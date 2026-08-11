package order

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupCreditSumTestDB 创建公司维度未偿敞口汇总测试库
func setupCreditSumTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE users (id TEXT PRIMARY KEY, company_id TEXT, deleted_at DATETIME)`,
		`CREATE TABLE orders (id TEXT PRIMARY KEY, user_id TEXT, status TEXT, total_amount REAL DEFAULT 0, deleted_at DATETIME)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

func seedCreditSumRows(t *testing.T, db *gorm.DB, rows []string) {
	t.Helper()
	for _, r := range rows {
		if err := db.Exec(r).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

// G20 r3: SumOpenOrderTotalsByCompany aggregates the open-order total across
// EVERY user of a buyer company (users.company_id → orders.user_id), not just the
// confirming user. This is what closes the cross-user credit-stacking bypass: two
// buyer users of one company each confirming up to the full per-user-sum limit.
func TestSumOpenOrderTotalsByCompany_AggregatesAcrossCompanyUsers(t *testing.T) {
	db := setupCreditSumTestDB(t)
	seedCreditSumRows(t, db, []string{
		`INSERT INTO users (id, company_id) VALUES ('u-a', 'company-1')`,
		`INSERT INTO users (id, company_id) VALUES ('u-b', 'company-1')`,
		`INSERT INTO users (id, company_id) VALUES ('u-other', 'company-2')`,
		`INSERT INTO orders (id, user_id, status, total_amount) VALUES ('o-a1', 'u-a', 'pending', 300)`,
		`INSERT INTO orders (id, user_id, status, total_amount) VALUES ('o-a2', 'u-a', 'confirmed', 200)`,
		`INSERT INTO orders (id, user_id, status, total_amount) VALUES ('o-b1', 'u-b', 'confirmed', 800)`,
		`INSERT INTO orders (id, user_id, status, total_amount) VALUES ('o-b2', 'u-b', 'cancelled', 9999)`, // terminal, excluded
		`INSERT INTO orders (id, user_id, status, total_amount) VALUES ('o-other1', 'u-other', 'confirmed', 5000)`, // other company, excluded
	})
	repo := NewOrderRepository(db)

	got, err := repo.SumOpenOrderTotalsByCompany(context.Background(), "company-1", "")
	if err != nil {
		t.Fatalf("SumOpenOrderTotalsByCompany: %v", err)
	}
	// 300 (o-a1) + 200 (o-a2) + 800 (o-b1) = 1300; cancelled o-b2 and company-2's
	// o-other1 are excluded.
	if got != 1300 {
		t.Fatalf("expected company-1 open exposure 1300, got %v", got)
	}
}

// G20 r3: the excluded order ID (the order being confirmed, whose confirmed total
// the caller adds separately) is dropped from the sum.
func TestSumOpenOrderTotalsByCompany_ExcludesConfirmingOrder(t *testing.T) {
	db := setupCreditSumTestDB(t)
	seedCreditSumRows(t, db, []string{
		`INSERT INTO users (id, company_id) VALUES ('u-a', 'company-1')`,
		`INSERT INTO orders (id, user_id, status, total_amount) VALUES ('o-keep', 'u-a', 'confirmed', 900)`,
		`INSERT INTO orders (id, user_id, status, total_amount) VALUES ('ord-confirm', 'u-a', 'pending_confirmation', 200)`,
	})
	repo := NewOrderRepository(db)

	got, err := repo.SumOpenOrderTotalsByCompany(context.Background(), "company-1", "ord-confirm")
	if err != nil {
		t.Fatalf("SumOpenOrderTotalsByCompany: %v", err)
	}
	if got != 900 {
		t.Fatalf("expected 900 with the confirming order excluded, got %v", got)
	}
}

// G20 r3: no orders for the company → sum 0 (not an error).
func TestSumOpenOrderTotalsByCompany_EmptyCompany(t *testing.T) {
	db := setupCreditSumTestDB(t)
	seedCreditSumRows(t, db, []string{
		`INSERT INTO users (id, company_id) VALUES ('u-a', 'company-1')`,
	})
	repo := NewOrderRepository(db)

	got, err := repo.SumOpenOrderTotalsByCompany(context.Background(), "company-1", "")
	if err != nil {
		t.Fatalf("SumOpenOrderTotalsByCompany: %v", err)
	}
	if got != 0 {
		t.Fatalf("expected 0 for a company with no open orders, got %v", got)
	}
}

// G20 r3: soft-deleted orders are excluded (GORM scope on the Order model).
func TestSumOpenOrderTotalsByCompany_ExcludesSoftDeleted(t *testing.T) {
	db := setupCreditSumTestDB(t)
	seedCreditSumRows(t, db, []string{
		`INSERT INTO users (id, company_id) VALUES ('u-a', 'company-1')`,
		`INSERT INTO orders (id, user_id, status, total_amount) VALUES ('o-live', 'u-a', 'confirmed', 100)`,
		`INSERT INTO orders (id, user_id, status, total_amount, deleted_at) VALUES ('o-deleted', 'u-a', 'confirmed', 9999, datetime('now'))`,
	})
	repo := NewOrderRepository(db)

	got, err := repo.SumOpenOrderTotalsByCompany(context.Background(), "company-1", "")
	if err != nil {
		t.Fatalf("SumOpenOrderTotalsByCompany: %v", err)
	}
	if got != 100 {
		t.Fatalf("expected 100 with the soft-deleted order excluded, got %v", got)
	}
}

// Compile-time guard that the concrete repository satisfies the service-side
// extension interface used by SumOpenOrderTotalsByCompany.
var _ companySummer = (*OrderRepository)(nil)

type companySummer interface {
	SumOpenOrderTotalsByCompany(ctx context.Context, companyID, excludeOrderID string) (float64, error)
}
