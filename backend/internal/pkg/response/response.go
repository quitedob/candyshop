package response

import (
	"log"
	"net/http"

	modelsCommon "candypro/api/internal/models/common"
	i18nutil "candypro/api/internal/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// ErrorResp writes a translated error response using the error code as the i18n key.
func ErrorResp(c *gin.Context, status int, code string) {
	msg := i18nutil.T(c, "errors."+code)
	c.JSON(status, modelsCommon.ErrorResponse{
		Error:   code,
		Message: msg,
	})
}

// ErrorRespDetail writes a translated error with additional details.
func ErrorRespDetail(c *gin.Context, status int, code string, detail any) {
	msg := i18nutil.T(c, "errors."+code)
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
