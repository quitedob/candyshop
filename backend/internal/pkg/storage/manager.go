package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	modelsCommon "candypro/api/internal/models/common"

	"github.com/google/uuid"
)

const (
	directUploadExpiry = 15 * time.Minute
	scanDownloadExpiry = 15 * time.Minute
)

type metadataRepository interface {
	Create(ctx context.Context, file *modelsCommon.UploadedFile) error
	Save(ctx context.Context, file *modelsCommon.UploadedFile) error
	FindByID(ctx context.Context, id string) (*modelsCommon.UploadedFile, error)
	FindByStorageKey(ctx context.Context, storageKey string) (*modelsCommon.UploadedFile, error)
}

type ScanDispatcher interface {
	Dispatch(ctx context.Context, file *modelsCommon.UploadedFile, downloadURL string) error
}

// DirectUpload pairs the tracked file record with its physical upload instruction.
type DirectUpload struct {
	File   *modelsCommon.UploadedFile `json:"file"`
	Upload *PresignedUpload           `json:"upload"`
}

// Manager tracks uploaded objects and dispatches asynchronous virus scans.
type Manager struct {
	backend  StorageService
	metadata metadataRepository
	scanner  ScanDispatcher
}

func NewManager(backend StorageService, metadata metadataRepository, scanner ScanDispatcher) *Manager {
	return &Manager{backend: backend, metadata: metadata, scanner: scanner}
}

func (m *Manager) Upload(ctx context.Context, reader io.Reader, opts UploadOptions) (string, error) {
	if m.metadata == nil {
		return "", ErrMetadataUnavailable
	}
	url, err := m.backend.Upload(ctx, reader, opts)
	if err != nil {
		return "", err
	}

	now := time.Now()
	file := &modelsCommon.UploadedFile{
		ID:           uuid.NewString(),
		StorageKey:   m.backend.KeyFromURL(url),
		URL:          url,
		OriginalName: opts.FileName,
		ContentType:  opts.ContentType,
		SizeBytes:    opts.SizeBytes,
		Folder:       opts.Folder,
		OwnerID:      opts.OwnerID,
		Visibility:   visibility(opts.IsPublic),
		Status:       modelsCommon.UploadedFileStatusUploaded,
		ScanStatus:   m.initialScanStatus(),
		UploadedAt:   &now,
	}
	if info, statErr := m.backend.Stat(ctx, url); statErr == nil {
		file.SizeBytes = info.SizeBytes
		if strings.TrimSpace(info.ContentType) != "" {
			file.ContentType = info.ContentType
		}
	}
	if err := m.metadata.Create(ctx, file); err != nil {
		if deleteErr := m.backend.Delete(ctx, url); deleteErr != nil {
			log.Printf("storage: metadata create failed and object cleanup also failed for %q: %v", url, deleteErr)
		}
		return "", fmt.Errorf("create upload metadata: %w", err)
	}
	m.queueScan(file)
	return url, nil
}

func (m *Manager) Delete(ctx context.Context, url string) error {
	if err := m.backend.Delete(ctx, url); err != nil {
		return err
	}
	if m.metadata == nil {
		return nil
	}
	file, err := m.metadata.FindByStorageKey(ctx, m.backend.KeyFromURL(url))
	if err != nil {
		return nil
	}
	file.Status = modelsCommon.UploadedFileStatusDeleted
	return m.metadata.Save(ctx, file)
}

func (m *Manager) GetPresignedURL(ctx context.Context, url string, expiry time.Duration) (string, error) {
	if m.metadata != nil {
		file, err := m.metadata.FindByStorageKey(ctx, m.backend.KeyFromURL(url))
		if err == nil && scanBlocksDownload(file.ScanStatus) {
			return "", ErrFileQuarantined
		}
	}
	return m.backend.GetPresignedURL(ctx, url, expiry)
}

func (m *Manager) PresignUpload(ctx context.Context, opts PresignUploadOptions, expiry time.Duration) (*PresignedUpload, error) {
	return m.backend.PresignUpload(ctx, opts, expiry)
}

func (m *Manager) Stat(ctx context.Context, url string) (*ObjectInfo, error) {
	return m.backend.Stat(ctx, url)
}

func (m *Manager) KeyFromURL(url string) string {
	return m.backend.KeyFromURL(url)
}

