package admin

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminGetTradeTransactions lists all trade transactions with optional status filter.
func (h *Handler) AdminGetTradeTransactions(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}

	page, limit := utils.ParsePagination(c, 20, 100)
	status := strings.TrimSpace(c.Query("status"))

	transactions, total, err := h.services.Trade.ListAllTransactions(c.Request.Context(), page, limit, status)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "trade_fetch_failed")
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, modelsProduct.PaginatedResponse{
		Data: transactions,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// AdminGetTradeTransaction returns a single trade transaction.
func (h *Handler) AdminGetTradeTransaction(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}

	transaction, err := h.services.Trade.GetTransaction(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "trade_not_found")
		return
	}

	c.JSON(http.StatusOK, transaction)
}

type adminUpdateTradeStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// AdminUpdateTradeStatus updates a trade transaction's status.
func (h *Handler) AdminUpdateTradeStatus(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}

	var req adminUpdateTradeStatusRequest
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}

	status := strings.TrimSpace(req.Status)
	if err := h.services.Trade.UpdateTransactionStatus(c.Request.Context(), uint(id), status); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "trade_status_update_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Trade status updated",
		"id":      id,
		"status":  status,
	})
}

// AdminGetTradeDocuments returns all documents for a trade transaction.
func (h *Handler) AdminGetTradeDocuments(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.InvalidResp(c, "invalid_transaction_id")
		return
	}

	docs, err := h.services.Trade.ListDocuments(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "trade_doc_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, docs)
}

type adminUpdateTradeDocumentRequest struct {
	Status string `json:"status"`
}

// AdminUpdateTradeDocument updates a trade document.
func (h *Handler) AdminUpdateTradeDocument(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}

	docIDStr := c.Param("docId")
	docID, err := strconv.ParseUint(docIDStr, 10, 64)
	if err != nil {
		utils.InvalidResp(c, "invalid_document_id")
		return
	}

	doc, err := h.services.Trade.GetDocument(c.Request.Context(), uint(docID))
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}

	var req adminUpdateTradeDocumentRequest
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}

	if req.Status != "" {
		doc.Status = strings.TrimSpace(req.Status)
	}

	if err := h.services.Trade.UpdateDocument(c.Request.Context(), doc); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "trade_doc_update_failed")
		return
	}

	c.JSON(http.StatusOK, doc)
}
