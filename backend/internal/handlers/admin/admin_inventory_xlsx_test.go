package admin

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	modelsProduct "candypro/api/internal/models/product"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// H11 follow-up regression tests (arguer round). The round-4 importer migration
// switched AdminApplyInventoryImport to UpdateProductColumns keyed by "row key
// present" — but the frontend applyImport (pages/admin/inventory/index.vue)
// sends every mapped cell as a string, so a blank cell arrives as "" and a
// present key was written as 0/"" (base_price=0, stock_quantity=0, blanked
// lead_time/…), the same data-wipe failure class that refuted round 1, now at
// per-row granularity. presentProductColumns now carries a column only when the
// cell holds a meaningful value: blank cells and unparseable formatted values
// ("$12.50", "1,234.50") are skipped (stored value preserved), while an
// explicit zero ("0", "false") is still written (H11 stays closed).

func runApplyImport(t *testing.T, db *gorm.DB, body string) map[string]interface{} {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, w := productUpdateTestCtx(body)
	h := newProductUpdateTestHandler(t, db)
	h.AdminApplyInventoryImport(c)
	if w.Code != http.StatusOK {
		t.Fatalf("AdminApplyInventoryImport: status=%d, want 200, body=%s", w.Code, w.Body.String())
	}
	var res map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("parse apply response %q: %v", w.Body.String(), err)
	}
	return res
}

func seedInventoryProduct(t *testing.T, db *gorm.DB) {
	t.Helper()
	seedProductUpdateRow(t, db, &modelsProduct.Product{
		ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears",
		Featured: true, HalalCertified: true, StockQuantity: 5,
		BasePrice: 12.5, LeadTime: "2 weeks", HSCode: "1704.90",
		Summary: "fruit gummies", Status: modelsProduct.ProductStatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
}

// Blank cells in numeric columns must NOT be written as zero — pre-fix the row
// key was present and presentProductColumns carried base_price/moq/
// stock_quantity, so the importer wrote 0 over stored values.
func TestPresentProductColumns_SkipsBlankNumericCells(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{
		"basePrice":     "",
		"moq":           "",
		"stockQuantity": "",
	})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	for _, c := range []string{"base_price", "moq", "stock_quantity"} {
		if set[c] {
			t.Fatalf("blank numeric cell leaked column %q into %v", c, cols)
		}
	}
	if len(cols) != 1 || !set["updated_at"] {
		t.Fatalf("expected only updated_at, got %v", cols)
	}
}

// A formatted currency cell ("$12.50", "1,234.50") cannot be parsed by
// toFloat64/toInt and must not wipe the stored value to 0.
func TestPresentProductColumns_SkipsFormattedCurrency(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{
		"basePrice":     "$12.50",
		"stockQuantity": "1,234.50",
	})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	if set["base_price"] || set["stock_quantity"] {
		t.Fatalf("formatted currency cell leaked into %v", cols)
	}
	if len(cols) != 1 || !set["updated_at"] {
		t.Fatalf("expected only updated_at, got %v", cols)
	}
}

// Blank cells in string columns must NOT clear stored text.
func TestPresentProductColumns_SkipsBlankStringCells(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{
		"leadTime": "", "hsCode": "", "shelfLife": "",
		"storage": "", "summary": "",
	})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	for _, c := range []string{"lead_time", "hs_code", "shelf_life", "storage", "summary"} {
		if set[c] {
			t.Fatalf("blank string cell leaked column %q into %v", c, cols)
		}
	}
	if len(cols) != 1 || !set["updated_at"] {
		t.Fatalf("expected only updated_at, got %v", cols)
	}
}

// Explicit zero strings must still be carried so the importer can clear
// stock/price/bools (H11). Guards against over-correcting the skip.
func TestPresentProductColumns_CarriesExplicitZeroString(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{
		"stockQuantity":  "0",
		"basePrice":      "0",
		"halalCertified": "false",
		"leadTime":       "2 weeks",
	})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	for _, c := range []string{"stock_quantity", "base_price", "halal_certified", "lead_time", "updated_at"} {
		if !set[c] {
			t.Fatalf("expected column %q in %v", c, cols)
		}
	}
}

// End-to-end: a partial sheet with blank price/stock/lead-time cells must
// preserve the stored values. Pre-fix this wrote base_price=0, stock_quantity=0
// and blanked lead_time while reporting success.
func TestAdminApplyInventoryImport_BlankCellsPreserveStoredValues(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedInventoryProduct(t, db)

	body := `{"rows":[{"id":"p1","name":"Gummy Bears","stockQuantity":"","basePrice":"","leadTime":""}],"imageColumns":[]}`
	res := runApplyImport(t, db, body)
	if res["errors"] != float64(0) || res["updated"] != float64(1) {
		t.Fatalf("unexpected apply result: %v", res)
	}
	got := mustReloadProduct(t, db, "p1")
	if got.StockQuantity != 5 {
		t.Fatalf("blank stock cell overwrote stored stock: got %d, want 5", got.StockQuantity)
	}
	if got.BasePrice != 12.5 {
		t.Fatalf("blank price cell overwrote stored price: got %v, want 12.5", got.BasePrice)
	}
	if got.LeadTime != "2 weeks" {
		t.Fatalf("blank lead time cell cleared stored value: got %q", got.LeadTime)
	}
	if got.Summary != "fruit gummies" || got.HSCode != "1704.90" {
		t.Fatalf("untouched fields corrupted: %+v", got)
	}
}

