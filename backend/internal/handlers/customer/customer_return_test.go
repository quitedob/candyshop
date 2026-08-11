package customer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	modelsOrder "candypro/api/internal/models/order"
	orderRepo "candypro/api/internal/repository/order"
	servicesCommon "candypro/api/internal/services/common"
	orderService "candypro/api/internal/services/order"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupReturnTestDB creates an in-memory SQLite DB with the return tables so the
// ReturnRepository.Create / FindByUserID paths used by CustomerCreateReturn work.
func setupReturnTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsOrder.ReturnRequest{}, &modelsOrder.ReturnItem{}); err != nil {
		t.Fatalf("migrate return tables: %v", err)
	}
	// :memory: creates a fresh DB per pooled connection; pin one connection so
	// every statement (including the concurrent double-return test) shares the
	// same in-memory database.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	return db
}

// buildTestReturnHandler wires a handler whose Order service reads from a seeded
// fakeOrderRepo and whose Return repo is backed by the in-memory DB. The same
// fakeOrderRepo is used by customer_orders_write_test.go.
func buildTestReturnHandler(db *gorm.DB, seedOrder *modelsOrder.Order) (*Handler, *fakeOrderRepo, *orderRepo.ReturnRepository) {
	orderFake := &fakeOrderRepo{}
	if seedOrder != nil {
		cp := *seedOrder
		orderFake.createdOrder = &cp
	}
	returnRepo := orderRepo.NewReturnRepository(db)
	svcs := &servicesCommon.UserPortalServices{
		Order:  orderService.NewOrderService(orderFake),
		Return: returnRepo,
	}
	return NewHandler(nil, svcs, nil, nil), orderFake, returnRepo
}

func performCustomerCreateReturn(handler *Handler, userID, orderID string, body map[string]interface{}) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/customer/orders/"+orderID+"/returns", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: orderID}}
	c.Set("userID", userID)
	handler.CustomerCreateReturn(c)
	return rec
}

func validReturnBody() map[string]interface{} {
	return map[string]interface{}{
		"reason": "wrong_item",
		"items": []map[string]interface{}{
			{
				"orderItemIdx": 0,
				"productId":    "p-a",
				"quantity":     2,
				"reasonCode":   "wrong_item",
				"refundAmount": 10,
			},
		},
	}
}

// deliveredOrder returns a shipped-and-delivered order with one line of product
// p-a qty 10 @ $5/unit, owned by u-owner.
func deliveredOrder() *modelsOrder.Order {
	return &modelsOrder.Order{
		ID:     "ord-delivered",
		UserID: "u-owner",
		Status: modelsOrder.OrderStatusDelivered,
		Items: modelsOrder.OrderItemArray{
			{ProductID: "p-a", Quantity: 10, UnitPrice: 5},
		},
		TotalAmount: 50,
	}
}

// TestCustomerCreateReturn_OwnershipForbidden (H2 IDOR): an authenticated
// customer must not be able to file a return against another user's order.
func TestCustomerCreateReturn_OwnershipForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	handler, _, _ := buildTestReturnHandler(db, deliveredOrder())

	rec := performCustomerCreateReturn(handler, "u-attacker", "ord-delivered", validReturnBody())
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-user return, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "forbidden" {
		t.Fatalf("expected error=forbidden, got %q", got)
	}
	var count int64
	if err := db.Model(&modelsOrder.ReturnRequest{}).Count(&count).Error; err != nil {
		t.Fatalf("count returns: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no return persisted on IDOR attempt, got %d", count)
	}
}

// TestCustomerCreateReturn_OrderNotFound: unknown order id must 404, not create.
func TestCustomerCreateReturn_OrderNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	handler, _, _ := buildTestReturnHandler(db, deliveredOrder())

	rec := performCustomerCreateReturn(handler, "u-owner", "ord-missing", validReturnBody())
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown order, got %d body=%s", rec.Code, rec.Body.String())
	}
	var count int64
	if err := db.Model(&modelsOrder.ReturnRequest{}).Count(&count).Error; err != nil {
		t.Fatalf("count returns: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no return persisted, got %d", count)
	}
}

// TestCustomerCreateReturn_StatusNotAllowed: pending orders cannot be returned.
func TestCustomerCreateReturn_StatusNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	order := deliveredOrder()
	order.Status = modelsOrder.OrderStatusPending
	handler, _, _ := buildTestReturnHandler(db, order)

	rec := performCustomerCreateReturn(handler, "u-owner", "ord-delivered", validReturnBody())
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for pending order return, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "return_status_not_allowed" {
		t.Fatalf("expected error=return_status_not_allowed, got %q", got)
	}
}

