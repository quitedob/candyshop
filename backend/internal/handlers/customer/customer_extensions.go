package customer

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"log"

	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/storage"
	"candypro/api/internal/pkg/upload"

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
// Accepts multipart/form-data with optional file attachments.
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

	// Parse multipart form (64MB max，支持视频附件)
	if err := c.Request.ParseMultipartForm(64 << 20); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "form_parse_failed")
		return
	}

	form := c.Request.MultipartForm
	getVal := func(key string) string {
		if values, ok := form.Value[key]; ok && len(values) > 0 {
			return strings.TrimSpace(values[0])
		}
		return ""
	}
	getValues := func(key string) []string {
		if values, ok := form.Value[key]; ok {
			result := make([]string, 0, len(values))
			for _, item := range values {
				if item = strings.TrimSpace(item); item != "" {
					result = append(result, item)
				}
			}
			return result
		}
		return []string{}
	}

	companyName := getVal("companyName")
	contactPerson := getVal("contactPerson")
	email := getVal("email")

	if companyName == "" || contactPerson == "" || email == "" {
		response.ErrorResp(c, http.StatusBadRequest, "inquiry_fields_required")
		return
	}

	// Handle file uploads
	var files []string
	if fileHeaders, ok := form.File["files"]; ok {
		cfg := h.cfg
		for _, fileHeader := range fileHeaders {
			if fileHeader.Size > cfg.Upload.MaxFileSize {
				response.ErrorRespDetail(c, http.StatusBadRequest, "file_size_exceeded", gin.H{
					"filename": fileHeader.Filename,
					"maxBytes": cfg.Upload.MaxFileSize,
				})
				return
			}
			contentType := fileHeader.Header.Get("Content-Type")
			if !upload.IsAllowedInquiryAttachment(contentType, fileHeader.Filename, cfg.Upload.AllowedTypes) {
				response.ErrorRespDetail(c, http.StatusBadRequest, "file_type_not_allowed", gin.H{
					"contentType": contentType,
				})
				return
			}
			f, fErr := fileHeader.Open()
			if fErr != nil {
				response.ErrorRespDetail(c, http.StatusBadRequest, "file_open_failed", gin.H{
					"filename": fileHeader.Filename,
				})
				return
			}
			url, uploadErr := h.storage.Upload(c.Request.Context(), f, storage.UploadOptions{
				FileName: fileHeader.Filename,
			})
			f.Close()
			if uploadErr != nil {
				response.ErrorRespDetail(c, http.StatusInternalServerError, "file_save_failed", gin.H{
					"filename": fileHeader.Filename,
				})
				return
			}
			files = append(files, url)
		}
	}

	oemNeeded := getVal("oemNeeded") == "true"

	inquiry := &modelsProduct.Inquiry{
		ID:                    crypto.GenerateID(),
		UserID:                &userID,
		CompanyName:           companyName,
		ContactPerson:         contactPerson,
		Email:                 email,
		WhatsApp:              getVal("whatsapp"),
		TargetCountry:         getVal("targetCountry"),
		EstimatedQuantity:     getVal("estimatedQuantity"),
		InterestedProducts:    modelsCommon.StringArray(getValues("interestedProducts")),
		ProductIDs:            modelsCommon.StringArray(getValues("productIds")),
		PackagingRequirements: getVal("packagingRequirements"),
		FlavorRequirements:    getVal("flavorRequirements"),
		OEMNeeded:             oemNeeded,
		ExpectedDelivery:      getVal("expectedDelivery"),
		Message:               getVal("message"),
		Files:                 files,
		Status:                "pending",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := h.services.Inquiry.SubmitInquiry(c.Request.Context(), inquiry, c.GetString("locale")); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inquiry_create_failed")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"message":   "Inquiry submitted successfully",
		"inquiryId": inquiry.ID,
	})
}

