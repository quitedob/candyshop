package auth

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	modelsProduct "candypro/api/internal/models/product"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/roles"
	"candypro/api/internal/utils"
	emailsvc "candypro/api/internal/services/content"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Register handles user registration
// @Summary User registration
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Registration credentials"
// @Success 201 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResponse(c)
		return
	}

	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=8"`
		FirstName   string `json:"firstName" binding:"required"`
		LastName    string `json:"lastName" binding:"required"`
		CompanyName string `json:"companyName"`
	}

	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	// H9: Enforce password strength beyond min=8
	if msg := utils.ValidatePasswordStrength(req.Password); msg != "" {
		utils.InvalidRequestResponse(c, msg)
		return
	}

	exists, err := h.services.User.EmailExists(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Registration failed",
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
		Company:      req.CompanyName,
		Status:       "pending",
	}

	if err := h.services.Auth.Register(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Registration failed",
		})
		return
	}

	verifyToken, err := utils.GenerateJWT(jwt.MapClaims{
		"sub":     user.ID,
		"purpose": "verify_email",
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}, h.cfg.JWT.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Registration succeeded, but verification token generation failed",
		})
		return
	}

	// SEC-5: Send verification email asynchronously (was TODO — now implemented)
	go func() {
		if h.cfg == nil {
			return
		}
		emailSvc := emailsvc.NewEmailService(h.cfg.Email)
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}
		verifyURL := frontendURL + "/auth/verify-email?token=" + verifyToken
		subject := "Verify your CandyPro OEM account"
		body := "Dear " + user.FirstName + ",\n\nPlease verify your email by clicking the link below:\n\n" + verifyURL + "\n\nThis link expires in 24 hours.\n\nBest regards,\nCandyPro OEM Team"
		_ = emailSvc.SendEmail(context.Background(), user.Email, subject, body)
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. Please check your email to verify your account.",
		"userId":  user.ID,
	})
}

// Login authenticates a user and returns JWT tokens
// @Summary User login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResponse(c)
		return
	}

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	// SEC-7: Per-account brute-force protection — check lockout before credential check
	if h.loginTracker != nil {
		if locked, retryAfter := h.loginTracker.IsLocked(req.Email); locked {
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.JSON(http.StatusTooManyRequests, modelsProduct.ErrorResponse{
				Error:   "account_locked",
				Message: "Too many failed login attempts. Please try again later.",
			})
			return
		}
	}

	user, err := h.services.Auth.ValidateCredentials(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// SEC-7: Record failed attempt
		if h.loginTracker != nil {
			h.loginTracker.RecordFailure(req.Email)
		}
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid email or password",
		})
		return
	}

	if user.Status == "suspended" || user.Status == "deleted" {
		c.JSON(http.StatusForbidden, modelsProduct.ErrorResponse{
			Error:   "forbidden",
			Message: "Account is " + user.Status,
		})
		return
	}

	// For safety: initialize user Role if nil
	if user.Role == nil {
		user.Role = &modelsUser.RoleSnapshot{Name: roles.User} // Fallback role if missing
	}

	accessToken, err := h.services.JWT.GenerateAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate token",
		})
		return
	}

	refreshToken, err := h.services.JWT.GenerateRefreshToken(c.Request.Context(), user, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate refresh token",
		})
		return
	}

	// SEC-7: Clear failed attempt counter on successful login
	if h.loginTracker != nil {
		h.loginTracker.RecordSuccess(req.Email)
	}

	// Async update last login
	go h.services.User.UpdateLastLogin(context.Background(), user.ID)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    h.cfg.JWT.AccessTokenDuration * 60,
		"user": gin.H{
			"id":            user.ID,
			"email":         user.Email,
			"firstName":     user.FirstName,
			"lastName":      user.LastName,
			"role":          user.Role.Name,
			"status":        user.Status,
			"emailVerified": user.EmailVerified,
		},
	})
}

// RefreshToken provides a new access token and rotates the refresh token
// @Summary Refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResponse(c)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	userID, err := h.services.Auth.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve user",
		})
		return
	}

	// R4-08: Reject refresh for suspended/deleted users
	if user.Status == "suspended" || user.Status == "deleted" {
		c.JSON(http.StatusForbidden, modelsProduct.ErrorResponse{
			Error:   "forbidden",
			Message: "Account is " + user.Status,
		})
		return
	}

	if user.Role == nil {
		user.Role = &modelsUser.RoleSnapshot{Name: roles.User}
	}

	accessToken, err := h.services.JWT.GenerateAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate access token",
		})
		return
	}

	// H10: Rotate refresh token — revoke old, issue new
	_ = h.services.Auth.RevokeRefreshToken(c.Request.Context(), req.RefreshToken)
	newRefreshToken, err := h.services.JWT.GenerateRefreshToken(c.Request.Context(), user, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		// Non-fatal: return access token without new refresh token
		c.JSON(http.StatusOK, gin.H{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   h.cfg.JWT.AccessTokenDuration * 60,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    h.cfg.JWT.AccessTokenDuration * 60,
	})
}

// Logout revokes the refresh token
// @Summary User logout
// @Tags auth
// @Produce json
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResponse(c)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Even if error, just return OK to clear state in frontend
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
		return
	}

	_ = h.services.Auth.Logout(c.Request.Context(), req.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// ForgotPassword initiates the password reset flow
// @Summary Forgot password
// @Tags auth
// @Accept json
// @Produce json
// @Router /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid email",
		})
		return
	}

	user, err := h.services.User.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		// Do not reveal if email exists or not
		c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a reset link has been sent."})
		return
	}

	// SEC-14: Use 256-bit token (32 chars) instead of 64-bit GenerateID for security tokens
	token := utils.GenerateRandomString(32)
	expires := time.Now().Add(1 * time.Hour)

	// C6: Store hashed token — never store plaintext reset tokens
	tokenHash := utils.HashResetToken(token)
	user.ResetToken = &tokenHash
	user.ResetTokenExpiresAt = &expires

	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Error processing request",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If that email exists, a reset link has been sent.",
	})
}

