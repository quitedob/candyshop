package order

import (
	modelsOrder "candypro/api/internal/models/order"
	tradesvc "candypro/api/internal/services/trade"
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"gorm.io/datatypes"
)

// UpdateOrderForAdmin 管理员更新订单；当进入 confirmed 时同事务写入发件箱，供 Relay 异步创建 Trade
func (s *OrderService) UpdateOrderForAdmin(ctx context.Context, order *modelsOrder.Order, previousStatus, targetStatus string) error {
	prev := strings.ToLower(strings.TrimSpace(previousStatus))
	next := strings.ToLower(strings.TrimSpace(targetStatus))
	if next == "confirmed" && prev != "confirmed" {
		payload, err := json.Marshal(map[string]string{"orderId": order.ID})
		if err != nil {
			return err
		}
		evt := &modelsOrder.EventOutbox{
			EventType:    modelsOrder.OutboxEventOrderConfirmedCreateTrade,
			AggregateKey: order.ID,
			Payload:      datatypes.JSON(payload),
			Status:       modelsOrder.OutboxStatusPending,
			CreatedAt:    time.Now(),
		}
		return s.repo.UpdateWithOutbox(ctx, order, evt)
	}
	return s.repo.Update(ctx, order)
}

// ProcessPendingTradeOutbox 消费 pending 的 order_confirmed_create_trade 事件（幂等：已存在 Trade 则跳过）
func (s *OrderService) ProcessPendingTradeOutbox(ctx context.Context, trade *tradesvc.TradeService, limit int) (processed int, err error) {
	if trade == nil {
		return 0, nil
	}
	rows, err := s.repo.ListPendingOutbox(ctx, modelsOrder.OutboxEventOrderConfirmedCreateTrade, limit)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	for _, row := range rows {
		var p struct {
			OrderID string `json:"orderId"`
		}
		if e := json.Unmarshal(row.Payload, &p); e != nil {
			_ = s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusFailed, e.Error(), &now)
			continue
		}
		if err := s.repo.IncrementOutboxAttempt(ctx, row.ID); err != nil {
			log.Printf("Warning: outbox relay failed to increment attempt for event %d: %v", row.ID, err)
			continue
		}
		attempt := row.Attempts + 1

		order, e := s.repo.FindByID(ctx, p.OrderID)
		if e != nil || order == nil {
			_ = s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusFailed, "order not found", &now)
			continue
		}

		has, e := trade.HasTransactionForOrder(ctx, order.ID)
		if e != nil {
			continue
		}
		if has {
			_ = s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusProcessed, "", &now)
			processed++
			continue
		}

		inc := ""
		notes := ""
		if order.InquiryID != nil && strings.TrimSpace(*order.InquiryID) != "" {
			var e error
			inc, notes, e = s.repo.FindInquiryTradeHints(ctx, strings.TrimSpace(*order.InquiryID))
			if e != nil {
				inc, notes = "", ""
			}
		}
		trans := BuildTradeTransactionFromOrderWithHints(order, inc, notes)
		if e := trade.CreateTransaction(ctx, trans); e != nil {
			if attempt >= 5 {
				_ = s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusFailed, e.Error(), &now)
			}
			continue
		}
		_ = s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusProcessed, "", &now)
		processed++
	}
	return processed, nil
}
