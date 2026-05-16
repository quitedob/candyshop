package trade

import (
	"context"
	"fmt"
	"strings"
)

// GenerateQuotationText generates a B2B candy OEM quotation draft.
// costStackInjection is optional pre-formatted cost stack text appended to the prompt.
func (s *AIService) GenerateQuotationText(ctx context.Context, prompt, currency, inquiryID, targetCountry, customerRequirements string, costStackInjection string) (string, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "USD"
	}

	if strings.TrimSpace(prompt) == "" {
		prompt = fmt.Sprintf(
			"Generate a B2B candy OEM quotation draft in %s for inquiry %s. Target country: %s. Requirements: %s. Include pricing assumptions, MOQ, lead time, payment terms, validity and exclusions.",
			currency,
			strings.TrimSpace(inquiryID),
			strings.TrimSpace(targetCountry),
			strings.TrimSpace(customerRequirements),
		)
	}

	if costStackInjection != "" {
		prompt += "\n" + costStackInjection
	}

	return s.Generate(ctx, prompt)
}