// A present-but-formatted price cell ("$12.50") must not wipe the price to 0.
func TestAdminApplyInventoryImport_FormattedPricePreserved(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedInventoryProduct(t, db)

	body := `{"rows":[{"id":"p1","name":"Gummy Bears","basePrice":"$12.50"}],"imageColumns":[]}`
	res := runApplyImport(t, db, body)
	if res["errors"] != float64(0) || res["updated"] != float64(1) {
		t.Fatalf("unexpected apply result: %v", res)
	}
	got := mustReloadProduct(t, db, "p1")
	if got.BasePrice != 12.5 {
		t.Fatalf("formatted price cell wiped stored price: got %v, want 12.5", got.BasePrice)
	}
}

// The H11 gap must stay closed for the importer: an explicit "0" stock cell is
// still written to 0.
func TestAdminApplyInventoryImport_ExplicitZeroPersists(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedInventoryProduct(t, db)

	body := `{"rows":[{"id":"p1","name":"Gummy Bears","stockQuantity":"0"}],"imageColumns":[]}`
	res := runApplyImport(t, db, body)
	if res["errors"] != float64(0) || res["updated"] != float64(1) {
		t.Fatalf("unexpected apply result: %v", res)
	}
	got := mustReloadProduct(t, db, "p1")
	if got.StockQuantity != 0 {
		t.Fatalf("explicit \"0\" stock cell did not persist: got %d", got.StockQuantity)
	}
}

// updated_by audit must be written on importer updates (it was dead before the
// column-scoped migration because the assignment was never carried).
func TestAdminApplyInventoryImport_WritesUpdatedByAudit(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedInventoryProduct(t, db)

	body := `{"rows":[{"id":"p1","name":"Gummy Bears","stockQuantity":"10"}],"imageColumns":[]}`
	res := runApplyImport(t, db, body)
	if res["errors"] != float64(0) || res["updated"] != float64(1) {
		t.Fatalf("unexpected apply result: %v", res)
	}
	got := mustReloadProduct(t, db, "p1")
	if got.UpdatedBy == nil || *got.UpdatedBy != "admin-1" {
		t.Fatalf("importer update did not write updated_by audit: %v", got.UpdatedBy)
	}
}

// ── H11 round 6 (arguer refutation): int-cell predicate must match the writer ──

// intCellMeaningful gates moq/stock_quantity and must use the writer's parser
// (toInt/Atoi). Pre-round-6 numericCellMeaningful used ParseFloat, so a
// decimal/exponent-formatted Excel stock/MOQ cell ("10.5", "100.0", "1e2",
// "12.5") was deemed meaningful while toInt parsed it to 0 and the endpoint
// silently wiped the stored value.
func TestIntCellMeaningful_MatchesToInt(t *testing.T) {
	for _, s := range []string{"10.5", "100.0", "1e2", "12.5", "$12.50", "1,234.50", ""} {
		if intCellMeaningful(s) {
			t.Fatalf("intCellMeaningful(%q) = true, want false (writer toInt would parse to 0)", s)
		}
		if toInt(s) != 0 {
			t.Fatalf("sanity: toInt(%q) = %d, want 0", s, toInt(s))
		}
	}
	for _, s := range []string{"0", "10"} {
		if !intCellMeaningful(s) {
			t.Fatalf("intCellMeaningful(%q) = false, want true (explicit integer/zero carried, H11)", s)
		}
	}
	// Numeric (non-string) types mirror toInt's application/truncation.
	if !intCellMeaningful(float64(10.5)) || !intCellMeaningful(int(7)) {
		t.Fatalf("numeric non-string values should be meaningful")
	}
}

// Decimal-formatted integer cells must NOT leak into the column set.
func TestPresentProductColumns_SkipsDecimalIntegerCells(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{
		"stockQuantity": "10.5",
		"moq":           "100.0",
	})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	if set["stock_quantity"] || set["moq"] {
		t.Fatalf("decimal-formatted integer cell leaked into %v", cols)
	}
	if len(cols) != 1 || !set["updated_at"] {
		t.Fatalf("expected only updated_at, got %v", cols)
	}
}

// Exponent-formatted integer cells ("1e2", Excel Scientific format) must NOT
// leak either.
func TestPresentProductColumns_SkipsExponentIntegerCells(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{
		"stockQuantity": "1e2",
		"moq":           "1e2",
	})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	if set["stock_quantity"] || set["moq"] {
		t.Fatalf("exponent-formatted integer cell leaked into %v", cols)
	}
	if len(cols) != 1 || !set["updated_at"] {
		t.Fatalf("expected only updated_at, got %v", cols)
	}
}

