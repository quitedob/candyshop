package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/orderpolicy"
	"candypro/api/internal/pkg/orderwarehouse"
	"candypro/api/internal/pkg/response"
	orderSvc "candypro/api/internal/services/order"
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// validateOrderItemsCompliance 校验订单行项目在目的国的合规性。
//
// 注意：当 country 为空时，目的国合规规则（清真要求/禁用成分等）无法判定。
// 调用方应在 country 为空但订单包含可能受目的国限制的商品时拒绝订单，而不是
// 把空 country 视作"无目的地→自动通过"。这里返回 true 仅表示"无违规可判定"，
// 不代表订单合规。
func (h *Handler) validateOrderItemsCompliance(c *gin.Context, country string, items modelsOrder.OrderItemArray) bool {
	country = strings.TrimSpace(country)
	if country == "" || len(items) == 0 || h.services == nil || h.services.Product == nil {
		return true
	}
	// R2 E-3: batch product lookups. Previously this was an N+1 loop calling
	// GetProductByID per line; now a single GetProductsByIDs round-trip serves
	// the whole order regardless of size.
	seen := make(map[string]struct{}, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		pid := strings.TrimSpace(item.ProductID)
		if pid == "" {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		ids = append(ids, pid)
	}
	if len(ids) == 0 {
		return true
	}
	products, err := h.services.Product.GetProductsByIDs(c.Request.Context(), ids)
	if err != nil || len(products) == 0 {
		return true
	}
	compliance := h.services.Product.ValidateComplianceWithMarketProfiles(c.Request.Context(), country, products)
	if len(compliance.Violations) > 0 {
		response.ErrorRespDetail(c, http.StatusUnprocessableEntity, "compliance_violation", gin.H{
			"country":    compliance.Country,
			"violations": compliance.Violations,
			"warnings":   compliance.Warnings,
		})
		return false
	}
	return true
}

// buildOrderInventoryWarnings 订单行库存警告（低库存/售完风险）
func (h *Handler) buildOrderInventoryWarnings(c *gin.Context, order *modelsOrder.Order) []string {
	if h.services == nil || h.services.Product == nil || order == nil || len(order.Items) == 0 {
		return nil
	}
	// R2 E-2: collect unique product IDs first, then fetch all products with
	// one batch query. Previously this looped GetProductByID per item.
	seen := make(map[string]struct{}, len(order.Items))
	ids := make([]string, 0, len(order.Items))
	for _, item := range order.Items {
		pid := strings.TrimSpace(item.ProductID)
		if pid == "" {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		ids = append(ids, pid)
	}
	if len(ids) == 0 {
		return nil
	}
	productByID := make(map[string]modelsProduct.Product, len(ids))
	if products, err := h.services.Product.GetProductsByIDs(c.Request.Context(), ids); err == nil {
		for i := range products {
			productByID[products[i].ID] = products[i]
		}
	}
	// R2 A-11: previously the error was silently dropped, producing a false
	// "no inventory issues" warning set if the underlying query failed.
	sellable, sellErr := h.services.Product.EffectiveSellableByProducts(c.Request.Context(), ids, modelsProduct.ChannelWebstore)
	if sellErr != nil {
		slog.Warn("buildOrderInventoryWarnings: sellable lookup failed",
			"orderID", order.ID, "error", sellErr)
		// Fall through with the (likely empty) map so we don't block the order
		// view. The downstream loop treats missing keys as zero availability,
		// which yields a correct "sold out" warning rather than a false clear.
	}
	var warnings []orderSvc.InventoryWarning
	for _, item := range order.Items {
		pid := strings.TrimSpace(item.ProductID)
		if pid == "" {
			continue
		}
		avail := sellable[pid]
		remaining := avail - item.Quantity
		product := productByID[pid]
		if remaining <= 0 {
			warnings = append(warnings, orderSvc.InventoryWarning{
				Code: orderSvc.InventoryWarningSoldOutAdmin, ProductID: pid,
				Available: avail, Ordered: item.Quantity,
			})
		} else if product.MOQ > 0 && remaining < product.MOQ {
			warnings = append(warnings, orderSvc.InventoryWarning{
				Code: orderSvc.InventoryWarningBelowMOQAfterOrder, ProductID: pid,
				Remaining: remaining, MOQ: product.MOQ,
			})
		} else if remaining < 1000 {
			warnings = append(warnings, orderSvc.InventoryWarning{
				Code: orderSvc.InventoryWarningLowStock, ProductID: pid, Remaining: remaining,
			})
		}
	}
	return formatAdminInventoryWarnings(c, warnings, productByID)
}

// assignOrderWarehouseID sets order.WarehouseID from default warehouse when available.
func (h *Handler) assignOrderWarehouseID(c *gin.Context, orderWarehouseID **string) {
	if h.services == nil || h.services.Product == nil {
		return
	}
	enableMulti := h.cfg != nil && h.cfg.Security.EnableMultiWarehouse
	orderwarehouse.AssignDefault(
		c.Request.Context(),
		enableMulti,
		h.services.Product,
		nil,
		orderWarehouseID,
	)
}

// emitLifecycleEvent fires webhook subscribers and event-bus hooks for an arbitrary payload.
//
// R2 A-12: delegates to orderpolicy so admin and customer portals share one
// implementation; previously each portal had its own copy.
func (h *Handler) emitLifecycleEvent(c *gin.Context, eventType, aggregateKey string, payload any) {
	if h.services == nil {
		return
	}
	d := orderpolicy.EventDispatcher{}
	if h.services.Webhook != nil {
		d.Webhook = h.services.Webhook
	}
	if h.services.EventBus != nil {
		bus := h.services.EventBus
		d.EventBus = orderpolicy.EmitterFunc(func(ctx context.Context, eventType string, payload any) (any, error) {
			return bus.Emit(ctx, eventType, payload)
		})
	}
	d.Emit(c.Request.Context(), eventType, aggregateKey, payload)
}

// syncFulfillmentTrackingToOrderAndTrade 履约发货后同步运单到订单及关联贸易 BOL 记录。
func (h *Handler) syncFulfillmentTrackingToOrderAndTrade(c *gin.Context, fulfillmentID, trackingNumber, carrier string) {
	if h.services == nil || h.services.Fulfillment == nil {
		return
	}
	f, _, err := h.services.Fulfillment.FindByID(c.Request.Context(), fulfillmentID)
	if err != nil || f == nil {
		return
	}
	if h.services.Order != nil {
		order, oerr := h.services.Order.GetOrder(c.Request.Context(), f.OrderID)
		if oerr == nil && order != nil {
			order.TrackingNumber = strings.TrimSpace(trackingNumber)
			// R2 A-8: surface the failure instead of swallowing it. The handler
			// continues so the BOL sync below still runs even if the order
			// projection failed (e.g. version conflict).
			if uerr := h.services.Order.UpdateOrder(c.Request.Context(), order); uerr != nil {
				slog.Warn("syncFulfillmentTracking: order tracking update failed",
					"orderID", order.ID, "trackingNumber", trackingNumber, "error", uerr)
			}
		}
	}
	if h.services.Trade == nil || h.services.Shipment == nil {
		return
	}
	trade, terr := h.services.Trade.GetFirstTransactionByOrderID(c.Request.Context(), f.OrderID)
	if terr != nil || trade == nil {
		return
	}
	shipments, serr := h.services.Shipment.GetByTransactionID(c.Request.Context(), trade.ID)
	if serr != nil || len(shipments) == 0 {
		return
	}
	for i := range shipments {
		sh := &shipments[i]
		if sh.Status != "PENDING" && sh.Status != "DISPATCHED" && sh.Status != "IN_TRANSIT" {
			continue
		}
		if trackingNumber != "" {
			sh.BillOfLadingNo = trackingNumber
		}
		if carrier != "" {
			sh.CarrierName = carrier
		}
		if sh.Status == "PENDING" {
			sh.Status = "DISPATCHED"
		}
		// R2 A-8: log shipment write failures so silently-stuck BOL records
		// surface in observability instead of staying in PENDING with the
		// new tracking lost.
		if uerr := h.services.Shipment.UpdateShipment(c.Request.Context(), sh); uerr != nil {
			slog.Warn("syncFulfillmentTracking: shipment update failed",
				"shipmentID", sh.ID, "tradeID", trade.ID, "error", uerr)
		}
		break
	}
}

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


// resolveAdminContractPriceListID returns the contract price list bound to the
// user's company, if any. Returns nil when no contract is configured.
func (h *Handler) resolveAdminContractPriceListID(c *gin.Context, userID string) *string {
	if h.services == nil || h.services.User == nil || h.services.Company == nil {
		return nil
	}
	usr, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil || usr == nil || usr.CompanyID == nil || strings.TrimSpace(*usr.CompanyID) == "" {
		return nil
	}
	company, err := h.services.Company.GetCompany(c.Request.Context(), *usr.CompanyID)
	if err != nil || company == nil || company.PriceListID == nil || strings.TrimSpace(*company.PriceListID) == "" {
		return nil
	}
	return company.PriceListID
}

// validateAdminLineMinQuantity enforces contract-price-list MOQ for an admin-created order.
// Returns false and writes the error response when the line falls below MOQ.
//
// R2 A-12: delegates to orderpolicy so the rule lives in one place.
func (h *Handler) validateAdminLineMinQuantity(c *gin.Context, productID string, quantity int, contractPriceListID *string) bool {
	if h.services == nil {
		return true
	}
	return orderpolicy.ValidateLineMinQuantity(c, h.services.Price, productID, quantity, contractPriceListID)
}

// checkAdminCompanyCreditLimit enforces the buyer company's credit limit against the order total.
// Returns false and writes the error response when the limit is exceeded.
//
// R2 A-12: delegates to orderpolicy.CheckCompanyCreditLimit.
func (h *Handler) checkAdminCompanyCreditLimit(c *gin.Context, userID string, totalAmount float64) bool {
	if h.services == nil {
		return true
	}
	return orderpolicy.CheckCompanyCreditLimit(c, adminCompanyProvider{h: h}, userID, totalAmount)
}

// adminCompanyProvider adapts admin-portal services to the
// orderpolicy.UserCompanyProvider interface.
type adminCompanyProvider struct {
	h *Handler
}

func (p adminCompanyProvider) GetUserCompanyID(ctx context.Context, userID string) (string, bool) {
	if p.h.services == nil || p.h.services.User == nil {
		return "", false
	}
	usr, err := p.h.services.User.GetByID(ctx, userID)
	if err != nil || usr == nil || usr.CompanyID == nil || strings.TrimSpace(*usr.CompanyID) == "" {
		return "", false
	}
	return *usr.CompanyID, true
}

func (p adminCompanyProvider) GetCompanyCreditLimit(ctx context.Context, companyID string) (float64, bool) {
	if p.h.services == nil || p.h.services.Company == nil {
		return 0, false
	}
	company, err := p.h.services.Company.GetCompany(ctx, companyID)
	if err != nil || company == nil {
		return 0, false
	}
	return company.CreditLimit, true
}

// validateAdminInventory verifies that all order lines respect available sellable stock,
// MOQ, and channel caps. Returns false and writes the error response on violation.
func (h *Handler) validateAdminInventory(c *gin.Context, items modelsOrder.OrderItemArray, productByID map[string]modelsProduct.Product) bool {
	if h.services == nil || h.services.Order == nil || h.services.Product == nil || len(items) == 0 {
		return true
	}
	ids := make([]string, 0, len(productByID))
	for id := range productByID {
		ids = append(ids, id)
	}
	sellable, _ := h.services.Product.EffectiveSellableByProducts(c.Request.Context(), ids, modelsProduct.ChannelWebstore)
	inventory := h.services.Order.ValidateInventoryWithSellable(items, productByID, sellable)
	if len(inventory.Violations) > 0 {
		response.ErrorRespDetail(c, http.StatusUnprocessableEntity, "inventory_violation", gin.H{
			"violations": inventory.Violations,
			"warnings":   formatAdminInventoryWarnings(c, inventory.Warnings, productByID),
		})
		return false
	}
	return true
}

// resolveAdminOrderItems validates each requested order line, resolves the unit price
// server-side (contract price list → product base price → cross-border cost stack),
// enforces MOQ, and returns the canonical OrderItem slice plus a productByID map for
// downstream inventory/compliance checks. On failure it writes the error response and
// returns ok=false.
//
// This mirrors the customer checkout path (CustomerCreateOrder) so that admin orders
// cannot bypass server-side pricing or per-line MOQ enforcement.
func (h *Handler) resolveAdminOrderItems(
	c *gin.Context,
	requested []modelsOrder.OrderItem,
	contractPriceListID *string,
	country string,
) (items modelsOrder.OrderItemArray, productByID map[string]modelsProduct.Product, products []modelsProduct.Product, ok bool) {
	if h.services == nil || h.services.Product == nil {
		response.ServiceUnavailableResp(c)
		return nil, nil, nil, false
	}
	if len(requested) == 0 {
		response.InvalidResp(c, "invalid_request")
		return nil, nil, nil, false
	}

	items = make(modelsOrder.OrderItemArray, 0, len(requested))
	productByID = make(map[string]modelsProduct.Product, len(requested))
	products = make([]modelsProduct.Product, 0, len(requested))

	// R2 E-4: batch product lookups up front. Previously this loop called
	// GetProductByID per line, producing N round-trips on the admin checkout
	// path. We keep per-line validation (status/MOQ/price) but resolve every
	// product with a single FindByIDs call.
	uniqueIDs := make([]string, 0, len(requested))
	idSeen := make(map[string]struct{}, len(requested))
	for _, raw := range requested {
		pid := strings.TrimSpace(raw.ProductID)
		if pid == "" {
			continue
		}
		if _, ok := idSeen[pid]; ok {
			continue
		}
		idSeen[pid] = struct{}{}
		uniqueIDs = append(uniqueIDs, pid)
	}
	productLookup := make(map[string]modelsProduct.Product, len(uniqueIDs))
	if len(uniqueIDs) > 0 {
		batch, batchErr := h.services.Product.GetProductsByIDs(c.Request.Context(), uniqueIDs)
		if batchErr != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "product_lookup_failed")
			return nil, nil, nil, false
		}
		for i := range batch {
			productLookup[batch[i].ID] = batch[i]
		}
	}

	for _, raw := range requested {
		productID := strings.TrimSpace(raw.ProductID)
		if productID == "" {
			response.InvalidResp(c, "invalid_request")
			return nil, nil, nil, false
		}
		if raw.Quantity < 1 {
			response.InvalidResp(c, "quantity_min_1")
			return nil, nil, nil, false
		}
		if !h.validateAdminLineMinQuantity(c, productID, raw.Quantity, contractPriceListID) {
			return nil, nil, nil, false
		}

		product, found := productLookup[productID]
		if !found {
			response.ErrorResp(c, http.StatusNotFound, "product_not_found")
			return nil, nil, nil, false
		}
		if status := strings.ToLower(strings.TrimSpace(product.Status)); status != "" && status != "active" {
			response.ErrorResp(c, http.StatusBadRequest, "product_unavailable")
			return nil, nil, nil, false
		}

		// Resolve unit price server-side. Client-supplied unitPrice is ignored.
		var unitPrice float64
		if contractPriceListID != nil && h.services.Price != nil {
			if cp, priceErr := h.services.Price.GetPriceForProduct(c.Request.Context(), productID, *contractPriceListID, raw.Quantity); priceErr == nil && cp > 0 {
				unitPrice = cp
			}
		}
		if unitPrice <= 0 {
			unitPrice = product.BasePrice
		}
		if unitPrice <= 0 {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "no_price")
			return nil, nil, nil, false
		}
		unitPrice = h.services.Product.ResolveCheckoutUnitPrice(c.Request.Context(), &product, unitPrice, country)

		items = append(items, modelsOrder.OrderItem{
			ProductID:      productID,
			Quantity:       raw.Quantity,
			UnitPrice:      unitPrice,
			Specifications: strings.TrimSpace(raw.Specifications),
		})
		productByID[productID] = product
		products = append(products, product)
	}
	return items, productByID, products, true
}
