// 东方财富 API 共用的辅助函数。
// 从 map 中提取值、secid 转换等通用逻辑，供 client / f10_client / news_client 共用。
package eastmoney

import (
	"fmt"
	"strconv"
	"strings"
)

// getString 从 map 中安全提取字符串值。
func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// getInt 从 map 中安全提取整数值。
func getInt(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}

// getFloat 从 map 中安全提取浮点值。
func getFloat(m map[string]any, key string) float64 {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// secidToCode 从 secid 提取纯代码：1.600519 → 600519
func secidToCode(secid string) string {
	if idx := strings.Index(secid, "."); idx >= 0 {
		return secid[idx+1:]
	}
	return secid
}
