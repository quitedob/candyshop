package product

import (
	"context"
	"strings"

	modelsProduct "candypro/api/internal/models/product"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
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
	if len(categorySlug) > 0 && categorySlug[0] != "" {
		query = query.Where("category_slug = ?", categorySlug[0])
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

// FindFeatured returns featured products
func (r *ProductRepository) FindFeatured(ctx context.Context, limit int) ([]modelsProduct.Product, error) {
	var products []modelsProduct.Product
	if err := r.db.WithContext(ctx).Where("featured = ?", true).Limit(limit).Order("created_at DESC").Find(&products).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Where("category_slug = ? AND slug != ?", product.CategorySlug, slug).
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
		WHERE p.status = 'active' AND pe.embedding IS NOT NULL
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

// Update updates a product
func (r *ProductRepository) Update(ctx context.Context, product *modelsProduct.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// Delete deletes a product
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.Product{}, "id = ?", id).Error
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
