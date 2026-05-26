package system

import (
	"context"

	einotool "candypro/api/internal/pkg/eino/tool"
)

// pricingAdapter 将 ProductService 适配为 DeepAgent 定价 Tool
type pricingAdapter struct {
	getProduct func(ctx context.Context, productID string) (map[string]any, error)
}

func (a *pricingAdapter) GetPriceTiers(ctx context.Context, productID string) ([]map[string]any, error) {
	p, err := a.getProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	return []map[string]any{{
		"minQuantity": p["moq"], "unitPrice": p["basePrice"], "currency": "USD",
	}}, nil
}

func (a *pricingAdapter) ComputeQuoteDraft(ctx context.Context, productID string, quantity int, currency string) (map[string]any, error) {
	p, err := a.getProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	unitPrice, _ := p["basePrice"].(float64)
	subtotal := unitPrice * float64(quantity)
	if currency == "" {
		currency = "USD"
	}
	return map[string]any{
		"productId": productID, "quantity": quantity, "currency": currency,
		"unitPrice": unitPrice, "subtotal": subtotal, "requiresHumanReview": subtotal > 10000,
	}, nil
}

func newPricingAdapter(h *Handler) einotool.PricingQuoter {
	if h.services == nil || h.services.Product == nil {
		return nil
	}
	return &pricingAdapter{
		getProduct: func(ctx context.Context, productID string) (map[string]any, error) {
			product, err := h.services.Product.GetProduct(ctx, productID)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"moq": product.MOQ, "basePrice": product.BasePrice,
			}, nil
		},
	}
}
