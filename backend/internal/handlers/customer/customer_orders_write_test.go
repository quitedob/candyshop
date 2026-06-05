package customer

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"candypro/api/internal/config"
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	modelsUser "candypro/api/internal/models/user"
	orderRepo "candypro/api/internal/repository/order"
	servicesCommon "candypro/api/internal/services/common"
	inquiryService "candypro/api/internal/services/inquiry"
	orderService "candypro/api/internal/services/order"
	productService "candypro/api/internal/services/product"
	userService "candypro/api/internal/services/user"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// fakeUserRepo 测试用用户仓储：默认返回已激活用户，满足 KYB 校验
type fakeUserRepo struct{}

func (f *fakeUserRepo) FindByID(ctx context.Context, id string) (*modelsUser.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &modelsUser.User{ID: id, Status: "active"}, nil
}
func (f *fakeUserRepo) FindByEmail(ctx context.Context, email string) (*modelsUser.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeUserRepo) FindByResetToken(ctx context.Context, token string) (*modelsUser.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeUserRepo) FindAll(ctx context.Context, page, limit int) ([]modelsUser.User, int64, error) {
	return nil, 0, nil
}
func (f *fakeUserRepo) FindRecent(ctx context.Context, limit int) ([]modelsUser.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) Create(ctx context.Context, user *modelsUser.User) error {
	return nil
}
func (f *fakeUserRepo) Update(ctx context.Context, user *modelsUser.User) error {
	return nil
}
func (f *fakeUserRepo) Delete(ctx context.Context, id string) error {
	return nil
}
func (f *fakeUserRepo) CountAll(ctx context.Context) (int64, error) {
	return 0, nil
}
func (f *fakeUserRepo) CountCreatedSince(ctx context.Context, since time.Time) (int64, error) {
	return 0, nil
}
func (f *fakeUserRepo) ActivateUsersByCompanyID(ctx context.Context, companyID string) error {
	return nil
}
func (f *fakeUserRepo) FindByRoleNames(ctx context.Context, names []string) ([]modelsUser.User, error) {
	return []modelsUser.User{}, nil
}

type fakeProductRepo struct {
	products map[string]modelsProduct.Product
}

func (f *fakeProductRepo) FindAll(ctx context.Context, page, limit int, categorySlug string) ([]modelsProduct.Product, int64, error) {
	return []modelsProduct.Product{}, 0, nil
}

func (f *fakeProductRepo) FindAllForAdmin(ctx context.Context, page, limit int, categorySlug, status, search string) ([]modelsProduct.Product, int64, error) {
	return []modelsProduct.Product{}, 0, nil
}

func (f *fakeProductRepo) FindBySlug(ctx context.Context, slug string) (*modelsProduct.Product, error) {
	for _, p := range f.products {
		if p.Slug == slug {
			cp := p
			return &cp, nil
		}
	}
	return nil, context.Canceled
}

func (f *fakeProductRepo) FindByID(ctx context.Context, id string) (*modelsProduct.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return nil, context.Canceled
	}
	cp := p
	return &cp, nil
}

func (f *fakeProductRepo) FindByIDs(ctx context.Context, ids []string) ([]modelsProduct.Product, error) {
	if len(ids) == 0 || f.products == nil {
		return nil, nil
	}
	out := make([]modelsProduct.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := f.products[id]; ok {
			cp := p
			out = append(out, cp)
		}
	}
	return out, nil
}

func (f *fakeProductRepo) FindFeatured(ctx context.Context, limit int) ([]modelsProduct.Product, error) {
	return []modelsProduct.Product{}, nil
}

func (f *fakeProductRepo) FindRelated(ctx context.Context, slug string, limit int) ([]modelsProduct.Product, error) {
	return []modelsProduct.Product{}, nil
}

func (f *fakeProductRepo) Create(ctx context.Context, product *modelsProduct.Product) error {
	return nil
}

func (f *fakeProductRepo) Update(ctx context.Context, product *modelsProduct.Product) error {
	return nil
}

