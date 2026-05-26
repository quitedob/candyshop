package eino

import (
	"testing"
)

func TestParseTranslationFieldsJSON_stringValues(t *testing.T) {
	raw := `{"name":"产品名","summary":"摘要"}`
	got, err := ParseTranslationFieldsJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got["name"] != "产品名" || got["summary"] != "摘要" {
		t.Fatalf("unexpected: %#v", got)
	}
}

func TestParseTranslationFieldsJSON_arrayValues(t *testing.T) {
	raw := `{"name":"Gummies","flavors":["Strawberry","Apple"]}`
	got, err := ParseTranslationFieldsJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got["name"] != "Gummies" {
		t.Fatalf("name: %q", got["name"])
	}
	if got["flavors"] != `["Strawberry","Apple"]` {
		t.Fatalf("flavors: %q", got["flavors"])
	}
}

func TestParseTranslationFieldsJSON_translationsWrapper(t *testing.T) {
	raw := `{"translations":{"name":"Name","flavors":["A","B"]}}`
	got, err := ParseTranslationFieldsJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got["name"] != "Name" {
		t.Fatalf("name: %q", got["name"])
	}
	if got["flavors"] != `["A","B"]` {
		t.Fatalf("flavors: %q", got["flavors"])
	}
}

func TestParseTranslationFieldsJSON_markdownBlock(t *testing.T) {
	raw := "```json\n{\"name\":\"Test\"}\n```"
	got, err := ParseTranslationFieldsJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got["name"] != "Test" {
		t.Fatalf("name: %q", got["name"])
	}
}
