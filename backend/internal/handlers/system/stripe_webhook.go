package system

import (
	"io"
	"log"
	"net/http"

	stripeAdapter "candypro/api/internal/pkg/payment/stripe"

	"github.com/gin-gonic/gin"
)

// HandleStripeWebhook receives Stripe webhook events and updates payment records accordingly.
func (h *Handler) HandleStripeWebhook(c *gin.Context) {
	if h.stripeAdapter == nil || !h.stripeAdapter.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stripe_not_configured"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot_read_body"})
		return
	}

	if _, err := h.stripeAdapter.ValidateWebhookPayload(payload, signature); err != nil {
		log.Printf("stripe webhook: signature verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_signature"})
		return
	}

	eventType, intent, parseErr := stripeAdapter.ParseWebhookEvent(payload)
	if parseErr != nil {
		log.Printf("stripe webhook: failed to parse event: %v", parseErr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	log.Printf("stripe webhook: received event type=%s", eventType)

	if intent != nil {
		paymentID, _ := intent.Metadata["payment_id"].(string)
		if paymentID == "" {
			c.JSON(http.StatusOK, gin.H{"status": "no_payment_id"})
			return
		}

		switch eventType {
		case "payment_intent.succeeded":
			h.handleStripePaymentSucceeded(c, paymentID, intent)
		case "payment_intent.payment_failed":
			h.handleStripePaymentFailed(c, paymentID)
		case "payment_intent.canceled":
			h.handleStripePaymentCanceled(c, paymentID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func (h *Handler) handleStripePaymentSucceeded(c *gin.Context, paymentID string, intent *stripeAdapter.PaymentIntent) {
	if h.services == nil || h.services.Payment == nil {
		return
	}
	ctx := c.Request.Context()

	switch intent.Status {
	case "requires_capture":
		// Authorize-only mode — payment authorized but not yet captured.
		if err := h.services.Payment.AuthorizePayment(ctx, paymentID); err != nil {
			log.Printf("stripe webhook: failed to authorize payment %s: %v", paymentID, err)
		} else {
			log.Printf("stripe webhook: payment %s authorized", paymentID)
		}
	case "succeeded":
		// One-step payment — directly confirm.
		if err := h.services.Payment.ConfirmPayment(ctx, paymentID, "stripe_webhook"); err != nil {
			log.Printf("stripe webhook: failed to confirm payment %s: %v", paymentID, err)
		} else {
			log.Printf("stripe webhook: payment %s confirmed", paymentID)
		}
	}
}

func (h *Handler) handleStripePaymentFailed(c *gin.Context, paymentID string) {
	if h.services == nil || h.services.Payment == nil {
		return
	}
	if err := h.services.Payment.FailPayment(c.Request.Context(), paymentID); err != nil {
		log.Printf("stripe webhook: failed to mark payment %s failed: %v", paymentID, err)
		return
	}
	h.compensateOrderForFailedPayment(c, paymentID)
}

func (h *Handler) handleStripePaymentCanceled(c *gin.Context, paymentID string) {
	if h.services == nil || h.services.Payment == nil {
		return
	}
	if err := h.services.Payment.FailPayment(c.Request.Context(), paymentID); err != nil {
		log.Printf("stripe webhook: failed to cancel payment %s: %v", paymentID, err)
		return
	}
	h.compensateOrderForFailedPayment(c, paymentID)
}

// compensateOrderForFailedPayment releases reserved stock and cancels the order
// if the failed payment was the only live payment AND the order is still in a
// state where compensation is meaningful (pending / pending_confirmation).
//
// H-21: previously a failed gateway payment left the order with reserved stock
// indefinitely. Customers who never retried saw stock leak until the draft
// expiry sweep eventually fired. With this hook the order is auto-cancelled
// and stock returns to circulation immediately.
func (h *Handler) compensateOrderForFailedPayment(c *gin.Context, paymentID string) {
	if h.services == nil || h.services.Payment == nil || h.services.Order == nil {
		return
	}
	ctx := c.Request.Context()
	pay, err := h.services.Payment.GetPayment(ctx, paymentID)
	if err != nil || pay == nil {
		return
	}
	hasLive, err := h.services.Payment.HasLivePaymentForOrder(ctx, pay.OrderID)
	if err != nil {
		log.Printf("stripe webhook: HasLivePaymentForOrder %s: %v", pay.OrderID, err)
		return
	}
	if hasLive {
		// Another payment attempt is still pending/authorized/confirmed — no
		// compensation needed; the order remains live.
		return
	}
	order, err := h.services.Order.GetOrder(ctx, pay.OrderID)
	if err != nil || order == nil {
		return
	}
	// Only auto-cancel orders that haven't moved past the pre-execution stage.
	switch order.Status {
	case "pending", "pending_confirmation":
	default:
		return
	}
	if order.StockReserved {
		if rerr := h.services.Order.ReleaseOrderStock(ctx, order); rerr != nil {
			log.Printf("stripe webhook: ReleaseOrderStock for order %s: %v", order.ID, rerr)
			return
		}
	}
	order.Status = "cancelled"
	if uerr := h.services.Order.UpdateOrder(ctx, order); uerr != nil {
		log.Printf("stripe webhook: cancel order %s: %v", order.ID, uerr)
		return
	}
	log.Printf("stripe webhook: order %s auto-cancelled after payment %s failure (H-21 compensation)", order.ID, paymentID)
}
