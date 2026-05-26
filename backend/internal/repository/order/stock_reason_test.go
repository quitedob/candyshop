package order_test

import (
	"testing"

	modelsOrder "candypro/api/internal/models/order"
)

// TestStockReasonConstants 验证三阶段库存 reason 常量值稳定
func TestStockReasonConstants(t *testing.T) {
	reasons := map[string]string{
		"reserve":  modelsOrder.StockReasonStockReserved,
		"release":  modelsOrder.StockReasonStockReleased,
		"deduct":   modelsOrder.StockReasonStockDeducted,
		"dispatch": modelsOrder.StockReasonDispatched,
	}
	for name, reason := range reasons {
		if reason == "" {
			t.Fatalf("stock reason %q must not be empty", name)
		}
	}
	// 预留与释放应成对
	if modelsOrder.StockReasonStockReserved == modelsOrder.StockReasonStockReleased {
		t.Fatal("reserved and released reasons must differ")
	}
}
