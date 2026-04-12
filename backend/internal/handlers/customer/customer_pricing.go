package customer

import (
	"candypro/api/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CustomerGetMyPriceList returns the price list assigned to the current user's company.
func (h *Handler) CustomerGetMyPriceList(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "User not found")
		return
	}

	if user.CompanyID == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "No company profile found")
		return
	}

	company, err := h.services.Company.GetCompany(c.Request.Context(), *user.CompanyID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Company not found")
		return
	}

	if company.PriceListID == nil {
		c.JSON(http.StatusOK, gin.H{"priceList": nil, "message": "No price list assigned to your company"})
		return
	}

	priceList, err := h.services.Price.GetPriceList(c.Request.Context(), *company.PriceListID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Price list not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"priceList": priceList})
}

// CustomerGetProductPrice returns the applicable price for a product given the user's price list.
func (h *Handler) CustomerGetProductPrice(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
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
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "User not found")
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