func (m *Manager) CreatePresignedUpload(ctx context.Context, opts PresignUploadOptions, ownerID string) (*DirectUpload, error) {
	if m.metadata == nil {
		return nil, ErrMetadataUnavailable
	}
	presigned, err := m.backend.PresignUpload(ctx, opts, directUploadExpiry)
	if err != nil {
		return nil, err
	}
	file := &modelsCommon.UploadedFile{
		ID:           uuid.NewString(),
		StorageKey:   presigned.Key,
		URL:          presigned.PublicURL,
		OriginalName: opts.FileName,
		ContentType:  opts.ContentType,
		SizeBytes:    opts.SizeBytes,
		Folder:       opts.Folder,
		OwnerID:      ownerID,
		Visibility:   visibility(opts.IsPublic),
		Status:       modelsCommon.UploadedFileStatusPending,
		ScanStatus:   modelsCommon.UploadScanStatusNotConfigured,
	}
	if err := m.metadata.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("create upload metadata: %w", err)
	}
	return &DirectUpload{File: file, Upload: presigned}, nil
}

func (m *Manager) CompletePresignedUpload(ctx context.Context, id, ownerID string, allowAnyOwner bool) (*modelsCommon.UploadedFile, error) {
	if m.metadata == nil {
		return nil, ErrMetadataUnavailable
	}
	file, err := m.metadata.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := authorizeFile(file, ownerID, allowAnyOwner); err != nil {
		return nil, err
	}
	if file.Status != modelsCommon.UploadedFileStatusPending {
		return file, nil
	}

	info, err := m.backend.Stat(ctx, file.URL)
	if err != nil {
		return nil, err
	}
	if file.SizeBytes > 0 && info.SizeBytes != file.SizeBytes {
		return nil, ErrUploadSizeMismatch
	}
	now := time.Now()
	file.Status = modelsCommon.UploadedFileStatusUploaded
	file.UploadedAt = &now
	file.SizeBytes = info.SizeBytes
	if strings.TrimSpace(info.ContentType) != "" {
		file.ContentType = info.ContentType
	}
	file.ScanStatus = m.initialScanStatus()
	if err := m.metadata.Save(ctx, file); err != nil {
		return nil, err
	}
	m.queueScan(file)
	return file, nil
}

func (m *Manager) GetUploadedFile(ctx context.Context, id, ownerID string, allowAnyOwner bool) (*modelsCommon.UploadedFile, error) {
	if m.metadata == nil {
		return nil, ErrMetadataUnavailable
	}
	file, err := m.metadata.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := authorizeFile(file, ownerID, allowAnyOwner); err != nil {
		return nil, err
	}
	return file, nil
}

func (m *Manager) RecordScanResult(ctx context.Context, id, status, message string) (*modelsCommon.UploadedFile, error) {
	if m.metadata == nil {
		return nil, ErrMetadataUnavailable
	}
	switch status {
	case modelsCommon.UploadScanStatusClean, modelsCommon.UploadScanStatusInfected, modelsCommon.UploadScanStatusError:
	default:
		return nil, fmt.Errorf("invalid scan status %q", status)
	}
	file, err := m.metadata.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	file.ScanStatus = status
	file.ScanMessage = strings.TrimSpace(message)
	file.ScannedAt = &now
	if err := m.metadata.Save(ctx, file); err != nil {
		return nil, err
	}
	return file, nil
}

func (m *Manager) initialScanStatus() string {
	if m.scanner == nil {
		return modelsCommon.UploadScanStatusNotConfigured
	}
	return modelsCommon.UploadScanStatusPending
}

func (m *Manager) queueScan(file *modelsCommon.UploadedFile) {
	if m.scanner == nil || file == nil {
		return
	}
	fileCopy := *file
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		downloadURL, err := m.backend.GetPresignedURL(ctx, fileCopy.URL, scanDownloadExpiry)
		if err == nil {
			err = m.scanner.Dispatch(ctx, &fileCopy, downloadURL)
		}
		if err == nil {
			return
		}
		log.Printf("storage: failed to dispatch virus scan for upload %s: %v", fileCopy.ID, err)
		current, findErr := m.metadata.FindByID(ctx, fileCopy.ID)
		if findErr != nil {
			return
		}
		now := time.Now()
		current.ScanStatus = modelsCommon.UploadScanStatusError
		current.ScanMessage = err.Error()
		current.ScannedAt = &now
		_ = m.metadata.Save(ctx, current)
	}()
}

func authorizeFile(file *modelsCommon.UploadedFile, ownerID string, allowAnyOwner bool) error {
	if allowAnyOwner {
		return nil
	}
	if file == nil || file.OwnerID == "" || file.OwnerID != ownerID {
		return ErrForbidden
	}
	return nil
}

func visibility(isPublic bool) string {
	if isPublic {
		return "public"
	}
	return "private"
}

func scanBlocksDownload(status string) bool {
	switch status {
	case modelsCommon.UploadScanStatusPending,
		modelsCommon.UploadScanStatusInfected,
		modelsCommon.UploadScanStatusError:
		return true
	default:
		return false
	}
}

var _ StorageService = (*Manager)(nil)
