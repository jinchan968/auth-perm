package unit

import (
	"sync"
	"testing"
	"time"

	"auth-perm/internal/common/monitoring"
	"auth-perm/internal/domain/auth/repo"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// createMockDB 创建用于测试的内存数据库
func createMockDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}
	return db
}

func TestSecurityMonitor_RecordLoginAttempt(t *testing.T) {
	alertHandler := &mockAlertHandler{}
	auditRepo := repo.NewAuditLogRepo(createMockDB())

	monitor := monitoring.NewSecurityMonitor(auditRepo, alertHandler)

	// 测试记录成功登录
	monitor.RecordLoginAttempt(true, "account1", "192.168.1.1")
	metrics := monitor.GetMetrics()

	if metrics.LoginAttempts != 1 {
		t.Errorf("Expected login attempts to be 1, got %d", metrics.LoginAttempts)
	}
	if metrics.LoginSuccesses != 1 {
		t.Errorf("Expected login successes to be 1, got %d", metrics.LoginSuccesses)
	}

	// 测试记录失败登录
	monitor.RecordLoginAttempt(false, "account2", "192.168.1.2")
	metrics = monitor.GetMetrics()

	if metrics.LoginFailures != 1 {
		t.Errorf("Expected login failures to be 1, got %d", metrics.LoginFailures)
	}
}

func TestSecurityMonitor_RecordTOTPVerification(t *testing.T) {
	alertHandler := &mockAlertHandler{}
	auditRepo := repo.NewAuditLogRepo(createMockDB())

	monitor := monitoring.NewSecurityMonitor(auditRepo, alertHandler)

	// 测试记录成功TOTP验证
	monitor.RecordTOTPVerification(true, "account1", "192.168.1.1")
	metrics := monitor.GetMetrics()

	if metrics.TOTPVerifications != 1 {
		t.Errorf("Expected TOTP verifications to be 1, got %d", metrics.TOTPVerifications)
	}
	if metrics.TOTPSuccesses != 1 {
		t.Errorf("Expected TOTP successes to be 1, got %d", metrics.TOTPSuccesses)
	}

	// 测试记录失败TOTP验证
	monitor.RecordTOTPVerification(false, "account2", "192.168.1.2")
	metrics = monitor.GetMetrics()

	if metrics.TOTPVerifications != 2 {
		t.Errorf("Expected TOTP verifications to be 2, got %d", metrics.TOTPVerifications)
	}
}

func TestSecurityMonitor_GetMetricsSummary(t *testing.T) {
	alertHandler := &mockAlertHandler{}
	auditRepo := repo.NewAuditLogRepo(createMockDB())

	monitor := monitoring.NewSecurityMonitor(auditRepo, alertHandler)

	// 记录一些测试数据
	monitor.RecordLoginAttempt(true, "account1", "192.168.1.1")
	monitor.RecordLoginAttempt(false, "account2", "192.168.1.2")
	monitor.RecordLoginAttempt(true, "account3", "192.168.1.3")
	monitor.RecordOAuthLogin("github", "account1", "192.168.1.1")
	monitor.RecordTOTPVerification(true, "account1", "192.168.1.1")

	summary := monitor.GetMetricsSummary()

	if summary["login_attempts"].(int64) != 3 {
		t.Errorf("Expected login attempts to be 3, got %v", summary["login_attempts"])
	}
	if summary["login_successes"].(int64) != 2 {
		t.Errorf("Expected login successes to be 2, got %v", summary["login_successes"])
	}
	if summary["oauth_logins"].(int64) != 1 {
		t.Errorf("Expected OAuth logins to be 1, got %v", summary["oauth_logins"])
	}
	if summary["totp_verifications"].(int64) != 1 {
		t.Errorf("Expected TOTP verifications to be 1, got %v", summary["totp_verifications"])
	}

	// 检查成功率
	successRate := summary["login_success_rate"].(float64)
	if successRate < 66 || successRate > 67 {
		t.Errorf("Expected login success rate to be around 66.67, got %f", successRate)
	}
}

func TestSecurityMonitor_RecordAccountLockout(t *testing.T) {
	alertHandler := &mockAlertHandler{}
	auditRepo := repo.NewAuditLogRepo(createMockDB())

	monitor := monitoring.NewSecurityMonitor(auditRepo, alertHandler)

	// 记录账户锁定
	monitor.RecordAccountLockout("account1", "192.168.1.1", "too many failed attempts")
	metrics := monitor.GetMetrics()

	if metrics.AccountLockouts != 1 {
		t.Errorf("Expected account lockouts to be 1, got %d", metrics.AccountLockouts)
	}
}

func TestSecurityMonitor_RecordAnomalousLogin(t *testing.T) {
	alertHandler := &mockAlertHandler{}
	auditRepo := repo.NewAuditLogRepo(createMockDB())

	monitor := monitoring.NewSecurityMonitor(auditRepo, alertHandler)

	// 记录异常登录
	monitor.RecordAnomalousLogin("account1", "192.168.1.1", "Mozilla/5.0", "Beijing, China")

	// 等待告警处理（异步）
	time.Sleep(100 * time.Millisecond)

	// 这应该触发一个告警，mockAlertHandler会记录它
	if len(alertHandler.GetAlerts()) == 0 {
		t.Error("Expected an alert to be triggered")
	}
}

func TestSecurityMonitor_GetCacheHitRate(t *testing.T) {
	alertHandler := &mockAlertHandler{}
	auditRepo := repo.NewAuditLogRepo(createMockDB())

	monitor := monitoring.NewSecurityMonitor(auditRepo, alertHandler)

	// 记录缓存命中和未命中
	monitor.RecordCacheHit()
	monitor.RecordCacheHit()
	monitor.RecordCacheMiss()
	monitor.RecordCacheMiss()
	monitor.RecordCacheMiss()

	hitRate := monitor.GetCacheHitRate()
	expectedRate := 40.0 // 2/5 = 40%

	if hitRate < expectedRate-1 || hitRate > expectedRate+1 {
		t.Errorf("Expected cache hit rate to be around %f, got %f", expectedRate, hitRate)
	}
}

// mockAlertHandler 模拟告警处理器
type mockAlertHandler struct {
	alerts []*monitoring.SecurityAlert
	mu     sync.Mutex
}

func (h *mockAlertHandler) HandleAlert(alert *monitoring.SecurityAlert) error {
	h.mu.Lock()
	h.alerts = append(h.alerts, alert)
	h.mu.Unlock()
	return nil
}

func (h *mockAlertHandler) GetAlerts() []*monitoring.SecurityAlert {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.alerts
}
