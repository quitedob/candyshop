package admin

import (
	"errors"
	"log"
	"net/http"

	modelsOrder "candypro/api/internal/models/order"
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

// AdminAcceptNegotiationOffer accepts a buyer's pending offer and converts the
// agreed terms into a draft order (P0.1 / G-ORD-1).
func (h *Handler) AdminAcceptNegotiationOffer(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	inquiryID := c.Param("id")
	offerID := c.Param("offerId")
	// H-3: 传入路径 inquiryID，服务层限定 offer 必须属于该询盘，防止把其他
	// 询盘的 offer 按当前询盘转成订单（询盘与报价条款错配）。
	offer, err := h.services.Negotiation.AcceptOffer(c.Request.Context(), offerID, inquiryID)
	if err != nil {
		status, code := translateNegotiationError(err)
		response.ErrorResp(c, status, code)
		return
	}

	order := h.createOrderFromAcceptedOffer(c, inquiryID, offer)
	resp := gin.H{"success": true, "offer": offer}
	if order != nil {
		resp["order"] = order
	}
	c.JSON(http.StatusOK, resp)
}

// createOrderFromAcceptedOffer builds a draft order from the accepted offer.
// Failures are logged but never fail the accept — the offer is already accepted
// and the order can be created manually via inquiry-convert as a fallback.
func (h *Handler) createOrderFromAcceptedOffer(c *gin.Context, inquiryID string, offer *modelsOrder.NegotiationOffer) *modelsOrder.Order {
	if h.services.OrderIntake == nil || h.services.Inquiry == nil {
		return nil
	}
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		log.Printf("negotiation accept: inquiry %s lookup failed: %v", inquiryID, err)
		return nil
	}
	opts := orderintake.Options{}
	if h.cfg != nil {
		opts.EnableMultiWarehouse = h.cfg.Security.EnableMultiWarehouse
	}
	res, ierr := h.services.OrderIntake.CreateOrderFromAcceptedOffer(c.Request.Context(), inquiry, offer, opts)
	if ierr != nil {
		log.Printf("negotiation accept: order creation from offer %s failed: %v", offer.ID, ierr)
		return nil
	}
	if res == nil || res.Order == nil {
		return nil
	}
	if mErr := h.services.OrderIntake.MarkInquiryWon(c.Request.Context(), inquiryID); mErr != nil {
		log.Printf("negotiation accept: failed to mark inquiry %s won: %v", inquiryID, mErr)
	}
	return res.Order
}

// AdminRejectNegotiationOffer rejects a buyer's pending offer.
func (h *Handler) AdminRejectNegotiationOffer(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID := c.GetString("userID")
	inquiryID := c.Param("id")
	offerID := c.Param("offerId")
	offer, err := h.services.Negotiation.RejectOffer(c.Request.Context(), offerID, inquiryID)
	if err != nil {
		status, code := translateNegotiationError(err)
		response.ErrorResp(c, status, code)
		return
	}
	// L1: record which operator rejected the offer (NegotiationOffer has no
	// operator column, so the audit log is the only attribution).
	h.logActivityAudit(c, "negotiation_offer_reject", "negotiation_offer", offerID, "", "rejected", userID)
	c.JSON(http.StatusOK, gin.H{"success": true, "offer": offer})
}
