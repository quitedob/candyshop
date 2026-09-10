package admin

import (
	"net/http"

	modelsTrade "candypro/api/internal/models/trade"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminListQuotationReviews lists the AI human-review queue, optionally filtered
// by status. GET /admin/quotation-reviews?status=&page=&limit=
func (h *Handler) AdminListQuotationReviews(c *gin.Context) {
	if h.services == nil || h.services.QuotationReview == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	rows, total, err := h.services.QuotationReview.List(c.Request.Context(), c.Query("status"), page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "quotation_reviews_list_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  rows,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// AdminApproveQuotationReview approves a pending quotation review.
// POST /admin/quotation-reviews/:id/approve
func (h *Handler) AdminApproveQuotationReview(c *gin.Context) {
	h.decideQuotationReview(c, modelsTrade.QuotationReviewStatusApproved)
}

// AdminRejectQuotationReview rejects a pending quotation review.
// POST /admin/quotation-reviews/:id/reject
func (h *Handler) AdminRejectQuotationReview(c *gin.Context) {
	h.decideQuotationReview(c, modelsTrade.QuotationReviewStatusRejected)
}

func (h *Handler) decideQuotationReview(c *gin.Context, status string) {
	if h.services == nil || h.services.QuotationReview == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		response.InvalidResp(c, "invalid_quotation_review_id")
		return
	}
	userID := c.GetString("userID")
	var req struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&req) // notes is optional; ignore bind errors
	if err := h.services.QuotationReview.Decide(c.Request.Context(), id, status, req.Notes, userID); err != nil {
		response.ErrorResp(c, http.StatusConflict, "quotation_review_decision_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Quotation review updated", "id": id, "status": status})
}
