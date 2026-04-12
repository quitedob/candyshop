package tool

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// ComplianceCheckRequest is for checking country compliance
type ComplianceCheckRequest struct {
	CountryCode string `json:"country_code" jsonschema_description:"Destination country code, e.g., US, UK, CN, JP"`
	ProductType string `json:"product_type" jsonschema_description:"Type of product, e.g., Gummy Candy, Hard Candy, Chocolate"`
	Ingredients string `json:"ingredients" jsonschema_description:"Comma-separated list of ingredients"`
}

type ComplianceCheckResponse struct {
	IsCompliant bool     `json:"is_compliant"`
	Flags       []string `json:"flags" jsonschema_description:"Specific compliance flags or warnings"`
	LanguageReq string   `json:"language_req" jsonschema_description:"Language required for the label"`
}

type complianceRule struct {
	LabelLanguage string
	RequiredItems []string
	Restricted    map[string]string
}

var complianceRules = map[string]complianceRule{
	"US": {
		LabelLanguage: "English",
		RequiredItems: []string{
			"FDA-compliant nutrition facts panel",
			"Major allergen declaration",
			"Ingredient list in descending order",
		},
		Restricted: map[string]string{
			"cyclamate": "Cyclamate is not permitted in US food products",
		},
	},
	"JP": {
		LabelLanguage: "Japanese",
		RequiredItems: []string{
			"MHLW-compliant additive declaration",
			"Allergen labeling for specified ingredients",
			"Manufacturer/importer address in Japanese",
		},
		Restricted: map[string]string{
			"red 2": "Color additive Red No.2 is prohibited in Japan",
		},
	},
	"SA": {
		LabelLanguage: "Arabic",
		RequiredItems: []string{
			"Halal certification and mark",
			"Arabic labeling on retail package",
			"SFDA importer registration details",
		},
		Restricted: map[string]string{
			"gelatin (porcine)": "Porcine-derived ingredients are not compliant for Saudi market",
			"alcohol":           "Alcohol-containing flavor carriers are restricted for Saudi market",
		},
	},
	"EU": {
		LabelLanguage: "Destination market language",
		RequiredItems: []string{
			"EU allergen highlight formatting",
			"Additives listed by functional class and E-number",
			"Traceability lot code on final label",
		},
		Restricted: map[string]string{
			"titanium dioxide": "Titanium dioxide (E171) is no longer authorized in EU food",
		},
	},
}

func normalizeCountry(code string) string {
	value := strings.ToUpper(strings.TrimSpace(code))
	switch value {
	case "USA":
		return "US"
	case "UK":
		return "EU"
	default:
		return value
	}
}

func parseIngredients(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		v := strings.ToLower(strings.TrimSpace(part))
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func hasToken(ingredients []string, token string) bool {
	token = strings.ToLower(strings.TrimSpace(token))
	for _, ingredient := range ingredients {
		if strings.Contains(ingredient, token) {
			return true
		}
	}
	return false
}

// NewComplianceCheckTool evaluates label/law fit based on destination and ingredients.
func NewComplianceCheckTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("check_compliance", "Check if the candy ingredients and product type comply with destination country regulations.",
		func(ctx context.Context, req *ComplianceCheckRequest) (*ComplianceCheckResponse, error) {
			if req == nil {
				return nil, fmt.Errorf("request is required")
			}

			countryCode := normalizeCountry(req.CountryCode)
			if countryCode == "" {
				return nil, fmt.Errorf("country_code is required")
			}

			rule, ok := complianceRules[countryCode]
			if !ok {
				rule = complianceRule{
					LabelLanguage: "Destination market official language",
					RequiredItems: []string{
						"Ingredient list and allergen statement",
						"Nutrition facts panel",
						"Importer details and lot traceability",
					},
				}
			}

			ingredients := parseIngredients(req.Ingredients)
			flags := make([]string, 0, 8)
			isCompliant := true

			if len(ingredients) == 0 {
				isCompliant = false
				flags = append(flags, "Ingredients list is empty; compliance screening cannot be completed.")
			}

			for restrictedToken, reason := range rule.Restricted {
				if hasToken(ingredients, restrictedToken) {
					isCompliant = false
					flags = append(flags, reason)
				}
			}

			productType := strings.ToLower(strings.TrimSpace(req.ProductType))
			if strings.Contains(productType, "sugar-free") && hasToken(ingredients, "sugar") {
				isCompliant = false
				flags = append(flags, "Product declared as sugar-free but sugar ingredient is present.")
			}

			for _, required := range rule.RequiredItems {
				flags = append(flags, "Required: "+required)
			}

			return &ComplianceCheckResponse{
				IsCompliant: isCompliant,
				Flags:       flags,
				LanguageReq: rule.LabelLanguage,
			}, nil
		})
}
