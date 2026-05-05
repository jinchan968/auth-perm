package constant

import "strings"

// BuildKey 构建缓存键：prefix + strings.Join(parts, ":")
func BuildKey(prefix string, parts ...string) string {
	if len(parts) == 0 {
		return prefix
	}
	return prefix + strings.Join(parts, ":")
}
