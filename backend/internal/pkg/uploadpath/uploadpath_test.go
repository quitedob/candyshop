package uploadpath

import (
	"path/filepath"
	"testing"
)

func TestLocalPathFromUploadURL(t *testing.T) {
	root := t.TempDir()
	p, err := LocalPathFromUploadURL(root, "/uploads/a/b.pdf")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "a", "b.pdf")
	if filepath.Clean(p) != filepath.Clean(want) {
		t.Fatalf("got %q want %q", p, want)
	}
	if _, err := LocalPathFromUploadURL(root, "/etc/passwd"); err == nil {
		t.Fatal("expected error for invalid url")
	}
}