// TestCustomerCreateReturn_QuantityExceedsOrdered: cannot return more than ordered.
func TestCustomerCreateReturn_QuantityExceedsOrdered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	handler, _, _ := buildTestReturnHandler(db, deliveredOrder())

	body := validReturnBody()
	body["items"].([]map[string]interface{})[0]["quantity"] = 11

	rec := performCustomerCreateReturn(handler, "u-owner", "ord-delivered", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for qty > ordered, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "return_item_quantity_exceeds_ordered" {
		t.Fatalf("expected error=return_item_quantity_exceeds_ordered, got %q", got)
	}
}

// TestCustomerCreateReturn_ProductMismatch: the return line must reference the
// same product as the order line at that index.
func TestCustomerCreateReturn_ProductMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	handler, _, _ := buildTestReturnHandler(db, deliveredOrder())

	body := validReturnBody()
	body["items"].([]map[string]interface{})[0]["productId"] = "p-other"

	rec := performCustomerCreateReturn(handler, "u-owner", "ord-delivered", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for product mismatch, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "return_item_invalid" {
		t.Fatalf("expected error=return_item_invalid, got %q", got)
	}
}

// TestCustomerCreateReturn_RefundExceedsLine: refund cannot exceed the line's
// unit price x quantity.
func TestCustomerCreateReturn_RefundExceedsLine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	handler, _, _ := buildTestReturnHandler(db, deliveredOrder())

	body := validReturnBody()
	body["items"].([]map[string]interface{})[0]["refundAmount"] = 99 // line total is 2 x $5 = $10

	rec := performCustomerCreateReturn(handler, "u-owner", "ord-delivered", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for refund > line, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "return_refund_exceeds_line" {
		t.Fatalf("expected error=return_refund_exceeds_line, got %q", got)
	}
}

