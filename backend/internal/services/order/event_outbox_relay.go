package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/dberror"
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

	if err := modelsOrder.ValidateOrderStatusTransition(prev, next); err != nil {
		return err
	}

	needsReserve := next == modelsOrder.OrderStatusConfirmed && !order.StockReserved && len(order.Items) > 0

	var outbox *modelsOrder.EventOutbox
	if next == "confirmed" && prev != "confirmed" {
		payload, err := json.Marshal(map[string]string{"orderId": order.ID})
		if err != nil {
			return err
		}
		outbox = &modelsOrder.EventOutbox{
			EventType:    modelsOrder.OutboxEventOrderConfirmedCreateTrade,
			AggregateKey: order.ID,
			Payload:      datatypes.JSON(payload),
			Status:       modelsOrder.OutboxStatusPending,
			CreatedAt:    time.Now(),
		}
	}

	if needsReserve {
		stockDeltas := buildOrderStockDeltas(order.Items)
		return s.repo.UpdateWithOptionalStockReservationAndOutbox(ctx, order, stockDeltas, true, outbox)
	}
	if outbox != nil {
		return s.repo.UpdateWithOutbox(ctx, order, outbox)
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
			if ue := s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusFailed, e.Error(), &now); ue != nil {
				log.Printf("event_outbox: UpdateOutboxResult failed for event %d: %v", row.ID, ue)
			}
			continue
		}
		// ListPendingOutbox already bumped attempts atomically when claiming the row
		// (see H-14 in order_outbox.go). row.Attempts therefore reflects the
		// post-claim value; no extra IncrementOutboxAttempt call is needed here.
		attempt := row.Attempts

		order, e := s.repo.FindByID(ctx, p.OrderID)
		if e != nil || order == nil {
			if ue := s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusFailed, "order not found", &now); ue != nil {
				log.Printf("event_outbox: UpdateOutboxResult failed for event %d: %v", row.ID, ue)
			}
			continue
		}

		has, e := trade.HasTransactionForOrder(ctx, order.ID)
		if e != nil {
			// R2 A-5: previously this branch just `continue`d, leaving the
			// outbox row pending forever for permanently-broken events. Apply
			// the same dead-letter cap as the CreateTransaction path so a
			// poison-pill row eventually moves to failed.
			if attempt >= 5 {
				if ue := s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusFailed, "has-tx check: "+e.Error(), &now); ue != nil {
					log.Printf("event_outbox: UpdateOutboxResult failed for event %d: %v", row.ID, ue)
				}
			}
			continue
		}
		if has {
			if ue := s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusProcessed, "", &now); ue != nil {
				log.Printf("event_outbox: UpdateOutboxResult failed for event %d: %v", row.ID, ue)
			}
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
			if dberror.IsDuplicateKeyError(e) {
				if ue := s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusProcessed, "", &now); ue != nil {
					log.Printf("event_outbox: UpdateOutboxResult failed for event %d: %v", row.ID, ue)
				}
				processed++
				continue
			}
			if attempt >= 5 {
				if ue := s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusFailed, e.Error(), &now); ue != nil {
					log.Printf("event_outbox: UpdateOutboxResult failed for event %d: %v", row.ID, ue)
				}
			}
			continue
		}
		if ue := s.repo.UpdateOutboxResult(ctx, row.ID, modelsOrder.OutboxStatusProcessed, "", &now); ue != nil {
			log.Printf("event_outbox: UpdateOutboxResult failed for event %d: %v", row.ID, ue)
		}
		processed++
	}
	return processed, nil
}
