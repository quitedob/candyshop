package product

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestSupplierAPIKey_DisclosedToOwner is the G23 regression: the self-service
// API key was tagged json:"-" so the registrant never learned their key in the
// self-registration response (or from their profile). It must be serialized so
// the owning supplier can retrieve it. It is still omitted when empty (e.g. an
// admin-created supplier that has no key yet), and the only endpoints that
// serialize a Supplier are admin endpoints and the owner-only supplier-portal
// endpoints, so disclosure is scoped to owner + admin.
func TestSupplierAPIKey_DisclosedToOwner(t *testing.T) {
	sup := Supplier{ID: "SUP-1", Name: "Acme", Email: "acme@example.com", APIKey: "deadbeefcafe"}
	raw, err := json.Marshal(sup)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"apiKey":"deadbeefcafe"`) {
		t.Fatalf("API key not disclosed in JSON: %s", raw)
	}

	// An unset key stays hidden so partial/legacy records do not emit a bogus
	// empty apiKey field.
	empty := Supplier{ID: "SUP-2", Name: "NoKey"}
	raw2, err := json.Marshal(empty)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw2), `"apiKey"`) {
		t.Fatalf("empty API key should be omitted, got: %s", raw2)
	}
}
