package admin

import (
	"candypro/api/internal/utils"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	modelsProduct "candypro/api/internal/models/product"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	url, err := saveFile(file, header, h.cfg.Upload.UploadPath, folder)
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
	url, err := saveFile(file, header, h.cfg.Upload.UploadPath, folder)
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
			errors = append(errors, fmt.Sprintf("Invalid type for %s", fileHeader.Filename))
			file.Close()
			continue
		}

		// Max 5MB per file
		if fileHeader.Size > 5*1024*1024 {
			errors = append(errors, fmt.Sprintf("File too large: %s", fileHeader.Filename))
			file.Close()
			continue
		}

		url, err := saveFile(file, fileHeader, h.cfg.Upload.UploadPath, folder)
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

func saveFile(file multipart.File, header *multipart.FileHeader, uploadPath, folder string) (string, error) {
	// R4-01: Sanitize folder — strip any path traversal attempts
	cleanFolder := filepath.Base(filepath.Clean(folder))
	if cleanFolder == "." || cleanFolder == "/" || strings.Contains(cleanFolder, "..") {
		cleanFolder = "uploads"
	}
	// Whitelist allowed folder names
	allowedFolders := map[string]bool{
		"images": true, "documents": true, "products": true,
		"certifications": true, "avatars": true, "uploads": true,
	}
	if !allowedFolders[cleanFolder] {
		cleanFolder = "uploads"
	}

	// R4-01: Validate file extension against whitelist
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".webp": true, ".gif": true, ".pdf": true,
	}
	if !allowedExts[ext] {
		return "", fmt.Errorf("file extension %s is not allowed", ext)
	}

	// R4-01: Validate magic bytes (first 8 bytes) to prevent MIME spoofing
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	buf = buf[:n]
	detectedType := http.DetectContentType(buf)
	if !allowedImageTypes[detectedType] && !allowedDocTypes[detectedType] {
		return "", fmt.Errorf("file content type %s is not allowed", detectedType)
	}
	// Seek back to start after reading magic bytes
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("failed to seek file: %w", err)
		}
	}

	// Create upload directory if not exists
	dir := filepath.Join(uploadPath, cleanFolder)
	// SEC-4: Ensure resolved dir is still under uploadPath — fail on path resolution errors
	absUploadPath, absUploadErr := filepath.Abs(uploadPath)
	if absUploadErr != nil {
		return "", fmt.Errorf("failed to resolve upload base path: %w", absUploadErr)
	}
	absDir, absDirErr := filepath.Abs(dir)
	if absDirErr != nil {
		return "", fmt.Errorf("failed to resolve upload directory path: %w", absDirErr)
	}
	if !strings.HasPrefix(absDir, absUploadPath) {
		return "", fmt.Errorf("invalid upload path")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// Generate unique filename — never use user-supplied filename
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)
	destPath := filepath.Join(dir, filename)

	// Create destination file
	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	// Return URL path
	urlPath := fmt.Sprintf("/uploads/%s/%s", cleanFolder, filename)
	return urlPath, nil
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

	// R4-13: Use path param, not query param (route is DELETE /upload/:filename)
	filename := c.Param("filename")
	if filename == "" {
		// Fallback to query param for backward compat
		filename = c.Query("path")
	}
	if filename == "" {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "File path is required",
		})
		return
	}

	// R4-01: Sanitize path — prevent traversal
	cleanPath := filepath.Clean("/" + strings.TrimPrefix(filename, "/"))
	if !strings.HasPrefix(cleanPath, "/uploads/") {
		c.JSON(http.StatusForbidden, modelsProduct.ErrorResponse{
			Error:   "forbidden",
			Message: "Can only delete files from uploads directory",
		})
		return
	}

	// Convert URL path to filesystem path
	fsPath := filepath.Join(h.cfg.Upload.UploadPath, strings.TrimPrefix(cleanPath, "/uploads/"))

	// SEC-4: Ensure resolved path is still under uploadPath — fail on path resolution errors
	absUploadPath, absUploadErr := filepath.Abs(h.cfg.Upload.UploadPath)
	if absUploadErr != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to resolve upload path",
		})
		return
	}
	absFsPath, absFsErr := filepath.Abs(fsPath)
	if absFsErr != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to resolve file path",
		})
		return
	}
	if !strings.HasPrefix(absFsPath, absUploadPath) {
		c.JSON(http.StatusForbidden, modelsProduct.ErrorResponse{
			Error:   "forbidden",
			Message: "Invalid file path",
		})
		return
	}

	// Check if file exists
	if _, err := os.Stat(fsPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "File not found",
		})
		return
	}

	// Delete file
	if err := os.Remove(fsPath); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
		"path":    cleanPath,
	})
}
