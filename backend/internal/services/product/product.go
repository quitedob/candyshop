package product

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
)

type productRepository interface {
	FindAll(ctx context.Context, page, limit int, categorySlug string) ([]modelsProduct.Product, int64, error)
	FindAllFiltered(ctx context.Context, page, limit int, halal, oemOnly, featuredOnly bool, search, sort string, minMOQ, maxMOQ int) ([]modelsProduct.Product, int64, error)
	FindBySlug(ctx context.Context, slug string) (*modelsProduct.Product, error)
	FindByID(ctx context.Context, id string) (*modelsProduct.Product, error)
	FindFeatured(ctx context.Context, limit int) ([]modelsProduct.Product, error)
	FindRelated(ctx context.Context, slug string, limit int) ([]modelsProduct.Product, error)
	Create(ctx context.Context, product *modelsProduct.Product) error
	Update(ctx context.Context, product *modelsProduct.Product) error
	Delete(ctx context.Context, id string) error
	FindVariantsByProductID(ctx context.Context, productID string) ([]modelsProduct.ProductVariant, error)
}

// ProductService handles product business logic.
type ProductService struct {
	repo productRepository
}

// NewProductService creates a new ProductService.
func NewProductService(repo productRepository) *ProductService {
	return &ProductService{repo: repo}
}

// GetProducts returns paginated products.
func (s *ProductService) GetProducts(ctx context.Context, page, limit int, categorySlug string) (*modelsProduct.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 12
	}

	products, total, err := s.repo.FindAll(ctx, page, limit, categorySlug)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &modelsProduct.PaginatedResponse{
		Data: products,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// GetProductsFiltered returns paginated products with full filter support.
func (s *ProductService) GetProductsFiltered(ctx context.Context, page, limit int, halal, oemOnly, featuredOnly bool, search, sort string, minMOQ, maxMOQ int) (*modelsProduct.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 12
	}

	products, total, err := s.repo.FindAllFiltered(ctx, page, limit, halal, oemOnly, featuredOnly, search, sort, minMOQ, maxMOQ)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &modelsProduct.PaginatedResponse{
		Data: products,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// GetProduct returns a product by slug or ID (UUID fallback).
func (s *ProductService) GetProduct(ctx context.Context, slugOrID string) (*modelsProduct.Product, error) {
	// Try slug first
	p, err := s.repo.FindBySlug(ctx, slugOrID)
	if err == nil {
		return p, nil
	}
	// Fallback: try by ID (UUID format)
	return s.repo.FindByID(ctx, slugOrID)
}

// GetProductByID returns a product by ID.
func (s *ProductService) GetProductByID(ctx context.Context, id string) (*modelsProduct.Product, error) {
	return s.repo.FindByID(ctx, id)
}

// GetFeaturedProducts returns featured products.
func (s *ProductService) GetFeaturedProducts(ctx context.Context, limit int) ([]modelsProduct.Product, error) {
	if limit <= 0 {
		limit = 8
	}
	return s.repo.FindFeatured(ctx, limit)
}

// GetRelatedProducts returns related products.
func (s *ProductService) GetRelatedProducts(ctx context.Context, slug string, limit int) ([]modelsProduct.Product, error) {
	if limit <= 0 {
		limit = 4
	}
	return s.repo.FindRelated(ctx, slug, limit)
}

// CreateProduct creates a new product.
func (s *ProductService) CreateProduct(ctx context.Context, product *modelsProduct.Product) error {
	return s.repo.Create(ctx, product)
}

// UpdateProduct updates a product.
func (s *ProductService) UpdateProduct(ctx context.Context, product *modelsProduct.Product) error {
	return s.repo.Update(ctx, product)
}

// DeleteProduct deletes a product.
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetProductVariants returns active SKU variants for a product.
func (s *ProductService) GetProductVariants(ctx context.Context, productID string) ([]modelsProduct.ProductVariant, error) {
	return s.repo.FindVariantsByProductID(ctx, productID)
}
