package dto

import (
	"time"
)

// AuditLogEntryDTO 审计日志条目
type AuditLogEntryDTO struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id"`
	UserID       string            `json:"user_id"`
	AccountID    string            `json:"account_id"`
	Action       string            `json:"action"`
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id"`
	OldValues    AuditLogValuesDTO `json:"old_values,omitempty"`
	NewValues    AuditLogValuesDTO `json:"new_values,omitempty"`
	IPAddress    string            `json:"ip_address"`
	UserAgent    string            `json:"user_agent"`
	Success      bool              `json:"success"`
	ErrorMessage string            `json:"error_message,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

// AuditLogStatsDTO 审计日志统计
type AuditLogStatsDTO struct {
	ActionStats   map[string]int64 `json:"action_stats"`   // 操作统计
	ErrorStats    map[string]int64 `json:"error_stats"`    // 错误统计
	SuccessStats  map[string]int64 `json:"success_stats"`  // 成功/失败统计
	ResourceStats map[string]int64 `json:"resource_stats"` // 资源统计
}

// ==================== 业务方法 ====================

// SetOldValues 设置旧值
func (a *AuditLogEntryDTO) SetOldValues(values *AuditLogValuesDTO) {
	if values == nil {
		a.OldValues = AuditLogValuesDTO{}
	} else {
		a.OldValues = *values
	}
}

// SetNewValues 设置新值
func (a *AuditLogEntryDTO) SetNewValues(values *AuditLogValuesDTO) {
	if values == nil {
		a.NewValues = AuditLogValuesDTO{}
	} else {
		a.NewValues = *values
	}
}

// SetFailure 设置失败状态
func (a *AuditLogEntryDTO) SetFailure(errorMessage string) {
	a.Success = false
	a.ErrorMessage = errorMessage
}

// SetIPAddress 设置IP地址
func (a *AuditLogEntryDTO) SetIPAddress(ip string) {
	a.IPAddress = ip
}

// SetUserAgent 设置用户代理
func (a *AuditLogEntryDTO) SetUserAgent(userAgent string) {
	a.UserAgent = userAgent
}
