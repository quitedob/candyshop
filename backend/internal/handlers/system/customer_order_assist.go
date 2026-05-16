package system

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/kyb"
	"candypro/api/internal/pkg/response"
	tradeSvc "candypro/api/internal/services/trade"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type customerOrderAssistRequest struct {
	Prompt                 string              `json:"prompt" binding:"required"`
	TargetCountry          string              `json:"targetCountry" binding:"required"`
	Quantity               int                 `json:"quantity" binding:"required,min=1"`
	Budget                 float64             `json:"budget"`
	Currency               string              `json:"currency"`
	TaxAmount              float64             `json:"taxAmount"`
	ShippingAmount         float64             `json:"shippingAmount"`
	ShippingAddress        modelsOrder.Address `json:"shippingAddress"`
	AdditionalRequirements string              `json:"additionalRequirements"`
	InquiryID              *string             `json:"inquiryId"`
}


// CustomerAIAssistOrder selects products via AI, creates a draft order, and requires user confirmation.
func (h *Handler) CustomerAIAssistOrder(c *gin.Context) {
	if h.aiUnavailable(c) {
		return
	}
	if h.services == nil || h.services.Order == nil || h.services.Search == nil || h.services.Product == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, _ := authContext(c)
	if userID == "" {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req customerOrderAssistRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	if req.TaxAmount < 0 || req.ShippingAmount < 0 {
		response.InvalidResp(c, "tax_shipping_negative")
		return
	}
	if msg := validateAssistShippingAddress(req.ShippingAddress); msg != "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	targetCountry := strings.TrimSpace(req.TargetCountry)
	if targetCountry == "" {
		response.InvalidResp(c, "target_country_required")
		return
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	queryParts := []string{
		strings.TrimSpace(req.Prompt),
		strings.TrimSpace(req.AdditionalRequirements),
		targetCountry,
	}
	searchQuery := strings.TrimSpace(strings.Join(filterNonEmpty(queryParts), " "))
	searchRes, err := h.services.Search.Search(c.Request.Context(), searchQuery, "products", 8)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_product_search_failed")
		return
	}
	if len(searchRes.Products) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":          "No products matched your request. Try broadening your description or adjusting the target country.",
			"order":            nil,
			"selectedProducts": []any{},
			"ai":               gin.H{"summary": "", "missingInformation": []string{}},
			"compliance":       gin.H{"targetCountry": targetCountry, "checklist": baselineComplianceChecklist(targetCountry), "warnings": []string{}, "mustConfirm": false},
			"inventory":        gin.H{"warnings": []string{}},
			"suggestions": []string{
				"Use broader product terms (e.g. 'gummy candy' instead of a specific brand name)",
				"Check spelling of product names or categories",
				"Remove very specific requirements and try again",
			},
		})
		return
	}

	inquiryID, inquiryErr := h.validateAssistInquiry(c, req.InquiryID, userID)
	if inquiryErr != nil {
		return
	}

	candidateJSON, err := json.Marshal(buildCandidatePayload(searchRes.Products))
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_prepare_payload_failed")
		return
	}

	draft, analyzeErr := h.aiService.OrderAssistAnalyze(c.Request.Context(), tradeSvc.OrderAssistParams{
		Prompt:                 req.Prompt,
		TargetCountry:          targetCountry,
		Quantity:               req.Quantity,
		Budget:                 req.Budget,
		Currency:               currency,
		AdditionalRequirements: req.AdditionalRequirements,
		ShippingCountry:        req.ShippingAddress.Country,
		CandidateJSON:          string(candidateJSON),
	})
	if analyzeErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_draft_generate_failed")
		return
	}

	items, selectedProducts, selectedProductModels, selectionWarnings := buildAssistOrderItems(draft.RecommendedProducts, searchRes.Products, req.Quantity)
	if len(items) == 0 {
		response.ErrorResp(c, http.StatusBadRequest, "no_valid_order_items")
		return
	}

	complianceValidation := h.services.Product.ValidateComplianceWithMarketProfiles(c.Request.Context(), targetCountry, selectedProductModels)
	if len(complianceValidation.Violations) > 0 {
		response.ErrorRespDetail(c, http.StatusUnprocessableEntity, "order_draft_blocked_compliance", gin.H{
			"country":    complianceValidation.Country,
			"violations": complianceValidation.Violations,
			"warnings":   complianceValidation.Warnings,
		})
		return
	}
	productByID := make(map[string]modelsProduct.Product, len(selectedProductModels))
	for _, product := range selectedProductModels {
		productByID[product.ID] = product
	}
	omsIDs := make([]string, 0, len(selectedProductModels))
	for _, p := range selectedProductModels {
		if id := strings.TrimSpace(p.ID); id != "" {
			omsIDs = append(omsIDs, id)
		}
	}
	sellableOMS, _ := h.services.Product.EffectiveSellableByProducts(c.Request.Context(), omsIDs, modelsProduct.ChannelWebstore)
	inventoryValidation := h.services.Order.ValidateInventoryWithSellable(items, productByID, sellableOMS)
	if len(inventoryValidation.Violations) > 0 {
		response.ErrorRespDetail(c, http.StatusUnprocessableEntity, "order_draft_blocked_inventory", gin.H{
			"violations": inventoryValidation.Violations,
			"warnings":   inventoryValidation.Warnings,
		})
		return
	}

	complianceLookup := h.aiService.LookupCompliance(
		targetCountry,
		buildComplianceQuery(targetCountry, selectedProductModels),
		5,
	)

	referencePayload := make([]gin.H, 0)
	retrievalSummary := ""
	hasOfficialEvidence := false
	if complianceLookup != nil {
		retrievalSummary = complianceLookup.Answer
		for _, ref := range complianceLookup.References {
			if ref.Official {
				hasOfficialEvidence = true
			}
			referencePayload = append(referencePayload, gin.H{
				"source":   ref.Source,
				"section":  ref.Section,
				"snippet":  ref.Snippet,
				"score":    ref.Score,
				"url":      ref.URL,
				"official": ref.Official,
			})
		}
	}

	subtotal := 0.0
	for _, item := range items {
		subtotal += float64(item.Quantity) * item.UnitPrice
	}

	now := time.Now()
	order := &modelsOrder.Order{
		ID:                         crypto.GenerateID(),
		OrderNumber:                fmt.Sprintf("AID-%s-%s", now.Format("20060102"), strings.ToUpper(crypto.GenerateSlug())),
		UserID:                     userID,
		InquiryID:                  inquiryID,
		Status:                     "pending_confirmation",
		PaymentStatus:              "unpaid",
		Items:                      items,
		StockReserved:              false,
		ComplianceOfficialEvidence: hasOfficialEvidence,
		Subtotal:                   subtotal,
		TaxAmount:                  req.TaxAmount,
		ShippingAmount:             req.ShippingAmount,
		TotalAmount:                subtotal + req.TaxAmount + req.ShippingAmount,
		Currency:                   currency,
		ShippingAddress: modelsOrder.Address{
			Street:  strings.TrimSpace(req.ShippingAddress.Street),
			City:    strings.TrimSpace(req.ShippingAddress.City),
			State:   strings.TrimSpace(req.ShippingAddress.State),
			ZipCode: strings.TrimSpace(req.ShippingAddress.ZipCode),
			Country: strings.TrimSpace(req.ShippingAddress.Country),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if !h.ensureActiveOrKYBBypassAmount(c, userID, order.TotalAmount, kyb.LineProductIDs(items)...) {
		return
	}

	if err := h.services.Order.CreateOrder(c.Request.Context(), order); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_draft_create_failed")
		return
	}

	warnings := make([]string, 0, 8)
	warnings = append(warnings, selectionWarnings...)
	warnings = append(warnings, complianceValidation.Warnings...)
	warnings = append(warnings, inventoryValidation.Warnings...)
	if req.Budget > 0 && order.TotalAmount > req.Budget {
		warnings = append(warnings, fmt.Sprintf("Estimated total %.2f %s exceeds budget %.2f %s.", order.TotalAmount, currency, req.Budget, currency))
	}

	complianceChecklist := append([]string{}, baselineComplianceChecklist(targetCountry)...)
	complianceChecklist = append(complianceChecklist, draft.ComplianceChecklist...)
	complianceChecklist = append(complianceChecklist, draft.RequiredCertificates...)
	complianceChecklist = dedupeNonEmpty(complianceChecklist)

	c.JSON(http.StatusCreated, gin.H{
		"message":          "AI order draft created. Please review and confirm before processing.",
		"order":            order,
		"selectedProducts": selectedProducts,
		"ai": gin.H{
			"summary":            strings.TrimSpace(draft.Summary),
			"missingInformation": dedupeNonEmpty(draft.MissingInformation),
		},
		"compliance": gin.H{
			"targetCountry":       targetCountry,
			"country":             complianceValidation.Country,
			"checklist":           complianceChecklist,
			"warnings":            dedupeNonEmpty(append(draft.Warnings, warnings...)),
			"paymentPolicy":       complianceValidation.Payment,
			"references":          referencePayload,
			"retrievalNote":       retrievalSummary,
			"hasOfficialEvidence": hasOfficialEvidence,
			"mustConfirm":         true,
		},
		"inventory": gin.H{"warnings": inventoryValidation.Warnings},
	})
}

