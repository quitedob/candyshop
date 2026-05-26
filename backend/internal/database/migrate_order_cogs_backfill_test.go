package database

import (
	"encoding/json"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupBackfillCOGSTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE products (
			id TEXT PRIMARY KEY,
			slug TEXT,
			base_price REAL DEFAULT 0,
			weighted_avg_cost REAL DEFAULT 0,
			stock_quantity INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE orders (
			id TEXT PRIMARY KEY,
			order_number TEXT UNIQUE NOT NULL,
			user_id TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			payment_status TEXT DEFAULT 'unpaid',
			stock_reserved INTEGER DEFAULT 0,
			items BLOB NOT NULL,
			subtotal REAL,
			cogs REAL DEFAULT 0,
			total_amount REAL,
			currency TEXT DEFAULT 'USD',
			created_at DATETIME,
			updated_at DATETIME,
			version INTEGER DEFAULT 0,
			deleted_at DATETIME
		)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

func TestBackfillOrderCOGS_FixesInflatedOrder(t *testing.T) {
	db := setupBackfillCOGSTestDB(t)
	now := time.Now()

	if err := db.Exec(`INSERT INTO products (id, slug, base_price, weighted_avg_cost, stock_quantity)
		VALUES ('p-gummy', '4d-fruit-gummy', 0.99, 3.40, 100),
		       ('p-crystal', 'crystal-hard-candy', 0.99, 2.04, 100),
		       ('p-belt', 'sour-belt', 0.12, 1.68, 100)`).Error; err != nil {
		t.Fatal(err)
	}

	items := []modelsOrder.OrderItem{
		{ProductID: "p-gummy", Quantity: 6000, UnitPrice: 0.99},
		{ProductID: "p-crystal", Quantity: 10000, UnitPrice: 0.99},
		{ProductID: "p-belt", Quantity: 8000, UnitPrice: 0.12},
	}
	itemsJSON, _ := json.Marshal(items)
	subtotal := 6000*0.99 + 10000*0.99 + 8000*0.12
	inflatedCOGS := 6000*3.40 + 10000*2.04 + 8000*1.68

	if err := db.Exec(`INSERT INTO orders (id, order_number, user_id, status, payment_status, stock_reserved, items, subtotal, cogs, total_amount, currency, created_at, updated_at)
		VALUES ('ord-cogs', 'ORD-COGS', 'user-1', 'confirmed', 'paid', 1, ?, ?, ?, ?, 'USD', ?, ?)`,
		itemsJSON, subtotal, inflatedCOGS, subtotal, now, now).Error; err != nil {
		t.Fatal(err)
	}

	if err := BackfillOrderCOGS(db); err != nil {
		t.Fatalf("BackfillOrderCOGS: %v", err)
	}

	var cogs float64
	if err := db.Raw(`SELECT cogs FROM orders WHERE id = 'ord-cogs'`).Scan(&cogs).Error; err != nil {
		t.Fatal(err)
	}
	if cogs >= subtotal {
		t.Fatalf("expected cogs %v < subtotal %v", cogs, subtotal)
	}
	if cogs >= inflatedCOGS {
		t.Fatalf("expected cogs reduced from %v, got %v", inflatedCOGS, cogs)
	}
	want := backfillComputeOrderCOGS(db, map[string]productCOGSCache{}, items)
	if cogs < want-0.05 || cogs > want+0.05 {
		t.Fatalf("cogs = %v, want ~%v", cogs, want)
	}
}
