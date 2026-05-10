package dberror

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// IsDuplicateKeyError 检测唯一约束 / 重复键（PostgreSQL + GORM）
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique constraint")
}
