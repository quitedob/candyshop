package trade

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsTrade "candypro/api/internal/models/trade"
	"candypro/api/internal/pkg/shipmenttrack"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrShipmentNotInTransaction guards against cross-tenant shipment-timeline
// enumeration: a customer may only read the timeline of a shipment that belongs
// to a trade transaction they own (H4/MEDIUM-3).
var ErrShipmentNotInTransaction = errors.New("shipment does not belong to transaction")

type logisticsShipmentRepo interface {
	FindByID(ctx context.Context, id uint) (*modelsTrade.ShipmentTracking, error)
	FindByTransactionID(ctx context.Context, transactionID uint) ([]modelsTrade.ShipmentTracking, error)
	FindTrackable(ctx context.Context, limit int) ([]modelsTrade.ShipmentTracking, error)
	Update(ctx context.Context, shipment *modelsTrade.ShipmentTracking) error
	UpdateStatus(ctx context.Context, id uint, fromStatus, toStatus string, extra map[string]interface{}) error
	Dispatch(ctx context.Context, db *gorm.DB, id uint, dispatchedAt time.Time) (bool, error)
}

type logisticsEventRepo interface {
	Create(ctx context.Context, event *modelsTrade.ShipmentEvent) error
	FindByShipmentID(ctx context.Context, shipmentID uint) ([]modelsTrade.ShipmentEvent, error)
	FindByShipmentIDs(ctx context.Context, shipmentIDs []uint) ([]modelsTrade.ShipmentEvent, error)
}

type logisticsOrderRepo interface {
	FindByID(ctx context.Context, id string) (*modelsOrder.Order, error)
	Update(ctx context.Context, order *modelsOrder.Order) error
	UpdateStatusGuarded(ctx context.Context, id, fromStatus, toStatus string, extra map[string]interface{}) error
}

type logisticsTradeRepo interface {
	GetTransactionByID(ctx context.Context, id uint) (*modelsTrade.TradeTransaction, error)
}

// LogisticsService orchestrates shipment dispatch, tracking events, and delivery confirmation.
type LogisticsService struct {
	shipmentRepo logisticsShipmentRepo
	eventRepo    logisticsEventRepo
	orderRepo    logisticsOrderRepo
	tradeRepo    logisticsTradeRepo
	db           *gorm.DB
}

// NewLogisticsService creates a new LogisticsService.
func NewLogisticsService(
	sr logisticsShipmentRepo,
	er logisticsEventRepo,
	or logisticsOrderRepo,
	tr logisticsTradeRepo,
	db *gorm.DB,
) *LogisticsService {
	return &LogisticsService{
		shipmentRepo: sr,
		eventRepo:    er,
		orderRepo:    or,
		tradeRepo:    tr,
		db:           db,
	}
}

