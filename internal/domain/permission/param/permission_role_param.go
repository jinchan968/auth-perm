package param

// ==================== Permission 参数 ====================

// CreatePermissionParams 创建权限参数
type CreatePermissionParams struct {
	TenantID    string `json:"tenant_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

// Validate 验证创建权限参数
func (p *CreatePermissionParams) Validate() error {
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	// Code字段改为可选，后端自动生成
	if p.Name == "" {
		return errEmpty("权限名称")
	}
	return nil
}

// UpdatePermissionParams 更新权限参数
type UpdatePermissionParams struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

// Validate 验证更新权限参数
func (p *UpdatePermissionParams) Validate() error {
	if p.ID == "" {
		return errEmpty("权限ID")
	}
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// DeletePermissionParams 删除权限参数
type DeletePermissionParams struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
}

// Validate 验证删除权限参数
func (p *DeletePermissionParams) Validate() error {
	if p.ID == "" {
		return errEmpty("权限ID")
	}
	// TenantID改为可选，如果为空则从权限中获取
	return nil
}

// GetPermissionParams 获取权限详情参数
type GetPermissionParams struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
}

// Validate 验证获取权限详情参数
func (p *GetPermissionParams) Validate() error {
	if p.ID == "" {
		return errEmpty("权限ID")
	}
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// ListPermissionParams 权限列表查询参数
type ListPermissionParams struct {
	TenantID   string `form:"tenant_id"`
	Code       string `form:"code"`
	Name       string `form:"name"`
	IsActive   *bool  `form:"is_active"`
	IsSystem   *bool  `form:"is_system"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	IncludeAll bool   `form:"include_all"` // 是否包含所有（包括未激活）
}

// NewListPermissionParams 创建权限列表查询参数
func NewListPermissionParams(tenantID string, page, pageSize int) *ListPermissionParams {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &ListPermissionParams{
		TenantID: tenantID,
		Page:     page,
		PageSize: pageSize,
	}
}

// Validate 验证列表查询参数
func (p *ListPermissionParams) Validate() error {
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// ==================== Role 参数 ====================

// CreateRoleParams 创建角色参数
type CreateRoleParams struct {
	TenantID    string `json:"tenant_id"`
	OrgID       string `json:"org_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	IsSystem    bool   `json:"is_system"`
}

// Validate 验证创建角色参数
func (p *CreateRoleParams) Validate() error {
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	// Code字段改为可选，后端自动生成
	if p.Name == "" {
		return errEmpty("角色名称")
	}
	return nil
}

// UpdateRoleParams 更新角色参数
type UpdateRoleParams struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	IsActive    *bool  `json:"is_active"`
}

// Validate 验证更新角色参数
func (p *UpdateRoleParams) Validate() error {
	if p.ID == "" {
		return errEmpty("角色ID")
	}
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// DeleteRoleParams 删除角色参数
type DeleteRoleParams struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
}

// Validate 验证删除角色参数
func (p *DeleteRoleParams) Validate() error {
	if p.ID == "" {
		return errEmpty("角色ID")
	}
	// TenantID改为可选，如果为空则从角色中获取
	return nil
}

// GetRoleParams 获取角色详情参数
type GetRoleParams struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
}

// Validate 验证获取角色详情参数
func (p *GetRoleParams) Validate() error {
	if p.ID == "" {
		return errEmpty("角色ID")
	}
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// ListRoleParams 角色列表查询参数
type ListRoleParams struct {
	TenantID   string `form:"tenant_id"`
	OrgID      string `form:"org_id"`
	Code       string `form:"code"`
	Name       string `form:"name"`
	IsActive   *bool  `form:"is_active"`
	IsSystem   *bool  `form:"is_system"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	IncludeAll bool   `form:"include_all"`
}

// NewListRoleParams 创建角色列表查询参数
func NewListRoleParams(tenantID string, page, pageSize int) *ListRoleParams {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &ListRoleParams{
		TenantID: tenantID,
		Page:     page,
		PageSize: pageSize,
	}
}

// Validate 验证列表查询参数
func (p *ListRoleParams) Validate() error {
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// ==================== Role-Permission 关联参数 ====================

// AssignPermissionToRoleParams 分配权限给角色参数
type AssignPermissionToRoleParams struct {
	RoleID        string   `json:"role_id"`
	PermissionIDs []string `json:"permission_ids"`
	TenantID      string   `json:"tenant_id"`
}

// Validate 验证分配权限参数
func (p *AssignPermissionToRoleParams) Validate() error {
	if p.RoleID == "" {
		return errEmpty("角色ID")
	}
	// PermissionIDs 允许为空（表示清空该角色的所有权限）
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// RemovePermissionFromRoleParams 移除角色权限参数
type RemovePermissionFromRoleParams struct {
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
	TenantID     string `json:"tenant_id"`
}

// Validate 验证移除权限参数
func (p *RemovePermissionFromRoleParams) Validate() error {
	if p.RoleID == "" {
		return errEmpty("角色ID")
	}
	if p.PermissionID == "" {
		return errEmpty("权限ID")
	}
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// ==================== Account-Role 关联参数 ====================

// AssignRoleToAccountParams 分配角色给账户参数
type AssignRoleToAccountParams struct {
	AccountID string   `json:"account_id"`
	RoleIDs   []string `json:"role_ids"`
	TenantID  string   `json:"tenant_id"`
}

// Validate 验证分配角色参数
func (p *AssignRoleToAccountParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	if len(p.RoleIDs) == 0 {
		return errEmpty("角色ID列表")
	}
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// RemoveRoleFromAccountParams 移除账户角色参数
type RemoveRoleFromAccountParams struct {
	AccountID string `json:"account_id"`
	RoleID    string `json:"role_id"`
	TenantID  string `json:"tenant_id"`
}

// Validate 验证移除角色参数
func (p *RemoveRoleFromAccountParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	if p.RoleID == "" {
		return errEmpty("角色ID")
	}
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	return nil
}
