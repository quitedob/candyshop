package customer

import (
	"errors"
	"log"
	"net/http"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/response"
	orderSvc "candypro/api/internal/services/order"
	"candypro/api/internal/services/orderintake"

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

// CustomerGetNegotiationOffers returns all negotiation offers for user's inquiry.
func (h *Handler) CustomerGetNegotiationOffers(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	inquiryID := c.Param("id")

	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}
	if inquiry.UserID == nil || *inquiry.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	offers, err := h.services.Negotiation.GetOffers(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": offers})
}

// CustomerCreateNegotiationOffer creates a counter-offer from buyer to admin.
func (h *Handler) CustomerCreateNegotiationOffer(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	inquiryID := c.Param("id")

	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}
	if inquiry.UserID == nil || *inquiry.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	var req orderSvc.CreateOfferRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.TotalAmount <= 0 {
		response.ErrorResp(c, http.StatusBadRequest, "offer_amount_required")
		return
	}
	offer, err := h.services.Negotiation.CreateOffer(c.Request.Context(), inquiryID, userID, "customer", req)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "offer_create_failed")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "offer": offer})
}

// CustomerAcceptNegotiationOffer accepts an admin's pending offer and converts
// the agreed terms into a draft order (P0.1 / G-ORD-1). Previously this only
// flipped the offer status, leaving the negotiated price as dead data.
func (h *Handler) CustomerAcceptNegotiationOffer(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	inquiryID := c.Param("id")
	offerID := c.Param("offerId")

	// Verify ownership before mutating anything.
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}
	if inquiry.UserID == nil || *inquiry.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	offer, err := h.services.Negotiation.AcceptOffer(c.Request.Context(), offerID, userID)
	if err != nil {
		status, code := translateNegotiationError(err)
		response.ErrorResp(c, status, code)
		return
	}

	order := h.createOrderFromAcceptedOffer(c, inquiry, offer)
	resp := gin.H{"success": true, "offer": offer}
	if order != nil {
		resp["order"] = order
	}
	c.JSON(http.StatusOK, resp)
}

// createOrderFromAcceptedOffer builds a draft order from the accepted offer.
// Failures are logged but do not fail the accept itself — the offer is already
// accepted and an admin can convert manually. Returns the order on success.
func (h *Handler) createOrderFromAcceptedOffer(c *gin.Context, inquiry *modelsProduct.Inquiry, offer *modelsOrder.NegotiationOffer) *modelsOrder.Order {
	if h.services.OrderIntake == nil {
		return nil
	}
	opts := orderintake.Options{}
	if h.cfg != nil {
		opts.EnableMultiWarehouse = h.cfg.Security.EnableMultiWarehouse
	}
	res, err := h.services.OrderIntake.CreateOrderFromAcceptedOffer(c.Request.Context(), inquiry, offer, opts)
	if err != nil {
		log.Printf("negotiation accept: order creation from offer %s failed: %v", offer.ID, err)
		return nil
	}
	if res == nil || res.Order == nil {
		return nil
	}
	if mErr := h.services.OrderIntake.MarkInquiryWon(c.Request.Context(), inquiry.ID); mErr != nil {
		log.Printf("negotiation accept: failed to mark inquiry %s won: %v", inquiry.ID, mErr)
	}
	return res.Order
}

// CustomerRejectNegotiationOffer rejects an admin's pending offer.
func (h *Handler) CustomerRejectNegotiationOffer(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	offerID := c.Param("offerId")
	offer, err := h.services.Negotiation.RejectOffer(c.Request.Context(), offerID, userID)
	if err != nil {
		status, code := translateNegotiationError(err)
		response.ErrorResp(c, status, code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "offer": offer})
}