// DispatchShipment performs goods-issue: deducts stock via FEFO, updates shipment and order status.
func (s *LogisticsService) DispatchShipment(ctx context.Context, shipmentID uint, operatorID string) error {
	shipment, err := s.shipmentRepo.FindByID(ctx, shipmentID)
	if err != nil {
		return fmt.Errorf("shipment not found: %w", err)
	}
	if shipment.Status != "PENDING" {
		return fmt.Errorf("shipment %d is not in PENDING status (current: %s)", shipmentID, shipment.Status)
	}

	// Resolve trade transaction -> order
	trans, err := s.tradeRepo.GetTransactionByID(ctx, shipment.TransactionID)
	if err != nil {
		return fmt.Errorf("trade transaction not found: %w", err)
	}
	if trans.OrderID == nil || *trans.OrderID == "" {
		return fmt.Errorf("trade transaction has no linked order")
	}

	// Validate the order exists up front (cheap, no lock) so a genuinely missing
	// order fails before any transaction is opened. The authoritative, locked
	// copy used for the deduction math is re-read inside the transaction below.
	if _, err := s.orderRepo.FindByID(ctx, *trans.OrderID); err != nil {
		return fmt.Errorf("linked order not found: %w", err)
	}

	now := time.Now()

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Atomically claim the PENDING -> DISPATCHED transition. Only one
		// concurrent dispatch may win; the loser fails here before touching
		// stock or emitting an event, and its transaction rolls back cleanly.
		won, dispatchErr := s.shipmentRepo.Dispatch(ctx, tx, shipmentID, now)
		if dispatchErr != nil {
			return dispatchErr
		}
		if !won {
			return fmt.Errorf("shipment %d is no longer PENDING (concurrent dispatch)", shipmentID)
		}

		// G21-d: re-read the linked order inside the transaction under a FOR UPDATE
		// row lock before computing the remaining deduction. FulfillmentRepository.Create
		// takes the same order-row lock, so a fulfillment that commits between the
		// FindByID snapshot above and this point either (a) wins the order lock first —
		// its FulfilledQuantity is reflected in this fresh read and only the unfulfilled
		// remainder is deducted — or (b) blocks until this transaction commits and then
		// sees the StockReasonDispatched audit rows written below and skips its own
		// deduction. Without the lock and the in-transaction re-read, the dispatch would
		// iterate the stale snapshot and could re-deduct the full quantity.
		var order modelsOrder.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", *trans.OrderID).First(&order).Error; err != nil {
			return fmt.Errorf("re-read order %s for dispatch: %w", *trans.OrderID, err)
		}
		warehouseID := ""
		if order.WarehouseID != nil {
			warehouseID = *order.WarehouseID
		}

		// Dispatch is idempotent per order (G21-d, dispatch/dispatch actor pair).
		// ShipmentTracking carries no quantity field, so the FIRST dispatch on an
		// order issues the entire unfulfilled remainder as goods-issue
		// (StockReasonDispatched audit rows) and a second shipment on the same order
		// must not re-deduct that same stock. The order-row FOR UPDATE lock above
		// serializes two concurrent dispatches on the same order, so this count sees
		// the first dispatch's committed audit rows before the second can proceed —
		// without it, two shipments on one order silently deduct the order's stock
		// twice (or fail loudly when warehouse stock < 2x qty) even though the
		// multi-shipment flow is a designed path.
		var priorDispatch int64
		if err := tx.Model(&modelsOrder.StockTransaction{}).
			Where("reference_id = ? AND reason = ?", order.ID, modelsOrder.StockReasonDispatched).
			Count(&priorDispatch).Error; err != nil {
			return fmt.Errorf("dispatch idempotency check failed for order %s: %w", order.ID, err)
		}

		if priorDispatch == 0 {
			// Deduct only the quantity not already issued by a fulfillment
			// (FulfilledQuantity); a fulfillment's deduction was already applied by
			// FulfillmentRepository.Create. Deduction records are written as
			// StockReasonDispatched audit rows so a later fulfillment on the same order
			// can detect the deduction and skip it too.
			var dispatchAudit []*modelsOrder.StockTransaction
			for _, item := range order.Items {
				remaining := item.Quantity - item.FulfilledQuantity
				if remaining <= 0 {
					continue
				}
				recs, stockErr := logisticsDeductStockForProductLine(tx, warehouseID, item.ProductID, remaining, order.StockReserved, modelsOrder.StockReasonDispatched, order.ID, operatorID, now)
				if stockErr != nil {
					return fmt.Errorf("stock deduction failed for product %s: %w", item.ProductID, stockErr)
				}
				dispatchAudit = append(dispatchAudit, recs...)
			}
			if len(dispatchAudit) > 0 {
				if ae := tx.Create(&dispatchAudit).Error; ae != nil {
					return fmt.Errorf("dispatch stock audit failed: %w", ae)
				}
			}
		}

		// Create dispatch event
		event := &modelsTrade.ShipmentEvent{
			ShipmentID:  shipmentID,
			EventType:   modelsTrade.EventDispatched,
			Description: "Shipment dispatched from warehouse",
			EventTime:   now,
			OperatorID:  operatorID,
			CreatedAt:   now,
		}
		if createErr := tx.Create(event).Error; createErr != nil {
			return createErr
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Check if all shipments for this transaction are dispatched+
	s.checkAndAdvanceOrderStatus(ctx, shipment.TransactionID, "DISPATCHED", "shipped", now)

	return nil
}

