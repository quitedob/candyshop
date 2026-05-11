// Package admin — XLSX export handlers with currency auto-conversion.
// Uses excelize (pure Go, no CGo). Supports ?currency=EUR|CNY|GBP|JPY|SAR|AED  query param
// to auto-convert all monetary values from base USD using configured exchange rates.
package admin

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ── Currency Helpers ──

func (h *Handler) resolveCurrency(c *gin.Context) (string, func(float64) float64) {
	target := strings.ToUpper(strings.TrimSpace(c.Query("currency")))
	if target == "" || target == "USD" {
		return "USD", func(v float64) float64 { return v }
	}
	if h.cfg == nil {
		return "USD", func(v float64) float64 { return v }
	}
	rate, ok := h.cfg.ExchangeRates[target]
	if !ok || rate <= 0 {
		return "USD", func(v float64) float64 { return v }
	}
	return target, func(v float64) float64 { return v * rate }
}

func formatMoney(amount float64, currency string) string {
	symbols := map[string]string{
		"USD": "$", "EUR": "€", "CNY": "¥", "GBP": "£", "JPY": "¥",
		"SAR": "﷼", "AED": "د.إ", "KRW": "₩", "SGD": "S$",
		"MYR": "RM", "THB": "฿", "VND": "₫", "IDR": "Rp", "PHP": "₱",
	}
	sym := symbols[currency]
	if sym == "" {
		sym = currency + " "
	}
	return fmt.Sprintf("%s%.2f", sym, amount)
}

// ── Product List Export ──

// AdminExportProductsXLSX exports the full product catalog as XLSX.
// GET /admin/xlsx/export/products?currency=EUR
func (h *Handler) AdminExportProductsXLSX(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	currency, convert := h.resolveCurrency(c)

	resp, err := h.services.Product.GetProducts(c.Request.Context(), 1, 10000, "")
	if err != nil || resp == nil {
		response.ErrorResp(c, http.StatusInternalServerError, "xlsx_products_fetch_failed")
		return
	}
	products, _ := resp.Data.([]modelsProduct.Product)

	f := excelize.NewFile()
	sheet := "Products"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Product ID", "Name", "Category", fmt.Sprintf("Base Price (%s)", currency), "MOQ",
		"Stock", "Lead Time", "Halal", "OEM", "HS Code", "Shelf Life", "Status", "Created At", "Certifications"}
	f.SetSheetRow(sheet, "A1", &headers)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Times New Roman"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F0F0F0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border:    []excelize.Border{{Type: "bottom", Color: "000000", Style: 1}},
	})
	f.SetRowStyle(sheet, 1, 1, headerStyle)

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Family: "Times New Roman"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})

	for i, p := range products {
		row := i + 2
		halal, oem := "No", "No"
		if p.HalalCertified {
			halal = "Yes"
		}
		if p.OEMAvailable {
			oem = "Yes"
		}
		price := formatMoney(convert(p.BasePrice), currency)
		f.SetSheetRow(sheet, fmt.Sprintf("A%d", row), &[]string{
			p.ID, p.Name, p.Category, price, strconv.Itoa(p.MOQ),
			strconv.Itoa(p.StockQuantity), p.LeadTime, halal, oem,
			p.HSCode, p.ShelfLife, p.Status, p.CreatedAt.Format("2006-01-02"),
			strings.Join(p.Certifications, ", "),
		})
		f.SetRowStyle(sheet, row, row, dataStyle)
	}

	for col := 'A'; col <= 'N'; col++ {
		_ = f.SetColWidth(sheet, string(col), string(col), 18)
	}

	serveXLSXFile(c, f, fmt.Sprintf("candypro_products_%s.xlsx", time.Now().Format("20060102")))
}

// ── Orders Export ──