func (f *fakeProductRepo) UpdateStockWithLock(ctx context.Context, productID string, newQty int) (int, error) {
	return newQty, nil
}

func (f *fakeProductRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (f *fakeProductRepo) FindVariantsByProductID(ctx context.Context, productID string) ([]modelsProduct.ProductVariant, error) {
	return []modelsProduct.ProductVariant{}, nil
}

func (f *fakeProductRepo) FindAllFiltered(ctx context.Context, page, limit int, halal, oemOnly, featuredOnly bool, search, sort string, minMOQ, maxMOQ int, categorySlug ...string) ([]modelsProduct.Product, int64, error) {
	return []modelsProduct.Product{}, 0, nil
}

func (f *fakeProductRepo) FindMarketProfilesForProducts(ctx context.Context, productIDs []string, marketCode string) ([]modelsProduct.ProductMarketProfile, error) {
	return nil, nil
}

func (f *fakeProductRepo) UpsertProductMarketProfile(ctx context.Context, row *modelsProduct.ProductMarketProfile) error {
	return nil
}

func (f *fakeProductRepo) FindMarketCostStacksForProduct(ctx context.Context, productID string) ([]modelsProduct.ProductMarketCostStack, error) {
	return nil, nil
}

func (f *fakeProductRepo) UpsertProductMarketCostStack(ctx context.Context, row *modelsProduct.ProductMarketCostStack) error {
	return nil
}

func (f *fakeProductRepo) ListWarehouses(ctx context.Context) ([]modelsProduct.Warehouse, error) {
	return nil, nil
}

func (f *fakeProductRepo) SaveWarehouse(ctx context.Context, w *modelsProduct.Warehouse) error {
	return nil
}

func (f *fakeProductRepo) UpsertWarehouseStock(ctx context.Context, row *modelsProduct.WarehouseStock) error {
	return nil
}

func (f *fakeProductRepo) SaveOEMProjectInventoryHold(ctx context.Context, row *modelsProduct.OEMProjectInventoryHold) error {
	return nil
}

func (f *fakeProductRepo) ListOEMInventoryHoldsByProject(ctx context.Context, projectID string) ([]modelsProduct.OEMProjectInventoryHold, error) {
	return nil, nil
}

func (f *fakeProductRepo) UpdateOEMHoldStatusIfMatches(ctx context.Context, id uint, expected, target string) (int64, error) {
	return 0, nil
}

func (f *fakeProductRepo) SumActiveOEMHoldsForProduct(ctx context.Context, productID string) (int64, error) {
	return 0, nil
}

// SumActiveOEMHoldsByProductIDs E-1 批量查询：测试用空 map 返回零预留。
func (f *fakeProductRepo) SumActiveOEMHoldsByProductIDs(ctx context.Context, productIDs []string) (map[string]int64, error) {
	return map[string]int64{}, nil
}

func (f *fakeProductRepo) ListChannelInventoriesForProduct(ctx context.Context, productID string) ([]modelsProduct.ChannelInventory, error) {
	return nil, nil
}

func (f *fakeProductRepo) FindChannelInventory(ctx context.Context, productID, channelCode string) (*modelsProduct.ChannelInventory, error) {
	return nil, gorm.ErrRecordNotFound
}

// FindChannelInventoriesByProductIDs E-1 批量查询：测试用空 map 表示无渠道库存约束。
func (f *fakeProductRepo) FindChannelInventoriesByProductIDs(ctx context.Context, productIDs []string, channelCode string) (map[string]*modelsProduct.ChannelInventory, error) {
	return map[string]*modelsProduct.ChannelInventory{}, nil
}

func (f *fakeProductRepo) UpsertChannelInventory(ctx context.Context, row *modelsProduct.ChannelInventory) error {
	return nil
}

func (f *fakeProductRepo) GetDefaultWarehouseID(ctx context.Context) (string, error) {
	return "", nil
}

func (f *fakeProductRepo) ComputeWeightedAvgCost(ctx context.Context, productID string) float64 {
	return 0
}

type fakeInquiryRepo struct{}

func (f *fakeInquiryRepo) Create(ctx context.Context, inquiry *modelsProduct.Inquiry) error {
	return nil
}
func (f *fakeInquiryRepo) FindByID(ctx context.Context, id string) (*modelsProduct.Inquiry, error) {
	return nil, context.Canceled
}
func (f *fakeInquiryRepo) FindAll(ctx context.Context, page, limit int) ([]modelsProduct.Inquiry, int64, error) {
	return []modelsProduct.Inquiry{}, 0, nil
}
func (f *fakeInquiryRepo) FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsProduct.Inquiry, int64, error) {
	return []modelsProduct.Inquiry{}, 0, nil
}
func (f *fakeInquiryRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	return nil
}
func (f *fakeInquiryRepo) Update(ctx context.Context, inquiry *modelsProduct.Inquiry) error {
	return nil
}
func (f *fakeInquiryRepo) Delete(ctx context.Context, id string) error {
	return nil
}
func (f *fakeInquiryRepo) CountAll(ctx context.Context) (int64, error) {
	return 0, nil
}
func (f *fakeInquiryRepo) CountByStatus(ctx context.Context, status string) (int64, error) {
	return 0, nil
}
func (f *fakeInquiryRepo) CountByStatusGrouped(ctx context.Context) (map[string]int64, error) {
	return nil, nil
}
func (f *fakeInquiryRepo) FindRecent(ctx context.Context, limit int) ([]modelsProduct.Inquiry, error) {
	return []modelsProduct.Inquiry{}, nil
}

