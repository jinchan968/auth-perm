package monitoring

import (
	"runtime"
	"sync"
	"time"
)

// Metrics 监控指标
type Metrics struct {
	mu sync.RWMutex

	// HTTP指标
	HTTPRequests       int64            `json:"http_requests"`
	HTTPErrors         int64            `json:"http_errors"`
	HTTPResponseTime   time.Duration    `json:"http_response_time"`
	HTTPRequestsByCode map[string]int64 `json:"http_requests_by_code"`
	HTTPRequestsByPath map[string]int64 `json:"http_requests_by_path"`

	// 业务指标
	UserLogins        int64 `json:"user_logins"`
	UserRegistrations int64 `json:"user_registrations"`
	PasswordResets    int64 `json:"password_resets"`
	OAuthLogins       int64 `json:"oauth_logins"`
	PermissionChecks  int64 `json:"permission_checks"`
	CacheHits         int64 `json:"cache_hits"`
	CacheMisses       int64 `json:"cache_misses"`
	DatabaseQueries   int64 `json:"database_queries"`
	DatabaseErrors    int64 `json:"database_errors"`

	// 系统指标
	StartTime        time.Time `json:"start_time"`
	LastActivityTime time.Time `json:"last_activity_time"`
	GoroutineCount   int64     `json:"goroutine_count"`
	MemoryUsage      int64     `json:"memory_usage_mb"`
	HeapSize         int64     `json:"heap_size_mb"`
	GCCount          uint32    `json:"gc_count"`

	// 认证相关指标
	ActiveSessions     int64 `json:"active_sessions"`
	ExpiredSessions    int64 `json:"expired_sessions"`
	TokenValidations   int64 `json:"token_validations"`
	TokenInvalidations int64 `json:"token_invalidations"`

	// 权限相关指标
	RoleAssignments   int64 `json:"role_assignments"`
	PermissionGranted int64 `json:"permission_granted"`
	PermissionDenied  int64 `json:"permission_denied"`

	// 缓存相关指标
	CacheSize        int64   `json:"cache_size"`
	CacheEvictions   int64   `json:"cache_evictions"`
	CacheHitRate     float64 `json:"cache_hit_rate"`
	RedisConnections int64   `json:"redis_connections"`
	RedisMemoryUsage int64   `json:"redis_memory_usage_mb"`

	// 数据库连接指标
	DBConnections     int64 `json:"db_connections"`
	DBIdleConnections int64 `json:"db_idle_connections"`
	DBOpenConnections int64 `json:"db_open_connections"`
}

// NewMetrics FUTURE: 指标创建 - 在实现指标收集时使用
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime:          time.Now(),
		LastActivityTime:   time.Now(),
		HTTPRequestsByCode: make(map[string]int64),
		HTTPRequestsByPath: make(map[string]int64),
	}
}

// RecordHTTPRequest 记录HTTP请求
func (m *Metrics) RecordHTTPRequest(code int, path string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.HTTPRequests++
	m.HTTPResponseTime = duration
	m.HTTPRequestsByCode[statusCodeToString(code)]++
	m.HTTPRequestsByPath[path]++
	m.LastActivityTime = time.Now()
}

// RecordHTTPError 记录HTTP错误
func (m *Metrics) RecordHTTPError(code int, path string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.HTTPErrors++
	m.HTTPRequestsByCode[statusCodeToString(code)]++
	m.HTTPRequestsByPath[path]++
	m.LastActivityTime = time.Now()
}

// RecordUserLogin 记录用户登录
func (m *Metrics) RecordUserLogin() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UserLogins++
	m.LastActivityTime = time.Now()
}

// RecordUserRegistration 记录用户注册
func (m *Metrics) RecordUserRegistration() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UserRegistrations++
	m.LastActivityTime = time.Now()
}

// RecordPasswordReset 记录密码重置
func (m *Metrics) RecordPasswordReset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PasswordResets++
	m.LastActivityTime = time.Now()
}

// RecordOAuthLogin 记录OAuth登录
func (m *Metrics) RecordOAuthLogin(provider string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.OAuthLogins++
	m.LastActivityTime = time.Now()
}

// RecordPermissionCheck 记录权限检查
func (m *Metrics) RecordPermissionCheck(granted bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PermissionChecks++
	if granted {
		m.PermissionGranted++
	} else {
		m.PermissionDenied++
	}
	m.LastActivityTime = time.Now()
}

// RecordCacheHit 记录缓存命中
func (m *Metrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheHits++
	m.updateCacheHitRate()
}

// RecordCacheMiss 记录缓存未命中
func (m *Metrics) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheMisses++
	m.updateCacheHitRate()
}

// RecordDatabaseQuery 记录数据库查询
func (m *Metrics) RecordDatabaseQuery() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DatabaseQueries++
	m.LastActivityTime = time.Now()
}

// RecordDatabaseError 记录数据库错误
func (m *Metrics) RecordDatabaseError() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DatabaseErrors++
	m.LastActivityTime = time.Now()
}

// RecordActiveSession 记录活跃会话
func (m *Metrics) RecordActiveSession() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ActiveSessions++
}

// RecordExpiredSession 记录过期会话
func (m *Metrics) RecordExpiredSession() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ExpiredSessions++
}

// RecordTokenValidation 记录令牌验证
func (m *Metrics) RecordTokenValidation(valid bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if valid {
		m.TokenValidations++
	} else {
		m.TokenInvalidations++
	}
}

// RecordRoleAssignment 记录角色分配
func (m *Metrics) RecordRoleAssignment() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RoleAssignments++
}

