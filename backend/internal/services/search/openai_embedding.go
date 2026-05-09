package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// openAIEmbed 调用 OpenAI Embeddings API，供混合语义检索使用
func openAIEmbed(ctx context.Context, apiKey, model, text string) ([]float32, error) {
	text = strings.TrimSpace(text)
	if text == "" || strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("missing embedding input")
	}
	if strings.TrimSpace(model) == "" {
		model = "text-embedding-3-small"
	}
	payload, err := json.Marshal(map[string]interface{}{
		"model": model,
		"input": text,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		msg := resp.Status
		if out.Error != nil && out.Error.Message != "" {
			msg = out.Error.Message
		}
		return nil, fmt.Errorf("embeddings API: %s", msg)
	}
	if len(out.Data) == 0 || len(out.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return out.Data[0].Embedding, nil
}
