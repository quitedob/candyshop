package graph

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	tradeModels "candypro/api/internal/models/trade"
	einotool "candypro/api/internal/pkg/eino/tool"

	"github.com/cloudwego/eino/compose"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// fakePersister records persisted documents and returns a canned PI reference.
type fakePersister struct {
	docs  []*tradeModels.TradeDocument
	piRef string
	err   error
}

func (f *fakePersister) AddDocument(_ context.Context, doc *tradeModels.TradeDocument) error {
	if f.err != nil {
		return f.err
	}
	f.docs = append(f.docs, doc)
	return nil
}

func (f *fakePersister) ResolveProformaInvoiceRef(_ context.Context, _ uint) (string, error) {
	return f.piRef, f.err
}

// runPipeline compiles and invokes the graph, returning the generated docs.
func runPipeline(t *testing.T, persister einotool.DocumentPersister, input DocGenerationInput) DocGenerationOutput {
	t.Helper()
	g, err := NewDocumentPipelineGraph(context.Background(), persister)
	if err != nil {
		t.Fatalf("NewDocumentPipelineGraph: %v", err)
	}
	runner, err := g.Compile(context.Background(), compose.WithGraphName("TestDocPipeline"))
	if err != nil {
		t.Fatalf("compile graph: %v", err)
	}
	out, err := runner.Invoke(context.Background(), input)
	if err != nil {
		t.Fatalf("invoke graph: %v", err)
	}
	return out
}

func findDoc(docs []GeneratedDocInfo, docType string) (GeneratedDocInfo, bool) {
	for _, d := range docs {
		if d.DocType == docType {
			return d, true
		}
	}
	return GeneratedDocInfo{}, false
}

func TestDocumentBatch_CIReferencesSharedPI(t *testing.T) {
	persister := &fakePersister{}
	out := runPipeline(t, persister, DocGenerationInput{
		TradeID:  7,
		DocTypes: []string{tradeModels.DocTypeProformaInvoice, tradeModels.DocTypeCommercialInvoice, tradeModels.DocTypePackingList},
	})
	if out.Status == "failed" {
		t.Fatalf("pipeline failed: %v", out.Errors)
	}

	pi, ok := findDoc(out.GeneratedDocs, tradeModels.DocTypeProformaInvoice)
	if !ok {
		t.Fatal("PI not generated")
	}
	ci, ok := findDoc(out.GeneratedDocs, tradeModels.DocTypeCommercialInvoice)
	if !ok {
		t.Fatal("CI not generated")
	}

	// PI and CI must share the same batch suffix (same Date + Seq).
	piSuffix := pi.DocNumber[len("PI-"):]
	ciSuffix := ci.DocNumber[len("CI-"):]
	if piSuffix != ciSuffix {
		t.Fatalf("PI/CI suffixes differ within a batch: PI=%s CI=%s", pi.DocNumber, ci.DocNumber)
	}

	// The CI's embedded pi_ref must equal the actual PI doc number.
	var ciContent struct {
		PiRef string `json:"pi_ref"`
	}
	if err := json.Unmarshal([]byte(ci.Content), &ciContent); err != nil {
		t.Fatalf("unmarshal CI content: %v", err)
	}
	if ciContent.PiRef != pi.DocNumber {
		t.Fatalf("CI pi_ref %q != PI doc number %q", ciContent.PiRef, pi.DocNumber)
	}

	// Persisted docs must match the generated numbers.
	if len(persister.docs) != 3 {
		t.Fatalf("expected 3 persisted docs, got %d", len(persister.docs))
	}
	for _, pd := range persister.docs {
		if pd.DocNumber == "" {
			t.Fatalf("persisted doc %s has empty doc_number", pd.Type)
		}
	}
}

