package docxgen

import (
	"fmt"
	"time"
)

// SalesContractData holds the data for generating a sales contract DOCX.
type SalesContractData struct {
	ContractNo      string
	BuyerName       string
	SellerName      string
	Incoterms       string
	TermsOfPayment  string
	QualityStandard string
	TotalAmount     float64
	Currency        string
	SignDate        time.Time
	ValidUntil      time.Time
	OrderRef        string
	Items           []ContractLineItem
}

// ContractLineItem represents a line item in a contract.
type ContractLineItem struct {
	ProductName string
	Quantity    int
	Unit        string
	UnitPrice   float64
	TotalPrice  float64
}

// GenerateSalesContract generates a .docx sales contract.
func GenerateSalesContract(data SalesContractData) ([]byte, error) {
	var body string

	body += Paragraph("SALES CONTRACT", ParBold(), ParFontSize(32), ParCenter())
	body += Spacer()

	body += Paragraph(fmt.Sprintf("Contract No: %s", data.ContractNo), ParFontSize(22), ParCenter())
	body += Paragraph(fmt.Sprintf("Date: %s", DateStr(data.SignDate)), ParFontSize(22), ParCenter())
	body += Spacer()

	body += Paragraph("PARTIES", ParBold(), ParFontSize(24))
	body += Spacer()
	body += Paragraph(fmt.Sprintf("Seller: %s", data.SellerName), ParFontSize(22))
	body += Paragraph(fmt.Sprintf("Buyer: %s", data.BuyerName), ParFontSize(22))
	if data.OrderRef != "" {
		body += Paragraph(fmt.Sprintf("Order Reference: %s", data.OrderRef), ParFontSize(22))
	}
	body += Spacer()

	body += Paragraph("TERMS AND CONDITIONS", ParBold(), ParFontSize(24))
	body += Spacer()
	body += Paragraph(fmt.Sprintf("Incoterms: %s", data.Incoterms), ParFontSize(22))
	body += Paragraph(fmt.Sprintf("Terms of Payment: %s", data.TermsOfPayment), ParFontSize(22))
	if data.QualityStandard != "" {
		body += Paragraph(fmt.Sprintf("Quality Standard: %s", data.QualityStandard), ParFontSize(22))
	}
	body += Spacer()

	if len(data.Items) > 0 {
		body += Paragraph("PRODUCTS", ParBold(), ParFontSize(24))
		body += Spacer()

		headers := []string{"Product", "Quantity", "Unit", "Unit Price", "Total"}
		rows := make([][]string, 0, len(data.Items))
		for _, item := range data.Items {
			rows = append(rows, []string{
				item.ProductName,
				fmt.Sprintf("%d", item.Quantity),
				item.Unit,
				FormatMoney(item.UnitPrice, data.Currency),
				FormatMoney(item.TotalPrice, data.Currency),
			})
		}
		body += Table(headers, rows, ColWidths("30", "15", "10", "20", "25"))
		body += Spacer()
	}

	body += Paragraph(
		fmt.Sprintf("Total Amount: %s", FormatMoney(data.TotalAmount, data.Currency)),
		ParBold(), ParFontSize(24), ParRight(),
	)
	body += Spacer()

	body += Paragraph("VALIDITY", ParBold(), ParFontSize(24))
	body += Spacer()
	body += Paragraph(fmt.Sprintf("This contract is valid until: %s", DateStr(data.ValidUntil)), ParFontSize(22))
	body += Spacer()

	body += Paragraph("AUTHORIZED SIGNATURES", ParBold(), ParFontSize(24))
	body += Spacer()
	body += Paragraph("_________________________", ParFontSize(22))
	body += Paragraph("Seller Signature & Stamp", ParItalic(), ParFontSize(20))
	body += Spacer()
	body += Paragraph("_________________________", ParFontSize(22))
	body += Paragraph("Buyer Signature & Stamp", ParItalic(), ParFontSize(20))

	return GenerateDocx(body)
}
