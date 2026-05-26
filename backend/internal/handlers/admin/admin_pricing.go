package admin

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// --- Price Lists ---

// AdminGetPriceLists returns all price lists with pagination.
func (h *Handler) AdminGetPriceLists(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	page, limit := pagination.ParsePagination(c, 20, 100)
	result, err := h.services.Price.GetPriceLists(c.Request.Context(), page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "price_fetch_failed")
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
		response.ServiceUnavailableResp(c)
		return
	}

	var req adminCreatePriceListRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	currency := strings.TrimSpace(req.Currency)
	if currency == "" {
		currency = "USD"
	}

	list := &modelsProduct.PriceList{
		ID:          crypto.GenerateID(),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Currency:    currency,
		Status:      strings.TrimSpace(req.Status),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.services.Price.CreatePriceList(c.Request.Context(), list); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "price_list_create_failed")
		return
	}

	h.logActivityAudit(c, "create", "pricing", list.ID, "", list.Name)
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
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	list, err := h.services.Price.GetPriceList(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "price_list_not_found")
		return
	}
	oldName := list.Name

	var req adminUpdatePriceListRequest
	if !response.BindJSONOrInvalid(c, &req) {
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
		response.ErrorResp(c, http.StatusInternalServerError, "price_list_update_failed")
		return
	}

	h.logActivityAudit(c, "update", "pricing", id, oldName, list.Name)
	c.JSON(http.StatusOK, list)
}

// AdminDeletePriceList deletes a price list.
func (h *Handler) AdminDeletePriceList(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	rules, err := h.services.Price.GetPriceListRules(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "price_fetch_failed")
		return
	}
	if len(rules) > 0 {
		response.InvalidResp(c, "price_list_has_rules")
		return
	}
	if err := h.services.Price.DeletePriceList(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "price_list_not_found")
		return
	}

	h.logActivityAudit(c, "delete", "pricing", id, id, "")
	c.JSON(http.StatusOK, gin.H{"message": "Price list deleted", "id": id})
}

// --- Product Prices ---

// AdminGetProductPrices returns all price rules for a product.
func (h *Handler) AdminGetProductPrices(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	productID := c.Param("id")
	rules, err := h.services.Price.GetProductPrices(c.Request.Context(), productID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "price_fetch_failed")
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
		response.ServiceUnavailableResp(c)
		return
	}

	productID := c.Param("id")
	var req adminSetProductPriceRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	// Verify product exists
	if _, err := h.services.Product.GetProductByID(c.Request.Context(), productID); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "product_not_found")
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

	priceListID := strings.TrimSpace(req.PriceListID)
	if _, err := h.services.Price.GetPriceList(c.Request.Context(), priceListID); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "price_list_not_found")
		return
	}

	var oldPrice float64
	if existing, err := h.services.Price.GetProductPrices(c.Request.Context(), productID); err == nil {
		for _, r := range existing {
			if r.PriceListID == priceListID && r.MinQuantity == minQty {
				oldPrice = r.UnitPrice
				break
			}
		}
	}

	rule := &modelsProduct.PriceRule{
		ID:          crypto.GenerateID(),
		ProductID:   productID,
		PriceListID: priceListID,
		MinQuantity: minQty,
		UnitPrice:   req.UnitPrice,
		Currency:    currency,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.services.Price.SetProductPrice(c.Request.Context(), rule); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "price_set_failed")
		return
	}

	h.logActivityAudit(c, "update", "pricing", productID, fmt.Sprintf("%.2f", oldPrice),
		fmt.Sprintf("%s:%.2f", priceListID, req.UnitPrice))
	c.JSON(http.StatusCreated, rule)
}

// AdminDeleteProductPrice deletes a specific price rule for a product.
func (h *Handler) AdminDeleteProductPrice(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	priceID := c.Param("priceId")
	productID := c.Param("id")
	if err := h.services.Price.DeleteProductPrice(c.Request.Context(), priceID); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "price_rule_not_found")
		return
	}
	h.logActivityAudit(c, "price_delete", productID, "", priceID, "")

	c.JSON(http.StatusOK, gin.H{"message": "Price rule deleted", "id": priceID})
}
