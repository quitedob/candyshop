package admin

import (
	"strings"
	"testing"
)

// TestSanitizeHTMLFields_StripsActiveContentOnlyInNamedFields verifies the
// write-time sanitization used by the AI-translate handlers: active HTML is
// stripped from the named rich-text fields, while plain-text fields are left
// byte-for-byte untouched (so a legitimate '&' is never rewritten to '&amp;').
func TestSanitizeHTMLFields_StripsActiveContentOnlyInNamedFields(t *testing.T) {
	fields := map[string]map[string]string{
		"en": {
			"name":        "Popcorn <script>alert(1)</script>",
			"description": "<p>Fresh <script>alert(1)</script> candy</p>",
			"summary":     "Sweet & fruity",
		},
		"fr": {
			"description": "<p>Bonbons</p>",
		},
	}

	sanitizeHTMLFields(fields, "description")

	enDesc := fields["en"]["description"]
	if strings.Contains(enDesc, "<script") || strings.Contains(enDesc, "alert(1)") {
		t.Fatalf("description still contains active content: %q", enDesc)
	}
	if !strings.Contains(enDesc, "Fresh") || !strings.Contains(enDesc, "candy") {
		t.Fatalf("description lost legit text: %q", enDesc)
	}
	if got := fields["fr"]["description"]; got != "<p>Bonbons</p>" {
		t.Fatalf("clean fr description mangled: %q", got)
	}

	// Non-HTML fields must be left untouched.
	if got := fields["en"]["name"]; got != "Popcorn <script>alert(1)</script>" {
		t.Fatalf("non-html field 'name' was sanitized: %q", got)
	}
	if got := fields["en"]["summary"]; got != "Sweet & fruity" {
		t.Fatalf("non-html field 'summary' was sanitized: %q", got)
	}
}