func (h *Handler) validateAssistInquiry(c *gin.Context, rawInquiryID *string, userID string) (*string, error) {
	if rawInquiryID == nil || strings.TrimSpace(*rawInquiryID) == "" {
		return nil, nil
	}

	inquiryID := strings.TrimSpace(*rawInquiryID)
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return nil, err
	}
	if inquiry.UserID == nil || *inquiry.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "inquiry_no_access")
		return nil, fmt.Errorf("inquiry does not belong to current user")
	}
	return &inquiryID, nil
}

func buildCandidatePayload(products []modelsProduct.Product) []gin.H {
	payload := make([]gin.H, 0, len(products))
	for _, product := range products {
		payload = append(payload, gin.H{
			"id":             product.ID,
			"slug":           product.Slug,
			"name":           product.Name,
			"summary":        product.Summary,
			"moq":            product.MOQ,
			"stockQuantity":  product.StockQuantity,
			"leadTime":       product.LeadTime,
			"category":       product.Category,
			"certifications": product.Certifications,
			"halalCertified": product.HalalCertified,
		})
	}
	return payload
}

func buildAssistOrderItems(recommended []tradeSvc.OrderAssistItem, candidates []modelsProduct.Product, requestedQuantity int) (modelsOrder.OrderItemArray, []gin.H, []modelsProduct.Product, []string) {
	candidateByID := make(map[string]modelsProduct.Product, len(candidates))
	for _, product := range candidates {
		candidateByID[product.ID] = product
	}

	items := make(modelsOrder.OrderItemArray, 0, len(recommended))
	selectedProducts := make([]gin.H, 0, len(recommended))
	selectedModels := make([]modelsProduct.Product, 0, len(recommended))
	warnings := make([]string, 0, 4)
	seen := make(map[string]struct{}, len(recommended))

	for _, rec := range recommended {
		productID := strings.TrimSpace(rec.ProductID)
		if productID == "" {
			continue
		}
		if _, ok := seen[productID]; ok {
			continue
		}

		product, ok := candidateByID[productID]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("AI selected unknown productId: %s (ignored).", productID))
			continue
		}

		qty := rec.Quantity
		if qty < requestedQuantity {
			qty = requestedQuantity
		}
		if product.MOQ > 0 && qty < product.MOQ {
			qty = product.MOQ
		}
		if qty < 1 {
			qty = 1
		}

		price := rec.UnitPrice
		if price < 0 {
			price = 0
			warnings = append(warnings, fmt.Sprintf("AI returned negative unitPrice for product %s; reset to 0.", productID))
		}

		items = append(items, modelsOrder.OrderItem{
			ProductID:      productID,
			Quantity:       qty,
			UnitPrice:      price,
			Specifications: strings.TrimSpace(rec.Specifications),
		})

		selectedProducts = append(selectedProducts, gin.H{
			"id":             product.ID,
			"name":           product.Name,
			"slug":           product.Slug,
			"quantity":       qty,
			"unitPrice":      price,
			"moq":            product.MOQ,
			"stockQuantity":  product.StockQuantity,
			"leadTime":       product.LeadTime,
			"reason":         strings.TrimSpace(rec.Reason),
			"specifications": strings.TrimSpace(rec.Specifications),
		})
		selectedModels = append(selectedModels, product)
		seen[productID] = struct{}{}
	}

	if len(items) > 0 {
		return items, selectedProducts, selectedModels, warnings
	}

	// Deterministic fallback: pick the first matched product.
	product := candidates[0]
	qty := requestedQuantity
	if product.MOQ > 0 && qty < product.MOQ {
		qty = product.MOQ
	}
	if qty < 1 {
		qty = 1
	}

	items = append(items, modelsOrder.OrderItem{
		ProductID:      product.ID,
		Quantity:       qty,
		UnitPrice:      0,
		Specifications: "AI fallback selection; unit price pending quotation confirmation.",
	})
	selectedProducts = append(selectedProducts, gin.H{
		"id":             product.ID,
		"name":           product.Name,
		"slug":           product.Slug,
		"quantity":       qty,
		"unitPrice":      0,
		"moq":            product.MOQ,
		"stockQuantity":  product.StockQuantity,
		"leadTime":       product.LeadTime,
		"reason":         "Fallback selection based on relevance and MOQ constraints.",
		"specifications": "Please confirm unit price and final specs before payment.",
	})
	selectedModels = append(selectedModels, product)
	warnings = append(warnings, "AI recommendations were empty. Applied deterministic fallback product selection.")
	return items, selectedProducts, selectedModels, warnings
}

