// Package orderintake centralises the orchestration that turns an inquiry
// (optionally with an accepted negotiation offer) into a draft order plus a
// linked trade transaction.
//
// This composition lives above the leaf services on purpose: the product
// service imports the order service, so the order service cannot import the
// product service. Order building needs product + price validation, therefore
// the orchestration has to sit in a package that depends on all of them and is
// itself imported by nobody else (the handler scopes wire it in).
//
// Both the customer and admin negotiation-accept handlers reuse this so the
// negotiate → buy path produces a real financial artifact instead of leaving
// the agreed terms as dead data (P0.1 / G-ORD-1).
package orderintake

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/orderwarehouse"
	inquirySvc "candypro/api/internal/services/inquiry"
	orderSvc "candypro/api/internal/services/order"
	productSvc "candypro/api/internal/services/product"
	tradeSvc "candypro/api/internal/services/trade"
)

// Sentinel errors. Handlers translate these into stable user-facing codes so
// internal implementation detail is not leaked (C-10).
var (
	// ErrNoItems means the inquiry referenced no usable, active products.
	ErrNoItems = errors.New("orderintake: inquiry has no usable items")
	// ErrComplianceViolation means the destination country blocks one or more
	// products. Details are carried on IntakeResult.ComplianceViolations.
	ErrComplianceViolation = errors.New("orderintake: compliance violation")
	// ErrInventoryViolation means stock is insufficient for one or more lines.
	// Details are carried on IntakeResult.InventoryViolations.
	ErrInventoryViolation = errors.New("orderintake: inventory violation")
	// ErrMissingCountry means no destination country could be resolved.
	ErrMissingCountry = errors.New("orderintake: destination country required")
)

// Service composes the leaf services needed to build an order from an inquiry.
type Service struct {
	product *productSvc.ProductService
	price   *productSvc.PriceService
	order   *orderSvc.OrderService
	trade   *tradeSvc.TradeService
	inquiry *inquirySvc.InquiryService
}

// NewService wires the intake composition. price and trade may be nil; the
// service degrades gracefully (falls back to base price, skips trade creation).
func NewService(
	product *productSvc.ProductService,
	price *productSvc.PriceService,
	order *orderSvc.OrderService,
	trade *tradeSvc.TradeService,
	inquiry *inquirySvc.InquiryService,
) *Service {
	return &Service{product: product, price: price, order: order, trade: trade, inquiry: inquiry}
}

// Options tunes how an order is constructed.
type Options struct {
	// EnableMultiWarehouse mirrors config.Security.EnableMultiWarehouse so the
	// default-warehouse resolution matches the rest of the order paths.
	EnableMultiWarehouse bool
	// ShippingAddress, when set, overrides the destination derived from the
	// inquiry's TargetCountry.
	ShippingAddress *modelsOrder.Address
}

// IntakeResult carries the outcome of an order-build attempt. When err is a
// validation sentinel (compliance/inventory), Order is nil and the relevant
// detail slice is populated so the caller can surface actionable feedback.
type IntakeResult struct {
	Order                *modelsOrder.Order
	ComplianceViolations []string
	ComplianceWarnings   []string
	InventoryViolations  []string
}

