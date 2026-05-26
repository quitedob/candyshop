package customer

import (
	"net/http"
	"strings"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type gatewayPaymentRequest struct {
	Method string  `json:"method" binding:"required"`
	Amount float64 `json:"amount"`
}

// CustomerCreateGatewayPayment 创建 Stripe/PayPal 网关支付会话
func (h *Handler) CustomerCreateGatewayPayment(c *gin.Context) {
	if h.services == nil || h.services.GatewayPayment == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	orderID := c.Param("id")
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	ord, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}
	if ord.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	var req gatewayPaymentRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	method := strings.ToLower(strings.TrimSpace(req.Method))
	if method != modelsOrder.PaymentMethodStripe && method != modelsOrder.PaymentMethodPayPal {
		response.InvalidResp(c, "invalid_request")
		return
	}
	result, err := h.services.GatewayPayment.CreateCheckout(c.Request.Context(), ord, userID, method, req.Amount)
	if err != nil {
		switch err.Error() {
		case "stripe_not_configured", "paypal_not_configured":
			response.ErrorResp(c, http.StatusServiceUnavailable, "gateway_not_configured")
		case "forbidden":
			response.ErrorResp(c, http.StatusForbidden, "forbidden")
		default:
			if strings.Contains(err.Error(), "exceeds remaining balance") {
				response.InvalidResp(c, "invalid_request")
			} else {
				response.ErrorResp(c, http.StatusInternalServerError, "payment_create_failed")
			}
		}
		return
	}
	c.JSON(http.StatusCreated, result)
}
