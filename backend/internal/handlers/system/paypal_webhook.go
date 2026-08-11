package system

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	orderSvc "candypro/api/internal/services/order"

	"github.com/gin-gonic/gin"
)

// processedPayPalEvents remembers recently seen PayPal webhook event IDs so that
// replayed deliveries of the same event are skipped (G16). PayPal retries
// deliveries and can double-post the same event; entries expire after 24h to
// bound memory. It is a first-line dedupe only — the authoritative cross-restart
// idempotency guard is the payment status state machine, which rejects
// confirm/fail transitions on already-processed records.
var (
	processedPayPalEventsMu sync.Mutex
	processedPayPalEvents   = map[string]time.Time{}
)

const processedPayPalEventTTL = 24 * time.Hour

// isPayPalEventProcessed reports whether a PayPal event ID was seen recently.
// Empty IDs are never treated as processed so malformed payloads are not
// accidentally swallowed.
func isPayPalEventProcessed(eventID string) bool {
	if eventID == "" {
		return false
	}
	processedPayPalEventsMu.Lock()
	defer processedPayPalEventsMu.Unlock()
	seenAt, ok := processedPayPalEvents[eventID]
	if !ok {
		return false
	}
	if time.Since(seenAt) > processedPayPalEventTTL {
		delete(processedPayPalEvents, eventID)
		return false
	}
	return true
}

// markPayPalEventProcessed records a PayPal event ID as processed, opportunistically
// purging expired entries to keep the map bounded without a background goroutine.
func markPayPalEventProcessed(eventID string) {
	if eventID == "" {
		return
	}
	processedPayPalEventsMu.Lock()
	defer processedPayPalEventsMu.Unlock()
	now := time.Now()
	for id, seenAt := range processedPayPalEvents {
		if now.Sub(seenAt) > processedPayPalEventTTL {
			delete(processedPayPalEvents, id)
		}
	}
	processedPayPalEvents[eventID] = now
}

// HandlePayPalWebhook receives PayPal webhook events and updates payment records.
func (h *Handler) HandlePayPalWebhook(c *gin.Context) {
	if h.services == nil || h.services.Payment == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service_unavailable"})
		return
	}
	if h.paypalAdapter == nil || !h.paypalAdapter.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "paypal_not_configured"})
		return
	}
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot_read_body"})
		return
	}

	// Verify the PayPal transmission signature BEFORE any payment mutation.
	// Previously this endpoint trusted any caller that knew a reference_id and
	// could falsely confirm/fail payments. Verification fails closed: missing
	// headers, an unknown webhook id, a stale transmission, or a bad signature
	// all reject the event.
	if _, err := h.paypalAdapter.VerifyWebhook(
		payload,
		c.GetHeader("PayPal-Transmission-Id"),
		c.GetHeader("PayPal-Transmission-Time"),
		c.GetHeader("PayPal-Transmission-Sig"),
		c.GetHeader("PayPal-Cert-Url"),
		c.GetHeader("PayPal-Auth-Algo"),
	); err != nil {
		log.Printf("paypal webhook: signature verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_signature"})
		return
	}

	var event struct {
		ID        string `json:"id"`
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
	log.Printf("paypal webhook: event id=%s type=%s", event.ID, event.EventType)

	// Skip replays of an event already processed (G16). PayPal retries
	// deliveries; without this a redelivered PAYMENT.CAPTURE.COMPLETED would
	// re-attempt the confirm (state machine would reject it, but noisily).
	if isPayPalEventProcessed(event.ID) {
		log.Printf("paypal webhook: skipping replayed event id=%s", event.ID)
		c.JSON(http.StatusOK, gin.H{"status": "duplicate"})
		return
	}
	markPayPalEventProcessed(event.ID)

	paymentID := event.Resource.ReferenceID
	if paymentID == "" {
		paymentID, _ = orderSvc.ParsePayPalWebhookResource(payload)
	}
	if paymentID == "" {
		c.JSON(http.StatusOK, gin.H{"status": "no_payment_id"})
		return
	}

	ctx := c.Request.Context()
	h.handlePayPalPaymentEvent(ctx, event.EventType, paymentID)
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// handlePayPalPaymentEvent applies a verified PayPal event to a payment record.
// It is split out from the HTTP handler so the confirm-vs-capture logic can be
// unit tested without the webhook signature machinery (G16).
//
// The key distinction fixed here: CHECKOUT.ORDER.APPROVED only means the buyer
// approved the order — funds are NOT captured yet — so it must never mark a
// payment confirmed (that would ship goods before money moves). The payment is
// only confirmed through a gateway capture that actually succeeds. Conversely
// PAYMENT.CAPTURE.COMPLETED is emitted only after funds are captured at PayPal,
// so it may confirm the local record directly.
func (h *Handler) handlePayPalPaymentEvent(ctx context.Context, eventType, paymentID string) {
	if h.services == nil || h.services.Payment == nil {
		return
	}
	switch eventType {
	case "PAYMENT.CAPTURE.COMPLETED":
		// Funds were actually captured at PayPal. Confirm directly; the payment
		// state machine (pending→confirmed) makes replays idempotent.
		if err := h.services.Payment.ConfirmPayment(ctx, paymentID, "paypal_webhook"); err != nil {
			log.Printf("paypal webhook: confirm payment %s: %v", paymentID, err)
		}
	case "CHECKOUT.ORDER.APPROVED":
		// Buyer approved the order but the money has NOT moved yet. Never call
		// ConfirmPayment here. Trigger a server-side capture; CaptureAndConfirm
		// only confirms the payment if the gateway capture succeeds (and, with
		// the PayPal adapter, the captured amount matches). Without a gateway
		// path there is nothing to capture — leave the payment pending.
		pay, err := h.services.Payment.GetPayment(ctx, paymentID)
		if err != nil || pay == nil {
			log.Printf("paypal webhook: payment %s not found: %v", paymentID, err)
			return
		}
		if h.services.GatewayPayment == nil || pay.GatewayTransactionID == nil {
			return
		}
		if err := h.services.GatewayPayment.CaptureAndConfirm(ctx, pay, "paypal_webhook"); err != nil {
			log.Printf("paypal webhook: capture/confirm payment %s: %v", paymentID, err)
		}
	case "PAYMENT.CAPTURE.DENIED", "CHECKOUT.ORDER.VOIDED":
		if err := h.services.Payment.FailPayment(ctx, paymentID); err != nil {
			log.Printf("paypal webhook: fail payment %s: %v", paymentID, err)
		}
	}
}
