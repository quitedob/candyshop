package admin

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/response"
	oemSvc "candypro/api/internal/services/oem"
	orderSvc "candypro/api/internal/services/order"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ===== OEM Sample lifecycle (P0.2 / G-OEM-1) =====

type adminAddOEMSampleRequest struct {
	Name string `json:"name" binding:"required"`
}

// AdminAddOEMSample creates a sample on an OEM project so the "sampling" stage
// can actually be driven by data. POST /oem-projects/:id/samples
func (h *Handler) AdminAddOEMSample(c *gin.Context) {
	if h.services == nil || h.services.Project == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	projectID := strings.TrimSpace(c.Param("id"))
	var req adminAddOEMSampleRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		response.InvalidResp(c, "oem_sample_name_required")
		return
	}
	sample, err := h.services.Project.AddSample(c.Request.Context(), projectID, req.Name)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_sample_create_failed")
		return
	}
	c.JSON(http.StatusCreated, sample)
}

type adminUpdateOEMSampleRequest struct {
	Status   string `json:"status" binding:"required"`
	Feedback string `json:"feedback"`
}

// AdminUpdateOEMSample advances a sample's lifecycle state.
// PUT /oem-projects/:id/samples/:sampleId
func (h *Handler) AdminUpdateOEMSample(c *gin.Context) {
	if h.services == nil || h.services.Project == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	projectID := strings.TrimSpace(c.Param("id"))
	sampleID := strings.TrimSpace(c.Param("sampleId"))
	var req adminUpdateOEMSampleRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	sample, err := h.services.Project.UpdateSampleStatus(c.Request.Context(), projectID, sampleID, req.Status, req.Feedback)
	if err != nil {
		switch {
		case errors.Is(err, oemSvc.ErrOEMSampleStatusInvalid):
			response.InvalidResp(c, "oem_sample_status_invalid")
		case errors.Is(err, oemSvc.ErrOEMSampleNotFound):
			response.ErrorResp(c, http.StatusNotFound, "oem_sample_not_found")
		default:
			response.ErrorResp(c, http.StatusInternalServerError, "oem_sample_update_failed")
		}
		return
	}
	c.JSON(http.StatusOK, sample)
}

// ===== OEM inventory hold release (P0.2 / G-OEM-3) =====

// AdminReleaseOEMInventoryHold releases an active OEM finished-goods hold so it
// stops eroding sellable stock. DELETE /oem-projects/:id/inventory-holds/:holdId
func (h *Handler) AdminReleaseOEMInventoryHold(c *gin.Context) {
	if h.services == nil || h.services.Product == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	holdIDRaw := strings.TrimSpace(c.Param("holdId"))
	holdID, err := strconv.ParseUint(holdIDRaw, 10, 64)
	if err != nil || holdID == 0 {
		response.InvalidResp(c, "invalid_request")
		return
	}
	released, rerr := h.services.Product.ReleaseOEMInventoryHold(c.Request.Context(), uint(holdID))
	if rerr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_hold_release_failed")
		return
	}
	if !released {
		// Already released or never existed/active — surface as conflict so the
		// UI refreshes rather than implying a state change happened.
		response.ErrorResp(c, http.StatusConflict, "oem_hold_not_active")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "holdId": holdID, "status": "released"})
}

// ===== OEM project → order conversion (P0.2 / G-OEM-2) =====

type adminConvertOEMRequest struct {
	ProductID       string               `json:"productId"`
	Quantity        int                  `json:"quantity"`
	UnitPrice       float64              `json:"unitPrice"`
	Currency        string               `json:"currency"`
	Incoterms       string               `json:"incoterms"`
	ShippingAddress *modelsOrder.Address `json:"shippingAddress"`
}

