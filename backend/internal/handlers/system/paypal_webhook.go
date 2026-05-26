package system

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	orderSvc "candypro/api/internal/services/order"

	"github.com/gin-gonic/gin"
)

// HandlePayPalWebhook receives PayPal webhook events and updates payment records.
func (h *Handler) HandlePayPalWebhook(c *gin.Context) {
	if h.services == nil || h.services.Payment == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service_unavailable"})
		return
	}
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot_read_body"})
		return
	}
	var event struct {
		EventType string `json:"event_type"`
		Resource  struct {
			ID          string `json:"id"`
			CustomID    string `json:"custom_id"`
			ReferenceID string `json:"reference_id"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	log.Printf("paypal webhook: event type=%s", event.EventType)

	paymentID := event.Resource.ReferenceID
	if paymentID == "" {
		paymentID, _ = orderSvc.ParsePayPalWebhookResource(payload)
	}
	if paymentID == "" {
		c.JSON(http.StatusOK, gin.H{"status": "no_payment_id"})
		return
	}

	ctx := c.Request.Context()
	switch event.EventType {
	case "CHECKOUT.ORDER.APPROVED", "PAYMENT.CAPTURE.COMPLETED":
		pay, err := h.services.Payment.GetPayment(ctx, paymentID)
		if err != nil || pay == nil {
			log.Printf("paypal webhook: payment %s not found: %v", paymentID, err)
			break
		}
		if h.services.GatewayPayment != nil && pay.GatewayTransactionID != nil {
			if err := h.services.GatewayPayment.CaptureAndConfirm(ctx, pay, "paypal_webhook"); err != nil {
				log.Printf("paypal webhook: capture/confirm payment %s: %v", paymentID, err)
			}
		} else if err := h.services.Payment.ConfirmPayment(ctx, paymentID, "paypal_webhook"); err != nil {
			log.Printf("paypal webhook: confirm payment %s: %v", paymentID, err)
		}
	case "PAYMENT.CAPTURE.DENIED", "CHECKOUT.ORDER.VOIDED":
		if err := h.services.Payment.FailPayment(ctx, paymentID); err != nil {
			log.Printf("paypal webhook: fail payment %s: %v", paymentID, err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}
