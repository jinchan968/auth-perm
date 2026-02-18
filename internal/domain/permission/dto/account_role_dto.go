package dto

import (
	"time"

	"github.com/google/uuid"
)

// AccountRoleDTO 账户角色关联数据传输对象
type AccountRoleDTO struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	RoleID    string    `json:"role_id"`
	TenantID  string    `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
}

// NewAccountRoleDTO 创建账户角色关联DTO
func NewAccountRoleDTO(accountID, roleID, tenantID string) *AccountRoleDTO {
	return &AccountRoleDTO{
		ID:        uuid.New().String(),
		AccountID: accountID,
		RoleID:    roleID,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
	}
}

// GetID 获取ID
func (a *AccountRoleDTO) GetID() string {
	return a.ID
}

// GetAccountID 获取账户ID
func (a *AccountRoleDTO) GetAccountID() string {
	return a.AccountID
}

// GetRoleID 获取角色ID
func (a *AccountRoleDTO) GetRoleID() string {
	return a.RoleID
}

// GetTenantID 获取租户ID
func (a *AccountRoleDTO) GetTenantID() string {
	return a.TenantID
}

// GetCreatedAt 获取创建时间
func (a *AccountRoleDTO) GetCreatedAt() time.Time {
	return a.CreatedAt
}
