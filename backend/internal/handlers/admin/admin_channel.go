package admin

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ── Multi-Channel Architecture ──

// AdminListChannels returns all sales channels.
func (h *Handler) AdminListChannels(c *gin.Context) {
	if h.services == nil || h.services.Channel == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	channels, err := h.services.Channel.List(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "channel_list_failed")
		return
	}
	if channels == nil {
		channels = []modelsProduct.Channel{}
	}
	c.JSON(http.StatusOK, channels)
}

// AdminGetChannel returns a single channel.
func (h *Handler) AdminGetChannel(c *gin.Context) {
	if h.services == nil || h.services.Channel == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_channel_id")
		return
	}
	ch, err := h.services.Channel.Get(c.Request.Context(), uint(id))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "channel_not_found")
		return
	}
	c.JSON(http.StatusOK, ch)
}

// AdminCreateChannel creates a new sales channel.
func (h *Handler) AdminCreateChannel(c *gin.Context) {
	if h.services == nil || h.services.Channel == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		Code             string                       `json:"code" binding:"required"`
		Name             string                       `json:"name" binding:"required"`
		Type             string                       `json:"type"`
		Currency         string                       `json:"currency"`
		TaxConfig        *modelsProduct.ChannelTaxJSON `json:"taxConfig"`
		FulfillmentMode  string                       `json:"fulfillmentMode"`
		PriceMultiplier  *float64                      `json:"priceMultiplier"`
		DefaultWarehouse *string                       `json:"defaultWarehouse"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	ch := &modelsProduct.Channel{
		Code:            strings.TrimSpace(req.Code),
		Name:            strings.TrimSpace(req.Name),
		Type:            defaultValue(req.Type, "marketplace"),
		Status:          "active",
		Currency:        defaultValue(req.Currency, "USD"),
		FulfillmentMode: defaultValue(req.FulfillmentMode, "self"),
		PriceMultiplier: 1.0,
		DefaultWarehouse: req.DefaultWarehouse,
	}
	if req.PriceMultiplier != nil {
		ch.PriceMultiplier = *req.PriceMultiplier
	}
	if req.TaxConfig != nil {
		ch.TaxConfig = *req.TaxConfig
	}
	if err := h.services.Channel.Create(c.Request.Context(), ch); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "channel_create_failed")
		return
	}
	c.JSON(http.StatusCreated, ch)
}

// AdminUpdateChannel updates a sales channel.
func (h *Handler) AdminUpdateChannel(c *gin.Context) {
	if h.services == nil || h.services.Channel == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_channel_id")
		return
	}
	ch, err := h.services.Channel.Get(c.Request.Context(), uint(id))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "channel_not_found")
		return
	}

	var req struct {
		Name             *string                       `json:"name"`
		Type             *string                       `json:"type"`
		Currency         *string                       `json:"currency"`
		TaxConfig        *modelsProduct.ChannelTaxJSON `json:"taxConfig"`
		FulfillmentMode  *string                       `json:"fulfillmentMode"`
		PriceMultiplier  *float64                      `json:"priceMultiplier"`
		DefaultWarehouse *string                       `json:"defaultWarehouse"`
		Status           *string                       `json:"status"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Name != nil {
		ch.Name = strings.TrimSpace(*req.Name)
	}
	if req.Type != nil {
		ch.Type = strings.TrimSpace(*req.Type)
	}
	if req.Currency != nil {
		ch.Currency = strings.TrimSpace(*req.Currency)
	}
	if req.TaxConfig != nil {
		ch.TaxConfig = *req.TaxConfig
	}
	if req.FulfillmentMode != nil {
		ch.FulfillmentMode = strings.TrimSpace(*req.FulfillmentMode)
	}
	if req.PriceMultiplier != nil {
		ch.PriceMultiplier = *req.PriceMultiplier
	}
	if req.DefaultWarehouse != nil {
		ch.DefaultWarehouse = req.DefaultWarehouse
	}
	if req.Status != nil {
		s := strings.TrimSpace(*req.Status)
		if s != "active" && s != "inactive" {
			response.InvalidResp(c, "invalid_channel_status")
			return
		}
		ch.Status = s
	}
	if err := h.services.Channel.Update(c.Request.Context(), ch); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "channel_update_failed")
		return
	}
	c.JSON(http.StatusOK, ch)
}

// AdminDeleteChannel deletes a sales channel.
func (h *Handler) AdminDeleteChannel(c *gin.Context) {
	if h.services == nil || h.services.Channel == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_channel_id")
		return
	}
	if err := h.services.Channel.Delete(c.Request.Context(), uint(id)); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "channel_delete_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func defaultValue(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
