package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/utils"
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	totalOrders, err := h.services.Order.CountOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to count orders",
		})
		return
	}
	totalSales, err := h.services.Order.SumSales(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to aggregate sales",
		})
		return
	}

	monthStart := time.Now()
	monthStart = time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, monthStart.Location())
	monthSales, err := h.services.Order.SumSalesSince(c.Request.Context(), monthStart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to aggregate monthly sales",
		})
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	totalInquiries, err := h.services.Inquiry.CountInquiries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to count inquiries",
		})
		return
	}

	statuses := []string{"pending", "contacted", "quoted", "negotiating", "won", "lost", "closed"}
	byStatus := make(map[string]int64, len(statuses))
	for _, status := range statuses {
		count, countErr := h.services.Inquiry.CountInquiriesByStatus(c.Request.Context(), status)
		if countErr != nil {
			c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to count inquiries by status",
			})
			return
		}
		byStatus[status] = count
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
		utils.ServiceUnavailableResponse(c)
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
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	exists, err := h.services.User.EmailExists(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to validate email",
		})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, modelsProduct.ErrorResponse{
			Error:   "email_exists",
			Message: "Email is already registered",
		})
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
			c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
				Error:   "invalid_request",
				Message: roleErr.Error(),
			})
			return
		}
		user.RoleID = resolvedRoleID
	}

	if user.Status == "" {
		user.Status = "active"
	}

	if err := h.services.Auth.Register(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create user",
		})
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID := c.Param("id")
	if _, err := h.services.User.GetByID(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	if err := h.services.User.DeleteUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete user",
		})
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID := c.Param("id")
	var req struct {
		RoleID   string `json:"roleId"`
		RoleName string `json:"roleName"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	resolvedRoleID, resolveErr := h.services.Auth.ResolveRoleID(c.Request.Context(), req.RoleID, req.RoleName)
	if resolveErr != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: resolveErr.Error(),
		})
		return
	}

	user.RoleID = resolvedRoleID
	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update role",
		})
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	inquiryID := c.Param("id")
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Inquiry not found",
		})
		return
	}
	c.JSON(http.StatusOK, inquiry)
}

// AdminAssignInquiry assigns inquiry to a sales/admin user.
func (h *Handler) AdminAssignInquiry(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	inquiryID := c.Param("id")
	var req struct {
		AssignedTo string `json:"assignedTo" binding:"required"`
		Priority   string `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Inquiry not found",
		})
		return
	}

	inquiry.AssignedTo = &req.AssignedTo
	if req.Priority != "" {
		inquiry.Priority = req.Priority
	}
	inquiry.UpdatedAt = time.Now()

	if err := h.services.Inquiry.UpdateInquiry(c.Request.Context(), inquiry); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to assign inquiry",
		})
		return
	}

	c.JSON(http.StatusOK, inquiry)
}

// AdminAnalyzeInquiry returns a lightweight AI-style analysis payload.
func (h *Handler) AdminAnalyzeInquiry(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	inquiryID := c.Param("id")
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Inquiry not found",
		})
		return
	}

	risk := "low"
	if inquiry.EstimatedQuantity == "" || inquiry.TargetCountry == "" {
		risk = "medium"
	}
	if strings.Contains(strings.ToLower(inquiry.Message), "urgent") {
		risk = "high"
	}
	recommendedAction := "follow_up_with_quote"
	if risk == "high" {
		recommendedAction = "contact_within_2_hours"
	}
	summary := "Inquiry has sufficient information for normal sales follow-up."
	if risk == "medium" {
		summary = "Inquiry requires clarification on quantity, market, or specs before quoting."
	}
	if risk == "high" {
		summary = "Inquiry contains urgency signals and should be prioritized immediately."
	}

	c.JSON(http.StatusOK, gin.H{
		"inquiryId": inquiry.ID,
		"analysis": gin.H{
			"priority":          inquiry.Priority,
			"riskLevel":         risk,
			"recommendedAction": recommendedAction,
			"summary":           summary,
		},
	})
}

// AdminQuoteInquiry writes quote result back to inquiry.
func (h *Handler) AdminQuoteInquiry(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	inquiryID := c.Param("id")
	var req struct {
		QuotedAmount  float64  `json:"quotedAmount" binding:"required"`
		ValidUntil    *string  `json:"validUntil"`
		Products      []string `json:"products"`
		CustomerNotes string   `json:"customerNotes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Inquiry not found",
		})
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
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to save quote",
		})
		return
	}

	c.JSON(http.StatusOK, inquiry)
}

// AdminGetProducts returns paginated product list for admin.
func (h *Handler) AdminGetProducts(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	page, limit := utils.ParsePagination(c, 20, 100)
	category := strings.TrimSpace(c.Query("category"))
	products, err := h.services.Product.GetProducts(c.Request.Context(), page, limit, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch products",
		})
		return
	}

	c.JSON(http.StatusOK, products)
}

// AdminGetProduct returns a single product for admin.
func (h *Handler) AdminGetProduct(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	productID := c.Param("id")
	product, err := h.services.Product.GetProductByID(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Product not found",
		})
		return
	}
	c.JSON(http.StatusOK, product)
}

// AdminUpdateProductStatus updates product status only.
func (h *Handler) AdminUpdateProductStatus(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	productID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	product, err := h.services.Product.GetProductByID(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Product not found",
		})
		return
	}

	product.Status = req.Status
	product.UpdatedAt = time.Now()
	if err := h.services.Product.UpdateProduct(c.Request.Context(), product); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update product status",
		})
		return
	}

	c.JSON(http.StatusOK, product)
}
