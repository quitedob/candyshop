package upload

import "testing"

func TestIsAllowedPaymentProof(t *testing.T) {
	cases := []struct {
		mime, name string
		want       bool
	}{
		{"application/pdf", "receipt.pdf", true},
		{"image/jpeg", "slip.jpg", true},
		{"", "proof.png", true},
		{"application/msword", "doc.doc", false},
		{"video/mp4", "clip.mp4", false},
	}
	for _, tc := range cases {
		if got := IsAllowedPaymentProof(tc.mime, tc.name); got != tc.want {
			t.Fatalf("IsAllowedPaymentProof(%q,%q)=%v want %v", tc.mime, tc.name, got, tc.want)
		}
	}
}
