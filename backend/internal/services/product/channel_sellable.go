package product

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"strings"
)

// EffectiveSellableByProducts 可售量：聚合库存 − OEM 预留，并按渠道 ListedQuantity 封顶（渠道行为空则跳过渠道表）
func (s *ProductService) EffectiveSellableByProducts(ctx context.Context, productIDs []string, channelCode string) (map[string]int, error) {
	out := make(map[string]int, len(productIDs))
	ch := strings.TrimSpace(channelCode)
	for _, pid := range productIDs {
		pid = strings.TrimSpace(pid)
		if pid == "" {
			continue
		}
		p, err := s.repo.FindByID(ctx, pid)
		if err != nil || p == nil {
			continue
		}
		eff := p.StockQuantity
		holds, err := s.repo.SumActiveOEMHoldsForProduct(ctx, pid)
		if err == nil {
			eff -= int(holds)
		}
		if eff < 0 {
			eff = 0
		}
		if ch != "" {
			row, err := s.repo.FindChannelInventory(ctx, pid, ch)
			if err == nil && row != nil && row.ListedQuantity > 0 {
				cap := row.ListedQuantity - row.ReservedForChannel
				if cap < 0 {
					cap = 0
				}
				if eff > cap {
					eff = cap
				}
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