// AdminExportOrdersXLSX exports orders as XLSX.
// GET /admin/xlsx/export/orders?currency=EUR&status=confirmed
func (h *Handler) AdminExportOrdersXLSX(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	currency, convert := h.resolveCurrency(c)

	status := c.Query("status")
	resp, err := h.services.Order.GetOrders(c.Request.Context(), 1, 10000, status, "", "", "")
	if err != nil || resp == nil {
		response.ErrorResp(c, http.StatusInternalServerError, "xlsx_orders_fetch_failed")
		return
	}
	orders, _ := resp.Data.([]modelsOrder.Order)

	f := excelize.NewFile()
	sheet := "Orders"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Order Number", "Status", "Payment",
		fmt.Sprintf("Subtotal (%s)", currency), fmt.Sprintf("Tax (%s)", currency),
		fmt.Sprintf("Shipping (%s)", currency), fmt.Sprintf("Total (%s)", currency),
		"Items", "Tracking", "Created At"}
	f.SetSheetRow(sheet, "A1", &headers)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Times New Roman"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F0F0F0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border:    []excelize.Border{{Type: "bottom", Color: "000000", Style: 1}},
	})
	f.SetRowStyle(sheet, 1, 1, headerStyle)

	for i, o := range orders {
		row := i + 2
		f.SetSheetRow(sheet, fmt.Sprintf("A%d", row), &[]string{
			o.OrderNumber, o.Status, o.PaymentStatus,
			formatMoney(convert(o.Subtotal), currency), formatMoney(convert(o.TaxAmount), currency),
			formatMoney(convert(o.ShippingAmount), currency), formatMoney(convert(o.TotalAmount), currency),
			strconv.Itoa(len(o.Items)), o.TrackingNumber,
			o.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	for col := 'A'; col <= 'J'; col++ {
		_ = f.SetColWidth(sheet, string(col), string(col), 16)
	}

	serveXLSXFile(c, f, fmt.Sprintf("candypro_orders_%s.xlsx", time.Now().Format("20060102")))
}

// ── Revenue Report Export ──

// AdminExportRevenueXLSX exports the financial revenue report as XLSX.
// GET /admin/xlsx/export/revenue?currency=EUR&months=12
func (h *Handler) AdminExportRevenueXLSX(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	currency, convert := h.resolveCurrency(c)

	months, _ := strconv.Atoi(c.DefaultQuery("months", "12"))
	revData, err := h.services.Order.RevenueByMonth(c.Request.Context(), months)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "xlsx_revenue_fetch_failed")
		return
	}

	totalOrders, _ := h.services.Order.CountOrders(c.Request.Context())
	totalSales, _ := h.services.Order.SumSales(c.Request.Context())

	f := excelize.NewFile()
	sheet := fmt.Sprintf("Revenue (%s)", currency)
	f.SetSheetName("Sheet1", sheet)

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Family: "Times New Roman"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.MergeCell(sheet, "A1", "D1")
	f.SetCellValue(sheet, "A1", fmt.Sprintf("CandyPro OEM — Revenue Report (%s)", currency))
	f.SetCellStyle(sheet, "A1", "D1", titleStyle)

	f.SetSheetRow(sheet, "A3", &[]string{"Metric", "Value"})
	f.SetSheetRow(sheet, "A4", &[]string{"Total Orders", strconv.FormatInt(totalOrders, 10)})
	f.SetSheetRow(sheet, "A5", &[]string{fmt.Sprintf("Total Sales (%s)", currency), formatMoney(convert(totalSales), currency)})
	f.SetSheetRow(sheet, "A6", &[]string{"Report Period (Months)", strconv.Itoa(months)})
	f.SetSheetRow(sheet, "A7", &[]string{"Generated At", time.Now().Format("2006-01-02 15:04")})

	tableStart := 9
	f.SetSheetRow(sheet, fmt.Sprintf("A%d", tableStart), &[]string{"Month", fmt.Sprintf("Revenue (%s)", currency), "Orders", "Avg Order Value"})

	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true, Size: 11, Family: "Times New Roman"},
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"#F0F0F0"}, Pattern: 1},
		Border: []excelize.Border{{Type: "bottom", Color: "000000", Style: 1}},
	})
	f.SetRowStyle(sheet, tableStart, tableStart, tableHeaderStyle)

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10, Family: "Times New Roman"},
	})

	for i, m := range revData {
		row := tableStart + 1 + i
		month, _ := m["month"].(string)
		revenue, _ := m["revenue"].(float64)
		var orderCount float64
		switch v := m["order_count"].(type) {
		case float64:
			orderCount = v
		case int64:
			orderCount = float64(v)
		}
		avgOrder := 0.0
		if orderCount > 0 {
			avgOrder = revenue / orderCount
		}
		f.SetSheetRow(sheet, fmt.Sprintf("A%d", row), &[]string{
			month, formatMoney(convert(revenue), currency),
			fmt.Sprintf("%.0f", orderCount), formatMoney(convert(avgOrder), currency),
		})
		f.SetRowStyle(sheet, row, row, dataStyle)
	}

	f.SetColWidth(sheet, "A", "A", 15)
	f.SetColWidth(sheet, "B", "B", 20)
	f.SetColWidth(sheet, "C", "C", 15)
	f.SetColWidth(sheet, "D", "D", 20)

	serveXLSXFile(c, f, fmt.Sprintf("candypro_revenue_%s_%s.xlsx", currency, time.Now().Format("20060102")))
}

