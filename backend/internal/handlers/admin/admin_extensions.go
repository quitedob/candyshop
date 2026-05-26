package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/password"
	"candypro/api/internal/pkg/response"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetDashboard returns admin dashboard payload (alias endpoint).
func (h *Handler) GetDashboard(c *gin.Context) {
	h.GetDashboardStats(c)
}

// GetSalesReport returns aggregate sales report.
func (h *Handler) GetSalesReport(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	totalOrders, err := h.services.Order.CountOrders(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}
	totalSales, err := h.services.Order.SumSales(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	monthStart := time.Now()
	monthStart = time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, monthStart.Location())
	monthSales, err := h.services.Order.SumSalesSince(c.Request.Context(), monthStart)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"totalOrders":     totalOrders,
		"totalSales":      totalSales,
		"salesThisMonth":  monthSales,
		"generatedAt":     time.Now(),
		"currencyDefault": "USD",
	})
}

// GetInquiriesReport returns aggregate inquiry report.
func (h *Handler) GetInquiriesReport(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	totalInquiries, err := h.services.Inquiry.CountInquiries(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}

	// M-1: pull all status counts in a single GROUP BY query rather than firing
	// one COUNT per status. Missing statuses surface as 0.
	statuses := []string{"pending", "contacted", "quoted", "negotiating", "won", "lost", "closed"}
	grouped, gerr := h.services.Inquiry.CountInquiriesByStatusGrouped(c.Request.Context())
	if gerr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "dashboard_fetch_failed")
		return
	}
	byStatus := make(map[string]int64, len(statuses))
	for _, status := range statuses {
		byStatus[status] = grouped[status]
	}

	c.JSON(http.StatusOK, gin.H{
		"totalInquiries": totalInquiries,
		"byStatus":       byStatus,
		"generatedAt":    time.Now(),
	})
}

// AdminCreateUser creates a new user.
func (h *Handler) AdminCreateUser(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,min=8"`
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
		Company   string `json:"company"`
		Phone     string `json:"phone"`
		RoleID    string `json:"roleId"`
		RoleName  string `json:"roleName"`
		Status    string `json:"status"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	exists, err := h.services.User.EmailExists(c.Request.Context(), req.Email)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}
	if exists {
		response.ErrorResp(c, http.StatusConflict, "email_exists")
		return
	}

	if msg := password.ValidatePasswordStrength(req.Password); msg != "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	user := &modelsUser.User{
		Email:        req.Email,
		PasswordHash: req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Company:      req.Company,
		Phone:        req.Phone,
		Status:       strings.TrimSpace(req.Status),
	}

	if strings.TrimSpace(req.RoleID) != "" || strings.TrimSpace(req.RoleName) != "" {
		resolvedRoleID, roleErr := h.services.Auth.ResolveRoleID(c.Request.Context(), req.RoleID, req.RoleName)
		if roleErr != nil {
			response.InvalidResp(c, "invalid_request")
			return
		}
		user.RoleID = resolvedRoleID
	}

	if user.Status == "" {
		user.Status = "active"
	}

	if err := h.services.Auth.Register(c.Request.Context(), user); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "user_create_failed")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"company":   user.Company,
		"phone":     user.Phone,
		"status":    user.Status,
		"roleId":    user.RoleID,
	})
}

// AdminDeleteUser deletes a user by id.
func (h *Handler) AdminDeleteUser(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID := c.Param("id")
	if _, err := h.services.User.GetByID(c.Request.Context(), userID); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	if err := h.services.User.DeleteUser(c.Request.Context(), userID); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "user_delete_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
		"id":      userID,
	})
}

// AdminUpdateUserRole updates user role.
func (h *Handler) AdminUpdateUserRole(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID := c.Param("id")
	var req struct {
		RoleID   string `json:"roleId"`
		RoleName string `json:"roleName"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	resolvedRoleID, resolveErr := h.services.Auth.ResolveRoleID(c.Request.Context(), req.RoleID, req.RoleName)
	if resolveErr != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}

	user.RoleID = resolvedRoleID
	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "user_role_update_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
		"id":      user.ID,
		"roleId":  resolvedRoleID,
	})
}