func (f *fakeInquiryRepo) ConversionByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

type fakeOrderRepo struct {
	createdOrder         *modelsOrder.Order
	forceConfirmNotFound bool
}

func (f *fakeOrderRepo) FindAll(ctx context.Context, page, limit int, status, userID, dateFrom, dateTo string) ([]modelsOrder.Order, int64, error) {
	return []modelsOrder.Order{}, 0, nil
}
func (f *fakeOrderRepo) FindByID(ctx context.Context, id string) (*modelsOrder.Order, error) {
	if f.createdOrder == nil || f.createdOrder.ID != id {
		return nil, context.Canceled
	}
	cp := *f.createdOrder
	return &cp, nil
}
func (f *fakeOrderRepo) FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsOrder.Order, int64, error) {
	return []modelsOrder.Order{}, 0, nil
}
func (f *fakeOrderRepo) Create(ctx context.Context, order *modelsOrder.Order) error {
	cp := *order
	f.createdOrder = &cp
	return nil
}
func (f *fakeOrderRepo) CreateWithStockReservation(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	cp := *order
	f.createdOrder = &cp
	return nil
}
func (f *fakeOrderRepo) Update(ctx context.Context, order *modelsOrder.Order) error {
	return nil
}
func (f *fakeOrderRepo) UpdateWithOutbox(ctx context.Context, order *modelsOrder.Order, outbox *modelsOrder.EventOutbox) error {
	return nil
}
func (f *fakeOrderRepo) UpdateWithOptionalStockReservationAndOutbox(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int, reserve bool, outbox *modelsOrder.EventOutbox) error {
	if reserve {
		order.StockReserved = true
	}
	return nil
}
func (f *fakeOrderRepo) ListPendingOutbox(ctx context.Context, eventType string, limit int) ([]modelsOrder.EventOutbox, error) {
	return nil, nil
}
func (f *fakeOrderRepo) IncrementOutboxAttempt(ctx context.Context, id uint) error {
	return nil
}
func (f *fakeOrderRepo) UpdateOutboxResult(ctx context.Context, id uint, status, lastErr string, processedAt *time.Time) error {
	return nil
}
func (f *fakeOrderRepo) FindInquiryTradeHints(ctx context.Context, inquiryID string) (incoterms, commercialNotes string, err error) {
	return "", "", nil
}
func (f *fakeOrderRepo) UpdateWithStockAdjustment(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return nil
}
func (f *fakeOrderRepo) Delete(ctx context.Context, id string) error {
	return nil
}
func (f *fakeOrderRepo) ReleaseStockForOrder(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	cp := *order
	cp.StockReserved = false
	f.createdOrder = &cp
	return nil
}
func (f *fakeOrderRepo) ReserveStockForOrder(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	cp := *order
	cp.StockReserved = true
	f.createdOrder = &cp
	return nil
}
func (f *fakeOrderRepo) DeleteWithStockRestore(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	f.createdOrder = nil
	return nil
}
func (f *fakeOrderRepo) ConfirmPendingOrder(ctx context.Context, id string, confirmedAt time.Time) error {
	if f.createdOrder == nil || f.createdOrder.ID != id {
		return context.Canceled
	}
	if f.forceConfirmNotFound {
		f.createdOrder.Status = "cancelled"
		f.createdOrder.StockReserved = false
		f.createdOrder.UpdatedAt = confirmedAt
		return gorm.ErrRecordNotFound
	}
	f.createdOrder.Status = "pending"
	f.createdOrder.StockReserved = true
	f.createdOrder.ConfirmedAt = &confirmedAt
	f.createdOrder.UpdatedAt = confirmedAt
	return nil
}
func (f *fakeOrderRepo) ConfirmAndReserveStock(ctx context.Context, id string, stockDeltas map[string]int, confirmedAt time.Time) error {
	if f.createdOrder == nil || f.createdOrder.ID != id {
		return context.Canceled
	}
	if f.forceConfirmNotFound {
		f.createdOrder.Status = "cancelled"
		f.createdOrder.StockReserved = false
		f.createdOrder.UpdatedAt = confirmedAt
		return gorm.ErrRecordNotFound
	}
	f.createdOrder.Status = "pending"
	f.createdOrder.StockReserved = true
	f.createdOrder.ConfirmedAt = &confirmedAt
	f.createdOrder.UpdatedAt = confirmedAt
	return nil
}
func (f *fakeOrderRepo) ReleaseExpiredPendingConfirmationOrders(ctx context.Context, olderThan time.Time, limit int) (int, error) {
	return 0, nil
}
func (f *fakeOrderRepo) CountAll(ctx context.Context) (int64, error) {
	return 0, nil
}
func (f *fakeOrderRepo) CountByStatuses(ctx context.Context, statuses []string) (int64, error) {
	return 0, nil
}
func (f *fakeOrderRepo) SumTotalAmount(ctx context.Context) (float64, error) {
	return 0, nil
}
func (f *fakeOrderRepo) SumTotalAmountSince(ctx context.Context, since time.Time) (float64, error) {
	return 0, nil
}
func (f *fakeOrderRepo) FindRecent(ctx context.Context, limit int) ([]modelsOrder.Order, error) {
	return []modelsOrder.Order{}, nil
}

