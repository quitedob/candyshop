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

	if user.CompanyID == nil {
		response.ErrorResp(c, http.StatusNotFound, "no_company_profile")
		return
	}

	company, err := h.services.Company.GetCompany(c.Request.Context(), *user.CompanyID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "company_not_found")
		return
	}

	if company.PriceListID == nil {
		c.JSON(http.StatusOK, gin.H{"priceList": nil, "message": "No price list assigned to your company"})
		return
	}

	priceList, err := h.services.Price.GetPriceList(c.Request.Context(), *company.PriceListID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "price_list_not_found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"priceList": priceList})
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

	if user.CompanyID == nil {
		c.JSON(http.StatusOK, gin.H{"unitPrice": nil, "message": "No company profile found"})
		return
	}

	company, err := h.services.Company.GetCompany(c.Request.Context(), *user.CompanyID)
	if err != nil || company.PriceListID == nil {
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
