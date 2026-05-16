package customer

import (
	"encoding/csv"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ---- Requisition Lists ----

func (h *Handler) CustomerCreateRequisitionList(c *gin.Context) {
	if h.services == nil || h.services.RequisitionList == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.InvalidResp(c, "unauthorized")
		return
	}
	var req struct {
		Name  string `json:"name" binding:"required"`
		Notes string `json:"notes"`
		Items []struct {
			ProductID string `json:"productId" binding:"required"`
			Quantity  int    `json:"quantity" binding:"required,min=1"`
		} `json:"items"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	list := &modelsOrder.RequisitionList{
		ID:        crypto.GenerateID(),
		UserID:    userID,
		Name:      strings.TrimSpace(req.Name),
		Notes:     strings.TrimSpace(req.Notes),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	items := make([]modelsOrder.RequisitionListItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = modelsOrder.RequisitionListItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		}
	}
	if err := h.services.RequisitionList.Create(c.Request.Context(), list, items); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "requisition_list_create_failed")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"list": list, "items": items})
}

func (h *Handler) CustomerListRequisitionLists(c *gin.Context) {
	if h.services == nil || h.services.RequisitionList == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.InvalidResp(c, "unauthorized")
		return
	}
	lists, err := h.services.RequisitionList.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "requisition_list_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, lists)
}

func (h *Handler) CustomerGetRequisitionList(c *gin.Context) {
	if h.services == nil || h.services.RequisitionList == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	list, err := h.services.RequisitionList.FindByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "requisition_list_not_found")
		return
	}
	userID, _ := contextUserID(c)
	if list.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	items, _ := h.services.RequisitionList.FindItems(c.Request.Context(), id)
	c.JSON(http.StatusOK, gin.H{"list": list, "items": items})
}

func (h *Handler) CustomerDeleteRequisitionList(c *gin.Context) {
	if h.services == nil || h.services.RequisitionList == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	list, err := h.services.RequisitionList.FindByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "requisition_list_not_found")
		return
	}
	userID, _ := contextUserID(c)
	if list.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	if err := h.services.RequisitionList.Delete(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "requisition_list_delete_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// CustomerConvertRequisitionToOrder creates a draft order from a requisition list.
func (h *Handler) CustomerConvertRequisitionToOrder(c *gin.Context) {
	if h.services == nil || h.services.RequisitionList == nil || h.services.Order == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	list, err := h.services.RequisitionList.FindByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "requisition_list_not_found")
		return
	}
	userID, ok := contextUserID(c)
	if !ok || list.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	items, err := h.services.RequisitionList.FindItems(c.Request.Context(), id)
	if err != nil || len(items) == 0 {
		response.InvalidResp(c, "requisition_list_empty")
		return
	}

	orderItems := make([]modelsOrder.OrderItem, len(items))
	for i, it := range items {
		orderItems[i] = modelsOrder.OrderItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		}
	}
	order := &modelsOrder.Order{
		ID:        crypto.GenerateID(),
		UserID:    userID,
		Status:    modelsOrder.OrderStatusPending,
		Items:     orderItems,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.services.Order.CreateOrder(c.Request.Context(), order); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "order_create_failed")
		return
	}
	c.JSON(http.StatusCreated, order)
}

// ---- CSV Bulk Order ----

// CustomerCreateBulkOrder parses a CSV upload and creates draft orders.
// CSV format: product_id,quantity[,specifications]
func (h *Handler) CustomerCreateBulkOrder(c *gin.Context) {
	if h.services == nil || h.services.Order == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.InvalidResp(c, "unauthorized")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.InvalidResp(c, "missing_file")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		response.InvalidResp(c, "invalid_file_type")
		return
	}

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// Read header row
	_, recErr := reader.Read()
	if recErr != nil {
		response.InvalidResp(c, "empty_file")
		return
	}

	var items []modelsOrder.OrderItem
	lineNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		lineNum++
		if err != nil {
			response.InvalidResp(c, "csv_parse_error")
			return
		}
		if len(record) < 2 {
			continue
		}
		productID := strings.TrimSpace(record[0])
		quantity, convErr := strconv.Atoi(strings.TrimSpace(record[1]))
		if convErr != nil || quantity <= 0 || productID == "" {
			continue
		}
		specifications := ""
		if len(record) >= 3 {
			specifications = strings.TrimSpace(record[2])
		}
		items = append(items, modelsOrder.OrderItem{
			ProductID:      productID,
			Quantity:       quantity,
			Specifications: specifications,
		})
	}

	if len(items) == 0 {
		response.InvalidResp(c, "no_valid_rows")
		return
	}

	order := &modelsOrder.Order{
		ID:        crypto.GenerateID(),
		UserID:    userID,
		Status:    modelsOrder.OrderStatusPending,
		Items:     items,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.services.Order.CreateOrder(c.Request.Context(), order); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "bulk_order_create_failed")
		return
	}

	c.JSON(http.StatusCreated, order)
}

// ---- Reorder from History ----

// CustomerReorderFromHistory creates a new draft order from a previous order's items.
func (h *Handler) CustomerReorderFromHistory(c *gin.Context) {
	if h.services == nil || h.services.Order == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.InvalidResp(c, "unauthorized")
		return
	}

	sourceOrderID := c.Param("id")
	sourceOrder, err := h.services.Order.GetOrder(c.Request.Context(), sourceOrderID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}
	if sourceOrder.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	newOrder := &modelsOrder.Order{
		ID:        crypto.GenerateID(),
		UserID:    userID,
		Status:    modelsOrder.OrderStatusPending,
		Items:     sourceOrder.Items,
		Currency:  sourceOrder.Currency,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.services.Order.CreateOrder(c.Request.Context(), newOrder); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "reorder_failed")
		return
	}
	c.JSON(http.StatusCreated, newOrder)
}
