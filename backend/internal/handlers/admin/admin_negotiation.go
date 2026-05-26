package admin

import (
	"errors"
	"net/http"

	orderSvc "candypro/api/internal/services/order"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// translateNegotiationError maps NegotiationService sentinel errors to stable
// user-facing error codes; unknown errors collapse to a generic code so internal
// implementation details are not leaked to clients (C-10).
func translateNegotiationError(err error) (int, string) {
	switch {
	case errors.Is(err, orderSvc.ErrNegotiationOfferNotFound):
		return http.StatusNotFound, "offer_not_found"
	case errors.Is(err, orderSvc.ErrNegotiationOfferNotPending):
		return http.StatusConflict, "offer_not_pending"
	default:
		return http.StatusInternalServerError, "internal_error"
	}
}

// AdminGetNegotiationOffers returns all negotiation offers for an inquiry.
func (h *Handler) AdminGetNegotiationOffers(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	inquiryID := c.Param("id")
	offers, err := h.services.Negotiation.GetOffers(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": offers})
}

// AdminCreateNegotiationOffer creates a counter-offer from admin to buyer.
func (h *Handler) AdminCreateNegotiationOffer(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID := c.GetString("userID")
	inquiryID := c.Param("id")

	var req orderSvc.CreateOfferRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.TotalAmount <= 0 {
		response.ErrorResp(c, http.StatusBadRequest, "offer_amount_required")
		return
	}
	offer, err := h.services.Negotiation.CreateOffer(c.Request.Context(), inquiryID, userID, "admin", req)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "offer_create_failed")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "offer": offer})
}

// AdminAcceptNegotiationOffer accepts a buyer's pending offer.
func (h *Handler) AdminAcceptNegotiationOffer(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID := c.GetString("userID")
	offerID := c.Param("offerId")
	offer, err := h.services.Negotiation.AcceptOffer(c.Request.Context(), offerID, userID)
	if err != nil {
		status, code := translateNegotiationError(err)
		response.ErrorResp(c, status, code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "offer": offer})
}

// AdminRejectNegotiationOffer rejects a buyer's pending offer.
func (h *Handler) AdminRejectNegotiationOffer(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID := c.GetString("userID")
	offerID := c.Param("offerId")
	offer, err := h.services.Negotiation.RejectOffer(c.Request.Context(), offerID, userID)
	if err != nil {
		status, code := translateNegotiationError(err)
		response.ErrorResp(c, status, code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "offer": offer})
}
