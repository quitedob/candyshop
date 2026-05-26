package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"candypro/api/internal/pkg/timeutil"

	"golang.org/x/crypto/bcrypt"
)

// sha256Sum 是 sha256.Sum256 的小封装，便于在不引入额外别名的情况下使用。
func sha256Sum(b []byte) [32]byte {
	return sha256.Sum256(b)
}

// GenerateID generates a cryptographically secure random ID
func GenerateID() string {
	b := make([]byte, 16)
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
	return fmt.Sprintf("%s_%s%s", timeutil.Now().Format("20060102_150405"), GenerateRandomString(8), ext)
}

// FormatResetToken 组合 userID 与随机段，便于 bcrypt 校验后定位用户。
//
// Deprecated: 仅保留以兼容尚未消费完的老链接（M-22）。新调用方应使用
// 不含 userID 的随机令牌 + DeriveResetTokenLookupKey 进行索引查找。
func FormatResetToken(userID, randomPart string) string {
	return userID + "." + randomPart
}

// ParseResetTokenUserID 从组合令牌中解析 userID（兼容老格式）。
//
// Deprecated: 见 FormatResetToken。新格式不带 userID，调用方应改走
// PasswordResetTokenRepository.FindActiveByLookupKey。
func ParseResetTokenUserID(token string) (string, bool) {
	token = strings.TrimSpace(token)
	idx := strings.Index(token, ".")
	if idx <= 0 || idx+1 >= len(token) {
		return "", false
	}
	return token[:idx], true
}

// DeriveResetTokenLookupKey 从明文令牌派生稳定的查表键。
//
// 使用 SHA-256 的前 32 hex 字符（128 bit），确保：
//   1. 数据库可以走唯一索引快速命中（O(1)）；
//   2. 即使 lookup_key 泄露也无法反推出原始令牌（仍需通过 bcrypt 校验）；
//   3. 同一令牌在不同请求中产生相同的 lookup_key（必要的可重复性）。
//
// 注意：这是查表键，不是身份验证凭证；攻击者拿到 lookup_key 仍需碰撞
// 出原始令牌的 bcrypt 哈希才能完成重置（M-22）。
func DeriveResetTokenLookupKey(token string) string {
	sum := sha256Sum([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:16])
}

// HashResetToken 使用 bcrypt 哈希重置令牌（含盐，防彩虹表）。
func HashResetToken(token string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(token), 12)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyResetToken 校验明文令牌与 bcrypt 哈希是否匹配。
func VerifyResetToken(token, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(token)) == nil
}
