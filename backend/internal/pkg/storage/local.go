package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LocalStorageService stores files on the local filesystem.
type LocalStorageService struct {
	uploadPath string
	uploadURL  string
}

// NewLocalStorageService creates a new local storage service.
func NewLocalStorageService(uploadPath, uploadURL string) *LocalStorageService {
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	if uploadURL == "" {
		uploadURL = "/uploads"
	}
	return &LocalStorageService{uploadPath: uploadPath, uploadURL: uploadURL}
}

// Allowed image content types
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
	"image/svg+xml": true,
}

// Allowed document content types
var allowedDocTypes = map[string]bool{
	"application/pdf": true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
}

// Allowed media content types (for inquiry attachments)
var allowedMediaTypes = map[string]bool{
	"text/plain":        true,
	"video/x-matroska":  true,
	"video/mp4":         true,
	"audio/mpeg":        true,
	"audio/mp3":         true,
}

// Whitelisted folder names
var allowedFolders = map[string]bool{
	"images": true, "documents": true, "products": true,
	"certifications": true, "avatars": true, "uploads": true,
	"payment-proofs": true, "kyb": true,
}

// Whitelisted file extensions
var allowedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true,
	".webp": true, ".gif": true, ".pdf": true,
	".docx": true, ".doc": true, ".txt": true,
	".svg": true, ".mkv": true, ".mp4": true, ".mp3": true,
}

// Upload stores a file on the local filesystem and returns the public URL path.
func (s *LocalStorageService) Upload(ctx context.Context, reader io.Reader, opts UploadOptions) (string, error) {
	cleanFolder := filepath.Base(filepath.Clean(opts.Folder))
	if cleanFolder == "." || cleanFolder == "/" || strings.Contains(cleanFolder, "..") {
		cleanFolder = "uploads"
	}
	if !allowedFolders[cleanFolder] {
		cleanFolder = "uploads"
	}

	ext := strings.ToLower(filepath.Ext(opts.FileName))
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("file extension %s is not allowed", ext)
	}

	// Read first 512 bytes for magic byte detection
	magicBuf := make([]byte, 512)
	n, readErr := reader.Read(magicBuf)
	if readErr != nil && readErr != io.EOF {
		return "", fmt.Errorf("failed to read file: %w", readErr)
	}
	magicBuf = magicBuf[:n]
	detectedType := http.DetectContentType(magicBuf)
	// Strip charset suffix for comparison (e.g. "text/plain; charset=utf-8" → "text/plain")
	if idx := strings.Index(detectedType, ";"); idx != -1 {
		detectedType = strings.TrimSpace(detectedType[:idx])
	}
	if !allowedImageTypes[detectedType] && !allowedDocTypes[detectedType] && !allowedMediaTypes[detectedType] {
		return "", fmt.Errorf("file content type %s is not allowed", detectedType)
	}

	// Reconstruct reader with the consumed magic bytes
	var fileReader io.Reader
	if seeker, ok := reader.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("failed to seek file: %w", err)
		}
		fileReader = reader
	} else {
		fileReader = io.MultiReader(strings.NewReader(string(magicBuf)), reader)
	}

	dir := filepath.Join(s.uploadPath, cleanFolder)
	absUploadPath, err1 := filepath.Abs(s.uploadPath)
	if err1 != nil {
		return "", fmt.Errorf("failed to resolve upload base path: %w", err1)
	}
	absDir, err2 := filepath.Abs(dir)
	if err2 != nil {
		return "", fmt.Errorf("failed to resolve upload directory path: %w", err2)
	}
	if !strings.HasPrefix(absDir, absUploadPath) {
		return "", fmt.Errorf("invalid upload path")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)
	destPath := filepath.Join(dir, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, fileReader); err != nil {
		return "", err
	}

	urlPath := fmt.Sprintf("%s/%s/%s", s.uploadURL, cleanFolder, filename)
	return urlPath, nil
}

// Delete removes a file from the local filesystem by its URL.
func (s *LocalStorageService) Delete(ctx context.Context, url string) error {
	cleanPath := filepath.Clean("/" + strings.TrimPrefix(url, "/"))
	if !strings.HasPrefix(cleanPath, "/uploads/") && !strings.HasPrefix(cleanPath, s.uploadURL+"/") {
		return fmt.Errorf("can only delete files from uploads directory")
	}

	relPath := strings.TrimPrefix(cleanPath, "/uploads/")
	relPath = strings.TrimPrefix(relPath, strings.TrimPrefix(s.uploadURL, "/")+"/")
	fsPath := filepath.Join(s.uploadPath, relPath)

	absUploadPath, err1 := filepath.Abs(s.uploadPath)
	if err1 != nil {
		return fmt.Errorf("failed to resolve upload path: %w", err1)
	}
	absFsPath, err2 := filepath.Abs(fsPath)
	if err2 != nil {
		return fmt.Errorf("failed to resolve file path: %w", err2)
	}
	if !strings.HasPrefix(absFsPath, absUploadPath) {
		return fmt.Errorf("invalid file path")
	}

	if _, err := os.Stat(fsPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}
	return os.Remove(fsPath)
}

// GetPresignedURL returns the public URL directly (local files are always accessible).
func (s *LocalStorageService) GetPresignedURL(ctx context.Context, url string, expiry time.Duration) (string, error) {
	return url, nil
}

// KeyFromURL extracts the storage key from a public URL.
func KeyFromURL(url string) string {
	url = strings.TrimPrefix(url, "/uploads/")
	url = strings.TrimPrefix(url, "/")
	return url
}
