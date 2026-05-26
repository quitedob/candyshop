package customer

import (
	"candypro/api/internal/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CustomerGetMyPriceList returns the price list assigned to the current user's company.
func (h *Handler) CustomerGetMyPriceList(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	company, err := h.services.Company.EnsureCompanyForUser(c.Request.Context(), user)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "company_provision_failed")
		return
	}
	if company == nil {
		c.JSON(http.StatusOK, gin.H{"priceList": nil, "status": "no_company_profile"})
		return
	}

	if company.PriceListID == nil {
		c.JSON(http.StatusOK, gin.H{"priceList": nil, "status": "no_price_list"})
		return
	}

	priceList, err := h.services.Price.GetPriceList(c.Request.Context(), *company.PriceListID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"priceList": nil, "status": "price_list_not_found"})
		return
	}

	rules, err := h.services.Price.GetPriceListRules(c.Request.Context(), *company.PriceListID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "price_rules_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"priceList": gin.H{
			"id":          priceList.ID,
			"name":        priceList.Name,
			"description": priceList.Description,
			"currency":    priceList.Currency,
			"status":      priceList.Status,
			"createdAt":   priceList.CreatedAt,
			"updatedAt":   priceList.UpdatedAt,
			"prices":      rules,
		},
	})
}

// CustomerGetProductPrice returns the applicable price for a product given the user's price list.
func (h *Handler) CustomerGetProductPrice(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	productID := c.Param("id")
	quantityStr := c.DefaultQuery("quantity", "1")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil || quantity < 1 {
		quantity = 1
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	company, err := h.services.Company.EnsureCompanyForUser(c.Request.Context(), user)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "company_provision_failed")
		return
	}
	if company == nil || company.PriceListID == nil {
		c.JSON(http.StatusOK, gin.H{"unitPrice": nil, "message": "No price list assigned"})
		return
	}

	unitPrice, err := h.services.Price.GetPriceForProduct(c.Request.Context(), productID, *company.PriceListID, quantity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"unitPrice": nil, "message": "No specific price found for this product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"productId":   productID,
		"priceListId": *company.PriceListID,
		"quantity":    quantity,
		"unitPrice":   unitPrice,
	})
}