func validateAssistShippingAddress(address modelsOrder.Address) string {
	if strings.TrimSpace(address.Street) == "" {
		return "shippingAddress.street is required"
	}
	if strings.TrimSpace(address.City) == "" {
		return "shippingAddress.city is required"
	}
	if strings.TrimSpace(address.Country) == "" {
		return "shippingAddress.country is required"
	}
	return ""
}

func filterNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func dedupeNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func baselineComplianceChecklist(country string) []string {
	normalized := strings.ToLower(strings.TrimSpace(country))

	switch normalized {
	case "usa", "us", "united states", "united states of america":
		return []string{
			"Ensure importer workflow includes FDA food facility registration and prior notice handling.",
			"Label in English with Nutrition Facts and mandatory allergen declaration.",
			"Keep ingredient list and traceability records aligned with FSMA preventive control expectations.",
		}
	case "eu", "europe", "european union", "germany", "france", "netherlands", "belgium", "italy", "spain", "poland":
		return []string{
			"Apply destination-language labeling under Regulation (EU) No 1169/2011.",
			"Highlight allergens and list additives by functional class and E-number.",
			"Verify restricted additives for target market; for example E171 is not authorized in EU food.",
		}
	case "japan", "jp":
		return []string{
			"Prepare Japanese label text compliant with Food Labeling Standards (Consumer Affairs Agency).",
			"Ensure import notification package is complete for Food Sanitation Act screening (MHLW quarantine).",
			"Confirm additive usage and allergen declarations are acceptable for Japan before shipment.",
		}
	case "saudi arabia", "saudi", "ksa":
		return []string{
			"Verify SFDA import requirements and product registration workflow before shipment.",
			"Use Arabic-compliant labeling for retail packs and include importer details.",
			"Confirm halal compliance and avoid prohibited non-halal ingredients or alcohol carriers.",
		}
	case "india", "in", "bharat":
		return []string{
			"Prepare India-compliant labeling and dossier under FSSAI import requirements before shipment.",
			"Confirm ingredient/additive disclosure and allergen declaration are complete for customs review.",
			"Apply risk-control payment policy: full prepayment required before production scheduling.",
		}
	case "pakistan", "pk":
		return []string{
			"Confirm product falls within applicable PSQCA/IPO mandatory standards scope before export.",
			"Use halal-compliant formulation and certificate pack for market entry and buyer compliance checks.",
			"Apply risk-control payment policy: full prepayment required before production scheduling.",
		}
	default:
		return []string{
			"Validate destination-market import registration, labeling language, and allergen declarations.",
			"Confirm product formula/additives are allowed in target country before production lock.",
			"Ensure certificate package is complete (origin, health/safety docs, and importer records).",
		}
	}
}

