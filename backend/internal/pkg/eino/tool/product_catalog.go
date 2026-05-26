// Package tool — 产品目录只读 Tool，供 DeepAgent 子 agent 查询
package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// ProductCatalogSearcher 产品目录查询接口（由 system handler 注入）
type ProductCatalogSearcher interface {
	SearchProducts(ctx context.Context, query string, limit int) ([]map[string]any, error)
	GetProductDetail(ctx context.Context, productID string) (map[string]any, error)
}

type searchProductsRequest struct {
	Query string `json:"query" jsonschema_description:"Product name, category, or keyword"`
	Limit int    `json:"limit" jsonschema_description:"Max results, default 5"`
}

type searchProductsResponse struct {
	Results string `json:"results" jsonschema_description:"JSON array of matching products"`
}

type getProductDetailRequest struct {
	ProductID string `json:"product_id" jsonschema_description:"Product ID or slug"`
}

type getProductDetailResponse struct {
	Detail string `json:"detail" jsonschema_description:"JSON object with product fields"`
}

// NewSearchProductsTool 创建产品搜索 Tool
func NewSearchProductsTool(ctx context.Context, searcher ProductCatalogSearcher) (tool.BaseTool, error) {
	return utils.InferTool("search_products", "Search the CandyPro product catalog by keyword.",
		func(ctx context.Context, req *searchProductsRequest) (*searchProductsResponse, error) {
			if searcher == nil {
				return nil, fmt.Errorf("product catalog not configured")
			}
			if req == nil {
				return nil, fmt.Errorf("request is required")
			}
			limit := req.Limit
			if limit <= 0 || limit > 20 {
				limit = 5
			}
			results, err := searcher.SearchProducts(ctx, req.Query, limit)
			if err != nil {
				return nil, err
			}
			b, err := json.Marshal(results)
			if err != nil {
				return nil, err
			}
			return &searchProductsResponse{Results: string(b)}, nil
		})
}

// NewGetProductDetailTool 创建产品详情 Tool
func NewGetProductDetailTool(ctx context.Context, searcher ProductCatalogSearcher) (tool.BaseTool, error) {
	return utils.InferTool("get_product_detail", "Get detailed product information by product ID.",
		func(ctx context.Context, req *getProductDetailRequest) (*getProductDetailResponse, error) {
			if searcher == nil {
				return nil, fmt.Errorf("product catalog not configured")
			}
			if req == nil || req.ProductID == "" {
				return nil, fmt.Errorf("product_id is required")
			}
			detail, err := searcher.GetProductDetail(ctx, req.ProductID)
			if err != nil {
				return nil, err
			}
			b, err := json.Marshal(detail)
			if err != nil {
				return nil, err
			}
			return &getProductDetailResponse{Detail: string(b)}, nil
		})
}
