package cache

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
	"time"

	"auth-perm/internal/common/constant"
)

// GenerateCacheKey FUTURE: 缓存键生成 - 在实现高级缓存功能时使用
func GenerateCacheKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + ":" + key
}

// GenerateHashKey FUTURE: 哈希键生成 - 在实现分布式缓存时使用
func GenerateHashKey(prefix, key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	hash := hex.EncodeToString(hasher.Sum(nil))

	if prefix == "" {
		return hash
	}
	return prefix + ":" + hash
}

// BuildUserCacheKey FUTURE: 用户缓存键构建 - 在实现用户数据缓存时使用
func BuildUserCacheKey(userID string, suffix string) string {
	if suffix == "" {
		return GenerateCacheKey(constant.CacheKeyUser, userID)
	}
	return GenerateCacheKey(constant.CacheKeyUser, userID+":"+suffix)
}

// BuildSessionCacheKey FUTURE: 会话缓存键构建 - 在实现会话管理时使用
func BuildSessionCacheKey(sessionID string) string {
	return GenerateCacheKey(constant.CacheKeySession, sessionID)
}

// BuildPermissionCacheKey FUTURE: 权限缓存键构建 - 在实现权限缓存时使用
func BuildPermissionCacheKey(userID, orgID, resource string) string {
	key := fmt.Sprintf("perm:%s:%s:%s", userID, orgID, resource)
	return GenerateCacheKey(constant.CacheKeyPermission, key)
}

// BuildRoleCacheKey FUTURE: 角色缓存键构建 - 在实现角色缓存时使用
func BuildRoleCacheKey(userID, orgID string) string {
	key := fmt.Sprintf("roles:%s:%s", userID, orgID)
	return GenerateCacheKey(constant.CacheKeyRole, key)
}

// BuildTokenCacheKey FUTURE: 令牌缓存键构建 - 在实现令牌管理时使用
func BuildTokenCacheKey(tokenHash string) string {
	return GenerateCacheKey("token", tokenHash)
}

// BuildRateLimitKey FUTURE: 限流缓存键构建 - 在实现高级限流时使用
func BuildRateLimitKey(identifier string, window string) string {
	key := fmt.Sprintf("%s:%s", identifier, window)
	return GenerateCacheKey("rate_limit", key)
}

// HashString FUTURE: 字符串哈希 - 在实现数据加密时使用
func HashString(s string) string {
	hasher := fnv.New32a()
	hasher.Write([]byte(s))
	return fmt.Sprintf("%x", hasher.Sum32())
}

// IsValidKey FUTURE: 键有效性检查 - 在实现缓存验证时使用
func IsValidKey(key string) bool {
	if len(key) == 0 || len(key) > 250 {
		return false
	}

	// 检查是否包含无效字符
	invalidChars := regexp.MustCompile(`[^\w\-\.:_@]`)
	return !invalidChars.MatchString(key)
}

// SanitizeKey FUTURE: 键清理 - 在实现缓存安全时使用
func SanitizeKey(key string) string {
	// 替换无效字符为下划线
	reg := regexp.MustCompile(`[^\w\-\.:_@]`)
	sanitized := reg.ReplaceAllString(key, "_")

	// 确保不为空
	if sanitized == "" {
		sanitized = "default"
	}

	// 限制长度
	if len(sanitized) > 250 {
		sanitized = sanitized[:250]
	}

	return sanitized
}

// ParseTTL FUTURE: TTL解析 - 在实现TTL配置时使用
func ParseTTL(ttlStr string) (time.Duration, error) {
	if ttlStr == "" {
		return 0, fmt.Errorf("empty TTL string")
	}

	// 尝试直接解析
	if duration, err := time.ParseDuration(ttlStr); err == nil {
		return duration, nil
	}

	// 解析常见的TTL格式
	switch strings.ToLower(ttlStr) {
	case "never", "0":
		return 0, nil
	case "1min", "1m":
		return time.Minute, nil
	case "5min", "5m":
		return 5 * time.Minute, nil
	case "10min", "10m":
		return 10 * time.Minute, nil
	case "30min", "30m":
		return 30 * time.Minute, nil
	case "1hour", "1h":
		return time.Hour, nil
	case "6hour", "6h":
		return 6 * time.Hour, nil
	case "12hour", "12h":
		return 12 * time.Hour, nil
	case "1day", "1d":
		return 24 * time.Hour, nil
	case "1week", "1w":
		return 7 * 24 * time.Hour, nil
	case "1month", "1M":
		return 30 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported TTL format: %s", ttlStr)
	}
}

// FormatDuration FUTURE: 持续时间格式化 - 在实现时间显示时使用
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "never"
	}

	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}

	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}

	if d < 24*time.Hour {
		return fmt.Sprintf("%.0fh", d.Hours())
	}

	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// CalculateTTL FUTURE: TTL计算 - 在实现动态TTL时使用
