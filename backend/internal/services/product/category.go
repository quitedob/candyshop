package product

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
)

type categoryRepository interface {
	FindAll(ctx context.Context) ([]modelsProduct.Category, error)
	FindBySlug(ctx context.Context, slug string) (*modelsProduct.Category, error)
	FindBySlugWithProducts(ctx context.Context, slug string, page, limit int) (*modelsProduct.Category, []modelsProduct.Product, error)
	Create(ctx context.Context, category *modelsProduct.Category) error
	Update(ctx context.Context, category *modelsProduct.Category) error
	Delete(ctx context.Context, slug string) error
}

// CategoryService handles category business logic.
type CategoryService struct {
	repo categoryRepository
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(repo categoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

// GetCategories returns all categories.
func (s *CategoryService) GetCategories(ctx context.Context) ([]modelsProduct.Category, error) {
	return s.repo.FindAll(ctx)
}

// GetCategory returns a category by slug.
func (s *CategoryService) GetCategory(ctx context.Context, slug string) (*modelsProduct.Category, error) {
	return s.repo.FindBySlug(ctx, slug)
}

// GetCategoryWithProducts returns a category with its products.
func (s *CategoryService) GetCategoryWithProducts(ctx context.Context, slug string, page, limit int) (*modelsProduct.Category, []modelsProduct.Product, error) {
	return s.repo.FindBySlugWithProducts(ctx, slug, page, limit)
}

// CreateCategory creates a new category.
func (s *CategoryService) CreateCategory(ctx context.Context, category *modelsProduct.Category) error {
	return s.repo.Create(ctx, category)
}

// UpdateCategory updates a category.
func (s *CategoryService) UpdateCategory(ctx context.Context, category *modelsProduct.Category) error {
	return s.repo.Update(ctx, category)
}

// DeleteCategory deletes a category.
func (s *CategoryService) DeleteCategory(ctx context.Context, slug string) error {
	return s.repo.Delete(ctx, slug)
}
