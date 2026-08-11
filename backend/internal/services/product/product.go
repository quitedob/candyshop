package product

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/catalog"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	ordersvc "candypro/api/internal/services/order"
)

type productRepository interface {
	FindAll(ctx context.Context, page, limit int, categorySlug string) ([]modelsProduct.Product, int64, error)
	FindAllForAdmin(ctx context.Context, page, limit int, categorySlug, status, search string) ([]modelsProduct.Product, int64, error)
	FindAllFiltered(ctx context.Context, page, limit int, halal, oemOnly, featuredOnly bool, search, sort string, minMOQ, maxMOQ int, categorySlug ...string) ([]modelsProduct.Product, int64, error)
	FindBySlug(ctx context.Context, slug string) (*modelsProduct.Product, error)
	FindByID(ctx context.Context, id string) (*modelsProduct.Product, error)
	FindByIDs(ctx context.Context, ids []string) ([]modelsProduct.Product, error)
	FindFeatured(ctx context.Context, limit int) ([]modelsProduct.Product, error)
	FindRelated(ctx context.Context, slug string, limit int) ([]modelsProduct.Product, error)
	Create(ctx context.Context, product *modelsProduct.Product) error
	Update(ctx context.Context, product *modelsProduct.Product) error
	UpdateStockWithLock(ctx context.Context, productID string, newQty int) (int, error)
	Delete(ctx context.Context, id string) error
	FindVariantsByProductID(ctx context.Context, productID string) ([]modelsProduct.ProductVariant, error)
	FindMarketProfilesForProducts(ctx context.Context, productIDs []string, marketCode string) ([]modelsProduct.ProductMarketProfile, error)
	UpsertProductMarketProfile(ctx context.Context, row *modelsProduct.ProductMarketProfile) error
	FindMarketCostStacksForProduct(ctx context.Context, productID string) ([]modelsProduct.ProductMarketCostStack, error)
	UpsertProductMarketCostStack(ctx context.Context, row *modelsProduct.ProductMarketCostStack) error
	ListWarehouses(ctx context.Context) ([]modelsProduct.Warehouse, error)
	GetDefaultWarehouseID(ctx context.Context) (string, error)
	SaveWarehouse(ctx context.Context, w *modelsProduct.Warehouse) error
	UpsertWarehouseStock(ctx context.Context, row *modelsProduct.WarehouseStock) error
	SaveOEMProjectInventoryHold(ctx context.Context, row *modelsProduct.OEMProjectInventoryHold) error
	ListOEMInventoryHoldsByProject(ctx context.Context, projectID string) ([]modelsProduct.OEMProjectInventoryHold, error)
	UpdateOEMHoldStatusIfMatches(ctx context.Context, id uint, expected, target string) (int64, error)
	SumActiveOEMHoldsForProduct(ctx context.Context, productID string) (int64, error)
	SumActiveOEMHoldsByProductIDs(ctx context.Context, productIDs []string) (map[string]int64, error)
	ListChannelInventoriesForProduct(ctx context.Context, productID string) ([]modelsProduct.ChannelInventory, error)
	FindChannelInventory(ctx context.Context, productID, channelCode string) (*modelsProduct.ChannelInventory, error)
	FindChannelInventoriesByProductIDs(ctx context.Context, productIDs []string, channelCode string) (map[string]*modelsProduct.ChannelInventory, error)
	UpsertChannelInventory(ctx context.Context, row *modelsProduct.ChannelInventory) error
	ComputeWeightedAvgCost(ctx context.Context, productID string) float64
}

// productZeroValueRepository is the subset of ProductRepository that can write
// zero values (H11): clearing featured/halal, zeroing stock/price, or emptying
// a text field. It is kept separate from productRepository so existing test
// fakes that only model the partial-update surface keep compiling;
// ProductService type-asserts the injected repository at construction and
// exposes the zero-capable paths when available (always true for the
// production *ProductRepository, which implements both).
type productZeroValueRepository interface {
	UpdateAll(ctx context.Context, product *modelsProduct.Product) error
	UpdateColumns(ctx context.Context, product *modelsProduct.Product, columns ...string) error
}

// ProductService handles product business logic.
type ProductService struct {
	repo productRepository
	zero productZeroValueRepository
}