// AdminConvertOEMProjectToOrder turns an OEM project that has reached the
// contract stage into a real draft order + linked trade transaction, so the
// OEM pipeline produces a financial artifact instead of ending at a status
// string (P0.2 / G-OEM-2). POST /oem-projects/:id/convert-to-order
func (h *Handler) AdminConvertOEMProjectToOrder(c *gin.Context) {
	if h.services == nil || h.services.Project == nil || h.services.Product == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	projectID := strings.TrimSpace(c.Param("id"))
	project, err := h.services.Project.GetProject(c.Request.Context(), projectID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "oem_project_not_found")
		return
	}

	// Only projects that have reached an agreed commercial stage can convert.
	switch project.Status {
	case modelsProduct.OEMStatusQuotation,
		modelsProduct.OEMStatusContract,
		modelsProduct.OEMStatusProduction:
	default:
		response.ErrorResp(c, http.StatusUnprocessableEntity, "oem_status_not_convertible")
		return
	}

	// Idempotency: a project converts into at most one order.
	if project.OrderID != nil && strings.TrimSpace(*project.OrderID) != "" {
		response.ErrorRespDetail(c, http.StatusConflict, "oem_already_converted", gin.H{
			"orderId": *project.OrderID,
		})
		return
	}

	var req adminConvertOEMRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	// Resolve the product: explicit override → project's linked product.
	productID := strings.TrimSpace(req.ProductID)
	if productID == "" && project.ProductID != nil {
		productID = strings.TrimSpace(*project.ProductID)
	}
	if productID == "" {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "oem_convert_product_required")
		return
	}

	product, prodErr := h.services.Product.GetProductByID(c.Request.Context(), productID)
	if prodErr != nil {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "product_not_found")
		return
	}
	if st := strings.ToLower(strings.TrimSpace(product.Status)); st != "" && st != "active" {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "product_unavailable")
		return
	}

	// Quantity: request → quoted → requirement MOQ.
	quantity := req.Quantity
	if quantity <= 0 {
		quantity = project.QuotedQuantity
	}
	if quantity <= 0 {
		quantity = project.Requirements.MOQ
	}
	if quantity <= 0 {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "oem_convert_quantity_required")
		return
	}

	// Unit price: request → quoted → product base price.
	unitPrice := req.UnitPrice
	if unitPrice <= 0 {
		unitPrice = project.QuotedUnitPrice
	}
	if unitPrice <= 0 {
		unitPrice = product.BasePrice
	}
	if unitPrice <= 0 {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "no_price")
		return
	}

	// Destination country: address → product market resolution baseline.
	destCountry := ""
	if req.ShippingAddress != nil {
		destCountry = strings.TrimSpace(req.ShippingAddress.Country)
	}
	if destCountry == "" {
		destCountry = strings.TrimSpace(project.Requirements.TargetMarket)
	}
	if destCountry == "" {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "oem_convert_country_required")
		return
	}

	// Apply destination market cost-stack to the unit price (parity with the
	// inquiry conversion path).
	unitPrice = h.services.Product.ResolveCheckoutUnitPrice(c.Request.Context(), product, unitPrice, destCountry)

	items := []modelsOrder.OrderItem{{
		ProductID:      productID,
		Quantity:       quantity,
		UnitPrice:      unitPrice,
		Specifications: oemSpecsSummary(project),
	}}

	// Destination compliance (hard block).
	compliance := h.services.Product.ValidateComplianceWithMarketProfiles(c.Request.Context(), destCountry, []modelsProduct.Product{*product})
	if len(compliance.Violations) > 0 {
		response.ErrorRespDetail(c, http.StatusUnprocessableEntity, "compliance_violation", gin.H{
			"country":    compliance.Country,
			"violations": compliance.Violations,
			"warnings":   compliance.Warnings,
		})
		return
	}

	subtotal := unitPrice * float64(quantity)
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	shippingAddr := modelsOrder.Address{Country: destCountry}
	if req.ShippingAddress != nil {
		shippingAddr = *req.ShippingAddress
	}

	orderNumber := "ORD-" + time.Now().Format("20060102") + "-" + strings.ToUpper(crypto.GenerateSlug())
	now := time.Now()
	order := &modelsOrder.Order{
		ID:              crypto.GenerateID(),
		OrderNumber:     orderNumber,
		UserID:          project.UserID,
		InquiryID:       project.InquiryID,
		Source:          modelsOrder.OrderSourceInquiry,
		Status:          modelsOrder.OrderStatusPendingConfirm,
		PaymentStatus:   modelsOrder.PaymentStatusUnpaid,
		Items:           items,
		StockReserved:   false,
		Subtotal:        subtotal,
		TotalAmount:     subtotal,
		Currency:        currency,
		ShippingAddress: shippingAddr,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	h.assignOrderWarehouseID(c, &order.WarehouseID)

	linked, err := h.services.Project.CreateLinkedOrder(c.Request.Context(), projectID, order)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "order_create_failed")
		return
	}

	// Only the request that atomically claimed the project may create its order.
	if !linked {
		detail := gin.H{}
		if current, currentErr := h.services.Project.GetProject(c.Request.Context(), projectID); currentErr == nil && current.OrderID != nil {
			detail["orderId"] = *current.OrderID
		}
		response.ErrorRespDetail(c, http.StatusConflict, "oem_already_converted", detail)
		return
	}

	// Best-effort trade transaction so the OEM order joins the trade pipeline.
	if h.services.Trade != nil {
		incoterms := strings.TrimSpace(req.Incoterms)
		trade := orderSvc.BuildTradeTransactionFromOEMProject(order, project, incoterms)
		if terr := h.services.Trade.CreateTransaction(c.Request.Context(), trade); terr != nil {
			c.Header("X-Trade-Create-Warning", "trade_creation_failed")
		}
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "order": order})
}

func oemSpecsSummary(project *modelsProduct.OEMProject) string {
	var parts []string
	if v := strings.TrimSpace(project.Requirements.Flavor); v != "" {
		parts = append(parts, "Flavor: "+v)
	}
	if v := strings.TrimSpace(project.Requirements.Shape); v != "" {
		parts = append(parts, "Shape: "+v)
	}
	if v := strings.TrimSpace(project.Requirements.Packaging); v != "" {
		parts = append(parts, "Packaging: "+v)
	}
	return strings.Join(parts, "; ")
}
