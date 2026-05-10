package uploadpath

import (
	"fmt"
	"path/filepath"
	"strings"
)

// LocalPathFromUploadURL 将对外路径 /uploads/... 转为本地磁盘路径并防止路径穿越。
func LocalPathFromUploadURL(uploadRoot string, uploadURL string) (string, error) {
	uploadRoot = strings.TrimSpace(uploadRoot)
	if uploadRoot == "" {
		uploadRoot = "./uploads"
	}
	u := strings.TrimSpace(uploadURL)
	if u == "" {
		return "", fmt.Errorf("empty upload url")
	}
	if !strings.HasPrefix(u, "/uploads/") {
		return "", fmt.Errorf("invalid upload url")
	}
	rel := strings.TrimPrefix(u, "/uploads/")
	rel = filepath.FromSlash(rel)
	if rel == "" || rel == "." || strings.Contains(rel, "..") {
		return "", fmt.Errorf("invalid path")
	}
	full := filepath.Join(uploadRoot, rel)
	absRoot, err1 := filepath.Abs(uploadRoot)
	absFull, err2 := filepath.Abs(full)
	if err1 != nil || err2 != nil {
		return "", fmt.Errorf("path resolution failed")
	}
	sep := string(filepath.Separator)
	if absFull != absRoot && !strings.HasPrefix(absFull, absRoot+sep) {
		return "", fmt.Errorf("path outside upload root")
	}
	return full, nil
}
