package param

import (
	"fmt"
)

// CreatePermissionResourceParams 创建权限资源关联参数
type CreatePermissionResourceParams struct {
	PermissionID string `json:"permission_id"`
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`
	TenantID     string `json:"tenant_id"`
}

// Validate 验证创建参数
func (p *CreatePermissionResourceParams) Validate() error {
	if p.PermissionID == "" {
		return errEmpty("权限ID")
	}
	if p.ResourceID == "" {
		return errEmpty("资源ID")
	}
	if p.ResourceType == "" {
		return errEmpty("资源类型")
	}
	// tenant_id 为可选字段
	return nil
}

// UpdatePermissionResourceParams 更新权限资源关联参数
type UpdatePermissionResourceParams struct {
	ID           string `json:"id"`
	ResourceName string `json:"resource_name"`
}

// Validate 验证更新参数
func (p *UpdatePermissionResourceParams) Validate() error {
	if p.ID == "" {
		return errEmpty("ID")
	}
	return nil
}

// DeletePermissionResourceParams 删除权限资源关联参数
type DeletePermissionResourceParams struct {
	ID string `json:"id"`
}

// Validate 验证删除参数
func (p *DeletePermissionResourceParams) Validate() error {
	if p.ID == "" {
		return errEmpty("ID")
	}
	return nil
}

// ListPermissionResourceParams 列表查询参数
type ListPermissionResourceParams struct {
	PermissionID string `form:"permission_id"`
	ResourceType string `form:"resource_type"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

// NewListPermissionResourceParams 创建列表查询参数
func NewListPermissionResourceParams(permissionID, resourceType string, page, pageSize int) *ListPermissionResourceParams {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &ListPermissionResourceParams{
		PermissionID: permissionID,
		ResourceType: resourceType,
		Page:         page,
		PageSize:     pageSize,
	}
}

// Validate 验证列表查询参数
func (p *ListPermissionResourceParams) Validate() error {
	if p.PermissionID == "" {
		return errEmpty("权限ID")
	}
	return nil
}

// BindPermissionResourcesParams 绑定权限资源参数（批量）
type BindPermissionResourcesParams struct {
	PermissionID string                `json:"permission_id"`
	Resources    []ResourceBindingItem `json:"resources"`
	TenantID     string                `json:"tenant_id"`
}

// ResourceBindingItem 资源绑定项
type ResourceBindingItem struct {
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`
}

// Validate 验证绑定参数
func (p *BindPermissionResourcesParams) Validate() error {
	if p.PermissionID == "" {
		return errEmpty("权限ID")
	}
	if len(p.Resources) == 0 {
		return errEmpty("资源列表")
	}
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	for i, r := range p.Resources {
		if r.ResourceID == "" {
			return &ValidationError{Field: fmt.Sprintf("resources[%d].resource_id", i), Message: "不能为空"}
		}
		if r.ResourceType == "" {
			return &ValidationError{Field: fmt.Sprintf("resources[%d].resource_type", i), Message: "不能为空"}
		}
	}
	return nil
}

// UnbindPermissionResourcesParams 解绑权限资源参数
type UnbindPermissionResourcesParams struct {
	PermissionID string   `json:"permission_id"`
	ResourceIDs  []string `json:"resource_ids"`
}

// Validate 验证解绑参数
func (p *UnbindPermissionResourcesParams) Validate() error {
	if p.PermissionID == "" {
		return errEmpty("权限ID")
	}
	if len(p.ResourceIDs) == 0 {
		return errEmpty("资源ID列表")
	}
	return nil
}

// GetAccountResourcesParams 获取账户可访问资源参数
type GetAccountResourcesParams struct {
	AccountID    string `form:"account_id"`
	ResourceType string `form:"resource_type"`
}

// NewGetAccountResourcesParams 创建获取账户资源参数
func NewGetAccountResourcesParams(accountID, resourceType string) *GetAccountResourcesParams {
	return &GetAccountResourcesParams{
		AccountID:    accountID,
		ResourceType: resourceType,
	}
}

// Validate 验证获取账户资源参数
func (p *GetAccountResourcesParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	return nil
}

// CheckResourcePermissionParams 检查资源权限参数
type CheckResourcePermissionParams struct {
	AccountID    string `json:"account_id"`
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
}

// NewCheckResourcePermissionParams 创建检查资源权限参数
func NewCheckResourcePermissionParams(accountID, resourceID, resourceType string) *CheckResourcePermissionParams {
	return &CheckResourcePermissionParams{
		AccountID:    accountID,
		ResourceID:   resourceID,
		ResourceType: resourceType,
	}
}

// Validate 验证检查资源权限参数
func (p *CheckResourcePermissionParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	if p.ResourceID == "" {
		return errEmpty("资源ID")
	}
	if p.ResourceType == "" {
		return errEmpty("资源类型")
	}
	return nil
}