// AddTrackingEvent adds a tracking event and optionally advances shipment status.
func (s *LogisticsService) AddTrackingEvent(ctx context.Context, shipmentID uint, eventType, location, description, operatorID string, eventTime time.Time) (*modelsTrade.ShipmentTracking, error) {
	shipment, err := s.shipmentRepo.FindByID(ctx, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("shipment not found: %w", err)
	}
	if shipment.Status == "DELIVERED" {
		return nil, fmt.Errorf("cannot add events to a delivered shipment")
	}

	event := &modelsTrade.ShipmentEvent{
		ShipmentID:  shipmentID,
		EventType:   eventType,
		Location:    location,
		Description: description,
		EventTime:   eventTime,
		OperatorID:  operatorID,
		CreatedAt:   time.Now(),
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	// Auto-advance shipment status based on event type
	now := time.Now()
	fromStatus := shipment.Status
	needsUpdate := false
	switch eventType {
	case modelsTrade.EventInTransit, modelsTrade.EventPickedUp:
		if shipment.Status == "DISPATCHED" || shipment.Status == "PENDING" {
			needsUpdate = true
		}
	case modelsTrade.EventArrivedAtPort, modelsTrade.EventCustomsCleared, modelsTrade.EventOutForDelivery:
		if shipment.Status != "IN_TRANSIT" {
			needsUpdate = true
		}
	}
	if needsUpdate {
		// M3: guarded conditional UPDATE keyed on the loaded status so a concurrent
		// writer cannot be silently overwritten.
		if updateErr := s.shipmentRepo.UpdateStatus(ctx, shipmentID, fromStatus, "IN_TRANSIT", nil); updateErr != nil {
			return shipment, updateErr
		}
		shipment.Status = "IN_TRANSIT"
		shipment.UpdatedAt = now
	}

	return shipment, nil
}

// ConfirmDelivery marks a shipment as delivered with proof of delivery.
func (s *LogisticsService) ConfirmDelivery(ctx context.Context, shipmentID uint, proofURL, signedBy string, deliveredAt time.Time) error {
	shipment, err := s.shipmentRepo.FindByID(ctx, shipmentID)
	if err != nil {
		return fmt.Errorf("shipment not found: %w", err)
	}
	if shipment.Status == "DELIVERED" {
		return fmt.Errorf("shipment %d is already delivered", shipmentID)
	}

	now := time.Now()
	// M3: guarded conditional UPDATE keyed on the loaded status so a concurrent
	// delivery cannot both win.
	if err := s.shipmentRepo.UpdateStatus(ctx, shipmentID, shipment.Status, "DELIVERED", map[string]interface{}{
		"delivered_at":       &deliveredAt,
		"delivery_proof_url": proofURL,
		"signed_by":          signedBy,
	}); err != nil {
		return err
	}

	// Create delivery event
	event := &modelsTrade.ShipmentEvent{
		ShipmentID:  shipmentID,
		EventType:   modelsTrade.EventDelivered,
		Description: "Shipment delivered",
		EventTime:   deliveredAt,
		OperatorID:  "",
		CreatedAt:   now,
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return err
	}

	// Check if all shipments delivered -> mark order delivered
	s.checkAndAdvanceOrderStatus(ctx, shipment.TransactionID, "DELIVERED", "delivered", now)

	return nil
}

// GetShipmentTimeline returns all tracking events for a shipment.
func (s *LogisticsService) GetShipmentTimeline(ctx context.Context, shipmentID uint) ([]modelsTrade.ShipmentEvent, error) {
	return s.eventRepo.FindByShipmentID(ctx, shipmentID)
}

// GetTransactionShipmentTimeline returns the tracking events for a shipment, but
// only when the shipment belongs to the given trade transaction. This is the
// customer-portal entry point: verifying the shipment's owning transaction inside
// the service (rather than trusting a caller-provided shipment ID) prevents an
// authenticated customer from enumerating another tenant's timeline by guessing
// shipment IDs (H4/MEDIUM-3).
func (s *LogisticsService) GetTransactionShipmentTimeline(ctx context.Context, transactionID, shipmentID uint) ([]modelsTrade.ShipmentEvent, error) {
	shipment, err := s.shipmentRepo.FindByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	if shipment.TransactionID != transactionID {
		return nil, ErrShipmentNotInTransaction
	}
	return s.eventRepo.FindByShipmentID(ctx, shipmentID)
}

// GetTransactionTimeline returns all shipments and their events for a trade transaction.
func (s *LogisticsService) GetTransactionTimeline(ctx context.Context, transactionID uint) ([]modelsTrade.ShipmentTracking, []modelsTrade.ShipmentEvent, error) {
	shipments, err := s.shipmentRepo.FindByTransactionID(ctx, transactionID)
	if err != nil {
		return nil, nil, err
	}
	if len(shipments) == 0 {
		return shipments, nil, nil
	}
	ids := make([]uint, len(shipments))
	for i, sh := range shipments {
		ids[i] = sh.ID
	}
	events, err := s.eventRepo.FindByShipmentIDs(ctx, ids)
	if err != nil {
		return shipments, nil, err
	}
	return shipments, events, nil
}

// checkAndAdvanceOrderStatus checks if all shipments have reached the target status
// and advances the linked order status accordingly.
func (s *LogisticsService) checkAndAdvanceOrderStatus(ctx context.Context, transactionID uint, targetShipmentStatus, targetOrderStatus string, now time.Time) {
	shipments, err := s.shipmentRepo.FindByTransactionID(ctx, transactionID)
	if err != nil {
		slog.Warn("advance order status: list shipments failed", "transactionID", transactionID, "error", err)
		return
	}
	if len(shipments) == 0 {
		return
	}

	allReached := true
	for _, sh := range shipments {
		reached := sh.Status == targetShipmentStatus || sh.Status == "DELIVERED"
		if targetShipmentStatus == "DISPATCHED" {
			reached = sh.Status == "DISPATCHED" || sh.Status == "IN_TRANSIT" || sh.Status == "DELIVERED"
		}
		if !reached {
			allReached = false
			break
		}
	}
	if !allReached {
		return
	}

	trans, err := s.tradeRepo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		slog.Warn("advance order status: trade transaction lookup failed", "transactionID", transactionID, "error", err)
		return
	}
	if trans.OrderID == nil {
		return
	}
	order, err := s.orderRepo.FindByID(ctx, *trans.OrderID)
	if err != nil {
		slog.Warn("advance order status: order lookup failed", "orderID", *trans.OrderID, "error", err)
		return
	}
	// H-7: enforce the same status-transition matrix as every other order mutation
	// path. Previously this directly assigned `order.Status = targetOrderStatus`,
	// allowing logistics events to silently push orders into states the matrix
	// would normally block (e.g. cancelled→shipped).
	if err := modelsOrder.ValidateOrderStatusTransition(order.Status, targetOrderStatus); err != nil {
		slog.Warn("advance order status: invalid transition skipped",
			"orderID", order.ID,
			"currentStatus", order.Status,
			"targetStatus", targetOrderStatus,
			"error", err,
		)
		return
	}
	extra := map[string]interface{}{}
	switch targetOrderStatus {
	case "shipped":
		extra["shipped_at"] = &now
	case "delivered":
		extra["delivered_at"] = &now
	}
	// M3: guarded conditional UPDATE keyed on the loaded status so a concurrent
	// order mutation cannot be silently overwritten by the logistics advance.
	if err := s.orderRepo.UpdateStatusGuarded(ctx, order.ID, order.Status, targetOrderStatus, extra); err != nil {
		slog.Warn("advance order status: order update failed", "orderID", order.ID, "targetStatus", targetOrderStatus, "error", err)
	}
}

