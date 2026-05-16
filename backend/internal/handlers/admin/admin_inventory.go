package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminGetInventory returns paginated product inventory for admin.
func (h *Handler) AdminGetInventory(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	result, err := h.services.Product.GetProductsFiltered(c.Request.Context(), page, limit, false, false, false, "", "", 0, 0)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inventory_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, result)
}

// AdminUpdateInventory adjusts stock quantity for a product.
func (h *Handler) AdminUpdateInventory(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	productID := c.Param("productId")
	var req struct {
		StockQuantity int    `json:"stockQuantity"`
		Reason        string `json:"reason"`
		Notes         string `json:"notes"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	product, err := h.services.Product.GetProductByID(c.Request.Context(), productID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "product_not_found")
		return
	}

	oldQty := product.StockQuantity
	product.StockQuantity = req.StockQuantity
	product.UpdatedAt = time.Now()

	if err := h.services.Product.UpdateProduct(c.Request.Context(), product); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inventory_update_failed")
		return
	}

	if h.services.StockTransaction != nil {
		change := req.StockQuantity - oldQty
		reason := strings.TrimSpace(req.Reason)
		if reason == "" {
			reason = "manual_adjustment"
		}
		operatorID := adminActorID(c)
		if err := h.services.StockTransaction.Record(c.Request.Context(), &modelsOrder.StockTransaction{
			ProductID:   productID,
			Change:      change,
			StockBefore: oldQty,
			StockAfter:  req.StockQuantity,
			Reason:      reason,
			ReferenceID: req.Notes,
			OperatorID:  operatorID,
		}); err != nil {
			log.Printf("Warning: failed to record stock transaction for product %s: %v", productID, err)
		}
	}

	c.JSON(http.StatusOK, product)
}

// AdminBatchUpdateInventory applies bulk field updates to selected products.
// POST /admin/inventory/batch-update
func (h *Handler) AdminBatchUpdateInventory(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req struct {
		IDs     []string               `json:"ids"`
		Updates map[string]interface{} `json:"updates"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if len(req.IDs) == 0 {
		response.InvalidResp(c, "xlsx_import_no_rows")
		return
	}
	if len(req.Updates) == 0 {
		response.InvalidResp(c, "batch_update_no_fields")
		return
	}

	operatorID := adminActorID(c)
	var updated, errors int
	var errorMessages []string

	for _, id := range req.IDs {
		product, err := h.services.Product.GetProductByID(c.Request.Context(), id)
		if err != nil {
			errors++
			errorMessages = append(errorMessages, fmt.Sprintf("%s: product not found", id))
			continue
		}

		oldStock := product.StockQuantity
		changed := false

		for field, val := range req.Updates {
			switch field {
			case "category":
				if s, ok := val.(string); ok && s != product.Category {
					product.Category = strings.TrimSpace(s)
					if product.CategorySlug == "" || product.CategorySlug == normalizeSlug(product.Category) {
						product.CategorySlug = normalizeSlug(s)
					}
					changed = true
				}
			case "categorySlug":
				if s, ok := val.(string); ok {
					product.CategorySlug = strings.TrimSpace(s)
					changed = true
				}
			case "status":
				if s, ok := val.(string); ok && modelsProduct.IsValidProductStatus(s) {
					product.Status = s
					changed = true
				}
			case "stockQuantity":
				if v, ok := toFloat64Safe(val); ok {
					qty := int(v)
					if qty >= 0 && qty != product.StockQuantity {
						product.StockQuantity = qty
						changed = true
					}
				}
			case "basePrice":
				if v, ok := toFloat64Safe(val); ok && v >= 0 {
					product.BasePrice = v
					changed = true
				}
			case "moq":
				if v, ok := toFloat64Safe(val); ok {
					qty := int(v)
					if qty >= 0 {
						product.MOQ = qty
						changed = true
					}
				}
			case "leadTime":
				if s, ok := val.(string); ok {
					product.LeadTime = strings.TrimSpace(s)
					changed = true
				}
			case "halalCertified":
				if b, ok := val.(bool); ok {
					product.HalalCertified = b
					changed = true
				}
			case "oemAvailable":
				if b, ok := val.(bool); ok {
					product.OEMAvailable = b
					changed = true
				}
			case "hsCode":
				if s, ok := val.(string); ok {
					product.HSCode = strings.TrimSpace(s)
					changed = true
				}
			case "shelfLife":
				if s, ok := val.(string); ok {
					product.ShelfLife = strings.TrimSpace(s)
					changed = true
				}
			case "storage":
				if s, ok := val.(string); ok {
					product.Storage = strings.TrimSpace(s)
					changed = true
				}
			}
		}

		if !changed {
			continue
		}

		product.UpdatedAt = time.Now()
		product.UpdatedBy = &operatorID
		if err := h.services.Product.UpdateProduct(c.Request.Context(), product); err != nil {
			errors++
			errorMessages = append(errorMessages, fmt.Sprintf("%s: update failed - %v", product.Name, err))
			continue
		}
		updated++

		// Record stock transaction if stock changed
		if product.StockQuantity != oldStock && h.services.StockTransaction != nil {
			change := product.StockQuantity - oldStock
			if err := h.services.StockTransaction.Record(c.Request.Context(), &modelsOrder.StockTransaction{
				ProductID:   product.ID,
				Change:      change,
				StockBefore: oldStock,
				StockAfter:  product.StockQuantity,
				Reason:      "batch_adjustment",
				ReferenceID: "batch_update",
				OperatorID:  operatorID,
			}); err != nil {
				log.Printf("Warning: failed to record batch stock tx for %s: %v", product.ID, err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"updated":       updated,
		"errors":        errors,
		"errorMessages": errorMessages,
	})
}

// AdminBatchDeleteInventory deletes multiple products by ID.
// POST /admin/inventory/batch-delete
func (h *Handler) AdminBatchDeleteInventory(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req struct {
		IDs []string `json:"ids"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if len(req.IDs) == 0 {
		response.InvalidResp(c, "xlsx_import_no_rows")
		return
	}

	operatorID := adminActorID(c)
	var deleted, errors int
	var errorMessages []string

	for _, id := range req.IDs {
		product, err := h.services.Product.GetProductByID(c.Request.Context(), id)
		if err != nil {
			errors++
			errorMessages = append(errorMessages, fmt.Sprintf("%s: product not found", id))
			continue
		}

		// Log stock transaction before deletion for audit trail
		if h.services.StockTransaction != nil {
			if err := h.services.StockTransaction.Record(c.Request.Context(), &modelsOrder.StockTransaction{
				ProductID:   id,
				Change:      -product.StockQuantity,
				StockBefore: product.StockQuantity,
				StockAfter:  0,
				Reason:      "product_deleted",
				ReferenceID: "batch_delete",
				OperatorID:  operatorID,
			}); err != nil {
				log.Printf("Warning: failed to record stock tx for deleted product %s: %v", id, err)
			}
		}

		if err := h.services.Product.DeleteProduct(c.Request.Context(), id); err != nil {
			errors++
			errorMessages = append(errorMessages, fmt.Sprintf("%s: delete failed - %v", product.Name, err))
			continue
		}
		deleted++
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted":       deleted,
		"errors":        errors,
		"errorMessages": errorMessages,
	})
}

func toFloat64Safe(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f, true
		}
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// AdminGetInventoryHistory returns stock transaction history for a product.
func (h *Handler) AdminGetInventoryHistory(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	productID := c.Param("productId")
	if h.services.StockTransaction == nil {
		c.JSON(http.StatusOK, []any{})
		return
	}
	page, limit := pagination.ParsePagination(c, 1, 50)
	txs, _, err := h.services.StockTransaction.FindByProductID(c.Request.Context(), productID, page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "stock_history_fetch_failed")
		return
	}

	type historyEntry struct {
		ID          uint      `json:"id"`
		Change      int       `json:"change"`
		PreviousQty int       `json:"previousQty"`
		NewQty      int       `json:"newQty"`
		Reason      string    `json:"reason"`
		Notes       string    `json:"notes"`
		CreatedAt   time.Time `json:"createdAt"`
	}
	entries := make([]historyEntry, len(txs))
	for i, tx := range txs {
		entries[i] = historyEntry{
			ID:          tx.ID,
			Change:      tx.Change,
			PreviousQty: tx.StockBefore,
			NewQty:      tx.StockAfter,
			Reason:      tx.Reason,
			Notes:       tx.ReferenceID,
			CreatedAt:   tx.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, entries)
}

// AdminCreateStockTransfer handles POST /api/v1/admin/inventory/transfer
func (h *Handler) AdminCreateStockTransfer(c *gin.Context) {
	if h.services == nil || h.services.StockTransfer == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		FromWarehouseID string `json:"fromWarehouseId" binding:"required"`
		ToWarehouseID   string `json:"toWarehouseId" binding:"required"`
		Notes           string `json:"notes"`
		Items           []struct {
			ProductID string `json:"productId" binding:"required"`
			Quantity  int    `json:"quantity" binding:"required,min=1"`
			BatchID   string `json:"batchId"`
		} `json:"items" binding:"required,min=1,dive"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	userID := c.GetString("userID")
	if userID == "" {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	transfer := &modelsOrder.StockTransfer{
		ID:              fmt.Sprintf("ST%d", time.Now().UnixNano()),
		TransferNumber:  fmt.Sprintf("TR-%d", time.Now().UnixNano()),
		FromWarehouseID: req.FromWarehouseID,
		ToWarehouseID:   req.ToWarehouseID,
		Status:          modelsOrder.StockTransferStatusCompleted,
		Notes:           req.Notes,
		CreatedBy:       userID,
	}
	items := make([]modelsOrder.StockTransferItem, len(req.Items))
	for i, it := range req.Items {
		var batchID *string
		if it.BatchID != "" {
			batchID = &it.BatchID
		}
		items[i] = modelsOrder.StockTransferItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			BatchID:   batchID,
		}
	}

	if err := h.services.StockTransfer.Create(c.Request.Context(), transfer, items); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "transfer_create_failed")
		return
	}
	c.JSON(http.StatusCreated, transfer)
}

// AdminListStockTransfers handles GET /api/v1/admin/inventory/transfers
func (h *Handler) AdminListStockTransfers(c *gin.Context) {
	if h.services == nil || h.services.StockTransfer == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	transfers, total, err := h.services.StockTransfer.FindAll(c.Request.Context(), page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "transfer_list_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       transfers,
		"pagination": pagination.BuildPagination(total, page, limit),
	})
}

// AdminGetStockTransfer handles GET /api/v1/admin/inventory/transfers/:id
func (h *Handler) AdminGetStockTransfer(c *gin.Context) {
	if h.services == nil || h.services.StockTransfer == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	if id == "" {
		response.InvalidResp(c, "missing_id")
		return
	}
	transfer, items, err := h.services.StockTransfer.FindByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "transfer_not_found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"transfer": transfer,
		"items":    items,
	})
}
