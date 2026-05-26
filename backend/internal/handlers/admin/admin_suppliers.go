package admin

import (
	"fmt"
	"net/http"
	"time"

	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ---- Suppliers ----

func (h *Handler) AdminCreateSupplier(c *gin.Context) {
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		Name          string  `json:"name" binding:"required"`
		ContactPerson string  `json:"contactPerson"`
		Email         string  `json:"email"`
		Phone         string  `json:"phone"`
		Address       string  `json:"address"`
		Country       string  `json:"country"`
		PaymentTerms  string  `json:"paymentTerms"`
		Rating        float64 `json:"rating"`
		Notes         string  `json:"notes"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	s := &modelsProduct.Supplier{
		ID:            fmt.Sprintf("SUP%d", time.Now().UnixNano()),
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		Country:       req.Country,
		PaymentTerms:  req.PaymentTerms,
		Rating:        req.Rating,
		Notes:         req.Notes,
	}
	if err := h.services.Supplier.CreateSupplier(c.Request.Context(), s); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "supplier_create_failed")
		return
	}
	c.JSON(http.StatusCreated, s)
}

func (h *Handler) AdminListSuppliers(c *gin.Context) {
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	suppliers, total, err := h.services.Supplier.FindAllSuppliers(c.Request.Context(), page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "supplier_list_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": suppliers, "pagination": pagination.BuildPagination(total, page, limit)})
}

func (h *Handler) AdminGetSupplier(c *gin.Context) {
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	s, err := h.services.Supplier.FindSupplierByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "supplier_not_found")
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) AdminUpdateSupplier(c *gin.Context) {
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	s, err := h.services.Supplier.FindSupplierByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "supplier_not_found")
		return
	}
	if !response.BindJSONOrInvalid(c, s) {
		return
	}
	s.ID = id
	if err := h.services.Supplier.UpdateSupplier(c.Request.Context(), s); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "supplier_update_failed")
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) AdminDeleteSupplier(c *gin.Context) {
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	if err := h.services.Supplier.DeleteSupplier(c.Request.Context(), c.Param("id")); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "supplier_delete_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ---- Purchase Orders ----

func (h *Handler) AdminCreatePO(c *gin.Context) {
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID := c.GetString("userID")
	var req struct {
		SupplierID   string `json:"supplierId" binding:"required"`
		WarehouseID  string `json:"warehouseId"`
		Notes        string `json:"notes"`
		ExpectedDate string `json:"expectedDate"`
		Items        []struct {
			ProductID string  `json:"productId" binding:"required"`
			Quantity  int     `json:"quantity" binding:"required,min=1"`
			UnitCost  float64 `json:"unitCost"`
		} `json:"items" binding:"required,min=1,dive"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	po := &modelsProduct.PurchaseOrder{
		ID:          fmt.Sprintf("PO%d", time.Now().UnixNano()),
		PONumber:    fmt.Sprintf("PO-%d", time.Now().UnixNano()),
		SupplierID:  req.SupplierID,
		WarehouseID: req.WarehouseID,
		Notes:       req.Notes,
		CreatedBy:   userID,
	}
	if req.ExpectedDate != "" {
		// R2 A-3: previously the error was discarded, silently dropping the
		// expected delivery date on malformed input. Reject so the buyer sees
		// the typo instead of a PO with no commitment date.
		t, err := time.Parse(time.RFC3339, req.ExpectedDate)
		if err != nil {
			response.InvalidResp(c, "invalid_expected_date")
			return
		}
		po.ExpectedDate = &t
	}
	items := make([]modelsProduct.PurchaseOrderItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = modelsProduct.PurchaseOrderItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitCost:  it.UnitCost,
		}
	}
	if err := h.services.Supplier.CreatePO(c.Request.Context(), po, items); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "po_create_failed")
		return
	}
	c.JSON(http.StatusCreated, po)
}

func (h *Handler) AdminListPOs(c *gin.Context) {
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	pos, total, err := h.services.Supplier.FindAllPOs(c.Request.Context(), page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "po_list_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pos, "pagination": pagination.BuildPagination(total, page, limit)})
}

func (h *Handler) AdminGetPO(c *gin.Context) {
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	po, items, err := h.services.Supplier.FindPOByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "po_not_found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"po": po, "items": items})
}

func (h *Handler) AdminReceivePO(c *gin.Context) {
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	var req struct {
		WarehouseID string         `json:"warehouseId" binding:"required"`
		Items       map[string]int `json:"items" binding:"required"` // productID -> quantity
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if err := h.services.Supplier.ReceivePO(c.Request.Context(), id, req.WarehouseID, req.Items); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "po_receive_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}
