package monitoring

import (
	"fmt"
	"sync"
	"time"

	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/repo"
)

// SecurityMetrics 安全指标
type SecurityMetrics struct {
	LoginAttempts     int64 `json:"login_attempts"`
	LoginSuccesses    int64 `json:"login_successes"`
	LoginFailures     int64 `json:"login_failures"`
	OAuthLogins       int64 `json:"oauth_logins"`
	TOTPVerifications int64 `json:"totp_verifications"`
	TOTPSuccesses     int64 `json:"totp_successes"`
	AccountLockouts   int64 `json:"account_lockouts"`
	FailedAttempts    int64 `json:"failed_attempts"`
	ActiveSessions    int64 `json:"active_sessions"`
	CacheHits         int64 `json:"cache_hits"`
	CacheMisses       int64 `json:"cache_misses"`
}

// SecurityMonitor 安全监控
type SecurityMonitor struct {
	metrics      SecurityMetrics
	mu           sync.RWMutex
	auditRepo    *repo.AuditLogRepo
	alertHandler AlertHandler
	startTime    time.Time
}

// AlertHandler 告警处理接口
type AlertHandler interface {
	HandleAlert(alert *SecurityAlert) error
}

// SecurityAlert 安全告警
type SecurityAlert struct {
	Type      AlertType              `json:"type"`
	Severity  Severity               `json:"severity"`
	Message   string                 `json:"message"`
	AccountID string                 `json:"account_id,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AlertType 告警类型
type AlertType string

const (
	AlertBruteForce       AlertType = "brute_force"
	AlertAccountLockout   AlertType = "account_lockout"
	AlertAnomalousLogin   AlertType = "anomalous_login"
	AlertHighRiskLogin    AlertType = "high_risk_login"
	AlertMultipleFailures AlertType = "multiple_failures"
)

// Severity 严重级别
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// NewSecurityMonitor FUTURE: 安全监控创建 - 在实现安全监控时使用
func NewSecurityMonitor(auditRepo *repo.AuditLogRepo, alertHandler AlertHandler) *SecurityMonitor {
	return &SecurityMonitor{
		metrics:      SecurityMetrics{},
		auditRepo:    auditRepo,
		alertHandler: alertHandler,
		startTime:    time.Now(),
	}
}

// RecordLoginAttempt 记录登录尝试
func (m *SecurityMonitor) RecordLoginAttempt(success bool, accountID, ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.LoginAttempts++
	if success {
		m.metrics.LoginSuccesses++
	} else {
		m.metrics.LoginFailures++
	}

	// 检查是否需要告警
	if !success {
		// 连续失败检查
		failCount := m.getRecentFailures(accountID)
		if failCount >= 5 {
			m.triggerAlert(&SecurityAlert{
				Type:      AlertMultipleFailures,
				Severity:  SeverityWarning,
				Message:   fmt.Sprintf("账户 %s 在15分钟内连续失败 %d 次", accountID, failCount),
				AccountID: accountID,
				IPAddress: ip,
				Timestamp: time.Now(),
				Metadata:  map[string]interface{}{"failures": failCount},
			})
		}
	}
}

// RecordOAuthLogin 记录OAuth登录
func (m *SecurityMonitor) RecordOAuthLogin(provider string, accountID, ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.OAuthLogins++

	m.triggerAlert(&SecurityAlert{
		Type:      AlertAnomalousLogin,
		Severity:  SeverityInfo,
		Message:   fmt.Sprintf("OAuth登录: %s", provider),
		AccountID: accountID,
		IPAddress: ip,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"provider": provider},
	})
}

// RecordTOTPVerification 记录TOTP验证
func (m *SecurityMonitor) RecordTOTPVerification(success bool, accountID, ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.TOTPVerifications++
	if success {
		m.metrics.TOTPSuccesses++
	}

	if !success {
		failCount := m.getRecentTOTPFailures(accountID)
		if failCount >= 3 {
			m.triggerAlert(&SecurityAlert{
				Type:      AlertBruteForce,
				Severity:  SeverityCritical,
				Message:   fmt.Sprintf("账户 %s TOTP验证连续失败 %d 次", accountID, failCount),
				AccountID: accountID,
				IPAddress: ip,
				Timestamp: time.Now(),
				Metadata:  map[string]interface{}{"failures": failCount},
			})
		}
	}
}

// RecordAccountLockout 记录账户锁定
func (m *SecurityMonitor) RecordAccountLockout(accountID, ip, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.AccountLockouts++

	m.triggerAlert(&SecurityAlert{
		Type:      AlertAccountLockout,
		Severity:  SeverityCritical,
		Message:   fmt.Sprintf("账户 %s 被锁定: %s", accountID, reason),
		AccountID: accountID,
		IPAddress: ip,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"reason": reason},
	})
}

// RecordAnomalousLogin 记录异常登录
func (m *SecurityMonitor) RecordAnomalousLogin(accountID, ip, userAgent, location string) {
	m.triggerAlert(&SecurityAlert{
		Type:      AlertAnomalousLogin,
		Severity:  SeverityWarning,
		Message:   fmt.Sprintf("检测到异常登录: %s 从 %s", accountID, location),
		AccountID: accountID,
		IPAddress: ip,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"user_agent": userAgent, "location": location},
	})
}

// RecordHighRiskLogin 记录高风险登录
func (m *SecurityMonitor) RecordHighRiskLogin(accountID, ip, riskFactor string) {
	m.triggerAlert(&SecurityAlert{
		Type:      AlertHighRiskLogin,
		Severity:  SeverityWarning,
		Message:   fmt.Sprintf("高风险登录: %s 风险因子: %s", accountID, riskFactor),
		AccountID: accountID,
		IPAddress: ip,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"risk_factor": riskFactor},
	})
}

// GetMetrics 获取指标
func (m *SecurityMonitor) GetMetrics() SecurityMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.metrics
}

// GetMetricsSummary 获取指标摘要
func (m *SecurityMonitor) GetMetricsSummary() map[string]interface{} {
	metrics := m.GetMetrics()
	uptime := time.Since(m.startTime).Hours()

	successRate := float64(0)
	if metrics.LoginAttempts > 0 {
		successRate = float64(metrics.LoginSuccesses) / float64(metrics.LoginAttempts) * 100
	}

	totpSuccessRate := float64(0)
	if metrics.TOTPVerifications > 0 {
		totpSuccessRate = float64(metrics.TOTPSuccesses) / float64(metrics.TOTPVerifications) * 100
	}

	return map[string]interface{}{
		"uptime_hours":       uptime,
		"login_attempts":     metrics.LoginAttempts,
		"login_successes":    metrics.LoginSuccesses,
		"login_failures":     metrics.LoginFailures,
		"login_success_rate": successRate,
		"oauth_logins":       metrics.OAuthLogins,
		"totp_verifications": metrics.TOTPVerifications,
		"totp_successes":     metrics.TOTPSuccesses,
		"totp_success_rate":  totpSuccessRate,
		"account_lockouts":   metrics.AccountLockouts,
		"active_sessions":    metrics.ActiveSessions,
		"cache_hits":         metrics.CacheHits,
		"cache_misses":       metrics.CacheMisses,
		"cache_hit_rate":     m.getCacheHitRate(),
	}
}

// GetCacheHitRate 获取缓存命中率
func (m *SecurityMonitor) GetCacheHitRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.metrics.CacheHits + m.metrics.CacheMisses
	if total == 0 {
		return 0
	}

	return float64(m.metrics.CacheHits) / float64(total) * 100
}

// RecordCacheHit 记录缓存命中
func (m *SecurityMonitor) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics.CacheHits++
}

// RecordCacheMiss 记录缓存未命中
func (m *SecurityMonitor) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics.CacheMisses++
}

// RecordActiveSession 记录活跃会话
func (m *SecurityMonitor) RecordActiveSession(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics.ActiveSessions = count
}

// 辅助方法

func (m *SecurityMonitor) getRecentFailures(accountID string) int64 {
	// 这里应该查询数据库获取最近15分钟的失败次数
	// 简化实现，返回0
	return 0
}

func (m *SecurityMonitor) getRecentTOTPFailures(accountID string) int64 {
	// 这里应该查询缓存或数据库获取最近TOTP失败次数
	// 简化实现，返回0
	return 0
}

func (m *SecurityMonitor) triggerAlert(alert *SecurityAlert) {
	if m.alertHandler != nil {
		go func() {
			_ = m.alertHandler.HandleAlert(alert)
		}()
	}

	// 同时记录到审计日志（异步）
	entry := &dto.AuditLogEntryDTO{
		Action:       string(alert.Type),
		ResourceType: "security",
		ResourceID:   alert.AccountID,
		Success:      false,
		IPAddress:    alert.IPAddress,
		UserAgent:    alert.Message,
	}
	m.auditRepo.LogAsync(entry)
}

func (m *SecurityMonitor) getCacheHitRate() float64 {
	total := m.metrics.CacheHits + m.metrics.CacheMisses
	if total == 0 {
		return 0
	}
	return float64(m.metrics.CacheHits) / float64(total) * 100
}