// NewProductService creates a new ProductService.
func NewProductService(repo productRepository) *ProductService {
	s := &ProductService{repo: repo}
	if z, ok := repo.(productZeroValueRepository); ok {
		s.zero = z
	}
	return s
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
func (s *ProductService) GetProductsFiltered(ctx context.Context, page, limit int, halal, oemOnly, featuredOnly bool, search, sort string, minMOQ, maxMOQ int, categorySlugs ...string) (*modelsProduct.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 12
	}

	products, total, err := s.repo.FindAllFiltered(ctx, page, limit, halal, oemOnly, featuredOnly, search, sort, minMOQ, maxMOQ, categorySlugs...)
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

// GetProductsForAdmin 管理员产品列表，支持 status 筛选
func (s *ProductService) GetProductsForAdmin(ctx context.Context, page, limit int, categorySlug, status, search string) (*modelsProduct.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	products, total, err := s.repo.FindAllForAdmin(ctx, page, limit, categorySlug, status, search)
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

// UpdateStockWithLock 行锁更新库存
func (s *ProductService) UpdateStockWithLock(ctx context.Context, productID string, newQty int) (int, error) {
	return s.repo.UpdateStockWithLock(ctx, productID, newQty)
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

// GetProductsByIDs returns the products matching the given IDs in a single batch.
// Order is not guaranteed; callers wanting a map should index the result themselves.
//
// Use this in hot paths (checkout, order confirm, COGS) instead of looping
// GetProductByID per line — that pattern was N+1 (H-1, H-2, H-3).
func (s *ProductService) GetProductsByIDs(ctx context.Context, ids []string) ([]modelsProduct.Product, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// Deduplicate to avoid sending duplicate IDs to the DB.
	seen := make(map[string]struct{}, len(ids))
	unique := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return s.repo.FindByIDs(ctx, unique)
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

// UpdateProduct updates a product using the partial, zero-skip Update.
func (s *ProductService) UpdateProduct(ctx context.Context, product *modelsProduct.Product) error {
	return s.repo.Update(ctx, product)
}

// UpdateProductAll writes the product with a full-row overwrite that persists
// zero values. The admin load-then-patch callers (AdminUpdateProduct,
// AdminUpdateProductStatus, AdminBatchUpdateInventory, AdminAITranslateProduct)
// load the current row first, so clearing featured/halal, zeroing stock/price,
// or emptying a text field must be written (H11) — the zero-skip Update would
// silently drop those changes while the UI reports success. Do NOT use with a
// fresh struct built from a partial column set.
func (s *ProductService) UpdateProductAll(ctx context.Context, product *modelsProduct.Product) error {
	if s.zero == nil {
		return fmt.Errorf("product: repository does not support zero-value full-row updates")
	}
	return s.zero.UpdateAll(ctx, product)
}

// UpdateProductColumns writes only the named DB columns, including zero values.
// Partial callers that must be able to clear/zero a specific field (the XLSX
// importer) pass the columns actually present in the row, so absent columns
// stay untouched while a carried zero value (e.g. stockQuantity:0) is persisted
// (H11). Add "updated_at" explicitly to bump the timestamp.
func (s *ProductService) UpdateProductColumns(ctx context.Context, product *modelsProduct.Product, columns ...string) error {
	if s.zero == nil {
		return fmt.Errorf("product: repository does not support column-scoped updates")
	}
	return s.zero.UpdateColumns(ctx, product, columns...)
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

// GetDefaultWarehouseID 返回默认仓库 ID
func (s *ProductService) GetDefaultWarehouseID(ctx context.Context) (string, error) {
	return s.repo.GetDefaultWarehouseID(ctx)
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

// ReleaseOEMInventoryHold 释放（取消）一条 active 的 OEM 成品预留，返回是否生效。
// 仅当当前状态为 active 时成功，避免重复释放（P0.2 / G-OEM-3）。
func (s *ProductService) ReleaseOEMInventoryHold(ctx context.Context, id uint) (bool, error) {
	rows, err := s.repo.UpdateOEMHoldStatusIfMatches(ctx, id, "active", "released")
	if err != nil {
		return false, err
	}
	return rows > 0, nil
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

// ComputeWeightedAvgCost 返回产品加权平均批次成本
func (s *ProductService) ComputeWeightedAvgCost(ctx context.Context, productID string) float64 {
	if s == nil || s.repo == nil {
		return 0
	}
	return s.repo.ComputeWeightedAvgCost(ctx, productID)
}

// GetProductBasePrice 返回产品 BasePrice（目录/合同标价字段）。
func (s *ProductService) GetProductBasePrice(ctx context.Context, productID string) float64 {
	if s == nil || s.repo == nil {
		return 0
	}
	p, err := s.repo.FindByID(ctx, productID)
	if err != nil || p == nil {
		return 0
	}
	return p.BasePrice
}

// GetCOGSReferencePrice 返回 COGS 比例折算用的目录参考价（不受合同价污染的 base_price 影响）。
func (s *ProductService) GetCOGSReferencePrice(ctx context.Context, productID string) float64 {
	if s == nil || s.repo == nil {
		return 0
	}
	p, err := s.repo.FindByID(ctx, productID)
	if err != nil || p == nil {
		return 0
	}
	weighted := s.repo.ComputeWeightedAvgCost(ctx, p.ID)
	return catalog.COGSReferencePrice(p.Slug, p.BasePrice, weighted)
}
