package dto

import (
	"auth-perm/internal/common/errors"
	permissionConstant "auth-perm/internal/domain/permission/constant"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RoleDTO 角色数据传输对象（包含业务逻辑）
type RoleDTO struct {
	// 基本信息
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	OrgID       string `json:"org_id"` // 组织ID，为空表示全局角色
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`

	// 状态信息
	IsSystem bool `json:"is_system"` // 是否为系统角色
	IsActive bool `json:"is_active"`
	Priority int  `json:"priority"` // 优先级，数字越大优先级越高

	// 关联信息
	PermissionCount int              `json:"permission_count"` // 关联的权限数量
	Permissions     []*PermissionDTO `json:"permissions"`      // 关联的权限列表

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewRoleDTO FUTURE: 角色DTO创建 - 在实现角色管理时使用
func NewRoleDTO(tenantID, code, name string) *RoleDTO {
	now := time.Now()
	return &RoleDTO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Code:      strings.ToLower(strings.TrimSpace(code)),
		Name:      strings.TrimSpace(name),
		IsActive:  true,
		IsSystem:  false,
		Priority:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewSystemRoleDTO FUTURE: 系统角色DTO创建 - 在实现角色管理时使用
func NewSystemRoleDTO(code, name, description string) *RoleDTO {
	role := NewRoleDTO("system", code, name)
	role.IsSystem = true
	role.Description = description
	return role
}

/*
// FromRoleDO 从DO创建DTO
func FromRoleDO(roleDO *dm.RoleDO) *RoleDTO {
	if roleDO == nil {
		return nil
	}
	return &RoleDTO{
		ID:          roleDO.ID,
		TenantID:    roleDO.TenantID,
		OrgID:       roleDO.OrgID,
		Name:        roleDO.Name,
		Code:        roleDO.Code,
		Description: roleDO.Description,
		IsSystem:    roleDO.IsSystem,
		IsActive:    roleDO.IsActive,
		Priority:    roleDO.Priority,
		CreatedAt:   roleDO.CreatedAt,
		UpdatedAt:   roleDO.UpdatedAt,
	}
}

// ToRoleDO 转换为DO
func (r *RoleDTO) ToRoleDO() *dm.RoleDO {
	return &dm.RoleDO{
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
}
*/

// ==================== Getter 方法 ====================

// GetID 获取角色ID
func (r *RoleDTO) GetID() string {
	return r.ID
}

// GetTenantID 获取租户ID
func (r *RoleDTO) GetTenantID() string {
	return r.TenantID
}

// GetOrgID 获取组织ID
func (r *RoleDTO) GetOrgID() string {
	return r.OrgID
}

// GetName 获取角色名称
func (r *RoleDTO) GetName() string {
	return r.Name
}

// GetCode 获取角色代码
func (r *RoleDTO) GetCode() string {
	return r.Code
}

// GetDescription 获取角色描述
func (r *RoleDTO) GetDescription() string {
	return r.Description
}

// GetIsSystem 获取是否为系统角色
func (r *RoleDTO) GetIsSystem() bool {
	return r.IsSystem
}

// GetIsActive 获取是否活跃
func (r *RoleDTO) GetIsActive() bool {
	return r.IsActive
}

// GetPriority 获取优先级
func (r *RoleDTO) GetPriority() int {
	return r.Priority
}

// GetCreatedAt 获取创建时间
func (r *RoleDTO) GetCreatedAt() time.Time {
	return r.CreatedAt
}

// GetPermissionCount 获取权限数量
func (r *RoleDTO) GetPermissionCount() int {
	return r.PermissionCount
}

// GetUpdatedAt 获取更新时间
func (r *RoleDTO) GetUpdatedAt() time.Time {
	return r.UpdatedAt
}

// ==================== 业务方法 ====================

// UpdateInfo 更新角色信息
func (r *RoleDTO) UpdateInfo(name, description string) error {
	if name != "" {
		if len(name) > 100 {
			return errors.NewValidationError("角色名称过长，最大100个字符")
		}
		r.Name = strings.TrimSpace(name)
	}

	if description != "" {
		if len(description) > 500 {
			return errors.NewValidationError("描述过长，最大500个字符")
		}
		r.Description = strings.TrimSpace(description)
	}

	r.UpdatedAt = time.Now()
	return nil
}

// SetPriority 设置角色优先级
func (r *RoleDTO) SetPriority(priority int) error {
	if priority < 0 {
		return errors.NewValidationError("优先级必须为非负数")
	}

	if r.IsSystem && priority < 100 {
		return errors.NewBusinessError("系统角色的优先级至少为100")
	}

	r.Priority = priority
	r.UpdatedAt = time.Now()
	return nil
}

// Activate 激活角色
func (r *RoleDTO) Activate() error {
	if r.IsActive {
		return errors.NewBusinessError("角色已经是激活状态")
	}

	r.IsActive = true
	r.UpdatedAt = time.Now()
	return nil
}

// Deactivate 停用角色
func (r *RoleDTO) Deactivate() error {
	if !r.IsActive {
		return errors.NewBusinessError("角色已经是停用状态")
	}

	if r.IsSystem {
		return errors.NewBusinessError("系统角色不能被停用")
	}

	r.IsActive = false
	r.UpdatedAt = time.Now()
	return nil
}

// IsGlobalRole 检查是否为全局角色
func (r *RoleDTO) IsGlobalRole() bool {
	return r.OrgID == ""
}

// IsOrgRole 检查是否为组织角色
func (r *RoleDTO) IsOrgRole() bool {
	return r.OrgID != ""
}

// CanModify 检查角色是否可以被修改
func (r *RoleDTO) CanModify() bool {
	return !r.IsSystem
}

// CanDelete 检查角色是否可以被删除
func (r *RoleDTO) CanDelete() bool {
	return !r.IsSystem
}

// GetRoleHierarchy 获取角色层级（用于角色继承）
func (r *RoleDTO) GetRoleHierarchy() int {
	// 根据角色代码确定层级
	// super_admin > admin > manager > user
	roleHierarchy := map[string]int{
		permissionConstant.RoleCodeSuperAdmin: 100,
		permissionConstant.RoleCodeAdmin:      80,
		permissionConstant.RoleCodeOrgAdmin:   70,
		permissionConstant.RoleCodeManager:    50,
		permissionConstant.RoleCodeUser:       10,
		permissionConstant.RoleCodeGuest:      0,
	}

	if level, exists := roleHierarchy[r.Code]; exists {
		return level
	}

	return r.Priority
}

// IsHigherThan 检查当前角色是否高于另一个角色
func (r *RoleDTO) IsHigherThan(other *RoleDTO) bool {
	return r.GetRoleHierarchy() > other.GetRoleHierarchy()
}

// Clone 克隆角色（用于创建类似角色）
func (r *RoleDTO) Clone(newCode, newName string) *RoleDTO {
	now := time.Now()
	clone := &RoleDTO{
		TenantID:    r.TenantID,
		OrgID:       r.OrgID,
		Code:        newCode,
		Name:        newName,
		Description: r.Description,
		IsSystem:    false, // 克隆的角色不是系统角色
		IsActive:    true,
		Priority:    r.Priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return clone
}

// ValidateCode 验证角色代码
func (r *RoleDTO) ValidateCode(code string) error {
	code = strings.TrimSpace(code)

	if code == "" {
		return errors.NewValidationError("角色代码不能为空")
	}

	if len(code) < 2 {
		return errors.NewValidationError("角色代码至少需要2个字符")
	}

	if len(code) > 50 {
		return errors.NewValidationError("角色代码过长，最大50个字符")
	}

	// 检查角色代码格式（只允许字母、数字、下划线）
	for _, char := range code {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return errors.NewValidationError("角色代码只能包含字母、数字和下划线")
		}
	}

	return nil
}

// ToDTO 转换为角色DTO
func (r *RoleDTO) ToDTO() *RoleDTO {
	return r
}
