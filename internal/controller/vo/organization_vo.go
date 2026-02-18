package vo

// ==================== Organization 请求/响应 ====================

// CreateOrganizationRequest 创建组织请求
type CreateOrganizationRequest struct {
	ParentID    string `json:"parent_id"`
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

// UpdateOrganizationRequest 更新组织请求
type UpdateOrganizationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
	IsActive    *bool  `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

// DeleteOrganizationRequest 删除组织请求
type DeleteOrganizationRequest struct {
	ID string `json:"id" binding:"required"`
}

// AssignAccountToOrgRequest 分配账户到组织请求
type AssignAccountToOrgRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	OrgID     string `json:"org_id" binding:"required"`
}

// RemoveAccountFromOrgRequest 从组织移除账户请求
type RemoveAccountFromOrgRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	OrgID     string `json:"org_id" binding:"required"`
}

// OrganizationResponse 组织响应
type OrganizationResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	ParentID    string `json:"parent_id,omitempty"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Level       int    `json:"level"`
	Path        string `json:"path"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	IsRoot      bool   `json:"is_root"`
	IsLeaf      bool   `json:"is_leaf"`
	UserCount   int    `json:"user_count"`
	ChildCount  int    `json:"child_count"`
}

// OrganizationListResponse 组织列表响应
type OrganizationListResponse struct {
	Data  []OrganizationResponse `json:"data"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Size  int                    `json:"size"`
}
