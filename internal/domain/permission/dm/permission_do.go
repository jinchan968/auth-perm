package dm

import (
	"strings"
	"time"

	"auth-perm/internal/domain/permission/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionDO 权限领域对象
type PermissionDO struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	TenantID    string `gorm:"column:tenant_id;type:uuid;not null;index;uniqueIndex:idx_tenant_code,priority:1"`
	Code        string `gorm:"column:code;not null;uniqueIndex:idx_tenant_code,priority:2"`
	Name        string `gorm:"column:name;not null"`
	Description string `gorm:"column:description"`
	Resource    string `gorm:"column:resource;index"` // 资源类型，如 users, posts, roles
	IsSystem    bool   `gorm:"column:is_system;default:false"`
	IsActive    bool   `gorm:"column:is_active;default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Roles []RoleDO `gorm:"many2many:role_permissions;"`
}

// NewPermission 创建新权限（数据库操作）
func NewPermission(tenantID, code, name string) *PermissionDO {
	now := time.Now()

	return &PermissionDO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Code:      strings.TrimSpace(code),
		Name:      strings.TrimSpace(name),
		Resource:  "",
		IsActive:  true,
		IsSystem:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewSystemPermission FUTURE: 系统权限创建 - 在实现权限管理时使用
func NewSystemPermission(code, name, description string) *PermissionDO {
	perm := NewPermission("system", code, name)
	perm.IsSystem = true
	perm.Description = description
	return perm
}

// BeforeCreate GORM钩子
func (p *PermissionDO) BeforeCreate(tx *gorm.DB) error {
	// 如果ID为空，自动生成UUID
	if p.ID == "" {
		p.ID = uuid.New().String()
	}

	return nil
}

// ToDTO 转换为DTO（避免循环导入，在dm层定义）
func (p *PermissionDO) ToDTO() *dto.PermissionDTO {
	if p == nil {
		return nil
	}
	return &dto.PermissionDTO{
		ID:          p.ID,
		TenantID:    p.TenantID,
		Code:        p.Code,
		Name:        p.Name,
		Description: p.Description,
		Resource:    p.Resource,
		IsSystem:    p.IsSystem,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// TableName 指定表名
func (p *PermissionDO) TableName() string {
	return "permissions"
}
