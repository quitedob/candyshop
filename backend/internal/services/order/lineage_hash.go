package order

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/money"
)

// InvoiceLineageOrderDerived 从订单派生发票时的 lineage 标记
const InvoiceLineageOrderDerived = "order_derived"

// lineageItem is a canonical line-item projection for hashing. Money fields
// (UnitPrice) are normalized to integer cents so sub-cent float noise cannot
// leak into the hash via json.Marshal of the raw OrderItem (which carries the
// float64 UnitPrice verbatim, so 19.99 vs 19.9900006 would serialize to
// different bytes and produce different hashes for the same nominal value).
type lineageItem struct {
	ProductID         string `json:"productId"`
	Quantity          int    `json:"quantity"`
	UnitPriceCents    int64  `json:"unitPriceCents"`
	FulfilledQuantity int    `json:"fulfilledQuantity,omitempty"`
	ShippedQuantity   int    `json:"shippedQuantity,omitempty"`
	Specifications    string `json:"specifications,omitempty"`
}

// canonicalLineageItems projects OrderItems into the stable cents-normalized
// form above. make(..., 0, n) guarantees a non-nil slice, so a nil Items (not
// preloaded) and an explicit empty [] serialize identically ("[]" not "null")
// — same order, same hash regardless of preload state.
func canonicalLineageItems(items modelsOrder.OrderItemArray) []lineageItem {
	out := make([]lineageItem, 0, len(items))
	for _, it := range items {
		out = append(out, lineageItem{
			ProductID:         it.ProductID,
			Quantity:          it.Quantity,
			UnitPriceCents:    money.MoneyToCentsInt(it.UnitPrice),
			FulfilledQuantity: it.FulfilledQuantity,
			ShippedQuantity:   it.ShippedQuantity,
			Specifications:    it.Specifications,
		})
	}
	return out
}

// OrderFinancialLineageHash 订单财务域与行项目规范序列化后的 SHA256（用于发票派生审计）
func OrderFinancialLineageHash(o *modelsOrder.Order) string {
	if o == nil {
		return ""
	}
	itemsJSON, err := json.Marshal(canonicalLineageItems(o.Items))
	if err != nil {
		itemsJSON = []byte("[]")
	}
	h := sha256.New()
	// Amounts (aggregate and per-line-item) are hashed as integer cents
	// (rounded to 2 decimals) rather than fmt.Sprintf("%f", ...) or raw float64
	// json.Marshal: those keep 6 decimals of float noise, so the same nominal
	// amount recomputed as 19.9900006 vs 19.99 produced different hashes (false
	// staleness) and %f-truncated neighbours could collide (wrong dedup).
	fmt.Fprintf(h, "%s|%s|%d|%d|%d|%s",
		strings.TrimSpace(o.ID),
		strings.TrimSpace(o.Currency),
		money.MoneyToCentsInt(o.Subtotal),
		money.MoneyToCentsInt(o.TaxAmount),
		money.MoneyToCentsInt(o.ShippingAmount),
		string(itemsJSON),
	)
	return hex.EncodeToString(h.Sum(nil))
}
