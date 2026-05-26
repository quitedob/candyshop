package order

import (
	"context"

	modelsOrder "candypro/api/internal/models/order"
)

// productCostLookup 查询产品成本与标价，用于 COGS 比例折算
type productCostLookup interface {
	ComputeWeightedAvgCost(ctx context.Context, productID string) float64
	GetCOGSReferencePrice(ctx context.Context, productID string) float64
}

// effectiveUnitCost 按实际售价比例折算单位成本，使 COGS 与合同价/BasePrice 尺度一致。
// basePrice 缺失时返回 0，避免未折算的加权成本虚高（COGS > 营收）。
//
// A-1: 加 sanity cap（不超过 unitPrice * 0.95），防止 batch UnitCost 数据录入错误
// 或币种不匹配时 weighted/base 比率爆炸——曾出现订单 COGS 是营收 16-20 倍的现象。
// 触发 cap 后回退到 unitPrice * 0.4 作为保守估值，保留 P&L 近似值并避免 NaN 放大。
const (
	cogsRatioCeiling  = 0.95 // unitCost <= unitPrice * cogsRatioCeiling
	cogsFallbackRatio = 0.40 // 触发 cap 时退化为 40% 毛成本估值
)

func effectiveUnitCost(unitPrice, weightedAvgCost, basePrice float64) float64 {
	if weightedAvgCost <= 0 {
		return 0
	}
	if unitPrice <= 0 || basePrice <= 0 {
		return 0
	}
	cost := unitPrice * (weightedAvgCost / basePrice)
	// 异常上限：成本不应大于售价的 cogsRatioCeiling，否则视为数据异常。
	if cost > unitPrice*cogsRatioCeiling {
		return unitPrice * cogsFallbackRatio
	}
	if cost < 0 {
		return 0
	}
	return cost
}

// ComputeOrderCOGS 汇总订单行 quantity × 有效单位成本
//
// 同一 productID 在订单中可能多次出现（不同规格/批次的拆分行），
// 这里按 productID 缓存 weighted 与 reference 查询结果，避免重复调用
// 数据库（H-1）。
func ComputeOrderCOGS(ctx context.Context, items []modelsOrder.OrderItem, lookup productCostLookup) float64 {
	if lookup == nil {
		return 0
	}
	type costPair struct {
		weighted float64
		base     float64
	}
	cache := make(map[string]costPair, len(items))
	var total float64
	for _, item := range items {
		if item.Quantity < 1 {
			continue
		}
		entry, ok := cache[item.ProductID]
		if !ok {
			entry = costPair{
				weighted: lookup.ComputeWeightedAvgCost(ctx, item.ProductID),
				base:     lookup.GetCOGSReferencePrice(ctx, item.ProductID),
			}
			cache[item.ProductID] = entry
		}
		unitCost := effectiveUnitCost(item.UnitPrice, entry.weighted, entry.base)
		total += float64(item.Quantity) * unitCost
	}
	return total
}
