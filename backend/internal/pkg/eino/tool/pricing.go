// Package tool — 定价只读 Tool，供 DeepAgent PricingExpert 查询
package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// PricingQuoter 定价查询接口（由 system handler 注入）
type PricingQuoter interface {
	GetPriceTiers(ctx context.Context, productID string) ([]map[string]any, error)
	ComputeQuoteDraft(ctx context.Context, productID string, quantity int, currency string) (map[string]any, error)
}

type getPriceTiersRequest struct {
	ProductID string `json:"product_id" jsonschema_description:"Product ID"`
}

type getPriceTiersResponse struct {
	Tiers string `json:"tiers" jsonschema_description:"JSON array of price tiers"`
}

type computeQuoteDraftRequest struct {
	ProductID string `json:"product_id" jsonschema_description:"Product ID"`
	Quantity  int    `json:"quantity" jsonschema_description:"Order quantity"`
	Currency  string `json:"currency" jsonschema_description:"ISO currency code, default USD"`
}

type computeQuoteDraftResponse struct {
	Quote string `json:"quote" jsonschema_description:"JSON quote draft with unit price and totals"`
}

// NewGetPriceTiersTool 创建价格阶梯查询 Tool
func NewGetPriceTiersTool(ctx context.Context, quoter PricingQuoter) (tool.BaseTool, error) {
	return utils.InferTool("get_price_tiers", "Get quantity-based price tiers for a product.",
		func(ctx context.Context, req *getPriceTiersRequest) (*getPriceTiersResponse, error) {
			if quoter == nil {
				return nil, fmt.Errorf("pricing service not configured")
			}
			if req == nil || req.ProductID == "" {
				return nil, fmt.Errorf("product_id is required")
			}
			tiers, err := quoter.GetPriceTiers(ctx, req.ProductID)
			if err != nil {
				return nil, err
			}
			b, err := json.Marshal(tiers)
			if err != nil {
				return nil, err
			}
			return &getPriceTiersResponse{Tiers: string(b)}, nil
		})
}

// NewComputeQuoteDraftTool 创建报价草稿 Tool
func NewComputeQuoteDraftTool(ctx context.Context, quoter PricingQuoter) (tool.BaseTool, error) {
	return utils.InferTool("compute_quote_draft", "Compute a quotation draft for a product and quantity.",
		func(ctx context.Context, req *computeQuoteDraftRequest) (*computeQuoteDraftResponse, error) {
			if quoter == nil {
				return nil, fmt.Errorf("pricing service not configured")
			}
			if req == nil || req.ProductID == "" || req.Quantity <= 0 {
				return nil, fmt.Errorf("product_id and positive quantity are required")
			}
			currency := req.Currency
			if currency == "" {
				currency = "USD"
			}
			quote, err := quoter.ComputeQuoteDraft(ctx, req.ProductID, req.Quantity, currency)
			if err != nil {
				return nil, err
			}
			b, err := json.Marshal(quote)
			if err != nil {
				return nil, err
			}
			return &computeQuoteDraftResponse{Quote: string(b)}, nil
		})
}
