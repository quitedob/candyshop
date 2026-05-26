package upload

import (
	"path/filepath"
	"strings"
)

// inquiryExtraMIME 询价附件额外允许的 MIME（在配置 AllowedTypes 之外）。
var inquiryExtraMIME = map[string]struct{}{
	"application/msword": {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
	"application/pdf": {},
	"text/csv":              {},
	"application/csv":       {},
	"text/comma-separated-values": {},
	"application/vnd.ms-excel": {},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": {},
	"image/png":  {},
	"image/jpeg": {},
	"video/x-matroska": {},
	"video/mp4": {},
}

// inquiryExtToMIME 扩展名兜底（浏览器未上报 Content-Type 时）。
var inquiryExtToMIME = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".csv":  "text/csv",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".png":  "image/png",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
	".mkv":  "video/x-matroska",
	".mp4":  "video/mp4",
}

// IsAllowedInquiryAttachment 校验询价附件类型（配置白名单 + 扩展名兜底）。
func IsAllowedInquiryAttachment(contentType, filename string, cfgAllowed []string) bool {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	if contentType != "" {
		for _, t := range cfgAllowed {
			if strings.EqualFold(t, contentType) {
				return true
			}
		}
		if _, ok := inquiryExtraMIME[contentType]; ok {
			return true
		}
	}
	ext := strings.ToLower(filepath.Ext(filename))
	_, ok := inquiryExtToMIME[ext]
	return ok
}
