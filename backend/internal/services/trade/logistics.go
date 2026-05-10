package trade

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsTrade "candypro/api/internal/models/trade"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type logisticsShipmentRepo interface {
	FindByID(ctx context.Context, id uint) (*modelsTrade.ShipmentTracking, error)
	FindByTransactionID(ctx context.Context, transactionID uint) ([]modelsTrade.ShipmentTracking, error)
	Update(ctx context.Context, shipment *modelsTrade.ShipmentTracking) error
}

type logisticsEventRepo interface {
	Create(ctx context.Context, event *modelsTrade.ShipmentEvent) error
	FindByShipmentID(ctx context.Context, shipmentID uint) ([]modelsTrade.ShipmentEvent, error)
	FindByShipmentIDs(ctx context.Context, shipmentIDs []uint) ([]modelsTrade.ShipmentEvent, error)
}

type logisticsOrderRepo interface {
	FindByID(ctx context.Context, id string) (*modelsOrder.Order, error)
	Update(ctx context.Context, order *modelsOrder.Order) error
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

	order, err := s.orderRepo.FindByID(ctx, *trans.OrderID)
	if err != nil {
		return fmt.Errorf("linked order not found: %w", err)
	}

	now := time.Now()

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Deduct stock for each order item only if not already reserved at order creation
		if !order.StockReserved {
			for _, item := range order.Items {
				_, stockErr := logisticsDeductStockForProductLine(tx, item.ProductID, item.Quantity, modelsOrder.StockReasonDispatched, order.ID, operatorID, now)
				if stockErr != nil {
					return fmt.Errorf("stock deduction failed for product %s: %w", item.ProductID, stockErr)
				}
			}
		}

		// Update shipment status
		shipment.Status = "DISPATCHED"
		shipment.UpdatedAt = now
		if saveErr := tx.Save(shipment).Error; saveErr != nil {
			return saveErr
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
	needsUpdate := false
	switch eventType {
	case modelsTrade.EventInTransit, modelsTrade.EventPickedUp:
		if shipment.Status == "DISPATCHED" || shipment.Status == "PENDING" {
			shipment.Status = "IN_TRANSIT"
			needsUpdate = true
		}
	case modelsTrade.EventArrivedAtPort, modelsTrade.EventCustomsCleared, modelsTrade.EventOutForDelivery:
		if shipment.Status != "IN_TRANSIT" {
			shipment.Status = "IN_TRANSIT"
			needsUpdate = true
		}
	}
	if needsUpdate {
		shipment.UpdatedAt = now
		if updateErr := s.shipmentRepo.Update(ctx, shipment); updateErr != nil {
			return shipment, updateErr
		}
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
	shipment.Status = "DELIVERED"
	shipment.DeliveredAt = &deliveredAt
	shipment.DeliveryProofURL = proofURL
	shipment.SignedBy = signedBy
	shipment.UpdatedAt = now

	if err := s.shipmentRepo.Update(ctx, shipment); err != nil {
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
	if err != nil || len(shipments) == 0 {
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
	if err != nil || trans.OrderID == nil {
		return
	}
	order, err := s.orderRepo.FindByID(ctx, *trans.OrderID)
	if err != nil {
		return
	}
	order.Status = targetOrderStatus
	order.UpdatedAt = now
	switch targetOrderStatus {
	case "shipped":
		order.ShippedAt = &now
	case "delivered":
		order.DeliveredAt = &now
	}
	if err := s.orderRepo.Update(ctx, order); err != nil {
		log.Printf("Warning: failed to update order %s status after shipment event: %v", order.ID, err)
	}
}