// TestCustomerCreateReturn_NegativeRefundRejected: refund amounts cannot be
// negative (a negative refund would let the total-amount cap be bypassed).
func TestCustomerCreateReturn_NegativeRefundRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	handler, _, _ := buildTestReturnHandler(db, deliveredOrder())

	body := validReturnBody()
	body["items"].([]map[string]interface{})[0]["refundAmount"] = -5

	rec := performCustomerCreateReturn(handler, "u-owner", "ord-delivered", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative refund, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestCustomerCreateReturn_Success: valid return persists with caller user id.
func TestCustomerCreateReturn_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	handler, _, _ := buildTestReturnHandler(db, deliveredOrder())

	rec := performCustomerCreateReturn(handler, "u-owner", "ord-delivered", validReturnBody())
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var ret modelsOrder.ReturnRequest
	if err := db.Preload("Items").Where("order_id = ?", "ord-delivered").First(&ret).Error; err != nil {
		t.Fatalf("load persisted return: %v", err)
	}
	if ret.UserID != "u-owner" {
		t.Fatalf("expected persisted userID=u-owner, got %q", ret.UserID)
	}
	if ret.Status != modelsOrder.ReturnStatusPending {
		t.Fatalf("expected status=pending, got %q", ret.Status)
	}
	if len(ret.Items) != 1 {
		t.Fatalf("expected 1 return item, got %d", len(ret.Items))
	}
	if ret.Items[0].Quantity != 2 || ret.Items[0].ProductID != "p-a" {
		t.Fatalf("unexpected return item: %+v", ret.Items[0])
	}
	// The create response must echo the persisted lines, not an empty items array.
	var createdResp struct {
		Items []struct {
			OrderItemIdx int    `json:"orderItemIdx"`
			ProductID    string `json:"productId"`
			Quantity     int    `json:"quantity"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createdResp); err != nil {
		t.Fatalf("parse create response: %v", err)
	}
	if len(createdResp.Items) != 1 || createdResp.Items[0].Quantity != 2 || createdResp.Items[0].ProductID != "p-a" {
		t.Fatalf("expected items in create response, got %+v", createdResp.Items)
	}
}

// TestCustomerCreateReturn_DoubleReturnRejected: returning the same line twice
// (full qty first, then more) must be rejected by the already-returned tally.
func TestCustomerCreateReturn_DoubleReturnRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	handler, _, _ := buildTestReturnHandler(db, deliveredOrder())

	first := validReturnBody()
	first["items"].([]map[string]interface{})[0]["quantity"] = 10 // claim the whole line
	rec1 := performCustomerCreateReturn(handler, "u-owner", "ord-delivered", first)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("expected first return 201, got %d body=%s", rec1.Code, rec1.Body.String())
	}

	second := validReturnBody() // qty 2 on the same, fully-claimed line
	rec2 := performCustomerCreateReturn(handler, "u-owner", "ord-delivered", second)
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("expected second return 400 (line fully claimed), got %d body=%s", rec2.Code, rec2.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "return_item_quantity_exceeds_ordered" {
		t.Fatalf("expected error=return_item_quantity_exceeds_ordered, got %q", got)
	}
}

// TestCustomerCreateReturn_ConcurrentFullClaim (H2 TOCTOU): two simultaneous
// requests claiming the full line must serialize on the per-process mutex so
// exactly one succeeds and the combined persisted quantity never exceeds the
// ordered line.
func TestCustomerCreateReturn_ConcurrentFullClaim(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	handler, _, _ := buildTestReturnHandler(db, deliveredOrder())

	body := validReturnBody()
	body["items"].([]map[string]interface{})[0]["quantity"] = 10 // claim the whole line

	var wg sync.WaitGroup
	codes := make([]int, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			codes[idx] = performCustomerCreateReturn(handler, "u-owner", "ord-delivered", body).Code
		}(i)
	}
	wg.Wait()

	var created, rejected int
	for _, code := range codes {
		switch code {
		case http.StatusCreated:
			created++
		case http.StatusBadRequest:
			rejected++
		}
	}
	if created != 1 || rejected != 1 {
		t.Fatalf("expected exactly one 201 and one 400 for concurrent full-line claim, got codes=%v", codes)
	}
	var items []modelsOrder.ReturnItem
	if err := db.Find(&items).Error; err != nil {
		t.Fatalf("load persisted items: %v", err)
	}
	sum := 0
	for _, it := range items {
		sum += it.Quantity
	}
	if sum != 10 {
		t.Fatalf("expected total persisted quantity 10 (single full-line claim), got %d", sum)
	}
}

// TestCustomerCreateReturn_ShippedQuantityCapsClaim (H2 edge): a claim may not
// exceed what the order line records as shipped/fulfilled.
func TestCustomerCreateReturn_ShippedQuantityCapsClaim(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	order := deliveredOrder()
	order.Items = modelsOrder.OrderItemArray{
		{ProductID: "p-a", Quantity: 10, UnitPrice: 5, ShippedQuantity: 5},
	}
	handler, _, _ := buildTestReturnHandler(db, order)

	over := validReturnBody()
	over["items"].([]map[string]interface{})[0]["quantity"] = 8 // > shipped 5
	if rec := performCustomerCreateReturn(handler, "u-owner", "ord-delivered", over); rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for qty > shipped, got %d body=%s", rec.Code, rec.Body.String())
	}

	at := validReturnBody()
	at["items"].([]map[string]interface{})[0]["quantity"] = 5 // == shipped 5
	if rec := performCustomerCreateReturn(handler, "u-owner", "ord-delivered", at); rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for qty == shipped, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestCustomerCreateReturn_TallyAcrossAllPages (H2 edge): the already-returned
// tally must aggregate over every page of the caller's returns, not just the
// first 1000. With 1001 prior pending returns each claiming 1 unit of a 1001-unit
// line, the line is fully claimed and any further claim must be rejected — the
// old limit=1000 read would have dropped the oldest claim and accepted it.
func TestCustomerCreateReturn_TallyAcrossAllPages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReturnTestDB(t)
	order := deliveredOrder()
	order.Items = modelsOrder.OrderItemArray{
		{ProductID: "p-a", Quantity: 1001, UnitPrice: 5},
	}
	handler, _, _ := buildTestReturnHandler(db, order)

	const n = 1001
	reqs := make([]modelsOrder.ReturnRequest, 0, n)
	items := make([]modelsOrder.ReturnItem, 0, n)
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("RET-prior-%d", i)
		reqs = append(reqs, modelsOrder.ReturnRequest{
			ID:      id,
			OrderID: order.ID,
			UserID:  "u-owner",
			Status:  modelsOrder.ReturnStatusPending,
			Reason:  "wrong_item",
		})
		items = append(items, modelsOrder.ReturnItem{
			ReturnID:     id,
			OrderItemIdx: 0,
			ProductID:    "p-a",
			Quantity:     1,
			ReasonCode:   "wrong_item",
		})
	}
	if err := db.Create(&reqs).Error; err != nil {
		t.Fatalf("seed return requests: %v", err)
	}
	if err := db.Create(&items).Error; err != nil {
		t.Fatalf("seed return items: %v", err)
	}

	body := validReturnBody()
	body["items"].([]map[string]interface{})[0]["quantity"] = 1
	body["items"].([]map[string]interface{})[0]["refundAmount"] = 1
	rec := performCustomerCreateReturn(handler, "u-owner", "ord-delivered", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (line fully claimed across >1000 returns), got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "return_item_quantity_exceeds_ordered" {
		t.Fatalf("expected error=return_item_quantity_exceeds_ordered, got %q", got)
	}
}
