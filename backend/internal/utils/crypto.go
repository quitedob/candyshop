package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// GenerateID generates a cryptographically secure random ID
func GenerateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// SEC-18: crypto/rand failure is critical — do not fall back to predictable values
		panic(fmt.Sprintf("CRITICAL: crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b)
}

// GenerateRandomString generates a cryptographically secure random string
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range b {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			// SEC-18: crypto/rand failure is critical — do not fall back to predictable values
			panic(fmt.Sprintf("CRITICAL: crypto/rand failed: %v", err))
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// GenerateSlug generates a URL-safe slug from a random string
func GenerateSlug() string {
	return GenerateRandomString(8)
}

// GenerateUniqueFilename generates a unique filename
func GenerateUniqueFilename(ext string) string {
	return fmt.Sprintf("%s_%s%s", Now().Format("20060102_150405"), GenerateRandomString(8), ext)
}

// HashResetToken returns a SHA-256 hex digest of a password reset token.
// Tokens are stored hashed so a database leak does not expose usable reset links.
func HashResetToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
