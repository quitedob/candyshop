package admin

import (
	modelsProduct "candypro/api/internal/models/product"
	"crypto/rand"
	"encoding/hex"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ── Supplier Self-Service Portal ──
// These endpoints enable self-service onboarding and management for suppliers.

// SupplierGetProfile returns the authenticated supplier's profile.
func (h *Handler) SupplierGetProfile(c *gin.Context) {
	supplierID := c.GetString("supplierID")
	if supplierID == "" {
		response.ErrorResp(c, http.StatusUnauthorized, "supplier_not_authenticated")
		return
	}
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	sup, err := h.services.Supplier.FindSupplierByID(c.Request.Context(), supplierID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "supplier_not_found")
		return
	}
	c.JSON(http.StatusOK, sup)
}

// SupplierUpdateProfile allows a supplier to update their own profile.
func (h *Handler) SupplierUpdateProfile(c *gin.Context) {
	supplierID := c.GetString("supplierID")
	if supplierID == "" {
		response.ErrorResp(c, http.StatusUnauthorized, "supplier_not_authenticated")
		return
	}
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	sup, err := h.services.Supplier.FindSupplierByID(c.Request.Context(), supplierID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "supplier_not_found")
		return
	}
	var req struct {
		Name          *string `json:"name"`
		ContactPerson *string `json:"contactPerson"`
		Email         *string `json:"email"`
		Phone         *string `json:"phone"`
		Address       *string `json:"address"`
		Country       *string `json:"country"`
		PaymentTerms  *string `json:"paymentTerms"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Name != nil {
		sup.Name = *req.Name
	}
	if req.ContactPerson != nil {
		sup.ContactPerson = *req.ContactPerson
	}
	if req.Email != nil {
		sup.Email = *req.Email
	}
	if req.Phone != nil {
		sup.Phone = *req.Phone
	}
	if req.Address != nil {
		sup.Address = *req.Address
	}
	if req.Country != nil {
		sup.Country = *req.Country
	}
	if req.PaymentTerms != nil {
		sup.PaymentTerms = *req.PaymentTerms
	}
	if err := h.services.Supplier.UpdateSupplier(c.Request.Context(), sup); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "supplier_update_failed")
		return
	}
	c.JSON(http.StatusOK, sup)
}

// SupplierGetMyPOs lists purchase orders for the authenticated supplier.
func (h *Handler) SupplierGetMyPOs(c *gin.Context) {
	supplierID := c.GetString("supplierID")
	if supplierID == "" {
		response.ErrorResp(c, http.StatusUnauthorized, "supplier_not_authenticated")
		return
	}
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	pos, err := h.services.Supplier.FindPOsBySupplierID(c.Request.Context(), supplierID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "supplier_pos_fetch_failed")
		return
	}
	if pos == nil {
		pos = []modelsProduct.PurchaseOrder{}
	}
	c.JSON(http.StatusOK, gin.H{"data": pos})
}

// SupplierUpdatePOStatus allows a supplier to update a PO status (e.g. accepted, shipped).
func (h *Handler) SupplierUpdatePOStatus(c *gin.Context) {
	supplierID := c.GetString("supplierID")
	if supplierID == "" {
		response.ErrorResp(c, http.StatusUnauthorized, "supplier_not_authenticated")
		return
	}
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	poID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	validStatuses := map[string]bool{
		modelsProduct.POStatusSent:     true,
		modelsProduct.POStatusPartial:  true,
		modelsProduct.POStatusReceived: true,
	}
	if !validStatuses[req.Status] {
		response.InvalidResp(c, "invalid_po_status")
		return
	}
	if err := h.services.Supplier.UpdatePOStatus(c.Request.Context(), poID, supplierID, req.Status); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "po_status_update_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": req.Status})
}

// ── Supplier Registration (self-service onboarding) ──

// SupplierRegisterSelf allows a factory/supplier to register themselves.
func (h *Handler) SupplierRegisterSelf(c *gin.Context) {
	if h.cfg != nil && !h.cfg.Security.EnableSupplierPortal {
		response.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}
	if h.services == nil || h.services.Supplier == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		Name          string `json:"name" binding:"required"`
		ContactPerson string `json:"contactPerson"`
		Email         string `json:"email" binding:"required"`
		Phone         string `json:"phone"`
		Address       string `json:"address"`
		Country       string `json:"country"`
		PaymentTerms  string `json:"paymentTerms"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	count, err := h.services.Supplier.CountSuppliers(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "supplier_count_failed")
		return
	}

	sup := &modelsProduct.Supplier{
		ID:            "SUP-" + strconv.Itoa(int(count)+1),
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		Country:       req.Country,
		PaymentTerms:  req.PaymentTerms,
		IsActive:      true,
	}

	// Generate API key for self-service access
	keyBytes := make([]byte, 16)
	_, _ = rand.Read(keyBytes)
	sup.APIKey = hex.EncodeToString(keyBytes)

	if err := h.services.Supplier.CreateSupplier(c.Request.Context(), sup); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "supplier_register_failed")
		return
	}
	c.JSON(http.StatusCreated, sup)
}
