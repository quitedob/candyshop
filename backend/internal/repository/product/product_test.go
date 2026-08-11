package product

import (
	"context"
	"strings"
	"testing"
	"time"

	modelsProduct "candypro/api/internal/models/product"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// embeddingWithDim returns an embedding vector of the given dimension with the
// leading values set (rest zero). product_embeddings.embedding is vector(1536),
// and UpsertEmbedding now validates the dimension, so tests must use 1536-dim
// vectors (G24b).
func embeddingWithDim(dim int, values ...float32) []float32 {
	v := make([]float32, dim)
	copy(v, values)
	return v
}

// setupProductTestDB creates an in-memory SQLite with the products table
// migrated from the Product model so full-row / column-scoped updates can write
// every column (H11).
func setupProductTestDB(t *testing.T) *gorm.DB {
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

func seedProduct(t *testing.T, repo *ProductRepository, p *modelsProduct.Product) {
	t.Helper()
	if err := repo.Create(context.Background(), p); err != nil {
		t.Fatalf("create: %v", err)
	}
}

// TestUpdate_SkipsZeroFields pins the partial-update contract that GORM's
// struct-Updates skips zero-valued fields. The XLSX importer
// (AdminApplyInventoryImport) builds a fresh struct from a subset of columns
// and relies on this to leave absent columns (featured, halal, price, summary,
// dietary flags, …) untouched. A repository-level Select("*") on this shared
// method would wipe that marketing/dietary/translation data on a live route
// (H11 adversarial review).
func TestUpdate_SkipsZeroFields(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	now := time.Now()
	seedProduct(t, repo, &modelsProduct.Product{
		ID:             "p1",
		Slug:           "gummy-bears",
		Name:           "Gummy Bears",
		Summary:        "fruit gummies",
		Featured:       true,
		HalalCertified: true,
		OEMAvailable:   true,
		StockQuantity:  500,
		BasePrice:      12.5,
		Status:         modelsProduct.ProductStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	// Importer-style partial struct: only id + name + stock, all other fields
	// left at their zero value.
	partial := &modelsProduct.Product{
		ID:            "p1",
		Name:          "Gummy Bears V2",
		StockQuantity: 200,
		UpdatedAt:     time.Now(),
	}
	if err := repo.Update(ctx, partial); err != nil {
		t.Fatalf("update: %v", err)
	}

	reloaded, err := repo.FindByID(ctx, "p1")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.Name != "Gummy Bears V2" {
		t.Fatalf("name not updated: %q", reloaded.Name)
	}
	if reloaded.StockQuantity != 200 {
		t.Fatalf("stock not updated: %d", reloaded.StockQuantity)
	}
	if !reloaded.Featured {
		t.Fatal("featured was wiped by partial update")
	}
	if !reloaded.HalalCertified {
		t.Fatal("halal_certified was wiped by partial update")
	}
	if reloaded.Summary != "fruit gummies" {
		t.Fatalf("summary was wiped: %q", reloaded.Summary)
	}
	if reloaded.BasePrice != 12.5 {
		t.Fatalf("base_price was wiped: %v", reloaded.BasePrice)
	}
}

// TestUpdateAll_PersistsZeroValues H11: UpdateAll must persist zero values.
// The admin edit path loads the current row then patches pointer fields, so
// clearing featured/halal, zeroing stock/price, or emptying a text field must
// be written — GORM's default zero-skipping silently no-ops those changes while
// the UI reports success.
func TestUpdateAll_PersistsZeroValues(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	now := time.Now()
	seedProduct(t, repo, &modelsProduct.Product{
		ID:             "p1",
		Slug:           "gummy-bears",
		Name:           "Gummy Bears",
		Summary:        "fruit gummies",
		Featured:       true,
		HalalCertified: true,
		OEMAvailable:   true,
		StockQuantity:  500,
		SafetyStock:    50,
		BasePrice:      12.5,
		MOQ:            100,
		Status:         modelsProduct.ProductStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	loaded, err := repo.FindByID(ctx, "p1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}

	// Admin clears the featured flag, zeroes stock and base price, and empties
	// a string field that previously held text.
	loaded.Featured = false
	loaded.HalalCertified = false
	loaded.StockQuantity = 0
	loaded.BasePrice = 0
	loaded.Summary = ""
	loaded.UpdatedAt = time.Now()

	if err := repo.UpdateAll(ctx, loaded); err != nil {
		t.Fatalf("UpdateAll: %v", err)
	}

	reloaded, err := repo.FindByID(ctx, "p1")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.Featured {
		t.Fatal("expected featured=false to persist")
	}
	if reloaded.HalalCertified {
		t.Fatal("expected halal_certified=false to persist")
	}
	if reloaded.StockQuantity != 0 {
		t.Fatalf("expected stock_quantity=0, got %d", reloaded.StockQuantity)
	}
	if reloaded.BasePrice != 0 {
		t.Fatalf("expected base_price=0, got %v", reloaded.BasePrice)
	}
	if reloaded.Summary != "" {
		t.Fatalf("expected summary=\"\" to persist, got %q", reloaded.Summary)
	}
	// Unchanged fields survive the full-row write.
	if reloaded.Name != "Gummy Bears" || reloaded.Slug != "gummy-bears" || !reloaded.OEMAvailable {
		t.Fatalf("full-row write corrupted unchanged fields: %+v", reloaded)
	}
}

// TestUpdateColumns_WritesNamedColumns H11: the column-scoped variant writes
// only the named columns, including zero values, and leaves absent columns
// untouched. This is the safe path for a partial caller that must be able to
// clear/zero a specific field (the XLSX importer after migrating to pass its
// present-column list).
func TestUpdateColumns_WritesNamedColumns(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	now := time.Now()
	seedProduct(t, repo, &modelsProduct.Product{
		ID:             "p1",
		Slug:           "gummy-bears",
		Name:           "Gummy Bears",
		Summary:        "fruit gummies",
		Featured:       true,
		HalalCertified: true,
		StockQuantity:  500,
		BasePrice:      12.5,
		Status:         modelsProduct.ProductStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	patch := &modelsProduct.Product{
		ID:            "p1",
		Featured:      false,
		StockQuantity: 0,
	}
	if err := repo.UpdateColumns(ctx, patch, "featured", "stock_quantity"); err != nil {
		t.Fatalf("UpdateColumns: %v", err)
	}

	reloaded, err := repo.FindByID(ctx, "p1")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.Featured {
		t.Fatal("expected featured=false to persist via UpdateColumns")
	}
	if reloaded.StockQuantity != 0 {
		t.Fatalf("expected stock_quantity=0, got %d", reloaded.StockQuantity)
	}
	// Absent columns must stay untouched.
	if reloaded.Summary != "fruit gummies" {
		t.Fatalf("summary was touched: %q", reloaded.Summary)
	}
	if !reloaded.HalalCertified {
		t.Fatal("halal_certified was touched")
	}
	if reloaded.BasePrice != 12.5 {
		t.Fatalf("base_price was touched: %v", reloaded.BasePrice)
	}

	// An empty column list is rejected rather than silently writing a no-op.
	if err := repo.UpdateColumns(ctx, patch); err == nil {
		t.Fatal("expected error for empty column list")
	}
}

// setupEmbeddingTestDB migrates products + product_embeddings so the vector-write
// methods can be exercised. FK constraints are disabled because ProductEmbedding
// declares an OnDelete:CASCADE FK to products.
func setupEmbeddingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsProduct.Product{}, &modelsProduct.ProductEmbedding{}); err != nil {
		t.Fatalf("migrate products + product_embeddings: %v", err)
	}
	return db
}

// TestUpsertEmbedding_InsertsThenUpdates pins the G24b fix: UpsertEmbedding must
// write product_embeddings (previously the table was never written anywhere) and
// re-running it for the same product must update the single row, not duplicate.
func TestUpsertEmbedding_InsertsThenUpdates(t *testing.T) {
	db := setupEmbeddingTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	now := time.Now()
	seedProduct(t, repo, &modelsProduct.Product{
		ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears", Summary: "fruit gummies",
		Status: modelsProduct.ProductStatusActive, CreatedAt: now, UpdatedAt: now,
	})

	if err := repo.UpsertEmbedding(ctx, "p1", embeddingWithDim(modelsProduct.EmbeddingDim, 0.1, 0.2, 0.3)); err != nil {
		t.Fatalf("upsert embedding (insert): %v", err)
	}
	var row modelsProduct.ProductEmbedding
	if err := db.First(&row, "product_id = ?", "p1").Error; err != nil {
		t.Fatalf("load embedding: %v", err)
	}
	if row.Embedding == nil || len(row.Embedding.Slice()) != modelsProduct.EmbeddingDim {
		t.Fatalf("embedding not stored correctly: %+v", row.Embedding)
	}

	// Update path must upsert, not duplicate the row.
	if err := repo.UpsertEmbedding(ctx, "p1", embeddingWithDim(modelsProduct.EmbeddingDim, 0.9, 0.8, 0.7)); err != nil {
		t.Fatalf("upsert embedding (update): %v", err)
	}
	var count int64
	if err := db.Model(&modelsProduct.ProductEmbedding{}).Where("product_id = ?", "p1").Count(&count).Error; err != nil {
		t.Fatalf("count embeddings: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 embedding row after update-uppsert, got %d", count)
	}
	var updated modelsProduct.ProductEmbedding
	if err := db.First(&updated, "product_id = ?", "p1").Error; err != nil {
		t.Fatalf("reload embedding after update: %v", err)
	}
	if updated.Embedding == nil || len(updated.Embedding.Slice()) == 0 || updated.Embedding.Slice()[0] != 0.9 {
		t.Fatalf("embedding not updated: %+v", updated.Embedding)
	}
}

// TestFindProductsMissingEmbeddings pins the G24b backfill query: only active,
// non-deleted products without an embedding row are returned (draft products and
// already-embedded products are excluded).
func TestFindProductsMissingEmbeddings(t *testing.T) {
	db := setupEmbeddingTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	now := time.Now()
	seedProduct(t, repo, &modelsProduct.Product{
		ID: "p1", Slug: "a", Name: "A", Status: modelsProduct.ProductStatusActive, CreatedAt: now, UpdatedAt: now,
	})
	seedProduct(t, repo, &modelsProduct.Product{
		ID: "p2", Slug: "b", Name: "B", Status: modelsProduct.ProductStatusActive, CreatedAt: now, UpdatedAt: now,
	})
	seedProduct(t, repo, &modelsProduct.Product{
		ID: "p3", Slug: "c", Name: "C", Status: modelsProduct.ProductStatusDraft, CreatedAt: now, UpdatedAt: now,
	})

	if err := repo.UpsertEmbedding(ctx, "p1", embeddingWithDim(modelsProduct.EmbeddingDim, 0.1, 0.2)); err != nil {
		t.Fatalf("seed embedding for p1: %v", err)
	}

	missing, err := repo.FindProductsMissingEmbeddings(ctx, 10)
	if err != nil {
		t.Fatalf("FindProductsMissingEmbeddings: %v", err)
	}
	if len(missing) != 1 || missing[0].ID != "p2" {
		t.Fatalf("expected only p2 (active, no embedding), got %+v", missing)
	}
}

// TestUpsertEmbedding_RejectsWrongDimension pins the G24b dimension guard: an
// embedding whose length is not modelsProduct.EmbeddingDim must be rejected with
// a descriptive error and must not write a row. Previously this only failed on
// Postgres (vector(1536)) and was silently accepted by SQLite, leaving semantic
// search inert without a code error.
func TestUpsertEmbedding_RejectsWrongDimension(t *testing.T) {
	db := setupEmbeddingTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	now := time.Now()
	seedProduct(t, repo, &modelsProduct.Product{
		ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears", Status: modelsProduct.ProductStatusActive, CreatedAt: now, UpdatedAt: now,
	})

	err := repo.UpsertEmbedding(ctx, "p1", []float32{0.1, 0.2, 0.3})
	if err == nil {
		t.Fatal("expected dimension-mismatch error")
	}
	if !strings.Contains(err.Error(), "1536") {
		t.Fatalf("error should name the expected dimension, got: %v", err)
	}
	var count int64
	if err := db.Model(&modelsProduct.ProductEmbedding{}).Where("product_id = ?", "p1").Count(&count).Error; err != nil {
		t.Fatalf("count embeddings: %v", err)
	}
	if count != 0 {
		t.Fatalf("no row should be written on dimension mismatch, got %d", count)
	}
}

// TestDropLegacySlugUniqueIndex_UnblocksReuse pins the G24a deployed-DB fix: on
// an already-migrated database the pre-G24a full-table unique index
// (idx_products_slug) survives AutoMigrate and blocks re-creating a soft-deleted
// slug with a unique-constraint 5xx. DropLegacySlugUniqueIndex removes exactly
// that index, after which reuse succeeds on the partial index.
func TestDropLegacySlugUniqueIndex_UnblocksReuse(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	// Simulate an already-migrated deployment: the legacy full-table unique index
	// is still present alongside the new partial index.
	if err := db.Exec("CREATE UNIQUE INDEX idx_products_slug ON products(slug)").Error; err != nil {
		t.Fatalf("create legacy index: %v", err)
	}
	if !db.Migrator().HasIndex(&modelsProduct.Product{}, modelsProduct.LegacyUniqueIndexSlug) {
		t.Fatal("legacy index should be present on the migrated DB")
	}

	now := time.Now()
	first := &modelsProduct.Product{ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears", Status: modelsProduct.ProductStatusActive, CreatedAt: now, UpdatedAt: now}
	seedProduct(t, repo, first)
	if err := db.Delete(first).Error; err != nil {
		t.Fatalf("soft delete first product: %v", err)
	}

	second := &modelsProduct.Product{ID: "p2", Slug: "gummy-bears", Name: "Gummy Bears 2", Status: modelsProduct.ProductStatusActive, CreatedAt: now, UpdatedAt: now}
	if err := repo.Create(ctx, second); err == nil {
		t.Fatal("expected legacy full-table unique index to block slug reuse (pre-fix production symptom)")
	}

	if err := modelsProduct.DropLegacySlugUniqueIndex(db); err != nil {
		t.Fatalf("drop legacy index: %v", err)
	}
	if db.Migrator().HasIndex(&modelsProduct.Product{}, modelsProduct.LegacyUniqueIndexSlug) {
		t.Fatal("legacy index should be gone after drop")
	}

	if err := repo.Create(ctx, second); err != nil {
		t.Fatalf("reuse after dropping the legacy index must succeed: %v", err)
	}
}
