package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"candypro/api/internal/pkg/docxgen"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

var safeFilenameRE = regexp.MustCompile("[^a-zA-Z0-9._-]")

func sanitizeFilename(s string) string {
	return safeFilenameRE.ReplaceAllString(s, "_")
}

// invoiceItemJSON 解析发票行 JSON（兼容订单快照与手工录入两种格式）
type invoiceItemJSON struct {
	ProductID      string  `json:"productId"`
	Quantity       int     `json:"quantity"`
	UnitPrice      float64 `json:"unitPrice"`
	Specifications string  `json:"specifications,omitempty"`
	Name           string  `json:"name"`
	Qty            int     `json:"qty"`
	Total          float64 `json:"total"`
}

// resolveInvoiceBuyerName 从订单关联用户解析购买方名称
func (h *Handler) resolveInvoiceBuyerName(c *gin.Context, orderID string, fallback string) string {
	if orderID == "" || h.services == nil || h.services.Order == nil {
		return fallback
	}
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil || order == nil {
		return fallback
	}
	name := ""
	if h.services.User != nil && order.UserID != "" {
		if user, uerr := h.services.User.GetByID(c.Request.Context(), order.UserID); uerr == nil && user != nil {
			name = strings.TrimSpace(user.Company)
			if name == "" {
				name = strings.TrimSpace(strings.TrimSpace(user.FirstName + " " + user.LastName))
			}
			if name == "" {
				name = strings.TrimSpace(user.Email)
			}
		}
	}
	if name == "" {
		name = strings.TrimSpace(order.ShippingAddress.Country)
	}
	if name == "" {
		return fallback
	}
	return name
}

// parseInvoiceItems parses the Invoice.Items JSON text into DOCX line items.
func parseInvoiceItems(itemsJSON string) []docxgen.InvoiceLineItem {
	if itemsJSON == "" {
		return nil
	}
	var raw []invoiceItemJSON
	if err := json.Unmarshal([]byte(itemsJSON), &raw); err != nil {
		return nil
	}
	items := make([]docxgen.InvoiceLineItem, 0, len(raw))
	for _, r := range raw {
		qty := r.Quantity
		if qty == 0 {
			qty = r.Qty
		}
		unitPrice := r.UnitPrice
		total := r.Total
		if total == 0 && qty > 0 {
			total = unitPrice * float64(qty)
		}
		name := r.Name
		if name == "" {
			name = r.ProductID
		}
		if name == "" {
			name = r.Specifications
		}
		items = append(items, docxgen.InvoiceLineItem{
			ProductName: name,
			Quantity:    qty,
			UnitPrice:   unitPrice,
			TotalPrice:  total,
		})
	}
	return items
}

// AdminExportTradeContract exports a trade's sales contract as DOCX.
// GET /admin/trades/:id/export/contract
// GET /admin/trades/:id/export/sales-contract (alias)
func (h *Handler) AdminExportTradeContract(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	tradeID, err := parseUintParam(c, "id")
	if err != nil {
		response.InvalidResp(c, "invalid_transaction_id")
		return
	}

	// Fetch sales contract data via the trade document detail service
	sc, err := h.services.TradeDocDetail.GetSalesContract(c.Request.Context(), tradeID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "sales_contract_not_found")
		return
	}

	// Fetch trade transaction for additional context (order reference)
	transaction, _ := h.services.Trade.GetTransaction(c.Request.Context(), tradeID)
	orderRef := ""
	if transaction != nil {
		orderRef = transaction.Reference
	}

	now := time.Now()
	signDate := now
	if sc.SignDate != nil {
		signDate = *sc.SignDate
	}
	validUntil := now.AddDate(0, 0, 30)
	if sc.ValidUntil != nil {
		validUntil = *sc.ValidUntil
	}

	data := docxgen.SalesContractData{
		ContractNo:      sc.ContractNo,
		BuyerName:       sc.BuyerName,
		SellerName:      sc.SellerName,
		Incoterms:       sc.Incoterms,
		TermsOfPayment:  sc.TermsOfPayment,
		QualityStandard: sc.QualityStandard,
		TotalAmount:     sc.TotalAmount,
		Currency:        sc.Currency,
		SignDate:        signDate,
		ValidUntil:      validUntil,
		OrderRef:        orderRef,
		Items:           nil, // SalesContract does not store line items; they are managed via TradeDocument
	}

	docxBytes, err := docxgen.GenerateSalesContract(data)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "docx_generation_failed")
		return
	}

	filename := fmt.Sprintf("sales_contract_%s.docx", sanitizeFilename(sc.ContractNo))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
}

