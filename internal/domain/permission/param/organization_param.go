package param

// ==================== 组织参数 ====================

// CreateOrganizationParams 创建组织参数
type CreateOrganizationParams struct {
	TenantID    string `json:"tenant_id"`
	ParentID    string `json:"parent_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

// Validate 验证创建组织参数
func (p *CreateOrganizationParams) Validate() error {
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	if p.Code == "" {
		return errEmpty("组织编码")
	}
	if p.Name == "" {
		return errEmpty("组织名称")
	}
	return nil
}

// UpdateOrganizationParams 更新组织参数
type UpdateOrganizationParams struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
	IsActive    *bool  `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

// Validate 验证更新组织参数
func (p *UpdateOrganizationParams) Validate() error {
	if p.ID == "" {
		return errEmpty("组织ID")
	}
	return nil
}

// DeleteOrganizationParams 删除组织参数
type DeleteOrganizationParams struct {
	ID string `json:"id"`
}

// Validate 验证删除组织参数
func (p *DeleteOrganizationParams) Validate() error {
	if p.ID == "" {
		return errEmpty("组织ID")
	}
	return nil
}

// GetOrganizationParams 获取组织参数
type GetOrganizationParams struct {
	ID string `json:"id"`
}

// Validate 验证获取组织参数
func (p *GetOrganizationParams) Validate() error {
	if p.ID == "" {
		return errEmpty("组织ID")
	}
	return nil
}

// ListOrganizationsParams 列出组织参数
type ListOrganizationsParams struct {
	TenantID string `json:"tenant_id"`
	ParentID string `json:"parent_id"`
	Keyword  string `json:"keyword"`
	Page     int    `json:"page"`
	Size     int    `json:"size"`
}

// Validate 验证列出组织参数
func (p *ListOrganizationsParams) Validate() error {
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 || p.Size > 100 {
		p.Size = 10
	}
	return nil
}

// AssignAccountToOrgParams 分配账户到组织参数
type AssignAccountToOrgParams struct {
	AccountID string `json:"account_id"`
	OrgID     string `json:"org_id"`
	TenantID  string `json:"tenant_id"`
}

// Validate 验证分配账户到组织参数
func (p *AssignAccountToOrgParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	if p.OrgID == "" {
		return errEmpty("组织ID")
	}
	if p.TenantID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// RemoveAccountFromOrgParams 从组织移除账户参数
type RemoveAccountFromOrgParams struct {
	AccountID string `json:"account_id"`
	OrgID     string `json:"org_id"`
}

// Validate 验证从组织移除账户参数
func (p *RemoveAccountFromOrgParams) Validate() error {
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	if p.OrgID == "" {
		return errEmpty("组织ID")
	}
	return nil
}
