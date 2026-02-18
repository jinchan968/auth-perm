package monitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// AlertLevel 报警级别
type AlertLevel string

const (
	AlertLevelInfo    AlertLevel = "info"
	AlertLevelWarning AlertLevel = "warning"
	AlertLevelError   AlertLevel = "error"
	AlertLevelFatal   AlertLevel = "fatal"
)

// AlertStatus 报警状态
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
)

// Alert 报警信息
type Alert struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Level       AlertLevel             `json:"level"`
	Status      AlertStatus            `json:"status"`
	Message     string                 `json:"message"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]interface{} `json:"annotations"`
	StartsAt    time.Time              `json:"starts_at"`
	EndsAt      *time.Time             `json:"ends_at,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	Resolved    bool                   `json:"resolved"`
	WebhookURL  string                 `json:"-"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	mu          sync.RWMutex
}

// NewAlert FUTURE: 告警创建 - 在实现告警时使用
func NewAlert(name, message string, level AlertLevel) *Alert {
	return &Alert{
		ID:          generateAlertID(),
		Name:        name,
		Level:       level,
		Status:      AlertStatusFiring,
		Message:     message,
		Labels:      make(map[string]string),
		Annotations: make(map[string]interface{}),
		StartsAt:    time.Now(),
		Timeout:     5 * time.Minute, // 默认5分钟超时
		Resolved:    false,
		MaxRetries:  3,
	}
}

// AddLabel 添加标签
func (a *Alert) AddLabel(key, value string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Labels[key] = value
}

// AddAnnotation 添加注解
func (a *Alert) AddAnnotation(key string, value interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Annotations[key] = value
}

// Resolve 解决报警
func (a *Alert) Resolve() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	a.Status = AlertStatusResolved
	a.Resolved = true
	a.EndsAt = &now
}

// IsExpired 检查报警是否过期
func (a *Alert) IsExpired() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return time.Since(a.StartsAt) > a.Timeout
}

// ShouldRetry 检查是否应该重试
func (a *Alert) ShouldRetry() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return !a.Resolved && a.RetryCount < a.MaxRetries
}

// IncrementRetry 增加重试次数
func (a *Alert) IncrementRetry() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.RetryCount++
}

// ToJSON 转换为JSON
func (a *Alert) ToJSON() ([]byte, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return json.Marshal(a)
}

// AlertRule 报警规则
type AlertRule struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Condition   func(*Metrics) *Alert  `json:"-"`
	Enabled     bool                   `json:"enabled"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]interface{} `json:"annotations"`
}

// NewAlertRule FUTURE: 告警规则创建 - 在实现告警规则时使用
func NewAlertRule(name, description string, condition func(*Metrics) *Alert) *AlertRule {
	return &AlertRule{
		Name:        name,
		Description: description,
		Condition:   condition,
		Enabled:     true,
		Labels:      make(map[string]string),
		Annotations: make(map[string]interface{}),
	}
}

// SetLabel 设置规则标签
func (r *AlertRule) SetLabel(key, value string) {
	r.Labels[key] = value
}

// SetAnnotation 设置规则注解
func (r *AlertRule) SetAnnotation(key string, value interface{}) {
	r.Annotations[key] = value
}

// Evaluate 评估规则
func (r *AlertRule) Evaluate(metrics *Metrics) *Alert {
	if !r.Enabled {
		return nil
	}

	alert := r.Condition(metrics)
	if alert != nil {
		// 应用规则的标签和注解
		for k, v := range r.Labels {
			alert.AddLabel(k, v)
		}
		for k, v := range r.Annotations {
			alert.AddAnnotation(k, v)
		}
	}

	return alert
}

// AlertManager 报警管理器
type AlertManager struct {
	rules    map[string]*AlertRule
	active   map[string]*Alert
	history  []*Alert
	webhooks []string
	mu       sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewAlertManager FUTURE: 告警管理器创建 - 在实现告警管理时使用
func NewAlertManager() *AlertManager {
	am := &AlertManager{
		rules:    make(map[string]*AlertRule),
		active:   make(map[string]*Alert),
		history:  make([]*Alert, 0),
		stopChan: make(chan struct{}),
	}

	// 启动后台处理
	am.start()

	return am
}

// AddRule 添加报警规则
func (am *AlertManager) AddRule(rule *AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.rules[rule.Name] = rule
}

// RemoveRule 移除报警规则
func (am *AlertManager) RemoveRule(name string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	delete(am.rules, name)
}

// EnableRule 启用规则
func (am *AlertManager) EnableRule(name string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	rule, exists := am.rules[name]
	if !exists {
		return fmt.Errorf("rule '%s' not found", name)
	}

	rule.Enabled = true
	return nil
}

// DisableRule 禁用规则
func (am *AlertManager) DisableRule(name string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	rule, exists := am.rules[name]
	if !exists {
		return fmt.Errorf("rule '%s' not found", name)
	}

	rule.Enabled = false
	return nil
}

// AddWebhook 添加Webhook地址
func (am *AlertManager) AddWebhook(url string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.webhooks = append(am.webhooks, url)
}

// RemoveWebhook 移除Webhook地址
func (am *AlertManager) RemoveWebhook(url string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	for i, webhook := range am.webhooks {
		if webhook == url {
			am.webhooks = append(am.webhooks[:i], am.webhooks[i+1:]...)
			break
		}
	}
}

// EvaluateRules 评估所有规则
func (am *AlertManager) EvaluateRules(metrics *Metrics) {
	am.mu.RLock()
	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		rules = append(rules, rule)
	}
	am.mu.RUnlock()

	for _, rule := range rules {
		alert := rule.Evaluate(metrics)
		if alert != nil {
			am.TriggerAlert(alert)
		}
	}
}