func (f *fakeOrderRepo) RevenueByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (f *fakeOrderRepo) OrderCountByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (f *fakeOrderRepo) TopProductsByRevenue(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (f *fakeOrderRepo) DistinctOrderingUsers(ctx context.Context, since time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeOrderRepo) RevenueByDay(ctx context.Context, days int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}
func (f *fakeOrderRepo) SalesVelocity(ctx context.Context, months int) ([]orderRepo.SalesVelocityResult, error) {
	return nil, nil
}
func (f *fakeOrderRepo) RFMAnalysis(ctx context.Context) ([]orderRepo.RFMRecord, error) {
	return nil, nil
}
func (f *fakeOrderRepo) CustomerChurn(ctx context.Context, dormantDays int) ([]orderRepo.CustomerChurnResult, error) {
	return nil, nil
}
func (f *fakeOrderRepo) InventoryHealth(ctx context.Context, salesWindowDays int) ([]orderRepo.InventoryHealthResult, error) {
	return nil, nil
}
func (f *fakeOrderRepo) ProfitLossByPeriod(ctx context.Context, groupBy string, periods int) ([]orderRepo.ProfitLossResult, error) {
	return nil, nil
}
func (f *fakeOrderRepo) ReplenishmentSuggestions(ctx context.Context, cycleDays int, salesWindowDays int) ([]orderRepo.ReplenishmentItem, error) {
	return nil, nil
}

func TestCustomerCreateOrder_ComplianceViolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-eu-block": {
			ID:            "p-eu-block",
			Name:          "EU Block Candy",
			Status:        "active",
			MOQ:           10,
			StockQuantity: 1000,
			BasePrice:     1.2,
			Ingredients:   "Sugar, Titanium Dioxide (E171), Flavor",
			Allergens:     "none",
		},
	})

	rec := performCustomerCreateOrder(handler, "u-100", map[string]interface{}{
		"currency": "USD",
		"items": []map[string]interface{}{
			{
				"productId": "p-eu-block",
				"quantity":  100,
				"unitPrice": 1.2,
			},
		},
		"shippingAddress": map[string]interface{}{
			"street":  "123 Demo Street",
			"city":    "Paris",
			"state":   "IDF",
			"zipCode": "75001",
			"country": "France",
		},
	})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder != nil {
		t.Fatalf("order should not be created when compliance fails")
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "compliance_violation" {
		t.Fatalf("expected error=compliance_violation, got %q", got)
	}
}

