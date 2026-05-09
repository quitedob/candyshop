package search

import modelsProduct "candypro/api/internal/models/product"

// mergeProductResults 合并向量检索与 ILIKE 结果：语义命中优先，其次关键词，去重
func mergeProductResults(semantic, keyword []modelsProduct.Product, limit int) []modelsProduct.Product {
	if limit <= 0 {
		limit = 10
	}
	seen := make(map[string]struct{}, limit*2)
	out := make([]modelsProduct.Product, 0, limit)
	for _, p := range semantic {
		if len(out) >= limit || p.ID == "" {
			continue
		}
		if _, ok := seen[p.ID]; ok {
			continue
		}
		seen[p.ID] = struct{}{}
		out = append(out, p)
	}
	for _, p := range keyword {
		if len(out) >= limit || p.ID == "" {
			continue
		}
		if _, ok := seen[p.ID]; ok {
			continue
		}
		seen[p.ID] = struct{}{}
		out = append(out, p)
	}
	return out
}
