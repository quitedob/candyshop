package kyb

import (
	modelsOrder "candypro/api/internal/models/order"
	"strings"
)

// TierConfig 与 config.KYBConfig 对齐的免审档位（避免 config 包循环依赖）
type TierConfig struct {
	BypassMaxOrderUSD       float64
	BypassSampleMaxOrderUSD float64
	SampleProductIDs        []string
}

// PendingOrderAllowed pending 用户在额度内是否允许下单（已激活用户由调用方先行放行）
func PendingOrderAllowed(cfg TierConfig, orderUSD float64, lineProductIDs []string) bool {
	maxUSD := cfg.BypassMaxOrderUSD
	sampleMax := cfg.BypassSampleMaxOrderUSD
	sampleOnly := OnlySampleProducts(lineProductIDs, cfg.SampleProductIDs)
	if maxUSD <= 0 && !(sampleOnly && sampleMax > 0) {
		return false
	}
	if sampleOnly && sampleMax > 0 && orderUSD <= sampleMax {
		return true
	}
	if maxUSD > 0 && orderUSD <= maxUSD {
		return true
	}
	return false
}

// OnlySampleProducts 订单行产品是否全部属于样品 SKU 列表
func OnlySampleProducts(lineProductIDs, sampleIDs []string) bool {
	if len(sampleIDs) == 0 || len(lineProductIDs) == 0 {
		return false
	}
	set := make(map[string]bool, len(sampleIDs))
	for _, id := range sampleIDs {
		set[strings.TrimSpace(id)] = true
	}
	for _, pid := range lineProductIDs {
		if !set[strings.TrimSpace(pid)] {
			return false
		}
	}
	return true
}

// LineProductIDs 从订单行提取去重产品 ID
func LineProductIDs(items []modelsOrder.OrderItem) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(items))
	for _, it := range items {
		pid := strings.TrimSpace(it.ProductID)
		if pid == "" || seen[pid] {
			continue
		}
		seen[pid] = true
		out = append(out, pid)
	}
	return out
}

// CartLineProductIDs 购物车行与可选额外 SKU 合并去重（用于 KYB 样品档位）
func CartLineProductIDs(items []modelsOrder.CartItem, extra ...string) []string {
	seen := make(map[string]bool)
	for _, it := range items {
		if pid := strings.TrimSpace(it.ProductID); pid != "" {
			seen[pid] = true
		}
	}
	for _, e := range extra {
		if pid := strings.TrimSpace(e); pid != "" {
			seen[pid] = true
		}
	}
	out := make([]string, 0, len(seen))
	for pid := range seen {
		out = append(out, pid)
	}
	return out
}
