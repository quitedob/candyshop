package system

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"candypro/api/internal/config"

	"github.com/gin-gonic/gin"
)

func TestSystemAIConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	enabled := &Handler{cfg: &config.Config{AI: config.AIConfig{OpenAIModel: "gpt-4o", OpenAIAPIKey: "sk-test"}}}
	disabled := &Handler{cfg: &config.Config{AI: config.AIConfig{OpenAIModel: "gpt-4o"}}}
	noCfg := &Handler{cfg: nil}

	r := gin.New()
	r.GET("/ai/config", noCfg.SystemAIConfig) // same handler method, router is shared

	t.Run("enabled", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.GET("/enabled", enabled.SystemAIConfig)
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/enabled", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var body struct {
			Model   string `json:"model"`
			Enabled bool   `json:"enabled"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if body.Model != "gpt-4o" {
			t.Fatalf("expected model gpt-4o, got %q", body.Model)
		}
		if !body.Enabled {
			t.Fatal("expected enabled=true with an API key configured")
		}
	})

	t.Run("disabled", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.GET("/disabled", disabled.SystemAIConfig)
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/disabled", nil))
		var body struct {
			Model   string `json:"model"`
			Enabled bool   `json:"enabled"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if body.Model != "gpt-4o" {
			t.Fatalf("expected model gpt-4o, got %q", body.Model)
		}
		if body.Enabled {
			t.Fatal("expected enabled=false with empty API key")
		}
	})

	t.Run("nil config", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.GET("/nilcfg", noCfg.SystemAIConfig)
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/nilcfg", nil))
		var body struct {
			Model   string `json:"model"`
			Enabled bool   `json:"enabled"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if body.Model != "" || body.Enabled {
			t.Fatalf("expected empty model + disabled, got %+v", body)
		}
	})
}
