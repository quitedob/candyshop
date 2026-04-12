package utils

import (
	modelsProduct "candypro/api/internal/models/product"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ServiceUnavailableResponse returns a standard service unavailable response
func ServiceUnavailableResponse(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, modelsProduct.ErrorResponse{
		Error:   "service_unavailable",
		Message: "Service temporarily unavailable. Please try again later.",
	})
}

// ErrorResponse writes a standard API error payload.
func ErrorResponse(c *gin.Context, status int, code, message string) {
	c.JSON(status, modelsProduct.ErrorResponse{
		Error:   code,
		Message: message,
	})
}

// InvalidRequestResponse writes a standard 400 payload.
func InvalidRequestResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Invalid request payload"
	}
	ErrorResponse(c, http.StatusBadRequest, "invalid_request", message)
}

// BindJSONOrInvalidRequest binds JSON and writes standard 400 response on failure.
func BindJSONOrInvalidRequest(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		InvalidRequestResponse(c, err.Error())
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
