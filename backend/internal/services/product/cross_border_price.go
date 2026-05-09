package product

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"strings"

	ordersvc "candypro/api/internal/services/order"
)

// ApplyMarketCostToUnitPrice 在合同价/基础价上叠加物流、关税、标签、合规摊销与目标毛利率（管理端维护的 ProductMarketCostStack）
func ApplyMarketCostToUnitPrice(baseUnit float64, stack *modelsProduct.ProductMarketCostStack) float64 {
	if stack == nil || baseUnit <= 0 {
		return baseUnit
	}
	duty := stack.DutyRate * baseUnit
	surcharge := stack.LogisticsPerUnit + stack.LabelCostPerUnit + stack.CompliancePerUnit + duty
	subtotal := baseUnit + surcharge
	m := stack.TargetGrossMargin
	if m > 0 && m < 0.95 {
		return subtotal / (1 - m)
	}
	return subtotal
}

// ResolveCheckoutUnitPrice 结算单价：先取 baseUnit（合同价或 BasePrice），再按收货国映射 marketCode 套成本栈
func (s *ProductService) ResolveCheckoutUnitPrice(ctx context.Context, product *modelsProduct.Product, baseUnit float64, shippingCountry string) float64 {
	if s == nil || product == nil || baseUnit <= 0 {
		return baseUnit
	}
	mc := ordersvc.ComplianceProfileMarketCode(shippingCountry)
	if strings.TrimSpace(mc) == "" {
		return baseUnit
	}
	stacks, err := s.repo.FindMarketCostStacksForProduct(ctx, product.ID)
	if err != nil || len(stacks) == 0 {
		return baseUnit
	}
	var pick *modelsProduct.ProductMarketCostStack
	for i := range stacks {
		if strings.EqualFold(strings.TrimSpace(stacks[i].MarketCode), strings.TrimSpace(mc)) {
			pick = &stacks[i]
			break
		}
	}
	if pick == nil {
		pick = &stacks[0]
	}
	return ApplyMarketCostToUnitPrice(baseUnit, pick)
}
