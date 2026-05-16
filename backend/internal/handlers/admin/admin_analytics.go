package admin

import (
	"candypro/api/internal/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ── Sales Velocity ──

// GetSalesVelocity returns sales velocity by product over N months.
func (h *Handler) GetSalesVelocity(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	months, err := strconv.Atoi(c.DefaultQuery("months", "6"))
	if err != nil || months < 1 {
		months = 6
	}
	result, err := h.services.Order.SalesVelocity(c.Request.Context(), months)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "sales_velocity_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ── RFM Customer Analysis ──

// GetRFMAnalysis returns Recency/Frequency/Monetary segmentation for all customers.
func (h *Handler) GetRFMAnalysis(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	result, err := h.services.Order.RFMAnalysis(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "rfm_analysis_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ── Customer Churn ──

// GetCustomerChurn returns customers at risk of churning.
func (h *Handler) GetCustomerChurn(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	dormantDays, err := strconv.Atoi(c.DefaultQuery("dormantDays", "90"))
	if err != nil || dormantDays < 1 {
		dormantDays = 90
	}
	result, err := h.services.Order.CustomerChurn(c.Request.Context(), dormantDays)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "churn_analysis_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ── Inventory Health ──

// GetInventoryHealth returns inventory health metrics for all products.
func (h *Handler) GetInventoryHealth(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	windowDays, err := strconv.Atoi(c.DefaultQuery("windowDays", "30"))
	if err != nil || windowDays < 1 {
		windowDays = 30
	}
	result, err := h.services.Order.InventoryHealth(c.Request.Context(), windowDays)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inventory_health_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ── P&L Report ──

// GetProfitLoss returns profit-loss report grouped by period.
func (h *Handler) GetProfitLoss(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	groupBy := c.DefaultQuery("groupBy", "month")
	periods, err := strconv.Atoi(c.DefaultQuery("periods", "12"))
	if err != nil || periods < 1 {
		periods = 12
	}
	result, err := h.services.Order.ProfitLossByPeriod(c.Request.Context(), groupBy, periods)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "profit_loss_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ── Smart Replenishment ──

// GetReplenishmentSuggestions returns smart replenishment recommendations.
func (h *Handler) GetReplenishmentSuggestions(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	cycleDays, err := strconv.Atoi(c.DefaultQuery("cycleDays", "30"))
	if err != nil || cycleDays < 1 {
		cycleDays = 30
	}
	windowDays, err := strconv.Atoi(c.DefaultQuery("windowDays", "30"))
	if err != nil || windowDays < 1 {
		windowDays = 30
	}
	result, err := h.services.Order.ReplenishmentSuggestions(c.Request.Context(), cycleDays, windowDays)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "replenishment_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
