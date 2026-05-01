package dto

import (
	"strings"
	"time"

	"auth-perm/internal/domain/permission/constant"
)

// PermissionResourceDTO 权限资源关联传输对象
type PermissionResourceDTO struct {
	ID           string    `json:"id"`
	PermissionID string    `json:"permission_id"`
	ResourceID   string    `json:"resource_id"`
	ResourceType string    `json:"resource_type"` // api_path, menu, button
	ResourceName string    `json:"resource_name"`
	TenantID     string    `json:"tenant_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AllResourceTypes 所有资源类型（委托给 constant 包）
var AllResourceTypes = constant.AllResourceTypes

// IsValidResourceType 检查资源类型是否有效
func (pr *PermissionResourceDTO) IsValidResourceType() bool {
	for _, rt := range AllResourceTypes {
		if pr.ResourceType == rt {
			return true
		}
	}
	return false
}

// Sanitize 清理和标准化数据
func (pr *PermissionResourceDTO) Sanitize() {
	pr.ResourceID = strings.TrimSpace(pr.ResourceID)
	pr.ResourceType = strings.ToLower(strings.TrimSpace(pr.ResourceType))
	pr.ResourceName = strings.TrimSpace(pr.ResourceName)
}

// GetDisplayName 获取显示名称
func (pr *PermissionResourceDTO) GetDisplayName() string {
	if pr.ResourceName != "" {
		return pr.ResourceName
	}
	return pr.ResourceID
}

// IsAPIPath 检查是否为API路径
func (pr *PermissionResourceDTO) IsAPIPath() bool {
	return pr.ResourceType == constant.ResourceTypeAPIPath
}

// IsMenu 检查是否为菜单
func (pr *PermissionResourceDTO) IsMenu() bool {
	return pr.ResourceType == constant.ResourceTypeMenu
}

// IsButton 检查是否为按钮
func (pr *PermissionResourceDTO) IsButton() bool {
	return pr.ResourceType == constant.ResourceTypeButton
}

// IsField 检查是否为字段
func (pr *PermissionResourceDTO) IsField() bool {
	return pr.ResourceType == constant.ResourceTypeField
}

// IsOther 检查是否为其他资源
func (pr *PermissionResourceDTO) IsOther() bool {
	return pr.ResourceType == constant.ResourceTypeOther
}
