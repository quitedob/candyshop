package auth

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	modelsAuth "candypro/api/internal/models/auth"
	modelsCommon "candypro/api/internal/models/common"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/pkg/authcookie"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/jwtutil"
	"candypro/api/internal/pkg/password"
	"candypro/api/internal/pkg/response"
	emailsvc "candypro/api/internal/services/content"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// frontendURL 返回配置中的前端地址，用于邮件链接。
func (h *Handler) frontendURL() string {
	if h.cfg != nil && strings.TrimSpace(h.cfg.Security.FrontendURL) != "" {
		return strings.TrimRight(strings.TrimSpace(h.cfg.Security.FrontendURL), "/")
	}
	if u := os.Getenv("FRONTEND_URL"); strings.TrimSpace(u) != "" {
		return strings.TrimRight(strings.TrimSpace(u), "/")
	}
	return "http://localhost:3000"
}

// sendVerificationEmailAsync 异步发送邮箱验证邮件。
func (h *Handler) sendVerificationEmailAsync(user *modelsUser.User, verifyToken string) {
	go func() {
		if h.cfg == nil || user == nil {
			return
		}
		emailSvc := emailsvc.NewEmailService(h.cfg.Email)
		verifyURL := h.frontendURL() + "/auth/verify-email?token=" + verifyToken
		subject := "Verify your CandyPro OEM account"
		body := "Dear " + user.FirstName + ",\n\nPlease verify your email by clicking the link below:\n\n" + verifyURL + "\n\nThis link expires in 24 hours.\n\nBest regards,\nCandyPro OEM Team"
		if e := emailSvc.SendEmail(context.Background(), user.Email, subject, body); e != nil {
			log.Printf("auth: verification email failed for %s: %v", user.Email, e)
		}
	}()
}

// issueAuthSession 生成 JWT 并写入 Cookie，供注册/登录复用。
func (h *Handler) issueAuthSession(c *gin.Context, user *modelsUser.User) error {
	if user.Role == nil {
		user.Role = &modelsUser.RoleSnapshot{Name: modelsAuth.User}
	}
	accessToken, err := h.services.JWT.GenerateAccessToken(c.Request.Context(), user)
	if err != nil {
		return err
	}
	refreshToken, err := h.services.JWT.GenerateRefreshToken(c.Request.Context(), user, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		return err
	}
	authcookie.SetAuthCookies(c, h.cfg, accessToken, refreshToken)
	return nil
}

