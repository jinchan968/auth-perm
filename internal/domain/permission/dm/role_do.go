package dm

import (
	"strings"
	"time"

	"auth-perm/internal/domain/permission/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleDO 角色领域对象
type RoleDO struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	TenantID    string `gorm:"column:tenant_id;type:uuid;not null;uniqueIndex:idx_roles_tenant_code,priority:1"`
	OrgID       string `gorm:"column:org_id;type:uuid;index"` // 组织ID，为空表示全局角色
	Name        string `gorm:"column:name;not null"`
	Code        string `gorm:"column:code;uniqueIndex:idx_roles_tenant_code,priority:2;not null"`
	Description string `gorm:"column:description"`
	IsSystem    bool   `gorm:"column:is_system;default:false"` // 是否为系统角色
	IsActive    bool   `gorm:"column:is_active;default:true"`
	Priority    int    `gorm:"column:priority;default:0"` // 优先级，数字越大优先级越高
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Permissions []PermissionDO `gorm:"many2many:role_permissions;"`
}

// NewRole FUTURE: 角色创建 - 在实现角色管理时使用
func NewRole(tenantID, code, name string) *RoleDO {
	now := time.Now()
	return &RoleDO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Code:      strings.TrimSpace(code),
		Name:      strings.TrimSpace(name),
		IsActive:  true,
		IsSystem:  false,
		Priority:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewSystemRole FUTURE: 系统角色创建 - 在实现角色管理时使用
func NewSystemRole(code, name, description string) *RoleDO {
	role := NewRole("system", code, name)
	role.IsSystem = true
	role.Description = description
	return role
}

// BeforeCreate GORM钩子
func (r *RoleDO) BeforeCreate(tx *gorm.DB) error {
	// 如果ID为空，自动生成UUID
	if r.ID == "" {
		r.ID = uuid.New().String()
	}

	return nil
}

// ToDTO 转换为DTO（避免循环导入，在dm层定义）
func (r *RoleDO) ToDTO() *dto.RoleDTO {
	if r == nil {
		return nil
	}
	dto := &dto.RoleDTO{
		ID:          r.ID,
		TenantID:    r.TenantID,
		OrgID:       r.OrgID,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		IsActive:    r.IsActive,
		Priority:    r.Priority,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	// 如果有预加载的Permissions，则设置数量
	if len(r.Permissions) > 0 {
		dto.PermissionCount = len(r.Permissions)
	}
	return dto
}

// TableName 指定表名
func (r *RoleDO) TableName() string {
	return "roles"
}