// SyncShipmentStatusFromEvents 根据已有追踪事件推断并更新发货状态（供后台 worker 调用）。
func (s *LogisticsService) SyncShipmentStatusFromEvents(ctx context.Context, shipmentID uint) (bool, error) {
	shipment, err := s.shipmentRepo.FindByID(ctx, shipmentID)
	if err != nil {
		return false, err
	}
	if shipment.Status == "DELIVERED" || shipment.Status == "EXCEPTION" {
		return false, nil
	}

	events, err := s.eventRepo.FindByShipmentID(ctx, shipmentID)
	if err != nil {
		return false, err
	}

	lastDesc := ""
	var lastTime time.Time
	for _, ev := range events {
		if ev.EventTime.After(lastTime) {
			lastTime = ev.EventTime
			lastDesc = ev.Description
			if strings.TrimSpace(lastDesc) == "" {
				lastDesc = ev.EventType
			}
		}
	}

	resolved := shipmenttrack.ResolveTrackingStatus(lastDesc, shipment.ETA)
	newStatus := shipmenttrack.MapToDBStatus(resolved, shipment.Status)
	if newStatus == "" {
		return false, nil
	}

	now := time.Now()
	switch newStatus {
	case "DELIVERED":
		deliveredAt := lastTime
		if deliveredAt.IsZero() {
			deliveredAt = now
		}
		if err := s.ConfirmDelivery(ctx, shipmentID, shipment.DeliveryProofURL, shipment.SignedBy, deliveredAt); err != nil {
			return false, err
		}
		return true, nil
	default:
		// M3: guarded conditional UPDATE keyed on the loaded status so a concurrent
		// writer cannot be silently overwritten (IN_TRANSIT / EXCEPTION transition).
		fromStatus := shipment.Status
		if err := s.shipmentRepo.UpdateStatus(ctx, shipment.ID, fromStatus, newStatus, nil); err != nil {
			return false, err
		}
		return true, nil
	}
}

// SyncActiveShipmentTracking 批量同步活跃发货的追踪状态。
func (s *LogisticsService) SyncActiveShipmentTracking(ctx context.Context, batchSize int) (int, error) {
	shipments, err := s.shipmentRepo.FindTrackable(ctx, batchSize)
	if err != nil {
		return 0, err
	}
	updated := 0
	for _, sh := range shipments {
		ok, syncErr := s.SyncShipmentStatusFromEvents(ctx, sh.ID)
		if syncErr != nil {
			slog.Warn("shipment tracking sync failed", "shipmentID", sh.ID, "error", syncErr)
			continue
		}
		if ok {
			updated++
		}
	}
	return updated, nil
}
