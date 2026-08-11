package storage

import "testing"

// TestContentTypeAllowed_RejectsScriptExecutable formats that browsers render
// inline must be rejected even though .svg is on the extension/content-type
// whitelists in local.go/s3.go, and regardless of what the magic-byte detector
// reports (S3 stores the detected ContentType, so an image/svg+xml or text/html
// object is served inline even under a safe-looking filename).
func TestContentTypeAllowed_RejectsScriptExecutable(t *testing.T) {
	cases := []struct {
		detected, filename string
	}{
		{"image/svg+xml", "logo.svg"},
		{"text/xml", "icon.svg"},
		{"application/octet-stream", "logo.svg"},
		{"text/html", "page.html"},
		{"text/html", "evil.htm"},
		// SVG/HTML content smuggled under a safe extension.
		{"image/svg+xml", "photo.png"},
		{"text/html", "notes.txt"},
	}
	for _, tc := range cases {
		if got := contentTypeAllowed(tc.detected, tc.filename); got {
			t.Fatalf("contentTypeAllowed(%q,%q)=true, want false", tc.detected, tc.filename)
		}
	}
}

// TestContentTypeAllowed_NoBlindExtensionFallback verifies the extension
// whitelist no longer approves a mismatched detected type (an HTML payload named
// .png previously passed the `allowedExtensions[ext]` fallback).
func TestContentTypeAllowed_NoBlindExtensionFallback(t *testing.T) {
	if got := contentTypeAllowed("text/html", "photo.png"); got {
		t.Fatal("contentTypeAllowed(text/html, photo.png)=true, want false (HTML payload named .png)")
	}
	if got := contentTypeAllowed("application/x-executable", "script.png"); got {
		t.Fatal("contentTypeAllowed(application/x-executable, script.png)=true, want false")
	}
	if got := contentTypeAllowed("text/xml", "data.pdf"); got {
		t.Fatal("contentTypeAllowed(text/xml, data.pdf)=true, want false")
	}

	// Indeterminate magic (octet-stream) with a whitelisted safe extension still
	// passes so truncated/odd files are not outright rejected.
	if got := contentTypeAllowed("application/octet-stream", "odd.png"); !got {
		t.Fatal("contentTypeAllowed(octet-stream, odd.png)=false, want true (indeterminate magic fallback)")
	}
	// CSV is commonly detected as text/plain, not text/csv.
	if got := contentTypeAllowed("text/plain", "list.csv"); !got {
		t.Fatal("contentTypeAllowed(text/plain, list.csv)=false, want true")
	}
}
