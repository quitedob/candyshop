package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	modelsProduct "candypro/api/internal/models/product"
	productRepo "candypro/api/internal/repository/product"
	servicesCommon "candypro/api/internal/services/common"
	productService "candypro/api/internal/services/product"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// H11 acceptance tests. Before the fix, every admin product write routed
// through ProductService.UpdateProduct -> repo.Update, and GORM's struct-based
// Updates skips zero-valued fields: PUT /admin/products/{id} with
// {"featured": false}, {"stockQuantity": 0}, {"basePrice": 0} or
// {"summary": ""} silently no-oped at the DB level while the handler reported
// "Product updated successfully". These tests pin that the zero values now
// persist end-to-end (handler -> service -> repository -> sqlite row).

func setupProductUpdateTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsProduct.Product{}); err != nil {
		t.Fatalf("AutoMigrate products: %v", err)
	}
	return db
}

func newProductUpdateTestHandler(t *testing.T, db *gorm.DB) *Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	repo := productRepo.NewProductRepository(db)
	svcs := &servicesCommon.AdminPortalServices{
		Product: productService.NewProductService(repo),
	}
	return &Handler{services: svcs}
}

func seedProductUpdateRow(t *testing.T, db *gorm.DB, p *modelsProduct.Product) {
	t.Helper()
	if err := db.Create(p).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}
}

func productUpdateTestCtx(body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "admin-1")
	c.Request = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func mustReloadProduct(t *testing.T, db *gorm.DB, id string) modelsProduct.Product {
	t.Helper()
	var got modelsProduct.Product
	if err := db.Where("id = ?", id).First(&got).Error; err != nil {
		t.Fatalf("reload product %s: %v", id, err)
	}
	return got
}

