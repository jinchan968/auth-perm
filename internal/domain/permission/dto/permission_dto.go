package dto

import (
	"auth-perm/internal/common/errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PermissionDTO 权限数据传输对象（包含业务逻辑）
type PermissionDTO struct {
	// 基本信息
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// 权限信息
	Resource string `json:"resource"` // 资源类型，如 users, posts, roles

	// 状态信息
	IsSystem   bool `json:"is_system"`
	IsActive   bool `json:"is_active"`
	IsSelected bool `json:"is_selected"` // 角色关联状态，前端用于展示勾选

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewPermissionDTO 创建权限DTO
func NewPermissionDTO(tenantID, code, name string) *PermissionDTO {
	now := time.Now()

	return &PermissionDTO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Code:      strings.ToLower(strings.TrimSpace(code)),
		Name:      strings.TrimSpace(name),
		Resource:  "",
		IsActive:  true,
		IsSystem:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewSystemPermissionDTO 创建系统权限DTO
func NewSystemPermissionDTO(code, name, description string) *PermissionDTO {
	perm := NewPermissionDTO("system", code, name)
	perm.IsSystem = true
	perm.Description = description
	return perm
}

// ==================== Getter 方法 ====================

// GetID 获取权限ID
func (p *PermissionDTO) GetID() string {
	return p.ID
}

// GetTenantID 获取租户ID
func (p *PermissionDTO) GetTenantID() string {
	return p.TenantID
}

// GetCode 获取权限代码
func (p *PermissionDTO) GetCode() string {
	return p.Code
}

// GetName 获取权限名称
func (p *PermissionDTO) GetName() string {
	return p.Name
}

// GetDescription 获取权限描述
func (p *PermissionDTO) GetDescription() string {
	return p.Description
}

// GetResource 获取资源类型
func (p *PermissionDTO) GetResource() string {
	return p.Resource
}

// GetIsSystem 获取是否为系统权限
func (p *PermissionDTO) GetIsSystem() bool {
	return p.IsSystem
}

// GetIsActive 获取是否活跃
func (p *PermissionDTO) GetIsActive() bool {
	return p.IsActive
}

// GetCreatedAt 获取创建时间
func (p *PermissionDTO) GetCreatedAt() time.Time {
	return p.CreatedAt
}

// GetUpdatedAt 获取更新时间
func (p *PermissionDTO) GetUpdatedAt() time.Time {
	return p.UpdatedAt
}

// ==================== 业务方法 ====================

// UpdateInfo 更新权限信息
func (p *PermissionDTO) UpdateInfo(name, description string) error {
	if name != "" {
		if len(name) > 100 {
			return errors.NewValidationError("权限名称过长，最大100个字符")
		}
		p.Name = strings.TrimSpace(name)
	}

	if description != "" {
		if len(description) > 500 {
			return errors.NewValidationError("描述过长，最大500个字符")
		}
		p.Description = strings.TrimSpace(description)
	}

	p.UpdatedAt = time.Now()
	return nil
}

// Activate 激活权限
func (p *PermissionDTO) Activate() error {
	if p.IsActive {
		return errors.NewBusinessError("权限已经是激活状态")
	}

	p.IsActive = true
	p.UpdatedAt = time.Now()
	return nil
}

// Deactivate 停用权限
func (p *PermissionDTO) Deactivate() error {
	if !p.IsActive {
		return errors.NewBusinessError("权限已经是停用状态")
	}

	if p.IsSystem {
		return errors.NewBusinessError("系统权限不能被停用")
	}

	p.IsActive = false
	p.UpdatedAt = time.Now()
	return nil
}

// Matches 检查权限是否匹配指定的资源
func (p *PermissionDTO) Matches(resource string) bool {
	return p.Resource == resource
}

// IsWildcard 检查是否为通配符权限
func (p *PermissionDTO) IsWildcard() bool {
	return strings.Contains(p.Code, "*")
}

// CanModify 检查权限是否可以被修改
func (p *PermissionDTO) CanModify() bool {
	return !p.IsSystem
}

// CanDelete 检查权限是否可以被删除
func (p *PermissionDTO) CanDelete() bool {
	return !p.IsSystem
}

// ToDTO 转换为权限DTO
func (p *PermissionDTO) ToDTO() *PermissionDTO {
	return p
}
