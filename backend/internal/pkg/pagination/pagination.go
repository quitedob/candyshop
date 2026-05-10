package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

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
