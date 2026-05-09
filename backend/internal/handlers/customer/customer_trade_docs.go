package customer

import (
	"candypro/api/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) verifyTradeOwnership(c *gin.Context, tradeID uint, userID string) bool {
	trans, err := h.services.Trade.GetTransaction(c.Request.Context(), tradeID)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "trade_not_found")
		return false
	}
	if trans.UserID != userID {
		utils.ErrorResp(c, http.StatusForbidden, "forbidden")
		return false
	}
	return true
}

func (h *Handler) CustomerGetTradeDocuments(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if !h.verifyTradeOwnership(c, uint(id), userID) {
		return
	}
	docs, err := h.services.Trade.ListDocuments(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": docs})
}

func (h *Handler) CustomerGetTradeSalesContract(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if !h.verifyTradeOwnership(c, uint(id), userID) {
		return
	}
	sc, err := h.services.TradeDocDetail.GetSalesContract(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}
	c.JSON(http.StatusOK, sc)
}

func (h *Handler) CustomerGetTradePackingList(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if !h.verifyTradeOwnership(c, uint(id), userID) {
		return
	}
	pl, err := h.services.TradeDocDetail.GetPackingList(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}
	c.JSON(http.StatusOK, pl)
}

func (h *Handler) CustomerGetTradeCertificateOfOrigin(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if !h.verifyTradeOwnership(c, uint(id), userID) {
		return
	}
	coo, err := h.services.TradeDocDetail.GetCertificateOfOrigin(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}
	c.JSON(http.StatusOK, coo)
}

func (h *Handler) CustomerGetTradeHealthCertificate(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if !h.verifyTradeOwnership(c, uint(id), userID) {
		return
	}
	hc, err := h.services.TradeDocDetail.GetHealthCertificate(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}
	c.JSON(http.StatusOK, hc)
}

func (h *Handler) CustomerGetTradeProformaInvoice(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if !h.verifyTradeOwnership(c, uint(id), userID) {
		return
	}
	pi, err := h.services.TradeDocDetail.GetProformaInvoice(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}
	c.JSON(http.StatusOK, pi)
}

func (h *Handler) CustomerGetTradeCommercialInvoice(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if !h.verifyTradeOwnership(c, uint(id), userID) {
		return
	}
	ci, err := h.services.TradeDocDetail.GetCommercialInvoice(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}
	c.JSON(http.StatusOK, ci)
}

func (h *Handler) CustomerGetTradeBillOfLading(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if !h.verifyTradeOwnership(c, uint(id), userID) {
		return
	}
	bl, err := h.services.TradeDocDetail.GetBillOfLading(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}
	c.JSON(http.StatusOK, bl)
}
