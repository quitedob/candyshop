package order

import (
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"gorm.io/gorm"
)

// TestReleaseStockForOrderLine_FEFORestoresBatch is the M1 regression test.
//
// The FEFO reserve path (deductFEFOFromBatches) permanently decrements
// product_batches.quantity and records the exact lots taken as StockTransaction
// rows carrying a BatchID. Before the fix, releasing stock only restored the
// aggregate products.stock_quantity (restoreLegacyProductStock), leaving the
// batches depleted — so a later FEFO order failed "insufficient batch stock"
// despite aggregate stock being available. The fix wires restoreFEFOBatches into
// releaseStockForOrderLine so the exact reserved batches are restored.
func TestReleaseStockForOrderLine_FEFORestoresBatch(t *testing.T) {
	db := setupStockTestDB(t)
	expiry := time.Now().UTC().Add(24 * time.Hour)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p3', 20)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO product_batches (id, product_id, batch_number, quantity, expiry_date, is_expired) VALUES ('b1', 'p3', 'LOT-1', 20, ?, 0)`, expiry).Error; err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	refID := "ord-fefo-1"
	operator := "user-fefo-1"

	// 1) Reserve via the FEFO route. Returns the stock-transaction records that
	// the production caller writes to stock_transactions via writeStockAuditEntries.
	recs, err := reserveStockForOrderLine(db, "", "p3", 5, modelsOrder.StockReasonStockReserved, refID, operator, now)
	if err != nil {
		t.Fatalf("reserveStockForOrderLine: %v", err)
	}
	if len(recs) == 0 {
		t.Fatal("expected FEFO stock transaction records")
	}
	if recs[0].BatchID == nil || *recs[0].BatchID != "b1" {
		t.Fatalf("expected batch b1 on reserve record, got %v", recs[0].BatchID)
	}

	// Persist the reserve records exactly as production does.
	if err := writeStockAuditEntries(db, recs); err != nil {
		t.Fatalf("writeStockAuditEntries (reserve): %v", err)
	}

	// Assert both batch quantity and aggregate stock dropped.
	if got := batchQuantity(t, db, "b1"); got != 15 {
		t.Fatalf("batch quantity after reserve: expected 15, got %d", got)
	}
	if got := productStock(t, db, "p3"); got != 15 {
		t.Fatalf("product stock after reserve: expected 15, got %d", got)
	}

	// 2) Release. Must restore BOTH the aggregate product stock AND the exact
	// batches that were reserved (M1), so a later FEFO order can still be fulfilled.
	if _, err := releaseStockForOrderLine(db, "", "p3", 5, modelsOrder.StockReasonStockReleased, refID, operator, now); err != nil {
		t.Fatalf("releaseStockForOrderLine: %v", err)
	}

	if got := batchQuantity(t, db, "b1"); got != 20 {
		t.Fatalf("batch quantity after release: expected 20, got %d", got)
	}
	if got := productStock(t, db, "p3"); got != 20 {
		t.Fatalf("product stock after release: expected 20, got %d", got)
	}
}

// TestReleaseStockForOrderLine_FEFOOnlyRestoresRecordedBatches verifies the
// restore is scoped to the order's own reservation rows: an order that reserved
// from a different reference must not pull stock back into the batch.
func TestReleaseStockForOrderLine_FEFOOnlyRestoresRecordedBatches(t *testing.T) {
	db := setupStockTestDB(t)
	expiry := time.Now().UTC().Add(24 * time.Hour)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p4', 20)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO product_batches (id, product_id, batch_number, quantity, expiry_date, is_expired) VALUES ('b1', 'p4', 'LOT-1', 20, ?, 0)`, expiry).Error; err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	// Order A reserves 5 (records batch b1).
	recsA, err := reserveStockForOrderLine(db, "", "p4", 5, modelsOrder.StockReasonStockReserved, "ord-a", "user-a", now)
	if err != nil {
		t.Fatalf("reserve (ord-a): %v", err)
	}
	if err := writeStockAuditEntries(db, recsA); err != nil {
		t.Fatalf("writeStockAuditEntries (ord-a): %v", err)
	}

	// Releasing order A restores only A's 5 units back to the batch.
	if _, err := releaseStockForOrderLine(db, "", "p4", 5, modelsOrder.StockReasonStockReleased, "ord-a", "user-a", now); err != nil {
		t.Fatalf("release (ord-a): %v", err)
	}
	if got := batchQuantity(t, db, "b1"); got != 20 {
		t.Fatalf("batch after releasing ord-a: expected 20, got %d", got)
	}

	// Now reserve order B and release B's line — the batch must reflect the
	// symmetric restore (reserved 5 then restored 5 => back to 20).
	recsB, err := reserveStockForOrderLine(db, "", "p4", 3, modelsOrder.StockReasonStockReserved, "ord-b", "user-b", now)
	if err != nil {
		t.Fatalf("reserve (ord-b): %v", err)
	}
	if err := writeStockAuditEntries(db, recsB); err != nil {
		t.Fatalf("writeStockAuditEntries (ord-b): %v", err)
	}
	if got := batchQuantity(t, db, "b1"); got != 17 {
		t.Fatalf("batch after reserving ord-b: expected 17, got %d", got)
	}
	if _, err := releaseStockForOrderLine(db, "", "p4", 3, modelsOrder.StockReasonStockReleased, "ord-b", "user-b", now); err != nil {
		t.Fatalf("release (ord-b): %v", err)
	}
	if got := batchQuantity(t, db, "b1"); got != 20 {
		t.Fatalf("batch after releasing ord-b: expected 20, got %d", got)
	}
}

func batchQuantity(t *testing.T, db *gorm.DB, id string) int {
	t.Helper()
	var qty int
	if err := db.Raw(`SELECT quantity FROM product_batches WHERE id = ?`, id).Scan(&qty).Error; err != nil {
		t.Fatalf("read batch %s: %v", id, err)
	}
	return qty
}

func productStock(t *testing.T, db *gorm.DB, id string) int {
	t.Helper()
	var qty int
	if err := db.Raw(`SELECT stock_quantity FROM products WHERE id = ?`, id).Scan(&qty).Error; err != nil {
		t.Fatalf("read product %s: %v", id, err)
	}
	return qty
}
