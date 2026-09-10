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
			confirmed_at DATETIME,
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

func TestBackfillOrderCOGS_RepairsMissingCommittedCostsOnly(t *testing.T) {
	database := setupBackfillCOGSTestDB(t)
	if err := database.Exec(`INSERT INTO products (id, slug, base_price, weighted_avg_cost)
		VALUES ('costed-product', 'costed-product', 10, 4), ('uncosted-product', 'uncosted-product', 10, 0)`).Error; err != nil {
		t.Fatal(err)
	}
	confirmedAt := time.Now()
	for _, fixture := range []struct {
		name        string
		status      string
		productID   string
		confirmedAt *time.Time
		initialCOGS float64
		deletedAt   *time.Time
		wantCOGS    float64
	}{
		{"confirmed", modelsOrder.OrderStatusConfirmed, "costed-product", &confirmedAt, 0, nil, 8},
		{"production", modelsOrder.OrderStatusProduction, "costed-product", &confirmedAt, 0, nil, 8},
		{"partially-shipped", modelsOrder.OrderStatusPartiallyShipped, "costed-product", &confirmedAt, 0, nil, 8},
		{"delivered", modelsOrder.OrderStatusDelivered, "costed-product", &confirmedAt, 0, nil, 8},
		{"returned", modelsOrder.OrderStatusReturned, "costed-product", &confirmedAt, 0, nil, 8},
		{"committed-cart", modelsOrder.OrderStatusPending, "costed-product", &confirmedAt, 0, nil, 8},
		{"uncommitted-cart", modelsOrder.OrderStatusPending, "costed-product", nil, 0, nil, 0},
		{"draft", modelsOrder.OrderStatusPendingConfirm, "costed-product", nil, 0, nil, 0},
		{"approval", modelsOrder.OrderStatusPendingApproval, "costed-product", nil, 0, nil, 0},
		{"cancelled", modelsOrder.OrderStatusCancelled, "costed-product", &confirmedAt, 0, nil, 0},
		{"expired", modelsOrder.OrderStatusExpired, "costed-product", nil, 0, nil, 0},
		{"existing-cost", modelsOrder.OrderStatusConfirmed, "costed-product", &confirmedAt, 6, nil, 6},
		{"unknown-cost", modelsOrder.OrderStatusConfirmed, "uncosted-product", &confirmedAt, 0, nil, 0},
		{"missing-product", modelsOrder.OrderStatusConfirmed, "missing-product", &confirmedAt, 0, nil, 0},
		{"deleted", modelsOrder.OrderStatusConfirmed, "costed-product", &confirmedAt, 0, &confirmedAt, 0},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			itemsJSON, err := json.Marshal([]modelsOrder.OrderItem{{ProductID: fixture.productID, Quantity: 2, UnitPrice: 10}})
			if err != nil {
				t.Fatal(err)
			}
			if err := database.Exec(`INSERT INTO orders (id, order_number, user_id, status, items, subtotal, cogs,
				confirmed_at, deleted_at) VALUES (?, ?, 'buyer', ?, ?, 20, ?, ?, ?)`, fixture.name, fixture.name,
				fixture.status, itemsJSON, fixture.initialCOGS, fixture.confirmedAt, fixture.deletedAt).Error; err != nil {
				t.Fatal(err)
			}
			if err := BackfillOrderCOGS(database); err != nil {
				t.Fatal(err)
			}
			// A second startup must not reprice already repaired orders or bump
			// their version again.
			if err := BackfillOrderCOGS(database); err != nil {
				t.Fatal(err)
			}
			var persisted struct {
				COGS    float64
				Version int
			}
			if err := database.Raw(`SELECT cogs, version FROM orders WHERE id = ?`, fixture.name).Scan(&persisted).Error; err != nil {
				t.Fatal(err)
			}
			wantVersion := 0
			if fixture.wantCOGS != fixture.initialCOGS {
				wantVersion = 1
			}
			if persisted.COGS != fixture.wantCOGS || persisted.Version != wantVersion {
				t.Fatalf("backfilled snapshot = %+v, want COGS %v and version %d", persisted, fixture.wantCOGS, wantVersion)
			}
		})
	}
}

func TestBackfillOrderCOGS_RetriesOrdersWithIncompleteCostData(t *testing.T) {
	database := setupBackfillCOGSTestDB(t)
	if err := database.Exec(`INSERT INTO products (id, slug, base_price, weighted_avg_cost)
		VALUES ('known-cost', 'known-cost', 10, 4)`).Error; err != nil {
		t.Fatal(err)
	}
	itemsJSON, err := json.Marshal([]modelsOrder.OrderItem{
		{ProductID: "known-cost", Quantity: 2, UnitPrice: 10},
		{ProductID: "missing-cost", Quantity: 2, UnitPrice: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Exec(`INSERT INTO orders (id, order_number, user_id, status, items, subtotal, cogs)
		VALUES ('incomplete-order', 'INCOMPLETE-ORDER', 'buyer', ?, ?, 40, 0)`,
		modelsOrder.OrderStatusConfirmed, itemsJSON).Error; err != nil {
		t.Fatal(err)
	}
	for _, productCorrection := range []string{
		// First the product is absent, then present but still lacks its cost.
		`SELECT 1`,
		`INSERT INTO products (id, slug, base_price, weighted_avg_cost) VALUES ('missing-cost', 'missing-cost', 10, 0)`,
	} {
		if err := database.Exec(productCorrection).Error; err != nil {
			t.Fatal(err)
		}
		if err := BackfillOrderCOGS(database); err != nil {
			t.Fatal(err)
		}
		var partialCost float64
		if err := database.Raw(`SELECT cogs FROM orders WHERE id = 'incomplete-order'`).Scan(&partialCost).Error; err != nil {
			t.Fatal(err)
		}
		if partialCost != 0 {
			t.Fatalf("incomplete data must not finalize a partial cost: got %v", partialCost)
		}
	}
	if err := database.Exec(`UPDATE products SET weighted_avg_cost = 3 WHERE id = 'missing-cost'`).Error; err != nil {
		t.Fatal(err)
	}
	if err := BackfillOrderCOGS(database); err != nil {
		t.Fatal(err)
	}
	var completedCost float64
	if err := database.Raw(`SELECT cogs FROM orders WHERE id = 'incomplete-order'`).Scan(&completedCost).Error; err != nil {
		t.Fatal(err)
	}
	if completedCost != 14 {
		t.Fatalf("expected later repair to include both products (2 × 4 + 2 × 3), got %v", completedCost)
	}
}
