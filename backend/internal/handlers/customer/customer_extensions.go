package customer

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func contextUserID(c *gin.Context) (string, bool) {
	rawUserID, exists := c.Get("userID")
	if !exists || rawUserID == nil {
		return "", false
	}
	userID, ok := rawUserID.(string)
	return userID, ok && userID != ""
}

// CustomerGetDashboard returns quick summary for customer portal.
func (h *Handler) CustomerGetDashboard(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	orders, err := h.services.Order.GetUserOrders(c.Request.Context(), userID, 1, 5)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	inquiries, totalInquiries, err := h.services.Inquiry.GetUserInquiries(c.Request.Context(), userID, 1, 5)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders":    orders,
		"inquiries": inquiries,
		"summary": gin.H{
			"totalOrders":    orders.Pagination.Total,
			"totalInquiries": totalInquiries,
		},
	})
}

// CustomerCreateInquiry creates an authenticated customer inquiry.
func (h *Handler) CustomerCreateInquiry(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		CompanyName           string   `json:"companyName" binding:"required"`
		ContactPerson         string   `json:"contactPerson" binding:"required"`
		Email                 string   `json:"email" binding:"required,email"`
		WhatsApp              string   `json:"whatsapp"`
		TargetCountry         string   `json:"targetCountry"`
		EstimatedQuantity     string   `json:"estimatedQuantity"`
		InterestedProducts    []string `json:"interestedProducts"`
		PackagingRequirements string   `json:"packagingRequirements"`
		FlavorRequirements    string   `json:"flavorRequirements"`
		OEMNeeded             bool     `json:"oemNeeded"`
		ExpectedDelivery      string   `json:"expectedDelivery"`
		Message               string   `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}

	inquiry := &modelsProduct.Inquiry{
		ID:                    crypto.GenerateID(),
		UserID:                &userID,
		CompanyName:           req.CompanyName,
		ContactPerson:         req.ContactPerson,
		Email:                 req.Email,
		WhatsApp:              req.WhatsApp,
		TargetCountry:         req.TargetCountry,
		EstimatedQuantity:     req.EstimatedQuantity,
		InterestedProducts:    modelsCommon.StringArray(req.InterestedProducts),
		PackagingRequirements: req.PackagingRequirements,
		FlavorRequirements:    req.FlavorRequirements,
		OEMNeeded:             req.OEMNeeded,
		ExpectedDelivery:      req.ExpectedDelivery,
		Message:               req.Message,
		Status:                "pending",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := h.services.Inquiry.SubmitInquiry(c.Request.Context(), inquiry); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inquiry_create_failed")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Inquiry submitted successfully",
		"inquiryId": inquiry.ID,
	})
}

// CustomerUpdateInquiry updates an existing inquiry that belongs to current user.
func (h *Handler) CustomerUpdateInquiry(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	inquiryID := c.Param("id")
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}

	if inquiry.UserID == nil || *inquiry.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "inquiry_no_access")
		return
	}

	var req struct {
		EstimatedQuantity     *string  `json:"estimatedQuantity"`
		InterestedProducts    []string `json:"interestedProducts"`
		PackagingRequirements *string  `json:"packagingRequirements"`
		FlavorRequirements    *string  `json:"flavorRequirements"`
		ExpectedDelivery      *string  `json:"expectedDelivery"`
		Message               *string  `json:"message"`
		CustomerNotes         *string  `json:"customerNotes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}

	if req.EstimatedQuantity != nil {
		inquiry.EstimatedQuantity = *req.EstimatedQuantity
	}
	if req.InterestedProducts != nil {
		inquiry.InterestedProducts = modelsCommon.StringArray(req.InterestedProducts)
	}
	if req.PackagingRequirements != nil {
		inquiry.PackagingRequirements = *req.PackagingRequirements
	}
	if req.FlavorRequirements != nil {
		inquiry.FlavorRequirements = *req.FlavorRequirements
	}
	if req.ExpectedDelivery != nil {
		inquiry.ExpectedDelivery = *req.ExpectedDelivery
	}
	if req.Message != nil {
		inquiry.Message = *req.Message
	}
	if req.CustomerNotes != nil {
		inquiry.CustomerNotes = *req.CustomerNotes
	}
	inquiry.UpdatedAt = time.Now()

	if err := h.services.Inquiry.UpdateInquiry(c.Request.Context(), inquiry); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}

	c.JSON(http.StatusOK, inquiry)
}

// CustomerGetOrderProgress returns a normalized order progress payload.
func (h *Handler) CustomerGetOrderProgress(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	orderID := c.Param("id")
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}
	if order.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	steps := []string{"pending_confirmation", "pending", "confirmed", "production", "shipped", "delivered"}
	currentStep := 0
	for idx, step := range steps {
		if step == order.Status {
			currentStep = idx
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"orderId":      order.ID,
		"orderNumber":  order.OrderNumber,
		"status":       order.Status,
		"currentStep":  currentStep,
		"steps":        steps,
		"trackingCode": order.TrackingNumber,
		"timestamps": gin.H{
			"confirmedAt": order.ConfirmedAt,
			"shippedAt":   order.ShippedAt,
			"deliveredAt": order.DeliveredAt,
		},
	})
}

// CustomerGetQuotes returns quote-oriented inquiry records for current user.
func (h *Handler) CustomerGetQuotes(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := pagination.ParsePagination(c, 1, 20)
	inquiries, total, err := h.services.Inquiry.GetUserInquiries(c.Request.Context(), userID, page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}

	quotes := make([]modelsProduct.Inquiry, 0, len(inquiries))
	for _, inquiry := range inquiries {
		if inquiry.Status == "quoted" || inquiry.QuotedAmount > 0 {
			quotes = append(quotes, inquiry)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       quotes,
		"pagination": pagination.BuildPagination(total, page, limit),
	})
}

// CustomerGetNotifications returns persistent notification feed from database.
func (h *Handler) CustomerGetNotifications(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := 50
	notifications, err := h.services.Notification.GetUserNotifications(c.Request.Context(), userID, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	unreadCount := 0
	for _, n := range notifications {
		if !n.IsRead {
			unreadCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        notifications,
		"total":       len(notifications),
		"unreadCount": unreadCount,
	})
}

// CustomerMarkNotificationRead marks a single notification as read.
func (h *Handler) CustomerMarkNotificationRead(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := c.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil || id == 0 {
		response.InvalidResp(c, "invalid_request")
		return
	}

	if err := h.services.Notification.MarkRead(c.Request.Context(), id, userID); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "notification_mark_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// CustomerMarkAllNotificationsRead marks all notifications for the user as read.
func (h *Handler) CustomerMarkAllNotificationsRead(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.services.Notification.MarkAllRead(c.Request.Context(), userID); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "notification_mark_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}
