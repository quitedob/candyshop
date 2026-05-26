package customer

import (
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/storage"
	"candypro/api/internal/pkg/upload"
	"errors"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

var errAttachmentUploadAborted = errors.New("attachment upload aborted")

const maxCustomerMessageFiles = 3

// uploadCustomerAttachments 上传客户侧通用附件（询价/OEM/订单消息等）。
func (h *Handler) uploadCustomerAttachments(
	c *gin.Context,
	fileHeaders []*multipart.FileHeader,
	folder string,
	maxFiles int,
) ([]string, error) {
	if len(fileHeaders) == 0 {
		return nil, nil
	}
	if maxFiles <= 0 {
		maxFiles = 5
	}
	if len(fileHeaders) > maxFiles {
		response.ErrorResp(c, http.StatusBadRequest, "file_count_exceeded")
		return nil, errAttachmentUploadAborted
	}

	var urls []string
	for _, fh := range fileHeaders {
		if fh.Size > h.cfg.Upload.MaxFileSize {
			response.ErrorRespDetail(c, http.StatusBadRequest, "file_size_exceeded", gin.H{
				"filename": fh.Filename, "maxBytes": h.cfg.Upload.MaxFileSize,
			})
			return nil, errAttachmentUploadAborted
		}
		contentType := fh.Header.Get("Content-Type")
		if !upload.IsAllowedInquiryAttachment(contentType, fh.Filename, h.cfg.Upload.AllowedTypes) {
			response.ErrorRespDetail(c, http.StatusBadRequest, "file_type_not_allowed", gin.H{
				"contentType": contentType, "filename": fh.Filename,
			})
			return nil, errAttachmentUploadAborted
		}
		f, fErr := fh.Open()
		if fErr != nil {
			response.ErrorRespDetail(c, http.StatusBadRequest, "file_open_failed", gin.H{"filename": fh.Filename})
			return nil, errAttachmentUploadAborted
		}
		url, uploadErr := h.storage.Upload(c.Request.Context(), f, storage.UploadOptions{
			Folder:   folder,
			FileName: fh.Filename,
		})
		f.Close()
		if uploadErr != nil {
			log.Printf("WARN: failed to store attachment %s: %v", fh.Filename, uploadErr)
			response.ErrorRespDetail(c, http.StatusInternalServerError, "file_save_failed", gin.H{"filename": fh.Filename})
			return nil, errAttachmentUploadAborted
		}
		urls = append(urls, url)
	}
	return urls, nil
}