func TestDocumentBatch_CIWithoutPIResolvesExistingRef(t *testing.T) {
	persister := &fakePersister{piRef: "PI-20260811-00042"}
	out := runPipeline(t, persister, DocGenerationInput{
		TradeID:  9,
		DocTypes: []string{tradeModels.DocTypeCommercialInvoice},
	})

	ci, ok := findDoc(out.GeneratedDocs, tradeModels.DocTypeCommercialInvoice)
	if !ok {
		t.Fatal("CI not generated")
	}
	var ciContent struct {
		PiRef string `json:"pi_ref"`
	}
	if err := json.Unmarshal([]byte(ci.Content), &ciContent); err != nil {
		t.Fatalf("unmarshal CI content: %v", err)
	}
	if ciContent.PiRef != "PI-20260811-00042" {
		t.Fatalf("CI pi_ref %q != existing PI ref", ciContent.PiRef)
	}
}

func TestDocumentBatch_CIWithoutPIAndNoRefStaysEmpty(t *testing.T) {
	persister := &fakePersister{piRef: ""}
	out := runPipeline(t, persister, DocGenerationInput{
		TradeID:  10,
		DocTypes: []string{tradeModels.DocTypeCommercialInvoice},
	})

	ci, ok := findDoc(out.GeneratedDocs, tradeModels.DocTypeCommercialInvoice)
	if !ok {
		t.Fatal("CI not generated")
	}
	var ciContent struct {
		PiRef string `json:"pi_ref"`
	}
	if err := json.Unmarshal([]byte(ci.Content), &ciContent); err != nil {
		t.Fatalf("unmarshal CI content: %v", err)
	}
	if ciContent.PiRef != "" {
		t.Fatalf("expected empty pi_ref, got %q", ciContent.PiRef)
	}
}

// TestNewBatchDocSeq_Uniqueness asserts the collision-resistance invariant the
// old modulo scheme (seq := ms % 100000) violated: two batches in the same
// millisecond, or exactly 100,000 ms apart on the same day (residue wraps),
// must produce distinct Seq values. FAILS on the pre-fix path (both inputs
// returned the same residue).
func TestNewBatchDocSeq_Uniqueness(t *testing.T) {
	// Same millisecond: the residue scheme returned an identical Seq for both.
	now := time.UnixMilli(123456789)
	a := NewBatchDocSeq(now)
	b := NewBatchDocSeq(now)
	if a.Seq == b.Seq {
		t.Fatalf("same-ms batches produced identical Seq %d", a.Seq)
	}
	if a.Date != b.Date {
		t.Fatalf("same-ms batches disagree on Date: %q vs %q", a.Date, b.Date)
	}

	// Exactly 100,000 ms apart on the same day (both residues are 0, which the
	// old scheme coerced to Seq 1).
	c := NewBatchDocSeq(time.UnixMilli(1700000000000))
	d := NewBatchDocSeq(time.UnixMilli(1700000100000))
	if c.Seq == d.Seq {
		t.Fatalf("100s-apart same-day batches produced identical Seq %d", c.Seq)
	}
	if c.Date != d.Date {
		t.Fatalf("100s-apart batches must share Date, got %q vs %q", c.Date, d.Date)
	}

	// Seq stays positive and stays date-correlated with the clock input.
	if a.Seq < 1 {
		t.Fatalf("expected Seq >= 1, got %d", a.Seq)
	}
	if a.Date != now.Format("20060102") {
		t.Fatalf("unexpected Date %q", a.Date)
	}
}

func TestBatchContainsDocType(t *testing.T) {
	cases := []struct {
		types []string
		want  bool
	}{
		{[]string{"PROFORMA_INVOICE", "COMMERCIAL_INVOICE"}, true},
		{[]string{"proforma_invoice"}, true},
		{[]string{"PACKING_LIST"}, false},
		{nil, false},
	}
	for _, c := range cases {
		if got := batchContainsDocType(c.types, tradeModels.DocTypeProformaInvoice); got != c.want {
			t.Fatalf("batchContainsDocType(%v) = %v, want %v", c.types, got, c.want)
		}
	}
}

