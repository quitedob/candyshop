package order

import (
	"context"
	"errors"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupCouponTestDB creates an in-memory SQLite DB with the minimal tables the
// Apply/RemoveCouponFromCart paths touch: coupons, orders, order_discounts.
// Columns mirror the modelsOrder.{Coupon,Order,OrderDiscount} fields that the
// repository reads/writes (id, user_id, status, subtotal/tax/total, usage caps,
// discount rows). MaxUsesPerUser is seeded as 0 so the per-user JOIN check in
// ValidateCoupon is skipped.
func setupCouponTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE coupons (
		id TEXT PRIMARY KEY,
		code TEXT NOT NULL UNIQUE,
		type TEXT NOT NULL,
		value REAL,
		min_order_amount REAL DEFAULT 0,
		max_uses INTEGER DEFAULT 0,
		used_count INTEGER DEFAULT 0,
		max_uses_per_user INTEGER DEFAULT 1,
		starts_at DATETIME,
		expires_at DATETIME,
		status TEXT DEFAULT 'active',
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create coupons: %v", err)
	}
	if err := db.Exec(`CREATE TABLE orders (
		id TEXT PRIMARY KEY,
		order_number TEXT NOT NULL UNIQUE,
		user_id TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		subtotal REAL,
		tax_amount REAL DEFAULT 0,
		shipping_amount REAL DEFAULT 0,
		total_amount REAL,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create orders: %v", err)
	}
	if err := db.Exec(`CREATE TABLE order_discounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id TEXT NOT NULL,
		coupon_id TEXT,
		gift_card_id TEXT,
		type TEXT NOT NULL,
		amount REAL NOT NULL,
		label TEXT
	)`).Error; err != nil {
		t.Fatalf("create order_discounts: %v", err)
	}
	return db
}

// seedCouponAndOrder seeds the SAVE10 (10% off) coupon with one use already
// consumed and a pending draft order with booked totals.
func seedCouponAndOrder(t *testing.T, db *gorm.DB) {
	t.Helper()
	now := time.Now()
	if err := db.Exec(`INSERT INTO coupons
		(id, code, type, value, max_uses, used_count, max_uses_per_user, starts_at, expires_at, status)
		VALUES ('c1', 'SAVE10', 'percentage', 10, 5, 1, 0, ?, ?, 'active')`,
		now.Add(-time.Hour), now.Add(24*time.Hour)).Error; err != nil {
		t.Fatalf("seed coupon: %v", err)
	}
	if err := db.Exec(`INSERT INTO orders
		(id, order_number, user_id, status, subtotal, tax_amount, shipping_amount, total_amount)
		VALUES ('o1', 'ON-1001', 'u1', 'pending', 100, 10, 5, 115)`).Error; err != nil {
		t.Fatalf("seed order: %v", err)
	}
}

func couponUsedCount(t *testing.T, db *gorm.DB, id string) int {
	t.Helper()
	var n int
	if err := db.Raw(`SELECT used_count FROM coupons WHERE id = ?`, id).Scan(&n).Error; err != nil {
		t.Fatalf("read coupon used_count: %v", err)
	}
	return n
}

func orderTotals(t *testing.T, db *gorm.DB, id string) (tax, total float64) {
	t.Helper()
	if err := db.Raw(`SELECT tax_amount, total_amount FROM orders WHERE id = ?`, id).Row().Scan(&tax, &total); err != nil {
		t.Fatalf("read order totals: %v", err)
	}
	return tax, total
}

// TestApplyRemoveCoupon_RoundTrip is the M3 + M7 regression test:
//   - ApplyCouponToCart increments coupon.used_count, writes an order_discount
//     row, and recomputes tax on the post-discount net (M7).
//   - RemoveCouponFromCart refunds the usage (never below 0), deletes the
//     discount row, and restores tax + total to the pre-coupon snapshot so an
//     apply→remove sequence round-trips exactly.
func TestApplyRemoveCoupon_RoundTrip(t *testing.T) {
	db := setupCouponTestDB(t)
	seedCouponAndOrder(t, db)
	ctx := context.Background()
	repo := NewCouponRepository(db)

	// Apply.
	disc, err := repo.ApplyCouponToCart(ctx, "o1", "u1", "SAVE10")
	if err != nil {
		t.Fatalf("ApplyCouponToCart: %v", err)
	}
	if disc == nil || disc.Amount != 10 {
		t.Fatalf("expected discount amount 10, got %+v", disc)
	}
	if got := couponUsedCount(t, db, "c1"); got != 2 {
		t.Fatalf("used_count after apply: expected 2, got %d", got)
	}
	var saved modelsOrder.OrderDiscount
	if err := db.Where("order_id = ? AND type = ?", "o1", "coupon").First(&saved).Error; err != nil {
		t.Fatalf("order_discount row after apply: %v", err)
	}
	if saved.Amount != 10 {
		t.Fatalf("stored discount amount: expected 10, got %v", saved.Amount)
	}
	// M7: tax booked on the post-discount net (100-10=90), proportional to base:
	// 10 * 90/100 = 9; total = 90 + 9 + 5 = 104.
	tax, total := orderTotals(t, db, "o1")
	if tax != 9 {
		t.Fatalf("tax after apply: expected 9, got %v", tax)
	}
	if total != 104 {
		t.Fatalf("total after apply: expected 104, got %v", total)
	}

	// Remove.
	if err := repo.RemoveCouponFromCart(ctx, "o1", "u1"); err != nil {
		t.Fatalf("RemoveCouponFromCart: %v", err)
	}
	if got := couponUsedCount(t, db, "c1"); got != 1 {
		t.Fatalf("used_count after remove: expected 1, got %d", got)
	}
	var count int64
	if err := db.Model(&modelsOrder.OrderDiscount{}).Where("order_id = ?", "o1").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("discount row still present after remove: %d rows", count)
	}
	// Round-trip: tax back to 10, total back to 115.
	tax, total = orderTotals(t, db, "o1")
	if tax != 10 {
		t.Fatalf("tax after remove: expected 10, got %v", tax)
	}
	if total != 115 {
		t.Fatalf("total after remove: expected 115, got %v", total)
	}
}

// TestRemoveCoupon_ForbiddenForOtherUser is the M3 IDOR regression test: a user
// who does not own the order must receive ErrCouponOrderForbidden, never the
// missing-order error (which would let a cross-tenant probe distinguish rows).
func TestRemoveCoupon_ForbiddenForOtherUser(t *testing.T) {
	db := setupCouponTestDB(t)
	seedCouponAndOrder(t, db)
	ctx := context.Background()
	repo := NewCouponRepository(db)

	err := repo.RemoveCouponFromCart(ctx, "o1", "u2")
	if !errors.Is(err, ErrCouponOrderForbidden) {
		t.Fatalf("expected ErrCouponOrderForbidden, got %v", err)
	}
}

// TestRemoveCoupon_InvalidState is the M3 status-guard regression test: once an
// order leaves the mutable draft states, removing a coupon must be refused so a
// paid/confirmed order's booked total cannot be inflated.
func TestRemoveCoupon_InvalidState(t *testing.T) {
	db := setupCouponTestDB(t)
	seedCouponAndOrder(t, db)
	ctx := context.Background()
	repo := NewCouponRepository(db)

	if err := db.Model(&modelsOrder.Order{}).Where("id = ?", "o1").
		Update("status", modelsOrder.OrderStatusConfirmed).Error; err != nil {
		t.Fatalf("move order to confirmed: %v", err)
	}

	err := repo.RemoveCouponFromCart(ctx, "o1", "u1")
	if !errors.Is(err, ErrCouponOrderInvalidState) {
		t.Fatalf("expected ErrCouponOrderInvalidState, got %v", err)
	}

	// And the reverse direction: applying to a confirmed order is also refused.
	_, err = repo.ApplyCouponToCart(ctx, "o1", "u1", "SAVE10")
	if !errors.Is(err, ErrCouponOrderInvalidState) {
		t.Fatalf("apply to confirmed order: expected ErrCouponOrderInvalidState, got %v", err)
	}
}
