package storage

import "testing"

func TestContentTypeAllowed(t *testing.T) {
	cases := []struct {
		detected, filename string
		want               bool
	}{
		{"application/pdf", "a.pdf", true},
		{"application/zip", "spec.docx", true},
		{"application/octet-stream", "legacy.doc", true},
		{"video/mp4", "demo.mp4", true},
		{"application/zip", "evil.zip", false},
	}
	for _, tc := range cases {
		if got := contentTypeAllowed(tc.detected, tc.filename); got != tc.want {
			t.Fatalf("contentTypeAllowed(%q,%q)=%v want %v", tc.detected, tc.filename, got, tc.want)
		}
	}
}