// ResetPassword completes the password reset flow
// @Summary Reset password
// @Tags auth
// @Accept json
// @Produce json
// @Router /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=8"`
	}

	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	// H9: Enforce password strength on reset
	if msg := utils.ValidatePasswordStrength(req.NewPassword); msg != "" {
		utils.InvalidRequestResponse(c, msg)
		return
	}

	// C6: Hash the token before querying — tokens are stored hashed
	tokenHash := utils.HashResetToken(req.Token)
	user, err := h.services.User.GetByResetToken(c.Request.Context(), tokenHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid or expired reset token",
		})
		return
	}

	if user.ResetTokenExpiresAt == nil || user.ResetTokenExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "expired_token",
			Message: "Reset token has expired",
		})
		return
	}

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to hash password",
		})
		return
	}

	user.PasswordHash = hash
	user.ResetToken = nil
	user.ResetTokenExpiresAt = nil

	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been successfully reset",
	})
}

func authContextUserID(c *gin.Context) (string, bool) {
	rawUserID, exists := c.Get("userID")
	if !exists || rawUserID == nil {
		return "", false
	}
	userID, ok := rawUserID.(string)
	return userID, ok && userID != ""
}

// VerifyEmail marks user email as verified.
// @Summary Verify email
// @Tags auth
// @Accept json
// @Produce json
// @Router /auth/verify-email [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	var req struct {
		Token string `json:"token" binding:"required"`
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "token is required",
		})
		return
	}

	claims, err := utils.ValidateJWT(req.Token, h.cfg.JWT.Secret)
	if err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid verification token",
		})
		return
	}
	purpose, ok := claims["purpose"].(string)
	if !ok || purpose != "verify_email" {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_token",
			Message: "Token purpose mismatch",
		})
		return
	}
	userID, ok := claims["sub"].(string)
	if !ok || strings.TrimSpace(userID) == "" {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid verification token subject",
		})
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "invalid_token",
			Message: "User not found",
		})
		return
	}

	now := time.Now()
	user.EmailVerified = true
	user.EmailVerifiedAt = &now
	user.Status = "active"
	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to verify email",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

// Me returns current user profile.
// @Summary Get current user
// @Tags auth
// @Produce json
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := authContextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not identified",
		})
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

	role := roles.User
	if user.Role != nil && user.Role.Name != "" {
		role = user.Role.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            user.ID,
		"email":         user.Email,
		"firstName":     user.FirstName,
		"lastName":      user.LastName,
		"company":       user.Company,
		"phone":         user.Phone,
		"status":        user.Status,
		"role":          role,
		"emailVerified": user.EmailVerified,
	})
}

// UpdateProfile updates current authenticated user's profile.
// @Summary Update current user profile
// @Tags auth
// @Accept json
// @Produce json
// @Router /auth/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := authContextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not identified",
		})
		return
	}

	var req struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Company   string `json:"company"`
		Phone     string `json:"phone"`
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

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Company = req.Company
	user.Phone = req.Phone

	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

// ChangePassword updates current authenticated user's password.
// @Summary Change current user password
// @Tags auth
// @Accept json
// @Produce json
// @Router /auth/change-password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := authContextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not identified",
		})
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=8"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	// H9: Enforce password strength on new password
	if msg := utils.ValidatePasswordStrength(req.NewPassword); msg != "" {
		utils.InvalidRequestResponse(c, msg)
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

	if !utils.CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Current password is incorrect",
		})
		return
	}

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to hash password",
		})
		return
	}

	now := time.Now()
	user.PasswordHash = hash
	user.PasswordChangedAt = &now
	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update password",
		})
		return
	}

	// SEC-8: Revoke all refresh tokens after password change
	if h.services.Auth != nil {
		_ = h.services.Auth.RevokeAllUserTokens(c.Request.Context(), userID)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated successfully",
	})
}

// ResendVerificationEmail resends the email verification link to the authenticated user.
func (h *Handler) ResendVerificationEmail(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := authContextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not identified",
		})
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

	if user.EmailVerified {
		c.JSON(http.StatusConflict, modelsProduct.ErrorResponse{
			Error:   "already_verified",
			Message: "Email is already verified",
		})
		return
	}

	verifyToken, err := utils.GenerateJWT(jwt.MapClaims{
		"sub":     user.ID,
		"purpose": "verify_email",
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}, h.cfg.JWT.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate verification token",
		})
		return
	}

	// Send verification email asynchronously
	go func() {
		if h.cfg == nil {
			return
		}
		emailSvc := emailsvc.NewEmailService(h.cfg.Email)
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}
		verifyURL := frontendURL + "/auth/verify-email?token=" + verifyToken
		subject := "Verify your CandyPro OEM account"
		body := "Dear " + user.FirstName + ",\n\nPlease verify your email by clicking the link below:\n\n" + verifyURL + "\n\nThis link expires in 24 hours.\n\nBest regards,\nCandyPro OEM Team"
		_ = emailSvc.SendEmail(context.Background(), user.Email, subject, body)
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent. Please check your inbox.",
	})
}
