package trade

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	modelsOrder "candypro/api/internal/models/order"
	tradeModels "candypro/api/internal/models/trade"
)

// orderLLMContext 供 LLM 使用的订单摘要。
type orderLLMContext struct {
	OrderID         string                `json:"orderId"`
	OrderNumber     string                `json:"orderNumber"`
	Status          string                `json:"status"`
	PaymentStatus   string                `json:"paymentStatus"`
	Currency        string                `json:"currency"`
	Subtotal        float64               `json:"subtotal"`
	TaxAmount       float64               `json:"taxAmount"`
	ShippingAmount  float64               `json:"shippingAmount"`
	TotalAmount     float64               `json:"totalAmount"`
	ShippingAddress modelsOrder.Address   `json:"shippingAddress,omitempty"`
	Items           []orderItemLLMContext `json:"items"`
}

type orderItemLLMContext struct {
	ProductID      string  `json:"productId"`
	Quantity       int     `json:"quantity"`
	UnitPrice      float64 `json:"unitPrice"`
	Specifications string  `json:"specifications,omitempty"`
	LineTotal      float64 `json:"lineTotal"`
}

// BuildOrderContextJSON 将订单行项目等业务数据序列化为 LLM 可读 JSON。
func BuildOrderContextJSON(order *modelsOrder.Order) string {
	if order == nil {
		return ""
	}
	ctx := orderLLMContext{
		OrderID:         strings.TrimSpace(order.ID),
		OrderNumber:     strings.TrimSpace(order.OrderNumber),
		Status:          strings.TrimSpace(order.Status),
		PaymentStatus:   strings.TrimSpace(order.PaymentStatus),
		Currency:        strings.TrimSpace(order.Currency),
		Subtotal:        order.Subtotal,
		TaxAmount:       order.TaxAmount,
		ShippingAmount:  order.ShippingAmount,
		TotalAmount:     order.TotalAmount,
		ShippingAddress: order.ShippingAddress,
		Items:           make([]orderItemLLMContext, 0, len(order.Items)),
	}
	for _, item := range order.Items {
		ctx.Items = append(ctx.Items, orderItemLLMContext{
			ProductID:      strings.TrimSpace(item.ProductID),
			Quantity:       item.Quantity,
			UnitPrice:      item.UnitPrice,
			Specifications: strings.TrimSpace(item.Specifications),
			LineTotal:      float64(item.Quantity) * item.UnitPrice,
		})
	}
	b, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

// BuildTradeOrderContextBlock 拼接贸易 + 关联订单上下文，供 Agent prompt 使用。
func BuildTradeOrderContextBlock(trade *tradeModels.TradeTransaction, order *modelsOrder.Order) string {
	if trade == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Trade Context:\n")
	sb.WriteString("- Trade ID: ")
	if trade.Reference != "" {
		sb.WriteString(trade.Reference)
	} else {
		sb.WriteString(strconv.FormatUint(uint64(trade.ID), 10))
	}
	if trade.ID > 0 {
		sb.WriteString(" (internal id ")
		sb.WriteString(strconv.FormatUint(uint64(trade.ID), 10))
		sb.WriteString(")")
	}
	sb.WriteString("\n")
	sb.WriteString("- Status: ")
	sb.WriteString(trade.Status)
	sb.WriteString("\n- Currency: ")
	sb.WriteString(trade.Currency)
	sb.WriteString("\n- Total Amount: ")
	sb.WriteString(strconv.FormatFloat(trade.TotalAmount, 'f', 2, 64))
	sb.WriteString("\n- Incoterms/Terms: ")
	sb.WriteString(trade.Terms)
	if trade.CommercialNotes != "" {
		sb.WriteString("\n- Commercial Notes: ")
		sb.WriteString(trade.CommercialNotes)
	}
	if orderJSON := BuildOrderContextJSON(order); orderJSON != "" {
		sb.WriteString("\n\nLinked Order (read-only business data):\n")
		sb.WriteString(orderJSON)
	} else if trade.OrderID != nil && strings.TrimSpace(*trade.OrderID) != "" {
		sb.WriteString("\n\nLinked Order ID: ")
		sb.WriteString(strings.TrimSpace(*trade.OrderID))
		sb.WriteString(" (details unavailable)")
	}
	return sb.String()
}

// BuildOrderChatPrompt 将订单 JSON 与用户问题拼成 Chatbot / Agent 提示词。
func BuildOrderChatPrompt(order *modelsOrder.Order, userMessage string) string {
	userMessage = strings.TrimSpace(userMessage)
	block := BuildOrderContextJSON(order)
	if block == "" {
		return userMessage
	}
	return fmt.Sprintf(
		"You are a B2B confectionery order assistant. Answer using the read-only order data below. "+
			"If the data is insufficient, say so clearly.\n\nOrder data:\n%s\n\nUser question:\n%s",
		block, userMessage,
	)
}
