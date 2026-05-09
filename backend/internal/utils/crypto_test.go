package utils

import (
	"testing"
)

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()

	if id1 == "" {
		t.Error("GenerateID returned empty string")
	}

	if id1 == id2 {
		t.Error("GenerateID returned duplicate IDs")
	}

	if len(id1) != 32 {
		t.Errorf("GenerateID returned ID with unexpected length: %d", len(id1))
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"short", 5},
		{"medium", 16},
		{"long", 32},
		{"zero", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRandomString(tt.length)
			if len(result) != tt.length {
				t.Errorf("GenerateRandomString length = %d, want %d", len(result), tt.length)
			}

			// Generate another to check uniqueness
			result2 := GenerateRandomString(tt.length)
			if tt.length > 0 && result == result2 {
				t.Error("GenerateRandomString returned duplicate strings")
			}
		})
	}
}

func TestGenerateSlug(t *testing.T) {
	slug := GenerateSlug()

	if len(slug) != 8 {
		t.Errorf("GenerateSlug length = %d, want 8", len(slug))
	}
}

func TestGenerateUniqueFilename(t *testing.T) {
	filename := GenerateUniqueFilename(".jpg")

	if filename == "" {
		t.Error("GenerateUniqueFilename returned empty string")
	}

	if len(filename) < 10 {
		t.Errorf("GenerateUniqueFilename returned too short filename: %s", filename)
	}

	// Check extension
	ext := filename[len(filename)-4:]
	if ext != ".jpg" {
		t.Errorf("GenerateUniqueFilename extension = %s, want .jpg", ext)
	}
}