// Register handles user registration
// @Summary User registration
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Registration credentials"
// @Success 201 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=8"`
		FirstName   string `json:"firstName" binding:"required"`
		LastName    string `json:"lastName" binding:"required"`
		CompanyName string `json:"companyName"`
	}

	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	// H9: Enforce password strength beyond min=8
	if msg := password.ValidatePasswordStrength(req.Password); msg != "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	exists, err := h.services.User.EmailExists(c.Request.Context(), req.Email)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "registration_failed")
		return
	}
	if exists {
		response.ErrorResp(c, http.StatusConflict, "email_exists")
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
		response.ErrorResp(c, http.StatusInternalServerError, "registration_failed")
		return
	}

	// 注册时若填写了公司名称，自动创建 pending 公司并关联用户
	if h.services.Company != nil && strings.TrimSpace(req.CompanyName) != "" {
		if _, err := h.services.Company.EnsureCompanyForUser(c.Request.Context(), user); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "company_provision_failed")
			return
		}
	}

	verifyToken, err := jwtutil.GenerateJWT(jwt.MapClaims{
		"sub":     user.ID,
		"purpose": "verify_email",
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}, h.cfg.JWT.Secret)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "registration_token_failed")
		return
	}

	h.sendVerificationEmailAsync(user, verifyToken)

	role := modelsAuth.User
	if user.Role != nil && user.Role.Name != "" {
		role = user.Role.Name
	}
	if err := h.issueAuthSession(c, user); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "token_generation_failed")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. Please check your email to verify your account.",
		"userId":  user.ID,
		"user": gin.H{
			"id":            user.ID,
			"email":         user.Email,
			"firstName":     user.FirstName,
			"lastName":      user.LastName,
			"role":          role,
			"status":        user.Status,
			"emailVerified": user.EmailVerified,
		},
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
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	// SEC-7: Per-account brute-force protection — check lockout before credential check
	if h.loginTracker != nil {
		if locked, retryAfter := h.loginTracker.IsLocked(req.Email); locked {
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			response.ErrorResp(c, http.StatusTooManyRequests, "login_rate_limited")
			return
		}
	}

	user, err := h.services.Auth.ValidateCredentials(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// SEC-7: Record failed attempt
		if h.loginTracker != nil {
			h.loginTracker.RecordFailure(req.Email)
		}
		response.ErrorResp(c, http.StatusUnauthorized, "invalid_credentials")
		return
	}

	if user.Status == "suspended" || user.Status == "deleted" {
		c.JSON(http.StatusForbidden, modelsCommon.ErrorResponse{
			Error:   "account_status",
			Message: i18n.TWithVars(c, "errors.account_status", map[string]string{"status": user.Status}),
		})
		return
	}

	// For safety: initialize user Role if nil
	if user.Role == nil {
		user.Role = &modelsUser.RoleSnapshot{Name: modelsAuth.User} // Fallback role if missing
	}

	accessToken, err := h.services.JWT.GenerateAccessToken(c.Request.Context(), user)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "token_generation_failed")
		return
	}

	refreshToken, err := h.services.JWT.GenerateRefreshToken(c.Request.Context(), user, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "refresh_token_failed")
		return
	}

	// SEC-7: Clear failed attempt counter on successful login
	if h.loginTracker != nil {
		h.loginTracker.RecordSuccess(req.Email)
	}

	// Async update last login
	go func() {
		if err := h.services.User.UpdateLastLogin(context.Background(), user.ID); err != nil {
			log.Printf("Warning: failed to update last login for user %s: %v", user.ID, err)
		}
	}()

	// 必须在 c.JSON 之前写入 Cookie，否则 WriteHeader 后 Set-Cookie 会被丢弃
	authcookie.SetAuthCookies(c, h.cfg, accessToken, refreshToken)
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
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)
	refreshTok := authcookie.RefreshTokenFromRequest(c, req.RefreshToken)
	if refreshTok == "" {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := h.services.Auth.ValidateRefreshToken(c.Request.Context(), refreshTok)
	if err != nil {
		c.JSON(http.StatusUnauthorized, modelsCommon.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "user_retrieve_failed")
		return
	}

	// R4-08: Reject refresh for suspended/deleted users
	if user.Status == "suspended" || user.Status == "deleted" {
		c.JSON(http.StatusForbidden, modelsCommon.ErrorResponse{
			Error:   "account_status",
			Message: i18n.TWithVars(c, "errors.account_status", map[string]string{"status": user.Status}),
		})
		return
	}

	if user.Role == nil {
		user.Role = &modelsUser.RoleSnapshot{Name: modelsAuth.User}
	}

	h.services.JWT.RevokeAccessTokenByString(c.Request.Context(), authcookie.TokenFromRequest(c))

	accessToken, err := h.services.JWT.GenerateAccessToken(c.Request.Context(), user)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "access_token_failed")
		return
	}

	// H10: Rotate refresh token — issue new first, then revoke old
	newRefreshToken, err := h.services.JWT.GenerateRefreshToken(c.Request.Context(), user, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		// 新 refresh 写入失败：保留旧 token（尚未撤销），仅返回 access token
		authcookie.SetAuthCookies(c, h.cfg, accessToken, refreshTok)
		c.JSON(http.StatusOK, gin.H{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   h.cfg.JWT.AccessTokenDuration * 60,
		})
		return
	}
	if e := h.services.Auth.RevokeRefreshToken(c.Request.Context(), refreshTok); e != nil {
		log.Printf("auth: RevokeRefreshToken failed: %v", e)
	}

	authcookie.SetAuthCookies(c, h.cfg, accessToken, newRefreshToken)
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
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.RefreshToken = authcookie.RefreshTokenFromRequest(c, "")
	}

	refreshTok := authcookie.RefreshTokenFromRequest(c, req.RefreshToken)
	h.services.JWT.RevokeAccessTokenByString(c.Request.Context(), authcookie.TokenFromRequest(c))
	if refreshTok != "" {
		if e := h.services.Auth.Logout(c.Request.Context(), refreshTok); e != nil {
			log.Printf("auth: Logout failed: %v", e)
		}
	}
	authcookie.ClearAuthCookies(c, h.cfg)
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
		response.ServiceUnavailableResp(c)
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid_email")
		return
	}

	user, err := h.services.User.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		// Do not reveal if email exists or not
		c.JSON(http.StatusOK, gin.H{"message": i18n.T(c, "errors.forgot_password_email_sent")})
		return
	}

	// M-22: token now contains only random bytes — no userID prefix.
	// The user is resolved through the password_reset_tokens table by lookup_key.
	token := crypto.GenerateRandomString(48)
	tokenHash, hashErr := crypto.HashResetToken(token)
	if hashErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "request_processing_error")
		return
	}
	expires := time.Now().Add(1 * time.Hour)
	if h.services.PasswordResetToken != nil {
		// Invalidate any active tokens for this user before issuing a new one
		// (prevents stale links from working after a fresh request).
		_ = h.services.PasswordResetToken.InvalidateActiveForUser(c.Request.Context(), user.ID)
		if err := h.services.PasswordResetToken.Create(c.Request.Context(), &modelsAuth.PasswordResetToken{
			ID:             crypto.GenerateID(),
			UserID:         user.ID,
			TokenLookupKey: crypto.DeriveResetTokenLookupKey(token),
			TokenHash:      tokenHash,
			ExpiresAt:      expires,
			CreatedAt:      time.Now(),
		}); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "request_processing_error")
			return
		}
		// Clear legacy User.ResetToken so old links can't be reused. Best-effort:
		// failure is logged but does not block the new-token flow.
		user.ResetToken = nil
		user.ResetTokenExpiresAt = nil
		if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
			log.Printf("auth: clear legacy reset token failed for %s: %v", user.ID, err)
		}
	} else {
		// Fallback: legacy User.ResetToken column when the new repo isn't wired.
		legacyToken := crypto.FormatResetToken(user.ID, token)
		legacyHash, lerr := crypto.HashResetToken(legacyToken)
		if lerr != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "request_processing_error")
			return
		}
		user.ResetToken = &legacyHash
		user.ResetTokenExpiresAt = &expires
		if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "request_processing_error")
			return
		}
		token = legacyToken
	}

	// 异步发送密码重置邮件
	resetURL := h.frontendURL() + "/auth/reset-password?token=" + token
	go func(email, firstName, link string) {
		if h.cfg == nil {
			return
		}
		emailSvc := emailsvc.NewEmailService(h.cfg.Email)
		subject := "Reset your CandyPro OEM password"
		body := "Dear " + firstName + ",\n\nClick the link below to reset your password:\n\n" + link + "\n\nThis link expires in 1 hour.\n\nIf you did not request this, please ignore this email.\n\nBest regards,\nCandyPro OEM Team"
		if e := emailSvc.SendEmail(context.Background(), email, subject, body); e != nil {
			log.Printf("auth: reset email failed for %s: %v", email, e)
		}
	}(user.Email, user.FirstName, resetURL)

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(c, "errors.forgot_password_email_sent"),
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
		response.ServiceUnavailableResp(c)
		return
	}

	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=8"`
	}

	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	// H9: Enforce password strength on reset
	if msg := password.ValidatePasswordStrength(req.NewPassword); msg != "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	// M-22: try the new token table first; fall back to the legacy in-row token
	// for in-flight links from before the migration.
	if h.services.PasswordResetToken != nil {
		lookupKey := crypto.DeriveResetTokenLookupKey(req.Token)
		if rec, lookErr := h.services.PasswordResetToken.FindActiveByLookupKey(c.Request.Context(), lookupKey); lookErr == nil && rec != nil {
			if !crypto.VerifyResetToken(req.Token, rec.TokenHash) {
				response.ErrorResp(c, http.StatusBadRequest, "reset_token_invalid")
				return
			}
			user, uerr := h.services.User.GetByID(c.Request.Context(), rec.UserID)
			if uerr != nil {
				response.ErrorResp(c, http.StatusBadRequest, "reset_token_invalid")
				return
			}
			hash, herr := password.HashPassword(req.NewPassword)
			if herr != nil {
				response.ErrorResp(c, http.StatusInternalServerError, "password_hash_failed")
				return
			}
			user.PasswordHash = hash
			user.ResetToken = nil
			user.ResetTokenExpiresAt = nil
			if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
				response.ErrorResp(c, http.StatusInternalServerError, "password_update_failed")
				return
			}
			if err := h.services.PasswordResetToken.MarkUsed(c.Request.Context(), rec.ID); err != nil {
				log.Printf("auth: mark reset token used failed: %v", err)
			}
			// Defence-in-depth: invalidate any other active tokens for this user.
			_ = h.services.PasswordResetToken.InvalidateActiveForUser(c.Request.Context(), user.ID)
			c.JSON(http.StatusOK, gin.H{"message": "Password has been successfully reset"})
			return
		}
		// fall through to legacy path below
	}

	// Legacy path (pre-M-22): token format `<userID>.<random>` stored in user row.
	userID, ok := crypto.ParseResetTokenUserID(req.Token)
	if !ok {
		response.ErrorResp(c, http.StatusBadRequest, "reset_token_invalid")
		return
	}
	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "reset_token_invalid")
		return
	}
	if user.ResetToken == nil || !crypto.VerifyResetToken(req.Token, *user.ResetToken) {
		response.ErrorResp(c, http.StatusBadRequest, "reset_token_invalid")
		return
	}

	if user.ResetTokenExpiresAt == nil || user.ResetTokenExpiresAt.Before(time.Now()) {
		response.ErrorResp(c, http.StatusBadRequest, "reset_token_expired")
		return
	}

	hash, err := password.HashPassword(req.NewPassword)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "password_hash_failed")
		return
	}

	user.PasswordHash = hash
	user.ResetToken = nil
	user.ResetTokenExpiresAt = nil

	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "password_update_failed")
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
		response.ServiceUnavailableResp(c)
		return
	}

	var req struct {
		Token string `json:"token" binding:"required"`
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "token_required")
		return
	}

	claims, err := jwtutil.ValidateJWT(req.Token, h.cfg.JWT.Secret)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "verification_token_invalid")
		return
	}
	purpose, ok := claims["purpose"].(string)
	if !ok || purpose != "verify_email" {
		response.ErrorResp(c, http.StatusBadRequest, "token_purpose_mismatch")
		return
	}
	userID, ok := claims["sub"].(string)
	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorResp(c, http.StatusBadRequest, "verification_subject_invalid")
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	now := time.Now()
	user.EmailVerified = true
	user.EmailVerifiedAt = &now
	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "verification_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(c, "errors.verification_success_redirect"),
	})
}