// setupTradeDocDB opens an in-memory sqlite DB with the TradeDocument table
// (including the global uniqueIndex on doc_number) migrated, mirroring the
// repo-layer test harness used across the codebase.
func setupTradeDocDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&tradeModels.TradeDocument{}); err != nil {
		t.Fatalf("migrate TradeDocument: %v", err)
	}
	return db
}

// sqlitePersister is a minimal sqlite-backed DocumentPersister that reproduces
// the real TradeService.AddDocument path (Create -> UNIQUE constraint on
// doc_number) without importing services/trade (which would create an import
// cycle through pkg/eino).
type sqlitePersister struct {
	db *gorm.DB
}

func (p *sqlitePersister) AddDocument(ctx context.Context, doc *tradeModels.TradeDocument) error {
	return p.db.WithContext(ctx).Create(doc).Error
}

func (p *sqlitePersister) ResolveProformaInvoiceRef(ctx context.Context, transactionID uint) (string, error) {
	var docs []tradeModels.TradeDocument
	if err := p.db.WithContext(ctx).Where("transaction_id = ?", transactionID).Find(&docs).Error; err != nil {
		return "", err
	}
	for i := range docs {
		if docs[i].Type == tradeModels.DocTypeProformaInvoice {
			return docs[i].DocNumber, nil
		}
	}
	return "", nil
}

// TestDocumentBatch_CollisionResistantDocNumbersPersist drives two batch
// identities that the old modulo scheme collided on (same-day instants 100 s
// apart, both residues 0 -> Seq 1) through a real sqlite-backed persister and
// asserts they persist without tripping TradeDocument's global uniqueIndex.
// FAILS on the pre-fix path: generateDocument produced identical DocNumbers and
// the second AddDocument hit a UNIQUE constraint error.
func TestDocumentBatch_CollisionResistantDocNumbersPersist(t *testing.T) {
	db := setupTradeDocDB(t)
	svc := &sqlitePersister{db: db}

	a := NewBatchDocSeq(time.UnixMilli(1700000000000))
	b := NewBatchDocSeq(time.UnixMilli(1700000100000)) // +100 s, same date, same residue pre-fix

	piA, err := generateDocument(context.Background(), tradeModels.DocTypeProformaInvoice, DocGenerationInput{TradeID: 1}, a, "")
	if err != nil {
		t.Fatalf("generateDocument A: %v", err)
	}
	piB, err := generateDocument(context.Background(), tradeModels.DocTypeProformaInvoice, DocGenerationInput{TradeID: 2}, b, "")
	if err != nil {
		t.Fatalf("generateDocument B: %v", err)
	}
	if piA.DocNumber == piB.DocNumber {
		t.Fatalf("collision-resistant seq failed: both batches produced %q", piA.DocNumber)
	}

	docs := []*tradeModels.TradeDocument{
		{TransactionID: 1, Type: piA.DocType, DocNumber: piA.DocNumber, Status: tradeModels.TradeStatusDraft, IsAIGenerated: true, LineageSource: "ai_draft_graph"},
		{TransactionID: 2, Type: piB.DocType, DocNumber: piB.DocNumber, Status: tradeModels.TradeStatusDraft, IsAIGenerated: true, LineageSource: "ai_draft_graph"},
	}
	for _, d := range docs {
		if err := svc.AddDocument(context.Background(), d); err != nil {
			t.Fatalf("AddDocument(%q): %v", d.DocNumber, err)
		}
	}
	var n int64
	if err := db.Model(&tradeModels.TradeDocument{}).Count(&n).Error; err != nil || n != 2 {
		t.Fatalf("expected 2 persisted docs, got %d (err %v)", n, err)
	}
}

