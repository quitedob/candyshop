package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"encoding/json"
	"strings"

	"gorm.io/datatypes"
)

// OrderFinancialSnapshot 订单财务相关字段快照（用于检测是否需填写调整原因）
type OrderFinancialSnapshot struct {
	Subtotal       float64                `json:"subtotal"`
	TaxAmount      float64                `json:"taxAmount"`
	ShippingAmount float64                `json:"shippingAmount"`
	TotalAmount    float64                `json:"totalAmount"`
	Currency       string                 `json:"currency"`
	Items          modelsOrder.OrderItemArray `json:"items"`
}

// SnapshotOrderFinancial 从订单拷贝财务快照
func SnapshotOrderFinancial(o *modelsOrder.Order) OrderFinancialSnapshot {
	if o == nil {
		return OrderFinancialSnapshot{}
	}
	itemsCopy := append(modelsOrder.OrderItemArray{}, o.Items...)
	return OrderFinancialSnapshot{
		Subtotal:       o.Subtotal,
		TaxAmount:      o.TaxAmount,
		ShippingAmount: o.ShippingAmount,
		TotalAmount:    o.TotalAmount,
		Currency:       o.Currency,
		Items:          itemsCopy,
	}
}

// OrderFinancialChanged 比较两快照是否一致
func OrderFinancialChanged(a, b OrderFinancialSnapshot) bool {
	if a.Subtotal != b.Subtotal || a.TaxAmount != b.TaxAmount || a.ShippingAmount != b.ShippingAmount ||
		a.TotalAmount != b.TotalAmount || strings.TrimSpace(a.Currency) != strings.TrimSpace(b.Currency) {
		return true
	}
	return !orderItemsEqual(a.Items, b.Items)
}

func orderItemsEqual(a, b modelsOrder.OrderItemArray) bool {
	aj, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bj, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aj) == string(bj)
}

// OrderStatusRequiresFinancialReason 已确认及之后状态改财务须审计
func OrderStatusRequiresFinancialReason(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "confirmed", "production", "shipped", "delivered":
		return true
	default:
		return false
	}
}

// RecordOrderFinancialAdjustment 写入订单财务变更审计（可选 adj 为 nil 时跳过）
func (s *InvoiceService) RecordOrderFinancialAdjustment(ctx context.Context, orderID, actorUserID, reason string, before, after OrderFinancialSnapshot) error {
	if s.adj == nil || strings.TrimSpace(actorUserID) == "" {
		return nil
	}
	bj, err := json.Marshal(before)
	if err != nil {
		return err
	}
	aj, err := json.Marshal(after)
	if err != nil {
		return err
	}
	return s.adj.Create(ctx, &modelsOrder.DocumentAdjustment{
		DocType:        modelsOrder.DocAdjustmentDocTypeOrder,
		DocumentID:     strings.TrimSpace(orderID),
		RelatedOrderID: strings.TrimSpace(orderID),
		ActorUserID:    actorUserID,
		Action:         modelsOrder.DocAdjustmentActionOrderFinancialChange,
		Reason:         reason,
		BeforeSnapshot: datatypes.JSON(bj),
		AfterSnapshot:  datatypes.JSON(aj),
	})
}
