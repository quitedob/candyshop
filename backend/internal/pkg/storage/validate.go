package storage

import (
	"path/filepath"
	"strings"
)

// normalizeContentType 去掉 charset 等后缀。
func normalizeContentType(ct string) string {
	ct = strings.TrimSpace(strings.ToLower(ct))
	if idx := strings.Index(ct, ";"); idx != -1 {
		ct = strings.TrimSpace(ct[:idx])
	}
	return ct
}

// contentTypeAllowed 校验 magic byte 检测结果；对 Office 等误报类型按扩展名兜底。
func contentTypeAllowed(detectedType, filename string) bool {
	detectedType = normalizeContentType(detectedType)
	ext := strings.ToLower(filepath.Ext(filename))

	if allowedImageTypes[detectedType] || allowedDocTypes[detectedType] || allowedMediaTypes[detectedType] {
		return true
	}

	// OpenXML（docx/xlsx）常被识别为 application/zip
	if detectedType == "application/zip" && (ext == ".docx" || ext == ".xlsx") {
		return true
	}

	// 旧版 .doc 常为 octet-stream
	if (detectedType == "application/octet-stream" || detectedType == "application/x-msdownload") &&
		(ext == ".doc" || ext == ".xls") {
		return true
	}

	// CSV 等文本类
	if strings.HasPrefix(detectedType, "text/") && (ext == ".csv" || ext == ".txt") {
		return true
	}

	// 扩展名已在白名单且非 executable 类 octet-stream 时兜底
	if allowedExtensions[ext] && detectedType != "application/x-executable" {
		return true
	}

	return false
}
