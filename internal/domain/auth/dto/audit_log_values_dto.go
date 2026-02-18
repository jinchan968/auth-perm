package dto

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// AuditLogValuesDTO 审计日志值变化结构
type AuditLogValuesDTO struct {
	ChangedFields map[string]interface{} `json:"changed_fields"` // 变更字段
	Context       map[string]interface{} `json:"context"`        // 上下文信息
	Metadata      map[string]interface{} `json:"metadata"`       // 元数据
	Tags          []string               `json:"tags"`           // 标签
	IPAddress     string                 `json:"ip_address"`     // IP地址
	UserAgent     string                 `json:"user_agent"`     // 用户代理
	SessionID     string                 `json:"session_id"`     // 会话ID
	RequestID     string                 `json:"request_id"`     // 请求ID
	CorrelationID string                 `json:"correlation_id"` // 关联ID
}

// Scan 实现sql.Scanner接口
func (a *AuditLogValuesDTO) Scan(value interface{}) error {
	if value == nil {
		*a = AuditLogValuesDTO{
			ChangedFields: make(map[string]interface{}),
			Context:       make(map[string]interface{}),
			Metadata:      make(map[string]interface{}),
			Tags:          make([]string, 0),
		}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for AuditLogValuesDTO")
	}

	// 直接反序列化为AuditLogValuesDTO结构体
	var values AuditLogValuesDTO
	if err := json.Unmarshal(bytes, &values); err != nil {
		return err
	}

	// 处理nil的slice和map
	if values.ChangedFields == nil {
		values.ChangedFields = make(map[string]interface{})
	}
	if values.Context == nil {
		values.Context = make(map[string]interface{})
	}
	if values.Metadata == nil {
		values.Metadata = make(map[string]interface{})
	}
	if values.Tags == nil {
		values.Tags = make([]string, 0)
	}

	*a = values
	return nil
}

// Value 实现driver.Valuer接口
func (a AuditLogValuesDTO) Value() (driver.Value, error) {
	return json.Marshal(a)
}
