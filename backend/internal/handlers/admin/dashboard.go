package admin

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/response"

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
		response.ServiceUnavailableResp(c)
		return
	}

	ctx := c.Request.Context()
	locale := c.GetString("locale")
	if locale == "" {
		locale = i18n.DefaultLocale()
	}

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
		convertedInquiries, cerr := h.services.Inquiry.CountInquiriesByStatus(ctx, "converted")
		if cerr == nil {
			conversionRate = float64(convertedInquiries) / float64(totalInquiries) * 100
		}
	}

	revenueByDay, _ := h.services.Order.GetRevenueByDay(ctx, 30)

	recentActivity, err := h.buildRecentActivities(ctx, locale, 10)
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

func (h *Handler) buildRecentActivities(ctx context.Context, locale string, limit int) ([]dashboardActivity, error) {
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
			Type: "order",
			Message: translateActivityMsg(locale, "admin.dashboard.activity_order",
				"New order {{order_no}} placed by {{user}}",
				map[string]string{"order_no": orderNo, "user": userLabel}),
			CreatedAt: order.CreatedAt,
		})
	}

	for _, inquiry := range inquiries {
		company := strings.TrimSpace(inquiry.CompanyName)
		if company == "" {
			company = "unknown company"
		}
		activities = append(activities, dashboardActivity{
			Type: "inquiry",
			Message: translateActivityMsg(locale, "admin.dashboard.activity_inquiry",
				"New B2B inquiry from {{company}}",
				map[string]string{"company": company}),
			CreatedAt: inquiry.CreatedAt,
		})
	}

	for _, user := range users {
		name := strings.TrimSpace(user.FirstName + " " + user.LastName)
		if name == "" {
			name = user.Email
		}
		activities = append(activities, dashboardActivity{
			Type: "user",
			Message: translateActivityMsg(locale, "admin.dashboard.activity_user",
				"New user registered: {{user}}",
				map[string]string{"user": name}),
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
		activities[i].Time = formatRelativeTimeLocalized(locale, now, activities[i].CreatedAt)
	}

	return activities, nil
}

// translateActivityMsg tries the i18n engine; falls back to English format string if key not found.
func translateActivityMsg(locale, key, enFallback string, vars map[string]string) string {
	if result := i18n.TranslateWithVars(locale, key, vars); result != key {
		return result
	}
	// No DB translation exists yet — substitute into English fallback
	s := enFallback
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}

// timeRelLabel holds the i18n key and English fallback for a relative-time bucket.
type timeRelLabel struct {
	maxAge   time.Duration
	key      string
	enFormat string
}

var timeRelLabels = []timeRelLabel{
	{time.Minute, "admin.dashboard.time_just_now", "just now"},
	{time.Hour, "admin.dashboard.time_minutes_ago", "%d minutes ago"},
	{24 * time.Hour, "admin.dashboard.time_hours_ago", "%d hours ago"},
	{30 * 24 * time.Hour, "admin.dashboard.time_days_ago", "%d days ago"},
}

func formatRelativeTimeLocalized(locale string, now, ts time.Time) string {
	if ts.IsZero() {
		return "unknown"
	}
	d := now.Sub(ts)
	for _, b := range timeRelLabels {
		if d < b.maxAge {
			return translateTimeLabel(locale, b, d)
		}
	}
	return ts.Format("2006-01-02")
}

func translateTimeLabel(locale string, bucket timeRelLabel, d time.Duration) string {
	var count int
	switch {
	case d < time.Minute:
		count = 0
	case d < time.Hour:
		count = int(d.Minutes())
	case d < 24*time.Hour:
		count = int(d.Hours())
	default:
		count = int(d.Hours() / 24)
	}

	vars := map[string]string{"count": fmt.Sprintf("%d", count)}
	if result := i18n.TranslateWithVars(locale, bucket.key, vars); result != bucket.key {
		return result
	}
	if count == 0 {
		return bucket.enFormat
	}
	return fmt.Sprintf(bucket.enFormat, count)
}