// Me returns current user profile.
// @Summary Get current user
// @Tags auth
// @Produce json
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := authContextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "invalid_user_identity")
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	role := modelsAuth.User
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
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := authContextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "invalid_user_identity")
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
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := authContextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "invalid_user_identity")
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=8"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	// H9: Enforce password strength on new password
	if msg := password.ValidatePasswordStrength(req.NewPassword); msg != "" {
		response.InvalidResp(c, "invalid_request")
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

	now := time.Now()
	user.PasswordHash = hash
	user.PasswordChangedAt = &now
	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "password_update_failed")
		return
	}

	// SEC-8: Revoke all refresh tokens after password change
	if h.services.Auth != nil {
		if e := h.services.Auth.RevokeAllUserTokens(c.Request.Context(), userID); e != nil {
			log.Printf("auth: RevokeAllUserTokens failed for %s: %v", userID, e)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated successfully",
	})
}

// ResendVerificationEmail resends the email verification link to the authenticated user.
func (h *Handler) ResendVerificationEmail(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := authContextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsCommon.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not identified",
		})
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsCommon.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	if user.EmailVerified {
		c.JSON(http.StatusConflict, modelsCommon.ErrorResponse{
			Error:   "already_verified",
			Message: "Email is already verified",
		})
		return
	}

	verifyToken, err := jwtutil.GenerateJWT(jwt.MapClaims{
		"sub":     user.ID,
		"purpose": "verify_email",
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}, h.cfg.JWT.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsCommon.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate verification token",
		})
		return
	}

	h.sendVerificationEmailAsync(user, verifyToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent. Please check your inbox.",
	})
}

