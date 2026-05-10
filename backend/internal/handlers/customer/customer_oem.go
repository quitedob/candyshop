package customer

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CustomerGetOEMProjects(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, _ := c.Get("userID")
	userIDStr := ""
	if id, ok := userID.(string); ok {
		userIDStr = id
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	projects, total, err := h.services.OEM.GetUserProjects(c.Request.Context(), userIDStr, page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_fetch_failed")
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, modelsProduct.PaginatedResponse{
		Data: projects,
		Pagination: modelsProduct.Pagination{
			Total: int(total), Page: page, Limit: limit, TotalPages: totalPages,
		},
	})
}

type customerCreateOEMProjectRequest struct {
	ProductName  string                        `json:"productName" binding:"required"`
	InquiryID    *string                       `json:"inquiryId"`
	Requirements modelsProduct.OEMRequirements `json:"requirements"`
	Notes        string                        `json:"notes"`
}

func (h *Handler) CustomerCreateOEMProject(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, _ := c.Get("userID")
	userIDStr := ""
	if id, ok := userID.(string); ok {
		userIDStr = id
	}
	var req customerCreateOEMProjectRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	project := &modelsProduct.OEMProject{
		ID:           crypto.GenerateID(),
		UserID:       userIDStr,
		InquiryID:    req.InquiryID,
		ProductName:  strings.TrimSpace(req.ProductName),
		Status:       "inquiry",
		CurrentStep:  0,
		Requirements: req.Requirements,
		Notes:        strings.TrimSpace(req.Notes),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := h.services.OEM.CreateProject(c.Request.Context(), project); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_project_create_failed")
		return
	}
	c.JSON(http.StatusCreated, project)
}

func (h *Handler) CustomerGetOEMProject(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	project, err := h.services.OEM.GetProject(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "oem_project_not_found")
		return
	}
	userID, _ := c.Get("userID")
	userIDStr := ""
	if uid, ok := userID.(string); ok {
		userIDStr = uid
	}
	if project.UserID != userIDStr {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	c.JSON(http.StatusOK, project)
}
