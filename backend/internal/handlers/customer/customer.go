package customer

import (
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/password"
	"candypro/api/internal/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CustomerGetOrders returns the logged-in customer's orders.
// @Summary Customer get orders
// @Tags customer-orders
// @Produce json
// @Router /customer/orders [get]
func (h *Handler) CustomerGetOrders(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := pagination.ParsePagination(c, 10, 50)
	orders, err := h.services.Order.GetUserOrders(c.Request.Context(), userID, page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}

	c.JSON(http.StatusOK, orders)
}

// CustomerGetOrder returns details of a single customer order.
// @Summary Customer get order
// @Tags customer-orders
// @Produce json
// @Param id path string true "Order ID"
// @Router /customer/orders/{id} [get]
func (h *Handler) CustomerGetOrder(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := c.Param("id")
	order, err := h.services.Order.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}

	isOwner := order.UserID == userID
	canApprove := false
	if h.services.Approval != nil {
		if can, cerr := h.services.Approval.UserCanApproveOrder(c.Request.Context(), userID, order); cerr == nil {
			canApprove = can
		}
	}
	if !isOwner && !canApprove {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order": order,
		"canApprove": canApprove,
		"isOwner":    isOwner,
	})
}

// CustomerGetInquiries returns the logged-in customer's inquiries.
// @Summary Customer get inquiries
// @Tags customer-inquiries
// @Produce json
// @Router /customer/inquiries [get]
func (h *Handler) CustomerGetInquiries(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := pagination.ParsePagination(c, 10, 50)
	inquiries, total, err := h.services.Inquiry.GetUserInquiries(c.Request.Context(), userID, page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       inquiries,
		"pagination": pagination.BuildPagination(total, page, limit),
	})
}

// CustomerGetInquiry returns details of a single customer inquiry.
// @Summary Customer get inquiry
// @Tags customer-inquiries
// @Produce json
// @Param id path string true "Inquiry ID"
// @Router /customer/inquiries/{id} [get]
func (h *Handler) CustomerGetInquiry(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := c.Param("id")
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}

	if inquiry.UserID == nil || *inquiry.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "inquiry_no_access")
		return
	}

	c.JSON(http.StatusOK, inquiry)
}

// CustomerUpdateProfile updates the logged-in customer's profile.
// @Summary Customer update profile
// @Tags customer-profile
// @Accept json
// @Produce json
// @Router /customer/profile [put]
func (h *Handler) CustomerUpdateProfile(c *gin.Context) {
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
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Company   string `json:"company"`
		Phone     string `json:"phone"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Company = req.Company
	user.Phone = req.Phone

	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "user_update_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":        user.ID,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"company":   user.Company,
			"phone":     user.Phone,
			"email":     user.Email,
		},
	})
}

// CustomerChangePassword changes the logged-in customer's password.
// @Summary Customer change password
// @Tags customer-profile
// @Accept json
// @Produce json
// @Router /customer/change-password [post]
func (h *Handler) CustomerChangePassword(c *gin.Context) {
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
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=8"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	if !password.CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		response.ErrorResp(c, http.StatusUnauthorized, "invalid_credentials")
		return
	}

	hash, err := password.HashPassword(req.NewPassword)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "password_hash_failed")
		return
	}

	user.PasswordHash = hash
	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "password_update_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated successfully",
	})
}
