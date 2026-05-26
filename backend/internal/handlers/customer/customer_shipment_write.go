package customer

import (
	modelsCommon "candypro/api/internal/models/common"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxShipmentAttachments = 5

// verifyTradeShipmentOwnership 校验贸易与发货归属当前用户。
func (h *Handler) verifyTradeShipmentOwnership(c *gin.Context) (userID string, tradeID uint, shipmentID uint, ok bool) {
	userID, ok = contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return "", 0, 0, false
	}
	tradeID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidResp(c, "invalid_request")
		return "", 0, 0, false
	}
	shipmentID64, err := strconv.ParseUint(c.Param("shipmentId"), 10, 32)
	if err != nil {
		response.InvalidResp(c, "invalid_request")
		return "", 0, 0, false
	}
	tradeID = uint(tradeID64)
	shipmentID = uint(shipmentID64)

	trans, err := h.services.Trade.GetTransaction(c.Request.Context(), tradeID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "trade_not_found")
		return "", 0, 0, false
	}
	if trans.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return "", 0, 0, false
	}

	shipment, err := h.services.Shipment.GetShipment(c.Request.Context(), shipmentID)
	if err != nil || shipment.TransactionID != tradeID {
		response.ErrorResp(c, http.StatusNotFound, "shipment_not_found")
		return "", 0, 0, false
	}
	return userID, tradeID, shipmentID, true
}

// CustomerNudgeShipment 客户针对单笔发货发送催促提醒。
func (h *Handler) CustomerNudgeShipment(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, tradeID, shipmentID, ok := h.verifyTradeShipmentOwnership(c)
	if !ok {
		return
	}

	shipment, err := h.services.Shipment.GetShipment(c.Request.Context(), shipmentID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "shipment_not_found")
		return
	}

	nudgeable := map[string]bool{
		"PENDING":    true,
		"DISPATCHED": true,
		"IN_TRANSIT": true,
		"EXCEPTION":  true,
	}
	if !nudgeable[shipment.Status] {
		response.ErrorResp(c, http.StatusConflict, "cannot_nudge_shipment")
		return
	}

	blNo := strings.TrimSpace(shipment.BillOfLadingNo)
	if blNo == "" {
		blNo = strconv.FormatUint(uint64(shipmentID), 10)
	}
	ref := strconv.FormatUint(uint64(tradeID), 10)

	if h.services.Notification != nil {
		_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
			UserID:    userID,
			Type:      "shipment",
			Reference: ref,
			Title:     "Nudge Sent",
			Message:   "Your reminder for shipment " + blNo + " has been sent to our team.",
		})
	}

	if h.services.User != nil && h.services.Notification != nil {
		adminUsers, userErr := h.services.User.FindAdminUsers(c.Request.Context())
		if userErr == nil {
			for _, admin := range adminUsers {
				_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
					UserID:    admin.ID,
					Type:      "shipment",
					Reference: ref,
					Title:     "Customer Shipment Nudge",
					Message:   "Customer sent a reminder for trade #" + ref + " shipment " + blNo + " (status: " + shipment.Status + ").",
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nudge sent successfully"})
}

// CustomerUploadShipmentAttachment 客户上传发货相关附件（POD、水单等）。
func (h *Handler) CustomerUploadShipmentAttachment(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	_, _, shipmentID, ok := h.verifyTradeShipmentOwnership(c)
	if !ok {
		return
	}

	shipment, err := h.services.Shipment.GetShipment(c.Request.Context(), shipmentID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "shipment_not_found")
		return
	}
	if shipment.Status == "DELIVERED" {
		response.ErrorResp(c, http.StatusConflict, "shipment_already_delivered")
		return
	}

	if err := c.Request.ParseMultipartForm(64 << 20); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "form_parse_failed")
		return
	}

	existing := len(shipment.CustomerAttachments)
	fileHeaders := c.Request.MultipartForm.File["files"]
	if len(fileHeaders) == 0 {
		response.InvalidResp(c, "invalid_request")
		return
	}
	if existing+len(fileHeaders) > maxShipmentAttachments {
		response.ErrorResp(c, http.StatusBadRequest, "file_count_exceeded")
		return
	}

	newFiles, upErr := h.uploadCustomerAttachments(c, fileHeaders, "shipments", maxShipmentAttachments-existing)
	if upErr != nil {
		return
	}

	shipment.CustomerAttachments = append(shipment.CustomerAttachments, newFiles...)
	shipment.UpdatedAt = time.Now()
	if err := h.services.Shipment.UpdateShipment(c.Request.Context(), shipment); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "shipment_update_failed")
		return
	}

	if h.services.User != nil && h.services.Notification != nil {
		adminUsers, userErr := h.services.User.FindAdminUsers(c.Request.Context())
		if userErr == nil {
			ref := strconv.FormatUint(uint64(shipment.TransactionID), 10)
			for _, admin := range adminUsers {
				_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
					UserID:    admin.ID,
					Type:      "shipment",
					Reference: ref,
					Title:     "Shipment Attachment Uploaded",
					Message:   "Customer uploaded " + strconv.Itoa(len(newFiles)) + " file(s) for shipment #" + strconv.FormatUint(uint64(shipmentID), 10) + ".",
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"files":      newFiles,
		"totalFiles": len(shipment.CustomerAttachments),
	})
}