// base_price uses the float writer (toFloat64/ParseFloat), which DOES accept
// decimal/exponent strings — the column is carried because a real value is
// applied, not a wipe.
func TestPresentProductColumns_CarriesDecimalFormattedFloat(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{"basePrice": "100.0"})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	if !set["base_price"] {
		t.Fatalf("float writer accepts \"100.0\"; base_price should be carried, got %v", cols)
	}
}

// A non-string value to a scalar string column (direct API caller) is never
// applied by the writer (it type-asserts string); carrying the column would
// write "" and clear stored text.
func TestPresentProductColumns_SkipsNonStringScalarCell(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{
		"summary":  123,
		"leadTime": 12.5,
	})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	if set["summary"] || set["lead_time"] {
		t.Fatalf("non-string scalar cell leaked into %v", cols)
	}
	if len(cols) != 1 || !set["updated_at"] {
		t.Fatalf("expected only updated_at, got %v", cols)
	}
}

func TestStringCellMeaningful_RejectsNonString(t *testing.T) {
	if stringCellMeaningful(123) || stringCellMeaningful(12.5) || stringCellMeaningful(true) || stringCellMeaningful(nil) {
		t.Fatal("non-string values must not be treated as meaningful scalar text")
	}
	if !stringCellMeaningful("2 weeks") {
		t.Fatal("non-blank string should be meaningful")
	}
	if stringCellMeaningful("") || stringCellMeaningful("   ") {
		t.Fatal("blank/whitespace strings must not be meaningful")
	}
}

// Array columns (images/flavors/shapes/certifications) are written via
// toStringArray: a comma string with a real item or a []interface{} holding a
// string item is meaningful; a blank/empty-part string or a non-string array
// would produce nil and clear the stored list, so it is skipped.
func TestArrayCellMeaningful(t *testing.T) {
	if !arrayCellMeaningful("a, b") || !arrayCellMeaningful("a") {
		t.Fatal("non-blank comma string should be meaningful")
	}
	if arrayCellMeaningful("") || arrayCellMeaningful(" , ") {
		t.Fatal("blank/empty-part string must not be meaningful")
	}
	if !arrayCellMeaningful([]interface{}{"a", "b"}) {
		t.Fatal("[]interface{} with string items should be meaningful")
	}
	if arrayCellMeaningful([]interface{}{123}) || arrayCellMeaningful([]interface{}{}) || arrayCellMeaningful(nil) {
		t.Fatal("array without a string item must not be meaningful")
	}
}

func TestPresentProductColumns_ArrayColumns(t *testing.T) {
	cols := presentProductColumns(map[string]interface{}{
		"images":         "a.jpg, b.jpg",
		"certifications": []interface{}{"ISO"},
		"flavors":        "",
		"shapes":         []interface{}{123},
	})
	set := map[string]bool{}
	for _, c := range cols {
		set[c] = true
	}
	if !set["images"] || !set["certifications"] {
		t.Fatalf("array cells with content should be carried, got %v", cols)
	}
	if set["flavors"] || set["shapes"] {
		t.Fatalf("blank/empty array cells must not be carried, got %v", cols)
	}
}

// End-to-end: a decimal-formatted Excel stock cell ("10.5") must preserve the
// stored stock. Pre-round-6 this wrote stock_quantity=0 over the stored value
// while reporting success.
func TestAdminApplyInventoryImport_DecimalStockCellPreservesStoredValue(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedInventoryProduct(t, db)

	body := `{"rows":[{"id":"p1","name":"Gummy Bears","stockQuantity":"10.5"}],"imageColumns":[]}`
	res := runApplyImport(t, db, body)
	if res["errors"] != float64(0) || res["updated"] != float64(1) {
		t.Fatalf("unexpected apply result: %v", res)
	}
	got := mustReloadProduct(t, db, "p1")
	if got.StockQuantity != 5 {
		t.Fatalf("decimal stock cell overwrote stored stock: got %d, want 5", got.StockQuantity)
	}
}

// End-to-end: a non-string value to a scalar string column (direct API caller)
// must not clear the stored text. Pre-round-6 stringCellMeaningful treated any
// non-string as present, so {"summary":123} carried the column and wrote "".
func TestAdminApplyInventoryImport_NonStringScalarPreservesStoredText(t *testing.T) {
	db := setupProductUpdateTestDB(t)
	seedInventoryProduct(t, db)

	body := `{"rows":[{"id":"p1","name":"Gummy Bears","summary":123}],"imageColumns":[]}`
	res := runApplyImport(t, db, body)
	if res["errors"] != float64(0) || res["updated"] != float64(1) {
		t.Fatalf("unexpected apply result: %v", res)
	}
	got := mustReloadProduct(t, db, "p1")
	if got.Summary != "fruit gummies" {
		t.Fatalf("non-string scalar cleared stored summary: got %q, want %q", got.Summary, "fruit gummies")
	}
}
