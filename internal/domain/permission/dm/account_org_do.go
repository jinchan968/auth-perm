package dm

import (
	"time"
)

// AccountOrgDO 账户-组织关联领域对象
// 对应数据库表：account_org（替代 user_org）
type AccountOrgDO struct {
	AccountID      string    `gorm:"column:account_id;type:uuid;not null;index"`
	OrganizationID string    `gorm:"column:organization_id;type:uuid;not null;index"`
	TenantID       string    `gorm:"column:tenant_id;type:uuid;not null;index"`
	CreatedAt      time.Time `gorm:"column:created_at;not null"`
}

// TableName 指定表名
func (AccountOrgDO) TableName() string {
	return "account_org"
}

// NewAccountOrg 创建账户组织关联
func NewAccountOrg(accountID, organizationID, tenantID string) *AccountOrgDO {
	return &AccountOrgDO{
		AccountID:      accountID,
		OrganizationID: organizationID,
		TenantID:       tenantID,
		CreatedAt:      time.Now(),
	}
}
