package admin

import (
	"candypro/api/internal/storage"
	"candypro/api/internal/utils"
	"fmt"
	"net/http"
	"strings"

	modelsProduct "candypro/api/internal/models/product"

	"github.com/gin-gonic/gin"
)

// Allowed image types
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// Allowed document types
var allowedDocTypes = map[string]bool{
	"application/pdf": true,
}

type uploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
}

// AdminUploadImage handles image uploads
// @Summary Admin upload image
// @Tags admin-upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file"
// @Param folder formData string false "Upload folder (products, certifications, etc.)"
// @Router /admin/upload/image [post]
func (h *Handler) AdminUploadImage(c *gin.Context) {
	if h.cfg == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "No file provided",
		})
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_file_type",
			Message: "Only JPEG, PNG, WebP, and GIF images are allowed",
		})
		return
	}

	// Validate file size (max 5MB)
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "file_too_large",
			Message: "File size must be less than 5MB",
		})
		return
	}

	folder := c.DefaultPostForm("folder", "images")
	url, err := h.storage.Upload(c.Request.Context(), file, storage.UploadOptions{
		Folder:      folder,
		FileName:    header.Filename,
		MaxFileSize: h.cfg.Upload.MaxFileSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to save file",
		})
		return
	}

	c.JSON(http.StatusOK, uploadResponse{
		URL:      url,
		Filename: header.Filename,
		Size:     header.Size,
		Type:     contentType,
	})
}

// AdminUploadDocument handles document uploads (PDFs for certificates, etc.)
// @Summary Admin upload document
// @Tags admin-upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Document file"
// @Param folder formData string false "Upload folder"
// @Router /admin/upload/document [post]
func (h *Handler) AdminUploadDocument(c *gin.Context) {
	if h.cfg == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "No file provided",
		})
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !allowedDocTypes[contentType] {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_file_type",
			Message: "Only PDF documents are allowed",
		})
		return
	}

	// Validate file size (max 10MB)
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "file_too_large",
			Message: "File size must be less than 10MB",
		})
		return
	}

	folder := c.DefaultPostForm("folder", "documents")
	url, err := h.storage.Upload(c.Request.Context(), file, storage.UploadOptions{
		Folder:      folder,
		FileName:    header.Filename,
		MaxFileSize: h.cfg.Upload.MaxFileSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to save file",
		})
		return
	}

	c.JSON(http.StatusOK, uploadResponse{
		URL:      url,
		Filename: header.Filename,
		Size:     header.Size,
		Type:     contentType,
	})
}

// AdminUploadMultiple handles multiple file uploads
// @Summary Admin upload multiple files
// @Tags admin-upload
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Files"
// @Param folder formData string false "Upload folder"
// @Router /admin/upload/multiple [post]
func (h *Handler) AdminUploadMultiple(c *gin.Context) {
	if h.cfg == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "No files provided",
		})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "No files provided",
		})
		return
	}

	folder := c.DefaultPostForm("folder", "images")
	var responses []uploadResponse
	var errors []string

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to open %s", fileHeader.Filename))
			continue
		}

		contentType := fileHeader.Header.Get("Content-Type")
		if !allowedImageTypes[contentType] && !allowedDocTypes[contentType] {
			file.Close()
			errors = append(errors, fmt.Sprintf("Invalid type for %s", fileHeader.Filename))
			continue
		}

		// Max 5MB per file
		if fileHeader.Size > 5*1024*1024 {
			file.Close()
			errors = append(errors, fmt.Sprintf("File too large: %s", fileHeader.Filename))
			continue
		}

		url, err := h.storage.Upload(c.Request.Context(), file, storage.UploadOptions{
			Folder:      folder,
			FileName:    fileHeader.Filename,
			MaxFileSize: h.cfg.Upload.MaxFileSize,
		})
		file.Close()
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to save %s", fileHeader.Filename))
			continue
		}

		responses = append(responses, uploadResponse{
			URL:      url,
			Filename: fileHeader.Filename,
			Size:     fileHeader.Size,
			Type:     contentType,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"files":  responses,
		"errors": errors,
		"total":  len(responses),
	})
}

// AdminDeleteFile handles file deletion
// @Summary Admin delete file
// @Tags admin-upload
// @Produce json
// @Param filename path string true "Filename (relative to /uploads/)"
// @Router /admin/upload/:filename [delete]
func (h *Handler) AdminDeleteFile(c *gin.Context) {
	if h.cfg == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	filename := c.Param("filename")
	if filename == "" {
		filename = c.Query("path")
	}
	if filename == "" {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "File path is required",
		})
		return
	}

	if err := h.storage.Delete(c.Request.Context(), filename); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
				Error:   "not_found",
				Message: "File not found",
			})
			return
		}
		c.JSON(http.StatusForbidden, modelsProduct.ErrorResponse{
			Error:   "forbidden",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
		"path":    filename,
	})
}
