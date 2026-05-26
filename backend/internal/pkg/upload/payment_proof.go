package upload

import (
	"path/filepath"
	"strings"
)

const PaymentProofMaxBytes = 10 * 1024 * 1024 // 10MB

// paymentProofMIME 银行/付款凭证允许的 MIME。
var paymentProofMIME = map[string]struct{}{
	"application/pdf": {},
	"image/jpeg":      {},
	"image/png":       {},
}

// paymentProofExt 扩展名兜底。
var paymentProofExt = map[string]struct{}{
	".pdf":  {},
	".jpg":  {},
	".jpeg": {},
	".png":  {},
}

// IsAllowedPaymentProof 校验付款凭证文件类型（PDF / JPG / PNG）。
func IsAllowedPaymentProof(contentType, filename string) bool {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	if contentType != "" {
		if _, ok := paymentProofMIME[contentType]; ok {
			return true
		}
	}
	ext := strings.ToLower(filepath.Ext(filename))
	_, ok := paymentProofExt[ext]
	return ok
}
