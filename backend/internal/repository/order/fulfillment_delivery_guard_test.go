package order

import (
	"context"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"
)

func TestFulfillmentDeliver_ValidatesOrderAndRollsBack(t *testing.T) {
	for _, testCase := range []struct {
		status            string
		remainingShipment bool
		wantStatus        string
		wantError         bool
	}{
		{status: modelsOrder.OrderStatusCancelled, wantError: true},
		{status: modelsOrder.OrderStatusReturned, wantError: true},
		{status: modelsOrder.OrderStatusProduction, wantError: true},
		{status: modelsOrder.OrderStatusShipped, wantStatus: modelsOrder.OrderStatusDelivered},
		{status: modelsOrder.OrderStatusShipped, remainingShipment: true, wantStatus: modelsOrder.OrderStatusPartiallyDelivered},
		{status: modelsOrder.OrderStatusPartiallyDelivered, wantStatus: modelsOrder.OrderStatusDelivered},
	} {
		t.Run(testCase.status+testCase.wantStatus, func(t *testing.T) {
			database := setupStateGuardTestDB(t)
			if err := database.Exec(`CREATE TABLE orders (id TEXT PRIMARY KEY, status TEXT, version INTEGER NOT NULL DEFAULT 0, delivered_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`).Error; err != nil {
				t.Fatal(err)
			}
			if err := database.Exec(`INSERT INTO orders (id, status) VALUES (?, ?)`, "order-delivery", testCase.status).Error; err != nil {
				t.Fatal(err)
			}
			fulfillment := modelsOrder.Fulfillment{ID: "shipment-delivery", OrderID: "order-delivery", WarehouseID: "warehouse-test", Status: modelsOrder.FulfillmentStatusShipped}
			if err := database.Create(&fulfillment).Error; err != nil {
				t.Fatal(err)
			}
			if testCase.remainingShipment {
				remaining := fulfillment
				remaining.ID = "shipment-remaining"
				if err := database.Create(&remaining).Error; err != nil {
					t.Fatal(err)
				}
			}
			err := NewFulfillmentRepository(database).Deliver(context.Background(), fulfillment.ID, time.Now())
			if (err != nil) != testCase.wantError {
				t.Fatalf("Deliver error=%v, wantError=%v", err, testCase.wantError)
			}
			var order modelsOrder.Order
			if err := database.First(&order).Error; err != nil {
				t.Fatal(err)
			}
			if err := database.First(&fulfillment, "id = ?", fulfillment.ID).Error; err != nil {
				t.Fatal(err)
			}
			if testCase.wantError {
				if fulfillment.Status != modelsOrder.FulfillmentStatusShipped || order.Status != testCase.status || order.Version != 0 {
					t.Fatalf("rejected transition changed data: fulfillment=%s order=%s version=%d", fulfillment.Status, order.Status, order.Version)
				}
			} else if order.Status != testCase.wantStatus || order.Version != 1 || fulfillment.Status != modelsOrder.FulfillmentStatusDelivered {
				t.Fatalf("delivery did not advance both records/version: fulfillment=%s order=%s version=%d", fulfillment.Status, order.Status, order.Version)
			}
		})
	}
}
