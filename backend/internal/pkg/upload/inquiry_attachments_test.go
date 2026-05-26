package upload

import "testing"

func TestIsAllowedInquiryAttachment(t *testing.T) {
	cfg := []string{"image/jpeg", "image/png", "application/pdf"}

	cases := []struct {
		mime     string
		filename string
		want     bool
	}{
		{"application/pdf", "spec.pdf", true},
		{"", "spec.docx", true},
		{"", "data.csv", true},
		{"", "sheet.xlsx", true},
		{"video/mp4", "demo.mp4", true},
		{"", "clip.mkv", true},
		{"application/zip", "archive.zip", false},
	}

	for _, tc := range cases {
		if got := IsAllowedInquiryAttachment(tc.mime, tc.filename, cfg); got != tc.want {
			t.Fatalf("IsAllowedInquiryAttachment(%q, %q) = %v, want %v", tc.mime, tc.filename, got, tc.want)
		}
	}
}
