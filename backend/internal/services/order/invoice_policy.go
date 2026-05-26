package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/datatypes"
)

// Sentinel errors for invoice operations.
var (
	ErrAdjustmentReasonRequired = errors.New("adjustment_reason_required")
	ErrInvoiceOrderNotFound     = errors.New("order not found")
)

// invoiceLineSnapshot 发票行快照（字段与 admin_docx invoiceItemJSON 对齐）
type invoiceLineSnapshot struct {
	ProductID      string  `json:"productId"`
	Quantity       int     `json:"quantity"`
	UnitPrice      float64 `json:"unitPrice"`
	Specifications string  `json:"specifications,omitempty"`
	Name           string  `json:"name,omitempty"`
}

// buildInvoiceItemsJSONFromOrder 从订单行生成发票 items JSON
func buildInvoiceItemsJSONFromOrder(items modelsOrder.OrderItemArray) (string, error) {
	lines := make([]invoiceLineSnapshot, 0, len(items))
	for _, it := range items {
		lines = append(lines, invoiceLineSnapshot{
			ProductID:      strings.TrimSpace(it.ProductID),
			Quantity:       it.Quantity,
			UnitPrice:      it.UnitPrice,
			Specifications: it.Specifications,
			Name:           strings.TrimSpace(it.ProductID),
		})
	}
	b, err := json.Marshal(lines)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func invoiceSnapshotJSON(inv *modelsOrder.Invoice) (datatypes.JSON, error) {
	b, err := json.Marshal(inv)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(b), nil
}

func invoiceFinancialChanged(a, b *modelsOrder.Invoice) bool {
	if a == nil || b == nil {
		return true
	}
	return a.Amount != b.Amount ||
		a.TaxAmount != b.TaxAmount ||
		strings.TrimSpace(a.Items) != strings.TrimSpace(b.Items) ||
		strings.TrimSpace(a.Currency) != strings.TrimSpace(b.Currency)
}

// CreateInvoiceFromOrder 从已存在订单派生草稿发票（金额/行与订单一致）
func (s *InvoiceService) CreateInvoiceFromOrder(ctx context.Context, orderID string, actorUserID string) (*modelsOrder.Invoice, error) {
	if s.orders == nil {
		return nil, errors.New("order reader not configured")
	}
	oid := strings.TrimSpace(orderID)
	if oid == "" {
		return nil, errors.New("orderId is required")
	}
	ord, err := s.orders.FindByID(ctx, oid)
	if err != nil || ord == nil {
		return nil, ErrInvoiceOrderNotFound
	}
	items, err := buildInvoiceItemsJSONFromOrder(ord.Items)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	hash := OrderFinancialLineageHash(ord)
	// M-24: invoice total must mirror the order total formula
	// (subtotal + tax + shipping). Previously this dropped ShippingAmount,
	// causing derived invoices to under-bill by the freight component.
	inv := &modelsOrder.Invoice{
		OrderID:            oid,
		Type:               modelsOrder.InvoiceTypeCommercial,
		Status:             modelsOrder.InvoiceStatusDraft,
		Amount:             ord.Subtotal,
		TaxAmount:          ord.TaxAmount,
		Currency:           ord.Currency,
		Items:              items,
		TotalAmount:        ord.Subtotal + ord.TaxAmount + ord.ShippingAmount,
		LineageSource:      InvoiceLineageOrderDerived,
		DerivedAt:          &now,
		OrderFinancialHash: hash,
	}
	if err := s.CreateInvoice(ctx, inv); err != nil {
		return nil, err
	}
	if s.adj != nil && strings.TrimSpace(actorUserID) != "" {
		afterJ, err := invoiceSnapshotJSON(inv)
		if err != nil {
			log.Printf("Warning: failed to marshal invoice snapshot for adjustment audit (invoice %s): %v", inv.ID, err)
		}
		if err := s.adj.Create(ctx, &modelsOrder.DocumentAdjustment{
			DocType:        modelsOrder.DocAdjustmentDocTypeInvoice,
			DocumentID:     inv.ID,
			RelatedOrderID: oid,
			ActorUserID:    actorUserID,
			Action:         modelsOrder.DocAdjustmentActionDerivedFromOrder,
			Reason:         "Invoice created from order totals and line items",
			AfterSnapshot:  afterJ,
		}); err != nil {
			log.Printf("Warning: failed to record invoice adjustment audit (invoice %s): %v", inv.ID, err)
		}
	}
	return inv, nil
}

// ValidateAndPersistInvoiceUpdate 更新发票并校验：关联订单且非 draft 时改金额/税/行/币种须填调整原因
func (s *InvoiceService) ValidateAndPersistInvoiceUpdate(ctx context.Context, before, after *modelsOrder.Invoice, actorUserID, adjustmentReason string) error {
	if before == nil || after == nil {
		return errors.New("invalid invoice state")
	}
	after.TotalAmount = after.Amount + after.TaxAmount
	st := strings.ToLower(strings.TrimSpace(before.Status))
	orderID := strings.TrimSpace(before.OrderID)
	if orderID != "" && st != modelsOrder.InvoiceStatusDraft && invoiceFinancialChanged(before, after) {
		if strings.TrimSpace(adjustmentReason) == "" {
			return fmt.Errorf("%w: linked non-draft invoice financial fields require a reason", ErrAdjustmentReasonRequired)
		}
		if s.adj != nil && strings.TrimSpace(actorUserID) != "" {
			beforeJ, err := invoiceSnapshotJSON(before)
			if err != nil {
				return err
			}
			afterJ, err := invoiceSnapshotJSON(after)
			if err != nil {
				return err
			}
			if err := s.adj.Create(ctx, &modelsOrder.DocumentAdjustment{
				DocType:        modelsOrder.DocAdjustmentDocTypeInvoice,
				DocumentID:     before.ID,
				RelatedOrderID: orderID,
				ActorUserID:    actorUserID,
				Action:         modelsOrder.DocAdjustmentActionManualInvoiceEdit,
				Reason:         adjustmentReason,
				BeforeSnapshot: beforeJ,
				AfterSnapshot:  afterJ,
			}); err != nil {
				return err
			}
		}
	}
	return s.repo.Update(ctx, after)
}
