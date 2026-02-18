package cache

import (
	"container/list"
	"context"
	"sync"
	"time"
)

// LRUCache 本地LRU缓存实现
type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex
	stats    *CacheStats
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits         int64     `json:"hits"`
	Misses       int64     `json:"misses"`
	Evictions    int64     `json:"evictions"`
	TotalGets    int64     `json:"total_gets"`
	TotalSets    int64     `json:"total_sets"`
	LastHitTime  time.Time `json:"last_hit_time"`
	LastMissTime time.Time `json:"last_miss_time"`
}

// CacheItem 缓存项
type CacheItem struct {
	key       string
	value     interface{}
	expiresAt time.Time
	createdAt time.Time
	hitCount  int64
}

// NewLRUCache 创建LRU缓存
func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 100 // 默认容量
	}

	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
		stats:    &CacheStats{},
	}
}

// Get 获取缓存值
func (l *LRUCache) Get(ctx context.Context, key string) (interface{}, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.stats.TotalGets++

	element, exists := l.cache[key]
	if !exists {
		l.stats.Misses++
		l.stats.LastMissTime = time.Now()
		return nil, ErrKeyNotFound
	}

	item := element.Value.(*CacheItem)

	// 检查是否过期
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		l.removeElement(element)
		l.stats.Misses++
		l.stats.LastMissTime = time.Now()
		return nil, ErrKeyNotFound
	}

	// 移动到链表前端（最近使用）
	l.list.MoveToFront(element)
	item.hitCount++

	l.stats.Hits++
	l.stats.LastHitTime = time.Now()

	return item.value, nil
}

// Set 设置缓存值
func (l *LRUCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.stats.TotalSets++

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	// 检查键是否已存在
	if element, exists := l.cache[key]; exists {
		// 更新现有项
		item := element.Value.(*CacheItem)
		item.value = value
		item.expiresAt = expiresAt
		item.createdAt = time.Now()

		l.list.MoveToFront(element)
		return nil
	}

	// 创建新项
	item := &CacheItem{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
		createdAt: time.Now(),
		hitCount:  0,
	}

	// 添加到链表前端
	element := l.list.PushFront(item)
	l.cache[key] = element

	// 检查容量限制
	if l.list.Len() > l.capacity {
		l.evict()
	}
	return nil
}

// Delete 删除缓存项
func (l *LRUCache) Delete(ctx context.Context, key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	element, exists := l.cache[key]
	if !exists {
		return nil
	}

	l.removeElement(element)
	return nil
}

// Exists 检查键是否存在
func (l *LRUCache) Exists(ctx context.Context, key string) (bool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	element, exists := l.cache[key]
	if !exists {
		return false, nil
	}

	item := element.Value.(*CacheItem)

	// 检查是否过期
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return false, nil
	}

	return true, nil
}

// Clear 清空缓存
func (l *LRUCache) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cache = make(map[string]*list.Element)
	l.list = list.New()
}

// Size 获取缓存大小
func (l *LRUCache) Size() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.list.Len()
}

// Ping for health checks
func (l *LRUCache) Ping(ctx context.Context) error {
	return nil
}

// Capacity 获取缓存容量
func (l *LRUCache) Capacity() int {
	return l.capacity
}

// Resize 调整缓存容量
func (l *LRUCache) Resize(newCapacity int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if newCapacity <= 0 {
		return
	}

	l.capacity = newCapacity

	// 如果当前大小超过新容量，进行淘汰
	for l.list.Len() > l.capacity {
		l.evict()
	}
}

// Keys 获取所有键
func (l *LRUCache) Keys() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	keys := make([]string, 0, l.list.Len())
	for element := l.list.Front(); element != nil; element = element.Next() {
		item := element.Value.(*CacheItem)
		// 只返回未过期的键
		if item.expiresAt.IsZero() || time.Now().Before(item.expiresAt) {
			keys = append(keys, item.key)
		}
	}

	return keys
}

// GetStats 获取缓存统计信息
func (l *LRUCache) GetStats() CacheStats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return *l.stats
}

