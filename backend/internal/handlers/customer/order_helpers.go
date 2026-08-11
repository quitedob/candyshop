package customer

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/orderpolicy"
	"candypro/api/internal/pkg/orderwarehouse"
	"candypro/api/internal/pkg/response"
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// channelWarehouseResolver 适配 ChannelService 为 orderwarehouse 解析器。
type channelWarehouseResolver struct {
	h *Handler
}

func (r channelWarehouseResolver) ResolveWebstoreDefaultWarehouse(ctx context.Context) (string, error) {
	if r.h.services == nil || r.h.services.Channel == nil {
		return "", nil
	}
	ch, err := r.h.services.Channel.ResolveWebstoreChannel(ctx)
	if err != nil {
		return "", err
	}
	wh := r.h.services.Channel.DefaultWarehouseID(ch)
	if wh == "" {
		return "", nil
	}
	return wh, nil
}

// assignOrderWarehouseID sets order.WarehouseID from default warehouse (single-warehouse mode).
func (h *Handler) assignOrderWarehouseID(c *gin.Context, orderWarehouseID **string) {
	if h.services == nil || h.services.Product == nil {
		return
	}
	enableMulti := h.cfg != nil && h.cfg.Security.EnableMultiWarehouse
	var channel orderwarehouse.ChannelWarehouseResolver
	if enableMulti && h.services.Channel != nil {
		channel = channelWarehouseResolver{h: h}
	}
	orderwarehouse.AssignDefault(
		c.Request.Context(),
		enableMulti,
		h.services.Product,
		channel,
		orderWarehouseID,
	)
}

// emitLifecycleEvent fires webhook subscribers and event-bus hooks.
//
// R2 A-12: delegates to orderpolicy so admin and customer portals share one
// implementation; previously this body lived in two packages.
func (h *Handler) emitLifecycleEvent(c *gin.Context, eventType, aggregateKey string, payload any) {
	if h.services == nil {
		return
	}
	d := orderpolicy.EventDispatcher{}
	if h.services.Webhook != nil {
		d.Webhook = h.services.Webhook
	}
	if h.services.EventBus != nil {
		// Bridge the concrete *Event return to the interface's any via a small
		// closure — avoids importing the events model into orderpolicy.
		bus := h.services.EventBus
		d.EventBus = orderpolicy.EmitterFunc(func(ctx context.Context, eventType string, payload any) (any, error) {
			return bus.Emit(ctx, eventType, payload)
		})
	}
	d.Emit(c.Request.Context(), eventType, aggregateKey, payload)
}

// applyChannelUnitPrice applies webstore channel price multiplier.
func (h *Handler) applyChannelUnitPrice(c *gin.Context, price float64) float64 {
	if h.services == nil || h.services.Channel == nil {
		return price
	}
	ch, err := h.services.Channel.ResolveWebstoreChannel(c.Request.Context())
	if err != nil {
		return price
	}
	return h.services.Channel.ApplyPriceMultiplier(price, ch)
}

// notifyOrderApprovers sends in-app notifications and emails to org approvers when an order needs approval.
func (h *Handler) notifyOrderApprovers(c *gin.Context, order *modelsOrder.Order) {
	if h.services == nil || h.services.Approval == nil || h.services.Notification == nil || order == nil {
		return
	}
	members, err := h.services.Approval.GetUserOrganizations(c.Request.Context(), order.UserID)
	if err != nil {
		return
	}
	vars := map[string]string{"orderNumber": order.OrderNumber}
	title := i18n.TWithVars(c, "notifications.order_approval_title", vars)
	message := i18n.TWithVars(c, "notifications.order_approval_message", vars)

	seen := make(map[string]struct{})
	for _, m := range members {
		approvers, err := h.services.Approval.GetOrgApprovers(c.Request.Context(), m.OrganizationID)
		if err != nil {
			continue
		}
		for _, a := range approvers {
			if _, ok := seen[a.UserID]; ok {
				continue
			}
			seen[a.UserID] = struct{}{}
			_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
				UserID:    a.UserID,
				Type:      "order_approval",
				Reference: order.ID,
				Title:     title,
				Message:   message,
			})
			if h.services.User != nil && h.services.Order != nil {
				approver, uerr := h.services.User.GetByID(c.Request.Context(), a.UserID)
				if uerr == nil && approver.Email != "" {
					name := approver.FirstName
					if approver.LastName != "" {
						name += " " + approver.LastName
					}
					h.services.Order.SendOrderStatusEmail(order, approver.Email, name, "pending_approval")
				}
			}
		}
	}
}

// notifyOrderApprovalResult notifies the purchaser after approve/reject.
func (h *Handler) notifyOrderApprovalResult(c *gin.Context, order *modelsOrder.Order, approved bool) {
	if h.services == nil || order == nil || h.services.Notification == nil {
		return
	}
	vars := map[string]string{"orderNumber": order.OrderNumber}
	var titleKey, messageKey, emailStatus string
	if approved {
		titleKey = "notifications.order_approved_title"
		messageKey = "notifications.order_approved_message"
		emailStatus = "approval_approved"
	} else {
		titleKey = "notifications.order_rejected_title"
		messageKey = "notifications.order_rejected_message"
		emailStatus = "approval_rejected"
	}
	title := i18n.TWithVars(c, titleKey, vars)
	message := i18n.TWithVars(c, messageKey, vars)
	_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
		UserID:    order.UserID,
		Type:      "order",
		Reference: order.ID,
		Title:     title,
		Message:   message,
	})
	if h.services.User != nil && h.services.Order != nil {
		purchaser, err := h.services.User.GetByID(c.Request.Context(), order.UserID)
		if err == nil && purchaser.Email != "" {
			name := purchaser.FirstName
			if purchaser.LastName != "" {
				name += " " + purchaser.LastName
			}
			h.services.Order.SendOrderStatusEmail(order, purchaser.Email, name, emailStatus)
		}
	}
}

