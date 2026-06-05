package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	ErrDirectUploadUnsupported = errors.New("direct upload is not supported by this storage driver")
	ErrMetadataUnavailable     = errors.New("upload metadata repository is unavailable")
	ErrForbidden               = errors.New("forbidden")
	ErrUploadSizeMismatch      = errors.New("uploaded object size does not match the presigned request")
	ErrFileQuarantined         = errors.New("uploaded file is quarantined")
)

// UploadOptions carries metadata for an upload.
type UploadOptions struct {
	Folder      string // subfolder / key prefix
	FileName    string // original filename for extension detection
	IsPublic    bool   // whether the file should be publicly accessible
	MaxFileSize int64  // per-file size limit (0 = no override)
	ContentType string // declared MIME type when known
	SizeBytes   int64  // declared size when known
	OwnerID     string // authenticated uploader when available
}

// PresignUploadOptions describes a direct browser-to-object-storage upload.
type PresignUploadOptions struct {
	Folder      string
	FileName    string
	ContentType string
	SizeBytes   int64
	IsPublic    bool
	MaxFileSize int64
}

// PresignedUpload is the physical upload instruction returned by a driver.
type PresignedUpload struct {
	Key       string            `json:"key"`
	URL       string            `json:"uploadUrl"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	PublicURL string            `json:"url"`
	ExpiresAt time.Time         `json:"expiresAt"`
}

// ObjectInfo is the storage driver's view of an uploaded object.
type ObjectInfo struct {
	Key         string
	URL         string
	ContentType string
	SizeBytes   int64
}

// StorageService abstracts file storage operations.
type StorageService interface {
	// Upload stores a file and returns its public URL.
	Upload(ctx context.Context, reader io.Reader, opts UploadOptions) (url string, err error)
	// Delete removes a file by its URL or storage key.
	Delete(ctx context.Context, url string) error
	// GetPresignedURL returns a time-limited download URL for a stored file.
	GetPresignedURL(ctx context.Context, url string, expiry time.Duration) (string, error)
	// PresignUpload creates a time-limited direct PUT instruction.
	PresignUpload(ctx context.Context, opts PresignUploadOptions, expiry time.Duration) (*PresignedUpload, error)
	// Stat verifies that an object exists and returns its current metadata.
	Stat(ctx context.Context, url string) (*ObjectInfo, error)
	// KeyFromURL extracts the storage key from a URL produced by this driver.
	KeyFromURL(url string) string
}
