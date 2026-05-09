package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	modelsTrade "candypro/api/internal/models/trade"
	"strings"
)

// BuildTradeTransactionFromOrder 根据订单组装贸易主单（Incoterms 等由调用方传入）
func BuildTradeTransactionFromOrder(order *modelsOrder.Order, terms string) *modelsTrade.TradeTransaction {
	t := strings.TrimSpace(terms)
	if t == "" {
		t = "FOB"
	}
	cur := strings.TrimSpace(order.Currency)
	if cur == "" {
		cur = "USD"
	}
	tx := &modelsTrade.TradeTransaction{
		UserID:      order.UserID,
		OrderID:     &order.ID,
		Status:      modelsTrade.TradeStatusPending,
		Currency:    cur,
		TotalAmount: order.TotalAmount,
		Terms:       t,
	}
	if order.InquiryID != nil {
		tx.InquiryID = order.InquiryID
	}
	return tx
}

// CommercialNotesFromInquiry 将询盘中的包装与付款条款拼成贸易备注
func CommercialNotesFromInquiry(inq *modelsProduct.Inquiry) string {
	if inq == nil {
		return ""
	}
	var parts []string
	if s := strings.TrimSpace(inq.PackagingRequirements); s != "" {
		parts = append(parts, "Packaging: "+s)
	}
	if s := strings.TrimSpace(inq.NegotiatedPaymentTerms); s != "" {
		parts = append(parts, "Payment terms: "+s)
	}
	if s := strings.TrimSpace(inq.FlavorRequirements); s != "" {
		parts = append(parts, "Product specs: "+s)
	}
	return strings.Join(parts, "; ")
}

// BuildTradeTransactionFromOrderWithHints 使用 Incoterms 与商务备注构建 Trade
func BuildTradeTransactionFromOrderWithHints(order *modelsOrder.Order, incoterms, commercialNotes string) *modelsTrade.TradeTransaction {
	tx := BuildTradeTransactionFromOrder(order, incoterms)
	tx.CommercialNotes = strings.TrimSpace(commercialNotes)
	return tx
}

// BuildTradeTransactionFromInquiry 从询盘字段映射贸易条款与备注
func BuildTradeTransactionFromInquiry(order *modelsOrder.Order, inq *modelsProduct.Inquiry) *modelsTrade.TradeTransaction {
	terms := ""
	if inq != nil {
		terms = strings.TrimSpace(inq.Incoterms)
	}
	return BuildTradeTransactionFromOrderWithHints(order, terms, CommercialNotesFromInquiry(inq))
}
