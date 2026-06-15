package storage

import (
	"os"
	"strings"
)

func (fs *FileSystem) GetETag(path string) string {
	b, err := os.ReadFile(path + ".etag")
	if err != nil {
		return ""
	}
	return "\"" + string(b) + "\""
}

func isInternalFile(key string) bool {
	return strings.HasSuffix(key, ".etag") ||
		strings.HasSuffix(key, ".tmp") ||
		strings.Contains(key, ".tmp.")
}
