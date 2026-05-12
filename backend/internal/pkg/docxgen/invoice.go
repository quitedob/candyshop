package docxgen

import (
	"fmt"
	"time"
)

// InvoiceData holds the data for generating an invoice DOCX.
type InvoiceData struct {
	InvoiceNo   string
	InvoiceType string
	OrderID     string
	BuyerName   string
	SellerName  string
	Amount      float64
	TaxAmount   float64
	TotalAmount float64
	Currency    string
	DueDate     time.Time
	IssueDate   time.Time
	Notes       string
	Items       []InvoiceLineItem
}

// InvoiceLineItem represents a line item in an invoice.
type InvoiceLineItem struct {
	ProductName string
	Quantity    int
	UnitPrice   float64
	TotalPrice  float64
}

func invoiceTypeLabel(t string) string {
	switch t {
	case "proforma":
		return "PROFORMA INVOICE"
	case "commercial":
		return "COMMERCIAL INVOICE"
	case "credit_note":
		return "CREDIT NOTE"
	default:
		return "INVOICE"
	}
}

// GenerateInvoice generates a .docx invoice.
func GenerateInvoice(data InvoiceData) ([]byte, error) {
	var body string

	body += Paragraph(invoiceTypeLabel(data.InvoiceType), ParBold(), ParFontSize(32), ParCenter())
	body += Spacer()

	body += Paragraph(fmt.Sprintf("Invoice No: %s", data.InvoiceNo), ParFontSize(22), ParCenter())
	body += Paragraph(fmt.Sprintf("Issue Date: %s", DateStr(data.IssueDate)), ParFontSize(22), ParCenter())
	if !data.DueDate.IsZero() {
		body += Paragraph(fmt.Sprintf("Due Date: %s", DateStr(data.DueDate)), ParFontSize(22), ParCenter())
	}
	body += Spacer()

	body += Paragraph("FROM / TO", ParBold(), ParFontSize(24))
	body += Spacer()
	body += Paragraph(fmt.Sprintf("Seller: %s", data.SellerName), ParFontSize(22))
	body += Paragraph(fmt.Sprintf("Buyer: %s", data.BuyerName), ParFontSize(22))
	if data.OrderID != "" {
		body += Paragraph(fmt.Sprintf("Order Reference: %s", data.OrderID), ParFontSize(22))
	}
	body += Spacer()

	if len(data.Items) > 0 {
		body += Paragraph("LINE ITEMS", ParBold(), ParFontSize(24))
		body += Spacer()

		headers := []string{"Product", "Quantity", "Unit Price", "Total"}
		rows := make([][]string, 0, len(data.Items))
		for _, item := range data.Items {
			rows = append(rows, []string{
				item.ProductName,
				fmt.Sprintf("%d", item.Quantity),
				FormatMoney(item.UnitPrice, data.Currency),
				FormatMoney(item.TotalPrice, data.Currency),
			})
		}
		body += Table(headers, rows, ColWidths("35", "15", "20", "30"))
		body += Spacer()
	}

	totalRows := [][]string{
		{"Subtotal", "", "", FormatMoney(data.Amount, data.Currency)},
		{"Tax", "", "", FormatMoney(data.TaxAmount, data.Currency)},
	}
	body += Table([]string{"Description", "", "", "Amount"}, totalRows, ColWidths("35", "15", "20", "30"))
	body += Spacer()
	body += Paragraph(
		fmt.Sprintf("Total Due: %s", FormatMoney(data.TotalAmount, data.Currency)),
		ParBold(), ParFontSize(28), ParRight(),
	)
	body += Spacer()

	if data.Notes != "" {
		body += Paragraph("NOTES", ParBold(), ParFontSize(24))
		body += Spacer()
		body += Paragraph(data.Notes, ParFontSize(20), ParItalic())
		body += Spacer()
	}

	body += Paragraph("PAYMENT INSTRUCTIONS", ParBold(), ParFontSize(24))
	body += Spacer()
	body += Paragraph("Please remit payment to the bank account specified in the sales agreement.", ParFontSize(20))
	body += Paragraph("Bank: CandyPro Manufacturing - Corporate Banking", ParFontSize(20))
	body += Paragraph(fmt.Sprintf("Reference: %s", data.InvoiceNo), ParFontSize(20))

	return GenerateDocx(body)
}
