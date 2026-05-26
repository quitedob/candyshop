package admin

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type dashboardActivity struct {
	Type       string            `json:"type"`
	MessageKey string            `json:"messageKey"`
	Vars       map[string]string `json:"vars,omitempty"`
	CreatedAt  time.Time         `json:"createdAt"`
}

// GetDashboardStats returns statistics for the admin dashboard
// @Summary Get admin dashboard stats
// @Tags dashboard
// @Produce json
// @Router /admin/dashboard/stats [get]
func (h *Handler) GetDashboardStats(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	ctx := c.Request.Context()

	totalUsers, err := h.services.User.CountUsers(ctx)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	newUsersThisWeek, err := h.services.User.CountUsersSince(ctx, time.Now().AddDate(0, 0, -7))
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	totalOrders, err := h.services.Order.CountOrders(ctx)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	pendingOrders, err := h.services.Order.CountOrdersByStatuses(ctx, []string{"pending", "confirmed", "production", "processing"})
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	totalInquiries, err := h.services.Inquiry.CountInquiries(ctx)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	pendingInquiries, err := h.services.Inquiry.CountInquiriesByStatus(ctx, "pending")
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	totalSales, err := h.services.Order.SumSales(ctx)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	monthStart := time.Now()
	monthStart = time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, monthStart.Location())
	revenueThisMonth, err := h.services.Order.SumSalesSince(ctx, monthStart)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	// Enriched stats
	activeCustomers, _ := h.services.Order.GetDistinctOrderingUsers(ctx, monthStart)

	var avgOrderValue float64
	if totalOrders > 0 {
		avgOrderValue = totalSales / float64(totalOrders)
	}

	conversionRate := 0.0
	if totalInquiries > 0 {
		convertedInquiries, cerr := h.services.Inquiry.CountInquiriesByStatus(ctx, "won")
		if cerr == nil {
			conversionRate = float64(convertedInquiries) / float64(totalInquiries) * 100
		}
	}

	revenueByDay, _ := h.services.Order.GetRevenueByDay(ctx, 30)

	recentActivity, err := h.buildRecentActivities(ctx, 10)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_activity_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"totalUsers":       totalUsers,
		"newUsersThisWeek": newUsersThisWeek,
		"totalOrders":      totalOrders,
		"pendingOrders":    pendingOrders,
		"totalInquiries":   totalInquiries,
		"pendingInquiries": pendingInquiries,
		"totalSales":       totalSales,
		"revenueThisMonth": revenueThisMonth,
		"activeCustomers":  activeCustomers,
		"avgOrderValue":    avgOrderValue,
		"conversionRate":   conversionRate,
		"revenueByDay":     revenueByDay,
		"recentActivity":   recentActivity,
	})
}

func (h *Handler) buildRecentActivities(ctx context.Context, limit int) ([]dashboardActivity, error) {
	orders, err := h.services.Order.GetRecentOrders(ctx, limit)
	if err != nil {
		return nil, err
	}
	inquiries, err := h.services.Inquiry.GetRecentInquiries(ctx, limit)
	if err != nil {
		return nil, err
	}
	users, err := h.services.User.GetRecentUsers(ctx, limit)
	if err != nil {
		return nil, err
	}

	activities := make([]dashboardActivity, 0, len(orders)+len(inquiries)+len(users))

	for _, order := range orders {
		orderNo := strings.TrimSpace(order.OrderNumber)
		if orderNo == "" {
			orderNo = order.ID
		}
		userLabel := ""
		if order.User != nil {
			name := strings.TrimSpace(order.User.FirstName + " " + order.User.LastName)
			if name != "" {
				userLabel = name
			} else if order.User.Email != "" {
				userLabel = order.User.Email
			}
		}
		activities = append(activities, dashboardActivity{
			Type:       "order",
			MessageKey: "admin.dashboard.activity_order",
			Vars: map[string]string{
				"order_no": orderNo,
				"user":     userLabel,
			},
			CreatedAt: order.CreatedAt,
		})
	}

	for _, inquiry := range inquiries {
		company := strings.TrimSpace(inquiry.CompanyName)
		activities = append(activities, dashboardActivity{
			Type:       "inquiry",
			MessageKey: "admin.dashboard.activity_inquiry",
			Vars: map[string]string{
				"company": company,
			},
			CreatedAt: inquiry.CreatedAt,
		})
	}

	for _, user := range users {
		name := strings.TrimSpace(user.FirstName + " " + user.LastName)
		if name == "" {
			name = user.Email
		}
		activities = append(activities, dashboardActivity{
			Type:       "user",
			MessageKey: "admin.dashboard.activity_user",
			Vars: map[string]string{
				"user": name,
			},
			CreatedAt: user.CreatedAt,
		})
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].CreatedAt.After(activities[j].CreatedAt)
	})

	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}
