package admin

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	"log"
	"net/http"
	"strings"
	"time"

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
