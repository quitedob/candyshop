package admin

import (
	"candypro/api/internal/utils"

	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type dashboardActivity struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Time      string    `json:"time"`
	CreatedAt time.Time `json:"-"`
}

// GetDashboardStats returns statistics for the admin dashboard
// @Summary Get admin dashboard stats
// @Tags dashboard
// @Produce json
// @Router /admin/dashboard/stats [get]
func (h *Handler) GetDashboardStats(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResp(c)
		return
	}

	ctx := c.Request.Context()

	totalUsers, err := h.services.User.CountUsers(ctx)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	newUsersThisWeek, err := h.services.User.CountUsersSince(ctx, time.Now().AddDate(0, 0, -7))
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	totalOrders, err := h.services.Order.CountOrders(ctx)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	pendingOrders, err := h.services.Order.CountOrdersByStatuses(ctx, []string{"pending", "confirmed", "production", "processing"})
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	totalInquiries, err := h.services.Inquiry.CountInquiries(ctx)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	pendingInquiries, err := h.services.Inquiry.CountInquiriesByStatus(ctx, "pending")
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	totalSales, err := h.services.Order.SumSales(ctx)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	monthStart := time.Now()
	monthStart = time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, monthStart.Location())
	revenueThisMonth, err := h.services.Order.SumSalesSince(ctx, monthStart)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
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
		convertedInquiries, cerr := h.services.Inquiry.CountInquiriesByStatus(ctx, "converted")
		if cerr == nil {
			conversionRate = float64(convertedInquiries) / float64(totalInquiries) * 100
		}
	}

	revenueByDay, _ := h.services.Order.GetRevenueByDay(ctx, 30)

	recentActivity, err := h.buildRecentActivities(ctx, 10)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "dashboard_activity_fetch_failed")
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
		userLabel := "unknown customer"
		if order.User != nil {
			name := strings.TrimSpace(order.User.FirstName + " " + order.User.LastName)
			if name != "" {
				userLabel = name
			} else if order.User.Email != "" {
				userLabel = order.User.Email
			}
		}
		activities = append(activities, dashboardActivity{
			Type:      "order",
			Message:   fmt.Sprintf("New order %s placed by %s", orderNo, userLabel),
			CreatedAt: order.CreatedAt,
		})
	}

	for _, inquiry := range inquiries {
		company := strings.TrimSpace(inquiry.CompanyName)
		if company == "" {
			company = "unknown company"
		}
		activities = append(activities, dashboardActivity{
			Type:      "inquiry",
			Message:   fmt.Sprintf("New B2B inquiry from %s", company),
			CreatedAt: inquiry.CreatedAt,
		})
	}

	for _, user := range users {
		name := strings.TrimSpace(user.FirstName + " " + user.LastName)
		if name == "" {
			name = user.Email
		}
		activities = append(activities, dashboardActivity{
			Type:      "user",
			Message:   fmt.Sprintf("New user registered: %s", name),
			CreatedAt: user.CreatedAt,
		})
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].CreatedAt.After(activities[j].CreatedAt)
	})

	if len(activities) > limit {
		activities = activities[:limit]
	}

	now := time.Now()
	for i := range activities {
		activities[i].Time = formatRelativeTime(now, activities[i].CreatedAt)
	}

	return activities, nil
}

func formatRelativeTime(now, ts time.Time) string {
	if ts.IsZero() {
		return "unknown"
	}

	d := now.Sub(ts)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	}
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
	return ts.Format("2006-01-02")
}
