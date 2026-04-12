package public

import (
	modelsProduct "candypro/api/internal/models/product"
)

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"candypro/api/internal/utils"

	"github.com/gin-gonic/gin"
)

// ===== Inquiry =====

// SubmitInquiry handles inquiry form submissions
// @Summary Submit inquiry
// @Tags inquiry
// @Accept multipart/form-data
// @Produce json
// @Param companyName formData string true "Company name"
// @Param contactPerson formData string true "Contact person"
// @Param email formData string true "Email address"
// @Param whatsapp formData string false "WhatsApp number"
// @Param targetCountry formData string false "Target country"
// @Param estimatedQuantity formData string false "Estimated quantity"
// @Param interestedProducts formData []string false "Interested products"
// @Param packagingRequirements formData string false "Packaging requirements"
// @Param flavorRequirements formData string false "Flavor requirements"
// @Param oemNeeded formData boolean false "OEM needed"
// @Param expectedDelivery formData string false "Expected delivery"
// @Param message formData string false "Message"
// @Param files formData file false "Attached files"
// @Success 200 {object} map[string]interface{}
// @Router /inquiry [post]
func (h *Handler) SubmitInquiry(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResponse(c)
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to parse form data",
		})
		return
	}

	form := c.Request.MultipartForm

	// Extract and validate required fields
	companyName := getFormValue(form, "companyName")
	contactPerson := getFormValue(form, "contactPerson")
	email := getFormValue(form, "email")

	// Validate required fields
	if companyName == "" || contactPerson == "" || email == "" {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "bad_request",
			Message: "Missing required fields: companyName, contactPerson, and email are required",
		})
		return
	}

	// Validate email format
	emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	if !matched {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid email format",
		})
		return
	}

	// Extract optional fields
	interestedProducts := getFormValues(form, "interestedProducts")
	oemNeeded := getFormValue(form, "oemNeeded") == "true"

	// Handle file uploads
	var files []string
	if fileHeaders, ok := form.File["files"]; ok {
		uploadPath := h.cfg.Upload.UploadPath
		if err := os.MkdirAll(uploadPath, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to create upload directory",
			})
			return
		}

		for _, fileHeader := range fileHeaders {
			// Validate file size
			if fileHeader.Size > h.cfg.Upload.MaxFileSize {
				c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
					Error:   "bad_request",
					Message: fmt.Sprintf("File %s exceeds maximum size of %d bytes", fileHeader.Filename, h.cfg.Upload.MaxFileSize),
				})
				return
			}

			// Validate file type from header
			contentType := fileHeader.Header.Get("Content-Type")
			if !isAllowedType(contentType, h.cfg.Upload.AllowedTypes) {
				c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
					Error:   "bad_request",
					Message: fmt.Sprintf("File type %s is not allowed", contentType),
				})
				return
			}

			// SEC-9: Validate magic bytes to prevent MIME spoofing
			f, fErr := fileHeader.Open()
			if fErr != nil {
				c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
					Error:   "bad_request",
					Message: fmt.Sprintf("Failed to open file %s", fileHeader.Filename),
				})
				return
			}
			magicBuf := make([]byte, 512)
			n, _ := f.Read(magicBuf)
			f.Close()
			detectedType := http.DetectContentType(magicBuf[:n])
			if !isAllowedType(detectedType, h.cfg.Upload.AllowedTypes) {
				c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
					Error:   "bad_request",
					Message: fmt.Sprintf("File %s content does not match declared type", fileHeader.Filename),
				})
				return
			}

			// Generate unique filename using secure random
			ext := filepath.Ext(fileHeader.Filename)
			filename := utils.GenerateUniqueFilename(ext)
			// Sanitize filename to prevent directory traversal
			filename = filepath.Base(filename)
			if filename == "." || filename == ".." || filename == string(filepath.Separator) {
				c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
					Error:   "bad_request",
					Message: "Invalid filename generated",
				})
				return
			}
			filePath := filepath.Join(uploadPath, filename)

			// Save file
			if err := c.SaveUploadedFile(fileHeader, filePath); err != nil {
				c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
					Error:   "internal_error",
					Message: fmt.Sprintf("Failed to save file %s", fileHeader.Filename),
				})
				return
			}

			files = append(files, filePath)
		}
	}

	// Create inquiry record
	inquiry := &modelsProduct.Inquiry{
		ID:                    utils.GenerateID(),
		CompanyName:           companyName,
		ContactPerson:         contactPerson,
		Email:                 email,
		WhatsApp:              getFormValue(form, "whatsapp"),
		TargetCountry:         getFormValue(form, "targetCountry"),
		EstimatedQuantity:     getFormValue(form, "estimatedQuantity"),
		InterestedProducts:    interestedProducts,
		PackagingRequirements: getFormValue(form, "packagingRequirements"),
		FlavorRequirements:    getFormValue(form, "flavorRequirements"),
		OEMNeeded:             oemNeeded,
		ExpectedDelivery:      getFormValue(form, "expectedDelivery"),
		Message:               getFormValue(form, "message"),
		Files:                 files,
		Status:                "pending",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Submit inquiry (saves to database and sends email)
	if err := h.services.Inquiry.SubmitInquiry(c.Request.Context(), inquiry); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to submit inquiry",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Thank you for your inquiry. We will contact you within 24 hours.",
		"inquiryId": inquiry.ID,
	})
}

// ===== Helper Functions =====

// getFormValue gets a single value from multipart form
func getFormValue(form *multipart.Form, key string) string {
	if values, ok := form.Value[key]; ok && len(values) > 0 {
		return strings.TrimSpace(values[0])
	}
	return ""
}

// getFormValues gets multiple values from multipart form
func getFormValues(form *multipart.Form, key string) []string {
	if values, ok := form.Value[key]; ok {
		result := make([]string, 0, len(values))
		for _, v := range values {
			if v = strings.TrimSpace(v); v != "" {
				result = append(result, v)
			}
		}
		return result
	}
	return []string{}
}

// isAllowedType checks if content type is allowed
func isAllowedType(contentType string, allowedTypes []string) bool {
	for _, t := range allowedTypes {
		if t == contentType {
			return true
		}
	}
	return false
}
