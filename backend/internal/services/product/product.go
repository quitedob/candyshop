package product

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	ordersvc "candypro/api/internal/services/order"
)

type productRepository interface {
	FindAll(ctx context.Context, page, limit int, categorySlug string) ([]modelsProduct.Product, int64, error)
	FindAllFiltered(ctx context.Context, page, limit int, halal, oemOnly, featuredOnly bool, search, sort string, minMOQ, maxMOQ int, categorySlug ...string) ([]modelsProduct.Product, int64, error)
	FindBySlug(ctx context.Context, slug string) (*modelsProduct.Product, error)
	FindByID(ctx context.Context, id string) (*modelsProduct.Product, error)
	FindFeatured(ctx context.Context, limit int) ([]modelsProduct.Product, error)
	FindRelated(ctx context.Context, slug string, limit int) ([]modelsProduct.Product, error)
	Create(ctx context.Context, product *modelsProduct.Product) error
	Update(ctx context.Context, product *modelsProduct.Product) error
	Delete(ctx context.Context, id string) error
	FindVariantsByProductID(ctx context.Context, productID string) ([]modelsProduct.ProductVariant, error)
	FindMarketProfilesForProducts(ctx context.Context, productIDs []string, marketCode string) ([]modelsProduct.ProductMarketProfile, error)
	UpsertProductMarketProfile(ctx context.Context, row *modelsProduct.ProductMarketProfile) error
	FindMarketCostStacksForProduct(ctx context.Context, productID string) ([]modelsProduct.ProductMarketCostStack, error)
	UpsertProductMarketCostStack(ctx context.Context, row *modelsProduct.ProductMarketCostStack) error
	ListWarehouses(ctx context.Context) ([]modelsProduct.Warehouse, error)
	SaveWarehouse(ctx context.Context, w *modelsProduct.Warehouse) error
	UpsertWarehouseStock(ctx context.Context, row *modelsProduct.WarehouseStock) error
	SaveOEMProjectInventoryHold(ctx context.Context, row *modelsProduct.OEMProjectInventoryHold) error
	ListOEMInventoryHoldsByProject(ctx context.Context, projectID string) ([]modelsProduct.OEMProjectInventoryHold, error)
	SumActiveOEMHoldsForProduct(ctx context.Context, productID string) (int64, error)
	ListChannelInventoriesForProduct(ctx context.Context, productID string) ([]modelsProduct.ChannelInventory, error)
	FindChannelInventory(ctx context.Context, productID, channelCode string) (*modelsProduct.ChannelInventory, error)
	UpsertChannelInventory(ctx context.Context, row *modelsProduct.ChannelInventory) error
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

// GetProduct returns a product by slug or ID (UUID fallback). Only active products are returned for public access.
func (s *ProductService) GetProduct(ctx context.Context, slugOrID string) (*modelsProduct.Product, error) {
	// Try slug first
	p, err := s.repo.FindBySlug(ctx, slugOrID)
	if err == nil {
		if p.Status != modelsProduct.ProductStatusActive {
			return nil, fmt.Errorf("product not found")
		}
		return p, nil
	}
	// Fallback: try by ID (UUID format)
	p, err = s.repo.FindByID(ctx, slugOrID)
	if err == nil {
		if p.Status != modelsProduct.ProductStatusActive {
			return nil, fmt.Errorf("product not found")
		}
		return p, nil
	}
	return nil, err
}

// GetProductByID returns a product by ID.
func (s *ProductService) GetProductByID(ctx context.Context, id string) (*modelsProduct.Product, error) {
	return s.repo.FindByID(ctx, id)
}

// IncrementViewCount increments the view count for a product (best-effort).
func (s *ProductService) IncrementViewCount(ctx context.Context, productID string) error {
	product, err := s.repo.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	product.ViewCount++
	return s.repo.Update(ctx, product)
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

// ValidateComplianceWithMarketProfiles 目的国合规：基线规则 + 可配置市场画像
func (s *ProductService) ValidateComplianceWithMarketProfiles(ctx context.Context, country string, products []modelsProduct.Product) ordersvc.ComplianceValidationResult {
	ids := make([]string, 0, len(products))
	for _, p := range products {
		if id := strings.TrimSpace(p.ID); id != "" {
			ids = append(ids, id)
		}
	}
	mc := ordersvc.ComplianceProfileMarketCode(country)
	profiles, err := s.repo.FindMarketProfilesForProducts(ctx, ids, mc)
	if err != nil || len(profiles) == 0 {
		return ordersvc.ValidateCountryComplianceRules(country, products)
	}
	return ordersvc.ValidateCountryComplianceRulesWithProfiles(country, products, profiles)
}

// ListWarehouses 仓库列表（跨境）
func (s *ProductService) ListWarehouses(ctx context.Context) ([]modelsProduct.Warehouse, error) {
	return s.repo.ListWarehouses(ctx)
}

// SaveWarehouse 保存仓库
func (s *ProductService) SaveWarehouse(ctx context.Context, w *modelsProduct.Warehouse) error {
	return s.repo.SaveWarehouse(ctx, w)
}

// UpsertWarehouseStock 写入仓库存
func (s *ProductService) UpsertWarehouseStock(ctx context.Context, row *modelsProduct.WarehouseStock) error {
	return s.repo.UpsertWarehouseStock(ctx, row)
}

// FindMarketProfilesForProducts 查询市场画像
func (s *ProductService) FindMarketProfilesForProducts(ctx context.Context, productIDs []string, marketCode string) ([]modelsProduct.ProductMarketProfile, error) {
	return s.repo.FindMarketProfilesForProducts(ctx, productIDs, marketCode)
}

// UpsertProductMarketProfile 保存市场画像
func (s *ProductService) UpsertProductMarketProfile(ctx context.Context, row *modelsProduct.ProductMarketProfile) error {
	return s.repo.UpsertProductMarketProfile(ctx, row)
}

// FindMarketCostStacksForProduct 目的国成本栈
func (s *ProductService) FindMarketCostStacksForProduct(ctx context.Context, productID string) ([]modelsProduct.ProductMarketCostStack, error) {
	return s.repo.FindMarketCostStacksForProduct(ctx, productID)
}

// UpsertProductMarketCostStack 保存成本栈
func (s *ProductService) UpsertProductMarketCostStack(ctx context.Context, row *modelsProduct.ProductMarketCostStack) error {
	return s.repo.UpsertProductMarketCostStack(ctx, row)
}

// SaveOEMProjectInventoryHold OEM 成品预留
func (s *ProductService) SaveOEMProjectInventoryHold(ctx context.Context, row *modelsProduct.OEMProjectInventoryHold) error {
	return s.repo.SaveOEMProjectInventoryHold(ctx, row)
}

// ListOEMInventoryHoldsByProject 列出项目预留
func (s *ProductService) ListOEMInventoryHoldsByProject(ctx context.Context, projectID string) ([]modelsProduct.OEMProjectInventoryHold, error) {
	return s.repo.ListOEMInventoryHoldsByProject(ctx, projectID)
}

// BuildPricingCostContextJSON 供 AI 报价注入的成本栈摘要（只读）
func (s *ProductService) BuildPricingCostContextJSON(ctx context.Context, productID, marketCode string) (string, error) {
	stacks, err := s.repo.FindMarketCostStacksForProduct(ctx, productID)
	if err != nil {
		return "", err
	}
	mc := strings.ToUpper(strings.TrimSpace(marketCode))
	var pick *modelsProduct.ProductMarketCostStack
	for i := range stacks {
		if strings.ToUpper(strings.TrimSpace(stacks[i].MarketCode)) == mc {
			pick = &stacks[i]
			break
		}
	}
	if pick == nil && len(stacks) > 0 {
		pick = &stacks[0]
	}
	if pick == nil {
		return "{}", nil
	}
	b, err := json.Marshal(pick)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ApplyProductTranslations overlays locale-specific translations onto a product's display fields.
// Fields without a translation fall back to the original stored value.
func ApplyProductTranslations(product *modelsProduct.Product, locale string) {
	if product == nil || product.Translations == nil || len(product.Translations) == 0 {
		return
	}
	t, ok := product.Translations[locale]
	if !ok {
		return
	}
	if v := t["name"]; v != "" {
		product.Name = v
	}
	if v := t["summary"]; v != "" {
		product.Summary = v
	}
	if v := t["description"]; v != "" {
		product.Description = v
	}
	if v := t["category"]; v != "" {
		product.Category = v
	}
	if v := t["ingredients"]; v != "" {
		product.Ingredients = v
	}
	if v := t["allergens"]; v != "" {
		product.Allergens = v
	}
	if v := t["storage"]; v != "" {
		product.Storage = v
	}
	if v := t["leadTime"]; v != "" {
		product.LeadTime = v
	}
	if v := t["shelfLife"]; v != "" {
		product.ShelfLife = v
	}
	if v := t["flavors"]; v != "" {
		var arr []string
		if err := json.Unmarshal([]byte(v), &arr); err == nil {
			product.Flavors = arr
		}
	}
	if v := t["shapes"]; v != "" {
		var arr []string
		if err := json.Unmarshal([]byte(v), &arr); err == nil {
			product.Shapes = arr
		}
	}
}

// ApplyProductTranslationsBatch applies translations to a slice of products.
func ApplyProductTranslationsBatch(products []modelsProduct.Product, locale string) {
	for i := range products {
		ApplyProductTranslations(&products[i], locale)
	}
}