// AdminGetInquiry returns inquiry details for admins.
func (h *Handler) AdminGetInquiry(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	inquiryID := c.Param("id")
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}
	resp := inquiryDetailResponse{Inquiry: inquiry, StatusHistory: []statusHistoryEntry{}}
	if h.services.ActivityLog != nil {
		if logs, logErr := h.services.ActivityLog.FindByEntity(c.Request.Context(), "inquiry", inquiryID); logErr == nil {
			resp.StatusHistory = buildStatusHistoryFromLogs(logs, "inquiry_status_change")
		}
	}
	c.JSON(http.StatusOK, resp)
}

// AdminAssignInquiry assigns inquiry to a sales/admin user.
func (h *Handler) AdminAssignInquiry(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	inquiryID := c.Param("id")
	var req struct {
		AssignedTo string `json:"assignedTo" binding:"required"`
		Priority   string `json:"priority"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}

	inquiry.AssignedTo = &req.AssignedTo
	if req.Priority != "" {
		inquiry.Priority = req.Priority
	}
	inquiry.UpdatedAt = time.Now()

	if err := h.services.Inquiry.UpdateInquiry(c.Request.Context(), inquiry); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inquiry_update_failed")
		return
	}

	c.JSON(http.StatusOK, inquiry)
}

// AdminAnalyzeInquiry 调用 AIService.AnalyzeInquiry 生成真实 AI 分析（非规则占位）。
func (h *Handler) AdminAnalyzeInquiry(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	inquiryID := c.Param("id")
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}

	inquiryText := strings.TrimSpace(inquiry.Message)
	if inquiryText == "" {
		inquiryText = strings.TrimSpace(inquiry.InternalNotes)
	}
	if inquiryText == "" {
		inquiryText = strings.TrimSpace(inquiry.CustomerNotes)
	}
	if inquiryText == "" {
		inquiryText = fmt.Sprintf("Company: %s; Contact: %s; Products: %v; Packaging: %s; Flavor: %s",
			inquiry.CompanyName, inquiry.ContactPerson, inquiry.InterestedProducts,
			inquiry.PackagingRequirements, inquiry.FlavorRequirements)
	}

	targetCountry := strings.TrimSpace(inquiry.TargetCountry)
	if targetCountry == "" {
		targetCountry = "global"
	}

	if h.aiService == nil || !h.aiService.IsEnabled() {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
		return
	}

	analysis, aiErr := h.aiService.AnalyzeInquiry(c.Request.Context(), inquiryText, targetCountry)
	if aiErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_analyze_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inquiryId": inquiry.ID,
		"analysis": gin.H{
			"priority":          inquiry.Priority,
			"riskLevel":         inferInquiryRiskLevel(analysis, inquiry.Priority),
			"recommendedAction": "ai_generated",
			"summary":           strings.TrimSpace(analysis),
			"targetCountry":     targetCountry,
		},
	})
}

// inferInquiryRiskLevel 从 AI 分析文本推断风险等级（非简单复用 priority）。
func inferInquiryRiskLevel(analysis, fallback string) string {
	lower := strings.ToLower(strings.TrimSpace(analysis))
	highMarkers := []string{"high risk", "critical", "severe", "urgent compliance", "major concern", "significant risk"}
	for _, m := range highMarkers {
		if strings.Contains(lower, m) {
			return "high"
		}
	}
	lowMarkers := []string{"low risk", "straightforward", "minimal risk", "standard product"}
	for _, m := range lowMarkers {
		if strings.Contains(lower, m) {
			return "low"
		}
	}
	mediumMarkers := []string{"medium risk", "moderate", "caution", "some concerns"}
	for _, m := range mediumMarkers {
		if strings.Contains(lower, m) {
			return "medium"
		}
	}
	if fallback != "" {
		return fallback
	}
	return "medium"
}

// AdminQuoteInquiry writes quote result back to inquiry.
func (h *Handler) AdminQuoteInquiry(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	inquiryID := c.Param("id")
	var req struct {
		QuotedAmount       float64  `json:"quotedAmount" binding:"required"`
		ValidUntil         *string  `json:"validUntil"`
		Products           []string `json:"products"`
		CustomerNotes      string   `json:"customerNotes"`
		RequestHumanReview bool     `json:"requestHumanReview"` // 与 Eino submit_quotation_for_human_review 工具配合的人审标记
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}

	inquiry.QuotedAmount = req.QuotedAmount
	now := time.Now()
	inquiry.QuotedAt = &now
	inquiry.Status = "quoted"
	inquiry.Products = modelsCommon.StringArray(req.Products)
	inquiry.CustomerNotes = req.CustomerNotes

	if req.ValidUntil != nil && *req.ValidUntil != "" {
		if parsed, parseErr := time.Parse(time.RFC3339, *req.ValidUntil); parseErr == nil {
			inquiry.ValidUntil = &parsed
		}
	}

	if err := h.services.Inquiry.UpdateInquiry(c.Request.Context(), inquiry); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inquiry_update_failed")
		return
	}

	if req.RequestHumanReview {
		c.Header("X-Quotation-Human-Review", "1")
	}
	c.JSON(http.StatusOK, inquiry)
}


	// AdminConfirmInquiry confirms packaging/weight/standards from admin side.
	// If customer has already confirmed, transitions to "confirmed" status.
	// PUT /admin/inquiries/:id/confirm
	func (h *Handler) AdminConfirmInquiry(c *gin.Context) {
		if h.services == nil {
			response.ServiceUnavailableResp(c)
			return
		}

		inquiryID := c.Param("id")
		inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
		if err != nil {
			response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
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
		inquiry.AdminConfirmed = true

		if req.Notes != "" {
			if inquiry.ConfirmationNotes != "" {
				inquiry.ConfirmationNotes += "\n[Admin] " + req.Notes
			} else {
				inquiry.ConfirmationNotes = "[Admin] " + req.Notes
			}
		}

		if inquiry.CustomerConfirmed {
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
			"success":        true,
			"message":        "Admin specifications confirmed",
			"status":         inquiry.Status,
			"fullyConfirmed": inquiry.AdminConfirmed && inquiry.CustomerConfirmed,
		})
	}
// AdminGetProducts returns paginated product list for admin.
func (h *Handler) AdminGetProducts(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	page, limit := pagination.ParsePagination(c, 20, 100)
	category := strings.TrimSpace(c.Query("category"))
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))

	products, err := h.services.Product.GetProductsForAdmin(c.Request.Context(), page, limit, category, status, search)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "product_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, products)
}

// AdminGetProduct returns a single product for admin.
func (h *Handler) AdminGetProduct(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	productID := c.Param("id")
	product, err := h.services.Product.GetProductByID(c.Request.Context(), productID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "product_not_found")
		return
	}
	c.JSON(http.StatusOK, product)
}

// AdminUpdateProductStatus updates product status only.
func (h *Handler) AdminUpdateProductStatus(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	productID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	product, err := h.services.Product.GetProductByID(c.Request.Context(), productID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "product_not_found")
		return
	}

	oldStatus := product.Status

	product.Status = strings.TrimSpace(req.Status)
	if !modelsProduct.IsValidProductStatus(product.Status) {
		response.InvalidResp(c, "product_status_invalid")
		return
	}
	if err := modelsProduct.ValidateProductStatusTransition(oldStatus, product.Status); err != nil {
		response.InvalidResp(c, "product_status_transition_invalid")
		return
	}
	if product.Status == modelsProduct.ProductStatusActive && product.StockQuantity <= 0 {
		response.InvalidResp(c, "product_activation_no_stock")
		return
	}
	product.UpdatedAt = time.Now()

	if rawID, ok := c.Get("userID"); ok {
		if uid, ok2 := rawID.(string); ok2 {
			product.UpdatedBy = &uid
		}
	}
	if err := h.services.Product.UpdateProduct(c.Request.Context(), product); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "product_update_failed")
		return
	}

	c.JSON(http.StatusOK, product)
}
