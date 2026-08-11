package graph

import (
	"encoding/json"
	"strings"
	"testing"

	tradeModels "candypro/api/internal/models/trade"
)

// TestDocumentBatch_SanitizesUntrustedInputFields drives HTML-injected string
// fields through the pipeline and asserts the active content never reaches the
// generated (or persisted) document content. Fails on the pre-fix path where
// PrepareContext passed the raw values through to fmt.Sprintf.
func TestDocumentBatch_SanitizesUntrustedInputFields(t *testing.T) {
	persister := &fakePersister{}
	out := runPipeline(t, persister, DocGenerationInput{
		TradeID:           13,
		DocTypes:          []string{tradeModels.DocTypeProformaInvoice},
		BuyerName:         "Alice <script>alert(1)</script>",
		SellerName:        "CandyPro <img src=x onerror=alert(1)>",
		PaymentTerms:      "T/T <b>30 days</b>",
		PortOfLoading:     "Shanghai <script>alert(2)</script>",
		PortOfDestination: "Hamburg",
	})
	if out.Status == "failed" {
		t.Fatalf("pipeline failed: %v", out.Errors)
	}

	pi, ok := findDoc(out.GeneratedDocs, tradeModels.DocTypeProformaInvoice)
	if !ok {
		t.Fatal("PI not generated")
	}
	var m struct {
		Buyer        string `json:"buyer"`
		Seller       string `json:"seller"`
		PaymentTerms string `json:"payment_terms"`
	}
	if err := json.Unmarshal([]byte(pi.Content), &m); err != nil {
		t.Fatalf("unmarshal PI content: %v", err)
	}
	for field, v := range map[string]string{"buyer": m.Buyer, "seller": m.Seller} {
		if strings.Contains(v, "<script") || strings.Contains(v, "onerror") {
			t.Fatalf("PI %s still contains active HTML: %q", field, v)
		}
	}
	if !strings.Contains(m.Buyer, "Alice") {
		t.Fatalf("PI content lost legit buyer text: %s", pi.Content)
	}
	// Harmless inline formatting is preserved, not stripped.
	if m.PaymentTerms != "T/T <b>30 days</b>" {
		t.Fatalf("payment_terms = %q, want inline <b> preserved", m.PaymentTerms)
	}

	// The persisted document must carry the same sanitized content.
	if len(persister.docs) != 1 {
		t.Fatalf("expected 1 persisted doc, got %d", len(persister.docs))
	}
	raw := string(persister.docs[0].Content)
	if strings.Contains(raw, "<script") || strings.Contains(raw, "onerror") {
		t.Fatalf("persisted content still active: %s", raw)
	}
}
