// order.go 包含 Repo 层共用的排序和搜索辅助函数。
package repo

import (
	"strings"

	"auth-perm/internal/common/utils"
)

// ilikePattern 将关键词包装为 ILIKE 模式：%keyword%，同时转义 SQL 通配符。
// 例如 "AI%芯片" → "%AI\%芯片%"
func ilikePattern(keyword string) string {
	return utils.ILIKEPattern(keyword)
}

// sanitizeOrderBy 验证排序字段是否在白名单中，防止 SQL 注入。
// 输入格式："column" 或 "column desc"/"column asc"。
// 不在白名单中或格式非法时返回默认排序。
func sanitizeOrderBy(input string, allowed map[string]bool, defaultOrder string) string {
	if input == "" {
		return defaultOrder
	}
	parts := strings.Fields(strings.ToLower(input))
	if len(parts) == 0 || len(parts) > 2 {
		return defaultOrder
	}
	col := parts[0]
	if !allowed[col] {
		return defaultOrder
	}
	if len(parts) == 2 {
		dir := parts[1]
		if dir != "asc" && dir != "desc" {
			return defaultOrder
		}
		return col + " " + dir
	}
	return col + " DESC"
}

// themeOrderAllowlist 主题列表允许的排序字段
var themeOrderAllowlist = map[string]bool{
	"strength":      true,
	"strength_norm": true,
	"ticker_count":  true,
	"event_count":   true,
	"created_at":    true,
	"updated_at":    true,
	"name":          true,
}

// tickerOrderAllowlist 股票列表允许的排序字段
var tickerOrderAllowlist = map[string]bool{
	"hot_score":     true,
	"mention_count": true,
	"created_at":    true,
	"updated_at":    true,
	"symbol":        true,
	"name":          true,
}

// eventOrderAllowlist 事件列表允许的排序字段
var eventOrderAllowlist = map[string]bool{
	"importance": true,
	"event_time": true,
	"created_at": true,
	"updated_at": true,
	"title":      true,
}
