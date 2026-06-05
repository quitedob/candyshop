package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	modelsCommon "candypro/api/internal/models/common"
)

type managerTestBackend struct {
	uploadCalled bool
}

func (b *managerTestBackend) Upload(context.Context, io.Reader, UploadOptions) (string, error) {
	b.uploadCalled = true
	return "s3://bucket/uploads/server.pdf", nil
}

func (b *managerTestBackend) Delete(context.Context, string) error { return nil }

func (b *managerTestBackend) GetPresignedURL(context.Context, string, time.Duration) (string, error) {
	return "https://download.example/file", nil
}

func (b *managerTestBackend) PresignUpload(context.Context, PresignUploadOptions, time.Duration) (*PresignedUpload, error) {
	return &PresignedUpload{
		Key:       "uploads/direct.pdf",
		URL:       "https://upload.example/direct",
		Method:    "PUT",
		PublicURL: "s3://bucket/uploads/direct.pdf",
	}, nil
}

func (b *managerTestBackend) Stat(context.Context, string) (*ObjectInfo, error) {
	return &ObjectInfo{
		Key:         "uploads/direct.pdf",
		URL:         "s3://bucket/uploads/direct.pdf",
		ContentType: "application/pdf",
		SizeBytes:   12,
	}, nil
}

func (b *managerTestBackend) KeyFromURL(url string) string {
	return strings.TrimPrefix(url, "s3://bucket/")
}

type managerTestMetadata struct {
	byID  map[string]*modelsCommon.UploadedFile
	byKey map[string]*modelsCommon.UploadedFile
}

func newManagerTestMetadata() *managerTestMetadata {
	return &managerTestMetadata{
		byID:  make(map[string]*modelsCommon.UploadedFile),
		byKey: make(map[string]*modelsCommon.UploadedFile),
	}
}

func (m *managerTestMetadata) Create(_ context.Context, file *modelsCommon.UploadedFile) error {
	m.byID[file.ID] = file
	m.byKey[file.StorageKey] = file
	return nil
}

func (m *managerTestMetadata) Save(_ context.Context, file *modelsCommon.UploadedFile) error {
	m.byID[file.ID] = file
	m.byKey[file.StorageKey] = file
	return nil
}

func (m *managerTestMetadata) FindByID(_ context.Context, id string) (*modelsCommon.UploadedFile, error) {
	file, ok := m.byID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return file, nil
}

func (m *managerTestMetadata) FindByStorageKey(_ context.Context, key string) (*modelsCommon.UploadedFile, error) {
	file, ok := m.byKey[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return file, nil
}

type managerTestScanner struct{}

func (managerTestScanner) Dispatch(context.Context, *modelsCommon.UploadedFile, string) error {
	return nil
}

func TestManagerUploadRequiresMetadata(t *testing.T) {
	backend := &managerTestBackend{}
	manager := NewManager(backend, nil, nil)

	_, err := manager.Upload(context.Background(), bytes.NewBufferString("%PDF"), UploadOptions{FileName: "test.pdf"})
	if !errors.Is(err, ErrMetadataUnavailable) {
		t.Fatalf("Upload() error = %v, want ErrMetadataUnavailable", err)
	}
	if backend.uploadCalled {
		t.Fatal("backend upload was called before metadata availability was checked")
	}
}

func TestManagerDirectUploadBlocksDownloadUntilScanCompletes(t *testing.T) {
	ctx := context.Background()
	manager := NewManager(&managerTestBackend{}, newManagerTestMetadata(), managerTestScanner{})

	direct, err := manager.CreatePresignedUpload(ctx, PresignUploadOptions{
		FileName:    "direct.pdf",
		ContentType: "application/pdf",
		SizeBytes:   12,
	}, "customer-1")
	if err != nil {
		t.Fatalf("CreatePresignedUpload() error = %v", err)
	}

	file, err := manager.CompletePresignedUpload(ctx, direct.File.ID, "customer-1", false)
	if err != nil {
		t.Fatalf("CompletePresignedUpload() error = %v", err)
	}
	if file.ScanStatus != modelsCommon.UploadScanStatusPending {
		t.Fatalf("scan status = %q, want pending", file.ScanStatus)
	}
	if _, err := manager.GetPresignedURL(ctx, file.URL, time.Minute); !errors.Is(err, ErrFileQuarantined) {
		t.Fatalf("GetPresignedURL() error = %v, want ErrFileQuarantined", err)
	}

	if _, err := manager.RecordScanResult(ctx, file.ID, modelsCommon.UploadScanStatusClean, ""); err != nil {
		t.Fatalf("RecordScanResult() error = %v", err)
	}
	if _, err := manager.GetPresignedURL(ctx, file.URL, time.Minute); err != nil {
		t.Fatalf("GetPresignedURL() after clean scan error = %v", err)
	}
}