// TestDocumentBatch_CIPolAndBLPortsComeFromInput verifies the CI's pol/pod and
// the BL's port_of_loading/port_of_discharge are rendered from the actual port
// fields rather than the Incoterms term / a hardcoded literal.
func TestDocumentBatch_CIPolAndBLPortsComeFromInput(t *testing.T) {
	persister := &fakePersister{}
	out := runPipeline(t, persister, DocGenerationInput{
		TradeID:           11,
		DocTypes:          []string{tradeModels.DocTypeCommercialInvoice, tradeModels.DocTypeBillOfLading},
		Incoterms:         "CIF Rotterdam",
		PortOfLoading:     "Qingdao",
		PortOfDestination: "Hamburg",
	})
	if out.Status == "failed" {
		t.Fatalf("pipeline failed: %v", out.Errors)
	}

	ci, ok := findDoc(out.GeneratedDocs, tradeModels.DocTypeCommercialInvoice)
	if !ok {
		t.Fatal("CI not generated")
	}
	var ciContent struct {
		Pol       string `json:"pol"`
		Pod       string `json:"pod"`
		Incoterms string `json:"incoterms"`
	}
	if err := json.Unmarshal([]byte(ci.Content), &ciContent); err != nil {
		t.Fatalf("unmarshal CI content: %v", err)
	}
	if ciContent.Pol != "Qingdao" {
		t.Fatalf("CI pol = %q, want %q (previously populated from incoterms)", ciContent.Pol, "Qingdao")
	}
	if ciContent.Pod != "Hamburg" {
		t.Fatalf("CI pod = %q, want %q", ciContent.Pod, "Hamburg")
	}
	if ciContent.Incoterms != "CIF Rotterdam" {
		t.Fatalf("CI incoterms = %q, want %q (the term must stay on incoterms)", ciContent.Incoterms, "CIF Rotterdam")
	}

	bl, ok := findDoc(out.GeneratedDocs, tradeModels.DocTypeBillOfLading)
	if !ok {
		t.Fatal("BL not generated")
	}
	var blContent struct {
		PortOfLoading   string `json:"port_of_loading"`
		PortOfDischarge string `json:"port_of_discharge"`
	}
	if err := json.Unmarshal([]byte(bl.Content), &blContent); err != nil {
		t.Fatalf("unmarshal BL content: %v", err)
	}
	if blContent.PortOfLoading != "Qingdao" {
		t.Fatalf("BL port_of_loading = %q, want %q", blContent.PortOfLoading, "Qingdao")
	}
	if blContent.PortOfDischarge != "Hamburg" {
		t.Fatalf("BL port_of_discharge = %q, want %q", blContent.PortOfDischarge, "Hamburg")
	}
}

// TestDocumentBatch_PortsDefaultWhenAbsent verifies the defaults applied in
// PrepareContext: an absent port still never leaves the CI pol holding the
// Incoterms term.
func TestDocumentBatch_PortsDefaultWhenAbsent(t *testing.T) {
	persister := &fakePersister{}
	out := runPipeline(t, persister, DocGenerationInput{
		TradeID:   12,
		DocTypes:  []string{tradeModels.DocTypeCommercialInvoice},
		Incoterms: "FOB Shenzhen",
	})
	ci, ok := findDoc(out.GeneratedDocs, tradeModels.DocTypeCommercialInvoice)
	if !ok {
		t.Fatal("CI not generated")
	}
	var ciContent struct {
		Pol       string `json:"pol"`
		Pod       string `json:"pod"`
		Incoterms string `json:"incoterms"`
	}
	if err := json.Unmarshal([]byte(ci.Content), &ciContent); err != nil {
		t.Fatalf("unmarshal CI content: %v", err)
	}
	if ciContent.Pol != "Shanghai" {
		t.Fatalf("default CI pol = %q, want %q (not the incoterms term)", ciContent.Pol, "Shanghai")
	}
	if ciContent.Pol == ciContent.Incoterms {
		t.Fatalf("CI pol still carries the incoterms term %q", ciContent.Pol)
	}
	if ciContent.Pod != "Destination Port" {
		t.Fatalf("default CI pod = %q, want %q", ciContent.Pod, "Destination Port")
	}
}
