// 报价人审工具：与 admin AdminQuoteInquiry 的 requestHumanReview 及响应头 X-Quotation-Human-Review 协同。
package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// QuotationReviewRequest 结构化报价草稿，供人工审核与后续向量索引闭环
type QuotationReviewRequest struct {
	CustomerRef          string  `json:"customer_ref" jsonschema_description:"Customer or inquiry reference id"`
	Currency             string  `json:"currency" jsonschema_description:"ISO currency, e.g. USD"`
	TotalAmount          float64 `json:"total_amount" jsonschema_description:"Quoted total amount"`
	LineItemsJSON        string  `json:"line_items_json" jsonschema_description:"JSON array of line items with sku, qty, unit_price"`
	StrategyNotes        string  `json:"strategy_notes" jsonschema_description:"Pricing rationale: history, MOQ, raw material index, discount policy"`
	SuggestedDiscountPct float64 `json:"suggested_discount_pct" jsonschema_description:"Suggested discount percent 0-100"`
}

// QuotationReviewResponse 返回给模型与前端的状态载荷
type QuotationReviewResponse struct {
	Status  string `json:"status"`
	Summary string `json:"summary"`
}

// NewQuotationHumanReviewTool 注册「报价提交人工审核」工具，配合 Graph/HITL 扩展
func NewQuotationHumanReviewTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("submit_quotation_for_human_review",
		"After computing prices for a B2B quote, call this to queue the draft for sales manager approval before sending PI to the buyer. Include strategy_notes for audit and future RAG feedback.",
		func(ctx context.Context, req *QuotationReviewRequest) (*QuotationReviewResponse, error) {
			if req == nil {
				return nil, fmt.Errorf("empty request")
			}
			_, _ = json.Marshal(req)
			summary := fmt.Sprintf("Queued quotation review: %s %s total=%.2f (discount hint %.1f%%)",
				req.CustomerRef, req.Currency, req.TotalAmount, req.SuggestedDiscountPct)
			return &QuotationReviewResponse{
				Status:  "pending_human_review",
				Summary: summary,
			}, nil
		})
}
