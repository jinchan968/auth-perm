package vo

// ==================== Permission 请求/响应 ====================

// CheckPermissionRequest 检查权限请求
type CheckPermissionRequest struct {
	PermissionCode string `json:"permission_code" binding:"required"`
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	TenantID    string `json:"tenant_id" binding:"required"`
	Code        string `json:"code"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	ID          string `json:"id" binding:"required"`
	TenantID    string `json:"tenant_id" binding:"required"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

// DeletePermissionRequest 删除权限请求
type DeletePermissionRequest struct {
	ID       string `json:"id" binding:"required"`
	TenantID string `json:"tenant_id"`
}

// PermissionResponse 权限响应
type PermissionResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	IsSystem    bool   `json:"is_system"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// PermissionListResponse 权限列表响应
type PermissionListResponse struct {
	Data  []PermissionResponse `json:"data"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
}

// ==================== Role 请求/响应 ====================

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	TenantID    string `json:"tenant_id" binding:"required"`
	OrgID       string `json:"org_id"`
	Code        string `json:"code"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	IsSystem    bool   `json:"is_system"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	ID          string `json:"id" binding:"required"`
	TenantID    string `json:"tenant_id" binding:"required"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	IsActive    *bool  `json:"is_active"`
}

// DeleteRoleRequest 删除角色请求
type DeleteRoleRequest struct {
	ID       string `json:"id" binding:"required"`
	TenantID string `json:"tenant_id" binding:"required"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	OrgID       string `json:"org_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	IsSystem    bool   `json:"is_system"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Data  []RoleResponse `json:"data"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
}

// ==================== Role-Permission 关联请求/响应 ====================

// AssignPermissionToRoleRequest 分配权限给角色请求
type AssignPermissionToRoleRequest struct {
	RoleID        string   `json:"role_id" binding:"required"`
	PermissionIDs []string `json:"permission_ids" binding:"required"`
	TenantID      string   `json:"tenant_id" binding:"required"`
}

// RemovePermissionFromRoleRequest 移除角色权限请求
type RemovePermissionFromRoleRequest struct {
	RoleID       string `json:"role_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
	TenantID     string `json:"tenant_id" binding:"required"`
}

// ==================== Account-Role 关联请求/响应 ====================

// AssignRoleToAccountRequest 分配角色给账户请求
type AssignRoleToAccountRequest struct {
	AccountID string   `json:"account_id" binding:"required"`
	RoleIDs   []string `json:"role_ids" binding:"required"`
	TenantID  string   `json:"tenant_id" binding:"required"`
}

// RemoveRoleFromAccountRequest 移除账户角色请求
type RemoveRoleFromAccountRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	RoleID    string `json:"role_id" binding:"required"`
	TenantID  string `json:"tenant_id" binding:"required"`
}
