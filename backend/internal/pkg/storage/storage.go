package storage

import (
	"context"
	"io"
	"time"
)

// UploadOptions carries metadata for an upload.
type UploadOptions struct {
	Folder      string // subfolder / key prefix
	FileName    string // original filename for extension detection
	IsPublic    bool   // whether the file should be publicly accessible
	MaxFileSize int64  // per-file size limit (0 = no override)
}

// StorageService abstracts file storage operations.
type StorageService interface {
	// Upload stores a file and returns its public URL.
	Upload(ctx context.Context, reader io.Reader, opts UploadOptions) (url string, err error)
	// Delete removes a file by its URL or storage key.
	Delete(ctx context.Context, url string) error
	// GetPresignedURL returns a time-limited download URL for a stored file.
	GetPresignedURL(ctx context.Context, url string, expiry time.Duration) (string, error)
}