// AdminExportTradeSalesContract is an alias for AdminExportTradeContract routed at
// GET /admin/trades/:id/export/sales-contract
func (h *Handler) AdminExportTradeSalesContract(c *gin.Context) {
	h.AdminExportTradeContract(c)
}

// AdminExportInvoice exports an invoice as DOCX.
// GET /admin/invoices/:id/export
func (h *Handler) AdminExportInvoice(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	invoice, err := h.services.Invoice.GetInvoice(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "invoice_not_found")
		return
	}

	now := time.Now()
	issueDate := invoice.CreatedAt
	if invoice.SentAt != nil {
		issueDate = *invoice.SentAt
	}
	dueDate := now.AddDate(0, 0, 30)
	if invoice.DueDate != nil {
		dueDate = *invoice.DueDate
	}

	data := docxgen.InvoiceData{
		InvoiceNo:   invoice.InvoiceNo,
		InvoiceType: invoice.Type,
		OrderID:     invoice.OrderID,
		BuyerName:   invoice.Notes, // Fallback; actual buyer name comes from order
		SellerName:  "CandyPro Manufacturing",
		Amount:      invoice.Amount,
		TaxAmount:   invoice.TaxAmount,
		TotalAmount: invoice.TotalAmount,
		Currency:    invoice.Currency,
		DueDate:     dueDate,
		IssueDate:   issueDate,
		Notes:       invoice.Notes,
		Items:       parseInvoiceItems(invoice.Items),
	}

	// 关联订单时解析真实购买方名称
	if invoice.OrderID != "" {
		data.BuyerName = h.resolveInvoiceBuyerName(c, invoice.OrderID, data.BuyerName)
	}

	docxBytes, err := docxgen.GenerateInvoice(data)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "docx_generation_failed")
		return
	}

	filename := fmt.Sprintf("invoice_%s.docx", sanitizeFilename(invoice.InvoiceNo))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
}

// AdminExportInvoiceProforma exports an invoice as a proforma DOCX.
// GET /admin/invoices/:id/export/proforma-invoice
func (h *Handler) AdminExportInvoiceProforma(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	invoice, err := h.services.Invoice.GetInvoice(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "invoice_not_found")
		return
	}

	now := time.Now()
	issueDate := invoice.CreatedAt
	if invoice.SentAt != nil {
		issueDate = *invoice.SentAt
	}
	dueDate := now.AddDate(0, 0, 30)
	if invoice.DueDate != nil {
		dueDate = *invoice.DueDate
	}

	data := docxgen.InvoiceData{
		InvoiceNo:   invoice.InvoiceNo,
		InvoiceType: invoice.Type,
		OrderID:     invoice.OrderID,
		BuyerName:   invoice.Notes,
		SellerName:  "CandyPro Manufacturing",
		Amount:      invoice.Amount,
		TaxAmount:   invoice.TaxAmount,
		TotalAmount: invoice.TotalAmount,
		Currency:    invoice.Currency,
		DueDate:     dueDate,
		IssueDate:   issueDate,
		Notes:       invoice.Notes,
		Items:       parseInvoiceItems(invoice.Items),
	}

	if invoice.OrderID != "" {
		data.BuyerName = h.resolveInvoiceBuyerName(c, invoice.OrderID, data.BuyerName)
	}

	docxBytes, err := docxgen.GenerateInvoice(data)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "docx_generation_failed")
		return
	}

	filename := fmt.Sprintf("proforma_invoice_%s.docx", sanitizeFilename(invoice.InvoiceNo))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
}