// validateLineMinQuantity checks contract price list minimum order quantities.
//
// R2 A-12: delegates to orderpolicy.ValidateLineMinQuantity so the admin and
// customer portals share a single implementation.
func (h *Handler) validateLineMinQuantity(c *gin.Context, productID string, quantity int, contractPriceListID *string) bool {
	if h.services == nil {
		return true
	}
	return orderpolicy.ValidateLineMinQuantity(c, h.services.Price, productID, quantity, contractPriceListID)
}

// checkCompanyCreditLimit enforces the buyer company's credit limit against
// the order total. Delegates to the shared orderpolicy package (R2 A-12).
func (h *Handler) checkCompanyCreditLimit(c *gin.Context, userID string, totalAmount float64) bool {
	if h.services == nil {
		return true
	}
	return orderpolicy.CheckCompanyCreditLimit(c, customerCompanyProvider{h: h}, userID, totalAmount)
}

// checkCompanyCreditLimitCumulative enforces the buyer company's credit limit
// against cumulative outstanding exposure at confirm: the sum of every buyer
// user's open (non-terminal) order totals for the company, excluding the order
// being confirmed, plus this order's confirmed total, must not exceed the limit
// (G20). The per-order check at draft creation cannot catch a buyer stacking
// several under-limit orders, and bulk/requisition/reorder drafts bypass that
// check entirely, so this runs at confirm before stock is committed.
//
// The exposure sum is scoped to the COMPANY, not the user (G20 r3): the credit
// limit lives on the company, so a per-user sum let two buyer users of one
// company each confirm up to the full limit (e.g. limit 1000, users A and B
// each confirm 800 → 1600 outstanding with neither per-user confirm blocked).
// SumOpenOrderTotalsByCompany aggregates across all company users in a single
// query. Companies without an explicit limit are treated as unlimited,
// mirroring orderpolicy.CheckCompanyCreditLimit. When the company/limit cannot
// be resolved the check fails open (nothing to enforce); when the limit is
// known but the open-exposure sum fails, it fails CLOSED so a transient DB
// failure cannot silently lift the credit guard.
func (h *Handler) checkCompanyCreditLimitCumulative(c *gin.Context, userID, excludeOrderID string, totalAmount float64) bool {
	if h.services == nil || h.services.Order == nil {
		return true
	}
	provider := customerCompanyProvider{h: h}
	companyID, ok := provider.GetUserCompanyID(c.Request.Context(), userID)
	if !ok || companyID == "" {
		return true // no company → no limit applies
	}
	limit, ok := provider.GetCompanyCreditLimit(c.Request.Context(), companyID)
	if !ok || limit <= 0 {
		return true // unlimited
	}
	open, err := h.services.Order.SumOpenOrderTotalsByCompany(c.Request.Context(), companyID, excludeOrderID)
	if err != nil {
		// Fail CLOSED on the open-exposure sum. The resolution failures above are
		// different: no company / no limit means there is nothing to enforce, so
		// failing open mirrors the per-order check. But a failed sum means the
		// company's limit IS known and we simply cannot verify the cumulative
		// exposure fits — confirming would silently disable the credit guard on a
		// transient DB failure and let a buyer exceed their limit (G20 refutation).
		// Over-credit is the worse failure mode than a blocked confirm.
		slog.Error("checkCompanyCreditLimitCumulative: company open-order sum failed",
			"userID", userID, "companyID", companyID, "orderID", excludeOrderID, "error", err)
		response.ErrorResp(c, http.StatusInternalServerError, "order_confirm_failed")
		return false
	}
	if open+totalAmount > limit {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "credit_limit_exceeded")
		return false
	}
	return true
}

// customerCompanyProvider adapts customer-portal services to the
// orderpolicy.UserCompanyProvider interface.
type customerCompanyProvider struct {
	h *Handler
}

func (p customerCompanyProvider) GetUserCompanyID(ctx context.Context, userID string) (string, bool) {
	if p.h.services == nil || p.h.services.User == nil {
		return "", false
	}
	usr, err := p.h.services.User.GetByID(ctx, userID)
	if err != nil || usr.CompanyID == nil || *usr.CompanyID == "" {
		return "", false
	}
	return *usr.CompanyID, true
}

func (p customerCompanyProvider) GetCompanyCreditLimit(ctx context.Context, companyID string) (float64, bool) {
	if p.h.services == nil || p.h.services.Company == nil {
		return 0, false
	}
	company, err := p.h.services.Company.GetCompany(ctx, companyID)
	if err != nil || company == nil {
		return 0, false
	}
	return company.CreditLimit, true
}
