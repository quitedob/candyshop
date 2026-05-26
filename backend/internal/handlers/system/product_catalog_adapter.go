package system

import (
	"context"

	modelsProduct "candypro/api/internal/models/product"
	einotool "candypro/api/internal/pkg/eino/tool"
)

// productCatalogAdapter 将 ProductService 适配为 DeepAgent 只读目录 Tool
type productCatalogAdapter struct {
	search func(ctx context.Context, query string, limit int) ([]modelsProduct.Product, error)
	get    func(ctx context.Context, id string) (*modelsProduct.Product, error)
}

func (a *productCatalogAdapter) SearchProducts(ctx context.Context, query string, limit int) ([]map[string]any, error) {
	products, err := a.search(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(products))
	for _, p := range products {
		out = append(out, map[string]any{
			"id": p.ID, "name": p.Name, "category": p.Category,
			"moq": p.MOQ, "basePrice": p.BasePrice, "status": p.Status,
		})
	}
	return out, nil
}

func (a *productCatalogAdapter) GetProductDetail(ctx context.Context, productID string) (map[string]any, error) {
	p, err := a.get(ctx, productID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id": p.ID, "name": p.Name, "summary": p.Summary, "description": p.Description,
		"category": p.Category, "moq": p.MOQ, "basePrice": p.BasePrice,
		"leadTime": p.LeadTime, "certifications": p.Certifications, "status": p.Status,
	}, nil
}

func newProductCatalogAdapter(h *Handler) einotool.ProductCatalogSearcher {
	if h.services == nil || h.services.Product == nil {
		return nil
	}
	return &productCatalogAdapter{
		search: func(ctx context.Context, query string, limit int) ([]modelsProduct.Product, error) {
			res, err := h.services.Product.GetProductsFiltered(ctx, 1, limit, false, false, false, query, "", 0, 0)
			if err != nil || res == nil {
				return nil, err
			}
			if products, ok := res.Data.([]modelsProduct.Product); ok {
				return products, nil
			}
			return nil, nil
		},
		get: h.services.Product.GetProduct,
	}
}
