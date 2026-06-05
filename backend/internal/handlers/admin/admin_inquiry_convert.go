package admin

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/services/orderintake"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type adminConvertInquiryRequest struct {
	ShippingAddress *modelsOrder.Address `json:"shippingAddress"`
	Notes           string               `json:"notes"`
}

// AdminConvertInquiryToOrder converts a quoted/won inquiry into a draft order.
// It reuses the shared orderintake composition so the negotiation-accept path
// and this manual path build orders identically (P0.1).
func (h *Handler) AdminConvertInquiryToOrder(c *gin.Context) {
	if h.services == nil || h.services.OrderIntake == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	inquiryID := c.Param("id")
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}

	// Only quoted, negotiating, or won inquiries can be converted.
	if inquiry.Status != "quoted" && inquiry.Status != "won" && inquiry.Status != "negotiating" {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "inquiry_status_invalid")
		return
	}

	// Must have a user assigned.
	if inquiry.UserID == nil || strings.TrimSpace(*inquiry.UserID) == "" {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "inquiry_missing_user")
		return
	}

	var req adminConvertInquiryRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	opts := orderintake.Options{ShippingAddress: req.ShippingAddress}
	if h.cfg != nil {
		opts.EnableMultiWarehouse = h.cfg.Security.EnableMultiWarehouse
	}

	res, err := h.services.OrderIntake.CreateOrderFromAcceptedOffer(c.Request.Context(), inquiry, nil, opts)
	if err != nil {
		switch {
		case errors.Is(err, orderintake.ErrMissingCountry):
			response.ErrorResp(c, http.StatusUnprocessableEntity, "inquiry_missing_country")
		case errors.Is(err, orderintake.ErrNoItems):
			response.ErrorResp(c, http.StatusUnprocessableEntity, "inquiry_no_items")
		case errors.Is(err, orderintake.ErrComplianceViolation):
			response.ErrorRespDetail(c, http.StatusUnprocessableEntity, "compliance_violation", gin.H{
				"violations": res.ComplianceViolations,
				"warnings":   res.ComplianceWarnings,
			})
		case errors.Is(err, orderintake.ErrInventoryViolation):
			response.ErrorRespDetail(c, http.StatusUnprocessableEntity, "inventory_violation", gin.H{
				"violations": res.InventoryViolations,
			})
		default:
			response.ErrorResp(c, http.StatusInternalServerError, "order_create_failed")
		}
		return
	}

	// Mark the inquiry won so it cannot be re-converted.
	if mErr := h.services.OrderIntake.MarkInquiryWon(c.Request.Context(), inquiryID); mErr != nil {
		log.Printf("admin: failed to mark inquiry %s as won: %v", inquiryID, mErr)
	}

	c.JSON(http.StatusCreated, res.Order)
}
