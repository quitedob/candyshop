package storage

import (
	"path/filepath"
	"strings"
)

// scriptExecutableExtensions are formats that browsers render inline and that can
// carry executable content: SVG can embed <script>, and the HTML family runs
// markup. net/http serves these inline (no Content-Disposition: attachment), so
// a stored-XSS payload would execute in the browser. Files with these extensions
// must never be accepted even if they appear in the extension/content-type
// whitelists (which live in local.go/s3.go), so the guard lives here instead.
var scriptExecutableExtensions = map[string]bool{
	".svg":   true,
	".htm":   true,
	".html":  true,
	".xhtml": true,
	".shtml": true,
}

// normalizeContentType 去掉 charset 等后缀。
func normalizeContentType(ct string) string {
	ct = strings.TrimSpace(strings.ToLower(ct))
	if idx := strings.Index(ct, ";"); idx != -1 {
		ct = strings.TrimSpace(ct[:idx])
	}
	return ct
}

// contentTypeAllowed 校验 magic byte 检测结果；对 Office 等误报类型按扩展名兜底。
// 脚本可执行格式（.svg/.html 等）无条件拒绝，且扩展名白名单不再为任意检测类型兜底。
func contentTypeAllowed(detectedType, filename string) bool {
	detectedType = normalizeContentType(detectedType)
	ext := strings.ToLower(filepath.Ext(filename))

	// Defense-in-depth: reject script-executable formats outright, regardless of
	// filename or detected type. A whitelisted extension alone must never grant
	// .svg (served inline as image/svg+xml) or .html a pass, and S3 stores the
	// detected ContentType, so an image/svg+xml / text/html object would be
	// served inline even when the filename extension looks safe.
	if scriptExecutableExtensions[ext] ||
		detectedType == "image/svg+xml" ||
		strings.HasPrefix(detectedType, "text/html") {
		return false
	}

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

	// CSV/文本类：仅接受 text/plain 或 text/csv，避免 text/html 载荷以 .txt/.csv 绕过。
	if (detectedType == "text/plain" || detectedType == "text/csv") &&
		(ext == ".csv" || ext == ".txt") {
		return true
	}

	// 兜底：仅在 magic byte 无法识别（octet-stream）且扩展名在白名单时放行；
	// 不再允许"扩展名 + 任意检测类型"蒙混（例如伪装成 .png 的 HTML 载荷）。
	if detectedType == "application/octet-stream" && allowedExtensions[ext] {
		return true
	}

	return false
}
