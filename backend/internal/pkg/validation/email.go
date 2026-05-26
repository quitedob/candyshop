package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var emailValidator = validator.New()

// IsValidEmail 校验邮箱格式，与 Gin binding:"email" 规则一致。
func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	return emailValidator.Var(email, "email") == nil
}
