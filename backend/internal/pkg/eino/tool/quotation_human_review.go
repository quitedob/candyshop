// 报价人审工具：与 admin AdminQuoteInquiry 的 requestHumanReview 及响应头 X-Quotation-Human-Review 协同。
package tool

import (
	"context"
	"fmt"

	tradeModels "candypro/api/internal/models/trade"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"gorm.io/datatypes"
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

// QuotationReviewSaver persists a quotation review submission. TradeService
// satisfies this interface so the agent never needs to know the storage layer.
type QuotationReviewSaver interface {
	SaveQuotationReview(ctx context.Context, qr *tradeModels.QuotationReview) error
}

// NewQuotationHumanReviewTool 注册「报价提交人工审核」工具，配合 Graph/HITL 扩展。
// saver 为空时退化为仅返回 pending 状态（不持久化）；持久化失败返回
// review_queue_failed 而非硬失败，保证 agent 流程不中断。
func NewQuotationHumanReviewTool(ctx context.Context, saver QuotationReviewSaver) (tool.BaseTool, error) {
	return utils.InferTool("submit_quotation_for_human_review",
		"After computing prices for a B2B quote, call this to queue the draft for sales manager approval before sending PI to the buyer. Include strategy_notes for audit and future RAG feedback.",
		func(ctx context.Context, req *QuotationReviewRequest) (*QuotationReviewResponse, error) {
			if req == nil {
				return nil, fmt.Errorf("empty request")
			}
			summary := fmt.Sprintf("Queued quotation review: %s %s total=%.2f (discount hint %.1f%%)",
				req.CustomerRef, req.Currency, req.TotalAmount, req.SuggestedDiscountPct)

			qr := &tradeModels.QuotationReview{
				CustomerRef:          req.CustomerRef,
				Currency:             req.Currency,
				TotalAmount:          req.TotalAmount,
				StrategyNotes:        req.StrategyNotes,
				SuggestedDiscountPct: req.SuggestedDiscountPct,
				Status:               tradeModels.QuotationReviewStatusPending,
			}
			if raw := req.LineItemsJSON; raw != "" {
				qr.LineItems = datatypes.JSON([]byte(raw))
			}

			if saver != nil {
				if err := saver.SaveQuotationReview(ctx, qr); err != nil {
					return &QuotationReviewResponse{
						Status:  "review_queue_failed",
						Summary: fmt.Sprintf("%s (persist error: %v)", summary, err),
					}, nil
				}
			}
			return &QuotationReviewResponse{
				Status:  "pending_human_review",
				Summary: summary,
			}, nil
		})
}

