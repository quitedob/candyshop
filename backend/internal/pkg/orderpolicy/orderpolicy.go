// Package orderpolicy holds checkout-time policy checks shared between the
// admin-portal and customer-portal handler packages.
//
// R2 A-12: previously these helpers lived as methods on each portal's Handler
// type, which produced three copies (validateLineMinQuantity,
// checkCompanyCreditLimit, emitLifecycleEvent) that drifted independently.
// Each function now takes minimal interfaces so the same implementation backs
// both portals. The portal handlers retain thin method wrappers that adapt
// their service bag to these interfaces — this avoids touching every call
// site while collapsing the actual business logic.
package orderpolicy

import (
	"context"
	"net/http"

	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// PriceProvider exposes contract-price-list MOQ lookups.
type PriceProvider interface {
	MinQuantityForPriceList(ctx context.Context, productID, priceListID string) (int, error)
}

// UserCompanyProvider lets us look up a user's company without depending on
// the concrete user/company service bag (which differs between portals).
type UserCompanyProvider interface {
	GetUserCompanyID(ctx context.Context, userID string) (string, bool)
	GetCompanyCreditLimit(ctx context.Context, companyID string) (float64, bool)
}

// EventDispatcher fans out a lifecycle event to webhook subscribers and the
// in-process event bus. Both ports are optional; nil inputs are no-ops.
type EventDispatcher struct {
	Webhook  WebhookEmitter
	EventBus EventEmitter
}

type WebhookEmitter interface {
	Dispatch(ctx context.Context, eventType, aggregateKey string, payload any)
}

// EventEmitter must support the same shape as services/order.EventBus.Emit —
// the concrete return type is the bus's *Event record, but we don't import
// it here to avoid a dependency cycle. Adapter functions in the handler
// packages bridge the concrete type to this interface.
type EventEmitter interface {
	Emit(ctx context.Context, eventType string, payload any) (any, error)
}

// EmitterFunc adapts a free function (or closure that wraps a concrete bus
// method) to EventEmitter so callers don't need to declare a struct type just
// to bridge a single method signature.
type EmitterFunc func(ctx context.Context, eventType string, payload any) (any, error)

// Emit implements EventEmitter.
func (f EmitterFunc) Emit(ctx context.Context, eventType string, payload any) (any, error) {
	return f(ctx, eventType, payload)
}

// ValidateLineMinQuantity enforces the contract price-list MOQ for a single
// order line. On violation it writes the canonical 422 response and returns
// false so the caller can early-exit the handler. When no contract price list
// is in play this is a no-op returning true.
func ValidateLineMinQuantity(c *gin.Context, price PriceProvider, productID string, quantity int, contractPriceListID *string) bool {
	if contractPriceListID == nil || price == nil {
		return true
	}
	minQty, err := price.MinQuantityForPriceList(c.Request.Context(), productID, *contractPriceListID)
	if err != nil || minQty <= 1 {
		return true
	}
	if quantity < minQty {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "min_quantity_not_met")
		return false
	}
	return true
}

// CheckCompanyCreditLimit blocks orders that would push a buyer's company
// over its standing credit limit. Companies without an explicit limit are
// treated as unlimited so unbounded customers aren't blocked by accident.
func CheckCompanyCreditLimit(c *gin.Context, provider UserCompanyProvider, userID string, totalAmount float64) bool {
	if provider == nil {
		return true
	}
	companyID, ok := provider.GetUserCompanyID(c.Request.Context(), userID)
	if !ok || companyID == "" {
		return true
	}
	limit, ok := provider.GetCompanyCreditLimit(c.Request.Context(), companyID)
	if !ok {
		return true
	}
	if limit > 0 && totalAmount > limit {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "credit_limit_exceeded")
		return false
	}
	return true
}

// EmitLifecycleEvent forwards an order/return/etc lifecycle event to webhook
// subscribers and the in-process event bus. Either dispatcher is optional;
// missing pieces are quietly skipped so feature toggles don't crash callers.
func (d EventDispatcher) Emit(ctx context.Context, eventType, aggregateKey string, payload any) {
	if d.Webhook != nil {
		d.Webhook.Dispatch(ctx, eventType, aggregateKey, payload)
	}
	if d.EventBus != nil {
		_, _ = d.EventBus.Emit(ctx, eventType, payload)
	}
}
