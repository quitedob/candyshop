package customer

import (
	"candypro/api/internal/utils"
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
		utils.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := utils.ParsePagination(c, 10, 50)
	orders, err := h.services.Order.GetUserOrders(c.Request.Context(), userID, page, limit)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "internal_error")
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
		utils.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := c.Param("id")
	order, err := h.services.Order.GetOrder(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}

	if order.UserID != userID {
		utils.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	c.JSON(http.StatusOK, order)
}

// CustomerGetInquiries returns the logged-in customer's inquiries.
// @Summary Customer get inquiries
// @Tags customer-inquiries
// @Produce json
// @Router /customer/inquiries [get]
func (h *Handler) CustomerGetInquiries(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := utils.ParsePagination(c, 10, 50)
	inquiries, total, err := h.services.Inquiry.GetUserInquiries(c.Request.Context(), userID, page, limit)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       inquiries,
		"pagination": utils.BuildPagination(total, page, limit),
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
		utils.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := c.Param("id")
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}

	if inquiry.UserID == nil || *inquiry.UserID != userID {
		utils.ErrorResp(c, http.StatusForbidden, "inquiry_no_access")
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
		utils.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Company   string `json:"company"`
		Phone     string `json:"phone"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Company = req.Company
	user.Phone = req.Phone

	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "user_update_failed")
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
		utils.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=8"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	if !utils.CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		utils.ErrorResp(c, http.StatusUnauthorized, "invalid_credentials")
		return
	}

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "password_hash_failed")
		return
	}

	user.PasswordHash = hash
	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "password_update_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated successfully",
	})
}
