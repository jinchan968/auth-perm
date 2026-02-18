package dm

import (
	"time"

	"github.com/google/uuid"
)

// RolePermissionDO 角色-权限关联领域对象
// 对应数据库表：role_permissions
type RolePermissionDO struct {
	ID           string    `gorm:"primaryKey;type:uuid"`
	RoleID       string    `gorm:"column:role_id;type:uuid;not null;uniqueIndex:idx_role_permission"`
	PermissionID string    `gorm:"column:permission_id;type:uuid;not null;uniqueIndex:idx_role_permission"`
	TenantID     string    `gorm:"column:tenant_id;type:uuid;not null;index;default:'default'"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

// TableName 指定表名
func (RolePermissionDO) TableName() string {
	return "role_permissions"
}

// NewRolePermission 创建角色权限关联
func NewRolePermission(roleID, permissionID, tenantID string) *RolePermissionDO {
	return &RolePermissionDO{
		ID:           uuid.New().String(),
		RoleID:       roleID,
		PermissionID: permissionID,
		TenantID:     tenantID,
		CreatedAt:    time.Now(),
	}
}
