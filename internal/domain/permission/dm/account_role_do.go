package dm

import (
	"time"

	"github.com/google/uuid"
)

// AccountRoleDO 账户-角色关联领域对象
// 对应数据库表：account_roles（替代原来的 user_roles）
type AccountRoleDO struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	AccountID string    `gorm:"column:account_id;type:uuid;not null;index"`
	RoleID    string    `gorm:"column:role_id;type:uuid;not null;index"`
	TenantID  string    `gorm:"column:tenant_id;type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

// TableName 指定表名
func (AccountRoleDO) TableName() string {
	return "account_roles"
}

// NewAccountRole 创建账户角色关联
func NewAccountRole(accountID, roleID, tenantID string) *AccountRoleDO {
	return &AccountRoleDO{
		ID:        uuid.New().String(),
		AccountID: accountID,
		RoleID:    roleID,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
	}
}
