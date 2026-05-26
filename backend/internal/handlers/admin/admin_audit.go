package admin

import (
	"encoding/json"
	"log"
	"time"

	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"

	"github.com/gin-gonic/gin"
)

// logActivityAudit 写入通用审计日志；可选 userIDOverride 覆盖操作者
func (h *Handler) logActivityAudit(c *gin.Context, action, entityType, entityID, oldValue, newValue string, userIDOverride ...string) {
	if h.services == nil || h.services.ActivityLog == nil {
		return
	}
	userID := c.GetString("userID")
	if len(userIDOverride) > 0 && userIDOverride[0] != "" {
		userID = userIDOverride[0]
	}
	if userID == "" {
		return
	}
	details, marshalErr := json.Marshal(map[string]string{
		"oldValue": oldValue,
		"newValue": newValue,
	})
	// L-7: surface marshal failures rather than silently storing an empty Details
	// blob; the audit row is still useful (action+entity), but operators should
	// know the structured payload was lost.
	if marshalErr != nil {
		details = []byte(`{"error":"failed to marshal audit details"}`)
	}
	uid := userID
	if err := h.services.ActivityLog.LogActivity(c.Request.Context(), &modelsCommon.ActivityLog{
		ID:         crypto.GenerateID(),
		UserID:     &uid,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    string(details),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		CreatedAt:  time.Now(),
	}); err != nil {
		log.Printf("audit: failed to write %s/%s log: %v", entityType, action, err)
	}
}

func (h *Handler) logOrderAudit(c *gin.Context, action, orderID, userID, oldValue, newValue string) {
	h.logActivityAudit(c, action, "order", orderID, oldValue, newValue, userID)
}

func (h *Handler) logInquiryAudit(c *gin.Context, inquiryID, userID, oldValue, newValue string) {
	h.logActivityAudit(c, "inquiry_status_change", "inquiry", inquiryID, oldValue, newValue, userID)
}

// statusHistoryEntry 状态时间线条目（订单/询价共用）
type statusHistoryEntry struct {
	Status    string    `json:"status"`
	Note      string    `json:"note,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// activityHistoryEntry 订单活动日志条目（状态 + 付款等）
type activityHistoryEntry struct {
	Type      string    `json:"type"`
	Label     string    `json:"label"`
	Detail    string    `json:"detail,omitempty"`
	Amount    float64   `json:"amount,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// orderDetailResponse 管理员订单详情（含 statusHistory 与 activityHistory）
type orderDetailResponse struct {
	*modelsOrder.Order
	StatusHistory    []statusHistoryEntry   `json:"statusHistory"`
	ActivityHistory  []activityHistoryEntry `json:"activityHistory"`
	InventoryWarnings []string              `json:"inventoryWarnings,omitempty"`
}

// inquiryDetailResponse 管理员询价详情（含 statusHistory）
type inquiryDetailResponse struct {
	*modelsProduct.Inquiry
	StatusHistory []statusHistoryEntry `json:"statusHistory"`
}

// buildStatusHistoryFromLogs 从 ActivityLog 构建状态时间线
func buildStatusHistoryFromLogs(logs []modelsCommon.ActivityLog, action string) []statusHistoryEntry {
	var entries []statusHistoryEntry
	for _, entry := range logs {
		if entry.Action != action {
			continue
		}
		var details struct {
			OldValue string `json:"oldValue"`
			NewValue string `json:"newValue"`
		}
		if err := json.Unmarshal([]byte(entry.Details), &details); err != nil || details.NewValue == "" {
			continue
		}
		note := ""
		if details.OldValue != "" {
			note = details.OldValue + " -> " + details.NewValue
		}
		entries = append(entries, statusHistoryEntry{
			Status:    details.NewValue,
			Note:      note,
			Timestamp: entry.CreatedAt,
		})
	}
	return entries
}

// buildOrderStatusHistory 从 ActivityLog 构建订单状态时间线
func buildOrderStatusHistory(logs []modelsCommon.ActivityLog) []statusHistoryEntry {
	return buildStatusHistoryFromLogs(logs, "order_status_change")
}

// buildOrderActivityHistory 合并状态变更与付款等事件，按时间排序
func buildOrderActivityHistory(logs []modelsCommon.ActivityLog) []activityHistoryEntry {
	var entries []activityHistoryEntry
	for _, entry := range logs {
		switch entry.Action {
		case "order_status_change":
			var details struct {
				OldValue string `json:"oldValue"`
				NewValue string `json:"newValue"`
			}
			if err := json.Unmarshal([]byte(entry.Details), &details); err != nil || details.NewValue == "" {
				continue
			}
			detail := details.NewValue
			if details.OldValue != "" {
				detail = details.OldValue + " -> " + details.NewValue
			}
			entries = append(entries, activityHistoryEntry{
				Type: "status", Label: details.NewValue, Detail: detail, Timestamp: entry.CreatedAt,
			})
		case "payment_create", "payment_confirm", "payment_refund", "invoice_auto_created":
			var details struct {
				OrderID   string  `json:"orderId"`
				PaymentID string  `json:"paymentId"`
				Amount    float64 `json:"amount"`
			}
			_ = json.Unmarshal([]byte(entry.Details), &details)
			entries = append(entries, activityHistoryEntry{
				Type: entry.Action, Label: entry.Action, Detail: details.PaymentID, Amount: details.Amount, Timestamp: entry.CreatedAt,
			})
		}
	}
	// 简单按时间升序
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Timestamp.Before(entries[i].Timestamp) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	return entries
}
