package admin

import (
	"errors"
	modelsProduct "candypro/api/internal/models/product"
	orderSvc "candypro/api/internal/services/order"
	"candypro/api/internal/utils"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// adminActorID 当前管理员用户 ID（审计用）
func adminActorID(c *gin.Context) string {
	if v, ok := c.Get("userID"); ok {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

// AdminListWarehouses GET /warehouses
func (h *Handler) AdminListWarehouses(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	rows, err := h.services.Product.ListWarehouses(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to list warehouses")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// AdminUpsertWarehouse POST /warehouses
func (h *Handler) AdminUpsertWarehouse(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	var body modelsProduct.Warehouse
	if !utils.BindJSONOrInvalidRequest(c, &body) {
		return
	}
	if strings.TrimSpace(body.ID) == "" {
		utils.InvalidRequestResponse(c, "id is required")
		return
	}
	if strings.TrimSpace(body.Code) == "" {
		utils.InvalidRequestResponse(c, "code is required")
		return
	}
	now := time.Now()
	if body.CreatedAt.IsZero() {
		body.CreatedAt = now
	}
	body.UpdatedAt = now
	if err := h.services.Product.SaveWarehouse(c.Request.Context(), &body); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to save warehouse")
		return
	}
	c.JSON(http.StatusOK, body)
}

// AdminUpsertWarehouseStock PUT /warehouses/:id/stock
func (h *Handler) AdminUpsertWarehouseStock(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	wid := strings.TrimSpace(c.Param("id"))
	var req struct {
		ProductID string `json:"productId" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required"`
		Reserved  int    `json:"reserved"`
		VariantID string `json:"variantId"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}
	row := modelsProduct.WarehouseStock{
		WarehouseID: wid,
		ProductID:   strings.TrimSpace(req.ProductID),
		Quantity:    req.Quantity,
		Reserved:    req.Reserved,
		UpdatedAt:   time.Now(),
	}
	if v := strings.TrimSpace(req.VariantID); v != "" {
		row.VariantID = &v
	}
	if err := h.services.Product.UpsertWarehouseStock(c.Request.Context(), &row); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to upsert warehouse stock")
		return
	}
	c.JSON(http.StatusOK, row)
}

// AdminGetProductMarketProfile GET /products/:id/market-profile?marketCode=EU
func (h *Handler) AdminGetProductMarketProfile(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	mc := strings.TrimSpace(c.Query("marketCode"))
	rows, err := h.services.Product.FindMarketProfilesForProducts(c.Request.Context(), []string{pid}, mc)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to load profiles")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// AdminPutProductMarketProfile PUT /products/:id/market-profile
func (h *Handler) AdminPutProductMarketProfile(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	var req struct {
		MarketCode                string   `json:"marketCode" binding:"required"`
		DestinationCountries      []string `json:"destinationCountries"`
		BlockedIngredientPatterns []string `json:"blockedIngredientPatterns"`
		RequiredCertKeywords      []string `json:"requiredCertKeywords"`
		LabelTemplateID           string   `json:"labelTemplateId"`
		Notes                     string   `json:"notes"`
		RuleVersion               string   `json:"ruleVersion"`
		EffectiveFrom             string   `json:"effectiveFrom"`
		RuleSourceSummary         string   `json:"ruleSourceSummary"`
		Copilot                   bool     `json:"copilot"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}
	dc, err := json.Marshal(req.DestinationCountries)
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid destination countries data")
		return
	}
	bi, err := json.Marshal(req.BlockedIngredientPatterns)
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid blocked ingredient patterns data")
		return
	}
	rk, err := json.Marshal(req.RequiredCertKeywords)
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid required certification keywords data")
		return
	}
	now := time.Now()
	row := modelsProduct.ProductMarketProfile{
		ProductID:                 pid,
		MarketCode:                strings.TrimSpace(req.MarketCode),
		DestinationCountries:      datatypes.JSON(dc),
		BlockedIngredientPatterns: datatypes.JSON(bi),
		RequiredCertKeywords:      datatypes.JSON(rk),
		LabelTemplateID:           strings.TrimSpace(req.LabelTemplateID),
		Notes:                     req.Notes,
		RuleVersion:               strings.TrimSpace(req.RuleVersion),
		RuleSourceSummary:         strings.TrimSpace(req.RuleSourceSummary),
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
	if ts := strings.TrimSpace(req.EffectiveFrom); ts != "" {
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			row.EffectiveFrom = &t
		}
	}
	if err := h.services.Product.UpsertProductMarketProfile(c.Request.Context(), &row); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to save market profile")
		return
	}
	if req.Copilot && h.aiService != nil {
		out := gin.H{"profile": row}
		p, err := h.services.Product.GetProductByID(c.Request.Context(), pid)
		if err == nil && p != nil {
			q := strings.TrimSpace(p.Name) + " " + strings.TrimSpace(p.Ingredients) + " " + strings.TrimSpace(p.Allergens)
			cc := strings.TrimSpace(req.MarketCode)
			if len(req.DestinationCountries) > 0 {
				cc = strings.TrimSpace(req.DestinationCountries[0])
			}
			out["ragAssist"] = gin.H{
				"disclaimer": "RAG output is not legal advice; hard rules and profiles still govern release.",
				"lookup":     h.aiService.LookupCompliance(cc, q, 5),
			}
		}
		c.JSON(http.StatusOK, out)
		return
	}
	c.JSON(http.StatusOK, row)
}

