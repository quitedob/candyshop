package product

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"

	"gorm.io/gorm"
)

// CategoryRepository handles category data operations
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// FindAll returns all categories
func (r *CategoryRepository) FindAll(ctx context.Context) ([]modelsProduct.Category, error) {
	var categories []modelsProduct.Category
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	// Count products for each category (active only)
	for i := range categories {
		var count int64
		r.db.WithContext(ctx).Model(&modelsProduct.Product{}).Where("category_slug = ? AND status = ?", categories[i].Slug, "active").Count(&count)
		categories[i].ProductCount = int(count)
	}

	return categories, nil
}

// FindBySlug returns a category by slug with products
func (r *CategoryRepository) FindBySlug(ctx context.Context, slug string) (*modelsProduct.Category, error) {
	var category modelsProduct.Category
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&category).Error; err != nil {
		return nil, err
	}

	// Count products (active only)
	var count int64
	r.db.WithContext(ctx).Model(&modelsProduct.Product{}).Where("category_slug = ? AND status = ?", slug, "active").Count(&count)
	category.ProductCount = int(count)

	return &category, nil
}

// FindBySlugWithProducts returns a category by slug with its products
func (r *CategoryRepository) FindBySlugWithProducts(ctx context.Context, slug string, page, limit int) (*modelsProduct.Category, []modelsProduct.Product, error) {
	category, err := r.FindBySlug(ctx, slug)
	if err != nil {
		return nil, nil, err
	}

	var products []modelsProduct.Product
	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).Where("category_slug = ? AND status = ?", slug, "active").
		Offset(offset).Limit(limit).Order("created_at DESC").
		Find(&products).Error; err != nil {
		return nil, nil, err
	}

	return category, products, nil
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *modelsProduct.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, category *modelsProduct.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, slug string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.Category{}, "slug = ?", slug).Error
}