// TriggerAlert 触发报警
func (am *AlertManager) TriggerAlert(alert *Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// 检查是否已有相同的活跃报警
	if existing, exists := am.active[alert.ID]; exists {
		// 更新现有报警
		existing.Message = alert.Message
		existing.StartsAt = alert.StartsAt
		return
	}

	// 添加到活跃报警
	am.active[alert.ID] = alert
	am.history = append(am.history, alert)

	// 异步发送通知
	go am.sendNotification(alert)
}

// ResolveAlert 解决报警
func (am *AlertManager) ResolveAlert(alertID string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if alert, exists := am.active[alertID]; exists {
		alert.Resolve()
		delete(am.active, alertID)
		go am.sendNotification(alert)
	}
}

// GetActiveAlerts 获取活跃报警
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0, len(am.active))
	for _, alert := range am.active {
		alerts = append(alerts, alert)
	}

	return alerts
}

// GetAlertHistory 获取报警历史
func (am *AlertManager) GetAlertHistory(limit int) []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	historyLen := len(am.history)
	if limit > 0 && limit < historyLen {
		start := historyLen - limit
		return am.history[start:]
	}

	return am.history
}

// sendNotification 发送通知
func (am *AlertManager) sendNotification(alert *Alert) {
	if len(am.webhooks) == 0 {
		return
	}

	alertData, err := alert.ToJSON()
	if err != nil {
		log.Printf("Failed to marshal alert %s: %v", alert.ID, err)
		return
	}

	for _, webhook := range am.webhooks {
		go am.sendWebhook(webhook, alertData)
	}
}

// sendWebhook 发送Webhook
func (am *AlertManager) sendWebhook(webhook string, data []byte) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Post(webhook, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("Failed to send webhook to %s: %v", webhook, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("Webhook returned error status %d for %s", resp.StatusCode, webhook)
	}
}

// start 启动后台处理
func (am *AlertManager) start() {
	am.wg.Add(1)
	go am.cleanupExpiredAlerts()
}

// cleanupExpiredAlerts 清理过期报警
func (am *AlertManager) cleanupExpiredAlerts() {
	defer am.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			am.cleanupExpired()
		case <-am.stopChan:
			return
		}
	}
}

// cleanupExpired 清理过期报警
func (am *AlertManager) cleanupExpired() {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, alert := range am.active {
		if alert.IsExpired() {
			alert.Resolve()
			delete(am.active, alert.ID)
			go am.sendNotification(alert)
		}
	}
}

// Stop 停止报警管理器
func (am *AlertManager) Stop() {
	close(am.stopChan)
	am.wg.Wait()
}

// GetStats 获取报警统计
func (am *AlertManager) GetStats() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()

	levelCounts := make(map[string]int)
	for _, alert := range am.active {
		levelCounts[string(alert.Level)]++
	}

	return map[string]interface{}{
		"active_alerts": len(am.active),
		"total_alerts":  len(am.history),
		"webhooks":      len(am.webhooks),
		"level_counts":  levelCounts,
	}
}

// generateAlertID FUTURE: 告警ID生成 - 在实现告警ID时使用
func generateAlertID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}

// DefaultRules FUTURE: 默认告警规则 - 在实现默认告警时使用
func DefaultRules() []*AlertRule {
	var rules []*AlertRule

	// HTTP错误率报警
	rules = append(rules, NewAlertRule(
		"http_error_rate_high",
		"HTTP错误率过高",
		func(metrics *Metrics) *Alert {
			errorRate := metrics.GetHTTPErrorRate()
			if errorRate > 0.1 { // 10%
				alert := NewAlert("HTTP Error Rate High",
					fmt.Sprintf("HTTP error rate is %.2f%%", errorRate*100),
					AlertLevelError)
				alert.AddLabel("service", "http")
				alert.AddLabel("error_rate", fmt.Sprintf("%.2f", errorRate))
				return alert
			}
			return nil
		},
	))

	// 缓存命中率低报警
	rules = append(rules, NewAlertRule(
		"cache_hit_rate_low",
		"缓存命中率过低",
		func(metrics *Metrics) *Alert {
			hitRate := metrics.GetCacheHitRate()
			if hitRate < 0.8 && (metrics.CacheHits+metrics.CacheMisses) > 100 { // 有足够的数据
				alert := NewAlert("Cache Hit Rate Low",
					fmt.Sprintf("Cache hit rate is %.2f%%", hitRate*100),
					AlertLevelWarning)
				alert.AddLabel("service", "cache")
				alert.AddLabel("hit_rate", fmt.Sprintf("%.2f", hitRate))
				return alert
			}
			return nil
		},
	))

	// 数据库错误报警
	rules = append(rules, NewAlertRule(
		"database_errors",
		"数据库错误",
		func(metrics *Metrics) *Alert {
			if metrics.DatabaseErrors > 0 {
				alert := NewAlert("Database Errors",
					fmt.Sprintf("Database errors: %d", metrics.DatabaseErrors),
					AlertLevelError)
				alert.AddLabel("service", "database")
				alert.AddLabel("error_count", fmt.Sprintf("%d", metrics.DatabaseErrors))
				return alert
			}
			return nil
		},
	))

	return rules
}