func TestCustomerCreateOrder_CompliancePass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-us-ok": {
			ID:            "p-us-ok",
			Name:          "US OK Candy",
			Status:        "active",
			MOQ:           10,
			StockQuantity: 1000,
			BasePrice:     1.2,
			Ingredients:   "Sugar, Pectin, Flavor",
			Allergens:     "none",
		},
	})

	rec := performCustomerCreateOrder(handler, "u-200", map[string]interface{}{
		"currency": "USD",
		"items": []map[string]interface{}{
			{
				"productId": "p-us-ok",
				"quantity":  50,
				"unitPrice": 2.4,
			},
		},
		"shippingAddress": map[string]interface{}{
			"street":  "500 Main St",
			"city":    "New York",
			"state":   "NY",
			"zipCode": "10001",
			"country": "USA",
		},
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder == nil {
		t.Fatalf("expected order to be created")
	}
	if orderRepo.createdOrder.UserID != "u-200" {
		t.Fatalf("expected created order userID=u-200, got %s", orderRepo.createdOrder.UserID)
	}
	if orderRepo.createdOrder.StockReserved {
		t.Fatalf("expected created order StockReserved=false (deferred to admin confirmation)")
	}
}

func TestCustomerCreateOrder_InventoryViolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{
		"p-stock-low": {
			ID:             "p-stock-low",
			Name:           "Low Stock Candy",
			Status:         "active",
			MOQ:            10,
			StockQuantity:  20,
			BasePrice:      1.1,
			Ingredients:    "Sugar, pectin",
			Allergens:      "none",
			HalalCertified: true,
			Certifications: modelsCommon.StringArray{"Halal", "HACCP"},
		},
	})

	rec := performCustomerCreateOrder(handler, "u-300", map[string]interface{}{
		"currency": "USD",
		"items": []map[string]interface{}{
			{
				"productId": "p-stock-low",
				"quantity":  50,
				"unitPrice": 1.1,
			},
		},
		"shippingAddress": map[string]interface{}{
			"street":  "1 Demo Road",
			"city":    "Karachi",
			"state":   "Sindh",
			"zipCode": "74000",
			"country": "Pakistan",
		},
	})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder != nil {
		t.Fatalf("order should not be created when inventory fails")
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "inventory_violation" {
		t.Fatalf("expected error=inventory_violation, got %q", got)
	}
}

