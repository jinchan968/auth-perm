package dm

import (
	"auth-perm/internal/domain/auth/dto"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLogDO 审计日志领域对象
type AuditLogDO struct {
	ID           string                `gorm:"primaryKey;type:uuid"`
	TenantID     string                `gorm:"column:tenant_id;type:uuid;not null;index"`
	UserID       string                `gorm:"column:user_id;type:uuid;not null;index"`
	AccountID    *string               `gorm:"column:account_id;type:uuid;index"` // 可选字段，记录操作对应的账户
	Action       string                `gorm:"column:action;type:varchar(100);not null;index"`
	ResourceType string                `gorm:"column:resource_type;type:varchar(50);not null;index"`
	ResourceID   string                `gorm:"column:resource_id;type:varchar(255);index"`
	OldValues    dto.AuditLogValuesDTO `gorm:"column:old_values;type:jsonb;default:'{}'"`
	NewValues    dto.AuditLogValuesDTO `gorm:"column:new_values;type:jsonb;default:'{}'"`
	IPAddress    *string               `gorm:"column:ip_address;type:inet"`
	UserAgent    string                `gorm:"column:user_agent;type:text"`
	Success      bool                  `gorm:"column:success;default:true;index"`
	ErrorMessage string                `gorm:"column:error_message;type:text"`
	CreatedAt    time.Time             `gorm:"column:created_at;not null;index"`
	UpdatedAt    time.Time             `gorm:"column:updated_at;not null"`
	DeletedAt    gorm.DeletedAt        `gorm:"index"`

	// 关联关系
	User UserDO `gorm:"foreignKey:UserID"`
}

// NewAuditLog 创建审计日志
func NewAuditLog(tenantID, userID, action, resourceType, resourceID string) *AuditLogDO {
	return NewAuditLogWithAccount(tenantID, userID, "", action, resourceType, resourceID)
}

// NewAuditLogWithAccount 创建审计日志（带账户信息）
func NewAuditLogWithAccount(tenantID, userID, accountID, action, resourceType, resourceID string) *AuditLogDO {
	now := time.Now()
	return &AuditLogDO{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		UserID:       userID,
		AccountID:    &accountID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OldValues:    dto.AuditLogValuesDTO{},
		NewValues:    dto.AuditLogValuesDTO{},
		Success:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// SetOldValues 设置旧值
func (a *AuditLogDO) SetOldValues(values *dto.AuditLogValuesDTO) {
	a.OldValues = *values
}

// GetOldValues 获取旧值
func (a *AuditLogDO) GetOldValues() *dto.AuditLogValuesDTO {
	return &a.OldValues
}

// SetNewValues 设置新值
func (a *AuditLogDO) SetNewValues(values *dto.AuditLogValuesDTO) {
	a.NewValues = *values
}

// GetNewValues 获取新值
func (a *AuditLogDO) GetNewValues() *dto.AuditLogValuesDTO {
	return &a.NewValues
}

// BeforeCreate GORM钩子
func (a *AuditLogDO) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	if a.UpdatedAt.IsZero() {
		a.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate GORM钩子
func (a *AuditLogDO) BeforeUpdate(tx *gorm.DB) error {
	a.UpdatedAt = time.Now()
	return nil
}

// AuditLogFromDTO 从DTO创建AuditLogDO
func AuditLogFromDTO(d *dto.AuditLogEntryDTO) *AuditLogDO {
	if d == nil {
		return nil
	}

	var ipAddress *string
	if d.IPAddress != "" {
		ipAddress = &d.IPAddress
	}

	var accountID *string
	if d.AccountID != "" {
		accountID = &d.AccountID
	}

	return &AuditLogDO{
		ID:           d.ID,
		TenantID:     d.TenantID,
		UserID:       d.UserID,
		AccountID:    accountID,
		Action:       d.Action,
		ResourceType: d.ResourceType,
		ResourceID:   d.ResourceID,
		OldValues:    d.OldValues,
		NewValues:    d.NewValues,
		IPAddress:    ipAddress,
		UserAgent:    d.UserAgent,
		Success:      d.Success,
		ErrorMessage: d.ErrorMessage,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    time.Now(),
	}
}

// ToDTO 转换为DTO
func (a *AuditLogDO) ToDTO() *dto.AuditLogEntryDTO {
	if a == nil {
		return nil
	}

	var ipAddress string
	if a.IPAddress != nil {
		ipAddress = *a.IPAddress
	}

	var accountID string
	if a.AccountID != nil {
		accountID = *a.AccountID
	}

	return &dto.AuditLogEntryDTO{
		ID:           a.ID,
		TenantID:     a.TenantID,
		UserID:       a.UserID,
		AccountID:    accountID,
		Action:       a.Action,
		ResourceType: a.ResourceType,
		ResourceID:   a.ResourceID,
		OldValues:    *a.GetOldValues(),
		NewValues:    *a.GetNewValues(),
		IPAddress:    ipAddress,
		UserAgent:    a.UserAgent,
		Success:      a.Success,
		ErrorMessage: a.ErrorMessage,
		CreatedAt:    a.CreatedAt,
	}
}

// TableName 指定表名
func (a *AuditLogDO) TableName() string {
	return "audit_logs"
}
