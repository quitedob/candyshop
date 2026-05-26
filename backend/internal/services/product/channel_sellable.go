package product

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"strings"
)

// EffectiveSellableByProducts 可售量：聚合库存 − OEM 预留，并按渠道 ListedQuantity 封顶（渠道行为空则跳过渠道表）。
//
// E-1 性能修复：原先按行调用 FindByID + SumActiveOEMHoldsForProduct + FindChannelInventory，
// N 行产品产生 3N 个数据库往返。改为三次批量查询：FindByIDs / SumActiveOEMHoldsByProductIDs /
// FindChannelInventoriesByProductIDs，常数往返与产品数无关。
func (s *ProductService) EffectiveSellableByProducts(ctx context.Context, productIDs []string, channelCode string) (map[string]int, error) {
	out := make(map[string]int, len(productIDs))
	if len(productIDs) == 0 {
		return out, nil
	}

	// 去重 + trim：上游可能传入重复或空白 ID
	seen := make(map[string]struct{}, len(productIDs))
	cleaned := make([]string, 0, len(productIDs))
	for _, pid := range productIDs {
		pid = strings.TrimSpace(pid)
		if pid == "" {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		cleaned = append(cleaned, pid)
	}
	if len(cleaned) == 0 {
		return out, nil
	}

	// 1) 一次性加载产品（含 stock_quantity）
	products, err := s.repo.FindByIDs(ctx, cleaned)
	if err != nil {
		return nil, err
	}
	productByID := make(map[string]*modelsProduct.Product, len(products))
	for i := range products {
		productByID[products[i].ID] = &products[i]
	}

	// 2) 一次性聚合 OEM 预留
	holdsByID, err := s.repo.SumActiveOEMHoldsByProductIDs(ctx, cleaned)
	if err != nil {
		// 批量失败时不中止，按零预留处理（保持原行为容错性）
		holdsByID = map[string]int64{}
	}

	// 3) 渠道库存（仅当 channelCode 不为空）
	ch := strings.TrimSpace(channelCode)
	var channelByID map[string]*modelsProduct.ChannelInventory
	if ch != "" {
		channelByID, err = s.repo.FindChannelInventoriesByProductIDs(ctx, cleaned, ch)
		if err != nil {
			channelByID = map[string]*modelsProduct.ChannelInventory{}
		}
	}

	// 4) 按产品计算最终可售量
	for _, pid := range cleaned {
		p := productByID[pid]
		if p == nil {
			continue
		}
		eff := p.StockQuantity - int(holdsByID[pid])
		if eff < 0 {
			eff = 0
		}
		if row, ok := channelByID[pid]; ok && row != nil && row.ListedQuantity > 0 {
			cap := row.ListedQuantity - row.ReservedForChannel
			if cap < 0 {
				cap = 0
			}
			if eff > cap {
				eff = cap
			}
		}
		out[pid] = eff
	}
	return out, nil
}

// ListChannelInventoriesForProduct OMS 渠道行列表
func (s *ProductService) ListChannelInventoriesForProduct(ctx context.Context, productID string) ([]modelsProduct.ChannelInventory, error) {
	return s.repo.ListChannelInventoriesForProduct(ctx, productID)
}

// UpsertChannelInventory 写入渠道库存与同步元数据
func (s *ProductService) UpsertChannelInventory(ctx context.Context, row *modelsProduct.ChannelInventory) error {
	return s.repo.UpsertChannelInventory(ctx, row)
}
