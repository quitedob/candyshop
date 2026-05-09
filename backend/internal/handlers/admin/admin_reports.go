package admin

import (
	"candypro/api/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetRevenueTrends returns monthly revenue aggregates.
func (h *Handler) GetRevenueTrends(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	months, err := strconv.Atoi(c.DefaultQuery("months", "12"))
	if err != nil || months < 1 {
		months = 12
	}

	result, err := h.services.Order.RevenueByMonth(c.Request.Context(), months)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "revenue_trends_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetOrderTrends returns monthly order counts.
func (h *Handler) GetOrderTrends(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	months, err := strconv.Atoi(c.DefaultQuery("months", "12"))
	if err != nil || months < 1 {
		months = 12
	}

	result, err := h.services.Order.OrderCountByMonth(c.Request.Context(), months)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "order_trends_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetTopProducts returns top products by revenue.
func (h *Handler) GetTopProducts(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	result, err := h.services.Order.TopProductsByRevenue(c.Request.Context(), limit)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "top_products_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetConversionTrends returns monthly inquiry-to-order conversion rates.
func (h *Handler) GetConversionTrends(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	months, err := strconv.Atoi(c.DefaultQuery("months", "12"))
	if err != nil || months < 1 {
		months = 12
	}

	result, err := h.services.Inquiry.ConversionByMonth(c.Request.Context(), months)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "conversion_trends_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
