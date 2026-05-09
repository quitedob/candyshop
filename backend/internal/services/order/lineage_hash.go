package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// InvoiceLineageOrderDerived 从订单派生发票时的 lineage 标记
const InvoiceLineageOrderDerived = "order_derived"

// OrderFinancialLineageHash 订单财务域与行项目规范序列化后的 SHA256（用于发票派生审计）
func OrderFinancialLineageHash(o *modelsOrder.Order) string {
	if o == nil {
		return ""
	}
	itemsJSON, err := json.Marshal(o.Items)
	if err != nil {
		itemsJSON = []byte("[]")
	}
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%f|%f|%f|%s",
		strings.TrimSpace(o.ID),
		strings.TrimSpace(o.Currency),
		o.Subtotal,
		o.TaxAmount,
		o.ShippingAmount,
		string(itemsJSON),
	)
	return hex.EncodeToString(h.Sum(nil))
}