// AdminGetProductMarketCosts GET /products/:id/market-costs
func (h *Handler) AdminGetProductMarketCosts(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	rows, err := h.services.Product.FindMarketCostStacksForProduct(c.Request.Context(), pid)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to load cost stacks")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// AdminPutProductMarketCost PUT /products/:id/market-costs
func (h *Handler) AdminPutProductMarketCost(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	var req modelsProduct.ProductMarketCostStack
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}
	req.ProductID = pid
	if strings.TrimSpace(req.MarketCode) == "" {
		utils.InvalidRequestResponse(c, "marketCode is required")
		return
	}
	now := time.Now()
	req.CreatedAt = now
	req.UpdatedAt = now
	if err := h.services.Product.UpsertProductMarketCostStack(c.Request.Context(), &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to save cost stack")
		return
	}
	c.JSON(http.StatusOK, req)
}

// AdminCreateOEMInventoryHold POST /oem-projects/:id/inventory-holds
func (h *Handler) AdminCreateOEMInventoryHold(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	prj := strings.TrimSpace(c.Param("id"))
	var req struct {
		ProductID string `json:"productId" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required"`
		Notes     string `json:"notes"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}
	now := time.Now()
	row := modelsProduct.OEMProjectInventoryHold{
		ProjectID: prj,
		ProductID: strings.TrimSpace(req.ProductID),
		Quantity:  req.Quantity,
		Status:    "active",
		Notes:     req.Notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := h.services.Product.SaveOEMProjectInventoryHold(c.Request.Context(), &row); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to create hold")
		return
	}
	c.JSON(http.StatusCreated, row)
}

// AdminListOEMInventoryHolds GET /oem-projects/:id/inventory-holds
func (h *Handler) AdminListOEMInventoryHolds(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	prj := strings.TrimSpace(c.Param("id"))
	rows, err := h.services.Product.ListOEMInventoryHoldsByProject(c.Request.Context(), prj)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to list holds")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// AdminListProductChannelInventory GET /products/:id/channel-inventory
func (h *Handler) AdminListProductChannelInventory(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	rows, err := h.services.Product.ListChannelInventoriesForProduct(c.Request.Context(), pid)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to list channel inventory")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// AdminPutProductChannelInventory PUT /products/:id/channel-inventory
func (h *Handler) AdminPutProductChannelInventory(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	var body modelsProduct.ChannelInventory
	if !utils.BindJSONOrInvalidRequest(c, &body) {
		return
	}
	body.ProductID = pid
	if strings.TrimSpace(body.ChannelCode) == "" {
		utils.InvalidRequestResponse(c, "channelCode is required")
		return
	}
	now := time.Now()
	if body.CreatedAt.IsZero() {
		body.CreatedAt = now
	}
	body.UpdatedAt = now
	if err := h.services.Product.UpsertChannelInventory(c.Request.Context(), &body); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to save channel inventory")
		return
	}
	c.JSON(http.StatusOK, body)
}

// AdminProductComplianceSuggest POST /products/:id/compliance-suggest（RAG 建议，非法律依据）
func (h *Handler) AdminProductComplianceSuggest(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	if h.aiService == nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "ai_disabled", "AI service not available")
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	var req struct {
		TargetCountry string `json:"targetCountry" binding:"required"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}
	p, err := h.services.Product.GetProductByID(c.Request.Context(), pid)
	if err != nil || p == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Product not found")
		return
	}
	q := strings.TrimSpace(p.Name) + " " + strings.TrimSpace(p.Ingredients) + " " + strings.TrimSpace(p.Allergens)
	lookup := h.aiService.LookupCompliance(strings.TrimSpace(req.TargetCountry), q, 5)
	c.JSON(http.StatusOK, gin.H{
		"disclaimer": "Suggestions are not legal advice; confirm with qualified compliance staff.",
		"lookup":     lookup,
		"productId":  pid,
	})
}

// AdminCreateInvoiceFromOrder POST /invoices/from-order
func (h *Handler) AdminCreateInvoiceFromOrder(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	var req struct {
		OrderID string `json:"orderId" binding:"required"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}
	inv, err := h.services.Invoice.CreateInvoiceFromOrder(c.Request.Context(), strings.TrimSpace(req.OrderID), adminActorID(c))
	if err != nil {
		if errors.Is(err, orderSvc.ErrInvoiceOrderNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "not_found", err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	c.JSON(http.StatusCreated, inv)
}