// isAllowedInquiryType 已迁移至 pkg/upload.IsAllowedInquiryAttachment

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
		ProductIDs            []string `json:"productIds"`
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
	if req.ProductIDs != nil {
		inquiry.ProductIDs = modelsCommon.StringArray(req.ProductIDs)
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
	if order.Status == "cancelled" {
		currentStep = -1
	} else {
		for idx, step := range steps {
			if step == order.Status {
				currentStep = idx
				break
			}
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

// CustomerUploadInquiryAttachment uploads a file to an existing inquiry.
// POST /user/inquiries/:id/attachments
func (h *Handler) CustomerUploadInquiryAttachment(c *gin.Context) {
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
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	// Parse multipart
	if err := c.Request.ParseMultipartForm(64 << 20); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "form_parse_failed")
		return
	}

	form := c.Request.MultipartForm
	attachmentNote := ""
	if values, ok := form.Value["note"]; ok && len(values) > 0 {
		attachmentNote = strings.TrimSpace(values[0])
	}

	var newFiles []string
	if fileHeaders, ok := form.File["files"]; ok {
		for _, fh := range fileHeaders {
			if fh.Size > h.cfg.Upload.MaxFileSize {
				response.ErrorRespDetail(c, http.StatusBadRequest, "file_size_exceeded", gin.H{
					"filename": fh.Filename, "maxBytes": h.cfg.Upload.MaxFileSize,
				})
				return
			}
			contentType := fh.Header.Get("Content-Type")
			if !upload.IsAllowedInquiryAttachment(contentType, fh.Filename, h.cfg.Upload.AllowedTypes) {
				response.ErrorRespDetail(c, http.StatusBadRequest, "file_type_not_allowed", gin.H{
					"contentType": contentType,
				})
				return
			}
			f, fErr := fh.Open()
			if fErr != nil {
				log.Printf("WARN: failed to open uploaded file %s: %v", fh.Filename, fErr)
				continue
			}
			url, uploadErr := h.storage.Upload(c.Request.Context(), f, storage.UploadOptions{FileName: fh.Filename})
			f.Close()
			if uploadErr != nil {
				log.Printf("WARN: failed to store uploaded file %s: %v", fh.Filename, uploadErr)
				continue
			}
			newFiles = append(newFiles, url)
		}
	}

	// Append new file URLs
	existingFiles := []string(inquiry.Files)
	existingFiles = append(existingFiles, newFiles...)
	inquiry.Files = existingFiles

	// Append note to customer notes
	if attachmentNote != "" {
		if inquiry.CustomerNotes != "" {
			inquiry.CustomerNotes += "\n" + attachmentNote
		} else {
			inquiry.CustomerNotes = attachmentNote
		}
	}

	inquiry.UpdatedAt = time.Now()
	if err := h.services.Inquiry.UpdateInquiry(c.Request.Context(), inquiry); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inquiry_update_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Attachments uploaded",
		"files":     newFiles,
		"totalFiles": len(existingFiles),
	})
}

// CustomerConfirmInquiry confirms packaging/weight/standards for the inquiry.
// POST /user/inquiries/:id/confirm
func (h *Handler) CustomerConfirmInquiry(c *gin.Context) {
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
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	var req struct {
		PackagingType   string  `json:"packagingType"`
		PackagingWeight float64 `json:"packagingWeight"`
		PackagingSize   string  `json:"packagingSize"`
		QualityStandard string  `json:"qualityStandard"`
		Notes           string  `json:"notes"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	inquiry.PackagingType = strings.TrimSpace(req.PackagingType)
	inquiry.PackagingWeight = req.PackagingWeight
	inquiry.PackagingSize = strings.TrimSpace(req.PackagingSize)
	inquiry.QualityStandard = strings.TrimSpace(req.QualityStandard)
	inquiry.CustomerConfirmed = true

	if req.Notes != "" {
		if inquiry.ConfirmationNotes != "" {
			inquiry.ConfirmationNotes += "\n[Customer] " + req.Notes
		} else {
			inquiry.ConfirmationNotes = "[Customer] " + req.Notes
		}
	}

	// If admin has already confirmed, both are confirmed → confirm status
	if inquiry.AdminConfirmed {
		inquiry.Status = "confirmed"
		now := time.Now()
		inquiry.ConfirmedAt = &now
	} else {
		inquiry.Status = "pending_confirmation"
	}

	inquiry.UpdatedAt = time.Now()
	if err := h.services.Inquiry.UpdateInquiry(c.Request.Context(), inquiry); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inquiry_update_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Specifications confirmed",
		"status":     inquiry.Status,
		"fullyConfirmed": inquiry.AdminConfirmed && inquiry.CustomerConfirmed,
	})
}
