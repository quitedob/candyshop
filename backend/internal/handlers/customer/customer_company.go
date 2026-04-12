package customer

import (
modelsUser "candypro/api/internal/models/user"
"candypro/api/internal/utils"
"fmt"
"io"
"net/http"
"os"
"path/filepath"
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

	// SEC-9: Validate magic bytes to prevent MIME spoofing
	magicBuf := make([]byte, 512)
	n, _ := file.Read(magicBuf)
	detectedType := http.DetectContentType(magicBuf[:n])
	allowedDetected := map[string]bool{
		"application/pdf": true,
		"image/jpeg":      true,
		"image/png":       true,
	}
	if !allowedDetected[detectedType] {
		utils.InvalidRequestResponse(c, "File content does not match declared type")
		return
	}
	// Seek back to start after reading magic bytes
	if seeker, ok := file.(io.Seeker); ok {
		if _, seekErr := seeker.Seek(0, io.SeekStart); seekErr != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to process file")
			return
		}
	}

	// Save file using admin upload path
	uploadPath := h.cfg.Upload.UploadPath
	if uploadPath == "" {
		uploadPath = "./uploads"
	}

	ext := ".pdf"
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	}

	filename := fmt.Sprintf("kyb_%s_%d%s", (*user.CompanyID)[:8], time.Now().Unix(), ext)
	// SEC-10: Use filepath.Join instead of string concatenation
	dir := filepath.Join(uploadPath, "kyb")

	// Validate resolved path is under upload path
	absUploadPath, absUploadErr := filepath.Abs(uploadPath)
	if absUploadErr != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to resolve upload path")
		return
	}
	absDir, absDirErr := filepath.Abs(dir)
	if absDirErr != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to resolve directory path")
		return
	}
	if !strings.HasPrefix(absDir, absUploadPath) {
		utils.ErrorResponse(c, http.StatusForbidden, "forbidden", "Invalid upload path")
		return
	}

	if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to create upload directory")
		return
	}

	dst, createErr := os.Create(filepath.Join(dir, filename))
	if createErr != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to save file")
		return
	}
	defer dst.Close()

	if _, copyErr := io.Copy(dst, file); copyErr != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to write file")
		return
	}

	fileURL := "/uploads/kyb/" + filename

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
