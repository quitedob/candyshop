package order

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"strings"
	"testing"
)

func TestValidateCountryCompliance_EUBlocksE171(t *testing.T) {
	products := []modelsProduct.Product{
		{
			ID:          "p-eu-1",
			Name:        "Color Candy",
			Ingredients: "Sugar, Titanium Dioxide (E171), Flavor",
		},
	}

	result := validateCountryCompliance("EU", products)
	if result.Country != "eu" {
		t.Fatalf("unexpected canonical country: got %q want %q", result.Country, "eu")
	}
	if len(result.Violations) == 0 {
		t.Fatalf("expected at least one violation for E171")
	}

	joined := strings.ToLower(strings.Join(result.Violations, " | "))
	if !strings.Contains(joined, "e171") {
		t.Fatalf("expected E171 violation text, got: %v", result.Violations)
	}
}

func TestValidateCountryCompliance_SaudiRequiresHalalAndNoPorcine(t *testing.T) {
	products := []modelsProduct.Product{
		{
			ID:             "p-sa-1",
			Name:           "Gel Candy",
			Ingredients:    "Sugar, gelatin (porcine), flavor",
			HalalCertified: false,
			Certifications: modelsCommon.StringArray{"HACCP"},
		},
	}

	result := validateCountryCompliance("KSA", products)
	if result.Country != "saudi arabia" {
		t.Fatalf("unexpected canonical country: got %q want %q", result.Country, "saudi arabia")
	}
	if len(result.Violations) < 2 {
		t.Fatalf("expected at least two violations for Saudi checks, got: %v", result.Violations)
	}

	violationsText := strings.ToLower(strings.Join(result.Violations, " | "))
	if !strings.Contains(violationsText, "non-halal") && !strings.Contains(violationsText, "prohibited") {
		t.Fatalf("expected prohibited ingredient violation, got: %v", result.Violations)
	}
	if !strings.Contains(violationsText, "not marked as halal certified") {
		t.Fatalf("expected halal certification violation, got: %v", result.Violations)
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("expected warning for missing explicit halal certificate field")
	}
}

func TestValidateCountryCompliance_USCyclamateViolationAndLabelWarnings(t *testing.T) {
	products := []modelsProduct.Product{
		{
			ID:          "p-us-1",
			Name:        "Diet Candy",
			Ingredients: "Cyclamate, flavor",
			Allergens:   "",
		},
		{
			ID:          "p-us-2",
			Name:        "Plain Candy",
			Ingredients: "",
			Allergens:   "",
		},
	}

	result := validateCountryCompliance("USA", products)
	if result.Country != "usa" {
		t.Fatalf("unexpected canonical country: got %q want %q", result.Country, "usa")
	}
	if len(result.Violations) == 0 {
		t.Fatalf("expected cyclamate violation, got none")
	}

	violationsText := strings.ToLower(strings.Join(result.Violations, " | "))
	if !strings.Contains(violationsText, "cyclamate") {
		t.Fatalf("expected cyclamate violation text, got: %v", result.Violations)
	}
	if len(result.Warnings) < 2 {
		t.Fatalf("expected warnings for missing ingredient/allergen declarations, got: %v", result.Warnings)
	}
}

func TestValidateCountryCompliance_IndiaRequiresFullPrepaymentPolicy(t *testing.T) {
	products := []modelsProduct.Product{
		{
			ID:          "p-in-1",
			Name:        "India Candy",
			Ingredients: "Sugar, pectin",
			Allergens:   "none",
		},
	}

	result := validateCountryCompliance("India", products)
	if result.Country != "india" {
		t.Fatalf("unexpected canonical country: got %q want %q", result.Country, "india")
	}
	if !result.Payment.RequiresFullPrepayment {
		t.Fatalf("expected India to require full prepayment policy")
	}
	if len(result.Payment.AllowedTerms) == 0 {
		t.Fatalf("expected allowed payment terms for India policy")
	}
}

func TestValidateCountryCompliance_PakistanHalalAndPrepaymentPolicy(t *testing.T) {
	products := []modelsProduct.Product{
		{
			ID:             "p-pk-1",
			Name:           "PK Candy",
			Ingredients:    "Sugar, gelatin (porcine), flavor",
			HalalCertified: false,
			Certifications: modelsCommon.StringArray{"HACCP"},
		},
	}

	result := validateCountryCompliance("Pakistan", products)
	if result.Country != "pakistan" {
		t.Fatalf("unexpected canonical country: got %q want %q", result.Country, "pakistan")
	}
	if !result.Payment.RequiresFullPrepayment {
		t.Fatalf("expected Pakistan to require full prepayment policy")
	}
	if len(result.Violations) == 0 {
		t.Fatalf("expected Pakistan halal violations")
	}
}
