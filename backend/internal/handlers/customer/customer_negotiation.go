package customer

import (
	"net/http"

	orderSvc "candypro/api/internal/services/order"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

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

// CustomerAcceptNegotiationOffer accepts an admin's pending offer.
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
	offerID := c.Param("offerId")
	offer, err := h.services.Negotiation.AcceptOffer(c.Request.Context(), offerID, userID)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "offer": offer})
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
		response.ErrorResp(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "offer": offer})
}
