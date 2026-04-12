package admin

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// --- Price Lists ---

// AdminGetPriceLists returns all price lists with pagination.
func (h *Handler) AdminGetPriceLists(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	page, limit := utils.ParsePagination(c, 20, 100)
	result, err := h.services.Price.GetPriceLists(c.Request.Context(), page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch price lists")
		return
	}

	c.JSON(http.StatusOK, result)
}

type adminCreatePriceListRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
}

// AdminCreatePriceList creates a new price list.
func (h *Handler) AdminCreatePriceList(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	var req adminCreatePriceListRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	currency := strings.TrimSpace(req.Currency)
	if currency == "" {
		currency = "USD"
	}

	list := &modelsProduct.PriceList{
		ID:          utils.GenerateID(),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Currency:    currency,
		Status:      strings.TrimSpace(req.Status),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.services.Price.CreatePriceList(c.Request.Context(), list); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to create price list")
		return
	}

	c.JSON(http.StatusCreated, list)
}

type adminUpdatePriceListRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Currency    *string `json:"currency"`
	Status      *string `json:"status"`
}

// AdminUpdatePriceList updates a price list.
func (h *Handler) AdminUpdatePriceList(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")
	list, err := h.services.Price.GetPriceList(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Price list not found")
		return
	}

	var req adminUpdatePriceListRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	if req.Name != nil {
		list.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		list.Description = strings.TrimSpace(*req.Description)
	}
	if req.Currency != nil {
		list.Currency = strings.TrimSpace(*req.Currency)
	}
	if req.Status != nil {
		list.Status = strings.TrimSpace(*req.Status)
	}
	list.UpdatedAt = time.Now()

	if err := h.services.Price.UpdatePriceList(c.Request.Context(), list); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update price list")
		return
	}

	c.JSON(http.StatusOK, list)
}

// AdminDeletePriceList deletes a price list.
func (h *Handler) AdminDeletePriceList(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")
	if err := h.services.Price.DeletePriceList(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Price list not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Price list deleted", "id": id})
}

// --- Product Prices ---

// AdminGetProductPrices returns all price rules for a product.
func (h *Handler) AdminGetProductPrices(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	productID := c.Param("id")
	rules, err := h.services.Price.GetProductPrices(c.Request.Context(), productID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch product prices")
		return
	}

	c.JSON(http.StatusOK, rules)
}

type adminSetProductPriceRequest struct {
	PriceListID string  `json:"priceListId" binding:"required"`
	MinQuantity int     `json:"minQuantity"`
	UnitPrice   float64 `json:"unitPrice" binding:"required"`
	Currency    string  `json:"currency"`
}

// AdminSetProductPrice creates or updates a price rule for a product.
func (h *Handler) AdminSetProductPrice(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	productID := c.Param("id")
	var req adminSetProductPriceRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	// Verify product exists
	if _, err := h.services.Product.GetProductByID(c.Request.Context(), productID); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Product not found")
		return
	}

	currency := strings.TrimSpace(req.Currency)
	if currency == "" {
		currency = "USD"
	}
	minQty := req.MinQuantity
	if minQty < 1 {
		minQty = 1
	}

	rule := &modelsProduct.PriceRule{
		ID:          utils.GenerateID(),
		ProductID:   productID,
		PriceListID: strings.TrimSpace(req.PriceListID),
		MinQuantity: minQty,
		UnitPrice:   req.UnitPrice,
		Currency:    currency,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.services.Price.SetProductPrice(c.Request.Context(), rule); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to set product price")
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// AdminDeleteProductPrice deletes a specific price rule for a product.
func (h *Handler) AdminDeleteProductPrice(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	priceID := c.Param("priceId")
	if err := h.services.Price.DeleteProductPrice(c.Request.Context(), priceID); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Price rule not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Price rule deleted", "id": priceID})
}
