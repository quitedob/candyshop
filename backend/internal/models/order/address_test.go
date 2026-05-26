package order

import (
	"encoding/json"
	"testing"
)

func TestAddressUnmarshalJSON_ZipAlias(t *testing.T) {
	var addr Address
	if err := json.Unmarshal([]byte(`{"street":"123 Main","city":"New York","state":"NY","zip":"10001","country":"US"}`), &addr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if addr.ZipCode != "10001" {
		t.Fatalf("ZipCode = %q, want 10001", addr.ZipCode)
	}
}

func TestAddressUnmarshalJSON_ZipCodePreferred(t *testing.T) {
	var addr Address
	payload := `{"street":"123 Main","city":"New York","state":"NY","zipCode":"10002","zip":"10001","country":"US"}`
	if err := json.Unmarshal([]byte(payload), &addr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if addr.ZipCode != "10002" {
		t.Fatalf("ZipCode = %q, want 10002", addr.ZipCode)
	}
}
