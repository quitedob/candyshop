package customer

import (
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/storage"
	"candypro/api/internal/pkg/uploadpath"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CustomerGetCompany(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}
	company, err := h.services.Company.EnsureCompanyForUser(c.Request.Context(), user)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "company_provision_failed")
		return
	}
	if company == nil {
		response.ErrorResp(c, http.StatusNotFound, "no_company_profile")
		return
	}
	c.JSON(http.StatusOK, company)
}

func (h *Handler) CustomerUpdateCompany(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}
	company, err := h.services.Company.EnsureCompanyForUser(c.Request.Context(), user)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "company_provision_failed")
		return
	}
	if company == nil {
		response.ErrorResp(c, http.StatusNotFound, "no_company_profile")
		return
	}
	var req struct {
		Phone   *string                    `json:"phone"`
		Website *string                    `json:"website"`
		Address *modelsUser.CompanyAddress `json:"address"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Phone != nil {
		company.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.Website != nil {
		company.Website = strings.TrimSpace(*req.Website)
	}
	if req.Address != nil {
		company.Address = *req.Address
	}
	company.UpdatedAt = time.Now()
	if err := h.services.Company.UpdateCompany(c.Request.Context(), company); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "company_update_failed")
		return
	}
	c.JSON(http.StatusOK, company)
}

// CustomerUploadKYBDocument allows a customer to upload a KYB document (business license, etc.).
func (h *Handler) CustomerUploadKYBDocument(c *gin.Context) {
	if h.services == nil || h.cfg == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}
	if user.CompanyID == nil {
		response.ErrorResp(c, http.StatusNotFound, "no_company_profile")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.InvalidResp(c, "upload_no_file")
		return
	}
	defer file.Close()

	// Validate file size (max 10MB)
	if header.Size > 10*1024*1024 {
		response.InvalidResp(c, "file_size_exceeded")
		return
	}

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	allowed := map[string]bool{
		"application/pdf": true,
		"image/jpeg":      true,
		"image/png":       true,
	}
	if !allowed[contentType] {
		response.InvalidResp(c, "file_type_not_allowed")
		return
	}

	fileURL, uploadErr := h.storage.Upload(c.Request.Context(), file, storage.UploadOptions{
		Folder:   "kyb",
		FileName: header.Filename,
	})
	if uploadErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "upload_failed")
		return
	}

	// Update company business license URL
	company, compErr := h.services.Company.GetCompany(c.Request.Context(), *user.CompanyID)
	if compErr == nil {
		company.BusinessLicense = fileURL
		company.UpdatedAt = time.Now()
		// SEC-16: Handle the DB update error properly instead of silently ignoring it
		if updateErr := h.services.Company.UpdateCompany(c.Request.Context(), company); updateErr != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "company_update_failed")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "KYB document uploaded successfully",
		"url":     fileURL,
	})
}

// CustomerDownloadKYBDocumentFile returns KYB document file via authenticated stream.
func (h *Handler) CustomerDownloadKYBDocumentFile(c *gin.Context) {
	if h.services == nil || h.cfg == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil || user.CompanyID == nil {
		response.ErrorResp(c, http.StatusNotFound, "no_company_profile")
		return
	}
	company, err := h.services.Company.GetCompany(c.Request.Context(), *user.CompanyID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "company_not_found")
		return
	}
	if strings.TrimSpace(company.BusinessLicense) == "" {
		response.ErrorResp(c, http.StatusNotFound, "no_kyb_document")
		return
	}
	if h.cfg.Upload.StorageDriver == "s3" {
		presigned, err := h.storage.GetPresignedURL(c.Request.Context(), company.BusinessLicense, 15*time.Minute)
		if err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, presigned)
		return
	}
	local, err := uploadpath.LocalPathFromUploadURL(h.cfg.Upload.UploadPath, company.BusinessLicense)
	if err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}
	if _, statErr := os.Stat(local); statErr != nil {
		response.ErrorResp(c, http.StatusNotFound, "file_not_found")
		return
	}
	c.File(local)
}
