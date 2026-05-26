package trade

import (
	"fmt"
	"strconv"
	"strings"

	tradeModels "candypro/api/internal/models/trade"
)

// BuildTradeDocGenerationPrompt builds an AI prompt for generating a trade document.
func BuildTradeDocGenerationPrompt(docType string, trade *tradeModels.TradeTransaction, userPrompt, extraContext, orderContext string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("You are a trade document assistant. Generate a %s for the following trade transaction.\n\n", docTypeLabel(docType)))
	builder.WriteString(fmt.Sprintf("Trade ID: %d\n", trade.ID))
	builder.WriteString(fmt.Sprintf("Reference: %s\n", trade.Reference))
	builder.WriteString(fmt.Sprintf("Status: %s\n", trade.Status))
	builder.WriteString(fmt.Sprintf("Currency: %s\n", trade.Currency))
	builder.WriteString(fmt.Sprintf("Total Amount: %.2f\n", trade.TotalAmount))
	builder.WriteString(fmt.Sprintf("Terms: %s\n", trade.Terms))

	if strings.TrimSpace(orderContext) != "" {
		builder.WriteString("\nLinked Order Data (use line items for document content):\n")
		builder.WriteString(orderContext)
		builder.WriteString("\n")
	}

	if extraContext != "" {
		builder.WriteString(fmt.Sprintf("\nAdditional Context:\n%s\n", extraContext))
	}

	if userPrompt != "" {
		builder.WriteString(fmt.Sprintf("\nUser Instructions:\n%s\n", userPrompt))
	}

	builder.WriteString("\nCall the appropriate tool to generate this document. Provide the trade_id as ")
	builder.WriteString(strconv.FormatUint(uint64(trade.ID), 10))
	builder.WriteString(" in your tool call.\n")

	return builder.String()
}

func docTypeLabel(docType string) string {
	labels := map[string]string{
		"PROFORMA_INVOICE":              "Proforma Invoice (PI)",
		"COMMERCIAL_INVOICE":            "Commercial Invoice (CI)",
		"SALES_CONTRACT":                "Sales Contract (SC)",
		"PACKING_LIST":                  "Packing List (PL)",
		"ORIGIN_CERTIFICATE":            "Certificate of Origin (COO)",
		"HEALTH_CERTIFICATE":            "Health Certificate (HC)",
		"BILL_OF_LADING":                "Bill of Lading (B/L)",
		"INGREDIENTS_DECLARATION":       "Ingredients Declaration",
		"SHIPPER_LETTER_OF_INSTRUCTION": "Shipper's Letter of Instruction (SLI)",
		"INSURANCE_CERTIFICATE":         "Insurance Certificate",
	}
	if label, ok := labels[docType]; ok {
		return label
	}
	return docType
}
