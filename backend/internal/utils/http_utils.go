package utils

import (
	modelsCommon "candypro/api/internal/models/common"
	"candypro/api/internal/i18n"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// T returns the translated string for the given i18n key using the locale from gin.Context.
func T(c *gin.Context, key string) string {
	locale, _ := c.Get("locale")
	if loc, ok := locale.(string); ok && loc != "" {
		return i18n.Translate(loc, key)
	}
	return i18n.Translate(i18n.DefaultLocale(), key)
}

// TWithVars translates with variable substitution.
func TWithVars(c *gin.Context, key string, vars map[string]string) string {
	locale, _ := c.Get("locale")
	if loc, ok := locale.(string); ok && loc != "" {
		return i18n.TranslateWithVars(loc, key, vars)
	}
	return i18n.TranslateWithVars(i18n.DefaultLocale(), key, vars)
}

// ErrorResp writes a translated error response using the error code as the i18n key.
func ErrorResp(c *gin.Context, status int, code string) {
	msg := T(c, "errors."+code)
	c.JSON(status, modelsCommon.ErrorResponse{
		Error:   code,
		Message: msg,
	})
}

// ErrorRespDetail writes a translated error with additional details.
func ErrorRespDetail(c *gin.Context, status int, code string, detail any) {
	msg := T(c, "errors."+code)
	c.JSON(status, modelsCommon.ErrorResponse{
		Error:   code,
		Message: msg,
		Details: detail,
	})
}

// InvalidResp writes a standard 400 translated error.
func InvalidResp(c *gin.Context, code string) {
	ErrorResp(c, http.StatusBadRequest, code)
}

// ServiceUnavailableResp writes a translated 503 response.
func ServiceUnavailableResp(c *gin.Context) {
	ErrorResp(c, http.StatusServiceUnavailable, "service_unavailable")
}

// BindJSONOrInvalid binds JSON and writes a translated 400 on failure.
func BindJSONOrInvalid(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		log.Printf("bind error: %v", err)
		ErrorResp(c, http.StatusBadRequest, "invalid_request")
		return false
	}
	return true
}

// --- Legacy helpers (deprecated, use i18n-aware versions above) ---

// ServiceUnavailableResponse returns a standard service unavailable response.
// Deprecated: Use ServiceUnavailableResp instead.
func ServiceUnavailableResponse(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, modelsCommon.ErrorResponse{
		Error:   "service_unavailable",
		Message: "Service temporarily unavailable. Please try again later.",
	})
}

// ErrorResponse writes a standard API error payload.
// Deprecated: Use ErrorResp instead.
func ErrorResponse(c *gin.Context, status int, code, message string) {
	c.JSON(status, modelsCommon.ErrorResponse{
		Error:   code,
		Message: message,
	})
}

// InvalidRequestResponse writes a standard 400 payload.
// Deprecated: Use InvalidResp instead.
func InvalidRequestResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Invalid request payload"
	}
	ErrorResponse(c, http.StatusBadRequest, "invalid_request", message)
}

// BindJSONOrInvalidRequest binds JSON and writes standard 400 response on failure.
// Deprecated: Use BindJSONOrInvalid instead.
func BindJSONOrInvalidRequest(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		log.Printf("bind error: %v", err)
		InvalidRequestResponse(c, "Invalid request payload")
		return false
	}
	return true
}

// ParsePagination extracts page and limit from query parameters with defaults
// and range clamping.
func ParsePagination(c *gin.Context, defaultLimit, maxLimit int) (page, limit int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err = strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit < 1 || limit > maxLimit {
		limit = defaultLimit
	}
	return page, limit
}

// BuildPagination builds a consistent pagination payload for frontend consumers.
func BuildPagination(total int64, page, limit int) gin.H {
	totalPages := 0
	if limit > 0 && total > 0 {
		totalPages = int(total) / limit
		if int(total)%limit != 0 {
			totalPages++
		}
	}

	return gin.H{
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	}
}
