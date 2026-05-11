package documents

import (
	_ "embed"
	"strings"
)

//go:embed font_rules.md
var DocumentFontRule string

//go:embed pi.md
var piPrompt string

//go:embed ci.md
var ciPrompt string

//go:embed sc.md
var scPrompt string

//go:embed pl.md
var plPrompt string

//go:embed coo.md
var cooPrompt string

//go:embed hc.md
var hcPrompt string

//go:embed bl.md
var blPrompt string

//go:embed sli.md
var sliPrompt string

//go:embed ingredients.md
var ingredientsPrompt string

//go:embed insurance.md
var insurancePrompt string

//go:embed quotation.md
var quotationPrompt string

func init() {
	DocumentFontRule = strings.TrimSpace(DocumentFontRule)
}

// BuildDocumentPrompt returns a document generation prompt with font rules injected.
func BuildDocumentPrompt(template string) string {
	return strings.ReplaceAll(template, "{FONT_RULES}", DocumentFontRule)
}

// PIDocumentPrompt returns the Proforma Invoice prompt with font rules.
func PIDocumentPrompt() string { return BuildDocumentPrompt(piPrompt) }

// CIDocumentPrompt returns the Commercial Invoice prompt.
func CIDocumentPrompt() string { return BuildDocumentPrompt(ciPrompt) }

// SalesContractPrompt returns the Sales Contract prompt.
func SalesContractPrompt() string { return BuildDocumentPrompt(scPrompt) }

// PackingListPrompt returns the Packing List prompt.
func PackingListPrompt() string { return BuildDocumentPrompt(plPrompt) }

// COODocumentPrompt returns the Certificate of Origin prompt.
func COODocumentPrompt() string { return BuildDocumentPrompt(cooPrompt) }

// HealthCertPrompt returns the Health Certificate prompt.
func HealthCertPrompt() string { return BuildDocumentPrompt(hcPrompt) }

// BLDocumentPrompt returns the Bill of Lading prompt.
func BLDocumentPrompt() string { return BuildDocumentPrompt(blPrompt) }

// SLIDocumentPrompt returns the SLI prompt.
func SLIDocumentPrompt() string { return BuildDocumentPrompt(sliPrompt) }

// IngredientsDeclPrompt returns the Ingredients Declaration prompt.
func IngredientsDeclPrompt() string { return BuildDocumentPrompt(ingredientsPrompt) }

// InsuranceCertPrompt returns the Insurance Certificate prompt.
func InsuranceCertPrompt() string { return BuildDocumentPrompt(insurancePrompt) }

// QuotationPrompt returns the Quotation prompt.
func QuotationPrompt() string { return BuildDocumentPrompt(quotationPrompt) }