func TestAdminUpdateProduct_PersistsFeaturedFalse(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedProductUpdateRow(t, db, &modelsProduct.Product{
		ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears",
		Featured: true, HalalCertified: true, StockQuantity: 5,
		BasePrice: 12.5, Status: modelsProduct.ProductStatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	h := newProductUpdateTestHandler(t, db)

	c, w := productUpdateTestCtx(`{"featured": false}`)
	c.Params = gin.Params{{Key: "id", Value: "p1"}}
	h.AdminUpdateProduct(c)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT featured=false: status=%d, want 200, body=%s", w.Code, w.Body.String())
	}
	got := mustReloadProduct(t, db, "p1")
	if got.Featured {
		t.Fatal("PUT /admin/products/{id} with {\"featured\": false} did not persist: featured still true")
	}
	// Unchanged fields must survive the full-row write.
	if got.HalalCertified != true || got.StockQuantity != 5 || got.BasePrice != 12.5 {
		t.Fatalf("full-row write corrupted unchanged fields: %+v", got)
	}
}

func TestAdminUpdateProduct_PersistsStockZero(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedProductUpdateRow(t, db, &modelsProduct.Product{
		ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears",
		Featured: true, StockQuantity: 5, BasePrice: 12.5,
		Status: modelsProduct.ProductStatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	h := newProductUpdateTestHandler(t, db)

	c, w := productUpdateTestCtx(`{"stockQuantity": 0}`)
	c.Params = gin.Params{{Key: "id", Value: "p1"}}
	h.AdminUpdateProduct(c)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT stockQuantity=0: status=%d, want 200, body=%s", w.Code, w.Body.String())
	}
	got := mustReloadProduct(t, db, "p1")
	if got.StockQuantity != 0 {
		t.Fatalf("PUT /admin/products/{id} with {\"stockQuantity\": 0} did not persist: got %d", got.StockQuantity)
	}
}

func TestAdminUpdateProduct_PersistsBasePriceZero(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedProductUpdateRow(t, db, &modelsProduct.Product{
		ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears",
		Featured: true, StockQuantity: 5, BasePrice: 12.5,
		Status: modelsProduct.ProductStatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	h := newProductUpdateTestHandler(t, db)

	c, w := productUpdateTestCtx(`{"basePrice": 0}`)
	c.Params = gin.Params{{Key: "id", Value: "p1"}}
	h.AdminUpdateProduct(c)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT basePrice=0: status=%d, want 200, body=%s", w.Code, w.Body.String())
	}
	got := mustReloadProduct(t, db, "p1")
	if got.BasePrice != 0 {
		t.Fatalf("PUT /admin/products/{id} with {\"basePrice\": 0} did not persist: got %v", got.BasePrice)
	}
}

func TestAdminUpdateProduct_PersistsSummaryEmpty(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedProductUpdateRow(t, db, &modelsProduct.Product{
		ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears", Summary: "fruit gummies",
		Featured: true, StockQuantity: 5, BasePrice: 12.5,
		Status: modelsProduct.ProductStatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	h := newProductUpdateTestHandler(t, db)

	c, w := productUpdateTestCtx(`{"summary": ""}`)
	c.Params = gin.Params{{Key: "id", Value: "p1"}}
	h.AdminUpdateProduct(c)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT summary=\"\": status=%d, want 200, body=%s", w.Code, w.Body.String())
	}
	got := mustReloadProduct(t, db, "p1")
	if got.Summary != "" {
		t.Fatalf("PUT /admin/products/{id} with {\"summary\": \"\"} did not persist: got %q", got.Summary)
	}
}

// AdminBatchUpdateInventory is the finding's "inventory corrections" path: a
// zero-stock correction must persist and the StockTransaction delta must match
// what the DB actually stored. The full-row overwrite closes the previous
// silent no-op for {"stockQuantity": 0}.
func TestAdminBatchUpdateInventory_PersistsZeroStock(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedProductUpdateRow(t, db, &modelsProduct.Product{
		ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears",
		Featured: true, StockQuantity: 5, BasePrice: 12.5,
		Status: modelsProduct.ProductStatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	h := newProductUpdateTestHandler(t, db)

	body := `{"ids":["p1"],"updates":{"stockQuantity":0}}`
	c, w := productUpdateTestCtx(body)
	h.AdminBatchUpdateInventory(c)

	if w.Code != http.StatusOK {
		t.Fatalf("batch update: status=%d, want 200, body=%s", w.Code, w.Body.String())
	}
	got := mustReloadProduct(t, db, "p1")
	if got.StockQuantity != 0 {
		t.Fatalf("batch inventory correction to 0 did not persist: got %d", got.StockQuantity)
	}
}

// AdminUpdateProductStatus also load-then-patches; ensure a status change on a
// row that would otherwise carry zero values still writes (full-row path).
func TestAdminUpdateProductStatus_Persists(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedProductUpdateRow(t, db, &modelsProduct.Product{
		ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears",
		Featured: true, StockQuantity: 5, BasePrice: 12.5,
		Status: modelsProduct.ProductStatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	h := newProductUpdateTestHandler(t, db)

	body := `{"status":"inactive"}`
	c, w := productUpdateTestCtx(body)
	c.Params = gin.Params{{Key: "id", Value: "p1"}}
	h.AdminUpdateProductStatus(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status update: status=%d, want 200, body=%s", w.Code, w.Body.String())
	}
	got := mustReloadProduct(t, db, "p1")
	if got.Status != "inactive" {
		t.Fatalf("status update did not persist: got %q", got.Status)
	}
}

// The ordinary partial-edit path must not regress: updating a non-zero field
// (name) still persists, and fields not touched in the request survive.
func TestAdminUpdateProduct_NormalEditPreservesUntouchedFields(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedProductUpdateRow(t, db, &modelsProduct.Product{
		ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears",
		Featured: true, HalalCertified: true, StockQuantity: 5,
		BasePrice: 12.5, Summary: "fruit gummies",
		Status: modelsProduct.ProductStatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	h := newProductUpdateTestHandler(t, db)

	c, w := productUpdateTestCtx(`{"name": "Gummy Bears 2"}`)
	c.Params = gin.Params{{Key: "id", Value: "p1"}}
	h.AdminUpdateProduct(c)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT name: status=%d, want 200, body=%s", w.Code, w.Body.String())
	}
	got := mustReloadProduct(t, db, "p1")
	if got.Name != "Gummy Bears 2" {
		t.Fatalf("name update did not persist: got %q", got.Name)
	}
	if !got.Featured || !got.HalalCertified || got.StockQuantity != 5 || got.BasePrice != 12.5 || got.Summary != "fruit gummies" {
		t.Fatalf("untouched fields were corrupted by the update: %+v", got)
	}
}

// presentProductColumns must return exactly the DB columns whose row keys are
// present, plus updated_at, and never "id".
func TestPresentProductColumns(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{
		"stockQuantity": float64(0),
		"halalCertified": true,
		"name":           "x",
	})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	if !set["stock_quantity"] {
		t.Fatalf("expected stock_quantity in columns, got %v", cols)
	}
	if !set["halal_certified"] {
		t.Fatalf("expected halal_certified in columns, got %v", cols)
	}
	if !set["name"] {
		t.Fatalf("expected name in columns, got %v", cols)
	}
	if !set["updated_at"] {
		t.Fatalf("expected updated_at appended, got %v", cols)
	}
	if set["id"] || set["category"] {
		t.Fatalf("absent/unknown columns leaked into set: %v", cols)
	}
	if len(cols) != 4 {
		t.Fatalf("unexpected column count %d: %v", len(cols), cols)
	}
}

// presentProductColumns must skip "status" when the row carries a value the
// importer would not apply (mirrors AdminApplyInventoryImport's guard), so a
// present-but-invalid status cannot blank the stored status via the
// column-scoped write.
func TestPresentProductColumns_SkipsInvalidStatus(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{
		"stockQuantity": float64(0),
		"status":        "not-a-status",
	})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	if set["status"] {
		t.Fatalf("invalid status leaked into columns: %v", cols)
	}
	if !set["stock_quantity"] || !set["updated_at"] {
		t.Fatalf("expected stock_quantity + updated_at, got %v", cols)
	}
	if len(cols) != 2 {
		t.Fatalf("unexpected column count %d: %v", len(cols), cols)
	}
}

// A valid status value must still be carried.
func TestPresentProductColumns_ValidStatus(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{"status": "active"})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	if !set["status"] || !set["updated_at"] {
		t.Fatalf("expected status + updated_at, got %v", cols)
	}
}
