package product

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FindMarketProfilesForProducts 按产品 ID 列表与市场代码加载合规画像
func (r *ProductRepository) FindMarketProfilesForProducts(ctx context.Context, productIDs []string, marketCode string) ([]modelsProduct.ProductMarketProfile, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	mc := strings.TrimSpace(marketCode)
	var rows []modelsProduct.ProductMarketProfile
	q := r.db.WithContext(ctx).Where("product_id IN ?", productIDs)
	if mc != "" {
		q = q.Where("UPPER(market_code) = UPPER(?)", mc)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// UpsertProductMarketProfile 写入或更新市场画像
func (r *ProductRepository) UpsertProductMarketProfile(ctx context.Context, row *modelsProduct.ProductMarketProfile) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "product_id"}, {Name: "market_code"}},
		DoUpdates: clause.AssignmentColumns([]string{"destination_countries", "blocked_ingredient_patterns", "required_cert_keywords", "label_template_id", "notes", "rule_version", "effective_from", "rule_source_summary", "updated_at"}),
	}).Create(row).Error
}

// FindMarketCostStacksForProduct 读取某产品的全部目的国成本栈
func (r *ProductRepository) FindMarketCostStacksForProduct(ctx context.Context, productID string) ([]modelsProduct.ProductMarketCostStack, error) {
	var rows []modelsProduct.ProductMarketCostStack
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Find(&rows).Error
	return rows, err
}

// UpsertProductMarketCostStack 写入或更新成本栈
func (r *ProductRepository) UpsertProductMarketCostStack(ctx context.Context, row *modelsProduct.ProductMarketCostStack) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "product_id"}, {Name: "market_code"}},
		DoUpdates: clause.AssignmentColumns([]string{"logistics_per_unit", "duty_rate", "label_cost_per_unit", "compliance_per_unit", "target_gross_margin", "currency", "notes", "updated_at"}),
	}).Create(row).Error
}

// ListWarehouses 返回仓库列表
func (r *ProductRepository) ListWarehouses(ctx context.Context) ([]modelsProduct.Warehouse, error) {
	var rows []modelsProduct.Warehouse
	err := r.db.WithContext(ctx).Order("is_default DESC, code ASC").Find(&rows).Error
	return rows, err
}

// GetDefaultWarehouseID 返回默认活跃仓库 ID（is_default 优先，否则 MAIN）
func (r *ProductRepository) GetDefaultWarehouseID(ctx context.Context) (string, error) {
	var w modelsProduct.Warehouse
	if err := r.db.WithContext(ctx).Where("is_default = ? AND is_active = ?", true, true).First(&w).Error; err == nil {
		return w.ID, nil
	}
	if err := r.db.WithContext(ctx).Where("code = ? AND is_active = ?", "MAIN", true).First(&w).Error; err == nil {
		return w.ID, nil
	}
	return "", gorm.ErrRecordNotFound
}

// SaveWarehouse 创建或全量更新仓库
func (r *ProductRepository) SaveWarehouse(ctx context.Context, w *modelsProduct.Warehouse) error {
	return r.db.WithContext(ctx).Save(w).Error
}

// UpsertWarehouseStock 按 warehouse_id+product_id 写入库存行
func (r *ProductRepository) UpsertWarehouseStock(ctx context.Context, row *modelsProduct.WarehouseStock) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "warehouse_id"}, {Name: "product_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"quantity", "reserved", "variant_id", "updated_at"}),
	}).Create(row).Error
}

// SaveOEMProjectInventoryHold 创建 OEM 库存预留
func (r *ProductRepository) SaveOEMProjectInventoryHold(ctx context.Context, row *modelsProduct.OEMProjectInventoryHold) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// ListOEMInventoryHoldsByProject 列出项目下预留
func (r *ProductRepository) ListOEMInventoryHoldsByProject(ctx context.Context, projectID string) ([]modelsProduct.OEMProjectInventoryHold, error) {
	var rows []modelsProduct.OEMProjectInventoryHold
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("id ASC").Find(&rows).Error
	return rows, err
}

// SumActiveOEMHoldsForProduct 活跃 OEM 预留数量合计
func (r *ProductRepository) SumActiveOEMHoldsForProduct(ctx context.Context, productID string) (int64, error) {
	var sum int64
	err := r.db.WithContext(ctx).Model(&modelsProduct.OEMProjectInventoryHold{}).
		Where("product_id = ? AND status = ?", productID, "active").
		Select("COALESCE(SUM(quantity),0)").
		Scan(&sum).Error
	return sum, err
}

// ListChannelInventoriesForProduct 某 SKU 全渠道库存行
func (r *ProductRepository) ListChannelInventoriesForProduct(ctx context.Context, productID string) ([]modelsProduct.ChannelInventory, error) {
	var rows []modelsProduct.ChannelInventory
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("channel_code ASC").Find(&rows).Error
	return rows, err
}

// FindChannelInventory 按产品与渠道取单行；无记录时返回 gorm.ErrRecordNotFound（不触发 GORM error 日志）
func (r *ProductRepository) FindChannelInventory(ctx context.Context, productID, channelCode string) (*modelsProduct.ChannelInventory, error) {
	var row modelsProduct.ChannelInventory
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND UPPER(channel_code) = UPPER(?)", productID, channelCode).
		Limit(1).
		Find(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

// UpsertChannelInventory 写入或更新渠道库存
func (r *ProductRepository) UpsertChannelInventory(ctx context.Context, row *modelsProduct.ChannelInventory) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "product_id"}, {Name: "channel_code"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"warehouse_id", "listed_quantity", "reserved_for_channel", "external_sku",
			"sync_status", "last_synced_at", "sync_notes", "updated_at",
		}),
	}).Create(row).Error
}

// SumActiveOEMHoldsByProductIDs 批量返回多个产品当前 active OEM 预留量。
// 返回 map[productID]quantity；不在结果中视为 0。
//
// E-1: 取代每行调用 SumActiveOEMHoldsForProduct 的 N+1 模式。
func (r *ProductRepository) SumActiveOEMHoldsByProductIDs(ctx context.Context, productIDs []string) (map[string]int64, error) {
	if len(productIDs) == 0 {
		return map[string]int64{}, nil
	}
	type row struct {
		ProductID string
		Total     int64
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&modelsProduct.OEMProjectInventoryHold{}).
		Select("product_id AS product_id, COALESCE(SUM(quantity),0) AS total").
		Where("product_id IN ? AND status = ?", productIDs, "active").
		Group("product_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, r := range rows {
		out[r.ProductID] = r.Total
	}
	return out, nil
}

// FindChannelInventoriesByProductIDs 批量查询渠道库存行，按 (productID, channelCode) 索引。
// 仅返回查询的渠道；productIDs 为空或 channelCode 为空时返回空 map。
//
// E-1: 取代逐行 FindChannelInventory 的 N+1 模式。
func (r *ProductRepository) FindChannelInventoriesByProductIDs(ctx context.Context, productIDs []string, channelCode string) (map[string]*modelsProduct.ChannelInventory, error) {
	out := make(map[string]*modelsProduct.ChannelInventory)
	if len(productIDs) == 0 {
		return out, nil
	}
	ch := strings.TrimSpace(channelCode)
	if ch == "" {
		return out, nil
	}
	var rows []modelsProduct.ChannelInventory
	err := r.db.WithContext(ctx).
		Where("product_id IN ? AND UPPER(channel_code) = UPPER(?)", productIDs, ch).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for i := range rows {
		out[rows[i].ProductID] = &rows[i]
	}
	return out, nil
}
