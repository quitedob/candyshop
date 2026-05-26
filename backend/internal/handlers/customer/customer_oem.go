package customer

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"
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

	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		h.customerCreateOEMProjectMultipart(c, userIDStr)
		return
	}

	var req customerCreateOEMProjectRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	project := h.buildOEMProject(userIDStr, req.ProductName, req.InquiryID, req.Requirements, req.Notes, nil)
	if err := h.services.OEM.CreateProject(c.Request.Context(), project); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_project_create_failed")
		return
	}
	c.JSON(http.StatusCreated, project)
}

// customerCreateOEMProjectMultipart 支持 multipart 创建 OEM 项目并上传附件。
func (h *Handler) customerCreateOEMProjectMultipart(c *gin.Context, userIDStr string) {
	if err := c.Request.ParseMultipartForm(64 << 20); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "form_parse_failed")
		return
	}
	form := c.Request.MultipartForm
	getVal := func(key string) string {
		if values, ok := form.Value[key]; ok && len(values) > 0 {
			return strings.TrimSpace(values[0])
		}
		return ""
	}

	productName := getVal("productName")
	if productName == "" {
		response.ErrorResp(c, http.StatusBadRequest, "invalid_request")
		return
	}

	var inquiryID *string
	if id := getVal("inquiryId"); id != "" {
		inquiryID = &id
	}

	req := modelsProduct.OEMRequirements{
		Flavor:       getVal("flavor"),
		Shape:        getVal("shape"),
		Packaging:    getVal("packaging"),
		TargetMarket: getVal("targetMarket"),
		MOQ:          parseOEMMOQ(getVal("moq")),
	}
	if certs := getVal("certifications"); certs != "" {
		req.Certifications = splitCSV(certs)
	}
	if raw := getVal("requirements"); raw != "" {
		var parsed modelsProduct.OEMRequirements
		if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
			req = parsed
		}
	}

	attachments, err := h.uploadCustomerAttachments(c, form.File["files"], "documents", 5)
	if err != nil {
		return
	}

	project := h.buildOEMProject(userIDStr, productName, inquiryID, req, getVal("notes"), attachments)
	if err := h.services.OEM.CreateProject(c.Request.Context(), project); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_project_create_failed")
		return
	}
	c.JSON(http.StatusCreated, project)
}

func (h *Handler) buildOEMProject(
	userID, productName string,
	inquiryID *string,
	requirements modelsProduct.OEMRequirements,
	notes string,
	attachments []string,
) *modelsProduct.OEMProject {
	return &modelsProduct.OEMProject{
		ID:           crypto.GenerateID(),
		UserID:       userID,
		InquiryID:    inquiryID,
		ProductName:  strings.TrimSpace(productName),
		Status:       modelsProduct.OEMStatusInquiry,
		CurrentStep:  0,
		Requirements: requirements,
		Notes:        strings.TrimSpace(notes),
		Attachments:  attachments,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseOEMMOQ(raw string) int {
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return n
}

// CustomerUploadOEMProjectAttachment POST /user/oem-projects/:id/attachments
func (h *Handler) CustomerUploadOEMProjectAttachment(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	projectID := c.Param("id")
	project, err := h.services.OEM.GetProject(c.Request.Context(), projectID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "oem_project_not_found")
		return
	}
	if project.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	if err := c.Request.ParseMultipartForm(64 << 20); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "form_parse_failed")
		return
	}
	newFiles, upErr := h.uploadCustomerAttachments(c, c.Request.MultipartForm.File["files"], "documents", 5)
	if upErr != nil {
		return
	}
	project.Attachments = append([]string(project.Attachments), newFiles...)
	project.UpdatedAt = time.Now()
	if err := h.services.OEM.UpdateProject(c.Request.Context(), project); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_project_update_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"files":      newFiles,
		"totalFiles": len(project.Attachments),
	})
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