func CalculateTTL(priority Priority) time.Duration {
	switch priority {
	case PriorityHigh:
		return time.Hour
	case PriorityMedium:
		return 30 * time.Minute
	case PriorityLow:
		return 10 * time.Minute
	case PriorityDefault:
		return time.Minute
	default:
		return time.Minute
	}
}

// Priority 缓存优先级
type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityDefault
)

// CacheEntry 缓存条目
type CacheEntry struct {
	Key       string        `json:"key"`
	Value     interface{}   `json:"value"`
	TTL       time.Duration `json:"ttl"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`
	HitCount  int64         `json:"hit_count"`
	Size      int           `json:"size"`
}

// NewCacheEntry FUTURE: 缓存条目创建 - 在实现缓存封装时使用
func NewCacheEntry(key string, value interface{}, ttl time.Duration) *CacheEntry {
	now := time.Now()
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = now.Add(ttl)
	}

	return &CacheEntry{
		Key:       key,
		Value:     value,
		TTL:       ttl,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		HitCount:  0,
		Size:      estimateSize(value),
	}
}

// IsExpired 检查缓存条目是否过期
func (e *CacheEntry) IsExpired() bool {
	return !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt)
}

// GetTTL 获取剩余TTL
func (e *CacheEntry) GetTTL() time.Duration {
	if e.ExpiresAt.IsZero() {
		return -1 // 永不过期
	}

	ttl := time.Until(e.ExpiresAt)
	if ttl < 0 {
		return 0 // 已过期
	}

	return ttl
}

// IncrementHitCount 增加命中次数
func (e *CacheEntry) IncrementHitCount() {
	e.HitCount++
}

// MergeStats FUTURE: 统计信息合并 - 在实现缓存统计时使用
func MergeStats(stats1, stats2 CacheStats) CacheStats {
	return CacheStats{
		Hits:         stats1.Hits + stats2.Hits,
		Misses:       stats1.Misses + stats2.Misses,
		Evictions:    stats1.Evictions + stats2.Evictions,
		TotalGets:    stats1.TotalGets + stats2.TotalGets,
		TotalSets:    stats1.TotalSets + stats2.TotalSets,
		LastHitTime:  maxTime(stats1.LastHitTime, stats2.LastHitTime),
		LastMissTime: maxTime(stats1.LastMissTime, stats2.LastMissTime),
	}
}

// GetHitRate 获取命中率
func (s *CacheStats) GetHitRate() float64 {
	total := s.TotalGets
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total)
}

// maxTime FUTURE: 时间比较 - 在实现时间处理时使用
func maxTime(t1, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}

// SplitKeysByPrefix FUTURE: 键前缀分割 - 在实现缓存管理时使用
func SplitKeysByPrefix(keys []string) map[string][]string {
	result := make(map[string][]string)

	for _, key := range keys {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) >= 2 {
			prefix := parts[0]
			result[prefix] = append(result[prefix], key)
		} else {
			result["default"] = append(result["default"], key)
		}
	}

	return result
}

// FilterKeysByPattern FUTURE: 键模式过滤 - 在实现缓存查询时使用
func FilterKeysByPattern(keys []string, pattern string) []string {
	if pattern == "" {
		return keys
	}

	var result []string
	reg, err := regexp.Compile(pattern)
	if err != nil {
		// 如果正则表达式无效，使用字符串匹配
		for _, key := range keys {
			if strings.Contains(key, pattern) {
				result = append(result, key)
			}
		}
		return result
	}

	for _, key := range keys {
		if reg.MatchString(key) {
			result = append(result, key)
		}
	}

	return result
}

// BuildCacheKeyPattern FUTURE: 缓存键模式构建 - 在实现缓存搜索时使用
func BuildCacheKeyPattern(prefix string, pattern string) string {
	if prefix == "" {
		return pattern
	}
	if pattern == "" {
		return prefix + ":*"
	}
	return prefix + ":" + pattern
}

// ExtractKeyPrefix FUTURE: 键前缀提取 - 在实现缓存分析时使用
func ExtractKeyPrefix(key string) string {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) >= 2 {
		return parts[0]
	}
	return "default"
}

// GetCacheKeyType FUTURE: 缓存键类型获取 - 在实现缓存分类时使用
func GetCacheKeyType(key string) string {
	prefix := ExtractKeyPrefix(key)
	switch prefix {
	case "user":
		return "user"
	case "session":
		return "session"
	case "permission":
		return "permission"
	case "role":
		return "role"
	case "token":
		return "token"
	case "rate_limit":
		return "rate_limit"
	default:
		return "unknown"
	}
}