// CreateOrderFromAcceptedOffer builds a draft order (pending_confirmation) from
// an inquiry, snapshotting the negotiated terms from an accepted offer when one
// is provided. offer may be nil for a plain inquiry → order conversion.
//
// The order is created in pending_confirmation so the existing confirm flow
// (which atomically reserves stock and finalises financials) still applies. The
// negotiated total becomes the contract amount, preserving the agreed price.
func (s *Service) CreateOrderFromAcceptedOffer(
	ctx context.Context,
	inquiry *modelsProduct.Inquiry,
	offer *modelsOrder.NegotiationOffer,
	opts Options,
) (*IntakeResult, error) {
	if s == nil || s.product == nil || s.order == nil {
		return nil, fmt.Errorf("orderintake: service not initialised")
	}
	if inquiry == nil {
		return nil, fmt.Errorf("orderintake: inquiry required")
	}
	if inquiry.UserID == nil || strings.TrimSpace(*inquiry.UserID) == "" {
		return nil, fmt.Errorf("orderintake: inquiry missing user")
	}
	userID := strings.TrimSpace(*inquiry.UserID)

	// Resolve destination country: explicit address → inquiry target.
	destCountry := ""
	if opts.ShippingAddress != nil {
		destCountry = strings.TrimSpace(opts.ShippingAddress.Country)
	}
	if destCountry == "" {
		destCountry = strings.TrimSpace(inquiry.TargetCountry)
	}
	if destCountry == "" {
		return nil, ErrMissingCountry
	}

	// Per-line quantity: prefer the negotiated quantity, else parse the inquiry's
	// free-text estimate.
	quantity := parseQuantity(inquiry.EstimatedQuantity)
	if offer != nil && offer.Quantity > 0 {
		quantity = offer.Quantity
	}
	if quantity < 1 {
		quantity = 1
	}

	// Gather candidate product IDs (dedup, preserve order).
	ids := dedupeNonEmpty(inquiry.Products)
	if len(ids) == 0 {
		return nil, ErrNoItems
	}

	products, err := s.product.GetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("orderintake: product lookup failed: %w", err)
	}
	productByID := make(map[string]modelsProduct.Product, len(products))
	for i := range products {
		productByID[products[i].ID] = products[i]
	}

	// Build validated line items in the inquiry's product order.
	var items []modelsOrder.OrderItem
	selectedProducts := make([]modelsProduct.Product, 0, len(ids))
	for _, pid := range ids {
		product, ok := productByID[pid]
		if !ok {
			continue
		}
		if st := strings.ToLower(strings.TrimSpace(product.Status)); st != "" && st != "active" {
			continue
		}

		// A negotiated offer's unit price is the agreed price and applies to EVERY
		// matching line — not just single-product inquiries (H10). The previous
		// guard (len(items)==1) silently dropped the negotiated price from
		// multi-line orders, so confirm-time totals and trade documents disagreed
		// with the accepted offer. When no offer is present the catalog / contract
		// price resolved below stands.
		unitPrice := effectiveUnitPrice(s.resolveUnitPrice(ctx, product, quantity, destCountry), offer)
		if unitPrice <= 0 {
			continue
		}
		items = append(items, modelsOrder.OrderItem{
			ProductID: pid,
			Quantity:  quantity,
			UnitPrice: unitPrice,
		})
		selectedProducts = append(selectedProducts, product)
	}
	if len(items) == 0 {
		return nil, ErrNoItems
	}

	// Destination-country compliance (hard block — non-compliant goods cannot ship).
	compliance := s.product.ValidateComplianceWithMarketProfiles(ctx, destCountry, selectedProducts)
	if len(compliance.Violations) > 0 {
		return &IntakeResult{
			ComplianceViolations: compliance.Violations,
			ComplianceWarnings:   compliance.Warnings,
		}, ErrComplianceViolation
	}

	// Inventory validation against effective sellable quantity.
	inv := s.order.ValidateInventoryWithSellable(items, productByID, s.sellable(ctx, ids))
	if len(inv.Violations) > 0 {
		return &IntakeResult{InventoryViolations: inv.Violations}, ErrInventoryViolation
	}

	// Totals: negotiated total is the contract amount when present.
	var subtotal float64
	for _, item := range items {
		subtotal += item.UnitPrice * float64(item.Quantity)
	}
	total := subtotal
	if offer != nil && offer.TotalAmount > 0 {
		total = offer.TotalAmount
	}

	currency := "USD"
	if offer != nil && strings.TrimSpace(offer.Currency) != "" {
		currency = strings.ToUpper(strings.TrimSpace(offer.Currency))
	}

	shippingAddr := modelsOrder.Address{Country: destCountry}
	if opts.ShippingAddress != nil {
		shippingAddr = *opts.ShippingAddress
	}

	orderNumber := fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102"), strings.ToUpper(crypto.GenerateSlug()))
	inquiryID := inquiry.ID
	now := time.Now()
	order := &modelsOrder.Order{
		ID:              crypto.GenerateID(),
		OrderNumber:     orderNumber,
		UserID:          userID,
		InquiryID:       &inquiryID,
		Source:          modelsOrder.OrderSourceInquiry,
		Status:          modelsOrder.OrderStatusPendingConfirm,
		PaymentStatus:   modelsOrder.PaymentStatusUnpaid,
		Items:           items,
		StockReserved:   false,
		Subtotal:        subtotal,
		TaxAmount:       0,
		ShippingAmount:  0,
		TotalAmount:     total,
		Currency:        currency,
		ShippingAddress: shippingAddr,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	s.assignWarehouse(ctx, opts.EnableMultiWarehouse, &order.WarehouseID)

	if err := s.order.CreateOrder(ctx, order); err != nil {
		if errors.Is(err, modelsOrder.ErrInsufficientStock) {
			return &IntakeResult{InventoryViolations: []string{"insufficient_stock"}}, ErrInventoryViolation
		}
		return nil, fmt.Errorf("orderintake: create order failed: %w", err)
	}

	// Link a trade transaction, snapshotting the negotiated incoterms / payment terms.
	s.createTrade(ctx, order, inquiry, offer)

	return &IntakeResult{Order: order}, nil
}