// ── Trade Transactions Export ──

// AdminExportTradesXLSX exports trade transactions as XLSX.
// GET /admin/xlsx/export/trades?currency=EUR&status=confirmed
func (h *Handler) AdminExportTradesXLSX(c *gin.Context) {
	if h.services == nil || h.services.Trade == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	currency, convert := h.resolveCurrency(c)

	status := c.Query("status")
	trades, _, err := h.services.Trade.ListAllTransactions(c.Request.Context(), 1, 10000, status)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "xlsx_trades_fetch_failed")
		return
	}

	f := excelize.NewFile()
	sheet := "Trade Transactions"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Reference", "Status", "Currency",
		fmt.Sprintf("Total Amount (%s)", currency), "Terms", "Documents", "Created At"}
	f.SetSheetRow(sheet, "A1", &headers)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true, Size: 11, Family: "Times New Roman"},
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"#F0F0F0"}, Pattern: 1},
		Border: []excelize.Border{{Type: "bottom", Color: "000000", Style: 1}},
	})
	f.SetRowStyle(sheet, 1, 1, headerStyle)

	for i, t := range trades {
		row := i + 2
		f.SetSheetRow(sheet, fmt.Sprintf("A%d", row), &[]string{
			t.Reference, t.Status, t.Currency,
			formatMoney(convert(t.TotalAmount), currency), t.Terms,
			strconv.Itoa(len(t.Documents)), t.CreatedAt.Format("2006-01-02"),
		})
	}

	serveXLSXFile(c, f, fmt.Sprintf("candypro_trades_%s_%s.xlsx", currency, time.Now().Format("20060102")))
}

// ── Customers Export ──

// AdminExportCustomersXLSX exports customer list as XLSX.
// GET /admin/xlsx/export/customers
func (h *Handler) AdminExportCustomersXLSX(c *gin.Context) {
	if h.services == nil || h.services.User == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	users, _, err := h.services.User.GetUsers(c.Request.Context(), 1, 10000)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "xlsx_customers_fetch_failed")
		return
	}

	f := excelize.NewFile()
	sheet := "Customers"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"ID", "First Name", "Last Name", "Email", "Company", "Status", "Created At"}
	f.SetSheetRow(sheet, "A1", &headers)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true, Size: 11, Family: "Times New Roman"},
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"#F0F0F0"}, Pattern: 1},
		Border: []excelize.Border{{Type: "bottom", Color: "000000", Style: 1}},
	})
	f.SetRowStyle(sheet, 1, 1, headerStyle)

	for i, u := range users {
		row := i + 2
		f.SetSheetRow(sheet, fmt.Sprintf("A%d", row), &[]string{
			u.ID, u.FirstName, u.LastName, u.Email, u.Company,
			u.Status, u.CreatedAt.Format("2006-01-02"),
		})
	}

	serveXLSXFile(c, f, fmt.Sprintf("candypro_customers_%s.xlsx", time.Now().Format("20060102")))
}

// ── Helper ──

func serveXLSXFile(c *gin.Context, f *excelize.File, filename string) {
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "xlsx_write_failed")
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Length", strconv.Itoa(buf.Len()))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