// GetHitRate 获取缓存命中率
func (l *LRUCache) GetHitRate() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	total := l.stats.TotalGets
	if total == 0 {
		return 0
	}

	return float64(l.stats.Hits) / float64(total)
}

// CleanExpired 清理过期项
func (l *LRUCache) CleanExpired() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	expired := 0
	var next *list.Element

	for element := l.list.Back(); element != nil; element = next {
		next = element.Prev()
		item := element.Value.(*CacheItem)

		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			l.removeElement(element)
			expired++
		}
	}

	return expired
}

// GetHotKeys 获取热点键（按命中次数排序）
func (l *LRUCache) GetHotKeys(limit int) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	type keyHitCount struct {
		key   string
		count int64
	}

	var keys []keyHitCount
	for element := l.list.Front(); element != nil; element = element.Next() {
		item := element.Value.(*CacheItem)
		keys = append(keys, keyHitCount{
			key:   item.key,
			count: item.hitCount,
		})
	}

	// 按命中次数排序（简单实现，实际应用中可以用更高效的算法）
	if len(keys) > 1 {
		for i := 0; i < len(keys)-1; i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[j].count > keys[i].count {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}
	}

	// 返回前N个
	if limit > 0 && len(keys) > limit {
		keys = keys[:limit]
	}

	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key.key
	}

	return result
}

// GetItemInfo 获取缓存项详细信息
func (l *LRUCache) GetItemInfo(key string) *CacheItemInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()

	element, exists := l.cache[key]
	if !exists {
		return nil
	}

	item := element.Value.(*CacheItem)

	return &CacheItemInfo{
		Key:       item.key,
		CreatedAt: item.createdAt,
		ExpiresAt: item.expiresAt,
		HitCount:  item.hitCount,
		TTL:       getTTL(item.expiresAt),
		Size:      estimateSize(item.value),
	}
}

// CacheItemInfo 缓存项信息
type CacheItemInfo struct {
	Key       string        `json:"key"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`
	HitCount  int64         `json:"hit_count"`
	TTL       time.Duration `json:"ttl"`
	Size      int           `json:"size"`
}

// 私有方法

// removeElement 移除元素
func (l *LRUCache) removeElement(element *list.Element) {
	item := element.Value.(*CacheItem)
	delete(l.cache, item.key)
	l.list.Remove(element)
}

// evict 淘汰最久未使用的项
func (l *LRUCache) evict() {
	if l.list.Len() == 0 {
		return
	}

	element := l.list.Back()
	l.removeElement(element)
	l.stats.Evictions++
}

// getTTL 获取剩余TTL
func getTTL(expiresAt time.Time) time.Duration {
	if expiresAt.IsZero() {
		return -1 // 永不过期
	}

	ttl := time.Until(expiresAt)
	if ttl < 0 {
		return 0 // 已过期
	}

	return ttl
}

// estimateSize 估算值的大小（字节）
func estimateSize(value interface{}) int {
	switch v := value.(type) {
	case string:
		return len(v)
	case []byte:
		return len(v)
	case nil:
		return 0
	default:
		// 简单估算：指针大小 + 基础开销
		return 16
	}
}

// LockStats 获取锁的统计信息（用于调试）
func (l *LRUCache) LockStats() map[string]interface{} {
	return map[string]interface{}{
		"cache_size":  len(l.cache),
		"list_length": l.list.Len(),
		"capacity":    l.capacity,
	}
}

// Clone 创建缓存的副本（用于调试）
func (l *LRUCache) Clone() *LRUCache {
	l.mu.RLock()
	defer l.mu.RUnlock()

	clone := &LRUCache{
		capacity: l.capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
		stats:    &CacheStats{},
	}

	// 复制缓存项
	for element := l.list.Front(); element != nil; element = element.Next() {
		item := element.Value.(*CacheItem)
		newItem := &CacheItem{
			key:       item.key,
			value:     item.value,
			expiresAt: item.expiresAt,
			createdAt: item.createdAt,
			hitCount:  item.hitCount,
		}
		newElement := clone.list.PushBack(newItem)
		clone.cache[newItem.key] = newElement
	}

	// 复制统计信息
	*clone.stats = *l.stats

	return clone
}