// effectiveUnitPrice returns the unit price to persist for a line. An accepted
// negotiation offer carries the agreed unit price, which is authoritative and
// replaces whatever the catalog resolved — this is what keeps the negotiated
// contract price on the order lines for both single- and multi-line inquiries
// (H10). Without an offer (or when the offer carries no unit price) the catalog
// / contract price is used. It returns 0 when no usable price can be established.
func effectiveUnitPrice(catalogPrice float64, offer *modelsOrder.NegotiationOffer) float64 {
	if offer != nil && offer.UnitPrice > 0 {
		return offer.UnitPrice
	}
	if catalogPrice <= 0 {
		return 0
	}
	return catalogPrice
}

// resolveUnitPrice mirrors the convert handler: price service → base price,
// then market cost-stack resolution for the destination country.
func (s *Service) resolveUnitPrice(ctx context.Context, product modelsProduct.Product, quantity int, destCountry string) float64 {
	var unitPrice float64
	if s.price != nil {
		if p, err := s.price.GetPriceForProduct(ctx, product.ID, "", quantity); err == nil && p > 0 {
			unitPrice = p
		}
	}
	if unitPrice <= 0 {
		unitPrice = product.BasePrice
	}
	if unitPrice <= 0 {
		return 0
	}
	return s.product.ResolveCheckoutUnitPrice(ctx, &product, unitPrice, destCountry)
}

func (s *Service) sellable(ctx context.Context, ids []string) map[string]int {
	sellable, _ := s.product.EffectiveSellableByProducts(ctx, ids, modelsProduct.ChannelWebstore)
	return sellable
}

func (s *Service) assignWarehouse(ctx context.Context, enableMulti bool, dst **string) {
	if s.product == nil || dst == nil {
		return
	}
	// channel resolver omitted: AssignDefault falls back to the product default
	// warehouse, which is correct for the single-warehouse common case.
	orderwarehouse.AssignDefault(ctx, enableMulti, s.product, nil, dst)
}

func (s *Service) createTrade(ctx context.Context, order *modelsOrder.Order, inquiry *modelsProduct.Inquiry, offer *modelsOrder.NegotiationOffer) {
	if s.trade == nil {
		return
	}
	incoterms := strings.TrimSpace(inquiry.Incoterms)
	notes := orderSvc.CommercialNotesFromInquiry(inquiry)
	if offer != nil {
		if t := strings.TrimSpace(offer.Incoterms); t != "" {
			incoterms = t
		}
		if extra := offerCommercialNotes(offer); extra != "" {
			if notes != "" {
				notes += "; " + extra
			} else {
				notes = extra
			}
		}
	}
	trade := orderSvc.BuildTradeTransactionFromOrderWithHints(order, incoterms, notes)
	// Best-effort: a failed trade link must not undo the order. Surfaced upstream
	// only as a soft signal (the order itself is the durable artifact).
	_ = s.trade.CreateTransaction(ctx, trade)
}

// MarkInquiryWon flips the inquiry to "won" so it cannot be re-converted. Best
// effort: a failure here does not invalidate the order that was just created.
func (s *Service) MarkInquiryWon(ctx context.Context, inquiryID string) error {
	if s.inquiry == nil {
		return nil
	}
	return s.inquiry.UpdateInquiryStatus(ctx, inquiryID, "won")
}

func offerCommercialNotes(offer *modelsOrder.NegotiationOffer) string {
	var parts []string
	if t := strings.TrimSpace(offer.PaymentTerms); t != "" {
		parts = append(parts, "Negotiated payment terms: "+t)
	}
	if d := strings.TrimSpace(offer.DeliveryDate); d != "" {
		parts = append(parts, "Agreed delivery: "+d)
	}
	if m := strings.TrimSpace(offer.Message); m != "" {
		parts = append(parts, "Negotiation note: "+m)
	}
	return strings.Join(parts, "; ")
}

func dedupeNonEmpty(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, raw := range in {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// parseQuantity extracts a numeric value from a free-text quantity like
// "1000 cartons", "5 tons", "500 kg". Defaults to 1 when unparseable.
var quantityNumberPattern = regexp.MustCompile(`[-+]?\d[\d,]*(?:\.\d+)?`)

func parseQuantity(qty string) int {
	match := quantityNumberPattern.FindString(strings.TrimSpace(qty))
	if match == "" {
		return 1
	}
	num, err := strconv.ParseFloat(strings.ReplaceAll(match, ",", ""), 64)
	if err != nil || num <= 0 {
		return 1
	}
	return int(num)
}
