package customer

import (
	"errors"
	"net/http"
	"strings"

	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/storage"

	"github.com/gin-gonic/gin"
)

type customerDirectUploadRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
	SizeBytes   int64  `json:"sizeBytes" binding:"required"`
	Folder      string `json:"folder"`
	IsPublic    bool   `json:"isPublic"`
}

func (h *Handler) CustomerPresignUpload(c *gin.Context) {
	if h.storage == nil || h.cfg == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req customerDirectUploadRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if strings.TrimSpace(req.Folder) == "" {
		req.Folder = "uploads"
	}
	result, err := h.storage.CreatePresignedUpload(c.Request.Context(), storage.PresignUploadOptions{
		Folder:      req.Folder,
		FileName:    req.Filename,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
		IsPublic:    req.IsPublic,
		MaxFileSize: h.cfg.Upload.MaxFileSize,
	}, userID)
	if err != nil {
		writeCustomerDirectUploadError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) CustomerCompletePresignedUpload(c *gin.Context) {
	if h.storage == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	file, err := h.storage.CompletePresignedUpload(c.Request.Context(), c.Param("fileId"), userID, false)
	if err != nil {
		writeCustomerDirectUploadError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"file": file})
}

func (h *Handler) CustomerGetUploadedFile(c *gin.Context) {
	if h.storage == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	file, err := h.storage.GetUploadedFile(c.Request.Context(), c.Param("fileId"), userID, false)
	if err != nil {
		writeCustomerDirectUploadError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"file": file})
}

func writeCustomerDirectUploadError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, storage.ErrDirectUploadUnsupported):
		response.ErrorResp(c, http.StatusNotImplemented, "direct_upload_unsupported")
	case errors.Is(err, storage.ErrMetadataUnavailable):
		response.ServiceUnavailableResp(c)
	case errors.Is(err, storage.ErrForbidden):
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
	case errors.Is(err, storage.ErrUploadSizeMismatch):
		response.InvalidResp(c, "upload_size_mismatch")
	default:
		response.InvalidResp(c, "upload_failed")
	}
}
