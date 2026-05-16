package admin

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ── Webhook Config CRUD ──

// AdminListWebhooks returns all webhook configurations.
func (h *Handler) AdminListWebhooks(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	configs, err := h.services.Webhook.List(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "webhook_list_failed")
		return
	}

	if configs == nil {
		configs = []modelsOrder.WebhookConfig{}
	}
	c.JSON(http.StatusOK, configs)
}

type webhookCreateRequest struct {
	Name   string   `json:"name" binding:"required"`
	URL    string   `json:"url" binding:"required"`
	Events []string `json:"events" binding:"required,min=1"`
}

// AdminCreateWebhook creates a new webhook configuration.
func (h *Handler) AdminCreateWebhook(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req webhookCreateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	validEvents := []string{
		modelsOrder.WebhookEventOrderCreated,
		modelsOrder.WebhookEventOrderConfirmed,
		modelsOrder.WebhookEventOrderShipped,
		modelsOrder.WebhookEventOrderDelivered,
		modelsOrder.WebhookEventPaymentConfirmed,
		modelsOrder.WebhookEventReturnCreated,
	}
	for _, e := range req.Events {
		if !slices.Contains(validEvents, e) {
			response.InvalidResp(c, "invalid_webhook_event")
			return
		}
	}

	secretBytes := make([]byte, 32)
	_, _ = rand.Read(secretBytes)

	cfg := &modelsOrder.WebhookConfig{
		Name:      strings.TrimSpace(req.Name),
		URL:       strings.TrimSpace(req.URL),
		Secret:    hex.EncodeToString(secretBytes),
		Events:    req.Events,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.services.Webhook.Create(c.Request.Context(), cfg); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "webhook_create_failed")
		return
	}

	c.JSON(http.StatusCreated, cfg)
}

type webhookUpdateRequest struct {
	Name   *string   `json:"name"`
	URL    *string   `json:"url"`
	Events *[]string `json:"events"`
	Status *string   `json:"status"`
}

// AdminUpdateWebhook updates a webhook configuration.
func (h *Handler) AdminUpdateWebhook(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_webhook_id")
		return
	}

	cfg, err := h.services.Webhook.Get(c.Request.Context(), uint(id))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "webhook_not_found")
		return
	}

	var req webhookUpdateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	if req.Name != nil {
		cfg.Name = strings.TrimSpace(*req.Name)
	}
	if req.URL != nil {
		cfg.URL = strings.TrimSpace(*req.URL)
	}
	if req.Events != nil {
		cfg.Events = *req.Events
	}
	if req.Status != nil {
		status := strings.TrimSpace(*req.Status)
		if status != "active" && status != "inactive" {
			response.InvalidResp(c, "invalid_webhook_status")
			return
		}
		cfg.Status = status
	}

	if err := h.services.Webhook.Update(c.Request.Context(), cfg); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "webhook_update_failed")
		return
	}

	c.JSON(http.StatusOK, cfg)
}

// AdminDeleteWebhook deletes a webhook configuration.
func (h *Handler) AdminDeleteWebhook(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_webhook_id")
		return
	}

	if err := h.services.Webhook.Delete(c.Request.Context(), uint(id)); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "webhook_delete_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// AdminListWebhookDeliveries returns delivery logs for a webhook.
func (h *Handler) AdminListWebhookDeliveries(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	webhookID, _ := strconv.ParseUint(c.Query("webhook_id"), 10, 64)
	page, limit := pagination.ParsePagination(c, 20, 100)
	offset := (page - 1) * limit

	deliveries, total, err := h.services.Webhook.ListDeliveries(c.Request.Context(), uint(webhookID), limit, offset)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "webhook_deliveries_failed")
		return
	}

	if deliveries == nil {
		deliveries = []modelsOrder.WebhookDelivery{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       deliveries,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + int64(limit) - 1) / int64(limit),
	})
}