func TestCustomerConfirmOrder_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-confirm-ok",
		UserID:                     "u-confirm",
		Status:                     "pending_confirmation",
		StockReserved:              true,
		ComplianceOfficialEvidence: true,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-confirm", "ord-confirm-ok")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder.Status != "pending" {
		t.Fatalf("expected status=pending after confirm, got %s", orderRepo.createdOrder.Status)
	}
	if orderRepo.createdOrder.ConfirmedAt == nil {
		t.Fatalf("expected confirmedAt to be set")
	}
}

func TestCustomerConfirmOrder_ComplianceAckRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-confirm-ack-required",
		UserID:                     "u-ack-required",
		Status:                     "pending_confirmation",
		StockReserved:              true,
		ComplianceOfficialEvidence: false,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrder(handler, "u-ack-required", "ord-confirm-ack-required")
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder.Status != "pending_confirmation" {
		t.Fatalf("expected status to remain pending_confirmation, got %s", orderRepo.createdOrder.Status)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "compliance_ack_required" {
		t.Fatalf("expected error=compliance_ack_required, got %q", got)
	}
}

func TestCustomerConfirmOrder_ComplianceAckProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-confirm-ack-ok",
		UserID:                     "u-ack-ok",
		Status:                     "pending_confirmation",
		StockReserved:              true,
		ComplianceOfficialEvidence: false,
		CreatedAt:                  now.Add(-5 * time.Minute),
		UpdatedAt:                  now.Add(-5 * time.Minute),
	}

	rec := performCustomerConfirmOrderWithBody(handler, "u-ack-ok", "ord-confirm-ack-ok", map[string]interface{}{
		"complianceAck": true,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if orderRepo.createdOrder.Status != "pending" {
		t.Fatalf("expected status=pending after confirm with ack, got %s", orderRepo.createdOrder.Status)
	}
}

func TestCustomerConfirmOrder_ExpiredDraft(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, orderRepo := buildTestCustomerOrderHandler(map[string]modelsProduct.Product{})
	now := time.Now()
	orderRepo.createdOrder = &modelsOrder.Order{
		ID:                         "ord-confirm-expired",
		UserID:                     "u-expired",
		Status:                     "pending_confirmation",
		StockReserved:              true,
		ComplianceOfficialEvidence: true,
		CreatedAt:                  now.Add(-2 * time.Hour),
		UpdatedAt:                  now.Add(-2 * time.Hour),
	}
	orderRepo.forceConfirmNotFound = true

	rec := performCustomerConfirmOrder(handler, "u-expired", "ord-confirm-expired")
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got, _ := resp["error"].(string); got != "order_expired" {
		t.Fatalf("expected error=order_expired, got %q", got)
	}
}

func buildTestCustomerOrderHandler(products map[string]modelsProduct.Product) (*Handler, *fakeOrderRepo) {
	cfg := &config.Config{}
	orderRepo := &fakeOrderRepo{}

	svcs := &servicesCommon.UserPortalServices{
		User:    userService.NewUserService(&fakeUserRepo{}),
		Product: productService.NewProductService(&fakeProductRepo{products: products}),
		Inquiry: inquiryService.NewInquiryService(&fakeInquiryRepo{}, cfg),
		Order:   orderService.NewOrderService(orderRepo),
	}

	return NewHandler(cfg, svcs, nil, nil), orderRepo
}

func performCustomerCreateOrder(handler *Handler, userID string, body map[string]interface{}) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/customer/orders", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Set("userID", userID)

	handler.CustomerCreateOrder(c)
	return rec
}

func performCustomerConfirmOrder(handler *Handler, userID, orderID string) *httptest.ResponseRecorder {
	return performCustomerConfirmOrderWithBody(handler, userID, orderID, nil)
}

func performCustomerConfirmOrderWithBody(handler *Handler, userID, orderID string, body map[string]interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Reader
	if body == nil {
		reqBody = bytes.NewReader(nil)
	} else {
		raw, _ := json.Marshal(body)
		reqBody = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(http.MethodPost, "/customer/orders/"+orderID+"/confirm", reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: orderID}}
	c.Set("userID", userID)

	handler.CustomerConfirmOrder(c)
	return rec
}
