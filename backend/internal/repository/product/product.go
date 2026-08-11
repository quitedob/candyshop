package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	modelsProduct "candypro/api/internal/models/product"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProductRepository handles product data operations
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// FindAll returns paginated products with optional filters.
func (r *ProductRepository) FindAll(ctx context.Context, page, limit int, categorySlug string) ([]modelsProduct.Product, int64, error) {
	return r.FindAllFiltered(ctx, page, limit, false, false, false, "", "", 0, 0, categorySlug)
}

// FindAllFiltered returns paginated products with full filter support.
func (r *ProductRepository) FindAllFiltered(ctx context.Context, page, limit int, halal, oemOnly, featuredOnly bool, search, sort string, minMOQ, maxMOQ int, categorySlug ...string) ([]modelsProduct.Product, int64, error) {
	var products []modelsProduct.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsProduct.Product{}).Where("status = ?", "active")

	if search != "" {
		pattern := "%" + escapeLikePattern(search) + "%"
		query = query.Where("name ILIKE ? OR summary ILIKE ? OR description ILIKE ?", pattern, pattern, pattern)
	}
	if halal {
		query = query.Where("halal_certified = true")
	}
	if oemOnly {
		query = query.Where("oem_available = true")
	}
	if featuredOnly {
		query = query.Where("featured = true")
	}
	if len(categorySlug) > 0 {
		slugs := make([]string, 0, len(categorySlug))
		for _, s := range categorySlug {
			if s = strings.TrimSpace(s); s != "" {
				slugs = append(slugs, s)
			}
		}
		switch len(slugs) {
		case 1:
			query = query.Where("category_slug = ?", slugs[0])
		case 0:
			// no-op
		default:
			query = query.Where("category_slug IN ?", slugs)
		}
	}
	if minMOQ > 0 {
		query = query.Where("moq >= ?", minMOQ)
	}
	if maxMOQ > 0 {
		query = query.Where("moq <= ?", maxMOQ)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	switch sort {
	case "name":
		query = query.Order("name ASC")
	case "moq_low":
		query = query.Order("moq ASC")
	case "moq_high":
		query = query.Order("moq DESC")
	default: // popular
		query = query.Order("view_count DESC, created_at DESC")
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// FindAllForAdmin 管理员产品列表，支持 status 筛选（空则不过滤）
func (r *ProductRepository) FindAllForAdmin(ctx context.Context, page, limit int, categorySlug, status, search string) ([]modelsProduct.Product, int64, error) {
	var products []modelsProduct.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsProduct.Product{})
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "":
		// 默认排除 draft/inactive，避免运营列表混入未发布商品
		query = query.Where("status NOT IN ?", []string{"draft", "inactive"})
	case "all":
		// 显式请求全部状态
	default:
		query = query.Where("status = ?", status)
	}
	if search != "" {
		pattern := "%" + escapeLikePattern(search) + "%"
		query = query.Where("name ILIKE ? OR summary ILIKE ? OR description ILIKE ?", pattern, pattern, pattern)
	}
	if categorySlug != "" {
		query = query.Where("category_slug = ?", categorySlug)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

// FindBySlug returns a product by slug
func (r *ProductRepository) FindBySlug(ctx context.Context, slug string) (*modelsProduct.Product, error) {
	var product modelsProduct.Product
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// FindByID returns a product by ID.
func (r *ProductRepository) FindByID(ctx context.Context, id string) (*modelsProduct.Product, error) {
	var product modelsProduct.Product
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// FindByIDs returns the products matching the given IDs in a single query.
// Used by hot paths (checkout / order confirmation) to avoid N+1 lookups
// where one query was previously issued per order line (H-1, H-2, H-3).
//
// Soft-deleted products are excluded automatically by GORM. Caller should
// build a productByID map from the result and tolerate missing IDs.
func (r *ProductRepository) FindByIDs(ctx context.Context, ids []string) ([]modelsProduct.Product, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var products []modelsProduct.Product
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// FindFeatured returns featured products
func (r *ProductRepository) FindFeatured(ctx context.Context, limit int) ([]modelsProduct.Product, error) {
	var products []modelsProduct.Product
	if err := r.db.WithContext(ctx).Where("featured = ?", true).Where("status = ?", "active").Limit(limit).Order("created_at DESC").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// FindRelated returns related products by category
func (r *ProductRepository) FindRelated(ctx context.Context, slug string, limit int) ([]modelsProduct.Product, error) {
	var product modelsProduct.Product
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&product).Error; err != nil {
		return nil, err
	}

	var products []modelsProduct.Product
	if err := r.db.WithContext(ctx).Where("category_slug = ? AND slug != ? AND status = ?", product.CategorySlug, slug, "active").
		Limit(limit).Order("created_at DESC").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// Search searches products by query
func (r *ProductRepository) Search(ctx context.Context, query string, limit int) ([]modelsProduct.Product, error) {
	var products []modelsProduct.Product
	searchPattern := "%" + escapeLikePattern(query) + "%"
	if err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Where("name ILIKE ? OR summary ILIKE ? OR description ILIKE ?", searchPattern, searchPattern, searchPattern).
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// SearchVectorSimilar uses pgvector cosine distance (<=>) against the optional product_embeddings table.
func (r *ProductRepository) SearchVectorSimilar(ctx context.Context, embedding []float32, limit int) ([]modelsProduct.Product, error) {
	if len(embedding) == 0 || limit <= 0 {
		return nil, nil
	}

	vec := pgvector.NewVector(embedding)
	var products []modelsProduct.Product
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.*
		FROM product_embeddings pe
		JOIN products p ON p.id = pe.product_id
		WHERE p.status = 'active'
			AND p.deleted_at IS NULL
			AND pe.embedding IS NOT NULL
		ORDER BY pe.embedding <=> ?::vector
		LIMIT ?`, vec, limit).Scan(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

// Create creates a new product
func (r *ProductRepository) Create(ctx context.Context, product *modelsProduct.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// Update updates a product using GORM's struct-based Updates, which skips
// zero-valued fields. This preserves partial-update semantics: a caller that
// builds a fresh struct from a subset of columns (e.g. the XLSX importer,
// AdminApplyInventoryImport) leaves absent columns untouched.
//
// Callers that must persist zero values — clearing featured / halal_certified,
// zeroing stock_quantity or base_price, emptying a text field (H11) — must use
// UpdateAll (full-row overwrite) or UpdateColumns (explicit column list)
// instead; GORM's zero-skipping otherwise silently no-ops the change while the
// caller reports success.
//
// WARNING (H11 end-to-end status, 2026-08-11): the live admin edit paths
// AdminUpdateProduct (admin_products.go), AdminUpdateProductStatus
// (admin_extensions.go), AdminBatchUpdateInventory (admin_inventory.go) and
// AdminAITranslate (admin_ai_translate.go) still call this zero-skip Update via
// ProductService.UpdateProduct, so clearing featured/halal, zeroing
// stock/price, or emptying a text field from the admin UI still silently no-ops
// until those handlers migrate to UpdateAll. Do not add new callers here that
// require zero-value persistence.
func (r *ProductRepository) Update(ctx context.Context, product *modelsProduct.Product) error {
	return r.db.WithContext(ctx).Model(product).Updates(product).Error
}

// UpdateAll performs a full-row overwrite, persisting zero values.
//
// GORM's struct-based Updates skips zero-valued fields, so an admin clearing
// featured=false or halal_certified=false, zeroing stock_quantity or
// base_price, or emptying a text field would silently no-op while the UI
// reports success (H11). Select("*") writes every column of the struct, zero
// values included.
//
// Safe for the load-then-patch callers (AdminUpdateProduct,
// AdminUpdateProductStatus, the batch inventory update, the AI translate
// handler) that read the current row first and patch only the changed fields:
// every column is preserved from the loaded row and the patched zero values are
// written. Do NOT use it with a fresh struct built from a partial column set —
// the XLSX importer must merge the full existing row first or use UpdateColumns.
//
// Select("*") includes created_at and deleted_at in the UPDATE column set. For
// a non-deleted row that writes the existing values (NULL for deleted_at) and
// is harmless; soft-deleted rows are excluded from updates by GORM's default
// scope.
func (r *ProductRepository) UpdateAll(ctx context.Context, product *modelsProduct.Product) error {
	return r.db.WithContext(ctx).Model(product).Select("*").Updates(product).Error
}

// UpdateColumns writes only the named DB columns from the struct, including
// zero values. It is the column-scoped variant for partial updates that must be
// able to clear or zero a specific field: pass the columns that were actually
// supplied by the caller (add "updated_at" explicitly to bump the timestamp) so
// absent columns stay untouched. At least one column is required.
func (r *ProductRepository) UpdateColumns(ctx context.Context, product *modelsProduct.Product, columns ...string) error {
	if len(columns) == 0 {
		return errors.New("product: UpdateColumns requires at least one column")
	}
	return r.db.WithContext(ctx).Model(product).Select(columns).Updates(product).Error
}

// UpdateStockWithLock 行锁更新库存数量，返回旧库存
func (r *ProductRepository) UpdateStockWithLock(ctx context.Context, productID string, newQty int) (int, error) {
	var oldQty int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product modelsProduct.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", productID).First(&product).Error; err != nil {
			return err
		}
		oldQty = product.StockQuantity
		now := time.Now()
		return tx.Model(&product).Updates(map[string]interface{}{
			"stock_quantity": newQty,
			"updated_at":     now,
		}).Error
	})
	return oldQty, err
}

// Delete deletes a product
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.Product{}, "id = ?", id).Error
}

// UpsertEmbedding inserts or updates a product's semantic-search vector in the
// product_embeddings table (G24b). product_embeddings was previously never
// written anywhere, which left the pgvector JOIN in SearchVectorSimilar
// perpetually empty and semantic search silently falling back to keyword on
// every request. Callers that generate vectors (the search-service backfill)
// persist them here. A missing/blank product ID or an empty embedding is a
// no-op.
//
// The embedding must be exactly modelsProduct.EmbeddingDim (1536) floats: the
// column is declared vector(1536), so a different dimension can never be stored
// on Postgres and, worse, is silently accepted by SQLite. Validating here makes
// a misconfigured embedding model loud and actionable instead of reproducing
// the inert-semantic-search symptom without a code error (G24b).
func (r *ProductRepository) UpsertEmbedding(ctx context.Context, productID string, embedding []float32) error {
	if strings.TrimSpace(productID) == "" || len(embedding) == 0 {
		return nil
	}
	if len(embedding) != modelsProduct.EmbeddingDim {
		return fmt.Errorf(
			"product: embedding for %q has %d dimensions, want %d (vector(%d)); configure an embedding model with %d-dim output or migrate the product_embeddings column",
			productID, len(embedding), modelsProduct.EmbeddingDim, modelsProduct.EmbeddingDim, modelsProduct.EmbeddingDim)
	}
	vec := pgvector.NewVector(embedding)
	now := time.Now()
	row := modelsProduct.ProductEmbedding{
		ProductID: productID,
		Embedding: &vec,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "product_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"embedding", "updated_at"}),
	}).Create(&row).Error
}

// FindProductsMissingEmbeddings returns up to limit active, non-deleted products
// that have no row in product_embeddings. Used by the search-service backfill so
// the semantic index stays populated for newly created products.
func (r *ProductRepository) FindProductsMissingEmbeddings(ctx context.Context, limit int) ([]modelsProduct.Product, error) {
	if limit <= 0 {
		return nil, nil
	}
	var products []modelsProduct.Product
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Where("id NOT IN (?)", r.db.WithContext(ctx).Model(&modelsProduct.ProductEmbedding{}).Select("product_id")).
		Limit(limit).
		Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

// escapeLikePattern escapes special characters for SQL LIKE operator
func escapeLikePattern(pattern string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"%", "\\%",
		"_", "\\_",
	)
	return replacer.Replace(pattern)
}

// FindVariantsByProductID returns all active variants for a product.
func (r *ProductRepository) FindVariantsByProductID(ctx context.Context, productID string) ([]modelsProduct.ProductVariant, error) {
	var variants []modelsProduct.ProductVariant
	if err := r.db.WithContext(ctx).Where("product_id = ? AND is_active = true", productID).
		Order("flavor, weight").Find(&variants).Error; err != nil {
		return nil, err
	}
	return variants, nil
}

// ComputeWeightedAvgCost 按批次加权平均计算产品单位成本
//
// R2 E-11: previously this loaded every batch into Go memory and folded over
// them. We now push SUM(quantity * unit_cost) / NULLIF(SUM(quantity), 0)
// into PostgreSQL — one round-trip, no per-batch row materialisation. Falls
// back to the cached Product.WeightedAvgCost when no eligible batches exist
// (e.g. a brand-new SKU with no PO yet).
func (r *ProductRepository) ComputeWeightedAvgCost(ctx context.Context, productID string) float64 {
	var avg sql.NullFloat64
	err := r.db.WithContext(ctx).
		Model(&modelsProduct.ProductBatch{}).
		Where("product_id = ? AND quantity > 0 AND is_expired = false", productID).
		Select("SUM(quantity * unit_cost) / NULLIF(SUM(quantity), 0)").
		Scan(&avg).Error
	if err == nil && avg.Valid {
		return avg.Float64
	}
	var product modelsProduct.Product
	if err := r.db.WithContext(ctx).Select("weighted_avg_cost").Where("id = ?", productID).First(&product).Error; err == nil {
		return product.WeightedAvgCost
	}
	return 0
}