// buildComplianceQuery generates a dynamic compliance query based on actual product attributes.
// This replaces the hardcoded keyword string so the RAG lookup is product-specific.
func buildComplianceQuery(country string, products []modelsProduct.Product) string {
	parts := []string{"food import labeling registration"}

	// Add product-specific terms
	hasHalal := false
	categories := make(map[string]struct{})
	for _, p := range products {
		if p.HalalCertified {
			hasHalal = true
		}
		if p.Category != "" {
			categories[strings.ToLower(p.Category)] = struct{}{}
		}
		if p.Allergens != "" {
			parts = append(parts, "allergens declaration")
			break
		}
		if p.Ingredients != "" {
			parts = append(parts, "additives ingredients")
			break
		}
	}

	if hasHalal {
		parts = append(parts, "halal certification requirements")
	}
	for cat := range categories {
		parts = append(parts, cat)
	}

	// Add country-specific regulatory terms
	switch strings.ToLower(strings.TrimSpace(country)) {
	case "usa", "us", "united states":
		parts = append(parts, "FDA FSMA prior notice nutrition facts")
	case "eu", "europe", "european union":
		parts = append(parts, "EU 1169/2011 food labeling E-numbers additives")
	case "japan", "jp":
		parts = append(parts, "MHLW food sanitation act import notification")
	case "saudi arabia", "saudi", "ksa":
		parts = append(parts, "SFDA import registration halal")
	case "india", "in":
		parts = append(parts, "FSSAI import license food safety")
	case "pakistan", "pk":
		parts = append(parts, "PSQCA food standards halal")
	default:
		parts = append(parts, "customs clearance food safety certificate")
	}

	return strings.Join(dedupeNonEmpty(parts), " ")
}