// ResendVerificationEmailPublic 允许未登录用户通过邮箱重发验证邮件（限流保护）。
//
// R2 C-2: this endpoint must NOT distinguish between "user does not exist",
// "user exists but already verified", and "user exists and is unverified" —
// all branches return the same generic 200 response. Previously a 409
// {"error": "already_verified"} let an attacker enumerate which registered
// emails had completed verification.
func (h *Handler) ResendVerificationEmailPublic(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResponse(c)
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	// genericResponse is intentionally identical for every outcome below so the
	// caller cannot distinguish account states by HTTP status or body.
	genericResponse := func() {
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a verification link has been sent."})
	}

	user, err := h.services.User.GetByEmail(c.Request.Context(), req.Email)
	if err != nil || user == nil || user.EmailVerified {
		genericResponse()
		return
	}

	verifyToken, err := jwtutil.GenerateJWT(jwt.MapClaims{
		"sub":     user.ID,
		"purpose": "verify_email",
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}, h.cfg.JWT.Secret)
	if err != nil {
		// Even on internal error, return the generic response — surface the
		// failure through logs rather than the response code.
		log.Printf("auth: resend-verification token generation failed for %s: %v", user.ID, err)
		genericResponse()
		return
	}

	h.sendVerificationEmailAsync(user, verifyToken)
	genericResponse()
}
