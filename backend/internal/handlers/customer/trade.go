package customer

import (
	modelsTrade "candypro/api/internal/models/trade"
	"candypro/api/internal/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CustomerListTradeTransactions lists current user's trades with pagination.
func (h *Handler) CustomerListTradeTransactions(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
		return
	}

	page, limit := utils.ParsePagination(c, 10, 50)
	transactions, totalItems, err := h.services.Trade.ListTransactions(c.Request.Context(), userID, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch transactions")
		return
	}

	items := make([]gin.H, 0, len(transactions))
	for _, transaction := range transactions {
		items = append(items, toTradeResponse(transaction))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       items,
		"pagination": utils.BuildPagination(totalItems, page, limit),
	})
}

// CustomerGetTradeTransaction gets a single trade owned by current user.
func (h *Handler) CustomerGetTradeTransaction(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidRequestResponse(c, "invalid ID format")
		return
	}

	trans, err := h.services.Trade.GetTransaction(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Transaction not found")
		return
	}

	if trans.UserID != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "forbidden", "You do not have access to this transaction")
		return
	}

	c.JSON(http.StatusOK, toTradeResponse(*trans))
}

// CustomerCreateTradeTransaction creates a trade owned by current user.
func (h *Handler) CustomerCreateTradeTransaction(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
		return
	}

	var req struct {
		InquiryID   *string `json:"inquiryId"`
		Reference   string  `json:"reference"`
		Status      string  `json:"status"`
		Currency    string  `json:"currency"`
		TotalAmount float64 `json:"totalAmount"`
		Terms       string  `json:"terms"`
		Incoterms   string  `json:"incoterms"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	terms := strings.TrimSpace(req.Terms)
	if terms == "" {
		terms = strings.TrimSpace(req.Incoterms)
	}

	transaction := &modelsTrade.TradeTransaction{
		InquiryID:   req.InquiryID,
		UserID:      userID,
		Reference:   strings.TrimSpace(req.Reference),
		Status:      strings.TrimSpace(req.Status),
		Currency:    strings.TrimSpace(req.Currency),
		TotalAmount: req.TotalAmount,
		Terms:       terms,
	}

	if err := h.services.Trade.CreateTransaction(c.Request.Context(), transaction); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to create transaction")
		return
	}

	c.JSON(http.StatusCreated, toTradeResponse(*transaction))
}

func toTradeResponse(transaction modelsTrade.TradeTransaction) gin.H {
	docs := make([]gin.H, 0, len(transaction.Documents))
	for _, doc := range transaction.Documents {
		docs = append(docs, gin.H{
			"id":            doc.ID,
			"transactionId": doc.TransactionID,
			"type":          doc.Type,
			"docType":       doc.Type,
			"docNumber":     doc.DocNumber,
			"status":        doc.Status,
			"isAIGenerated": doc.IsAIGenerated,
			"aiVersion":     doc.AIVersion,
			"createdAt":     doc.CreatedAt,
			"updatedAt":     doc.UpdatedAt,
		})
	}

	return gin.H{
		"id":          transaction.ID,
		"inquiryId":   transaction.InquiryID,
		"userId":      transaction.UserID,
		"reference":   transaction.Reference,
		"status":      transaction.Status,
		"currency":    transaction.Currency,
		"totalAmount": transaction.TotalAmount,
		"terms":       transaction.Terms,
		"incoterms":   transaction.Terms,
		"documents":   docs,
		"createdAt":   transaction.CreatedAt,
		"updatedAt":   transaction.UpdatedAt,
	}
}
