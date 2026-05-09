package customer

import (
modelsUser "candypro/api/internal/models/user"
"candypro/api/internal/storage"
"candypro/api/internal/utils"
"net/http"
"os"
"strings"
"time"

"github.com/gin-gonic/gin"
)

func (h *Handler) CustomerGetCompany(c *gin.Context) {
if h.services == nil {
utils.ServiceUnavailableResponse(c)
return
}
userID, ok := contextUserID(c)
if !ok {
utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
return
}
user, err := h.services.User.GetByID(c.Request.Context(), userID)
if err != nil {
utils.ErrorResponse(c, http.StatusNotFound, "not_found", "User not found")
return
}
if user.CompanyID == nil {
utils.ErrorResponse(c, http.StatusNotFound, "not_found", "No company profile found")
return
}
company, err := h.services.Company.GetCompany(c.Request.Context(), *user.CompanyID)
if err != nil {
utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Company not found")
return
}
c.JSON(http.StatusOK, company)
}

func (h *Handler) CustomerUpdateCompany(c *gin.Context) {
if h.services == nil {
utils.ServiceUnavailableResponse(c)
return
}
userID, ok := contextUserID(c)
if !ok {
utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
return
}
user, err := h.services.User.GetByID(c.Request.Context(), userID)
if err != nil {
utils.ErrorResponse(c, http.StatusNotFound, "not_found", "User not found")
return
}
if user.CompanyID == nil {
utils.ErrorResponse(c, http.StatusNotFound, "not_found", "No company profile found")
return
}
company, err := h.services.Company.GetCompany(c.Request.Context(), *user.CompanyID)
if err != nil {
utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Company not found")
return
}
var req struct {
Phone   *string                    `json:"phone"`
Website *string                    `json:"website"`
Address *modelsUser.CompanyAddress `json:"address"`
}
if !utils.BindJSONOrInvalidRequest(c, &req) {
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
utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update company")
return
}
c.JSON(http.StatusOK, company)
}

// CustomerUploadKYBDocument allows a customer to upload a KYB document (business license, etc.).
func (h *Handler) CustomerUploadKYBDocument(c *gin.Context) {
	if h.services == nil || h.cfg == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "User not found")
		return
	}
	if user.CompanyID == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "No company profile found")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.InvalidRequestResponse(c, "No file provided")
		return
	}
	defer file.Close()

	// Validate file size (max 10MB)
	if header.Size > 10*1024*1024 {
		utils.InvalidRequestResponse(c, "File size must be less than 10MB")
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
		utils.InvalidRequestResponse(c, "Only PDF, JPEG, and PNG files are allowed")
		return
	}

	fileURL, uploadErr := h.storage.Upload(c.Request.Context(), file, storage.UploadOptions{
		Folder:   "kyb",
		FileName: header.Filename,
	})
	if uploadErr != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to save file")
		return
	}

	// Update company business license URL
	company, compErr := h.services.Company.GetCompany(c.Request.Context(), *user.CompanyID)
	if compErr == nil {
		company.BusinessLicense = fileURL
		company.UpdatedAt = time.Now()
		// SEC-16: Handle the DB update error properly instead of silently ignoring it
		if updateErr := h.services.Company.UpdateCompany(c.Request.Context(), company); updateErr != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "File uploaded but failed to update company record")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "KYB document uploaded successfully",
		"url":     fileURL,
	})
}

// CustomerDownloadKYBDocumentFile 认证用户下载本公司 KYB 证照文件（禁止直链 /uploads/kyb）。
func (h *Handler) CustomerDownloadKYBDocumentFile(c *gin.Context) {
	if h.services == nil || h.cfg == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
		return
	}
	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil || user.CompanyID == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "No company profile found")
		return
	}
	company, err := h.services.Company.GetCompany(c.Request.Context(), *user.CompanyID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Company not found")
		return
	}
	if strings.TrimSpace(company.BusinessLicense) == "" {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "No KYB document on file")
		return
	}
	if h.cfg.Upload.StorageDriver == "s3" {
		presigned, err := h.storage.GetPresignedURL(c.Request.Context(), company.BusinessLicense, 15*time.Minute)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to generate download link")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, presigned)
		return
	}
	local, err := utils.LocalPathFromUploadURL(h.cfg.Upload.UploadPath, company.BusinessLicense)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "Invalid document path")
		return
	}
	if _, statErr := os.Stat(local); statErr != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "File not found on server")
		return
	}
	c.File(local)
}
