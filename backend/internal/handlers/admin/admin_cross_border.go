package admin

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/response"
	orderSvc "candypro/api/internal/services/order"
	"encoding/json"
	"errors"
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
		response.ServiceUnavailableResp(c)
		return
	}
	rows, err := h.services.Product.ListWarehouses(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "warehouse_list_failed")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// AdminUpsertWarehouse POST /warehouses
func (h *Handler) AdminUpsertWarehouse(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var body modelsProduct.Warehouse
	if !response.BindJSONOrInvalid(c, &body) {
		return
	}
	if strings.TrimSpace(body.ID) == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	if strings.TrimSpace(body.Code) == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	now := time.Now()
	if body.CreatedAt.IsZero() {
		body.CreatedAt = now
	}
	body.UpdatedAt = now
	if err := h.services.Product.SaveWarehouse(c.Request.Context(), &body); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "warehouse_save_failed")
		return
	}
	c.JSON(http.StatusOK, body)
}

// AdminUpsertWarehouseStock PUT /warehouses/:id/stock
func (h *Handler) AdminUpsertWarehouseStock(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	wid := strings.TrimSpace(c.Param("id"))
	var req struct {
		ProductID string `json:"productId" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required"`
		Reserved  int    `json:"reserved"`
		VariantID string `json:"variantId"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
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
		response.ErrorResp(c, http.StatusInternalServerError, "warehouse_stock_upsert_failed")
		return
	}
	c.JSON(http.StatusOK, row)
}

// AdminGetProductMarketProfile GET /products/:id/market-profile?marketCode=EU
func (h *Handler) AdminGetProductMarketProfile(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	mc := strings.TrimSpace(c.Query("marketCode"))
	rows, err := h.services.Product.FindMarketProfilesForProducts(c.Request.Context(), []string{pid}, mc)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "market_profile_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// AdminPutProductMarketProfile PUT /products/:id/market-profile
func (h *Handler) AdminPutProductMarketProfile(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
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
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	dc, err := json.Marshal(req.DestinationCountries)
	if err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}
	bi, err := json.Marshal(req.BlockedIngredientPatterns)
	if err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}
	rk, err := json.Marshal(req.RequiredCertKeywords)
	if err != nil {
		response.InvalidResp(c, "invalid_request")
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
		response.ErrorResp(c, http.StatusInternalServerError, "market_profile_save_failed")
		return
	}
	c.JSON(http.StatusOK, row)
}

// AdminGetProductMarketCosts GET /products/:id/market-costs
func (h *Handler) AdminGetProductMarketCosts(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	rows, err := h.services.Product.FindMarketCostStacksForProduct(c.Request.Context(), pid)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "cost_stack_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// AdminPutProductMarketCost PUT /products/:id/market-costs
func (h *Handler) AdminPutProductMarketCost(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	var req modelsProduct.ProductMarketCostStack
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	req.ProductID = pid
	if strings.TrimSpace(req.MarketCode) == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	now := time.Now()
	req.UpdatedAt = now
	existing, _ := h.services.Product.FindMarketCostStacksForProduct(c.Request.Context(), pid)
	for _, row := range existing {
		if row.MarketCode == req.MarketCode {
			req.CreatedAt = row.CreatedAt
			break
		}
	}
	if req.CreatedAt.IsZero() {
		req.CreatedAt = now
	}
	if err := h.services.Product.UpsertProductMarketCostStack(c.Request.Context(), &req); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "cost_stack_save_failed")
		return
	}
	c.JSON(http.StatusOK, req)
}

// AdminCreateOEMInventoryHold POST /oem-projects/:id/inventory-holds
func (h *Handler) AdminCreateOEMInventoryHold(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	prj := strings.TrimSpace(c.Param("id"))
	var req struct {
		ProductID string `json:"productId" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required"`
		Notes     string `json:"notes"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
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
		response.ErrorResp(c, http.StatusInternalServerError, "oem_hold_create_failed")
		return
	}
	c.JSON(http.StatusCreated, row)
}

// AdminListOEMInventoryHolds GET /oem-projects/:id/inventory-holds
func (h *Handler) AdminListOEMInventoryHolds(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	prj := strings.TrimSpace(c.Param("id"))
	rows, err := h.services.Product.ListOEMInventoryHoldsByProject(c.Request.Context(), prj)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_hold_list_failed")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// AdminListProductChannelInventory GET /products/:id/channel-inventory
func (h *Handler) AdminListProductChannelInventory(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	rows, err := h.services.Product.ListChannelInventoriesForProduct(c.Request.Context(), pid)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "channel_inventory_list_failed")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// AdminPutProductChannelInventory PUT /products/:id/channel-inventory
func (h *Handler) AdminPutProductChannelInventory(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	pid := strings.TrimSpace(c.Param("id"))
	var body modelsProduct.ChannelInventory
	if !response.BindJSONOrInvalid(c, &body) {
		return
	}
	body.ProductID = pid
	if strings.TrimSpace(body.ChannelCode) == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	now := time.Now()
	if body.CreatedAt.IsZero() {
		body.CreatedAt = now
	}
	body.UpdatedAt = now
	if err := h.services.Product.UpsertChannelInventory(c.Request.Context(), &body); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "channel_inventory_save_failed")
		return
	}
	c.JSON(http.StatusOK, body)
}

// AdminProductComplianceSuggest 已停用（平台不使用 RAG 合规检索）。
func (h *Handler) AdminProductComplianceSuggest(c *gin.Context) {
	response.ErrorResp(c, http.StatusGone, "feature_disabled")
}

// AdminCreateInvoiceFromOrder POST /invoices/from-order
func (h *Handler) AdminCreateInvoiceFromOrder(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		OrderID string `json:"orderId" binding:"required"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	inv, err := h.services.Invoice.CreateInvoiceFromOrder(c.Request.Context(), strings.TrimSpace(req.OrderID), adminActorID(c))
	if err != nil {
		if errors.Is(err, orderSvc.ErrInvoiceOrderNotFound) {
			response.ErrorResp(c, http.StatusNotFound, "not_found")
			return
		}
		response.InvalidResp(c, "invalid_request")
		return
	}
	c.JSON(http.StatusCreated, inv)
}