// UpdateSystemMetrics 更新系统指标
func (m *Metrics) UpdateSystemMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新Goroutine数量
	m.GoroutineCount = int64(runtime.NumGoroutine())

	// 更新内存信息
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.MemoryUsage = int64(memStats.Alloc / 1024 / 1024)  // MB
	m.HeapSize = int64(memStats.HeapAlloc / 1024 / 1024) // MB
	m.GCCount = memStats.NumGC
}

// UpdateRedisMetrics 更新Redis指标
func (m *Metrics) UpdateRedisMetrics(connections, memoryUsage int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RedisConnections = connections
	m.RedisMemoryUsage = memoryUsage
}

// UpdateDatabaseMetrics 更新数据库指标
func (m *Metrics) UpdateDatabaseMetrics(idle, open int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DBIdleConnections = idle
	m.DBOpenConnections = open
}

// UpdateCacheMetrics 更新缓存指标
func (m *Metrics) UpdateCacheMetrics(size, evictions int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheSize = size
	m.CacheEvictions = evictions
}

// GetSnapshot 获取指标快照
func (m *Metrics) GetSnapshot() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 深拷贝
	snapshot := &Metrics{
		HTTPRequests:       m.HTTPRequests,
		HTTPErrors:         m.HTTPErrors,
		HTTPResponseTime:   m.HTTPResponseTime,
		UserLogins:         m.UserLogins,
		UserRegistrations:  m.UserRegistrations,
		PasswordResets:     m.PasswordResets,
		OAuthLogins:        m.OAuthLogins,
		PermissionChecks:   m.PermissionChecks,
		CacheHits:          m.CacheHits,
		CacheMisses:        m.CacheMisses,
		DatabaseQueries:    m.DatabaseQueries,
		DatabaseErrors:     m.DatabaseErrors,
		StartTime:          m.StartTime,
		LastActivityTime:   m.LastActivityTime,
		GoroutineCount:     m.GoroutineCount,
		MemoryUsage:        m.MemoryUsage,
		HeapSize:           m.HeapSize,
		GCCount:            m.GCCount,
		ActiveSessions:     m.ActiveSessions,
		ExpiredSessions:    m.ExpiredSessions,
		TokenValidations:   m.TokenValidations,
		TokenInvalidations: m.TokenInvalidations,
		RoleAssignments:    m.RoleAssignments,
		PermissionGranted:  m.PermissionGranted,
		PermissionDenied:   m.PermissionDenied,
		CacheSize:          m.CacheSize,
		CacheEvictions:     m.CacheEvictions,
		CacheHitRate:       m.CacheHitRate,
		RedisConnections:   m.RedisConnections,
		RedisMemoryUsage:   m.RedisMemoryUsage,
		DBConnections:      m.DBConnections,
		DBIdleConnections:  m.DBIdleConnections,
		DBOpenConnections:  m.DBOpenConnections,
	}

	// 深拷贝map
	snapshot.HTTPRequestsByCode = make(map[string]int64)
	snapshot.HTTPRequestsByPath = make(map[string]int64)

	for k, v := range m.HTTPRequestsByCode {
		snapshot.HTTPRequestsByCode[k] = v
	}

	for k, v := range m.HTTPRequestsByPath {
		snapshot.HTTPRequestsByPath[k] = v
	}

	return snapshot
}

// GetUptime 获取运行时间
func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.StartTime)
}

// GetHTTPErrorRate 获取HTTP错误率
func (m *Metrics) GetHTTPErrorRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.HTTPRequests == 0 {
		return 0
	}

	return float64(m.HTTPErrors) / float64(m.HTTPRequests)
}

// GetCacheHitRate 获取缓存命中率
func (m *Metrics) GetCacheHitRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.CacheHits + m.CacheMisses
	if total == 0 {
		return 0
	}

	return float64(m.CacheHits) / float64(total)
}

// Reset 重置指标
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	*m = Metrics{
		StartTime:          time.Now(),
		LastActivityTime:   time.Now(),
		HTTPRequestsByCode: make(map[string]int64),
		HTTPRequestsByPath: make(map[string]int64),
	}
}

// 私有方法

// updateCacheHitRate 更新缓存命中率
func (m *Metrics) updateCacheHitRate() {
	total := m.CacheHits + m.CacheMisses
	if total == 0 {
		m.CacheHitRate = 0
		return
	}

	m.CacheHitRate = float64(m.CacheHits) / float64(total)
}

// statusCodeToString 将状态码转换为字符串
func statusCodeToString(code int) string {
	switch {
	case code < 200:
		return "1xx"
	case code < 300:
		return "2xx"
	case code < 400:
		return "3xx"
	case code < 500:
		return "4xx"
	default:
		return "5xx"
	}
}

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	Collect() *Metrics
	Reset()
}

// DefaultMetricsCollector 默认指标收集器
type DefaultMetricsCollector struct {
	metrics *Metrics
}

// NewDefaultMetricsCollector FUTURE: 默认指标收集器 - 在实现指标收集时使用
func NewDefaultMetricsCollector() *DefaultMetricsCollector {
	return &DefaultMetricsCollector{
		metrics: NewMetrics(),
	}
}

// Collect 收集指标
func (c *DefaultMetricsCollector) Collect() *Metrics {
	c.metrics.UpdateSystemMetrics()
	return c.metrics.GetSnapshot()
}

// Reset 重置指标
func (c *DefaultMetricsCollector) Reset() {
	c.metrics.Reset()
}

// GetMetrics 获取指标
func (c *DefaultMetricsCollector) GetMetrics() *Metrics {
	return c.metrics
}
