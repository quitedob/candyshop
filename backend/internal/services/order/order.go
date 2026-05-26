package order

import (
	"candypro/api/internal/config"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	orderRepo "candypro/api/internal/repository/order"
	emailsvc "candypro/api/internal/services/content"
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

type orderRepository interface {
	FindAll(ctx context.Context, page, limit int, status, userID, dateFrom, dateTo string) ([]modelsOrder.Order, int64, error)
	FindByID(ctx context.Context, id string) (*modelsOrder.Order, error)
	FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsOrder.Order, int64, error)
	Create(ctx context.Context, order *modelsOrder.Order) error
	CreateWithStockReservation(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error
	Update(ctx context.Context, order *modelsOrder.Order) error
	UpdateWithOutbox(ctx context.Context, order *modelsOrder.Order, outbox *modelsOrder.EventOutbox) error
	UpdateWithOptionalStockReservationAndOutbox(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int, reserve bool, outbox *modelsOrder.EventOutbox) error
	ListPendingOutbox(ctx context.Context, eventType string, limit int) ([]modelsOrder.EventOutbox, error)
	IncrementOutboxAttempt(ctx context.Context, id uint) error
	UpdateOutboxResult(ctx context.Context, id uint, status, lastErr string, processedAt *time.Time) error
	FindInquiryTradeHints(ctx context.Context, inquiryID string) (incoterms, commercialNotes string, err error)
	UpdateWithStockAdjustment(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error
	Delete(ctx context.Context, id string) error
	ReleaseStockForOrder(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error
	DeleteWithStockRestore(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error
	ReserveStockForOrder(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error
	ConfirmPendingOrder(ctx context.Context, id string, confirmedAt time.Time) error
	ConfirmAndReserveStock(ctx context.Context, id string, stockDeltas map[string]int, confirmedAt time.Time) error
	ReleaseExpiredPendingConfirmationOrders(ctx context.Context, olderThan time.Time, limit int) (int, error)
	CountAll(ctx context.Context) (int64, error)
	CountByStatuses(ctx context.Context, statuses []string) (int64, error)
	SumTotalAmount(ctx context.Context) (float64, error)
	SumTotalAmountSince(ctx context.Context, since time.Time) (float64, error)
	FindRecent(ctx context.Context, limit int) ([]modelsOrder.Order, error)
	RevenueByMonth(ctx context.Context, months int) ([]map[string]interface{}, error)
	OrderCountByMonth(ctx context.Context, months int) ([]map[string]interface{}, error)
	TopProductsByRevenue(ctx context.Context, limit int) ([]map[string]interface{}, error)
	DistinctOrderingUsers(ctx context.Context, since time.Time) (int64, error)
	RevenueByDay(ctx context.Context, days int) ([]map[string]interface{}, error)
	SalesVelocity(ctx context.Context, months int) ([]orderRepo.SalesVelocityResult, error)
	RFMAnalysis(ctx context.Context) ([]orderRepo.RFMRecord, error)
	CustomerChurn(ctx context.Context, dormantDays int) ([]orderRepo.CustomerChurnResult, error)
	InventoryHealth(ctx context.Context, salesWindowDays int) ([]orderRepo.InventoryHealthResult, error)
	ProfitLossByPeriod(ctx context.Context, groupBy string, periods int) ([]orderRepo.ProfitLossResult, error)
	ReplenishmentSuggestions(ctx context.Context, cycleDays int, salesWindowDays int) ([]orderRepo.ReplenishmentItem, error)
}

// OrderService handles order business logic.
type OrderService struct {
	repo orderRepository
	cfg  *config.Config
}

// NewOrderService creates a new OrderService.
func NewOrderService(repo orderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// NewOrderServiceWithConfig creates a new OrderService with config for email notifications.
func NewOrderServiceWithConfig(repo orderRepository, cfg *config.Config) *OrderService {
	return &OrderService{repo: repo, cfg: cfg}
}

// GetOrders returns paginated orders with optional filters.
func (s *OrderService) GetOrders(ctx context.Context, page, limit int, status, userID, dateFrom, dateTo string) (*modelsProduct.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	orders, total, err := s.repo.FindAll(ctx, page, limit, status, userID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &modelsProduct.PaginatedResponse{
		Data: orders,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// GetOrder returns a single order by ID.
func (s *OrderService) GetOrder(ctx context.Context, id string) (*modelsOrder.Order, error) {
	return s.repo.FindByID(ctx, id)
}

// GetUserOrders returns paginated orders for a specific user.
func (s *OrderService) GetUserOrders(ctx context.Context, userID string, page, limit int) (*modelsProduct.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	orders, total, err := s.repo.FindByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &modelsProduct.PaginatedResponse{
		Data: orders,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// CreateOrder creates a new order.
func (s *OrderService) CreateOrder(ctx context.Context, order *modelsOrder.Order) error {
	return s.repo.Create(ctx, order)
}

// CreateOrderWithStockReservation creates an order and atomically reserves stock.
func (s *OrderService) CreateOrderWithStockReservation(ctx context.Context, order *modelsOrder.Order) error {
	stockDeltas := buildOrderStockDeltas(order.Items)
	return s.repo.CreateWithStockReservation(ctx, order, stockDeltas)
}

// UpdateOrder updates an existing order.
func (s *OrderService) UpdateOrder(ctx context.Context, order *modelsOrder.Order) error {
	return s.repo.Update(ctx, order)
}

// UpdateOrderWithStockAdjustment updates order and adjusts stock with provided deltas.
func (s *OrderService) UpdateOrderWithStockAdjustment(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return s.repo.UpdateWithStockAdjustment(ctx, order, stockDeltas)
}

// DeleteOrder deletes an order by ID.
func (s *OrderService) DeleteOrder(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ReserveOrderStock deducts stock for an existing order and marks it as reserved.
func (s *OrderService) ReserveOrderStock(ctx context.Context, order *modelsOrder.Order) error {
	stockDeltas := buildOrderStockDeltas(order.Items)
	return s.repo.ReserveStockForOrder(ctx, order, stockDeltas)
}

// ReleaseOrderStock restores reserved stock for an order and marks reservation released.
func (s *OrderService) ReleaseOrderStock(ctx context.Context, order *modelsOrder.Order) error {
	stockDeltas := buildOrderStockDeltas(order.Items)
	return s.repo.ReleaseStockForOrder(ctx, order, stockDeltas)
}

// DeleteOrderWithStockRestore restores reserved stock and deletes the order.
func (s *OrderService) DeleteOrderWithStockRestore(ctx context.Context, order *modelsOrder.Order) error {
	stockDeltas := buildOrderStockDeltas(order.Items)
	return s.repo.DeleteWithStockRestore(ctx, order, stockDeltas)
}

// ConfirmPendingOrder confirms an order only when it is still in pending_confirmation with reserved stock.
func (s *OrderService) ConfirmPendingOrder(ctx context.Context, id string, confirmedAt time.Time) error {
	if confirmedAt.IsZero() {
		confirmedAt = time.Now()
	}
	return s.repo.ConfirmPendingOrder(ctx, id, confirmedAt)
}

// ConfirmAndReserveOrder confirms an AI draft order and atomically reserves stock at confirmation time.
func (s *OrderService) ConfirmAndReserveOrder(ctx context.Context, id string, items []modelsOrder.OrderItem, confirmedAt time.Time) error {
	if confirmedAt.IsZero() {
		confirmedAt = time.Now()
	}
	stockDeltas := buildOrderStockDeltas(items)
	return s.repo.ConfirmAndReserveStock(ctx, id, stockDeltas, confirmedAt)
}

// ConfirmAndReserveOrderWithFinancials extends ConfirmAndReserveOrder so COGS,
// tax, shipping, total amount and currency are committed atomically with the
// status flip and stock reservation (H-15).
func (s *OrderService) ConfirmAndReserveOrderWithFinancials(ctx context.Context, id string, items []modelsOrder.OrderItem, confirmedAt time.Time, fin *orderRepo.OrderConfirmFinancials) error {
	if confirmedAt.IsZero() {
		confirmedAt = time.Now()
	}
	stockDeltas := buildOrderStockDeltas(items)
	if extRepo, ok := s.repo.(interface {
		ConfirmAndReserveStockWithFinancials(ctx context.Context, id string, stockDeltas map[string]int, confirmedAt time.Time, fin *orderRepo.OrderConfirmFinancials) error
	}); ok {
		return extRepo.ConfirmAndReserveStockWithFinancials(ctx, id, stockDeltas, confirmedAt, fin)
	}
	// Fallback for tests / mocks that don't implement the extended method.
	return s.repo.ConfirmAndReserveStock(ctx, id, stockDeltas, confirmedAt)
}

// ReleaseExpiredPendingConfirmationOrders cancels expired pending_confirmation orders and releases reserved stock.
func (s *OrderService) ReleaseExpiredPendingConfirmationOrders(ctx context.Context, olderThan time.Time, limit int) (int, error) {
	return s.repo.ReleaseExpiredPendingConfirmationOrders(ctx, olderThan, limit)
}

// CountOrders returns total order count.
func (s *OrderService) CountOrders(ctx context.Context) (int64, error) {
	return s.repo.CountAll(ctx)
}

// CountOrdersByStatuses returns order count for statuses.
func (s *OrderService) CountOrdersByStatuses(ctx context.Context, statuses []string) (int64, error) {
	return s.repo.CountByStatuses(ctx, statuses)
}

// SumSales returns total sales amount.
func (s *OrderService) SumSales(ctx context.Context) (float64, error) {
	return s.repo.SumTotalAmount(ctx)
}

// SumSalesSince returns sales amount since timestamp.
func (s *OrderService) SumSalesSince(ctx context.Context, since time.Time) (float64, error) {
	return s.repo.SumTotalAmountSince(ctx, since)
}

// GetRecentOrders returns latest orders.
func (s *OrderService) GetRecentOrders(ctx context.Context, limit int) ([]modelsOrder.Order, error) {
	return s.repo.FindRecent(ctx, limit)
}

func buildOrderStockDeltas(items []modelsOrder.OrderItem) map[string]int {
	stockDeltas := make(map[string]int, len(items))
	for _, item := range items {
		if item.Quantity < 1 {
			continue
		}
		stockDeltas[item.ProductID] += item.Quantity
	}
	return stockDeltas
}

// SendOrderStatusEmail sends a notification email for order status changes.
func (s *OrderService) SendOrderStatusEmail(order *modelsOrder.Order, userEmail, userName, newStatus string) {
	if s.cfg == nil || userEmail == "" {
		return
	}

	subject, body := s.buildOrderStatusEmail(order, userName, newStatus)
	if subject == "" {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered in order status email goroutine for order %s: %v", order.ID, r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		emailSvc := emailsvc.NewEmailService(s.cfg.Email)
		if err := emailSvc.SendEmail(ctx, userEmail, subject, body); err != nil {
			log.Printf("Failed to send order status email to %s: %v", userEmail, err)
		}
	}()
}

func (s *OrderService) buildOrderStatusEmail(order *modelsOrder.Order, userName, status string) (string, string) {
	// L-2: prefer the configured frontend URL; fall back to a generic placeholder
	// rather than http://localhost:3000, which leaks dev environment defaults
	// into production status emails when FrontendURL is unset.
	frontendURL := ""
	if s.cfg != nil && s.cfg.Security.FrontendURL != "" {
		frontendURL = strings.TrimRight(s.cfg.Security.FrontendURL, "/")
	}
	var orderLink string
	if frontendURL != "" {
		orderLink = fmt.Sprintf("%s/customer/orders/%s", frontendURL, order.ID)
	} else {
		orderLink = fmt.Sprintf("/customer/orders/%s", order.ID)
	}
	greeting := "Dear Customer"
	if userName != "" {
		greeting = fmt.Sprintf("Dear %s", userName)
	}

	switch strings.ToLower(status) {
	case "confirmed":
		subject := fmt.Sprintf("Order %s Confirmed - CandyPro OEM", order.OrderNumber)
		body := fmt.Sprintf(`%s,

Your order #%s has been confirmed and will soon enter production.

Order Details:
- Order Number: %s
- Total Amount: %.2f %s
- Items: %d product(s)

You can track your order status at: %s

Best regards,
CandyPro OEM Team`,
			greeting, order.OrderNumber, order.OrderNumber,
			order.TotalAmount, order.Currency, len(order.Items), orderLink)
		return subject, body

	case "shipped":
		subject := fmt.Sprintf("Order %s Shipped - CandyPro OEM", order.OrderNumber)
		trackingInfo := ""
		if order.TrackingNumber != "" {
			trackingInfo = fmt.Sprintf("\nTracking Number: %s", order.TrackingNumber)
		}
		body := fmt.Sprintf(`%s,

Your order #%s has been shipped!%s

Order Details:
- Order Number: %s
- Total Amount: %.2f %s

Track your shipment at: %s

Best regards,
CandyPro OEM Team`,
			greeting, order.OrderNumber, trackingInfo,
			order.OrderNumber, order.TotalAmount, order.Currency, orderLink)
		return subject, body

	case "delivered":
		subject := fmt.Sprintf("Order %s Delivered - CandyPro OEM", order.OrderNumber)
		body := fmt.Sprintf(`%s,

Your order #%s has been delivered successfully.

Order Details:
- Order Number: %s
- Total Amount: %.2f %s

Thank you for your business!

Best regards,
CandyPro OEM Team`,
			greeting, order.OrderNumber,
			order.OrderNumber, order.TotalAmount, order.Currency)
		return subject, body

	case "production":
		subject := fmt.Sprintf("Order %s In Production - CandyPro OEM", order.OrderNumber)
		body := fmt.Sprintf(`%s,

Your order #%s has entered production.

Order Details:
- Order Number: %s
- Total Amount: %.2f %s
- Items: %d product(s)

You can track your order status at: %s

Best regards,
CandyPro OEM Team`,
			greeting, order.OrderNumber, order.OrderNumber,
			order.TotalAmount, order.Currency, len(order.Items), orderLink)
		return subject, body

	case "cancelled":
		subject := fmt.Sprintf("Order %s Cancelled - CandyPro OEM", order.OrderNumber)
		body := fmt.Sprintf(`%s,

Your order #%s has been cancelled.

If you have questions about this cancellation, please contact our team.

Best regards,
CandyPro OEM Team`,
			greeting, order.OrderNumber)
		return subject, body

	case "pending_approval":
		subject := fmt.Sprintf("Order %s Awaiting Your Approval - CandyPro OEM", order.OrderNumber)
		body := fmt.Sprintf(`%s,

Order #%s requires your approval before it can proceed.

Order Details:
- Order Number: %s
- Total Amount: %.2f %s
- Items: %d product(s)

Review and approve at: %s

Best regards,
CandyPro OEM Team`,
			greeting, order.OrderNumber, order.OrderNumber,
			order.TotalAmount, order.Currency, len(order.Items), orderLink)
		return subject, body

	case "approval_approved":
		subject := fmt.Sprintf("Order %s Approved - CandyPro OEM", order.OrderNumber)
		body := fmt.Sprintf(`%s,

Your order #%s has been approved and is ready for confirmation.

Order Details:
- Order Number: %s
- Total Amount: %.2f %s

View your order at: %s

Best regards,
CandyPro OEM Team`,
			greeting, order.OrderNumber, order.OrderNumber,
			order.TotalAmount, order.Currency, orderLink)
		return subject, body

	case "approval_rejected":
		subject := fmt.Sprintf("Order %s Rejected - CandyPro OEM", order.OrderNumber)
		body := fmt.Sprintf(`%s,

Your order #%s was rejected by an approver and has been cancelled.

If you have questions, please contact our team.

Best regards,
CandyPro OEM Team`,
			greeting, order.OrderNumber)
		return subject, body
	}

	return "", ""
}

// RevenueByMonth returns monthly revenue for the last N months.
func (s *OrderService) RevenueByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return s.repo.RevenueByMonth(ctx, months)
}

// GetOrderCountByMonth returns monthly order counts for the last N months.
func (s *OrderService) GetOrderCountByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return s.repo.OrderCountByMonth(ctx, months)
}

// OrderCountByMonth returns monthly order counts for the last N months (alias for handler compatibility).
func (s *OrderService) OrderCountByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return s.repo.OrderCountByMonth(ctx, months)
}

// GetTopProductsByRevenue returns top products ranked by revenue.
func (s *OrderService) GetTopProductsByRevenue(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	return s.repo.TopProductsByRevenue(ctx, limit)
}

// TopProductsByRevenue returns top products ranked by revenue (alias for handler compatibility).
func (s *OrderService) TopProductsByRevenue(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	return s.repo.TopProductsByRevenue(ctx, limit)
}

// GetDistinctOrderingUsers counts unique users who placed orders since a date.
func (s *OrderService) GetDistinctOrderingUsers(ctx context.Context, since time.Time) (int64, error) {
	return s.repo.DistinctOrderingUsers(ctx, since)
}

// GetRevenueByDay returns daily revenue for the last N days.
func (s *OrderService) GetRevenueByDay(ctx context.Context, days int) ([]map[string]interface{}, error) {
	return s.repo.RevenueByDay(ctx, days)
}

// ── Analytics ──

// SalesVelocity returns per-product sales velocity for the last N months.
func (s *OrderService) SalesVelocity(ctx context.Context, months int) ([]orderRepo.SalesVelocityResult, error) {
	return s.repo.SalesVelocity(ctx, months)
}

// RFMAnalysis returns RFM values for all ordering users.
func (s *OrderService) RFMAnalysis(ctx context.Context) ([]orderRepo.RFMRecord, error) {
	return s.repo.RFMAnalysis(ctx)
}

// CustomerChurn finds customers at risk of churning.
func (s *OrderService) CustomerChurn(ctx context.Context, dormantDays int) ([]orderRepo.CustomerChurnResult, error) {
	return s.repo.CustomerChurn(ctx, dormantDays)
}

// InventoryHealth computes inventory health metrics for all products.
func (s *OrderService) InventoryHealth(ctx context.Context, salesWindowDays int) ([]orderRepo.InventoryHealthResult, error) {
	return s.repo.InventoryHealth(ctx, salesWindowDays)
}

// ProfitLossByPeriod returns P&L grouped by period.
func (s *OrderService) ProfitLossByPeriod(ctx context.Context, groupBy string, periods int) ([]orderRepo.ProfitLossResult, error) {
	return s.repo.ProfitLossByPeriod(ctx, groupBy, periods)
}

// ReplenishmentSuggestions computes smart replenishment for all products.
func (s *OrderService) ReplenishmentSuggestions(ctx context.Context, cycleDays int, salesWindowDays int) ([]orderRepo.ReplenishmentItem, error) {
	return s.repo.ReplenishmentSuggestions(ctx, cycleDays, salesWindowDays)
}
