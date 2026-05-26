package customer

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/orderpolicy"
	"candypro/api/internal/pkg/orderwarehouse"
	"context"

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
