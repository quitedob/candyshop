package system

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	modelsAuth "candypro/api/internal/models/auth"
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/response"
	tradeSvc "candypro/api/internal/services/trade"

	"github.com/gin-gonic/gin"
)

func isAdminPortalRole(role string) bool {
	for _, r := range modelsAuth.AdminPortal() {
		if r == role {
			return true
		}
	}
	return false
}

// resolveOrderForUser 按 JWT 角色校验订单访问（管理员全量，客户仅本人订单）。
func (h *Handler) resolveOrderForUser(c *gin.Context, orderID string) (*modelsOrder.Order, error) {
	if h.services == nil || h.services.Order == nil {
		return nil, errors.New("order service unavailable")
	}
	userID, role := authContext(c)
	order, err := h.services.Order.GetOrder(c.Request.Context(), strings.TrimSpace(orderID))
	if err != nil {
		return nil, err
	}
	if isAdminPortalRole(role) {
		return order, nil
	}
	if strings.TrimSpace(order.UserID) != userID {
		return nil, errors.New("forbidden")
	}
	return order, nil
}

// enrichOrderQuery 为 SSE Agent 注入订单业务上下文。
func (h *Handler) enrichOrderQuery(c *gin.Context, orderID, query string) string {
	order, err := h.resolveOrderForUser(c, orderID)
	if err != nil {
		return fmt.Sprintf("[Order context unavailable for %s]\n\n%s", orderID, query)
	}
	return tradeSvc.BuildOrderChatPrompt(order, query)
}

// ChatbotOrderContext 已登录用户带订单上下文的 Chatbot（非 RAG，直连 LLM）。
func (h *Handler) ChatbotOrderContext(c *gin.Context) {
	if h.aiUnavailable(c) {
		return
	}

	var req struct {
		Message string `json:"message" binding:"required"`
		OrderID string `json:"orderId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid_request")
		return
	}

	order, err := h.resolveOrderForUser(c, req.OrderID)
	if err != nil {
		if err.Error() == "forbidden" {
			response.ErrorResp(c, http.StatusForbidden, "forbidden")
			return
		}
		response.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}

	prompt := tradeSvc.BuildOrderChatPrompt(order, strings.TrimSpace(req.Message))
	reply, genErr := h.aiService.Generate(c.Request.Context(), prompt)
	if genErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_chatbot_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reply":       reply,
		"orderId":     order.ID,
		"orderNumber": order.OrderNumber,
		"generatedAt": time.Now(),
		"notice":      "AI-generated content is for general information only and is not legal, regulatory, or financial advice.",
	})
}
