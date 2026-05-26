package request

import "testing"

func TestFirstNonEmpty(t *testing.T) {
	if got := FirstNonEmpty("", "  ", "abc", "def"); got != "abc" {
		t.Fatalf("FirstNonEmpty = %q, want abc", got)
	}
	if got := FirstNonEmpty("", ""); got != "" {
		t.Fatalf("FirstNonEmpty = %q, want empty", got)
	}
}

func TestFirstOptionalNonEmpty(t *testing.T) {
	empty := ""
	paid := "paid"
	if got := FirstOptionalNonEmpty(nil, &empty, &paid); got == nil || *got != "paid" {
		t.Fatalf("FirstOptionalNonEmpty = %v, want paid", got)
	}
	if got := FirstOptionalNonEmpty(nil, nil); got != nil {
		t.Fatalf("FirstOptionalNonEmpty = %v, want nil", got)
	}
}
