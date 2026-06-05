package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	modelsCommon "candypro/api/internal/models/common"
)

type WebhookScanner struct {
	url    string
	secret string
	client *http.Client
}

type scanDispatchPayload struct {
	FileID       string `json:"fileId"`
	StorageKey   string `json:"storageKey"`
	Filename     string `json:"filename"`
	ContentType  string `json:"contentType"`
	SizeBytes    int64  `json:"sizeBytes"`
	DownloadURL  string `json:"downloadUrl"`
	CallbackPath string `json:"callbackPath"`
}

func NewWebhookScanner(url, secret string, client *http.Client) *WebhookScanner {
	if strings.TrimSpace(url) == "" {
		return nil
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &WebhookScanner{url: strings.TrimSpace(url), secret: secret, client: client}
}

func (s *WebhookScanner) Dispatch(ctx context.Context, file *modelsCommon.UploadedFile, downloadURL string) error {
	body, err := json.Marshal(scanDispatchPayload{
		FileID:       file.ID,
		StorageKey:   file.StorageKey,
		Filename:     file.OriginalName,
		ContentType:  file.ContentType,
		SizeBytes:    file.SizeBytes,
		DownloadURL:  downloadURL,
		CallbackPath: "/api/v1/system/uploads/" + file.ID + "/scan-result",
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Virus-Scan-Secret", s.secret)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("scan webhook returned %s", resp.Status)
	}
	return nil
}

var _ ScanDispatcher = (*WebhookScanner)(nil)
